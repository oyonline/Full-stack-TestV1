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

type BudgetFeeCategoryDetails struct {
	api.Api
}

// GetPage 获取预算费用明细列表
// @Summary 获取预算费用明细列表
// @Description 获取预算费用明细列表
// @Tags 预算费用明细
// @Param feeName query string false "feeName"
// @Param feeCode query string false "feeCode"
// @Param budgetFeeCategoryId query int false "budgetFeeCategoryId"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.BudgetFeeCategoryDetails}} "{"code": 200, "data": [...]}"
// @Router /api/v1/budget-fee-category-details/list [get]
// @Security Bearer
func (e BudgetFeeCategoryDetails) GetPage(c *gin.Context) {
	req := dto.BudgetFeeCategoryDetailsGetPageReq{}
	s := service.BudgetFeeCategoryDetails{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	list := make([]models.BudgetFeeCategoryDetails, 0)
	var count int64
	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取预算费用明细
// @Summary 获取预算费用明细
// @Description 获取预算费用明细
// @Tags 预算费用明细
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.BudgetFeeCategoryDetails} "{"code": 200, "data": [...]}"
// @Router /api/v1/budget-fee-category-details/get/{id} [get]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Get(c *gin.Context) {
	req := dto.BudgetFeeCategoryDetailsGetReq{}
	s := service.BudgetFeeCategoryDetails{}
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
	var object models.BudgetFeeCategoryDetails

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建预算费用明细
// @Summary 创建预算费用明细
// @Description 创建预算费用明细
// @Tags 预算费用明细
// @Accept application/json
// @Product application/json
// @Param data body dto.BudgetFeeCategoryDetailsInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/budget-fee-category-details/add [post]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Insert(c *gin.Context) {
	req := dto.BudgetFeeCategoryDetailsInsertReq{}
	s := service.BudgetFeeCategoryDetails{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	// 设置创建人
	req.SetCreateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	err = s.Insert(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("创建预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改预算费用明细
// @Summary 修改预算费用明细
// @Description 修改预算费用明细
// @Tags 预算费用明细
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.BudgetFeeCategoryDetailsUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/budget-fee-category-details/edit/{id} [put]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Update(c *gin.Context) {
	req := dto.BudgetFeeCategoryDetailsUpdateReq{}
	s := service.BudgetFeeCategoryDetails{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	err = s.Update(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("修改预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除预算费用明细
// @Summary 删除预算费用明细
// @Description 删除预算费用明细
// @Tags 预算费用明细
// @Param data body dto.BudgetFeeCategoryDetailsDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/budget-fee-category-details/remove [delete]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Delete(c *gin.Context) {
	s := service.BudgetFeeCategoryDetails{}
	req := dto.BudgetFeeCategoryDetailsDeleteReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	// req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	err = s.Remove(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("删除预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入预算费用明细
// @Summary 导入预算费用明细
// @Description 导入预算费用明细
// @Tags 预算费用明细
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param budgetFeeCategoryId formData int64 true "费用类别ID"
// @Param viewType formData int true "视图类型(1:金蝶科目视图 2:经营管理视图 3:项目管理视图)"
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/budget-fee-category-details/import [post]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Import(c *gin.Context) {
	var importReq dto.BudgetFeeCategoryDetailsImportReq
	s := service.BudgetFeeCategoryDetails{}
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
		e.Error(500, err, fmt.Sprintf("导入失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出预算费用明细
// @Summary 导出预算费用明细
// @Description 导出预算费用明细
// @Tags 预算费用明细
// @Param feeName query string false "feeName"
// @Param feeCode query string false "feeCode"
// @Param budgetFeeCategoryId query int false "budgetFeeCategoryId"
// @Success 200 {object} response.Response
// @Router /api/v1/budget-fee-category-details/export [get]
// @Security Bearer
func (e BudgetFeeCategoryDetails) Export(c *gin.Context) {
	req := dto.BudgetFeeCategoryDetailsGetPageReq{}
	s := service.BudgetFeeCategoryDetails{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载预算费用明细模板
// @Summary 下载预算费用明细模板
// @Description 下载预算费用明细模板
// @Tags 预算费用明细
// @Success 200 {object} response.Response
// @Router /api/v1/budget-fee-category-details/template [get]
// @Security Bearer
func (e BudgetFeeCategoryDetails) DownloadTemplate(c *gin.Context) {
	s := service.BudgetFeeCategoryDetails{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用明细失败，\r\n失败信息 %s", err.Error()))
		return
	}
}
