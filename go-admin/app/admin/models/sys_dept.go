package models

import (
	"go-admin/common/models"
	"go-admin/common/utils"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
)

type CurrentDept struct {
	DeptId             int     `json:"deptId" gorm:"primaryKey;autoIncrement;"` // 部门ID
	DeptName           string  `json:"deptName"  gorm:"size:128;"`              // 部门名称
	DeptPath           string  `json:"deptPath" gorm:"size:255;"`               // 部门路径ID
	LeaderUid          int     `json:"leaderUid" gorm:"size:20"`                // 负责人ID
	Leader             string  `json:"leader" gorm:"size:128;"`                 // 负责人
	ParentDepartmentId string  `json:"parentDepartmentId" gorm:"size:55"`       // 飞书上级部门ID
	OpenDepartmentId   string  `json:"openDepartmentId" gorm:"size:55"`         // 飞书系统部门ID
	DepartmentId       string  `json:"departmentId" gorm:"size:55"`             // 飞书自定义部门ID
	LeaderUserId       string  `json:"leaderUserId" gorm:"size:55"`             // 飞书部门管理ID
	LeaderUser         SysUser `json:"leaderUser" gorm:"foreignKey:LeaderUid;references:UserId"`
}

func (CurrentDept) TableName() string {
	return "sys_dept"
}

type SimpleDept struct {
	DeptId             int    `json:"deptId" gorm:"primaryKey;autoIncrement;"` // 部门ID
	DeptName           string `json:"deptName"  gorm:"size:128;"`              // 部门名称
	DeptPath           string `json:"deptPath" gorm:"size:255;"`               // 部门路径ID
	LeaderUid          int    `json:"leaderUid" gorm:"size:20"`                // 负责人ID
	Leader             string `json:"leader" gorm:"size:128;"`                 // 负责人
	ParentDepartmentId string `json:"parentDepartmentId" gorm:"size:55"`       // 飞书上级部门ID
	OpenDepartmentId   string `json:"openDepartmentId" gorm:"size:55"`         // 飞书系统部门ID
	DepartmentId       string `json:"departmentId" gorm:"size:55"`             // 飞书自定义部门ID
	LeaderUserId       string `json:"leaderUserId" gorm:"size:55"`             // 飞书部门管理ID
}

func (SimpleDept) TableName() string {
	return "sys_dept"
}

type SysDept struct {
	DeptId             int    `json:"deptId" gorm:"primaryKey;autoIncrement;"` //部门编码
	ParentId           int    `json:"parentId" gorm:""`                        //上级部门
	DeptType           int    `json:"deptType" gorm:"size:20"`                 // 部门类型 1业务部门 2职能部门
	DeptPath           string `json:"deptPath" gorm:"size:255;"`               //
	DeptPathName       string `json:"deptPathName" gorm:"size:255;"`           // 部门路径
	DeptName           string `json:"deptName"  gorm:"size:128;"`              //部门名称
	UserNumber         int    `json:"userNumber" gorm:"size:20"`               // 预算编制
	Sort               int    `json:"sort" gorm:"size:4;"`                     //排序
	LeaderUid          int    `json:"leaderUid" gorm:"size:20"`                // 负责人ID
	Leader             string `json:"leader" gorm:"size:128;"`                 //负责人
	Phone              string `json:"phone" gorm:"size:11;"`                   //手机
	Email              string `json:"email" gorm:"size:64;"`                   //邮箱
	Status             int    `json:"status" gorm:"size:4;"`                   //状态
	DeptCode           string `json:"deptCode" gorm:"size:100"`                // 部门编码
	Level              int    `json:"level" gorm:"size:20"`                    // 层级
	ParentDepartmentId string `json:"parentDepartmentId" gorm:"size:55"`       // 飞书上级部门ID
	OpenDepartmentId   string `json:"openDepartmentId" gorm:"size:55"`         // 飞书系统部门ID
	DepartmentId       string `json:"departmentId" gorm:"size:55"`             // 飞书自定义部门ID
	LeaderUserId       string `json:"leaderUserId" gorm:"size:55"`             // 飞书部门管理ID
	MemberCount        int    `json:"memberCount" gorm:"size:20"`              // 实际编制
	models.ControlBy
	models.ModelTime
	DataScope  string    `json:"dataScope" gorm:"-"`
	Params     string    `json:"params" gorm:"-"`
	Children   []SysDept `json:"children" gorm:"-"`
	CreateTime string    `json:"createTime" gorm:"-"`
	UpdateTime string    `json:"updateTime" gorm:"-"`
}

func (*SysDept) TableName() string {
	return "sys_dept"
}

func (e *SysDept) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SysDept) GetId() interface{} {
	return e.DeptId
}

func (model *SysDept) FeishuGenerate(v *larkcontact.Department) {
	model.DeptName = *v.Name
	model.ParentDepartmentId = *v.ParentDepartmentId
	model.OpenDepartmentId = *v.OpenDepartmentId
	model.DepartmentId = *v.DepartmentId
	model.MemberCount = *v.MemberCount
	if v.LeaderUserId != nil {
		model.LeaderUserId = *v.LeaderUserId
	}
	if *v.Status.IsDeleted {
		model.Status = 1
	} else {
		model.Status = 2
	}
	model.Sort = utils.ToInt(*v.Order)
}
