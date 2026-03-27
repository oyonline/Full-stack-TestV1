package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
)

type KingdeeCustomer struct {
	api.Api
}

// GetPage
// @Summary 金蝶店铺列表数据
// @Description 获取JSON
// @Tags 金蝶店铺
// @Param customerName query string false "customerName"
// @Param customerNumber query string false "customerNumber"
// @Param forbidStatus query string false "forbidStatus"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer [get]
// @Security Bearer
func (e KingdeeCustomer) GetPage(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	list := make([]models.KingdeeCustomer, 0)
	var count int64

	err = s.GetPage(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get
// @Summary 获取金蝶店铺信息
// @Description 获取JSON
// @Tags 金蝶店铺
// @Param id path int true "编码"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/{customerId} [get]
// @Security Bearer
func (e KingdeeCustomer) Get(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var object models.KingdeeCustomer

	err = s.Get(&req, &object)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("金蝶店铺信息获取失败！错误详情：%s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert
// @Summary 添加金蝶店铺
// @Description 获取JSON
// @Tags 金蝶店铺
// @Accept  application/json
// @Product application/json
// @Param data body dto.KingdeeCustomerInsertReq true "data"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer [post]
// @Security Bearer
func (e KingdeeCustomer) Insert(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerInsertReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	err = s.Insert(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("新建金蝶店铺失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.CustomerNumber, "创建成功")
}

// Update
// @Summary 修改金蝶店铺
// @Description 获取JSON
// @Tags 金蝶店铺
// @Accept  application/json
// @Product application/json
// @Param data body dto.KingdeeCustomerUpdateReq true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/{id} [put]
// @Security Bearer
func (e KingdeeCustomer) Update(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerUpdateReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	req.SetUpdateBy(user.GetUserId(c))

	err = s.Update(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("金蝶店铺更新失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetId(), "更新成功")
}

// Delete
// @Summary 删除金蝶店铺
// @Description 删除数据
// @Tags 金蝶店铺
// @Param id body dto.KingdeeCustomerDeleteReq true "请求参数"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer [delete]
// @Security Bearer
func (e KingdeeCustomer) Delete(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerDeleteReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	err = s.Remove(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("金蝶店铺删除失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// DownloadTemplate
// @Summary 下载金蝶店铺模板
// @Description 下载金蝶店铺模板
// @Tags 金蝶店铺
// @Success 200 {object} response.Response
// @Router /api/v1/kingdee-customer/template [get]
// @Security Bearer
func (e KingdeeCustomer) DownloadTemplate(c *gin.Context) {
	s := service.KingdeeCustomer{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取金蝶店铺失败，\r\n失败信息 %s", err.Error()))
		return
	}
}

// Import
// @Summary 导入金蝶店铺数据
// @Description 导入金蝶店铺信息
// @Tags 金蝶店铺
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/kingdee-customer/import [post]
// @Security Bearer
func (e KingdeeCustomer) Import(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerImport{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	file, _ := c.FormFile("file")
	if file == nil {
		e.Error(500, nil, "请选择文件")
		return
	}
	req.File = file
	err = s.ImportData(&req, user.GetUserId(c))
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导入失败信息 %s", err.Error()))
		return
	}
	e.OK(req, "导入成功")
}

// Export
// @Summary 导出金蝶店铺数据
// @Description 导出金蝶店铺信息
// @Tags 金蝶店铺
// @Param customerName query string false "customerName"
// @Param customerNumber query string false "customerNumber"
// @Param forbidStatus query string false "forbidStatus"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/export [get]
// @Security Bearer
func (e KingdeeCustomer) Export(c *gin.Context) {
	s := service.KingdeeCustomer{}
	req := dto.KingdeeCustomerPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	err = s.Export(c.Writer, &req)
	if err != nil {
		e.Error(500, err, "导出失败")
		return
	}
}

// PullKingdeeCustomers
// @Summary 拉取金蝶店铺信息
// @Description 拉取金蝶店铺列表
// @Tags 金蝶店铺
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/pull [get]
// @Security Bearer
func (e KingdeeCustomer) PullKingdeeCustomers(c *gin.Context) {
	s := service.KingdeeCustomer{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	createBy := user.GetUserId(c)
	err = s.PullKingdeeCustomers(&createBy)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("拉取金蝶店铺失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(createBy, "创建成功")
}

// PullKingdeeCustomerGroups
// @Summary 拉取金蝶客户分组信息
// @Description 拉取金蝶客户分组列表
// @Tags 金蝶店铺
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/group [post]
// @Security Bearer
func (e KingdeeCustomer) PullKingdeeCustomerGroups(c *gin.Context) {
	s := service.KingdeeCustomerGroup{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	err = s.PullKingdeeCustomerGroups()
	if err != nil {
		e.Error(500, err, fmt.Sprintf("拉取金蝶客户分组失败！错误详情：%s", err.Error()))
		return
	}
	e.OK("PullKingdeeCustomerGroups", "创建成功")
}

// GetKingdeeCustomerGroups
// @Summary 金蝶客户分组下拉信息
// @Description 金蝶客户分组下拉列表
// @Tags 金蝶店铺
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/group [get]
// @Security Bearer
func (e KingdeeCustomer) GetKingdeeCustomerGroups(c *gin.Context) {
	s := service.KingdeeCustomerGroup{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	list := make([]models.KingdeeCustomerGroup, 0)
	err = s.GetKingdeeCustomerGroups(&list)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取金蝶客户分组失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(list, "获取成功")
}

// PullKingdeeOrganizeInfos
// @Summary 拉取金蝶组织信息
// @Description 拉取金蝶组织列表
// @Tags 金蝶店铺
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/kingdee-customer/organize [get]
// @Security Bearer
func (e KingdeeCustomer) PullKingdeeOrganizeInfos(c *gin.Context) {
	s := service.KingdeeOrganizeInfo{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	err = s.PullKingdeeOrganizeInfos()
	if err != nil {
		e.Error(500, err, fmt.Sprintf("拉取金蝶组织失败！错误详情：%s", err.Error()))
		return
	}
	e.OK("PullKingdeeOrganizeInfos", "创建成功")
}
