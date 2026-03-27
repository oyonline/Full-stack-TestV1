package dto

import (
	"errors"
	"go-admin/app/admin/models"
	"sort"
	"strings"

	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type SysUserGetPageReq struct {
	dto.Pagination `search:"-"`
	UserId         int    `form:"userId" search:"type:exact;column:user_id;table:sys_user" comment:"用户ID"`
	Username       string `form:"username" search:"type:contains;column:username;table:sys_user" comment:"用户名"`
	NickName       string `form:"nickName" search:"type:contains;column:nick_name;table:sys_user" comment:"昵称"`
	Phone          string `form:"phone" search:"type:contains;column:phone;table:sys_user" comment:"手机号"`
	RoleId         string `form:"roleId" search:"type:exact;column:role_id;table:sys_user" comment:"角色ID"`
	RoleIds        string `form:"roleIds" search:"-" comment:"角色ID数组，逗号分隔"`
	Sex            string `form:"sex" search:"type:exact;column:sex;table:sys_user" comment:"性别"`
	Email          string `form:"email" search:"type:contains;column:email;table:sys_user" comment:"邮箱"`
	PostId         string `form:"postId" search:"type:exact;column:post_id;table:sys_user" comment:"岗位"`
	Status         string `form:"status" search:"type:exact;column:status;table:sys_user" comment:"状态"`
	DeptJoin       `search:"type:left;on:dept_id:dept_id;table:sys_user;join:sys_dept"`
	SysUserOrder
}

type SysUserOrder struct {
	UserIdOrder    string `search:"type:order;column:user_id;table:sys_user" form:"userIdOrder"`
	UsernameOrder  string `search:"type:order;column:username;table:sys_user" form:"usernameOrder"`
	StatusOrder    string `search:"type:order;column:status;table:sys_user" form:"statusOrder"`
	CreatedAtOrder string `search:"type:order;column:created_at;table:sys_user" form:"createdAtOrder"`
}

type DeptJoin struct {
	DeptId string `search:"type:contains;column:dept_path;table:sys_dept" form:"deptId"`
}

func (m *SysUserGetPageReq) GetNeedSearch() interface{} {
	return *m
}

func (m *SysUserGetPageReq) GetRoleIDList() []int {
	values := make([]int, 0)
	if m.RoleId != "" {
		if roleID, err := pkg.StringToInt(m.RoleId); err == nil && roleID > 0 {
			values = append(values, roleID)
		}
	}
	if m.RoleIds != "" {
		for _, part := range strings.Split(m.RoleIds, ",") {
			if roleID, err := pkg.StringToInt(strings.TrimSpace(part)); err == nil && roleID > 0 {
				values = append(values, roleID)
			}
		}
	}
	return uniqueSysUserRoleIDs(values)
}

type ResetSysUserPwdReq struct {
	UserId   int    `json:"userId" comment:"用户ID" vd:"$>0"` // 用户ID
	Password string `json:"password" comment:"密码" vd:"len($)>0"`
	common.ControlBy
}

func (s *ResetSysUserPwdReq) GetId() interface{} {
	return s.UserId
}

func (s *ResetSysUserPwdReq) Generate(model *models.SysUser) {
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Password = s.Password
}

type UpdateSysUserAvatarReq struct {
	UserId int    `json:"userId" comment:"用户ID" vd:"len($)>0"` // 用户ID
	Avatar string `json:"avatar" comment:"头像" vd:"len($)>0"`
	common.ControlBy
}

func (s *UpdateSysUserAvatarReq) GetId() interface{} {
	return s.UserId
}

func (s *UpdateSysUserAvatarReq) Generate(model *models.SysUser) {
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Avatar = s.Avatar
}

type UpdateSysUserProfileReq struct {
	UserId       int    `json:"userId" comment:"用户ID" vd:"$>0"`
	Introduction string `json:"introduction" comment:"个人简介"`
	common.ControlBy
}

func (s *UpdateSysUserProfileReq) GetId() interface{} {
	return s.UserId
}

func (s *UpdateSysUserProfileReq) Generate(model *models.SysUser) {
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Introduction = strings.TrimSpace(s.Introduction)
}

type UpdateSysUserStatusReq struct {
	UserId int    `json:"userId" comment:"用户ID" vd:"$>0"` // 用户ID
	Status string `json:"status" comment:"状态" vd:"len($)>0"`
	common.ControlBy
}

func (s *UpdateSysUserStatusReq) GetId() interface{} {
	return s.UserId
}

func (s *UpdateSysUserStatusReq) Generate(model *models.SysUser) {
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Status = s.Status
}

type SysUserInsertReq struct {
	UserId        int    `json:"userId" comment:"用户ID"` // 用户ID
	Username      string `json:"username" comment:"用户名" vd:"len($)>0"`
	Password      string `json:"password" comment:"密码"`
	NickName      string `json:"nickName" comment:"昵称" vd:"len($)>0"`
	Phone         string `json:"phone" comment:"手机号" vd:"len($)>0"`
	RoleId        int    `json:"roleId" comment:"角色ID（兼容字段）"`
	PrimaryRoleId int    `json:"primaryRoleId" comment:"主角色ID"`
	RoleIds       []int  `json:"roleIds" comment:"角色ID数组"`
	Avatar        string `json:"avatar" comment:"头像"`
	Sex           string `json:"sex" comment:"性别"`
	Email         string `json:"email" comment:"邮箱" vd:"len($)>0,email"`
	DeptId        int    `json:"deptId" comment:"部门" vd:"$>0"`
	PostId        int    `json:"postId" comment:"岗位"`
	Remark        string `json:"remark" comment:"备注"`
	Status        string `json:"status" comment:"状态" vd:"len($)>0" default:"2"`
	common.ControlBy
}

func (s *SysUserInsertReq) Generate(model *models.SysUser) {
	primaryRoleID, roleIDs, _ := s.NormalizeRoles()
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Username = s.Username
	model.Password = s.Password
	model.NickName = s.NickName
	model.Phone = s.Phone
	model.RoleId = primaryRoleID
	model.PrimaryRoleId = primaryRoleID
	model.RoleIds = roleIDs
	model.Avatar = s.Avatar
	model.Sex = s.Sex
	model.Email = s.Email
	model.DeptId = s.DeptId
	model.PostId = s.PostId
	model.Remark = s.Remark
	model.Status = s.Status
	model.CreateBy = s.CreateBy
}

func (s *SysUserInsertReq) GetId() interface{} {
	return s.UserId
}

func (s *SysUserInsertReq) NormalizeRoles() (int, []int, error) {
	return normalizeSysUserRoles(s.PrimaryRoleId, s.RoleId, s.RoleIds)
}

type SysUserUpdateReq struct {
	UserId        int    `json:"userId" comment:"用户ID"` // 用户ID
	Username      string `json:"username" comment:"用户名" vd:"len($)>0"`
	NickName      string `json:"nickName" comment:"昵称" vd:"len($)>0"`
	Phone         string `json:"phone" comment:"手机号" vd:"len($)>0"`
	RoleId        int    `json:"roleId" comment:"角色ID（兼容字段）"`
	PrimaryRoleId int    `json:"primaryRoleId" comment:"主角色ID"`
	RoleIds       []int  `json:"roleIds" comment:"角色ID数组"`
	Avatar        string `json:"avatar" comment:"头像"`
	Sex           string `json:"sex" comment:"性别"`
	Email         string `json:"email" comment:"邮箱" vd:"len($)>0,email"`
	DeptId        int    `json:"deptId" comment:"部门" vd:"$>0"`
	PostId        int    `json:"postId" comment:"岗位"`
	Remark        string `json:"remark" comment:"备注"`
	Status        string `json:"status" comment:"状态" default:"2"`
	common.ControlBy
}

func (s *SysUserUpdateReq) Generate(model *models.SysUser) {
	primaryRoleID, roleIDs, _ := s.NormalizeRoles()
	if s.UserId != 0 {
		model.UserId = s.UserId
	}
	model.Username = s.Username
	model.NickName = s.NickName
	model.Phone = s.Phone
	model.RoleId = primaryRoleID
	model.PrimaryRoleId = primaryRoleID
	model.RoleIds = roleIDs
	model.Avatar = s.Avatar
	model.Sex = s.Sex
	model.Email = s.Email
	model.DeptId = s.DeptId
	model.PostId = s.PostId
	model.Remark = s.Remark
	model.Status = s.Status
}

func (s *SysUserUpdateReq) GetId() interface{} {
	return s.UserId
}

func (s *SysUserUpdateReq) NormalizeRoles() (int, []int, error) {
	return normalizeSysUserRoles(s.PrimaryRoleId, s.RoleId, s.RoleIds)
}

type SysUserById struct {
	dto.ObjectById
	common.ControlBy
}

func (s *SysUserById) GetId() interface{} {
	if len(s.Ids) > 0 {
		s.Ids = append(s.Ids, s.Id)
		return s.Ids
	}
	return s.Id
}

func (s *SysUserById) GenerateM() (common.ActiveRecord, error) {
	return &models.SysUser{}, nil
}

// PassWord 密码
type PassWord struct {
	NewPassword string `json:"newPassword" vd:"len($)>0"`
	OldPassword string `json:"oldPassword" vd:"len($)>0"`
}

func normalizeSysUserRoles(primaryRoleID int, legacyRoleID int, roleIDs []int) (int, []int, error) {
	normalized := append([]int{}, roleIDs...)
	if len(normalized) == 0 && primaryRoleID > 0 {
		normalized = append(normalized, primaryRoleID)
	}
	if len(normalized) == 0 && legacyRoleID > 0 {
		normalized = append(normalized, legacyRoleID)
	}
	normalized = uniqueSysUserRoleIDs(normalized)
	if len(normalized) == 0 {
		return 0, nil, errors.New("至少需要选择一个角色")
	}
	if primaryRoleID <= 0 {
		if legacyRoleID > 0 {
			primaryRoleID = legacyRoleID
		} else {
			primaryRoleID = normalized[0]
		}
	}
	found := false
	for _, roleID := range normalized {
		if roleID == primaryRoleID {
			found = true
			break
		}
	}
	if !found {
		return 0, nil, errors.New("主角色必须属于已选角色")
	}
	return primaryRoleID, normalized, nil
}

func uniqueSysUserRoleIDs(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Ints(result)
	return result
}
