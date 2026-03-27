package models

import (
	"go-admin/common/models"
)

type BudgetFeeCategoryDetails struct {
	models.Model
	BudgetFeeCategoryId int64  `json:"budgetFeeCategoryId" gorm:"type:bigint(20);comment:费用类别id"`
	FeeName             string `json:"feeName" gorm:"type:varchar(255);comment:费用名称" excel:"费用名称(必填),sort:2,width:20,required:true"`
	FeeNameEn           string `json:"feeNameEn" gorm:"type:varchar(255);comment:费用名称英文" excel:"费用英文名称(必填),sort:3,width:20,required:true"`
	FeeCode             string `json:"feeCode" gorm:"type:varchar(150);comment:费用编码" excel:"费用编码(必填),sort:1,width:20,required:true"`
	Source              string `json:"source" gorm:"type:varchar(50);comment:数据来源"`
	Description         string `json:"description" gorm:"type:varchar(500);comment:描述" excel:"描述,sort:6,width:20"`
	ViewType            int    `json:"viewType" gorm:"type:tinyint(1);comment:1 金蝶科目视图 2 经营管理视图 3项目管理视图" excel:"视图类型(必填),sort:4,width:20,converter:1=金蝶科目视图|2=经营管理视图|3=项目管理视图,required:true"`
	Platform            string `json:"platform" gorm:"type:varchar(150);comment:平台" excel:"客户分组,sort:5,width:20"`

	models.ModelTime
	models.ControlBy
	//业务字段
	PathStr      string `json:"pathStr" gorm:"->;column:path_str" excel:"归属费用类别全路径,sort:7,width:20"`
	CategoryName string `json:"categoryName" gorm:"->;column:category_name"`
}

func (BudgetFeeCategoryDetails) TableName() string {
	return "budget_fee_category_details"
}

func (e *BudgetFeeCategoryDetails) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *BudgetFeeCategoryDetails) GetId() interface{} {
	return e.Id
}
