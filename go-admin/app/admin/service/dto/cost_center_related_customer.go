package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type CostCenterRelatedCustomerGetPageReq struct {
	dto.Pagination `search:"-"`
	CostCenterRelatedCustomerOrder
}

type CostCenterRelatedCustomerOrder struct {
	Id               string `form:"idOrder"  search:"type:order;column:id;table:cost_center_related_customer"`
	CostCenterInfoId string `form:"costCenterInfoIdOrder"  search:"type:order;column:cost_center_info_id;table:cost_center_related_customer"`
	GroupId          string `form:"groupIdOrder"  search:"type:order;column:group_id;table:cost_center_related_customer"`
	CreateBy         string `form:"createByOrder"  search:"type:order;column:create_by;table:cost_center_related_customer"`
	CreatedAt        string `form:"createdAtOrder"  search:"type:order;column:created_at;table:cost_center_related_customer"`
	UpdateBy         string `form:"updateByOrder"  search:"type:order;column:update_by;table:cost_center_related_customer"`
	UpdatedAt        string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:cost_center_related_customer"`
	DeletedAt        string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:cost_center_related_customer"`
}

func (m *CostCenterRelatedCustomerGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type CostCenterRelatedCustomerInsertReq struct {
	Id               int64 `json:"-" comment:"主键"` // 主键
	CostCenterInfoId int64 `json:"costCenterInfoId" comment:"成本中心ID"`
	GroupId          int64 `json:"groupId" comment:"客户分组ID"`
	common.ControlBy
}

func (s *CostCenterRelatedCustomerInsertReq) Generate(model *models.CostCenterRelatedCustomer) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostCenterInfoId = s.CostCenterInfoId
	model.GroupId = s.GroupId
	model.CreateBy = s.CreateBy // 添加这而，需要记录是被谁创建的
}

func (s *CostCenterRelatedCustomerInsertReq) GetId() interface{} {
	return s.Id
}

type CostCenterRelatedCustomerUpdateReq struct {
	Id               int64 `uri:"id" comment:"主键"` // 主键
	CostCenterInfoId int64 `json:"costCenterInfoId" comment:"成本中心ID"`
	GroupId          int64 `json:"groupId" comment:"客户分组ID"`
	common.ControlBy
}

func (s *CostCenterRelatedCustomerUpdateReq) Generate(model *models.CostCenterRelatedCustomer) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostCenterInfoId = s.CostCenterInfoId
	model.GroupId = s.GroupId
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
}

func (s *CostCenterRelatedCustomerUpdateReq) GetId() interface{} {
	return s.Id
}

// CostCenterRelatedCustomerGetReq 功能获取请求参数
type CostCenterRelatedCustomerGetReq struct {
	Id int64 `uri:"id"`
}

func (s *CostCenterRelatedCustomerGetReq) GetId() interface{} {
	return s.Id
}

// CostCenterRelatedCustomerDeleteReq 功能删除请求参数
type CostCenterRelatedCustomerDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *CostCenterRelatedCustomerDeleteReq) GetId() interface{} {
	return s.Ids
}

type CostCenterRelatedCustomerImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type CostCenterRelatedCustomerExport struct {
	Id               int64 `json:"-" comment:"主键"` // 主键
	CostCenterInfoId int64 `json:"costCenterInfoId" comment:"成本中心ID" excel:"成本中心ID,sort:0,width:20,required:true"`
	GroupId          int64 `json:"groupId" comment:"客户分组ID" excel:"客户分组ID,sort:0,width:20,required:true"`
}
