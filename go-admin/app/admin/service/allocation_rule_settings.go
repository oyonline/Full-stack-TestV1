package service

import (
	"errors"
	"fmt"
	"go-admin/common/utils/collectors"
	"gorm.io/gorm/clause"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/excelUtils"
)

type AllocationRuleSettings struct {
	service.Service
}

// GetPage 获取AllocationRuleSettings列表
func (e *AllocationRuleSettings) GetPage(c *dto.AllocationRuleSettingsGetPageReq, p *actions.DataPermission, list *[]models.AllocationRuleSettings, count *int64) error {
	var err error
	var data models.AllocationRuleSettings

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("AllocationRuleSettingsService GetPage error:%s \r\n", err)
		return err
	}

	feeDetailsMap, _ := e.getFeeDetailsMap()
	ids := collectors.DistinctField(*list, func(v models.AllocationRuleSettings) int64 {
		return v.Id
	})
	ruleDeptMap := e.getRuleSettingMap(ids)
	for _, v := range *list {
		if feeDetails, ok := feeDetailsMap[v.BudgetFeeCategoryDetailsId]; ok {
			v.FeeName = feeDetails.FeeName
			v.FeeCode = feeDetails.FeeCode
		}
		if ruleDept, ok := ruleDeptMap[v.Id]; ok {
			v.AllocationRuleSettingsDept = ruleDept
		}
	}
	return nil
}

// Get 获取AllocationRuleSettings对象
func (e *AllocationRuleSettings) Get(d *dto.AllocationRuleSettingsGetReq, p *actions.DataPermission, model *models.AllocationRuleSettings) error {

	err := e.Orm.Model(*model).Scopes(actions.Permission(model.TableName(), p)).First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetAllocationRuleSettings error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}

	feeDetailsMap, _ := e.getFeeDetailsMap()
	ruleDeptMap := e.getRuleSettingMap([]int64{model.Id})
	if ruleDept, ok := ruleDeptMap[model.Id]; ok {
		model.AllocationRuleSettingsDept = ruleDept
	}
	if feeDetails, ok := feeDetailsMap[model.BudgetFeeCategoryDetailsId]; ok {
		model.FeeName = feeDetails.FeeName
		model.FeeCode = feeDetails.FeeCode
	}
	return nil
}

