package dto

import (
	"go-admin/app/admin/models"
	common "go-admin/common/models"
	"time"
)

// SysDeptGetPageReq 列表或者搜索使用结构体
type SysDeptGetPageReq struct {
	DeptId   int    `form:"deptId" search:"type:exact;column:dept_id;table:sys_dept" comment:"id"`       //id
	ParentId int    `form:"parentId" search:"type:exact;column:parent_id;table:sys_dept" comment:"上级部门"` //上级部门
	DeptPath string `form:"deptPath" search:"type:exact;column:dept_path;table:sys_dept" comment:""`     //路径
	DeptName string `form:"deptName" search:"type:exact;column:dept_name;table:sys_dept" comment:"部门名称"` //部门名称
	Sort     int    `form:"sort" search:"type:exact;column:sort;table:sys_dept" comment:"排序"`            //排序
	Leader   string `form:"leader" search:"type:exact;column:leader;table:sys_dept" comment:"负责人"`       //负责人
	Phone    string `form:"phone" search:"type:exact;column:phone;table:sys_dept" comment:"手机"`          //手机
	Email    string `form:"email" search:"type:exact;column:email;table:sys_dept" comment:"邮箱"`          //邮箱
	Status   string `form:"status" search:"type:exact;column:status;table:sys_dept" comment:"状态"`        //状态
}

