package models

import (
	"encoding/json"
	"go-admin/common/models"
	"go-admin/common/utils/structsUtils"
	"time"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SysUserDept struct {
	Id               int    `json:"id" gorm:"id" gorm:"primaryKey;autoIncrement;"`
	DeptId           int    `json:"deptId" gorm:"column:dept_id"`
	UserId           int    `json:"userId" gorm:"column:user_id"`
	OpenDepartmentId string `json:"openDepartmentId" gorm:"column:open_department_id"`
}

func (SysUserDept) TableName() string {
	return "sys_user_depts"
}

type CurrentUser struct {
	UserId   int         `gorm:"primaryKey;autoIncrement;comment:编码"  json:"userId"`
	NickName string      `json:"nickName" gorm:"size:128;comment:昵称"`
	OpenId   string      `json:"openId" gorm:"size:55;comment:飞书用户应用ID"`
	CnName   string      `json:"cnName" gorm:"size:25;comment:飞书中文名"`
	Phone    string      `json:"phone" gorm:"size:20;comment:手机号"`
	DeptId   int         `json:"deptId" gorm:"size:20;comment:部门"`
	JobTitle string      `json:"jobTitle" gorm:"size:55;comment:飞书用户职务"`
	MainDept CurrentDept `json:"mainDept" gorm:"foreignKey:DeptId;references:DeptId"`
}

func (CurrentUser) TableName() string {
	return "sys_user"
}

type SimpleUser struct {
	UserId   int    `gorm:"primaryKey;autoIncrement;comment:编码"  json:"userId"`
	NickName string `json:"nickName" gorm:"size:128;comment:昵称"`
	OpenId   string `json:"openId" gorm:"size:55;comment:飞书用户应用ID"`
	CnName   string `json:"cnName" gorm:"size:25;comment:飞书中文名"`
	Phone    string `json:"phone" gorm:"size:20;comment:手机号"`
	DeptId   int    `json:"deptId" gorm:"size:20;comment:部门"`
	JobTitle string `json:"jobTitle" gorm:"size:55;comment:飞书用户职务"`
}

func (SimpleUser) TableName() string {
	return "sys_user"
}

type SysUser struct {
	UserId            int       `gorm:"primaryKey;autoIncrement;comment:编码"  json:"userId"`
	Username          string    `json:"username" gorm:"size:64;comment:用户名（飞书用户ID）"`
	Password          string    `json:"-" gorm:"size:128;comment:密码"`
	NickName          string    `json:"nickName" gorm:"size:128;comment:昵称"`
	Phone             string    `json:"phone" gorm:"size:11;comment:手机号"`
	RoleId            int       `json:"roleId" gorm:"size:20;comment:角色ID"`
	PrimaryRoleId     int       `json:"primaryRoleId" gorm:"-"`
	Salt              string    `json:"-" gorm:"size:255;comment:加盐"`
	Avatar            string    `json:"avatar" gorm:"size:255;comment:头像"`
	Sex               string    `json:"sex" gorm:"size:255;comment:性别"`
	Email             string    `json:"email" gorm:"size:128;comment:邮箱"`
	DeptId            int       `json:"deptId" gorm:"size:20;comment:部门"`
	PostId            int       `json:"postId" gorm:"size:20;comment:岗位"`
	Introduction      string    `json:"introduction" gorm:"size:255;comment:个人简介"`
	Remark            string    `json:"remark" gorm:"size:255;comment:备注"`
	Status            string    `json:"status" gorm:"size:4;comment:状态"`
	OpenId            string    `json:"openId" gorm:"size:55;comment:飞书用户应用ID"`
	JobTitle          string    `json:"jobTitle" gorm:"size:55;comment:飞书用户职务"`
	OpenDepartmentId  string    `json:"openDepartmentId" gorm:"size:55;comment:飞书系统部门ID"`
	OpenDepartmentIds string    `json:"openDepartmentIds" gorm:"size:255;comment:飞书系统多部门ID"`
	CnName            string    `json:"cnName" gorm:"size:25;comment:飞书中文名"`
	DeptIds           []int     `json:"deptIds" gorm:"-"`
	PostIds           []int     `json:"postIds" gorm:"-"`
	RoleIds           []int     `json:"roleIds" gorm:"-"`
	Roles             []SysRole `json:"roles" gorm:"-"`
	Dept              *SysDept  `json:"dept"`
	CreateTime        string    `json:"createTime" gorm:"-"`
	UpdateTime        string    `json:"updateTime" gorm:"-"`
	models.ControlBy
	models.ModelTime
}

func (*SysUser) TableName() string {
	return "sys_user"
}

func (e *SysUser) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SysUser) GetId() interface{} {
	return e.UserId
}

// Encrypt 加密
func (e *SysUser) Encrypt() (err error) {
	if e.Password == "" {
		return
	}

	var hash []byte
	if hash, err = bcrypt.GenerateFromPassword([]byte(e.Password), bcrypt.DefaultCost); err != nil {
		return
	} else {
		e.Password = string(hash)
		return
	}
}

func (e *SysUser) BeforeCreate(_ *gorm.DB) error {
	return e.Encrypt()
}

func (e *SysUser) BeforeUpdate(_ *gorm.DB) error {
	var err error
	if e.Password != "" {
		err = e.Encrypt()
	}
	return err
}

func (e *SysUser) AfterFind(_ *gorm.DB) error {
	e.DeptIds = []int{e.DeptId}
	e.PostIds = []int{e.PostId}
	if len(e.Roles) > 0 {
		e.RoleIds, _ = structsUtils.StructFieldValues[SysRole, int](&e.Roles, "DeptId")
	} else {
		e.RoleIds = []int{e.RoleId}
	}
	e.CreateTime = e.CreatedAt.Format(time.DateTime)
	e.UpdateTime = e.UpdatedAt.Format(time.DateTime)
	return nil
}

func (model *SysUser) FeishuGenerate(s *larkcontact.User) {
	model.Username = *s.UserId
	model.CnName = *s.Name
	model.NickName = *s.EnName
	model.Phone = *s.Mobile
	model.Avatar = *s.Avatar.AvatarOrigin
	if *s.Gender == 1 {
		model.Sex = "0"
	} else if *s.Gender == 2 {
		model.Sex = "1"
	} else {
		model.Sex = "2"
	}
	model.Email = *s.Email
	if *s.Status.IsFrozen {
		model.Status = "1"
	} else {
		model.Status = "2"
	}
	model.OpenId = *s.OpenId
	model.JobTitle = *s.JobTitle
	model.OpenDepartmentId = s.DepartmentIds[0]
	openDepartmentIds, _ := json.Marshal(s.DepartmentIds)
	model.OpenDepartmentIds = string(openDepartmentIds)
}
