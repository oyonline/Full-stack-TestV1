package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

type CostCenterInfo struct {
	api.Api
}

// GetPage 获取成本中心数据列表
// @Summary 获取成本中心数据列表
// @Description 获取成本中心数据列表
// @Tags 成本中心数据
// @Param costCenterName query string false "成本中心名称"
// @Param costCenterCode query string false "成本中心编码"
// @Param deptId query int64 false "上级部门"
// @Param status query int false "状态(1=停用|2=启用)"
// @Param pageSize query int false "页条数"
// @Param pageIndex query int false "页码"
// @Success 200 {object} response.Response{data=response.Page{list=[]models.CostCenterInfo}} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-info/list [get]
// @Security Bearer
func (e CostCenterInfo) GetPage(c *gin.Context) {
	req := dto.CostCenterInfoGetPageReq{}
	s := service.CostCenterInfo{}
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
	var count int64

	list, err := s.GetPage(&req, p, &count)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取成本中心数据
// @Summary 获取成本中心数据
// @Description 获取成本中心数据
// @Tags 成本中心数据
// @Param id path int false "id"
// @Success 200 {object} response.Response{data=models.CostCenterInfo} "{"code": 200, "data": [...]}"
// @Router /api/v1/cost-center-info/get/{id} [get]
// @Security Bearer
func (e CostCenterInfo) Get(c *gin.Context) {
	req := dto.CostCenterInfoGetReq{}
	s := service.CostCenterInfo{}
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
	object, err := s.Get(&req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("获取成本中心数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(object, "查询成功")
}

// Insert 创建成本中心数据
// @Summary 创建成本中心数据
// @Description 创建成本中心数据
// @Tags 成本中心数据
// @Accept application/json
// @Product application/json
// @Param data body dto.CostCenterInfoInsertReq true "data"
// @Success 200 {object} response.Response	"{"code": 200, "message": "添加成功"}"
// @Router /api/v1/cost-center-info/add [post]
// @Security Bearer
func (e CostCenterInfo) Insert(c *gin.Context) {
	req := dto.CostCenterInfoInsertReq{}
	s := service.CostCenterInfo{}
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
		e.Error(500, err, fmt.Sprintf("创建成本中心数据失败，\r\n失败信息 %s", err.Error()))
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update 修改成本中心数据
// @Summary 修改成本中心数据
// @Description 修改成本中心数据
// @Tags 成本中心数据
// @Accept application/json
// @Product application/json
// @Param id path int true "id"
// @Param data body dto.CostCenterInfoUpdateReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "修改成功"}"
// @Router /api/v1/cost-center-info/edit/{id} [put]
// @Security Bearer
func (e CostCenterInfo) Update(c *gin.Context) {
	req := dto.CostCenterInfoUpdateReq{}
	s := service.CostCenterInfo{}
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
		e.Error(500, err, fmt.Sprintf("修改成本中心数据失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除成本中心数据
// @Summary 删除成本中心数据
// @Description 删除成本中心数据
// @Tags 成本中心数据
// @Param data body dto.CostCenterInfoDeleteReq true "body"
// @Success 200 {object} response.Response	"{"code": 200, "message": "删除成功"}"
// @Router /api/v1/cost-center-info/remove [delete]
// @Security Bearer
func (e CostCenterInfo) Delete(c *gin.Context) {
	s := service.CostCenterInfo{}
	req := dto.CostCenterInfoDeleteReq{}
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
		e.Error(500, err, fmt.Sprintf("删除成本中心数据失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// Import 导入成本中心数据
// @Summary 导入成本中心数据
// @Description 导入成本中心数据
// @Tags 成本中心数据
// @Accept multipart/form-data
// @Product multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response	"{"code": 200, "message": "导入成功"}"
// @Router /api/v1/cost-center-info/import [post]
// @Security Bearer
func (e CostCenterInfo) Import(c *gin.Context) {
	var importReq dto.CostCenterInfoImportReq
	s := service.CostCenterInfo{}
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
		e.Error(500, err, fmt.Sprintf("导入成本中心数据失败信息 %s", err.Error()))
		return
	}
	e.OK(importReq, "导入成功")
}

// Export 导出成本中心数据
// @Summary 导出成本中心数据
// @Description 导出成本中心数据
// @Tags 成本中心数据
// @Param costCenterName query string false "成本中心名称"
// @Param costCenterCode query string false "成本中心编码"
// @Param deptId query int64 false "上级部门"
// @Param status query int false "状态(1=停用|2=启用)"
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-info/export [get]
// @Security Bearer
func (e CostCenterInfo) Export(c *gin.Context) {
	req := dto.CostCenterInfoGetPageReq{}
	s := service.CostCenterInfo{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	err = s.Export(c.Writer, &req, p)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("导出成本中心数据失败信息 %s", err.Error()))
		return
	}
}

// DownloadTemplate 下载成本中心数据模板
// @Summary 下载成本中心数据模板
// @Description 下载成本中心数据模板
// @Tags 成本中心数据
// @Success 200 {object} response.Response
// @Router /api/v1/cost-center-info/template [get]
// @Security Bearer
func (e CostCenterInfo) DownloadTemplate(c *gin.Context) {
	s := service.CostCenterInfo{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	err = s.DownloadTemplate(c.Writer)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("下载成本中心数据失败信息 %s", err.Error()))
		return
	}
}

func (e CostCenterInfo) CostCenterTimeTask(c *gin.Context) {
	s := service.CostCenterInfo{}
	err := e.MakeContext(c).MakeOrm().MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	userId := user.GetUserId(c)
	e.Logger.Info("开始执行CostCenterTimeTask{}", userId)
	err = s.CostCenterTimeTask()
	if err != nil {
		e.Error(500, err, fmt.Sprintf("CostCenterTimeTask失败，失败信息 %s", err.Error()))
		return
	}
	e.OK(nil, "执行CostCenterTimeTask成功")
}
