package dto

import (
	"errors"
	"strconv"
	"strings"

	"go-admin/app/platform/models"
	cDto "go-admin/common/dto"
	common "go-admin/common/models"
)

const (
	WorkflowStatusDraft    = "draft"
	WorkflowStatusReview   = "in_review"
	WorkflowStatusApproved = "approved"
	WorkflowStatusRejected = "rejected"
	WorkflowStatusCanceled = "cancelled"

	WorkflowTaskPending   = "pending"
	WorkflowTaskApproved  = "approved"
	WorkflowTaskRejected  = "rejected"
	WorkflowTaskCancelled = "cancelled"

	WorkflowActionStart    = "start"
	WorkflowActionApprove  = "approve"
	WorkflowActionReject   = "reject"
	WorkflowActionWithdraw = "withdraw"

	WorkflowNodeTypeStart   = "start"
	WorkflowNodeTypeApprove = "approve"
	WorkflowNodeTypeEnd     = "end"

	WorkflowApproverUser = "user"
	WorkflowApproverRole = "role"
)

type WorkflowDefinitionGetPageReq struct {
	cDto.Pagination `search:"-"`

	DefinitionName string `form:"definitionName"`
	ModuleKey      string `form:"moduleKey"`
	BusinessType   string `form:"businessType"`
	Status         string `form:"status"`
}

type WorkflowDefinitionGetReq struct {
	Id int `uri:"id"`
}

type WorkflowDefinitionInsertReq struct {
	DefinitionKey  string `json:"definitionKey" binding:"required"`
	DefinitionName string `json:"definitionName" binding:"required"`
	ModuleKey      string `json:"moduleKey" binding:"required"`
	BusinessType   string `json:"businessType" binding:"required"`
	Status         string `json:"status"`
	Version        int    `json:"version"`
	Remark         string `json:"remark"`
	common.ControlBy
}

func (m *WorkflowDefinitionInsertReq) Normalize() {
	m.DefinitionKey = strings.TrimSpace(m.DefinitionKey)
	m.DefinitionName = strings.TrimSpace(m.DefinitionName)
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.BusinessType = strings.TrimSpace(m.BusinessType)
	m.Remark = strings.TrimSpace(m.Remark)
	if m.Status == "" {
		m.Status = "2"
	}
	if m.Version <= 0 {
		m.Version = 1
	}
}

func (m *WorkflowDefinitionInsertReq) Generate(model *models.WorkflowDefinition) {
	model.DefinitionKey = m.DefinitionKey
	model.DefinitionName = m.DefinitionName
	model.ModuleKey = m.ModuleKey
	model.BusinessType = m.BusinessType
	model.Status = m.Status
	model.Version = m.Version
	model.Remark = m.Remark
}

type WorkflowDefinitionUpdateReq struct {
	DefinitionId   int    `json:"definitionId" binding:"required"`
	DefinitionKey  string `json:"definitionKey" binding:"required"`
	DefinitionName string `json:"definitionName" binding:"required"`
	ModuleKey      string `json:"moduleKey" binding:"required"`
	BusinessType   string `json:"businessType" binding:"required"`
	Status         string `json:"status"`
	Version        int    `json:"version"`
	Remark         string `json:"remark"`
	common.ControlBy
}

func (m *WorkflowDefinitionUpdateReq) Normalize() {
	m.DefinitionKey = strings.TrimSpace(m.DefinitionKey)
	m.DefinitionName = strings.TrimSpace(m.DefinitionName)
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.BusinessType = strings.TrimSpace(m.BusinessType)
	m.Remark = strings.TrimSpace(m.Remark)
	if m.Status == "" {
		m.Status = "2"
	}
	if m.Version <= 0 {
		m.Version = 1
	}
}

func (m *WorkflowDefinitionUpdateReq) Generate(model *models.WorkflowDefinition) {
	model.DefinitionId = m.DefinitionId
	model.DefinitionKey = m.DefinitionKey
	model.DefinitionName = m.DefinitionName
	model.ModuleKey = m.ModuleKey
	model.BusinessType = m.BusinessType
	model.Status = m.Status
	model.Version = m.Version
	model.Remark = m.Remark
}

