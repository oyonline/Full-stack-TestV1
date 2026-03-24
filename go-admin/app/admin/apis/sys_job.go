package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

type SysJob struct {
	api.Api
}

// GetPage
// @Summary 定时任务列表
// @Description 获取定时任务分页列表数据
// @Tags 定时任务
// @Param pageIndex query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param jobId query int false "任务ID"
// @Param jobName query string false "任务名称"
// @Param jobGroup query string false "任务分组"
// @Param cronExpression query string false "cron表达式"
// @Param invokeTarget query string false "调用目标"
// @Param status query int false "状态"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sysjob [get]
// @Security Bearer
func (e SysJob) GetPage(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobGetPageReq{}
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

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	list := make([]models.SysJob, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get
// @Summary 获取定时任务详情
// @Description 获取定时任务详情数据
// @Tags 定时任务
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sysjob/{id} [get]
// @Security Bearer
func (e SysJob) Get(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobById{}
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

	var object models.SysJob
	// 数据权限检查
	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(http.StatusUnprocessableEntity, err, "查询失败")
		return
	}
	e.OK(object, "查询成功")
}

// Insert
// @Summary 新增定时任务
// @Description 创建定时任务数据
// @Tags 定时任务
// @Accept application/json
// @Product application/json
// @Param data body dto.SysJobInsertReq true "定时任务数据"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sysjob [post]
// @Security Bearer
func (e SysJob) Insert(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobInsertReq{}
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

	// 设置创建人
	req.SetCreateBy(user.GetUserId(c))

	err = s.Insert(&req)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	e.OK(req.GetId(), "创建成功")
}

// Update
// @Summary 修改定时任务
// @Description 更新定时任务数据
// @Tags 定时任务
// @Accept application/json
// @Product application/json
// @Param data body dto.SysJobUpdateReq true "定时任务数据"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sysjob [put]
// @Security Bearer
func (e SysJob) Update(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobUpdateReq{}
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

	// 设置更新人
	req.SetUpdateBy(user.GetUserId(c))

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.Update(&req, p)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	e.OK(req.GetId(), "更新成功")
}

// Delete
// @Summary 删除定时任务
// @Description 删除定时任务数据（支持批量删除）
// @Tags 定时任务
// @Accept application/json
// @Product application/json
// @Param data body dto.SysJobById true "定时任务ID"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sysjob [delete]
// @Security Bearer
func (e SysJob) Delete(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobById{}
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

	// 设置更新人
	req.SetUpdateBy(user.GetUserId(c))

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.Remove(&req, p)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	e.OK(req.GetId(), "删除成功")
}

// StartJob
// @Summary 启动定时任务
// @Description 启动指定的定时任务
// @Tags 定时任务
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/job/start/{id} [get]
// @Security Bearer
func (e SysJob) StartJob(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobById{}
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

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.Start(&req, p)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	e.OK(nil, "启动成功")
}

// RemoveJob
// @Summary 停止定时任务
// @Description 停止/移除指定的定时任务
// @Tags 定时任务
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/job/remove/{id} [get]
// @Security Bearer
func (e SysJob) RemoveJob(c *gin.Context) {
	s := service.SysJob{}
	req := dto.SysJobById{}
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

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.RemoveJob(&req, p)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	e.OK(nil, "停止成功")
}