// Insert 创建AllocationRuleSettings对象
func (e *AllocationRuleSettings) Insert(c *dto.AllocationRuleSettingsInsertReq) error {
	var err error
	var data models.AllocationRuleSettings
	c.Generate(&data)

	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = tx.Create(&data).Error
	if err != nil {
		e.Log.Errorf("创建分摊规则失败:%s \r\n", err)
		tx.Rollback()
		return err
	}

	for _, v := range data.AllocationRuleSettingsDept {
		v.AllocationRuleSettingsId = data.Id
		v.CreateBy = data.CreateBy
		v.UpdateBy = data.UpdateBy
	}
	if err = tx.Save(&data.AllocationRuleSettingsDept).Error; err != nil {
		e.Log.Errorf("创建分摊规则映射费用承担部门失败:%s \r\n", err)
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Update 修改AllocationRuleSettings对象
func (e *AllocationRuleSettings) Update(c *dto.AllocationRuleSettingsUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.AllocationRuleSettings{}

	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	tx.Scopes(actions.Permission(data.TableName(), p)).First(&data, c.GetId())
	c.Generate(&data)

	if err = tx.Save(&data).Error; err != nil {
		e.Log.Errorf("编辑分摊规则失败:%s \r\n", err)
		tx.Rollback()
		return err
	}
	for _, v := range data.AllocationRuleSettingsDept {
		v.AllocationRuleSettingsId = data.Id
		v.CreateBy = data.CreateBy
		v.UpdateBy = data.UpdateBy
	}
	if err = tx.Unscoped().Where("allocation_rule_settings_id = ?", data.Id).Delete(&models.AllocationRuleSettingsDept{}).Error; err != nil {
		e.Log.Errorf("删除分摊规则映射费用承担部门失败:%s \r\n", err)
		tx.Rollback()
		return err
	}
	if err = tx.Save(&data.AllocationRuleSettingsDept).Error; err != nil {
		e.Log.Errorf("编辑分摊规则映射费用承担部门失败:%s \r\n", err)
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Remove 删除AllocationRuleSettings
func (e *AllocationRuleSettings) Remove(d *dto.AllocationRuleSettingsDeleteReq, p *actions.DataPermission) error {
	var data models.AllocationRuleSettings
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	db := tx.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Delete(&data, d.GetId())
	if err := db.Error; err != nil {
		e.Log.Errorf("删除分摊规则数据:%s \r\n", err)
		tx.Rollback()
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	if err := tx.Unscoped().Where("allocation_rule_settings_id in ?", d.GetId()).Delete(&models.AllocationRuleSettingsDept{}).Error; err != nil {
		e.Log.Errorf("删除分摊规则映射费用承担部门数据失败:%s \r\n", err)
		tx.Rollback()
		return err
	}
	return nil
}

func (e *AllocationRuleSettings) ImportData(req *dto.AllocationRuleSettingsImportReq, p *actions.DataPermission, userId int) error {
	var err error
	// 打开上传的文件
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	// 初始化字典映射
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.AllocationRuleSettingsExport](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	// 开始事务
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	_, feeDetailsCodeMap := e.getFeeDetailsMap()
	sysDeptMap := e.getSysDeptMap()
	// 批量更新
	for _, v := range detailData {
		var data1 models.AllocationRuleSettings
		data1.AllocationName = v.AllocationName
		if feeDetailsCode, ok := feeDetailsCodeMap[v.FeeCode]; ok {
			data1.BudgetFeeCategoryDetailsId = feeDetailsCode.Id
		} else {
			tx.Rollback()
			return errors.New(fmt.Sprintf("费用编码【%s】不存在", v.FeeCode))
		}
		data1.AllocationType = v.AllocationType
		data1.Status = v.Status
		data1.EffectiveDate = v.EffectiveDate
		data1.ExpiredDate = v.ExpiredDate
		data1.Description = v.Description
		data1.CreateBy = userId
		data1.UpdateBy = userId
		if err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "allocation_name"}, {Name: "budget_fee_category_details_id"}, {Name: "effective_date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"allocation_type",
				"status",
				"expired_date",
				"description",
				"update_by",
				"updated_at",
			}),
		}).Create(&data1).Error; err != nil {
			e.Log.Errorf("导入分摊规则数据失败:%s \r\n", err)
			tx.Rollback()
			return err
		}
		if v.RuleSettingsDeptStr != "" {
			var ruleDeptList []models.AllocationRuleSettingsDept
			for _, v1 := range strings.Split(v.RuleSettingsDeptStr, ",") {
				v2 := strings.Split(strings.TrimSpace(v1), "=")
				value1 := strings.TrimSpace(v2[0])
				value2 := strings.TrimSpace(v2[1])
				var dept1 models.AllocationRuleSettingsDept
				dept1.AllocationType = data1.AllocationType
				dept1.AllocationRuleSettingsId = data1.Id
				scaleSettings, _ := strconv.ParseFloat(value2, 64)
				dept1.ScaleSettings = scaleSettings
				if dept22, ok := sysDeptMap[value1]; ok {
					dept1.AssociationId = int64(dept22.DeptId)
				}
				dept1.CreateBy = userId
				dept1.UpdateBy = userId
				ruleDeptList = append(ruleDeptList, dept1)
			}
			if err = tx.Save(&ruleDeptList).Error; err != nil {
				e.Log.Errorf("导入分摊规则映射费用承担部门数据失败:%s \r\n", err)
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func (e *AllocationRuleSettings) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []dto.AllocationRuleSettingsExport
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.AllocationRuleSettingsExport](dictMappings)
	// 执行流式导出
	filename := "分摊规则设置模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *AllocationRuleSettings) Export(httpResponseWriter http.ResponseWriter, c *dto.AllocationRuleSettingsGetPageReq, p *actions.DataPermission) error {
	var err error
	var dataList []models.AllocationRuleSettings
	var exportData []dto.AllocationRuleSettingsExport
	var tableName = models.AllocationRuleSettings{}.TableName()
	err = e.Orm.Table(tableName).Scopes(cDto.MakeCondition(c.GetNeedSearch()), actions.Permission(tableName, p)).Find(&dataList).Error
	if err != nil {
		e.Log.Errorf("AllocationRuleSettingsService Export error:%s \r\n", err)
		return err
	}
	feeDetailsMap, _ := e.getFeeDetailsMap()
	ids := collectors.DistinctField(dataList, func(v models.AllocationRuleSettings) int64 {
		return v.Id
	})
	ruleSettingMap := e.getRuleSettingMap(ids)
	for _, v := range dataList {
		var data1 dto.AllocationRuleSettingsExport
		data1.AllocationName = v.AllocationName
		if feeDetails, ok := feeDetailsMap[v.BudgetFeeCategoryDetailsId]; ok {
			data1.FeeCode = feeDetails.FeeCode
		}
		data1.AllocationType = v.AllocationType
		data1.Status = v.Status
		data1.EffectiveDate = v.EffectiveDate
		data1.ExpiredDate = v.ExpiredDate
		data1.Description = v.Description
		if ruleDept, ok := ruleSettingMap[v.Id]; ok {
			var ruleSettingsDeptStrList []string
			for _, v1 := range ruleDept {
				ruleSettingsDeptStrList = append(ruleSettingsDeptStrList, fmt.Sprintf("%s=%g", v1.DeptPathName, v1.ScaleSettings))
			}
			data1.RuleSettingsDeptStr = strings.Join(ruleSettingsDeptStrList, ",")
		}
		exportData = append(exportData, data1)
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.AllocationRuleSettingsExport](dictMappings)
	// 执行流式导出
	filename := "分摊规则设置"
	return mapper.ExportToExcel(exportData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

func (e *AllocationRuleSettings) getFeeDetailsMap() (map[int64]models.BudgetFeeCategoryDetails, map[string]models.BudgetFeeCategoryDetails) {
	var feeDetailsList []models.BudgetFeeCategoryDetails
	e.Orm.Table("budget_fee_category_details").Where("view_type = ?", 2).Find(&feeDetailsList)
	feeDetailsMap := collectors.ToMapDynamic(feeDetailsList, func(v models.BudgetFeeCategoryDetails) int64 {
		return v.Id
	}, func(v models.BudgetFeeCategoryDetails) models.BudgetFeeCategoryDetails { return v })
	feeDetailsCodeMap := collectors.ToMapDynamic(feeDetailsList, func(v models.BudgetFeeCategoryDetails) string {
		return v.FeeCode
	}, func(v models.BudgetFeeCategoryDetails) models.BudgetFeeCategoryDetails { return v })
	return feeDetailsMap, feeDetailsCodeMap
}

func (e *AllocationRuleSettings) getRuleSettingMap(ids []int64) map[int64][]models.AllocationRuleSettingsDept {
	var ruleDeptList []models.AllocationRuleSettingsDept
	query := e.Orm.Table("allocation_rule_settings_dept t1").
		Joins("LEFT JOIN sys_dept t2 ON t2.dept_id = t1.association_id").
		Select("t1.*, t2.dept_name, t2.dept_path_name")

	// 如果 ID 集合不为空，添加 WHERE 条件
	if len(ids) > 0 {
		query = query.Where("t1.allocation_rule_settings_id IN ?", ids)
	}

	query.Scan(&ruleDeptList)
	ruleDeptMap := collectors.GroupBy(ruleDeptList, func(v models.AllocationRuleSettingsDept) int64 {
		return v.AllocationRuleSettingsId
	})
	return ruleDeptMap
}

func (e *AllocationRuleSettings) getSysDeptMap() map[string]models.SysDept {
	var sysDeptList []models.SysDept
	e.Orm.Table("sys_dept").Find(&sysDeptList)
	sysDeptMap := collectors.ToMapDynamic(sysDeptList, func(v models.SysDept) string {
		return v.DeptPathName
	}, func(v models.SysDept) models.SysDept { return v })
	return sysDeptMap
}
