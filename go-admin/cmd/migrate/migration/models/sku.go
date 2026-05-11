package models

type Sku struct {
	SkuId   int64   `gorm:"primaryKey;autoIncrement;column:sku_id" json:"skuId"`
	SpuId   int64   `gorm:"not null;index:idx_sku_spu_status,priority:1;column:spu_id" json:"spuId"`
	SkuCode string  `gorm:"size:64;not null;uniqueIndex;column:sku_code" json:"skuCode"`
	SkuName string  `gorm:"size:200;column:sku_name" json:"skuName"`
	Spec    string  `gorm:"size:255" json:"spec"`
	Unit    string  `gorm:"size:20" json:"unit"`
	Price   float64 `gorm:"type:decimal(12,2);not null;default:0" json:"price"`
	Status  int     `gorm:"size:4;not null;default:1;index:idx_sku_spu_status,priority:2" json:"status"`
	ControlBy
	ModelTime
}

func (Sku) TableName() string {
	return "sku"
}