type WorkflowDefinitionNodeSaveReq struct {
	DefinitionId int                               `uri:"id"`
	Nodes        []WorkflowDefinitionNodeUpsertReq `json:"nodes" binding:"required,min=1"`
	common.ControlBy
}

type WorkflowDefinitionNodeUpsertReq struct {
	NodeId        int    `json:"nodeId"`
	NodeKey       string `json:"nodeKey" binding:"required"`
	NodeName      string `json:"nodeName" binding:"required"`
	NodeType      string `json:"nodeType" binding:"required"`
	Sort          int    `json:"sort"`
	ApproverType  string `json:"approverType"`
	ApproverValue string `json:"approverValue"`
	ApproverName  string `json:"approverName"`
	Remark        string `json:"remark"`
}

func (m *WorkflowDefinitionNodeUpsertReq) Normalize() {
	m.NodeKey = strings.TrimSpace(m.NodeKey)
	m.NodeName = strings.TrimSpace(m.NodeName)
	m.NodeType = strings.TrimSpace(m.NodeType)
	m.ApproverType = strings.TrimSpace(m.ApproverType)
	m.ApproverValue = strings.TrimSpace(m.ApproverValue)
	m.ApproverName = strings.TrimSpace(m.ApproverName)
	m.Remark = strings.TrimSpace(m.Remark)
}

func (m *WorkflowDefinitionNodeUpsertReq) Validate() error {
	switch m.NodeType {
	case WorkflowNodeTypeStart, WorkflowNodeTypeApprove, WorkflowNodeTypeEnd:
	default:
		return errors.New("节点类型仅支持 start / approve / end")
	}
	if m.Sort <= 0 {
		return errors.New("节点排序必须大于 0")
	}
	if m.NodeType == WorkflowNodeTypeApprove {
		switch m.ApproverType {
		case WorkflowApproverUser, WorkflowApproverRole:
		default:
			return errors.New("审批节点仅支持 user / role")
		}
		if strings.TrimSpace(m.ApproverValue) == "" {
			return errors.New("审批节点必须设置审批对象")
		}
		if _, err := strconv.Atoi(m.ApproverValue); err != nil {
			return errors.New("审批对象必须为数值 ID")
		}
	}
	return nil
}

type WorkflowDefinitionDeleteReq struct {
	Id int `uri:"id" binding:"required"`
}

type WorkflowInstanceStartReq struct {
	DefinitionId int    `json:"definitionId" binding:"required"`
	ModuleKey    string `json:"moduleKey" binding:"required"`
	BusinessType string `json:"businessType" binding:"required"`
	BusinessId   string `json:"businessId" binding:"required"`
	BusinessNo   string `json:"businessNo"`
	Title        string `json:"title" binding:"required"`
	Remark       string `json:"remark"`
}

func (m *WorkflowInstanceStartReq) Normalize() {
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.BusinessType = strings.TrimSpace(m.BusinessType)
	m.BusinessId = strings.TrimSpace(m.BusinessId)
	m.BusinessNo = strings.TrimSpace(m.BusinessNo)
	m.Title = strings.TrimSpace(m.Title)
	m.Remark = strings.TrimSpace(m.Remark)
}

type WorkflowInstanceGetReq struct {
	Id int `uri:"id"`
}

type WorkflowTodoTaskGetPageReq struct {
	cDto.Pagination `search:"-"`

	Title        string `form:"title"`
	BusinessType string `form:"businessType"`
	BusinessNo   string `form:"businessNo"`
	Status       string `form:"status"`
}

type WorkflowStartedInstanceGetPageReq struct {
	cDto.Pagination `search:"-"`

	Title        string `form:"title"`
	BusinessType string `form:"businessType"`
	BusinessNo   string `form:"businessNo"`
	Status       string `form:"status"`
}

type WorkflowTaskActionReq struct {
	Id      int    `uri:"id" binding:"required"`
	Comment string `json:"comment"`
}

func (m *WorkflowTaskActionReq) Normalize() {
	m.Comment = strings.TrimSpace(m.Comment)
}

type WorkflowInstanceWithdrawReq struct {
	Id      int    `uri:"id" binding:"required"`
	Comment string `json:"comment"`
}

func (m *WorkflowInstanceWithdrawReq) Normalize() {
	m.Comment = strings.TrimSpace(m.Comment)
}
