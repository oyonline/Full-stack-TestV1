package handler

import (
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"gorm.io/gorm"
)

type Login struct {
	Username string `form:"UserName" json:"username" binding:"required"`
	Password string `form:"Password" json:"password" binding:"required"`
	Code     string `form:"Code" json:"code" binding:"required"`
	UUID     string `form:"UUID" json:"uuid" binding:"required"`
}

type userRoleBinding struct {
	RoleId    int  `gorm:"column:role_id"`
	IsPrimary bool `gorm:"column:is_primary"`
}

func (u *Login) GetUser(tx *gorm.DB) (user SysUser, role SysRole, roles []SysRole, err error) {
	err = tx.Table("sys_user").Where("username = ?  and status = '2'", u.Username).First(&user).Error
	if err != nil {
		log.Errorf("get user error, %s", err.Error())
		return
	}
	_, err = pkg.CompareHashAndPassword(user.Password, u.Password)
	if err != nil {
		log.Errorf("user login error, %s", err.Error())
		return
	}

	bindings := make([]userRoleBinding, 0)
	err = tx.Table("sys_user_role").
		Where("user_id = ?", user.UserId).
		Order("is_primary desc, role_id asc").
		Find(&bindings).Error
	if err != nil {
		log.Errorf("get user roles error, %s", err.Error())
		return
	}
	if len(bindings) == 0 && user.RoleId > 0 {
		bindings = append(bindings, userRoleBinding{
			RoleId:    user.RoleId,
			IsPrimary: true,
		})
	}
	roleIDs := make([]int, 0, len(bindings))
	primaryRoleID := user.RoleId
	for _, binding := range bindings {
		roleIDs = append(roleIDs, binding.RoleId)
		if binding.IsPrimary {
			primaryRoleID = binding.RoleId
		}
	}
	user.PrimaryRoleId = primaryRoleID
	user.RoleIds = roleIDs
	if len(roleIDs) > 0 {
		err = tx.Table("sys_role").Where("role_id IN ?", roleIDs).Find(&roles).Error
		if err != nil {
			log.Errorf("get roles error, %s", err.Error())
			return
		}
	}
	if primaryRoleID == 0 && len(roles) > 0 {
		primaryRoleID = roles[0].RoleId
	}
	for _, currentRole := range roles {
		if currentRole.RoleId == primaryRoleID {
			role = currentRole
			break
		}
	}
	if role.RoleId == 0 {
		err = tx.Table("sys_role").Where("role_id = ? ", primaryRoleID).First(&role).Error
		if err != nil {
			log.Errorf("get primary role error, %s", err.Error())
			return
		}
	}
	return
}
