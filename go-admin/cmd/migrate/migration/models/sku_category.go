package models

type SkuCategory struct {
	CategoryId   int64  `gorm:"primaryKey;autoIncrement;column:category_id" json:"categoryId"`
	CategoryName string `gorm:"size:64;not null;column:category_name" json:"categoryName"`
	ParentId     int64  `gorm:"not null;default:0;column:parent_id;index" json:"parentId"`
	Level        int    `gorm:"size:4;not null;default:1" json:"level"`
	Sort         int    `gorm:"not null;default:0" json:"sort"`
	Status       int    `gorm:"size:4;not null;default:2" json:"status"`
	ControlBy
	ModelTime
}

func (SkuCategory) TableName() string {
	return "sku_category"
}
