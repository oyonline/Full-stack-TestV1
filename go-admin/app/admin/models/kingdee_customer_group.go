package models

import "go-admin/common/models"

type KingdeeCustomerGroup struct {
	models.Model
	GroupId     int64  `gorm:"size:20;" json:"FID"`     // FID 客户分组ID
	GroupName   string `gorm:"size:55;" json:"FNAME"`   // FNAME 客户分组名称
	GroupNumber string `gorm:"size:25;" json:"FNUMBER"` // FNUMBER 客户分组编码
	models.ModelTime
}

func (*KingdeeCustomerGroup) TableName() string {
	return "kingdee_customer_group"
}

func (e *KingdeeCustomerGroup) GetId() interface{} {
	return e.Id
}
