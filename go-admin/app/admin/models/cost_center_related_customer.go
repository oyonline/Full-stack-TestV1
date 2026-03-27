package models

import (
	"go-admin/common/models"
)

type CostCenterRelatedCustomer struct {
	models.Model

	CostCenterInfoId int64 `json:"costCenterInfoId" gorm:"type:bigint(20);comment:成本中心ID"`
	GroupId          int64 `json:"groupId" gorm:"type:bigint(20);comment:客户分组ID"`
	models.ModelTime
	models.ControlBy
}

func (CostCenterRelatedCustomer) TableName() string {
	return "cost_center_related_customer"
}

func (e *CostCenterRelatedCustomer) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostCenterRelatedCustomer) GetId() interface{} {
	return e.Id
}
