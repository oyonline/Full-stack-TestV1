package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/platform/models"
	"go-admin/app/platform/service"
	"go-admin/app/platform/service/dto"
	"go-admin/common/middleware"
)

type ModuleRegistry struct {
	api.Api
}

func (e ModuleRegistry) GetPage(c *gin.Context) {
	s := service.ModuleRegistry{}
	req := dto.ModuleRegistryGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]models.ModuleRegistry, 0)
	var count int64
	if err = s.GetPage(&req, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e ModuleRegistry) Get(c *gin.Context) {
	s := service.ModuleRegistry{}
	req := dto.ModuleRegistryGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var data models.ModuleRegistry
	if err = s.Get(req.Id, &data); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(data, "查询成功")
}

func (e ModuleRegistry) Insert(c *gin.Context) {
	s := service.ModuleRegistry{}
	req := dto.ModuleRegistryInsertReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.CreateBy = user.GetUserId(c)
	if err = s.Insert(&req); err != nil {
		e.Error(500, err, "创建失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "模块注册",
		BusinessType:  middleware.AuditActionCreate,
		BusinessTypes: middleware.AuditCategoryModule,
		Method:        "platform.module.insert",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("模块编码", req.ModuleKey),
			middleware.AuditKV("模块名称", req.ModuleName),
			middleware.AuditKV("路由前缀", req.RouteBase),
			middleware.AuditKV("根菜单编码", req.MenuRootCode),
		),
	})
	e.OK(nil, "创建成功")
}

func (e ModuleRegistry) Update(c *gin.Context) {
	s := service.ModuleRegistry{}
	req := dto.ModuleRegistryUpdateReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.UpdateBy = user.GetUserId(c)
	if err = s.Update(&req); err != nil {
		e.Error(500, err, "更新失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "模块注册",
		BusinessType:  middleware.AuditActionUpdate,
		BusinessTypes: middleware.AuditCategoryModule,
		Method:        "platform.module.update",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("模块ID", req.ModuleId),
			middleware.AuditKV("模块编码", req.ModuleKey),
			middleware.AuditKV("模块名称", req.ModuleName),
			middleware.AuditKV("路由前缀", req.RouteBase),
			middleware.AuditKV("根菜单编码", req.MenuRootCode),
		),
	})
	e.OK(nil, "更新成功")
}

func (e ModuleRegistry) Delete(c *gin.Context) {
	s := service.ModuleRegistry{}
	req := dto.ModuleRegistryDeleteReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	if err = s.Delete(req.Id); err != nil {
		e.Error(500, err, fmt.Sprintf("删除失败,%s", err.Error()))
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "模块注册",
		BusinessType:  middleware.AuditActionDelete,
		BusinessTypes: middleware.AuditCategoryModule,
		Method:        "platform.module.delete",
		OperatorType:  middleware.AuditOperatorManage,
		Remark:        middleware.AuditKV("模块ID", req.Id),
	})
	e.OK(nil, "删除成功")
}
