package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type AllocationRuleSettingsDeptGetPageReq struct {
	dto.Pagination           `search:"-"`
	AllocationType           int   `form:"allocationType"  search:"type:exact;column:allocation_type;table:allocation_rule_settings_dept" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	AllocationRuleSettingsId int64 `form:"allocationRuleSettingsId"  search:"type:exact;column:allocation_rule_settings_id;table:allocation_rule_settings_dept" comment:"分摊规则设置ID"`
	AllocationRuleSettingsDeptOrder
}

type AllocationRuleSettingsDeptOrder struct {
	Id                       string `form:"idOrder"  search:"type:order;column:id;table:allocation_rule_settings_dept"`
	AllocationType           string `form:"allocationTypeOrder"  search:"type:order;column:allocation_type;table:allocation_rule_settings_dept"`
	AllocationRuleSettingsId string `form:"allocationRuleSettingsIdOrder"  search:"type:order;column:allocation_rule_settings_id;table:allocation_rule_settings_dept"`
	ScaleSettings            string `form:"scaleSettingsOrder"  search:"type:order;column:scale_settings;table:allocation_rule_settings_dept"`
	AssociationId            string `form:"associationIdOrder"  search:"type:order;column:association_id;table:allocation_rule_settings_dept"`
	CreateBy                 string `form:"createByOrder"  search:"type:order;column:create_by;table:allocation_rule_settings_dept"`
	CreatedAt                string `form:"createdAtOrder"  search:"type:order;column:created_at;table:allocation_rule_settings_dept"`
	UpdateBy                 string `form:"updateByOrder"  search:"type:order;column:update_by;table:allocation_rule_settings_dept"`
	UpdatedAt                string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:allocation_rule_settings_dept"`
	DeletedAt                string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:allocation_rule_settings_dept"`
}

func (m *AllocationRuleSettingsDeptGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type AllocationRuleSettingsDeptInsertReq struct {
	Id                       int64   `json:"-" comment:"主键"` // 主键
	AllocationType           int     `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	AllocationRuleSettingsId int64   `json:"allocationRuleSettingsId" comment:"分摊规则设置ID"`
	ScaleSettings            float64 `json:"scaleSettings" comment:"分摊比例"`
	AssociationId            int64   `json:"associationId" comment:"关联费用承担部门ID"`
	common.ControlBy
}

func (s *AllocationRuleSettingsDeptInsertReq) Generate(model *models.AllocationRuleSettingsDept) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.AllocationType = s.AllocationType
	model.AllocationRuleSettingsId = s.AllocationRuleSettingsId
	model.ScaleSettings = s.ScaleSettings
	model.AssociationId = s.AssociationId
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *AllocationRuleSettingsDeptInsertReq) GetId() interface{} {
	return s.Id
}

type AllocationRuleSettingsDeptUpdateReq struct {
	Id                       int64   `uri:"id" comment:"主键"` // 主键
	AllocationType           int     `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)"`
	AllocationRuleSettingsId int64   `json:"allocationRuleSettingsId" comment:"分摊规则设置ID"`
	ScaleSettings            float64 `json:"scaleSettings" comment:"分摊比例"`
	AssociationId            int64   `json:"associationId" comment:"关联费用承担部门ID"`
	common.ControlBy
}

func (s *AllocationRuleSettingsDeptUpdateReq) Generate(model *models.AllocationRuleSettingsDept) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.AllocationType = s.AllocationType
	model.AllocationRuleSettingsId = s.AllocationRuleSettingsId
	model.ScaleSettings = s.ScaleSettings
	model.AssociationId = s.AssociationId
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *AllocationRuleSettingsDeptUpdateReq) GetId() interface{} {
	return s.Id
}

// AllocationRuleSettingsDeptGetReq 功能获取请求参数
type AllocationRuleSettingsDeptGetReq struct {
	Id int64 `uri:"id"`
}

func (s *AllocationRuleSettingsDeptGetReq) GetId() interface{} {
	return s.Id
}

// AllocationRuleSettingsDeptDeleteReq 功能删除请求参数
type AllocationRuleSettingsDeptDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *AllocationRuleSettingsDeptDeleteReq) GetId() interface{} {
	return s.Ids
}

type AllocationRuleSettingsDeptImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type AllocationRuleSettingsDeptExport struct {
	Id                       int64   `json:"-" comment:"主键"` // 主键
	AllocationType           int     `json:"allocationType" comment:"分摊类型类型(1=固定比例|2=按销售额分摊)" excel:"分摊类型类型(1=固定比例|2=按销售额分摊),sort:0,width:20,required:true"`
	AllocationRuleSettingsId int64   `json:"allocationRuleSettingsId" comment:"分摊规则设置ID" excel:"分摊规则设置ID,sort:0,width:20,required:true"`
	ScaleSettings            float64 `json:"scaleSettings" comment:"分摊比例" excel:"分摊比例,sort:0,width:20,required:true"`
	AssociationId            int64   `json:"associationId" comment:"关联费用承担部门ID" excel:"关联费用承担部门ID,sort:0,width:20,required:true"`
}
