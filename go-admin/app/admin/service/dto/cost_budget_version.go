package dto

import (
	"time"

	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type CostBudgetVersionGetPageReq struct {
	dto.Pagination   `search:"-"`
	CostBudgetName   string `form:"costBudgetName"  search:"type:contains;column:cost_budget_name;table:cost_budget_version" comment:"版本名称"`
	CostBudgetCode   string `form:"costBudgetCode"  search:"type:contains;column:cost_budget_code;table:cost_budget_version" comment:"版本编码"`
	Years            int    `form:"years"  search:"type:exact;column:years;table:cost_budget_version" comment:"预算年度"`
	Status           int    `form:"status"  search:"type:exact;column:status;table:cost_budget_version" comment:"状态(1=草稿|2=生效中)"`
	CostCenterInfoId int64  `form:"costCenterInfoId"  search:"type:exact;column:cost_center_info_id;table:cost_budget_version" comment:"成本中心ID"`
	CostBudgetVersionOrder
}

type CostBudgetVersionOrder struct {
	Id string `form:"idOrder"  search:"type:order;column:id;table:cost_budget_version"`
	/*CostBudgetName   string `form:"costBudgetNameOrder"  search:"type:order;column:cost_budget_name;table:cost_budget_version"`
	CostBudgetCode   string `form:"costBudgetCodeOrder"  search:"type:order;column:cost_budget_code;table:cost_budget_version"`
	Years            string `form:"yearsOrder"  search:"type:order;column:years;table:cost_budget_version"`
	EffectiveDate    string `form:"effectiveDateOrder"  search:"type:order;column:effective_date;table:cost_budget_version"`
	BudgetAmount     string `form:"budgetAmountOrder"  search:"type:order;column:budget_amount;table:cost_budget_version"`
	Description      string `form:"descriptionOrder"  search:"type:order;column:description;table:cost_budget_version"`
	Status           string `form:"statusOrder"  search:"type:order;column:status;table:cost_budget_version"`
	CostCenterInfoId string `form:"costCenterInfoIdOrder"  search:"type:order;column:cost_center_info_id;table:cost_budget_version"`
	CreateBy         string `form:"createByOrder"  search:"type:order;column:create_by;table:cost_budget_version"`
	CreatedAt        string `form:"createdAtOrder"  search:"type:order;column:created_at;table:cost_budget_version"`
	UpdateBy         string `form:"updateByOrder"  search:"type:order;column:update_by;table:cost_budget_version"`
	UpdatedAt        string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:cost_budget_version"`
	DeletedAt        string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:cost_budget_version"`*/
}

func (m *CostBudgetVersionGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type CostBudgetVersionInsertReq struct {
	Id               int64     `json:"-" comment:"主键"` // 主键
	CostBudgetName   string    `json:"costBudgetName" comment:"版本名称" vd:"@:len($)>0; msg:'版本名称不能为空'"`
	CostBudgetCode   string    `json:"costBudgetCode" comment:"版本编码"`
	Years            int       `json:"years" comment:"预算年度" vd:"$!=0; msg:'预算年度不能为空'"`
	EffectiveDate    time.Time `json:"effectiveDate" comment:"生效日期" vd:"!$IsZero(); msg:'生效日期不能为空'"`
	BudgetAmount     float64   `json:"budgetAmount" comment:"总预算额"`
	Description      string    `json:"description" comment:"描述备注"`
	Status           int       `json:"status" comment:"状态(1=草稿|2=生效中)"`
	CostCenterInfoId int64     `json:"costCenterInfoId" comment:"成本中心ID" vd:"$!=0; msg:'成本中心不能为空'"`
	common.ControlBy
	// 文件上传字段（用于导入解析）
	File *multipart.FileHeader `json:"file" form:"file" comment:"导入文件"`
}

func (s *CostBudgetVersionInsertReq) Generate(model *models.CostBudgetVersion) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostBudgetName = s.CostBudgetName
	model.CostBudgetCode = s.CostBudgetCode
	model.Years = s.Years
	model.EffectiveDate = s.EffectiveDate
	model.BudgetAmount = s.BudgetAmount
	model.Description = s.Description
	model.Status = s.Status
	model.CostCenterInfoId = s.CostCenterInfoId
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *CostBudgetVersionInsertReq) GetId() interface{} {
	return s.Id
}

type CostBudgetVersionUpdateReq struct {
	Id               int64     `uri:"id" comment:"主键"` // 主键
	CostBudgetName   string    `json:"costBudgetName" comment:"版本名称" vd:"@:len($)>0; msg:'版本名称不能为空'"`
	CostBudgetCode   string    `json:"costBudgetCode" comment:"版本编码"`
	Years            int       `json:"years" comment:"预算年度" vd:"$!=0; msg:'预算年度不能为空'"`
	EffectiveDate    time.Time `json:"effectiveDate" comment:"生效日期" vd:"!$IsZero(); msg:'生效日期不能为空'"`
	BudgetAmount     float64   `json:"budgetAmount" comment:"总预算额"`
	Description      string    `json:"description" comment:"描述备注"`
	Status           int       `json:"status" comment:"状态(1=草稿|2=生效中)"`
	CostCenterInfoId int64     `json:"costCenterInfoId" comment:"成本中心ID" vd:"$!=0; msg:'成本中心不能为空'"`
	common.ControlBy
	// 文件上传字段（用于导入解析）
	File *multipart.FileHeader `json:"file" form:"file" comment:"导入文件"`
}

func (s *CostBudgetVersionUpdateReq) Generate(model *models.CostBudgetVersion) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostBudgetName = s.CostBudgetName
	model.CostBudgetCode = s.CostBudgetCode
	model.Years = s.Years
	model.EffectiveDate = s.EffectiveDate
	model.BudgetAmount = s.BudgetAmount
	model.Description = s.Description
	model.Status = s.Status
	model.CostCenterInfoId = s.CostCenterInfoId
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *CostBudgetVersionUpdateReq) GetId() interface{} {
	return s.Id
}

