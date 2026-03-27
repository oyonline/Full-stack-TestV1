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

type CostCenterRelatedCustomer struct {
	api.Api
}

// GetPage 获取成本中心关联客户分组列表
// @Summary 获取成本中心关联客户分组列表
// @Description 获取成本中心关联客户分组列表
// @Tags 成本中心关联客户分组
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.CostCenterRelatedCustomer}} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-related-customer/list [get]
// @Security Bearer
func (e CostCenterRelatedCustomer) GetPage(c *gin.Context) {
	req := dto.CostCenterRelatedCustomerGetPageReq{}
	s := service.CostCenterRelatedCustomer{}
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
	list := make([]models.CostCenterRelatedCustomer, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心关联客户分组失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取成本中心关联客户分组
// @Summary 获取成本中心关联客户分组
// @Description 获取成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.CostCenterRelatedCustomer} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-related-customer/get/{id} [get]
// @Security Bearer
func (e CostCenterRelatedCustomer) Get(c *gin.Context) {
	req := dto.CostCenterRelatedCustomerGetReq{}
	s := service.CostCenterRelatedCustomer{}
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
	var object models.CostCenterRelatedCustomer

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心关联客户分组失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建成本中心关联客户分组
// @Summary 创建成本中心关联客户分组
// @Description 创建成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Accept application/json
// @Product application/json
// @Param data body dto.CostCenterRelatedCustomerInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/cost-center-related-customer/add [post]
// @Security Bearer
func (e CostCenterRelatedCustomer) Insert(c *gin.Context) {
	req := dto.CostCenterRelatedCustomerInsertReq{}
	s := service.CostCenterRelatedCustomer{}
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
		e.Error(500, err, fmt.Sprintf("创建成本中心关联客户分组失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改成本中心关联客户分组
// @Summary 修改成本中心关联客户分组
// @Description 修改成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.CostCenterRelatedCustomerUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/cost-center-related-customer/edit/{id} [put]
// @Security Bearer
func (e CostCenterRelatedCustomer) Update(c *gin.Context) {
	req := dto.CostCenterRelatedCustomerUpdateReq{}
	s := service.CostCenterRelatedCustomer{}
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
		e.Error(500, err, fmt.Sprintf("修改成本中心关联客户分组失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除成本中心关联客户分组
// @Summary 删除成本中心关联客户分组
// @Description 删除成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Param data body dto.CostCenterRelatedCustomerDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/cost-center-related-customer/remove [delete]
// @Security Bearer
func (e CostCenterRelatedCustomer) Delete(c *gin.Context) {
	s := service.CostCenterRelatedCustomer{}
	req := dto.CostCenterRelatedCustomerDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除成本中心关联客户分组失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入成本中心关联客户分组
// @Summary 导入成本中心关联客户分组
// @Description 导入成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/cost-center-related-customer/import [post]
// @Security Bearer
func (e CostCenterRelatedCustomer) Import(c *gin.Context) {
	var importReq dto.CostCenterRelatedCustomerImportReq
	s := service.CostCenterRelatedCustomer{}
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
		e.Error(500, err, fmt.Sprintf("导入成本中心关联客户分组失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出成本中心关联客户分组
// @Summary 导出成本中心关联客户分组
// @Description 导出成本中心关联客户分组
// @Tags 成本中心关联客户分组
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-related-customer/export [get]
// @Security Bearer
func (e CostCenterRelatedCustomer) Export(c *gin.Context) {
	req := dto.CostCenterRelatedCustomerGetPageReq{}
	s := service.CostCenterRelatedCustomer{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出成本中心关联客户分组失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载成本中心关联客户分组模板
// @Summary 下载成本中心关联客户分组模板
// @Description 下载成本中心关联客户分组模板
// @Tags 成本中心关联客户分组
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-related-customer/template [get]
// @Security Bearer
func (e CostCenterRelatedCustomer) DownloadTemplate(c *gin.Context) {
	s := service.CostCenterRelatedCustomer{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载成本中心关联客户分组失败信息 %s", err.Error()))
		return
	}
}
