package models

import common "go-admin/common/models"

type WorkflowDefinition struct {
	DefinitionId   int    `json:"definitionId" gorm:"primaryKey;autoIncrement;comment:流程定义ID"`
	DefinitionKey  string `json:"definitionKey" gorm:"size:64;uniqueIndex;comment:流程定义编码"`
	DefinitionName string `json:"definitionName" gorm:"size:128;comment:流程定义名称"`
	ModuleKey      string `json:"moduleKey" gorm:"size:64;index;comment:模块编码"`
	BusinessType   string `json:"businessType" gorm:"size:64;index;comment:业务类型"`
	Status         string `json:"status" gorm:"size:4;default:2;comment:状态"`
	Version        int    `json:"version" gorm:"default:1;comment:版本"`
	Remark         string `json:"remark" gorm:"size:255;comment:备注"`
	common.ControlBy
	common.ModelTime
}

func (*WorkflowDefinition) TableName() string {
	return "wf_definition"
}