// CostBudgetVersionGetReq 功能获取请求参数
type CostBudgetVersionGetReq struct {
	Id int64 `uri:"id"`
}

func (s *CostBudgetVersionGetReq) GetId() interface{} {
	return s.Id
}

// CostBudgetVersionDeleteReq 功能删除请求参数
type CostBudgetVersionDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *CostBudgetVersionDeleteReq) GetId() interface{} {
	return s.Ids
}

type CostBudgetVersionImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type CostBudgetVersionExport struct {
	Id               int64     `json:"-" comment:"主键"` // 主键
	CostBudgetName   string    `json:"costBudgetName" comment:"版本名称" excel:"版本名称,sort:0,width:20,required:true"`
	CostBudgetCode   string    `json:"costBudgetCode" comment:"版本编码" excel:"版本编码,sort:0,width:20,required:true"`
	Years            int       `json:"years" comment:"预算年度" excel:"预算年度,sort:0,width:20,required:true"`
	EffectiveDate    time.Time `json:"effectiveDate" comment:"生效日期" excel:"生效日期,sort:0,width:20,required:true"`
	BudgetAmount     float64   `json:"budgetAmount" comment:"总预算额" excel:"总预算额,sort:0,width:20,required:true"`
	Description      string    `json:"description" comment:"描述备注" excel:"描述备注,sort:0,width:20,required:true"`
	Status           int       `json:"status" comment:"状态(1=草稿|2=生效中)" excel:"状态(1=草稿|2=生效中),sort:0,width:20,required:true"`
	CostCenterInfoId int64     `json:"costCenterInfoId" comment:"成本中心ID" excel:"成本中心ID,sort:0,width:20,required:true"`
}

type OriginalData struct {
	Id                  int64   `json:"id"`
	ParentId            int64   `json:"parentId"`
	CategoryName        string  `json:"categoryName"`
	CostBudgetVersionId int64   `json:"costBudgetVersionId"`
	YearsMonth          string  `json:"yearsMonth"`
	BudgetAmount        float64 `json:"budgetAmount"`
}

type MonthData struct {
	Id                  int64       `json:"id"`
	ParentId            int64       `json:"parentId"`
	CategoryName        string      `json:"categoryName"`
	CostBudgetVersionId int64       `json:"costBudgetVersionId"`
	Month1              float64     `json:"month1"`
	Month2              float64     `json:"month2"`
	Month3              float64     `json:"month3"`
	Month4              float64     `json:"month4"`
	Month5              float64     `json:"month5"`
	Month6              float64     `json:"month6"`
	Month7              float64     `json:"month7"`
	Month8              float64     `json:"month8"`
	Month9              float64     `json:"month9"`
	Month10             float64     `json:"month10"`
	Month11             float64     `json:"month11"`
	Month12             float64     `json:"month12"`
	MonthTotal          float64     `json:"monthTotal"`
	Children            []MonthData `json:"children"`
}

type GroupResult struct {
	CostBudgetVersion models.CostBudgetVersion `json:"costBudgetVersion"`
	TreeData          []MonthData              `json:"treeData"`
	Month1            float64                  `json:"month1"`
	Month2            float64                  `json:"month2"`
	Month3            float64                  `json:"month3"`
	Month4            float64                  `json:"month4"`
	Month5            float64                  `json:"month5"`
	Month6            float64                  `json:"month6"`
	Month7            float64                  `json:"month7"`
	Month8            float64                  `json:"month8"`
	Month9            float64                  `json:"month9"`
	Month10           float64                  `json:"month10"`
	Month11           float64                  `json:"month11"`
	Month12           float64                  `json:"month12"`
	MonthTotal        float64                  `json:"monthTotal"`
}

func (data *MonthData) SetMonthValue(month int, amount float64) {
	switch month {
	case 1:
		data.Month1 += amount
	case 2:
		data.Month2 += amount
	case 3:
		data.Month3 += amount
	case 4:
		data.Month4 += amount
	case 5:
		data.Month5 += amount
	case 6:
		data.Month6 += amount
	case 7:
		data.Month7 += amount
	case 8:
		data.Month8 += amount
	case 9:
		data.Month9 += amount
	case 10:
		data.Month10 += amount
	case 11:
		data.Month11 += amount
	case 12:
		data.Month12 += amount
	}
	data.MonthTotal += amount
}
