package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"
)

// Sku 具体可售单元服务。
//
// 标准 CRUD 同时强制 SKU 必须挂在已存在的 SPU 上：Insert / Update 入库前会校验 spu_id。
type Sku struct {
	service.Service
}

// GetPage 列表查询，支持按 spu_id / sku_code 等过滤（由 search tag 驱动）。
func (e *Sku) GetPage(c *dto.SkuPageReq, list *[]models.Sku, count *int64) error {
	var data models.Sku
	err := e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("Sku GetPage error: %s", err)
		return err
	}
	return nil
}

// Get 单条查询。
func (e *Sku) Get(c *dto.SkuGetReq, model *models.Sku) error {
	db := e.Orm.Model(&models.Sku{}).First(model, c.GetId())
	err := db.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("查看对象不存在或无权查看")
	}
	if err != nil {
		e.Log.Errorf("Sku Get error: %s", err)
		return err
	}
	return nil
}

// Insert 新增 SKU。校验 spu_id 真实存在。
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

// Update 修改 SKU。SpuId 变更时校验目标 SPU 存在。
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

// Remove 批量删除。
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
