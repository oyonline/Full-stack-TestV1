package apis

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

type BudgetFeeCategory struct {
	api.Api
}

// GetPage 获取预算费用类别列表
// @Summary 获取预算费用类别列表
// @Description 获取预算费用类别列表
// @Tags 预算费用类别
// @Param categoryName query string false "categoryName"
// @Param categoryCode query string false "categoryCode"
// @Param viewType query int false "1 金蝶科目视图 2 经营管理视图 3项目管理视图"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.BudgetFeeCategory}} "{"code": 200, "data": [...]}"
// @Router /api/v1/budget-fee-category/list [get]
// @Security Bearer
func (e BudgetFeeCategory) GetPage(c *gin.Context) {
	req := dto.BudgetFeeCategoryGetPageReq{}
	s := service.BudgetFeeCategory{}
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
	list := make([]models.BudgetFeeCategory, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用类别失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取预算费用类别
// @Summary 获取预算费用类别
// @Description 获取预算费用类别
// @Tags 预算费用类别
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.BudgetFeeCategory} "{"code": 200, "data": [...]}"
// @Router /api/v1/budget-fee-category/get/{id} [get]
// @Security Bearer
func (e BudgetFeeCategory) Get(c *gin.Context) {
	req := dto.BudgetFeeCategoryGetReq{}
	s := service.BudgetFeeCategory{}
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
	var object models.BudgetFeeCategory

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取预算费用类别失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建预算费用类别
// @Summary 创建预算费用类别
// @Description 创建预算费用类别
// @Tags 预算费用类别
// @Accept application/json
// @Product application/json
// @Param data body dto.BudgetFeeCategoryInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/budget-fee-category/add [post]
// @Security Bearer
func (e BudgetFeeCategory) Insert(c *gin.Context) {
	req := dto.BudgetFeeCategoryInsertReq{}
	s := service.BudgetFeeCategory{}
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
	p := actions.GetPermissionFromContext(c)
	err = s.Insert(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("创建预算费用类别失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改预算费用类别
// @Summary 修改预算费用类别
// @Description 修改预算费用类别
// @Tags 预算费用类别
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.BudgetFeeCategoryUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/budget-fee-category/edit/{id} [put]
// @Security Bearer
func (e BudgetFeeCategory) Update(c *gin.Context) {
	req := dto.BudgetFeeCategoryUpdateReq{}
	s := service.BudgetFeeCategory{}
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
		e.Error(500, err, fmt.Sprintf("修改预算费用类别失败： %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除预算费用类别
// @Summary 删除预算费用类别
// @Description 删除预算费用类别
// @Tags 预算费用类别
// @Param data body dto.BudgetFeeCategoryDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/budget-fee-category/remove [delete]
// @Security Bearer
func (e BudgetFeeCategory) Delete(c *gin.Context) {
	s := service.BudgetFeeCategory{}
	req := dto.BudgetFeeCategoryDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除预算费用类别失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// ListTree 获取预算费用类别树形列表
// @Summary 获取预算费用类别树形列表
// @Description 获取预算费用类别树形列表
// @Tags 预算费用类别
// @Param categoryName query string false "categoryName"
// @Param categoryCode query string false "categoryCode"
// @Param viewType query int false "1 金蝶科目视图 2 经营管理视图 3项目管理视图"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/budget-fee-category/listTree [get]
// @Security Bearer
func (e BudgetFeeCategory) ListTree(c *gin.Context) {
	s := service.BudgetFeeCategory{}
	req := dto.BudgetFeeCategoryGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	list, err := s.SetBudgetFeeCategoryListTree(&req, p)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(list, "")
}

// Import 导入预算费用类别
// @Summary 导入预算费用类别
// @Description 导入预算费用类别
// @Tags 预算费用类别
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/budget-fee-category/import [post]
// @Security Bearer
func (e BudgetFeeCategory) Import(c *gin.Context) {
	var importReq dto.BudgetFeeCategoryImportReq
	s := service.BudgetFeeCategory{}
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

// Export 导出预算费用类别
// @Summary 导出预算费用类别
// @Description 导出预算费用类别
// @Tags 预算费用类别
// @Param categoryName query string false "categoryName"
// @Param categoryCode query string false "categoryCode"
// @Param viewType query int false "1 金蝶科目视图 2 经营管理视图 3项目管理视图"
// @Success 200 {object} response.Response
// @Router /api/v1/budget-fee-category/export [get]
// @Security Bearer
func (e BudgetFeeCategory) Export(c *gin.Context) {
	req := dto.BudgetFeeCategoryGetPageReq{}
	s := service.BudgetFeeCategory{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出预算费用类别失败，\r\n失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载预算费用类别模板
// @Summary 下载预算费用类别模板
// @Description 下载预算费用类别模板
// @Tags 预算费用类别
// @Success 200 {object} response.Response
// @Router /api/v1/budget-fee-category/template [get]
// @Security Bearer
func (e BudgetFeeCategory) DownloadTemplate(c *gin.Context) {
	s := service.BudgetFeeCategory{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载预算费用类别模板失败，\r\n失败信息 %s", err.Error()))
		return
	}
}
