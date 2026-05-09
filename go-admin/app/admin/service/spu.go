package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	platformModels "go-admin/app/platform/models"
	platformService "go-admin/app/platform/service"
	platformDto "go-admin/app/platform/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

const (
	// SpuModuleKey / SpuBusinessType 是 SPU 在 platform workflow 中的标识。
	SpuModuleKey    = "admin"
	SpuBusinessType = "spu"

	// SpuDefaultDefinitionKey 创建审核默认流程定义 key。
	SpuDefaultDefinitionKey = "spu_create_review"
)

// Spu 产品标准单元服务。
//
// 关键能力：
//   - GetPage：接 dataScope 数据权限 + workflow_business_binding LEFT JOIN 暴露 workflow 状态。
//   - SubmitForReview：创建 workflow 实例 + 更新 SPU.status=2。
//
// SPU.status 终态由 platform workflow 在终态时通过 RegisterTerminalHandler 回调 spu_workflow_handler.go 中的 onSpuWorkflowTerminal 写回。
type Spu struct {
	service.Service
}

// GetPage 列表查询。
//
//   - dataScope：通过 actions.Permission 注入；data_scope=5 仅自己的，data_scope=3/4 按部门，
//     data_scope=1 全部（含 admin 默认行为）。
//   - workflow status：LEFT JOIN wf_business_binding 暴露当前工作流状态，便于前端按流程态过滤。
func (e *Spu) GetPage(c *dto.SpuPageReq, p *actions.DataPermission, list *[]dto.SpuListItem, count *int64) error {
	var data models.Spu
	q := e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		)

	if err := q.Count(count).Error; err != nil {
		e.Log.Errorf("Spu count: %s", err)
		return err
	}

	pageSize := c.GetPageSize()
	pageIndex := c.GetPageIndex()
	offset := (pageIndex - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	rows := make([]models.Spu, 0)
	if err := q.
		Order("spu.spu_id DESC").
		Limit(pageSize).Offset(offset).
		Find(&rows).Error; err != nil {
		e.Log.Errorf("Spu find: %s", err)
		return err
	}
	if len(rows) == 0 {
		*list = []dto.SpuListItem{}
		return nil
	}

	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.SpuId)
	}

	// 一次性拉 workflow binding（按 SPU id），避免 N+1
	type bindingRow struct {
		BusinessId     string
		WorkflowStatus string
		Title          string
	}
	bindings := make([]bindingRow, 0)
	if err := e.Orm.Table("wf_business_binding").
		Select("business_id, workflow_status, title").
		Where("module_key = ? AND business_type = ? AND business_id IN ?", SpuModuleKey, SpuBusinessType, idsToStrings(ids)).
		Find(&bindings).Error; err != nil {
		// binding 不存在不阻塞列表；旁路 workflow 视为空
		e.Log.Warnf("Spu list binding lookup: %s", err)
	}
	bindingMap := make(map[string]bindingRow, len(bindings))
	for _, b := range bindings {
		bindingMap[b.BusinessId] = b
	}

	out := make([]dto.SpuListItem, 0, len(rows))
	for _, r := range rows {
		item := dto.SpuListItem{Spu: r}
		if b, ok := bindingMap[int64ToString(r.SpuId)]; ok {
			item.WorkflowStatus = b.WorkflowStatus
			item.WorkflowTitle = b.Title
		}
		out = append(out, item)
	}
	*list = out
	return nil
}

// Get 单条查询，含 dataScope 校验。
func (e *Spu) Get(c *dto.SpuGetReq, p *actions.DataPermission, item *dto.SpuListItem) error {
	var spu models.Spu
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission((&models.Spu{}).TableName(), p)).
		First(&spu, c.GetId()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("SPU 不存在或无权查看")
		}
		e.Log.Errorf("Spu Get: %s", err)
		return err
	}
	out := dto.SpuListItem{Spu: spu}

	var binding platformModels.WorkflowBusinessBinding
	if err := e.Orm.Where("module_key = ? AND business_type = ? AND business_id = ?",
		SpuModuleKey, SpuBusinessType, int64ToString(spu.SpuId)).First(&binding).Error; err == nil {
		out.WorkflowStatus = binding.WorkflowStatus
		out.WorkflowTitle = binding.Title
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		e.Log.Warnf("Spu Get binding: %s", err)
	}
	*item = out
	return nil
}

