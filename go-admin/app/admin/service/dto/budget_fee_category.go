package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type BudgetFeeCategoryGetPageReq struct {
	dto.Pagination `search:"-"`
	CategoryName   string `form:"categoryName"  search:"type:contains;column:category_name;table:budget_fee_category" comment:"预算类别名称"`
	CategoryCode   string `form:"categoryCode"  search:"type:contains;column:category_code;table:budget_fee_category" comment:"预算类别编码"`
	ViewType       int    `form:"viewType"  search:"type:exact;column:view_type;table:budget_fee_category" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图"`
	BudgetFeeCategoryOrder
}

type BudgetFeeCategoryOrder struct {
	Id int64 `form:"idOrder"  search:"type:order;column:id;table:budget_fee_category"`
	/*ParentId       string `form:"parentIdOrder"  search:"type:order;column:parent_id;table:budget_fee_category"`
	CategoryName   string `form:"categoryNameOrder"  search:"type:order;column:category_name;table:budget_fee_category"`
	CategoryNameEn string `form:"categoryNameEnOrder"  search:"type:order;column:category_name_en;table:budget_fee_category"`
	CategoryCode   string `form:"categoryCodeOrder"  search:"type:order;column:category_code;table:budget_fee_category"`
	ViewType       string `form:"viewTypeOrder"  search:"type:order;column:view_type;table:budget_fee_category"`
	CreateBy       string `form:"createByOrder"  search:"type:order;column:create_by;table:budget_fee_category"`
	CreatedAt      string `form:"createdAtOrder"  search:"type:order;column:created_at;table:budget_fee_category"`
	UpdateBy       string `form:"updateByOrder"  search:"type:order;column:update_by;table:budget_fee_category"`
	UpdatedAt      string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:budget_fee_category"`
	DeletedAt      string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:budget_fee_category"`*/
}

func (m *BudgetFeeCategoryGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type BudgetFeeCategoryInsertReq struct {
	Id             int64  `json:"-" comment:""` //
	ParentId       int64  `json:"parentId" comment:""`
	CategoryName   string `json:"categoryName" comment:"类别名称" vd:"@:len($)>0; msg:'类别名称不能为空'"`
	CategoryNameEn string `json:"categoryNameEn" comment:"类别名称英文"`
	CategoryCode   string `json:"categoryCode" comment:"类别编码"`
	ViewType       int    `json:"viewType" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图" vd:"$>0; msg:'视图类型不能为空'"`
	common.ControlBy
}

func (s *BudgetFeeCategoryInsertReq) Generate(model *models.BudgetFeeCategory) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.ParentId = s.ParentId
	model.CategoryName = s.CategoryName
	model.CategoryNameEn = s.CategoryNameEn
	model.CategoryCode = s.CategoryCode
	model.ViewType = s.ViewType
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *BudgetFeeCategoryInsertReq) GetId() interface{} {
	return s.Id
}

type BudgetFeeCategoryUpdateReq struct {
	Id             int64  `uri:"id" comment:""` //
	ParentId       int64  `json:"parentId" comment:""`
	CategoryName   string `json:"categoryName" comment:"类别名称" vd:"@:len($)>0; msg:'类别名称不能为空'"`
	CategoryNameEn string `json:"categoryNameEn" comment:"类别名称英文"`
	CategoryCode   string `json:"categoryCode" comment:"类别编码"`
	ViewType       int    `json:"viewType" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图" vd:"$>0; msg:'视图类型不能为空'"`
	common.ControlBy
}

func (s *BudgetFeeCategoryUpdateReq) Generate(model *models.BudgetFeeCategory) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.ParentId = s.ParentId
	model.CategoryName = s.CategoryName
	model.CategoryNameEn = s.CategoryNameEn
	model.CategoryCode = s.CategoryCode
	model.ViewType = s.ViewType
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *BudgetFeeCategoryUpdateReq) GetId() interface{} {
	return s.Id
}

// BudgetFeeCategoryGetReq 功能获取请求参数
type BudgetFeeCategoryGetReq struct {
	Id int64 `uri:"id"`
}

func (s *BudgetFeeCategoryGetReq) GetId() interface{} {
	return s.Id
}

// BudgetFeeCategoryDeleteReq 功能删除请求参数
type BudgetFeeCategoryDeleteReq struct {
	Ids []int64 `json:"ids"`
}

func (s *BudgetFeeCategoryDeleteReq) GetId() interface{} {
	return s.Ids
}

type BudgetFeeCategoryListTree struct {
	Id             int64                       `gorm:"-" json:"id"`
	ParentId       int64                       `gorm:"-" json:"parentId"`
	CategoryName   string                      `gorm:"-" json:"categoryName"`
	CategoryNameEn string                      `gorm:"-" json:"categoryNameEn"`
	CategoryCode   string                      `gorm:"-" json:"categoryCode"`
	ViewType       int                         `gorm:"-" json:"viewType"`
	Children       []BudgetFeeCategoryListTree `gorm:"-" json:"children"`
}

type BudgetFeeCategoryImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}
