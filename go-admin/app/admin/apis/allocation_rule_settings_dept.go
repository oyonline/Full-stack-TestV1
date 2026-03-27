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

type AllocationRuleSettingsDept struct {
	api.Api
}

// GetPage 获取分摊规则设置-费用承担部门映射列表
// @Summary 获取分摊规则设置-费用承担部门映射列表
// @Description 获取分摊规则设置-费用承担部门映射列表
// @Tags 分摊规则设置-费用承担部门映射
// @Param allocationType query int false "分摊类型类型(1=固定比例|2=按销售额分摊)"
// @Param allocationRuleSettingsId query int64 false "分摊规则设置ID"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.AllocationRuleSettingsDept}} "{"code": 200, "data": [...]}"
// @Router /api/v1/allocation-rule-settings-dept/list [get]
// @Security Bearer
func (e AllocationRuleSettingsDept) GetPage(c *gin.Context) {
	req := dto.AllocationRuleSettingsDeptGetPageReq{}
	s := service.AllocationRuleSettingsDept{}
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
	list := make([]models.AllocationRuleSettingsDept, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取分摊规则设置-费用承担部门映射失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取分摊规则设置-费用承担部门映射
// @Summary 获取分摊规则设置-费用承担部门映射
// @Description 获取分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.AllocationRuleSettingsDept} "{"code": 200, "data": [...]}"
// @Router /api/v1/allocation-rule-settings-dept/get/{id} [get]
// @Security Bearer
func (e AllocationRuleSettingsDept) Get(c *gin.Context) {
	req := dto.AllocationRuleSettingsDeptGetReq{}
	s := service.AllocationRuleSettingsDept{}
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
	var object models.AllocationRuleSettingsDept

	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取分摊规则设置-费用承担部门映射失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建分摊规则设置-费用承担部门映射
// @Summary 创建分摊规则设置-费用承担部门映射
// @Description 创建分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Accept application/json
// @Product application/json
// @Param data body dto.AllocationRuleSettingsDeptInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/allocation-rule-settings-dept/add [post]
// @Security Bearer
func (e AllocationRuleSettingsDept) Insert(c *gin.Context) {
	req := dto.AllocationRuleSettingsDeptInsertReq{}
	s := service.AllocationRuleSettingsDept{}
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
		e.Error(500, err, fmt.Sprintf("创建分摊规则设置-费用承担部门映射失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改分摊规则设置-费用承担部门映射
// @Summary 修改分摊规则设置-费用承担部门映射
// @Description 修改分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.AllocationRuleSettingsDeptUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/allocation-rule-settings-dept/edit/{id} [put]
// @Security Bearer
func (e AllocationRuleSettingsDept) Update(c *gin.Context) {
	req := dto.AllocationRuleSettingsDeptUpdateReq{}
	s := service.AllocationRuleSettingsDept{}
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
		e.Error(500, err, fmt.Sprintf("修改分摊规则设置-费用承担部门映射失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除分摊规则设置-费用承担部门映射
// @Summary 删除分摊规则设置-费用承担部门映射
// @Description 删除分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Param data body dto.AllocationRuleSettingsDeptDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/allocation-rule-settings-dept/remove [delete]
// @Security Bearer
func (e AllocationRuleSettingsDept) Delete(c *gin.Context) {
	s := service.AllocationRuleSettingsDept{}
	req := dto.AllocationRuleSettingsDeptDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除分摊规则设置-费用承担部门映射失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入分摊规则设置-费用承担部门映射
// @Summary 导入分摊规则设置-费用承担部门映射
// @Description 导入分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/allocation-rule-settings-dept/import [post]
// @Security Bearer
func (e AllocationRuleSettingsDept) Import(c *gin.Context) {
	var importReq dto.AllocationRuleSettingsDeptImportReq
	s := service.AllocationRuleSettingsDept{}
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
		e.Error(500, err, fmt.Sprintf("导入分摊规则设置-费用承担部门映射失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出分摊规则设置-费用承担部门映射
// @Summary 导出分摊规则设置-费用承担部门映射
// @Description 导出分摊规则设置-费用承担部门映射
// @Tags 分摊规则设置-费用承担部门映射
// @Param allocationType query int false "分摊类型类型(1=固定比例|2=按销售额分摊)"
// @Param allocationRuleSettingsId query int64 false "分摊规则设置ID"
// @Success 200 {object} response.Response
// @Router /api/v1/allocation-rule-settings-dept/export [get]
// @Security Bearer
func (e AllocationRuleSettingsDept) Export(c *gin.Context) {
	req := dto.AllocationRuleSettingsDeptGetPageReq{}
	s := service.AllocationRuleSettingsDept{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出分摊规则设置-费用承担部门映射失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载分摊规则设置-费用承担部门映射模板
// @Summary 下载分摊规则设置-费用承担部门映射模板
// @Description 下载分摊规则设置-费用承担部门映射模板
// @Tags 分摊规则设置-费用承担部门映射
// @Success 200 {object} response.Response
// @Router /api/v1/allocation-rule-settings-dept/template [get]
// @Security Bearer
func (e AllocationRuleSettingsDept) DownloadTemplate(c *gin.Context) {
	s := service.AllocationRuleSettingsDept{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载分摊规则设置-费用承担部门映射失败信息 %s", err.Error()))
		return
	}
}