func (m *SysDeptGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type SysDeptInsertReq struct {
	DeptId             int    `uri:"id" comment:"编码"`                                         // 编码
	ParentId           int    `json:"parentId" comment:"上级部门" vd:"?"`                         //上级部门
	DeptPath           string `json:"deptPath" comment:""`                                    //路径
	DeptName           string `json:"deptName" comment:"部门名称" vd:"len($)>0"`                  //部门名称
	Sort               int    `json:"sort" comment:"排序" vd:"?"`                               //排序
	DeptCode           string `json:"deptCode" comment:"部门编码" vd:"len($)>0"`                  // 编码
	LeaderUid          int    `json:"leaderUid" gorm:"size:20"`                               // 负责人ID
	Leader             string `json:"leader" comment:"负责人" vd:"@:len($)>0; msg:'leader不能为空'"` //负责人
	Phone              string `json:"phone" comment:"手机" vd:"?"`                              //手机
	Email              string `json:"email" comment:"邮箱" vd:"?"`                              //邮箱
	Status             int    `json:"status" comment:"状态" vd:"$>0"`                           //状态
	UserNumber         int    `json:"userNumber" comment:"预算编制"`                              // 预算编制
	DeptType           int    `json:"deptType" comment:"部门类型"`                                // 部门类型 1业务部门 2职能部门
	ParentDepartmentId string `json:"parentDepartmentId" comment:"飞书上级部门ID"`                  // 飞书上级部门ID
	OpenDepartmentId   string `json:"openDepartmentId" comment:"飞书系统部门ID"`                    // 飞书系统部门ID
	DepartmentId       string `json:"departmentId" comment:"飞书自定义部门ID"`                       // 飞书自定义部门ID
	LeaderUserId       string `json:"leaderUserId" comment:"飞书部门管理ID"`                        // 飞书部门管理ID
	MemberCount        int    `json:"memberCount" comment:"实际编制"`                             // 实际编制
	common.ControlBy
}

func (s *SysDeptInsertReq) Generate(model *models.SysDept) {
	if s.DeptId != 0 {
		model.DeptId = s.DeptId
	}
	model.DeptName = s.DeptName
	model.ParentId = s.ParentId
	model.DeptPath = s.DeptPath
	model.Sort = s.Sort
	model.Leader = s.Leader
	model.Phone = s.Phone
	model.Email = s.Email
	model.Status = s.Status
	model.LeaderUid = s.LeaderUid
	model.DeptCode = s.DeptCode
	if s.UserNumber == 0 {
		s.UserNumber = 1
	}
	if s.DeptType == 0 {
		s.DeptType = 2
	}
	model.DeptType = s.DeptType
	model.ParentDepartmentId = s.ParentDepartmentId
	model.OpenDepartmentId = s.OpenDepartmentId
	model.DepartmentId = s.DepartmentId
	model.LeaderUserId = s.LeaderUserId
	model.MemberCount = s.MemberCount
	model.CreateBy = s.CreateBy
	model.UpdateBy = s.UpdateBy
}

// GetId 获取数据对应的ID
func (s *SysDeptInsertReq) GetId() interface{} {
	return s.DeptId
}

type SysDeptUpdateReq struct {
	DeptId     int    `uri:"id" comment:"编码"`                        // 编码
	DeptCode   string `json:"deptCode" comment:"部门编码" vd:"len($)>0"` // 部门编码
	Status     int    `json:"status" comment:"状态" vd:"$>0"`          // 状态
	UserNumber int    `json:"userNumber" comment:"预算编制"`             // 预算编制
	DeptType   int    `json:"deptType" comment:"部门类型"`               // 部门类型 1业务部门 2职能部门
	common.ControlBy
}

// Generate 结构体数据转化 从 SysDeptControl 至 SysDept 对应的模型
func (s *SysDeptUpdateReq) Generate(model *models.SysDept) {
	if s.DeptId != 0 {
		model.DeptId = s.DeptId
	}
	model.Status = s.Status
	model.DeptCode = s.DeptCode
	if s.UserNumber > 0 && s.UserNumber != model.UserNumber {
		model.UserNumber = s.UserNumber
	}
	if s.DeptType != 0 && s.DeptType != model.DeptType {
		model.DeptType = s.DeptType
	}
	model.UpdateBy = s.UpdateBy
}

// GetId 获取数据对应的ID
func (s *SysDeptUpdateReq) GetId() interface{} {
	return s.DeptId
}

type SysDeptGetReq struct {
	Id int `uri:"id"`
}

func (s *SysDeptGetReq) GetId() interface{} {
	return s.Id
}

type SysDeptDeleteReq struct {
	Ids []int `json:"ids"`
}

func (s *SysDeptDeleteReq) GetId() interface{} {
	return s.Ids
}

type DeptLabel struct {
	Id       int         `gorm:"-" json:"id"`
	Label    string      `gorm:"-" json:"label"`
	Children []DeptLabel `gorm:"-" json:"children"`
}

type DepartmentBatch struct {
	OpenDepartmentIds []string `form:"openDepartmentIds" comment:"飞书上级部门ID"`
	ParentId          int      `form:"parentId" comment:"上级部门"`
	common.ControlBy
}

func (s *DepartmentBatch) GetId() interface{} {
	return s.OpenDepartmentIds
}

type SysDeptExport struct {
	DeptPathName string    `json:"deptPathName" excel:"部门路径,sort:1,width:50,required:true"`
	DeptName     string    `json:"deptName" excel:"部门名称,sort:1,width:20,required:true"`
	DeptCode     string    `json:"deptCode" excel:"部门编码,sort:2,width:20,required:true"`
	DeptType     int       `json:"deptType" excel:"部门类型,sort:3,width:20,converter:1=业务部门|2=职能部门,required:true"`
	Leader       string    `json:"leader" excel:"负责人,sort:4,width:20,required:true"`
	UserNumber   int       `json:"userNumber" excel:"预算编制,sort:5,width:20,required:true"`
	MemberCount  int       `json:"memberCount" excel:"实际编制,sort:6,width:20,required:true"`
	Level        int       `json:"level" excel:"层级,sort:7,width:20,converter:1=L1|2=L2|3=L3|4=L4|5=L5,required:true"`
	UpdatedAt    time.Time `json:"updatedAt" excel:"更新时间,sort:8,width:20,required:true"`
	Status       int       `json:"status" excel:"状态,sort:9,width:20,converter:1=禁用|2=启用,required:true"`
}
