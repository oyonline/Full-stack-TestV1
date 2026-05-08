package models

import "go-admin/common/models"

const (
	SkuStatusDisabled = 1
	SkuStatusEnabled  = 2
)

type Sku struct {
	SkuId   int64   `gorm:"primaryKey;autoIncrement;column:sku_id" json:"skuId"`
	SpuId   int64   `gorm:"not null;index;column:spu_id" json:"spuId"`
	SkuCode string  `gorm:"size:64;not null;uniqueIndex;column:sku_code" json:"skuCode"`
	SkuName string  `gorm:"size:200;column:sku_name" json:"skuName"`
	Spec    string  `gorm:"size:255" json:"spec"`
	Unit    string  `gorm:"size:20" json:"unit"`
	Price   float64 `gorm:"type:decimal(12,2);not null;default:0" json:"price"`
	Status  int     `gorm:"size:4;not null;default:1" json:"status"`
	models.ControlBy
	models.ModelTime
}

func (*Sku) TableName() string {
	return "sku"
}

func (e *Sku) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *Sku) GetId() interface{} {
	return e.SkuId
}
