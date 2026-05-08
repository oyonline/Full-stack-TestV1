package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/middleware"
)

type SkuCategory struct {
	api.Api
}

// GetTree 类目树
func (e SkuCategory) GetTree(c *gin.Context) {
	s := service.SkuCategory{}
	req := dto.SkuCategoryPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	tree := make([]*dto.SkuCategoryTreeNode, 0)
	if err := s.GetTree(&req, &tree); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(tree, "查询成功")
}

// Insert 创建类目
func (e SkuCategory) Insert(c *gin.Context) {
	s := service.SkuCategory{}
	req := dto.SkuCategoryInsertReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	if err := s.Insert(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("类目创建失败：%s", err.Error()))
		return
	}
	middleware.AuditLogCreate(c,
		"类目管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySkuCategory,
			ID:    req.CategoryId,
			Label: req.CategoryName,
		},
		map[string]interface{}{
			"categoryName": req.CategoryName,
			"parentId":     req.ParentId,
			"status":       req.Status,
		},
		"admin.skuCategory.insert",
	)
	e.OK(req.CategoryId, "创建成功")
}

// Update 修改类目
func (e SkuCategory) Update(c *gin.Context) {
	s := service.SkuCategory{}
	req := dto.SkuCategoryUpdateReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Update(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("类目更新失败：%s", err.Error()))
		return
	}
	middleware.AuditLogUpdate(c,
		"类目管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySkuCategory,
			ID:    req.CategoryId,
			Label: req.CategoryName,
		},
		nil,
		map[string]interface{}{
			"categoryName": req.CategoryName,
			"parentId":     req.ParentId,
			"status":       req.Status,
		},
		"admin.skuCategory.update",
	)
	e.OK(req.CategoryId, "更新成功")
}

// Delete 批量删除类目
func (e SkuCategory) Delete(c *gin.Context) {
	s := service.SkuCategory{}
	req := dto.SkuCategoryDeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Remove(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("类目删除失败：%s", err.Error()))
		return
	}
	middleware.AuditLogDelete(c,
		"类目管理",
		middleware.AuditTarget{Type: middleware.AuditCategorySkuCategory, ID: req.Ids},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.skuCategory.delete",
	)
	e.OK(req.Ids, "删除成功")
}
