package models

import "time"

type WorkflowTask struct {
	TaskId          int        `json:"taskId" gorm:"primaryKey;autoIncrement;comment:审批任务ID"`
	InstanceId      int        `json:"instanceId" gorm:"index;comment:流程实例ID"`
	DefinitionId    int        `json:"definitionId" gorm:"index;comment:流程定义ID"`
	NodeId          int        `json:"nodeId" gorm:"index;comment:节点ID"`
	NodeKey         string     `json:"nodeKey" gorm:"size:64;comment:节点编码"`
	NodeName        string     `json:"nodeName" gorm:"size:128;comment:节点名称"`
	AssigneeType    string     `json:"assigneeType" gorm:"size:32;comment:审批人类型"`
	AssigneeId      int        `json:"assigneeId" gorm:"index;comment:审批对象ID"`
	AssigneeName    string     `json:"assigneeName" gorm:"size:128;comment:审批对象名称"`
	Status          string     `json:"status" gorm:"size:32;index;comment:任务状态"`
	Action          string     `json:"action" gorm:"size:32;comment:处理动作"`
	Comment         string     `json:"comment" gorm:"size:255;comment:处理意见"`
	ActionBy        int        `json:"actionBy" gorm:"index;comment:处理人ID"`
	ActionByName    string     `json:"actionByName" gorm:"size:128;comment:处理人名称"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"comment:创建时间"`
	ProcessedAt     *time.Time `json:"processedAt" gorm:"comment:处理时间"`
	CancelledReason string     `json:"cancelledReason" gorm:"size:255;comment:取消原因"`
}

func (*WorkflowTask) TableName() string {
	return "wf_task"
}
