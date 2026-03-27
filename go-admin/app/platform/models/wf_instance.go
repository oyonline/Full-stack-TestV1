package models

import (
	"time"

	common "go-admin/common/models"
)

type WorkflowInstance struct {
	InstanceId       int        `json:"instanceId" gorm:"primaryKey;autoIncrement;comment:流程实例ID"`
	DefinitionId     int        `json:"definitionId" gorm:"index;comment:流程定义ID"`
	DefinitionKey    string     `json:"definitionKey" gorm:"size:64;comment:流程定义编码"`
	DefinitionName   string     `json:"definitionName" gorm:"size:128;comment:流程定义名称"`
	ModuleKey        string     `json:"moduleKey" gorm:"size:64;index;comment:模块编码"`
	BusinessType     string     `json:"businessType" gorm:"size:64;index;comment:业务类型"`
	BusinessId       string     `json:"businessId" gorm:"size:64;index;comment:业务ID"`
	BusinessNo       string     `json:"businessNo" gorm:"size:128;comment:业务单号"`
	Title            string     `json:"title" gorm:"size:255;comment:流程标题"`
	Status           string     `json:"status" gorm:"size:32;index;comment:流程状态"`
	CurrentNodeId    int        `json:"currentNodeId" gorm:"comment:当前节点ID"`
	CurrentNodeKey   string     `json:"currentNodeKey" gorm:"size:64;comment:当前节点编码"`
	CurrentNodeName  string     `json:"currentNodeName" gorm:"size:128;comment:当前节点名称"`
	StarterId        int        `json:"starterId" gorm:"index;comment:发起人ID"`
	StarterName      string     `json:"starterName" gorm:"size:128;comment:发起人名称"`
	StartedAt        time.Time  `json:"startedAt" gorm:"comment:发起时间"`
	FinishedAt       *time.Time `json:"finishedAt" gorm:"comment:完成时间"`
	LastAction       string     `json:"lastAction" gorm:"size:32;comment:最后动作"`
	LastActionRemark string     `json:"lastActionRemark" gorm:"size:255;comment:最后动作说明"`
	common.ControlBy
	common.ModelTime
}

func (*WorkflowInstance) TableName() string {
	return "wf_instance"
}

func (e *WorkflowInstance) GetId() interface{} {
	return e.InstanceId
}
