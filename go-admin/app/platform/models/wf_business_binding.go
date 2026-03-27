package models

import common "go-admin/common/models"

type WorkflowBusinessBinding struct {
	BindingId       int    `json:"bindingId" gorm:"primaryKey;autoIncrement;comment:绑定ID"`
	ModuleKey       string `json:"moduleKey" gorm:"size:64;index;comment:模块编码"`
	BusinessType    string `json:"businessType" gorm:"size:64;index;comment:业务类型"`
	BusinessId      string `json:"businessId" gorm:"size:64;index;comment:业务ID"`
	BusinessNo      string `json:"businessNo" gorm:"size:128;comment:业务单号"`
	Title           string `json:"title" gorm:"size:255;comment:业务标题"`
	InstanceId      int    `json:"instanceId" gorm:"index;comment:流程实例ID"`
	WorkflowStatus  string `json:"workflowStatus" gorm:"size:32;comment:流程状态"`
	BusinessStatus  string `json:"businessStatus" gorm:"size:32;comment:业务状态"`
	LastAction      string `json:"lastAction" gorm:"size:32;comment:最后动作"`
	LastActionRemark string `json:"lastActionRemark" gorm:"size:255;comment:最后动作说明"`
	common.ControlBy
	common.ModelTime
}

func (*WorkflowBusinessBinding) TableName() string {
	return "wf_business_binding"
}

func (e *WorkflowBusinessBinding) GetId() interface{} {
	return e.BindingId
}
