package dto

import (
	"go-admin/app/admin/models"
	common "go-admin/common/models"
	"mime/multipart"

	"go-admin/common/dto"
)

// KingdeeCustomerPageReq 列表或者搜索使用结构体
type KingdeeCustomerPageReq struct {
	dto.Pagination `search:"-"`
	CustomerNumber string `form:"customerNumber" search:"type:contains;column:customer_number;table:kingdee_customer" comment:"店铺编码"`
	CustomerName   string `form:"customerName" search:"type:contains;column:customer_name;table:kingdee_customer" comment:"店铺名称"`
	GroupId        int64  `form:"groupId" search:"type:exact;column:group_id;table:kingdee_customer" comment:"客户分组ID"`
	Country        string `form:"country" search:"type:exact;column:country;table:kingdee_customer" comment:"国家"`
	ForbidStatus   string `form:"forbidStatus" search:"type:exact;column:forbid_status;table:kingdee_customer" comment:"禁用状态(A:启用,B:禁用)"`
	DeptId         int64  `form:"deptId" search:"type:exact;column:dept_id;table:kingdee_customer" comment:"归属部门ID"`
}

func (m *KingdeeCustomerPageReq) GetNeedSearch() interface{} {
	return *m
}

// KingdeeCustomerInsertReq 增使用的结构体
type KingdeeCustomerInsertReq struct {
	CustId         int64  `form:"custId"  comment:"金蝶客户ID"`
	CustomerNumber string `form:"customerNumber" comment:"客户编码"`
	CustomerName   string `form:"customerName"  comment:"客户名称"`
	Country        string `form:"country" comment:"国家"`
	CreateOrgId    int64  `form:"createOrgId" comment:"创建组织ID"`
	UseOrgId       int64  `form:"useOrgId" comment:"使用组织ID"`
	GroupId        int64  `form:"groupId" comment:"客户分组ID"`
	GroupName      string `form:"groupName" comment:"客户分组名称"`
	GroupNumber    string `form:"groupNumber" comment:"客户分组编码"`
	CustomerStatus string `form:"customerStatus" comment:"单据状态(A:创建,B:审核中,C:已审核,D:重新审核,Z:暂存)"`
	ForbidStatus   string `form:"forbidStatus" comment:"禁用状态(A:启用,B:禁用)"`
	CreateDate     string `form:"createDate" comment:"金蝶创建日期"`
	ModifyDate     string `form:"modifyDate" comment:"金蝶修改日期"`
	Remark         string `form:"remark" comment:"备注"`
	DeptId         int64  `form:"deptId" comment:"归属部门ID"`
	CostId         int64  `form:"costId" comment:"成本中心ID"`
	common.ControlBy
}

func (s *KingdeeCustomerInsertReq) Generate(model *models.KingdeeCustomer) {
	model.CustomerNumber = s.CustomerNumber
	model.CustomerName = s.CustomerName
	model.Country = s.Country
	model.CreateOrgId = s.CreateOrgId
	model.UseOrgId = s.UseOrgId
	model.GroupId = s.GroupId
	model.GroupName = s.GroupName
	model.GroupNumber = s.GroupNumber
	model.CustomerStatus = s.CustomerStatus
	model.ForbidStatus = s.ForbidStatus
	model.CreateDate = s.CreateDate
	model.ModifyDate = s.ModifyDate
	model.Remark = s.Remark
	model.DeptId = s.DeptId
	model.CostId = s.CostId
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

// KingdeeCustomerUpdateReq 改使用的结构体
type KingdeeCustomerUpdateReq struct {
	Id     int64  `uri:"id"  comment:"id"`
	Remark string `form:"forbidStatus" comment:"备注"`
	common.ControlBy
}

func (s *KingdeeCustomerUpdateReq) Generate(model *models.KingdeeCustomer) {
	model.Id = s.Id
	model.Remark = s.Remark
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *KingdeeCustomerUpdateReq) GetId() interface{} {
	return s.Id
}

// KingdeeCustomerGetReq 获取单个的结构体
type KingdeeCustomerGetReq struct {
	Id int64 `uri:"id"`
}

func (s *KingdeeCustomerGetReq) GetId() interface{} {
	return s.Id
}

// KingdeeCustomerDeleteReq 删除的结构体
type KingdeeCustomerDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *KingdeeCustomerDeleteReq) Generate(model *models.KingdeeCustomer) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *KingdeeCustomerDeleteReq) GetId() interface{} {
	return s.Ids
}

type KingdeeCustomerImport struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

type KingdeeCustomerExport struct {
	CustomerNumber string `json:"customerNumber" excel:"金蝶店铺编码,width:20,required:true"`
	CustomerName   string `json:"customerName" excel:"金蝶店铺名称,width:40,required:true"`
	GroupName      string `json:"groupName" excel:"客户分组,width:20,required:true"`
	UseOrgName     string `json:"useOrgName" excel:"金蝶归属组织,width:30,required:true"`
	ForbidStatus   string `json:"forbidStatus" excel:"状态,width:20,converter:A=启用|B=禁用,required:true"`
	Country        string `json:"country" excel:"所属地区,width:20,required:true"`
	Remark         string `json:"remark" excel:"备注,width:50,operation:export"`
}

type KingdeeCustomerPull struct {
	CustId         int64  `json:"FCustId"`             // FCustId 金蝶客户ID
	CustomerNumber string `json:"FNumber"`             // FNumber 客户编码
	CustomerName   string `json:"FName"`               // FName 客户名称
	Country        string `json:"FCountry.FDataValue"` // FCountry 国家
	CreateOrgId    int64  `json:"FCreateOrgId"`        // FCreateOrgId 创建组织ID
	UseOrgId       int64  `json:"FUseOrgId"`           // FUseOrgId 使用组织ID
	GroupId        int64  `json:"FGroup"`              // FGroup 客户分组ID
	GroupName      string `json:"FGroup.FName"`        // FName 客户分组名称
	GroupNumber    string `json:"FGroup.FNumber"`      // FNumber 客户分组编码
	CustomerStatus string `json:"FDocumentStatus"`     // FDocumentStatus 单据状态(A:创建,B:审核中,C:已审核,D:重新审核,Z:暂存)
	ForbidStatus   string `json:"FForbidStatus"`       // FForbidStatus 禁用状态(A:启用,B:禁用)
	CreateDate     string `json:"FCreateDate"`         // FCreateDate 金蝶创建日期
	ModifyDate     string `json:"FModifyDate"`         // FModifyDate 金蝶修改日期
}
