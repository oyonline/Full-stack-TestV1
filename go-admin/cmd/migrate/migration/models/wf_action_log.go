package models

import (
	"time"

	common "go-admin/common/models"
)

type WorkflowActionLog struct {
	LogId        int       `json:"logId" gorm:"primaryKey;autoIncrement;comment:动作日志ID"`
	InstanceId   int       `json:"instanceId" gorm:"index;comment:流程实例ID"`
	TaskId       int       `json:"taskId" gorm:"index;comment:任务ID"`
	Action       string    `json:"action" gorm:"size:32;comment:动作"`
	FromStatus   string    `json:"fromStatus" gorm:"size:32;comment:原状态"`
	ToStatus     string    `json:"toStatus" gorm:"size:32;comment:目标状态"`
	FromNodeKey  string    `json:"fromNodeKey" gorm:"size:64;comment:原节点编码"`
	FromNodeName string    `json:"fromNodeName" gorm:"size:128;comment:原节点名称"`
	ToNodeKey    string    `json:"toNodeKey" gorm:"size:64;comment:目标节点编码"`
	ToNodeName   string    `json:"toNodeName" gorm:"size:128;comment:目标节点名称"`
	OperatorId   int       `json:"operatorId" gorm:"index;comment:操作人ID"`
	OperatorName string    `json:"operatorName" gorm:"size:128;comment:操作人名称"`
	Comment      string    `json:"comment" gorm:"size:255;comment:备注"`
	common.ControlBy
	CreatedAt time.Time `json:"createdAt" gorm:"comment:创建时间"`
}

func (*WorkflowActionLog) TableName() string {
	return "wf_action_log"
}
