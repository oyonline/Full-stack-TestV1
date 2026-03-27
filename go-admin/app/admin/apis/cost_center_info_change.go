package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

type CostCenterInfoChange struct {
	api.Api
}

// GetPage 获取成本中心变更数据列表
// @Summary 获取成本中心变更数据列表
// @Description 获取成本中心变更数据列表
// @Tags 成本中心变更数据
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.CostCenterInfoChange}} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-info-change/list [get]
// @Security Bearer
func (e CostCenterInfoChange) GetPage(c *gin.Context) {
	req := dto.CostCenterInfoChangeGetPageReq{}
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	p := actions.GetPermissionFromContext(c)
	list := make([]models.CostCenterInfoChange, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心变更数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取成本中心变更数据
// @Summary 获取成本中心变更数据
// @Description 获取成本中心变更数据
// @Tags 成本中心变更数据
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.CostCenterInfoChange} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-info-change/get/{id} [get]
// @Security Bearer
func (e CostCenterInfoChange) Get(c *gin.Context) {
	req := dto.CostCenterInfoChangeGetReq{}
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var object models.CostCenterInfoChange

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心变更数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建成本中心变更数据
// @Summary 创建成本中心变更数据
// @Description 创建成本中心变更数据
// @Tags 成本中心变更数据
// @Accept application/json
// @Product application/json
// @Param data body dto.CostCenterInfoChangeInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/cost-center-info-change/add [post]
// @Security Bearer
func (e CostCenterInfoChange) Insert(c *gin.Context) {
	req := dto.CostCenterInfoChangeInsertReq{}
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	// 设置创建人
	req.SetCreateBy(user.GetUserId(c))

	err = s.Insert(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("创建成本中心变更数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改成本中心变更数据
// @Summary 修改成本中心变更数据
// @Description 修改成本中心变更数据
// @Tags 成本中心变更数据
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.CostCenterInfoChangeUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/cost-center-info-change/edit/{id} [put]
// @Security Bearer
func (e CostCenterInfoChange) Update(c *gin.Context) {
	req := dto.CostCenterInfoChangeUpdateReq{}
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)

	err = s.Update(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("修改成本中心变更数据失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除成本中心变更数据
// @Summary 删除成本中心变更数据
// @Description 删除成本中心变更数据
// @Tags 成本中心变更数据
// @Param data body dto.CostCenterInfoChangeDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/cost-center-info-change/remove [delete]
// @Security Bearer
func (e CostCenterInfoChange) Delete(c *gin.Context) {
	s := service.CostCenterInfoChange{}
	req := dto.CostCenterInfoChangeDeleteReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	// req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)

	err = s.Remove(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("删除成本中心变更数据失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入成本中心变更数据
// @Summary 导入成本中心变更数据
// @Description 导入成本中心变更数据
// @Tags 成本中心变更数据
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/cost-center-info-change/import [post]
// @Security Bearer
func (e CostCenterInfoChange) Import(c *gin.Context) {
	var importReq dto.CostCenterInfoChangeImportReq
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).MakeOrm().Bind(&importReq).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	file, _ := c.FormFile("file")
	if file == nil {
		e.Error(500, nil, "请选择文件")
		return
	}
	importReq.File = file
	err = s.ImportData(&importReq, p, user.GetUserId(c))
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导入成本中心变更数据失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出成本中心变更数据
// @Summary 导出成本中心变更数据
// @Description 导出成本中心变更数据
// @Tags 成本中心变更数据
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-info-change/export [get]
// @Security Bearer
func (e CostCenterInfoChange) Export(c *gin.Context) {
	req := dto.CostCenterInfoChangeGetPageReq{}
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出成本中心变更数据失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载成本中心变更数据模板
// @Summary 下载成本中心变更数据模板
// @Description 下载成本中心变更数据模板
// @Tags 成本中心变更数据
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-info-change/template [get]
// @Security Bearer
func (e CostCenterInfoChange) DownloadTemplate(c *gin.Context) {
	s := service.CostCenterInfoChange{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载成本中心变更数据失败信息 %s", err.Error()))
		return
	}
}
