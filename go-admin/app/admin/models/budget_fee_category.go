package models

import (
	"go-admin/common/models"
)

type BudgetFeeCategory struct {
	models.Model
	ParentId       int64  `json:"parentId" gorm:"type:bigint(20);comment:ParentId"`
	CategoryName   string `json:"categoryName" gorm:"type:varchar(50);comment:类别名称" excel:"类别名称(必填),sort:1,width:20,required:true"`
	CategoryNameEn string `json:"categoryNameEn" gorm:"type:varchar(100);comment:类别名称英文" excel:"类别名称英文,sort:2,width:20"`
	CategoryCode   string `json:"categoryCode" gorm:"type:varchar(150);comment:类别编码" excel:"类别编码,sort:3,width:20"`
	ViewType       int    `json:"viewType" gorm:"type:tinyint(1);comment:1 金蝶科目视图 2 经营管理视图 3项目管理视图" excel:"视图类型(必填),sort:4,width:20,converter:1=金蝶科目视图|2=经营管理视图|3=项目管理视图,required:true"`
	Level          int    `json:"level" gorm:"type:int(11);comment:层级"`
	PathStr        string `json:"pathStr" gorm:"type:varchar(255);comment:路径" excel:"类别全路径,sort:5,width:20,operation:export"`
	PathStrId      string `json:"pathStrId" gorm:"type:varchar(150);comment:路径Id"`
	models.ModelTime
	models.ControlBy

	ParentStr    string              `json:"parentStr" gorm:"-" excel:"父级类别,sort:6,width:20,operation:import"`
	ChildrenData []BudgetFeeCategory `json:"-" gorm:"-"`
}

func (BudgetFeeCategory) TableName() string {
	return "budget_fee_category"
}

func (e *BudgetFeeCategory) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *BudgetFeeCategory) GetId() interface{} {
	return e.Id
}
