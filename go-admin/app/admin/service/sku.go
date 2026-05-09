package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

// Sku 具体可售单元服务。
// SKU 主管理页只读；写动作通过 SPU 编辑页的子表完成。
type Sku struct {
	service.Service
}

// GetPage 列表查询。LEFT JOIN spu，按 spu.create_by 继承 SPU 的 dataScope 过滤。
func (e *Sku) GetPage(c *dto.SkuPageReq, p *actions.DataPermission, list *[]dto.SkuListItem, count *int64) error {
	q := e.Orm.Model(&models.Sku{}).
		Select("sku.*, spu.spu_code as spu_code, spu.spu_name as spu_name, spu.status as spu_status").
		Joins("LEFT JOIN spu ON sku.spu_id = spu.spu_id").
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission("spu", p),
		)

	if err := q.Count(count).Error; err != nil {
		e.Log.Errorf("Sku GetPage count: %s", err)
		return err
	}

	pageSize := c.GetPageSize()
	pageIndex := c.GetPageIndex()
	offset := (pageIndex - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	if err := q.Order("sku.sku_id DESC").Limit(pageSize).Offset(offset).Find(list).Error; err != nil {
		e.Log.Errorf("Sku GetPage find: %s", err)
		return err
	}
	return nil
}

// Get 单条查询。LEFT JOIN spu 验证 dataScope，并填充 SpuCode/SpuName/SpuStatus。
func (e *Sku) Get(c *dto.SkuGetReq, p *actions.DataPermission, item *dto.SkuListItem) error {
	err := e.Orm.Model(&models.Sku{}).
		Select("sku.*, spu.spu_code as spu_code, spu.spu_name as spu_name, spu.status as spu_status").
		Joins("LEFT JOIN spu ON sku.spu_id = spu.spu_id").
		Scopes(actions.Permission("spu", p)).
		Where("sku.sku_id = ?", c.GetId()).
		First(item).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("查看对象不存在或无权查看")
	}
	if err != nil {
		e.Log.Errorf("Sku Get error: %s", err)
		return err
	}
	return nil
}

// Insert 新增 SKU（内部接口，由 SPU 编辑页发起）。
func (e *Sku) Insert(c *dto.SkuInsertReq) error {
	if err := e.checkSpuExists(c.SpuId); err != nil {
		return err
	}
	var data models.Sku
	c.Generate(&data)
	if err := e.Orm.Create(&data).Error; err != nil {
		e.Log.Errorf("Sku Insert: %s", err)
		return err
	}
	c.SkuId = data.SkuId
	return nil
}

// Update 修改 SKU（内部接口，由 SPU 编辑页发起）。
func (e *Sku) Update(c *dto.SkuUpdateReq) error {
	if err := e.checkSpuExists(c.SpuId); err != nil {
		return err
	}
	var existing models.Sku
	if err := e.Orm.First(&existing, c.SkuId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("SKU 不存在")
		}
		return err
	}
	c.Generate(&existing)
	db := e.Orm.Save(&existing)
	if err := db.Error; err != nil {
		e.Log.Errorf("Sku Update: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 批量删除（内部接口，由 SPU 编辑页发起）。
func (e *Sku) Remove(c *dto.SkuDeleteReq) error {
	if len(c.Ids) == 0 {
		return errors.New("ids 不能为空")
	}
	var data models.Sku
	db := e.Orm.Model(&data).Delete(&data, c.Ids)
	if err := db.Error; err != nil {
		e.Log.Errorf("Sku Remove: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

func (e *Sku) checkSpuExists(spuId int64) error {
	if spuId <= 0 {
		return errors.New("spuId 不能为空")
	}
	var cnt int64
	if err := e.Orm.Model(&models.Spu{}).Where("spu_id = ?", spuId).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("所属 SPU 不存在或已删除")
	}
	return nil
}
