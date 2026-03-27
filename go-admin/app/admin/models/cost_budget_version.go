package models

import (
	"time"

	"go-admin/common/models"
)

type CostBudgetVersion struct {
	models.Model

	CostBudgetName   string    `json:"costBudgetName" gorm:"type:varchar(150);comment:版本名称"`
	CostBudgetCode   string    `json:"costBudgetCode" gorm:"type:varchar(50);comment:版本编码"`
	Years            int       `json:"years" gorm:"type:int(10);comment:预算年度"`
	EffectiveDate    time.Time `json:"effectiveDate" gorm:"type:date;comment:生效日期"`
	BudgetAmount     float64   `json:"budgetAmount" gorm:"type:decimal(18,2);comment:总预算额"`
	Description      string    `json:"description" gorm:"type:varchar(500);comment:描述备注"`
	Status           int       `json:"status" gorm:"type:tinyint(1);comment:状态(1=草稿|2=生效中|3=已归档)"`
	CostCenterInfoId int64     `json:"costCenterInfoId" gorm:"type:bigint(20);comment:成本中心ID"`
	models.ModelTime
	models.ControlBy

	CostCenterName string `json:"costCenterName" gorm:"-"`
	CostCenterCode string `json:"costCenterCode" gorm:"-"`
}

func (CostBudgetVersion) TableName() string {
	return "cost_budget_version"
}

func (e *CostBudgetVersion) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostBudgetVersion) GetId() interface{} {
	return e.Id
}
