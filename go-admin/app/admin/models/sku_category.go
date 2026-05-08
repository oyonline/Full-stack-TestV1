package models

import "go-admin/common/models"

const (
	SkuCategoryStatusDisabled = 1
	SkuCategoryStatusEnabled  = 2
)

type SkuCategory struct {
	CategoryId   int64  `gorm:"primaryKey;autoIncrement;column:category_id" json:"categoryId"`
	CategoryName string `gorm:"size:64;not null;column:category_name" json:"categoryName"`
	ParentId     int64  `gorm:"not null;default:0;column:parent_id;index" json:"parentId"`
	Level        int    `gorm:"size:4;not null;default:1" json:"level"`
	Sort         int    `gorm:"not null;default:0" json:"sort"`
	Status       int    `gorm:"size:4;not null;default:2" json:"status"`
	models.ControlBy
	models.ModelTime
}

func (*SkuCategory) TableName() string {
	return "sku_category"
}

func (e *SkuCategory) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SkuCategory) GetId() interface{} {
	return e.CategoryId
}
