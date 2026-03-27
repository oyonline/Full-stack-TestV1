package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/platform/models"
	"go-admin/app/platform/service"
	"go-admin/app/platform/service/dto"
	"go-admin/common/authctx"
	"go-admin/common/middleware"
)

type Workflow struct {
	api.Api
}

func (e Workflow) GetDefinitionPage(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]models.WorkflowDefinition, 0)
	var count int64
	if err = s.GetDefinitionPage(&req, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Workflow) GetDefinition(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	detail, err := s.GetDefinition(req.Id)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(detail, "查询成功")
}

func (e Workflow) GetTodoTaskPage(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowTodoTaskGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]service.WorkflowTodoTaskItem, 0)
	var count int64
	if err = s.GetTodoTaskPage(&req, &list, &count, user.GetUserId(c), authctx.GetRoleIDs(c)); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Workflow) GetStartedInstancePage(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowStartedInstanceGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]service.WorkflowStartedInstanceItem, 0)
	var count int64
	if err = s.GetStartedInstancePage(&req, &list, &count, user.GetUserId(c)); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Workflow) InsertDefinition(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionInsertReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.CreateBy = user.GetUserId(c)
	data, err := s.InsertDefinition(&req)
	if err != nil {
		e.Error(500, err, "创建失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流管理",
		BusinessType:  middleware.AuditActionCreate,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.definition.insert",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("流程编码", req.DefinitionKey),
			middleware.AuditKV("流程名称", req.DefinitionName),
			middleware.AuditKV("模块编码", req.ModuleKey),
			middleware.AuditKV("业务类型", req.BusinessType),
		),
	})
	e.OK(data, "创建成功")
}

func (e Workflow) UpdateDefinition(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionUpdateReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.UpdateBy = user.GetUserId(c)
	data, err := s.UpdateDefinition(&req)
	if err != nil {
		e.Error(500, err, "更新失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流管理",
		BusinessType:  middleware.AuditActionUpdate,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.definition.update",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("流程ID", req.DefinitionId),
			middleware.AuditKV("流程编码", req.DefinitionKey),
			middleware.AuditKV("流程名称", req.DefinitionName),
		),
	})
	e.OK(data, "更新成功")
}

func (e Workflow) DeleteDefinition(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionDeleteReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	if err = s.DeleteDefinition(req.Id); err != nil {
		e.Error(500, err, fmt.Sprintf("删除失败,%s", err.Error()))
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流管理",
		BusinessType:  middleware.AuditActionDelete,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.definition.delete",
		OperatorType:  middleware.AuditOperatorManage,
		Remark:        middleware.AuditKV("流程ID", req.Id),
	})
	e.OK(nil, "删除成功")
}

func (e Workflow) GetDefinitionNodes(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]models.WorkflowDefinitionNode, 0)
	if err = s.GetDefinitionNodes(req.Id, &list); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(list, "查询成功")
}

func (e Workflow) SaveDefinitionNodes(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowDefinitionNodeSaveReq{}
	err := e.MakeContext(c).
		MakeOrm().
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.CreateBy = user.GetUserId(c)
	req.UpdateBy = user.GetUserId(c)
	req.DefinitionId = 0
	if value := c.Param("id"); value != "" {
		fmt.Sscanf(value, "%d", &req.DefinitionId)
	}
	if err = s.SaveDefinitionNodes(req.DefinitionId, &req); err != nil {
		e.Error(500, err, "保存失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流管理",
		BusinessType:  middleware.AuditActionUpdate,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.definition.nodes.save",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("流程ID", req.DefinitionId),
			middleware.AuditCount("节点数量", len(req.Nodes)),
		),
	})
	e.OK(nil, "保存成功")
}

func (e Workflow) StartInstance(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowInstanceStartReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	detail, err := s.Start(c, &req)
	if err != nil {
		e.Error(500, err, "发起失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流实例",
		BusinessType:  middleware.AuditActionStart,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.instance.start",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("流程定义ID", req.DefinitionId),
			middleware.AuditKV("模块编码", req.ModuleKey),
			middleware.AuditKV("业务类型", req.BusinessType),
			middleware.AuditKV("业务ID", req.BusinessId),
			middleware.AuditKV("标题", req.Title),
		),
	})
	e.OK(detail, "发起成功")
}

func (e Workflow) GetInstance(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowInstanceGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	detail, err := s.GetInstanceDetail(req.Id)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.OK(detail, "查询成功")
}

func (e Workflow) ApproveTask(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowTaskActionReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.Normalize()
	if value := c.Param("id"); value != "" {
		fmt.Sscanf(value, "%d", &req.Id)
	}
	detail, err := s.Approve(c, req.Id, req.Comment)
	if err != nil {
		e.Error(500, err, "审批失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流实例",
		BusinessType:  middleware.AuditActionApprove,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.task.approve",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("任务ID", req.Id),
			middleware.AuditKV("审批意见", req.Comment),
		),
	})
	e.OK(detail, "审批成功")
}

func (e Workflow) RejectTask(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowTaskActionReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.Normalize()
	if value := c.Param("id"); value != "" {
		fmt.Sscanf(value, "%d", &req.Id)
	}
	detail, err := s.Reject(c, req.Id, req.Comment)
	if err != nil {
		e.Error(500, err, "驳回失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流实例",
		BusinessType:  middleware.AuditActionReject,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.task.reject",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("任务ID", req.Id),
			middleware.AuditKV("驳回意见", req.Comment),
		),
	})
	e.OK(detail, "驳回成功")
}

func (e Workflow) WithdrawInstance(c *gin.Context) {
	s := service.Workflow{}
	req := dto.WorkflowInstanceWithdrawReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.Normalize()
	if value := c.Param("id"); value != "" {
		fmt.Sscanf(value, "%d", &req.Id)
	}
	detail, err := s.Withdraw(c, req.Id, req.Comment)
	if err != nil {
		e.Error(500, err, "撤回失败,"+err.Error())
		return
	}
	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "审批流实例",
		BusinessType:  middleware.AuditActionWithdraw,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.workflow.instance.withdraw",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("流程实例ID", req.Id),
			middleware.AuditKV("撤回说明", req.Comment),
		),
	})
	e.OK(detail, "撤回成功")
}
