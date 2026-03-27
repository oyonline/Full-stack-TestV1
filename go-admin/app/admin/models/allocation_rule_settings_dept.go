package models

import (
	"go-admin/common/models"
)

type AllocationRuleSettingsDept struct {
	models.Model

	AllocationType           int     `json:"allocationType" gorm:"type:tinyint(1);comment:分摊类型类型(1=固定比例|2=按销售额分摊)"`
	AllocationRuleSettingsId int64   `json:"allocationRuleSettingsId" gorm:"type:bigint(20);comment:分摊规则设置ID"`
	ScaleSettings            float64 `json:"scaleSettings" gorm:"type:decimal(18,2);comment:分摊比例"`
	AssociationId            int64   `json:"associationId" gorm:"type:bigint(20);comment:关联费用承担部门ID"`
	models.ModelTime
	models.ControlBy

	DeptName     string `json:"deptName" gorm:"->;column:dept_name"`
	DeptPathName string `json:"deptPathName" gorm:"->;column:dept_path_name"`
}

func (AllocationRuleSettingsDept) TableName() string {
	return "allocation_rule_settings_dept"
}

func (e *AllocationRuleSettingsDept) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *AllocationRuleSettingsDept) GetId() interface{} {
	return e.Id
}
