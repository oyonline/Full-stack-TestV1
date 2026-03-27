package dto

import (
	"time"

	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type AllocationRuleSettingsGetPageReq struct {
	dto.Pagination             `search:"-"`
	AllocationName             string `form:"allocationName"  search:"type:contains;column:allocation_name;table:allocation_rule_settings" comment:"分摊规则名称"`
	BudgetFeeCategoryDetailsId int64  `form:"budgetFeeCategoryDetailsId"  search:"type:exact;column:budget_fee_category_details_id;table:allocation_rule_settings" comment:"费用明细来源ID"`
	AllocationType             int    `form:"allocationType"  search:"type:exact;column:allocation_type;table:allocation_rule_settings" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	Status                     int    `form:"status"  search:"type:exact;column:status;table:allocation_rule_settings" comment:"状态(1=停用|2=启用|3=待生效)"`
	AllocationRuleSettingsOrder
}

type AllocationRuleSettingsOrder struct {
	Id string `form:"idOrder"  search:"type:order;column:id;table:allocation_rule_settings"`
	/*AllocationName string `form:"allocationNameOrder"  search:"type:order;column:allocation_name;table:allocation_rule_settings"`
	  BudgetFeeCategoryDetailsId string `form:"budgetFeeCategoryDetailsIdOrder"  search:"type:order;column:budget_fee_category_details_id;table:allocation_rule_settings"`
	  AllocationType string `form:"allocationTypeOrder"  search:"type:order;column:allocation_type;table:allocation_rule_settings"`
	  Status string `form:"statusOrder"  search:"type:order;column:status;table:allocation_rule_settings"`
	  EffectiveDate string `form:"effectiveDateOrder"  search:"type:order;column:effective_date;table:allocation_rule_settings"`
	  ExpiredDate string `form:"expiredDateOrder"  search:"type:order;column:expired_date;table:allocation_rule_settings"`
	  Description string `form:"descriptionOrder"  search:"type:order;column:description;table:allocation_rule_settings"`
	  CreateBy string `form:"createByOrder"  search:"type:order;column:create_by;table:allocation_rule_settings"`
	  CreatedAt string `form:"createdAtOrder"  search:"type:order;column:created_at;table:allocation_rule_settings"`
	  UpdateBy string `form:"updateByOrder"  search:"type:order;column:update_by;table:allocation_rule_settings"`
	  UpdatedAt string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:allocation_rule_settings"`
	  DeletedAt string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:allocation_rule_settings"`*/

}

func (m *AllocationRuleSettingsGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type AllocationRuleSettingsInsertReq struct {
	Id                         int64      `json:"-" comment:"主键"` // 主键
	AllocationName             string     `json:"allocationName" comment:"分摊规则名称"`
	BudgetFeeCategoryDetailsId int64      `json:"budgetFeeCategoryDetailsId" comment:"费用明细来源ID"`
	AllocationType             int        `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	Status                     int        `json:"status" comment:"状态(1=停用|2=启用|3=待生效)"`
	EffectiveDate              *time.Time `json:"effectiveDate" comment:"生效日期"`
	ExpiredDate                *time.Time `json:"expiredDate" comment:"失效日期"`
	Description                string     `json:"description" comment:"描述备注"`
	common.ControlBy
	AllocationRuleSettingsDept []models.AllocationRuleSettingsDept `json:"allocationRuleSettingsDept"`
}

func (s *AllocationRuleSettingsInsertReq) Generate(model *models.AllocationRuleSettings) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.AllocationName = s.AllocationName
	model.BudgetFeeCategoryDetailsId = s.BudgetFeeCategoryDetailsId
	model.AllocationType = s.AllocationType
	model.Status = s.Status
	model.EffectiveDate = s.EffectiveDate
	model.ExpiredDate = s.ExpiredDate
	model.Description = s.Description
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
	model.AllocationRuleSettingsDept = s.AllocationRuleSettingsDept
}

func (s *AllocationRuleSettingsInsertReq) GetId() interface{} {
	return s.Id
}

type AllocationRuleSettingsUpdateReq struct {
	Id                         int64      `uri:"id" comment:"主键"` // 主键
	AllocationName             string     `json:"allocationName" comment:"分摊规则名称"`
	BudgetFeeCategoryDetailsId int64      `json:"budgetFeeCategoryDetailsId" comment:"费用明细来源ID"`
	AllocationType             int        `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	Status                     int        `json:"status" comment:"状态(1=停用|2=启用|3=待生效)"`
	EffectiveDate              *time.Time `json:"effectiveDate" comment:"生效日期"`
	ExpiredDate                *time.Time `json:"expiredDate" comment:"失效日期"`
	Description                string     `json:"description" comment:"描述备注"`
	common.ControlBy
	AllocationRuleSettingsDept []models.AllocationRuleSettingsDept `json:"allocationRuleSettingsDept"`
}

func (s *AllocationRuleSettingsUpdateReq) Generate(model *models.AllocationRuleSettings) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.AllocationName = s.AllocationName
	model.BudgetFeeCategoryDetailsId = s.BudgetFeeCategoryDetailsId
	model.AllocationType = s.AllocationType
	model.Status = s.Status
	model.EffectiveDate = s.EffectiveDate
	model.ExpiredDate = s.ExpiredDate
	model.Description = s.Description
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
	model.AllocationRuleSettingsDept = s.AllocationRuleSettingsDept
}

func (s *AllocationRuleSettingsUpdateReq) GetId() interface{} {
	return s.Id
}

// AllocationRuleSettingsGetReq 功能获取请求参数
type AllocationRuleSettingsGetReq struct {
	Id int64 `uri:"id"`
}

func (s *AllocationRuleSettingsGetReq) GetId() interface{} {
	return s.Id
}

// AllocationRuleSettingsDeleteReq 功能删除请求参数
type AllocationRuleSettingsDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *AllocationRuleSettingsDeleteReq) GetId() interface{} {
	return s.Ids
}

type AllocationRuleSettingsImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type AllocationRuleSettingsExport struct {
	AllocationName      string     `json:"allocationName" comment:"分摊规则名称" excel:"分摊规则名称,sort:1,width:20,required:true"`
	FeeCode             string     `json:"feeCode" comment:"费用来源" excel:"费用来源编码,sort:2,width:20,required:true"`
	AllocationType      int        `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)" excel:"分摊类型类型,sort:3,width:20,required:true,converter:1=固定比例|2=按销售额分摊"`
	Status              int        `json:"status" comment:"状态(1=停用|2=启用|3=待生效)" excel:"状态,sort:4,width:20,required:true,converter:1=停用|2=启用|3=待生效"`
	EffectiveDate       *time.Time `json:"effectiveDate" comment:"生效日期" excel:"生效日期,sort:5,width:20,required:true"`
	ExpiredDate         *time.Time `json:"expiredDate" comment:"失效日期" excel:"失效日期,sort:6,width:20"`
	Description         string     `json:"description" comment:"描述备注" excel:"描述备注,sort:8,width:20"`
	RuleSettingsDeptStr string     `json:"ruleSettingsDeptStr" comment:"费用承担部门" excel:"费用承担部门,sort:7,width:20"`
}