// Insert 新增 SPU。SPU 默认 status=Draft；SubmitForReview 才会推进到 Review。
func (e *Spu) Insert(c *dto.SpuInsertReq) (int64, error) {
	if strings.TrimSpace(c.SpuCode) == "" {
		return 0, errors.New("SPU 编码不能为空")
	}
	var data models.Spu
	c.Generate(&data)
	if err := e.Orm.Create(&data).Error; err != nil {
		e.Log.Errorf("Spu Insert: %s", err)
		return 0, err
	}
	c.SpuId = data.SpuId
	return data.SpuId, nil
}

// Update 修改 SPU。审核中的 SPU 不允许直接编辑（应先撤回审批）。
func (e *Spu) Update(c *dto.SpuUpdateReq, p *actions.DataPermission) error {
	var existing models.Spu
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission((&models.Spu{}).TableName(), p)).
		First(&existing, c.SpuId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("SPU 不存在或无权修改")
		}
		return err
	}
	if existing.Status == models.SpuStatusReviewing {
		return errors.New("SPU 处于审核中，不可直接修改，请先撤回审批")
	}
	c.Generate(&existing)
	if err := e.Orm.Save(&existing).Error; err != nil {
		e.Log.Errorf("Spu Update: %s", err)
		return err
	}
	return nil
}

// Remove 批量删除（dataScope 过滤后再删）。
func (e *Spu) Remove(c *dto.SpuDeleteReq, p *actions.DataPermission) error {
	if len(c.Ids) == 0 {
		return errors.New("ids 不能为空")
	}
	tableName := (&models.Spu{}).TableName()
	var allowed []int64
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission(tableName, p)).
		Where(tableName+".spu_id IN ?", c.Ids).
		Pluck(tableName+".spu_id", &allowed).Error; err != nil {
		return err
	}
	if len(allowed) == 0 {
		return errors.New("SPU 不存在或无权删除")
	}
	// 审核中不允许删除
	var reviewingCount int64
	if err := e.Orm.Model(&models.Spu{}).
		Where("spu_id IN ? AND status = ?", allowed, models.SpuStatusReviewing).
		Count(&reviewingCount).Error; err != nil {
		return err
	}
	if reviewingCount > 0 {
		return errors.New("存在审核中的 SPU，不能删除")
	}
	return e.Orm.Transaction(func(tx *gorm.DB) error {
		// 关联 SKU 也级联软删
		if err := tx.Where("spu_id IN ?", allowed).Delete(&models.Sku{}).Error; err != nil {
			return err
		}
		var data models.Spu
		if err := tx.Delete(&data, allowed).Error; err != nil {
			return err
		}
		return nil
	})
}

// SubmitForReview 提交 SPU 进入审核流程。
//
// 流程：
//  1. 加载 SPU + dataScope 校验；status 必须为 Draft 或 Rejected。
//  2. 解析流程定义（优先 req.DefinitionId；否则按 SpuDefaultDefinitionKey 取启用版本）。
//  3. 调用 platform Workflow.Start：创建 wf_instance + wf_business_binding。
//  4. 更新 SPU：status=Reviewing、submitted_at=now、workflow_instance_id=instance.id。
//
// 终态由 platform workflow 在 Approve/Reject/Withdraw 时通过 RegisterTerminalHandler 回调本模块写回。
func (e *Spu) SubmitForReview(c *gin.Context, p *actions.DataPermission, req *dto.SpuSubmitReq) (int, error) {
	var spu models.Spu
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission((&models.Spu{}).TableName(), p)).
		First(&spu, req.SpuId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("SPU 不存在或无权提交")
		}
		return 0, err
	}
	if spu.Status != models.SpuStatusDraft && spu.Status != models.SpuStatusRejected {
		return 0, errors.New("SPU 当前状态不允许提交审核")
	}

	defID, err := e.resolveDefinition(req.DefinitionId)
	if err != nil {
		return 0, err
	}

	wf := &platformService.Workflow{Service: e.Service}
	wfReq := &platformDto.WorkflowInstanceStartReq{
		DefinitionId: defID,
		ModuleKey:    SpuModuleKey,
		BusinessType: SpuBusinessType,
		BusinessId:   int64ToString(spu.SpuId),
		BusinessNo:   spu.SpuCode,
		Title:        spu.SpuName,
		Remark:       req.Remark,
	}
	detail, err := wf.Start(c, wfReq)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	if err := e.Orm.Model(&models.Spu{}).
		Where("spu_id = ?", spu.SpuId).
		Updates(map[string]interface{}{
			"status":               models.SpuStatusReviewing,
			"submitted_at":         &now,
			"workflow_instance_id": detail.Instance.InstanceId,
		}).Error; err != nil {
		e.Log.Errorf("Spu SubmitForReview update spu: %s", err)
		return detail.Instance.InstanceId, err
	}
	return detail.Instance.InstanceId, nil
}

