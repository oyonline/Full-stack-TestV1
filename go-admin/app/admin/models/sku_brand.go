package models

import "go-admin/common/models"

const (
	SkuBrandStatusDisabled = 1
	SkuBrandStatusEnabled  = 2
)

type SkuBrand struct {
	BrandId      int64  `gorm:"primaryKey;autoIncrement;column:brand_id" json:"brandId"`
	BrandName    string `gorm:"size:64;not null;uniqueIndex;column:brand_name" json:"brandName"`
	BrandLogoUrl string `gorm:"size:512;column:brand_logo_url" json:"brandLogoUrl"`
	Sort         int    `gorm:"not null;default:0" json:"sort"`
	Status       int    `gorm:"size:4;not null;default:2" json:"status"`
	models.ControlBy
	models.ModelTime
}

func (*SkuBrand) TableName() string {
	return "sku_brand"
}

func (e *SkuBrand) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SkuBrand) GetId() interface{} {
	return e.BrandId
}
