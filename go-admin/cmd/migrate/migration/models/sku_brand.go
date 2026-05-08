package models

type SkuBrand struct {
	BrandId      int64  `gorm:"primaryKey;autoIncrement;column:brand_id" json:"brandId"`
	BrandName    string `gorm:"size:64;not null;uniqueIndex;column:brand_name" json:"brandName"`
	BrandLogoUrl string `gorm:"size:512;column:brand_logo_url" json:"brandLogoUrl"`
	Sort         int    `gorm:"not null;default:0" json:"sort"`
	Status       int    `gorm:"size:4;not null;default:2" json:"status"`
	ControlBy
	ModelTime
}

func (SkuBrand) TableName() string {
	return "sku_brand"
}
