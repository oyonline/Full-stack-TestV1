package models

import (
	"time"

	"go-admin/common/models"
)

type AllocationRuleSettings struct {
	models.Model

	AllocationName             string     `json:"allocationName" gorm:"type:varchar(150);comment:分摊规则名称"`
	BudgetFeeCategoryDetailsId int64      `json:"budgetFeeCategoryDetailsId" gorm:"type:bigint(20);comment:费用明细来源ID"`
	AllocationType             int        `json:"allocationType" gorm:"type:tinyint(1);comment:分摊类型类型(1=固定比例|2=按销售额分摊)"`
	Status                     int        `json:"status" gorm:"type:tinyint(1);comment:状态(1=停用|2=启用|3=待生效)"`
	EffectiveDate              *time.Time `json:"effectiveDate" gorm:"type:date;comment:生效日期"`
	ExpiredDate                *time.Time `json:"expiredDate" gorm:"type:date;comment:失效日期"`
	Description                string     `json:"description" gorm:"type:varchar(500);comment:描述备注"`
	models.ModelTime
	models.ControlBy

	AllocationRuleSettingsDept []AllocationRuleSettingsDept `json:"allocationRuleSettingsDept" gorm:"-"`
	FeeName                    string                       `json:"feeName" gorm:"-"`
	FeeCode                    string                       `json:"feeCode" gorm:"-"`
}

func (AllocationRuleSettings) TableName() string {
	return "allocation_rule_settings"
}

func (e *AllocationRuleSettings) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *AllocationRuleSettings) GetId() interface{} {
	return e.Id
}
