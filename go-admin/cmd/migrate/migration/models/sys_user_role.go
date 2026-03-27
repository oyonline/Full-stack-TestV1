package models

import "time"

type SysUserRole struct {
	UserId    int       `json:"userId" gorm:"primaryKey;autoIncrement:false;comment:用户ID"`
	RoleId    int       `json:"roleId" gorm:"primaryKey;autoIncrement:false;comment:角色ID"`
	IsPrimary bool      `json:"isPrimary" gorm:"column:is_primary;default:false;comment:是否主角色"`
	CreatedAt time.Time `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"comment:最后更新时间"`
}

func (*SysUserRole) TableName() string {
	return "sys_user_role"
}
