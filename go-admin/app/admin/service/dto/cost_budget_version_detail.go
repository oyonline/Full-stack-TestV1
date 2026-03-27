package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type CostBudgetVersionDetailGetPageReq struct {
	dto.Pagination `search:"-"`
	CostBudgetVersionDetailOrder
}

type CostBudgetVersionDetailOrder struct {
	Id                  string `form:"idOrder"  search:"type:order;column:id;table:cost_budget_version_detail"`
	CostBudgetVersionId string `form:"costBudgetVersionIdOrder"  search:"type:order;column:cost_budget_version_id;table:cost_budget_version_detail"`
	BudgetFeeCategoryId string `form:"budgetFeeCategoryIdOrder"  search:"type:order;column:budget_fee_category_id;table:cost_budget_version_detail"`
	BudgetAmount        string `form:"budgetAmountOrder"  search:"type:order;column:budget_amount;table:cost_budget_version_detail"`
	YearsMonth          string `form:"yearsMonthOrder"  search:"type:order;column:years_month;table:cost_budget_version_detail"`
	CreateBy            string `form:"createByOrder"  search:"type:order;column:create_by;table:cost_budget_version_detail"`
	CreatedAt           string `form:"createdAtOrder"  search:"type:order;column:created_at;table:cost_budget_version_detail"`
	UpdateBy            string `form:"updateByOrder"  search:"type:order;column:update_by;table:cost_budget_version_detail"`
	UpdatedAt           string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:cost_budget_version_detail"`
	DeletedAt           string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:cost_budget_version_detail"`
}

func (m *CostBudgetVersionDetailGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type CostBudgetVersionDetailInsertReq struct {
	Id                  int64   `json:"-" comment:"主键"` // 主键
	CostBudgetVersionId int64   `json:"costBudgetVersionId" comment:"预算版本ID"`
	BudgetFeeCategoryId int64   `json:"budgetFeeCategoryId" comment:"费用类别ID"`
	BudgetAmount        float64 `json:"budgetAmount" comment:"总预算额"`
	YearsMonth          string  `json:"yearsMonth" comment:"年月"`
	common.ControlBy
}

func (s *CostBudgetVersionDetailInsertReq) Generate(model *models.CostBudgetVersionDetail) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostBudgetVersionId = s.CostBudgetVersionId
	model.BudgetFeeCategoryId = s.BudgetFeeCategoryId
	model.BudgetAmount = s.BudgetAmount
	model.YearsMonth = s.YearsMonth
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *CostBudgetVersionDetailInsertReq) GetId() interface{} {
	return s.Id
}

type CostBudgetVersionDetailUpdateReq struct {
	Id                  int64   `uri:"id" comment:"主键"` // 主键
	CostBudgetVersionId int64   `json:"costBudgetVersionId" comment:"预算版本ID"`
	BudgetFeeCategoryId int64   `json:"budgetFeeCategoryId" comment:"费用类别ID"`
	BudgetAmount        float64 `json:"budgetAmount" comment:"总预算额"`
	YearsMonth          string  `json:"yearsMonth" comment:"年月"`
	common.ControlBy
}

func (s *CostBudgetVersionDetailUpdateReq) Generate(model *models.CostBudgetVersionDetail) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostBudgetVersionId = s.CostBudgetVersionId
	model.BudgetFeeCategoryId = s.BudgetFeeCategoryId
	model.BudgetAmount = s.BudgetAmount
	model.YearsMonth = s.YearsMonth
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *CostBudgetVersionDetailUpdateReq) GetId() interface{} {
	return s.Id
}

// CostBudgetVersionDetailGetReq 功能获取请求参数
type CostBudgetVersionDetailGetReq struct {
	Id int64 `uri:"id"`
}

func (s *CostBudgetVersionDetailGetReq) GetId() interface{} {
	return s.Id
}

// CostBudgetVersionDetailDeleteReq 功能删除请求参数
type CostBudgetVersionDetailDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *CostBudgetVersionDetailDeleteReq) GetId() interface{} {
	return s.Ids
}

type CostBudgetVersionDetailImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type CostBudgetVersionDetailExport struct {
	CategoryName   string  `json:"categoryName" comment:"费用类别" excel:"费用类别(必填),sort:1,width:20,required:true"`
	CostCenterName string  `json:"costCenterName" comment:"成本中心名称" excel:"成本中心名称(必填),sort:2,width:20,required:true"`
	Month1         float64 `json:"month1" comment:"1月" excel:"1月,sort:3,width:20"`
	Month2         float64 `json:"month2" comment:"2月" excel:"2月,sort:4,width:20"`
	Month3         float64 `json:"month3" comment:"3月" excel:"3月,sort:5,width:20"`
	Month4         float64 `json:"month4" comment:"4月" excel:"4月,sort:6,width:20"`
	Month5         float64 `json:"month5" comment:"5月" excel:"5月,sort:7,width:20"`
	Month6         float64 `json:"month6" comment:"6月" excel:"6月,sort:8,width:20"`
	Month7         float64 `json:"month7" comment:"7月" excel:"7月,sort:9,width:20"`
	Month8         float64 `json:"month8" comment:"8月" excel:"8月,sort:10,width:20"`
	Month9         float64 `json:"month9" comment:"9月" excel:"9月,sort:11,width:20"`
	Month10        float64 `json:"month10" comment:"10月" excel:"10月,sort:12,width:20"`
	Month11        float64 `json:"month11" comment:"11月" excel:"11月,sort:13,width:20"`
	Month12        float64 `json:"month12" comment:"12月" excel:"12月,sort:14,width:20"`
}

// GetMonthAmount 根据月份索引获取金额（1-12）
func (d *CostBudgetVersionDetailExport) GetMonthAmount(monthIndex int) float64 {
	months := []float64{d.Month1, d.Month2, d.Month3, d.Month4, d.Month5, d.Month6, d.Month7, d.Month8,
		d.Month9, d.Month10, d.Month11, d.Month12}

	if monthIndex < 1 || monthIndex > 12 {
		return 0
	}
	return months[monthIndex-1]
}

func (d *CostBudgetVersionDetailExport) TotalAmount() float64 {
	return d.Month1 + d.Month2 + d.Month3 + d.Month4 + d.Month5 + d.Month6 + d.Month7 + d.Month8 +
		d.Month9 + d.Month10 + d.Month11 + d.Month12
}
