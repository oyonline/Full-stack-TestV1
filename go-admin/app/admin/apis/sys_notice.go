package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	
	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
)

type SysNotice struct {
	api.Api
}

// GetPage 公告管理列表
// @Summary 公告管理列表
// @Description 公告管理列表
// @Tags 公告管理
// @Param title query string false "标题"
// @Param type query string false "类型"
// @Param status query string false "状态"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.SysNotice}} "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-notice [get]
// @Security Bearer
func (e SysNotice) GetPage(c *gin.Context) {
	s := service.SysNotice{}
	req := dto.SysNoticeGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		return
	}
	list := make([]models.SysNotice, 0)
	var count int64
	err = s.GetPage(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 公告管理详情
// @Summary 公告管理详情
// @Description 公告管理详情
// @Tags 公告管理
// @Param id path string false "id"
// @Success 200 {object} response.Response{data=models.SysNotice} "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-notice/{id} [get]
// @Security Bearer
func (e SysNotice) Get(c *gin.Context) {
	req := dto.SysNoticeGetReq{}
	s := service.SysNotice{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var object models.SysNotice
	err = s.Get(&req, &object)
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	e.OK(object, "查询成功")
}

// Insert 创建公告管理
// @Summary 创建公告管理
// @Description 创建公告管理
// @Tags 公告管理
// @Accept application/json
// @Product application/json
// @Param data body dto.SysNoticeControl true "body"
// @Success 200 {object} response.Response "{"code": 200, "message": "创建成功"}"
// @Router /api/v1/sys-notice [post]
// @Security Bearer
func (e SysNotice) Insert(c *gin.Context) {
	s := service.SysNotice{}
	req := dto.SysNoticeControl{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	err = s.Insert(&req)
	if err != nil {
		e.Error(500, err, "创建失败")
		return
	}
	e.OK(req.GetId(), "创建成功")
}

// Update 修改公告管理
// @Summary 修改公告管理
// @Description 修改公告管理
// @Tags 公告管理
// @Accept application/json
// @Product application/json
// @Param data body dto.SysNoticeControl true "body"
// @Success 200 {object} response.Response "{"code": 200, "message": "修改成功"}"
// @Router /api/v1/sys-notice/{id} [put]
// @Security Bearer
func (e SysNotice) Update(c *gin.Context) {
	s := service.SysNotice{}
	req := dto.SysNoticeControl{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	err = s.Update(&req)
	if err != nil {
		e.Error(500, err, "更新失败")
		return
	}
	e.OK(req.GetId(), "更新成功")
}

// Delete 删除公告管理
// @Summary 删除公告管理
// @Description 删除公告管理
// @Tags 公告管理
// @Param ids body []int false "ids"
// @Success 200 {object} response.Response "{"code": 200, "message": "删除成功"}"
// @Router /api/v1/sys-notice [delete]
// @Security Bearer
func (e SysNotice) Delete(c *gin.Context) {
	s := service.SysNotice{}
	req := dto.SysNoticeDeleteReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	err = s.Remove(&req)
	if err != nil {
		e.Error(500, err, "删除失败")
		return
	}
	e.OK(req.GetId(), "删除成功")
}
