package models

import (
	"go-admin/common/models"
)

type CostBudgetVersionDetail struct {
	models.Model

	CostBudgetVersionId int64   `json:"costBudgetVersionId" gorm:"type:bigint(20);comment:预算版本ID"`
	BudgetFeeCategoryId int64   `json:"budgetFeeCategoryId" gorm:"type:bigint(20);comment:费用类别ID"`
	BudgetAmount        float64 `json:"budgetAmount" gorm:"type:decimal(18,2);comment:总预算额"`
	BudgetUsed          float64 `json:"budgetUsed" gorm:"column:budget_used;type:decimal(10,2)" comment:"已使用预算"`
	YearsMonth          string  `json:"yearsMonth" gorm:"type:varchar(10);comment:年月"`
	models.ModelTime
	models.ControlBy
}

func (CostBudgetVersionDetail) TableName() string {
	return "cost_budget_version_detail"
}

func (e *CostBudgetVersionDetail) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostBudgetVersionDetail) GetId() interface{} {
	return e.Id
}
