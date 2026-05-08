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

type SkuBrand struct {
	api.Api
}

// GetPage 品牌列表
func (e SkuBrand) GetPage(c *gin.Context) {
	s := service.SkuBrand{}
	req := dto.SkuBrandPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	list := make([]models.SkuBrand, 0)
	var count int64
	if err := s.GetPage(&req, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 品牌详情
func (e SkuBrand) Get(c *gin.Context) {
	s := service.SkuBrand{}
	req := dto.SkuBrandGetReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var item models.SkuBrand
	if err := s.Get(&req, &item); err != nil {
		e.Error(500, err, fmt.Sprintf("品牌获取失败：%s", err.Error()))
		return
	}
	e.OK(item, "查询成功")
}

// Insert 创建品牌
func (e SkuBrand) Insert(c *gin.Context) {
	s := service.SkuBrand{}
	req := dto.SkuBrandInsertReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	if err := s.Insert(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("品牌创建失败：%s", err.Error()))
		return
	}
	middleware.AuditLogCreate(c,
		"品牌管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySkuBrand,
			ID:    req.BrandId,
			Label: req.BrandName,
		},
		map[string]interface{}{"brandName": req.BrandName, "status": req.Status},
		"admin.skuBrand.insert",
	)
	e.OK(req.BrandId, "创建成功")
}

// Update 修改品牌
func (e SkuBrand) Update(c *gin.Context) {
	s := service.SkuBrand{}
	req := dto.SkuBrandUpdateReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Update(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("品牌更新失败：%s", err.Error()))
		return
	}
	middleware.AuditLogUpdate(c,
		"品牌管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySkuBrand,
			ID:    req.BrandId,
			Label: req.BrandName,
		},
		nil,
		map[string]interface{}{"brandName": req.BrandName, "status": req.Status},
		"admin.skuBrand.update",
	)
	e.OK(req.BrandId, "更新成功")
}

// Delete 批量删除品牌
func (e SkuBrand) Delete(c *gin.Context) {
	s := service.SkuBrand{}
	req := dto.SkuBrandDeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Remove(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("品牌删除失败：%s", err.Error()))
		return
	}
	middleware.AuditLogDelete(c,
		"品牌管理",
		middleware.AuditTarget{Type: middleware.AuditCategorySkuBrand, ID: req.Ids},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.skuBrand.delete",
	)
	e.OK(req.Ids, "删除成功")
}
