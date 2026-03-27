package dto

import (
	"time"

	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"mime/multipart"
)

type CostCenterInfoGetPageReq struct {
	dto.Pagination `search:"-"`
	CostCenterName string `form:"costCenterName"  search:"type:contains;column:cost_center_name;table:cost_center_info" comment:"成本中心名称"`
	CostCenterCode string `form:"costCenterCode"  search:"type:contains;column:cost_center_code;table:cost_center_info" comment:"成本中心编码"`
	Status         int    `form:"status"  search:"type:exact;column:status;table:cost_center_info" comment:"状态(1=停用|2=启用)"`
	DeptId         int64  `form:"deptId"  gorm:"-" search:"-"`
	GroupId        int64  `form:"groupId"  gorm:"-" search:"-"`
	CostCenterInfoOrder
}

type CostCenterInfoOrder struct {
	Id string `form:"idOrder"  search:"type:order;column:id;table:cost_center_info"`
	/*CostCenterName string `form:"costCenterNameOrder"  search:"type:order;column:cost_center_name;table:cost_center_info"`
	  CostCenterCode string `form:"costCenterCodeOrder"  search:"type:order;column:cost_center_code;table:cost_center_info"`
	  CostCenterType string `form:"costCenterTypeOrder"  search:"type:order;column:cost_center_type;table:cost_center_info"`
	  DeptId string `form:"deptIdOrder"  search:"type:order;column:dept_id;table:cost_center_info"`
	  Description string `form:"descriptionOrder"  search:"type:order;column:description;table:cost_center_info"`
	  Status string `form:"statusOrder"  search:"type:order;column:status;table:cost_center_info"`
	  EffectiveDate string `form:"effectiveDateOrder"  search:"type:order;column:effective_date;table:cost_center_info"`
	  CreateBy string `form:"createByOrder"  search:"type:order;column:create_by;table:cost_center_info"`
	  CreatedAt string `form:"createdAtOrder"  search:"type:order;column:created_at;table:cost_center_info"`
	  UpdateBy string `form:"updateByOrder"  search:"type:order;column:update_by;table:cost_center_info"`
	  UpdatedAt string `form:"updatedAtOrder"  search:"type:order;column:updated_at;table:cost_center_info"`
	  DeletedAt string `form:"deletedAtOrder"  search:"type:order;column:deleted_at;table:cost_center_info"`*/
}

func (m *CostCenterInfoGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type CostCenterInfoInsertReq struct {
	Id               int64  `json:"-" comment:"主键"` // 主键
	CostCenterName   string `json:"costCenterName" comment:"成本中心名称" vd:"@:len($)>0; msg:'成本中心名称不能为空'"`
	CostCenterCode   string `json:"costCenterCode" comment:"成本中心编码" vd:"@:len($)>0; msg:'成本中心编码不能为空'"`
	CostCenterType   int    `json:"costCenterType" comment:"成本中心类型(1=事业部|2=成本中心|3=费用类别)"`
	DeptId           int64  `json:"deptId" comment:"上级部门" vd:"?"`
	Description      string `json:"description" comment:"描述备注"`
	Status           int    `json:"status" comment:"状态(1=停用|2=启用)"`
	EffectiveDateStr string `json:"effectiveDateStr" comment:"生效日期" vd:"@:len($)>0; msg:'生效日期不能为空'"`
	common.ControlBy
	GroupNameList []models.GroupNameInfoData `json:"groupNameList" gorm:"-"`
}

func (s *CostCenterInfoInsertReq) Generate(model *models.CostCenterInfo) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostCenterName = s.CostCenterName
	model.CostCenterCode = s.CostCenterCode
	model.DeptId = s.DeptId
	model.Description = s.Description
	model.Status = s.Status
	// 添加这而，需要记录是被谁创建的
	model.CreateBy = s.CreateBy
	model.UpdateBy = s.CreateBy
	model.GroupNameList = s.GroupNameList
	dateStr, err := time.Parse(time.DateOnly, s.EffectiveDateStr)
	if err != nil {
		panic(err)
	}
	model.EffectiveDate = dateStr
	model.CostCenterType = 2
}

func (s *CostCenterInfoInsertReq) GetId() interface{} {
	return s.Id
}

