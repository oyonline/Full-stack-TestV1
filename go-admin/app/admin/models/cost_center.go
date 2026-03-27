package models

import (
	"go-admin/common/models"
)

type CostCenter struct {
	models.Model

	ParentId    string `json:"parentId" gorm:"type:int(11);comment:上级ID"`
	Code        string `json:"code" gorm:"type:varchar(30);comment:编码"`
	Name        string `json:"name" gorm:"type:varchar(100);comment:名称"`
	GroupId     string `json:"groupId" gorm:"type:int(11);comment:归属客户分组kingdee_customer_group.id"`
	CostCode    string `json:"costCode" gorm:"type:varchar(50);comment:费用编码"`
	Type        string `json:"type" gorm:"type:int(11);comment:成本类型 1营销费用 2管理费用"`
	Description string `json:"description" gorm:"type:varchar(500);comment:描述"`
	models.ModelTime
	models.ControlBy
}

func (CostCenter) TableName() string {
	return "cost_center_info"
}

func (e *CostCenter) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostCenter) GetId() interface{} {
	return e.Id
}
