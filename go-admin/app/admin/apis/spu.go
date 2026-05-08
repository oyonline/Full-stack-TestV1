package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

type Spu struct {
	api.Api
}

// GetPage SPU 列表
func (e Spu) GetPage(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]dto.SpuListItem, 0)
	var count int64
	p := actions.GetPermissionFromContext(c)
	if err := s.GetPage(&req, p, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get SPU 详情
func (e Spu) Get(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuGetReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var item dto.SpuListItem
	p := actions.GetPermissionFromContext(c)
	if err := s.Get(&req, p, &item); err != nil {
		e.Error(500, err, fmt.Sprintf("SPU 获取失败：%s", err.Error()))
		return
	}
	e.OK(item, "查询成功")
}

// Insert 新增 SPU
func (e Spu) Insert(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuInsertReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	id, err := s.Insert(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SPU 创建失败：%s", err.Error()))
		return
	}
	middleware.AuditLogCreate(c,
		"SPU 管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySpu,
			ID:    id,
			Label: req.SpuName,
		},
		map[string]interface{}{
			"spuCode":    req.SpuCode,
			"spuName":    req.SpuName,
			"categoryId": req.CategoryId,
			"brandId":    req.BrandId,
			"status":     req.Status,
		},
		"admin.spu.insert",
	)
	e.OK(id, "创建成功")
}

// Update 修改 SPU
func (e Spu) Update(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuUpdateReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err := s.Update(&req, p); err != nil {
		e.Error(500, err, fmt.Sprintf("SPU 更新失败：%s", err.Error()))
		return
	}
	middleware.AuditLogUpdate(c,
		"SPU 管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategorySpu,
			ID:    req.SpuId,
			Label: req.SpuName,
		},
		nil,
		map[string]interface{}{
			"spuCode":    req.SpuCode,
			"spuName":    req.SpuName,
			"categoryId": req.CategoryId,
			"brandId":    req.BrandId,
			"status":     req.Status,
		},
		"admin.spu.update",
	)
	e.OK(req.SpuId, "更新成功")
}

// Delete 批量删除 SPU
func (e Spu) Delete(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuDeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err := s.Remove(&req, p); err != nil {
		e.Error(500, err, fmt.Sprintf("SPU 删除失败：%s", err.Error()))
		return
	}
	middleware.AuditLogDelete(c,
		"SPU 管理",
		middleware.AuditTarget{Type: middleware.AuditCategorySpu, ID: req.Ids},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.spu.delete",
	)
	e.OK(req.Ids, "删除成功")
}

// Submit SPU 提交审核
func (e Spu) Submit(c *gin.Context) {
	s := service.Spu{}
	req := dto.SpuSubmitReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	instanceId, err := s.SubmitForReview(c, p, &req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SPU 提交审核失败：%s", err.Error()))
		return
	}
	middleware.AuditLog(c, middleware.AuditEntry{
		Title:  "SPU 管理",
		Action: middleware.AuditActionStart,
		Target: middleware.AuditTarget{
			Type: middleware.AuditCategorySpu,
			ID:   req.SpuId,
		},
		After: map[string]interface{}{
			"action":     "submit-for-review",
			"instanceId": instanceId,
			"remark":     req.Remark,
		},
		Method: "admin.spu.submit",
	})
	e.OK(map[string]interface{}{"spuId": req.SpuId, "instanceId": instanceId}, "已提交审核")
}
