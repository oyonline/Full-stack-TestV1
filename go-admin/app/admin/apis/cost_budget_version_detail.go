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

type CostBudgetVersionDetail struct {
	api.Api
}

// GetPage 获取预算版本管理详情列表
// @Summary 获取预算版本管理详情列表
// @Description 获取预算版本管理详情列表
// @Tags 预算版本管理详情
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.CostBudgetVersionDetail}} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-budget-version-detail/list [get]
// @Security Bearer
func (e CostBudgetVersionDetail) GetPage(c *gin.Context) {
	req := dto.CostBudgetVersionDetailGetPageReq{}
	s := service.CostBudgetVersionDetail{}
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
	list := make([]models.CostBudgetVersionDetail, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算版本管理详情失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取预算版本管理详情
// @Summary 获取预算版本管理详情
// @Description 获取预算版本管理详情
// @Tags 预算版本管理详情
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.CostBudgetVersionDetail} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-budget-version-detail/get/{id} [get]
// @Security Bearer
func (e CostBudgetVersionDetail) Get(c *gin.Context) {
	req := dto.CostBudgetVersionDetailGetReq{}
	s := service.CostBudgetVersionDetail{}
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
	var object models.CostBudgetVersionDetail

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算版本管理详情失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建预算版本管理详情
// @Summary 创建预算版本管理详情
// @Description 创建预算版本管理详情
// @Tags 预算版本管理详情
// @Accept application/json
// @Product application/json
// @Param data body dto.CostBudgetVersionDetailInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/cost-budget-version-detail/add [post]
// @Security Bearer
func (e CostBudgetVersionDetail) Insert(c *gin.Context) {
	req := dto.CostBudgetVersionDetailInsertReq{}
	s := service.CostBudgetVersionDetail{}
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
		e.Error(500, err, fmt.Sprintf("创建预算版本管理详情失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改预算版本管理详情
// @Summary 修改预算版本管理详情
// @Description 修改预算版本管理详情
// @Tags 预算版本管理详情
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.CostBudgetVersionDetailUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/cost-budget-version-detail/edit/{id} [put]
// @Security Bearer
func (e CostBudgetVersionDetail) Update(c *gin.Context) {
	req := dto.CostBudgetVersionDetailUpdateReq{}
	s := service.CostBudgetVersionDetail{}
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
		e.Error(500, err, fmt.Sprintf("修改预算版本管理详情失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除预算版本管理详情
// @Summary 删除预算版本管理详情
// @Description 删除预算版本管理详情
// @Tags 预算版本管理详情
// @Param data body dto.CostBudgetVersionDetailDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/cost-budget-version-detail/remove [delete]
// @Security Bearer
func (e CostBudgetVersionDetail) Delete(c *gin.Context) {
	s := service.CostBudgetVersionDetail{}
	req := dto.CostBudgetVersionDetailDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除预算版本管理详情失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入预算版本管理详情
// @Summary 导入预算版本管理详情
// @Description 导入预算版本管理详情
// @Tags 预算版本管理详情
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/cost-budget-version-detail/import [post]
// @Security Bearer
func (e CostBudgetVersionDetail) Import(c *gin.Context) {
	var importReq dto.CostBudgetVersionDetailImportReq
	s := service.CostBudgetVersionDetail{}
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
		e.Error(500, err, fmt.Sprintf("导入预算版本管理详情失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出预算版本管理详情
// @Summary 导出预算版本管理详情
// @Description 导出预算版本管理详情
// @Tags 预算版本管理详情
// @Success 200 {object} response.Response
// @Router /api/v1/cost-budget-version-detail/export [get]
// @Security Bearer
func (e CostBudgetVersionDetail) Export(c *gin.Context) {
	req := dto.CostBudgetVersionDetailGetPageReq{}
	s := service.CostBudgetVersionDetail{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出预算版本管理详情失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载预算版本管理详情模板
// @Summary 下载预算版本管理详情模板
// @Description 下载预算版本管理详情模板
// @Tags 预算版本管理详情
// @Success 200 {object} response.Response
// @Router /api/v1/cost-budget-version-detail/template [get]
// @Security Bearer
func (e CostBudgetVersionDetail) DownloadTemplate(c *gin.Context) {
	s := service.CostBudgetVersionDetail{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载预算版本管理详情失败信息 %s", err.Error()))
		return
	}
}
