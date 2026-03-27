package models

import common "go-admin/common/models"

type WorkflowDefinitionNode struct {
	NodeId         int    `json:"nodeId" gorm:"primaryKey;autoIncrement;comment:流程节点ID"`
	DefinitionId   int    `json:"definitionId" gorm:"index;comment:流程定义ID"`
	NodeKey        string `json:"nodeKey" gorm:"size:64;comment:节点编码"`
	NodeName       string `json:"nodeName" gorm:"size:128;comment:节点名称"`
	NodeType       string `json:"nodeType" gorm:"size:32;comment:节点类型"`
	Sort           int    `json:"sort" gorm:"default:0;comment:排序"`
	ApproverType   string `json:"approverType" gorm:"size:32;comment:审批人类型"`
	ApproverValue  string `json:"approverValue" gorm:"size:128;comment:审批人配置值"`
	ApproverName   string `json:"approverName" gorm:"size:128;comment:审批人展示名"`
	Remark         string `json:"remark" gorm:"size:255;comment:备注"`
	common.ControlBy
	common.ModelTime
}

func (*WorkflowDefinitionNode) TableName() string {
	return "wf_definition_node"
}

func (e *WorkflowDefinitionNode) GetId() interface{} {
	return e.NodeId
}
