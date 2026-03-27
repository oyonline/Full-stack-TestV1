package dto

import (
	"time"

	"encoding/json"
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type CostCenterInfoChangeGetPageReq struct {
	dto.Pagination   `search:"-"`
	CostCenterInfoId int64 `form:"costCenterInfoId"  search:"type:exact;column:cost_center_info_id;table:cost_center_info_change" comment:"成本中心ID"`
	CostCenterInfoChangeOrder
}

type CostCenterInfoChangeOrder struct {
	Id string `form:"idOrder"  search:"type:order;column:id;table:cost_center_info_change"`
	/*ChangeOrder      string `form:"changeOrderOrder"  search:"type:order;column:change_order;table:cost_center_info_change"`
	ChangeType       string `form:"changeTypeOrder"  search:"type:order;column:change_type;table:cost_center_info_change"`
	Status           string `form:"statusOrder"  search:"type:order;column:status;table:cost_center_info_change"`
	EffectiveDate    string `form:"effectiveDateOrder"  search:"type:order;column:effective_date;table:cost_center_info_change"`
	VersionNumber    string `form:"versionNumberOrder"  search:"type:order;column:version_number;table:cost_center_info_change"`
	ChangeDetails    string `form:"changeDetailsOrder"  search:"type:order;column:change_details;table:cost_center_info_change"`
	CostCenterInfoId string `form:"costCenterInfoIdOrder"  search:"type:order;column:cost_center_info_id;table:cost_center_info_change"`
	CreateBy         string `form:"createByOrder"  search:"type:order;column:create_by;table:cost_center_info_change"`
	CreatedAt        string `form:"createdAtOrder"  search:"type:order;column:created_at;table:cost_center_info_change"`
	UpdateBy         string `form:"updateByOrder"  search:"type:order;column:update_by;table:cost_center_info_change"`
	UpdatedAt        string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:cost_center_info_change"`
	DeletedAt        string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:cost_center_info_change"`*/
}

func (m *CostCenterInfoChangeGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type CostCenterInfoChangeInsertReq struct {
	Id               int64           `json:"-" comment:"主键"` // 主键
	ChangeOrder      string          `json:"changeOrder" comment:"变更单号"`
	ChangeType       string          `json:"changeType" comment:"变更类型"`
	Status           int             `json:"status" comment:"状态(1=已生效，2=未生效)"`
	EffectiveDate    time.Time       `json:"effectiveDate" comment:"生效日期"`
	VersionNumber    string          `json:"versionNumber" comment:"版本号"`
	ChangeDetails    json.RawMessage `json:"changeDetails" comment:"变更内容"`
	CostCenterInfoId int64           `json:"costCenterInfoId" comment:"成本中心ID"`
	common.ControlBy
}

func (s *CostCenterInfoChangeInsertReq) Generate(model *models.CostCenterInfoChange) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.ChangeOrder = s.ChangeOrder
	model.ChangeType = s.ChangeType
	model.Status = s.Status
	model.EffectiveDate = s.EffectiveDate
	model.VersionNumber = s.VersionNumber
	model.ChangeDetails = s.ChangeDetails
	model.CostCenterInfoId = s.CostCenterInfoId
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *CostCenterInfoChangeInsertReq) GetId() interface{} {
	return s.Id
}

type CostCenterInfoChangeUpdateReq struct {
	Id               int64           `uri:"id" comment:"主键"` // 主键
	ChangeOrder      string          `json:"changeOrder" comment:"变更单号"`
	ChangeType       string          `json:"changeType" comment:"变更类型"`
	Status           int             `json:"status" comment:"状态(1=已生效，2=未生效)"`
	EffectiveDate    time.Time       `json:"effectiveDate" comment:"生效日期"`
	VersionNumber    string          `json:"versionNumber" comment:"版本号"`
	ChangeDetails    json.RawMessage `json:"changeDetails" comment:"变更内容"`
	CostCenterInfoId int64           `json:"costCenterInfoId" comment:"成本中心ID"`
	common.ControlBy
}

func (s *CostCenterInfoChangeUpdateReq) Generate(model *models.CostCenterInfoChange) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.ChangeOrder = s.ChangeOrder
	model.ChangeType = s.ChangeType
	model.Status = s.Status
	model.EffectiveDate = s.EffectiveDate
	model.VersionNumber = s.VersionNumber
	model.ChangeDetails = s.ChangeDetails
	model.CostCenterInfoId = s.CostCenterInfoId
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *CostCenterInfoChangeUpdateReq) GetId() interface{} {
	return s.Id
}

// CostCenterInfoChangeGetReq 功能获取请求参数
type CostCenterInfoChangeGetReq struct {
	Id int64 `uri:"id"`
}

func (s *CostCenterInfoChangeGetReq) GetId() interface{} {
	return s.Id
}

// CostCenterInfoChangeDeleteReq 功能删除请求参数
type CostCenterInfoChangeDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *CostCenterInfoChangeDeleteReq) GetId() interface{} {
	return s.Ids
}

type CostCenterInfoChangeImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type CostCenterInfoChangeExport struct {
	Id               int64           `json:"-" comment:"主键"` // 主键
	ChangeOrder      string          `json:"changeOrder" comment:"变更单号" excel:"变更单号,sort:0,width:20,required:true"`
	ChangeType       string          `json:"changeType" comment:"变更类型" excel:"变更类型,sort:0,width:20,required:true"`
	Status           int             `json:"status" comment:"状态(1=已生效，2=未生效)" excel:"状态(1=已生效，2=未生效),sort:0,width:20,required:true"`
	EffectiveDate    time.Time       `json:"effectiveDate" comment:"生效日期" excel:"生效日期,sort:0,width:20,required:true"`
	VersionNumber    string          `json:"versionNumber" comment:"版本号" excel:"版本号,sort:0,width:20,required:true"`
	ChangeDetails    json.RawMessage `json:"changeDetails" comment:"变更内容" excel:"变更内容,sort:0,width:20,required:true"`
	CostCenterInfoId int64           `json:"costCenterInfoId" comment:"成本中心ID" excel:"成本中心ID,sort:0,width:20,required:true"`
}
