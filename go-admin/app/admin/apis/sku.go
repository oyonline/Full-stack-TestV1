package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/middleware"
)

type Sku struct {
	api.Api
}

// GetPage SKU 列表
func (e Sku) GetPage(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]models.Sku, 0)
	var count int64
	if err := s.GetPage(&req, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get SKU 详情
func (e Sku) Get(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuGetReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var item models.Sku
	if err := s.Get(&req, &item); err != nil {
		e.Error(500, err, fmt.Sprintf("SKU 获取失败：%s", err.Error()))
		return
	}
	e.OK(item, "查询成功")
}

// Insert 新增 SKU
func (e Sku) Insert(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuInsertReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	if err := s.Insert(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("SKU 创建失败：%s", err.Error()))
		return
	}
	middleware.AuditLogCreate(c,
		"SKU 管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySku,
			ID:    req.SkuId,
			Label: req.SkuName,
		},
		map[string]interface{}{
			"spuId":   req.SpuId,
			"skuCode": req.SkuCode,
			"skuName": req.SkuName,
			"price":   req.Price,
			"status":  req.Status,
		},
		"admin.sku.insert",
	)
	e.OK(req.SkuId, "创建成功")
}

// Update 修改 SKU
func (e Sku) Update(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuUpdateReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Update(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("SKU 更新失败：%s", err.Error()))
		return
	}
	middleware.AuditLogUpdate(c,
		"SKU 管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySku,
			ID:    req.SkuId,
			Label: req.SkuName,
		},
		nil,
		map[string]interface{}{
			"spuId":   req.SpuId,
			"skuCode": req.SkuCode,
			"skuName": req.SkuName,
			"price":   req.Price,
			"status":  req.Status,
		},
		"admin.sku.update",
	)
	e.OK(req.SkuId, "更新成功")
}

// Delete 批量删除 SKU
func (e Sku) Delete(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuDeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Remove(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("SKU 删除失败：%s", err.Error()))
		return
	}
	middleware.AuditLogDelete(c,
		"SKU 管理",
		middleware.AuditTarget{Type: middleware.AuditCategorySku, ID: req.Ids},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.sku.delete",
	)
	e.OK(req.Ids, "删除成功")
}
