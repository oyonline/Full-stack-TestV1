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

type CostBudgetVersion struct {
	api.Api
}

// GetPage 获取预算版本管理列表
// @Summary 获取预算版本管理列表
// @Description 获取预算版本管理列表
// @Tags 预算版本管理
// @Param costBudgetName query string false "版本名称"
// @Param costBudgetCode query string false "版本编码"
// @Param years query int false "预算年度"
// @Param status query int false "状态(1=草稿|2=生效中)"
// @Param costCenterInfoId query int64 false "成本中心ID"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.CostBudgetVersion}} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-budget-version/list [get]
// @Security Bearer
func (e CostBudgetVersion) GetPage(c *gin.Context) {
	req := dto.CostBudgetVersionGetPageReq{}
	s := service.CostBudgetVersion{}
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
	list := make([]models.CostBudgetVersion, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算版本管理失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取预算版本管理
// @Summary 获取预算版本管理
// @Description 获取预算版本管理
// @Tags 预算版本管理
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.CostBudgetVersion} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-budget-version/get/{id} [get]
// @Security Bearer
func (e CostBudgetVersion) Get(c *gin.Context) {
	req := dto.CostBudgetVersionGetReq{}
	s := service.CostBudgetVersion{}
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
	groupResults, err := s.Get(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算版本管理失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(groupResults, "查询成功")
}

// Insert 创建预算版本管理
// @Summary 创建预算版本管理
// @Description 创建预算版本管理
// @Tags 预算版本管理
// @Accept application/json
// @Product application/json
// @Param data body dto.CostBudgetVersionInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/cost-budget-version/add [post]
// @Security Bearer
func (e CostBudgetVersion) Insert(c *gin.Context) {
	req := dto.CostBudgetVersionInsertReq{}
	s := service.CostBudgetVersion{}
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
		e.Error(500, err, fmt.Sprintf("创建预算版本管理失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改预算版本管理
// @Summary 修改预算版本管理
// @Description 修改预算版本管理
// @Tags 预算版本管理
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.CostBudgetVersionUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/cost-budget-version/edit/{id} [put]
// @Security Bearer
func (e CostBudgetVersion) Update(c *gin.Context) {
	req := dto.CostBudgetVersionUpdateReq{}
	s := service.CostBudgetVersion{}
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
		e.Error(500, err, fmt.Sprintf("修改预算版本管理失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除预算版本管理
// @Summary 删除预算版本管理
// @Description 删除预算版本管理
// @Tags 预算版本管理
// @Param data body dto.CostBudgetVersionDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/cost-budget-version/remove [delete]
// @Security Bearer
func (e CostBudgetVersion) Delete(c *gin.Context) {
	s := service.CostBudgetVersion{}
	req := dto.CostBudgetVersionDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除预算版本管理失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入预算版本管理
// @Summary 导入预算版本管理
// @Description 导入预算版本管理
// @Tags 预算版本管理
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/cost-budget-version/import [post]
// @Security Bearer
func (e CostBudgetVersion) Import(c *gin.Context) {
	var importReq dto.CostBudgetVersionImportReq
	s := service.CostBudgetVersion{}
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
		e.Error(500, err, fmt.Sprintf("导入预算版本管理失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出预算版本管理
// @Summary 导出预算版本管理
// @Description 导出预算版本管理
// @Tags 预算版本管理
// @Param costBudgetName query string false "版本名称"
// @Param costBudgetCode query string false "版本编码"
// @Param years query int false "预算年度"
// @Param status query int false "状态(1=草稿|2=生效中)"
// @Param costCenterInfoId query int64 false "成本中心ID"
// @Success 200 {object} response.Response
// @Router /api/v1/cost-budget-version/export [get]
// @Security Bearer
func (e CostBudgetVersion) Export(c *gin.Context) {
	req := dto.CostBudgetVersionGetPageReq{}
	s := service.CostBudgetVersion{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出预算版本管理失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载预算版本管理模板
// @Summary 下载预算版本管理模板
// @Description 下载预算版本管理模板
// @Tags 预算版本管理
// @Success 200 {object} response.Response
// @Router /api/v1/cost-budget-version/template [get]
// @Security Bearer
func (e CostBudgetVersion) DownloadTemplate(c *gin.Context) {
	s := service.CostBudgetVersion{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载预算版本管理失败信息 %s", err.Error()))
		return
	}
}
