package models

import "time"

type Spu struct {
	SpuId              int64      `gorm:"primaryKey;autoIncrement;column:spu_id" json:"spuId"`
	SpuCode            string     `gorm:"size:64;not null;uniqueIndex;column:spu_code" json:"spuCode"`
	SpuName            string     `gorm:"size:200;not null;column:spu_name" json:"spuName"`
	CategoryId         int64      `gorm:"index;column:category_id" json:"categoryId"`
	BrandId            int64      `gorm:"index;column:brand_id" json:"brandId"`
	Description        string     `gorm:"type:mediumtext;column:description" json:"description"`
	MainImageUrl       string     `gorm:"size:512;column:main_image_url" json:"mainImageUrl"`
	DetailImages       string     `gorm:"type:json;column:detail_images" json:"detailImages"`
	Status             int        `gorm:"size:4;not null;default:1" json:"status"`
	WorkflowInstanceId int64      `gorm:"not null;default:0;column:workflow_instance_id;index" json:"workflowInstanceId"`
	SubmittedAt        *time.Time `gorm:"column:submitted_at" json:"submittedAt"`
	ApprovedAt         *time.Time `gorm:"column:approved_at" json:"approvedAt"`
	CreatorId          int        `gorm:"index;column:creator_id" json:"creatorId"`
	DeptId             int        `gorm:"index;column:dept_id" json:"deptId"`
	ControlBy
	ModelTime
}

func (Spu) TableName() string {
	return "spu"
}