// resolveDefinition 优先用调用方提供的 DefinitionId；否则按默认 key 查启用版本。
func (e *Spu) resolveDefinition(defID int) (int, error) {
	if defID > 0 {
		var n int64
		if err := e.Orm.Model(&platformModels.WorkflowDefinition{}).
			Where("definition_id = ? AND status = ?", defID, "2").
			Count(&n).Error; err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, errors.New("流程定义不存在或未启用")
		}
		return defID, nil
	}
	var def platformModels.WorkflowDefinition
	if err := e.Orm.Where("definition_key = ? AND status = ?", SpuDefaultDefinitionKey, "2").
		Order("version DESC").First(&def).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("默认 SPU 创建审核流程未配置或未启用")
		}
		return 0, err
	}
	return def.DefinitionId, nil
}

// GoOffline 下架 SPU（status=3 && is_online=true → is_online=false），并级联下架所有 SKU。
//
// 状态机约束：GoOffline 仅允许在 status=Approved(3) && is_online=true 时执行。
// 级联：将该 SPU 下所有 SKU 置为 SkuStatusDisabled(1)。
// hook 位预留（架构师 §8.4）：before/after hook 点留空，后续按需扩展。
func (e *Spu) GoOffline(spuId int64, operatorId int, p *actions.DataPermission) error {
	// --- before-hook 位 ---

	var spu models.Spu
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission((&models.Spu{}).TableName(), p)).
		First(&spu, spuId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("SPU 不存在或无权操作")
		}
		return err
	}
	if spu.Status != models.SpuStatusApproved || !spu.IsOnline {
		return errors.New("SPU 当前状态不允许下架：仅 status=审核通过(3) 且 is_online=true 的 SPU 可下架")
	}

	return e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Spu{}).
			Where("spu_id = ?", spuId).
			Updates(map[string]interface{}{
				"is_online":  false,
				"update_by":  operatorId,
			}).Error; err != nil {
			e.Log.Errorf("Spu GoOffline update spu: %s", err)
			return err
		}
		if err := tx.Model(&models.Sku{}).
			Where("spu_id = ?", spuId).
			Updates(map[string]interface{}{
				"status":    models.SkuStatusDisabled,
				"update_by": operatorId,
			}).Error; err != nil {
			e.Log.Errorf("Spu GoOffline cascade sku: %s", err)
			return err
		}
		// --- after-hook 位 ---
		return nil
	})
}

// GoOnline 上架 SPU（status=3 && is_online=false → is_online=true）。
//
// 状态机约束：GoOnline 仅允许在 status=Approved(3) && is_online=false 时执行。
// 不反向恢复 SKU（架构师 §8.3）。
// hook 位预留（架构师 §8.4）：before/after hook 点留空，后续按需扩展。
func (e *Spu) GoOnline(spuId int64, operatorId int, p *actions.DataPermission) error {
	// --- before-hook 位 ---

	var spu models.Spu
	if err := e.Orm.Model(&models.Spu{}).
		Scopes(actions.Permission((&models.Spu{}).TableName(), p)).
		First(&spu, spuId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("SPU 不存在或无权操作")
		}
		return err
	}
	if spu.Status != models.SpuStatusApproved || spu.IsOnline {
		return errors.New("SPU 当前状态不允许上架：仅 status=审核通过(3) 且 is_online=false 的 SPU 可上架")
	}

	if err := e.Orm.Model(&models.Spu{}).
		Where("spu_id = ?", spuId).
		Updates(map[string]interface{}{
			"is_online": true,
			"update_by": operatorId,
		}).Error; err != nil {
		e.Log.Errorf("Spu GoOnline update spu: %s", err)
		return err
	}

	// --- after-hook 位 ---
	return nil
}

// idsToStrings / int64ToString —— wf_business_binding.business_id 是 string 类型，
// SPU.spu_id 是 int64，所以列表查询时需要做一次类型转换。
func idsToStrings(ids []int64) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, int64ToString(id))
	}
	return out
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
