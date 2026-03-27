package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
	"strings"
)

type BudgetFeeCategoryDetailsGetPageReq struct {
	dto.Pagination      `search:"-"`
	FeeName             string `form:"feeName" search:"type:contains;column:fee_name;table:budget_fee_category_details" comment:"费用名称"`
	FeeCode             string `form:"feeCode" search:"type:contains;column:fee_code;table:budget_fee_category_details" comment:"费用编码"`
	BudgetFeeCategoryId int64  `form:"budgetFeeCategoryId" search:"type:exact;column:budget_fee_category_id;table:budget_fee_category_details" comment:"费用类别id"`
	BudgetFeeCategoryDetailsOrder
}

type BudgetFeeCategoryDetailsOrder struct {
	Id int64 `form:"idOrder"  search:"type:order;column:id;table:budget_fee_category_details"`
}

func (m *BudgetFeeCategoryDetailsGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type BudgetFeeCategoryDetailsInsertReq struct {
	Id                  int64  `json:"-" comment:""` //
	BudgetFeeCategoryId int64  `json:"budgetFeeCategoryId" comment:"费用类别id" vd:"$>0; msg:'费用类别不能为空'"`
	FeeName             string `json:"feeName" comment:"费用名称" vd:"@:len($)>0; msg:'费用名称不能为空'"`
	FeeNameEn           string `json:"feeNameEn" comment:"费用名称英文"`
	FeeCode             string `json:"feeCode" comment:"费用编码" vd:"@:len($)>0; msg:'费用编码不能为空'"`
	Source              string `json:"source" comment:"数据来源"`
	Description         string `json:"description" comment:"描述"`
	ViewType            int    `json:"viewType" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图" vd:"$>0; msg:'视图类型不能为空'"`
	Platform            string `json:"platform" comment:"平台"`
	common.ControlBy
}

func (s *BudgetFeeCategoryDetailsInsertReq) Generate(model *models.BudgetFeeCategoryDetails) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.BudgetFeeCategoryId = s.BudgetFeeCategoryId
	model.FeeName = s.FeeName
	model.FeeNameEn = s.FeeNameEn
	model.FeeCode = s.FeeCode
	model.Source = "人工维护"
	model.Description = s.Description
	model.ViewType = s.ViewType
	model.Platform = "/" + strings.ReplaceAll(s.Platform, ",", "/") + "/"
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *BudgetFeeCategoryDetailsInsertReq) GetId() interface{} {
	return s.Id
}

type BudgetFeeCategoryDetailsUpdateReq struct {
	Id                  int64  `uri:"id" comment:""` //
	BudgetFeeCategoryId int64  `json:"budgetFeeCategoryId" comment:"费用类别id" vd:"$>0; msg:'费用类别不能为空'"`
	FeeName             string `json:"feeName" comment:"费用名称" vd:"@:len($)>0; msg:'费用名称不能为空'"`
	FeeNameEn           string `json:"feeNameEn" comment:"费用名称英文"`
	FeeCode             string `json:"feeCode" comment:"费用编码" vd:"@:len($)>0; msg:'费用编码不能为空'"`
	Source              string `json:"source" comment:"数据来源"`
	Description         string `json:"description" comment:"描述"`
	ViewType            int    `json:"viewType" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图" vd:"$>0; msg:'视图类型不能为空'"`
	Platform            string `json:"platform" comment:"平台"`
	common.ControlBy
}

func (s *BudgetFeeCategoryDetailsUpdateReq) Generate(model *models.BudgetFeeCategoryDetails) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.BudgetFeeCategoryId = s.BudgetFeeCategoryId
	model.FeeName = s.FeeName
	model.FeeNameEn = s.FeeNameEn
	model.FeeCode = s.FeeCode
	model.Source = "人工维护"
	model.Description = s.Description
	model.ViewType = s.ViewType
	model.Platform = "/" + strings.ReplaceAll(s.Platform, ",", "/") + "/"
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *BudgetFeeCategoryDetailsUpdateReq) GetId() interface{} {
	return s.Id
}

// BudgetFeeCategoryDetailsGetReq 功能获取请求参数
type BudgetFeeCategoryDetailsGetReq struct {
	Id int64 `uri:"id"`
}

func (s *BudgetFeeCategoryDetailsGetReq) GetId() interface{} {
	return s.Id
}

// BudgetFeeCategoryDetailsDeleteReq 功能删除请求参数
type BudgetFeeCategoryDetailsDeleteReq struct {
	Ids []int64 `json:"ids"`
}

func (s *BudgetFeeCategoryDetailsDeleteReq) GetId() interface{} {
	return s.Ids
}

type BudgetFeeCategoryDetailsImportReq struct {
	BudgetFeeCategoryId int64                 `form:"budgetFeeCategoryId" binding:"required,gte=1" comment:"费用类别id"`
	ViewType            int                   `form:"viewType" binding:"required,gte=1" comment:"1 金蝶科目视图 2 经营管理视图 3项目管理视图"`
	File                *multipart.FileHeader `form:"file" comment:"文件"`
}