type CostCenterInfoUpdateReq struct {
	Id               int64  `uri:"id" comment:"主键"` // 主键
	CostCenterName   string `json:"costCenterName" comment:"成本中心名称" vd:"@:len($)>0; msg:'成本中心名称不能为空'"`
	CostCenterCode   string `json:"costCenterCode" comment:"成本中心编码" vd:"@:len($)>0; msg:'成本中心编码不能为空'"`
	CostCenterType   int    `json:"costCenterType" comment:"成本中心类型(1=事业部|2=成本中心|3=费用类别)"`
	DeptId           int64  `json:"deptId" comment:"上级部门" vd:"?"`
	Description      string `json:"description" comment:"描述备注"`
	Status           int    `json:"status" comment:"状态(1=停用|2=启用)"`
	EffectiveDateStr string `json:"effectiveDateStr" comment:"生效日期" vd:"@:len($)>0; msg:'生效日期不能为空'"`
	common.ControlBy
	GroupNameList []models.GroupNameInfoData `json:"groupNameList" gorm:"-"`
}

func (s *CostCenterInfoUpdateReq) Generate(model *models.CostCenterInfo) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.CostCenterName = s.CostCenterName
	model.CostCenterCode = s.CostCenterCode
	model.CostCenterType = 2
	model.DeptId = s.DeptId
	model.Description = s.Description
	model.Status = s.Status
	dateStr, err := time.Parse(time.DateOnly, s.EffectiveDateStr)
	if err != nil {
		panic(err)
	}
	model.EffectiveDate = dateStr
	model.UpdateBy = s.UpdateBy // 添加这而，需要记录是被谁更新的
	model.GroupNameList = s.GroupNameList
}

func (s *CostCenterInfoUpdateReq) GetId() interface{} {
	return s.Id
}

// CostCenterInfoGetReq 功能获取请求参数
type CostCenterInfoGetReq struct {
	Id int64 `uri:"id"`
}

func (s *CostCenterInfoGetReq) GetId() interface{} {
	return s.Id
}

// CostCenterInfoDeleteReq 功能删除请求参数
type CostCenterInfoDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *CostCenterInfoDeleteReq) GetId() interface{} {
	return s.Ids
}

type CostCenterInfoImportReq struct {
	File *multipart.FileHeader `form:"file" comment:"文件"`
}

/*type CostCenterInfoExport struct {
	Id             int64     `json:"-" comment:"主键"` // 主键
	CostCenterName string    `json:"costCenterName" comment:"成本中心名称" excel:"成本中心名称,sort:1,width:20,required:true"`
	CostCenterCode string    `json:"costCenterCode" comment:"成本中心编码" excel:"成本中心编码,sort:2,width:20,required:true"`
	DeptPathName   string    `json:"deptPathName" comment:"上级部门" excel:"上级部门,sort:3,width:20,operation:export"`
	Description    string    `json:"description" comment:"描述备注" excel:"描述备注,sort:4,width:20"`
	Status         int       `json:"status" comment:"状态(1=停用|2=启用)" excel:"状态(1=停用|2=启用),sort:5,width:20,required:true"`
	EffectiveDate  time.Time `json:"effectiveDate" comment:"生效日期" excel:"生效日期,sort:6,width:20,required:true"`
	GroupNameStr   string    `json:"groupNameStr" comment:"客户分组名称" excel:"客户分组名称,sort:7,width:20"`
}

type ChangeDetailsCompare struct {
	CostCenterName string    `json:"costCenterName" comment:"成本中心名称"`
	CostCenterCode string    `json:"costCenterCode" comment:"成本中心编码"`
	DeptPathName   string    `json:"deptPathName" comment:"上级部门"`
	DeptId         int64     `json:"deptId" comment:"上级部门ID"`
	Description    string    `json:"description" comment:"描述备注"`
	Status         int       `json:"status" dict:"1=停用|2=启用" comment:"状态"`
	EffectiveDate  time.Time `json:"effectiveDate" comment:"生效日期"`
	GroupNameStr   string    `json:"groupNameStr" comment:"客户分组名称"`
	GroupIds       string    `json:"groupIds" comment:"客户分组ID"`
	Id             int64     `json:"id"` // 主键
}*/
