package service

import (
	"errors"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"sort"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

type SysUser struct {
	service.Service
}

// GetPage 获取SysUser列表
func (e *SysUser) GetPage(c *dto.SysUserGetPageReq, p *actions.DataPermission, list *[]models.SysUser, count *int64) error {
	var err error
	var data models.SysUser

	db := e.Orm.Debug().Model(&data).Preload("Dept").
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		)
	if roleIDs := c.GetRoleIDList(); len(roleIDs) > 0 {
		subQuery := e.Orm.Table("sys_user_role").Select("distinct user_id").Where("role_id IN ?", roleIDs)
		db = db.Where("sys_user.user_id IN (?)", subQuery)
	}
	err = db.Find(list).Limit(-1).Offset(-1).Count(count).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if err = e.hydrateUsers(list); err != nil {
		e.Log.Errorf("hydrate user roles error: %s", err)
		return err
	}
	return nil
}

// Get 获取SysUser对象
func (e *SysUser) Get(d *dto.SysUserById, p *actions.DataPermission, model *models.SysUser) error {
	var data models.SysUser

	err := e.Orm.Model(&data).Debug().
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if err = e.hydrateUser(model); err != nil {
		e.Log.Errorf("hydrate user roles error: %s", err)
		return err
	}
	return nil
}

// Insert 创建SysUser对象
func (e *SysUser) Insert(c *dto.SysUserInsertReq) error {
	primaryRoleID, roleIDs, err := c.NormalizeRoles()
	if err != nil {
		e.Log.Errorf("normalize role selection error: %s", err)
		return err
	}
	var data models.SysUser
	var i int64
	err = e.Orm.Model(&data).Where("username = ?", c.Username).Count(&i).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if i > 0 {
		err := errors.New("用户名已存在！")
		e.Log.Errorf("db error: %s", err)
		return err
	}
	c.Generate(&data)
	data.RoleId = primaryRoleID
	data.PrimaryRoleId = primaryRoleID
	data.RoleIds = roleIDs
	return e.Orm.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&data).Error; err != nil {
			e.Log.Errorf("db error: %s", err)
			return err
		}
		if err = e.syncUserRoles(tx, data.UserId, primaryRoleID, roleIDs); err != nil {
			e.Log.Errorf("sync user roles error: %s", err)
			return err
		}
		c.UserId = data.UserId
		c.RoleId = primaryRoleID
		c.PrimaryRoleId = primaryRoleID
		c.RoleIds = roleIDs
		return nil
	})
}

// Update 修改SysUser对象
func (e *SysUser) Update(c *dto.SysUserUpdateReq, p *actions.DataPermission) error {
	primaryRoleID, roleIDs, err := c.NormalizeRoles()
	if err != nil {
		e.Log.Errorf("normalize role selection error: %s", err)
		return err
	}
	var model models.SysUser
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateSysUser error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")

	}
	c.Generate(&model)
	model.RoleId = primaryRoleID
	model.PrimaryRoleId = primaryRoleID
	model.RoleIds = roleIDs
	return e.Orm.Transaction(func(tx *gorm.DB) error {
		update := tx.Model(&model).Where("user_id = ?", &model.UserId).Omit("password", "salt", "roles").Updates(&model)
		if err = update.Error; err != nil {
			e.Log.Errorf("db error: %s", err)
			return err
		}
		if update.RowsAffected == 0 {
			err = errors.New("update userinfo error")
			log.Warnf("db update error")
			return err
		}
		if err = e.syncUserRoles(tx, model.UserId, primaryRoleID, roleIDs); err != nil {
			e.Log.Errorf("sync user roles error: %s", err)
			return err
		}
		c.RoleId = primaryRoleID
		c.PrimaryRoleId = primaryRoleID
		c.RoleIds = roleIDs
		return nil
	})
}

// UpdateAvatar 更新用户头像
func (e *SysUser) UpdateAvatar(c *dto.UpdateSysUserAvatarReq, p *actions.DataPermission) error {
	var err error
	var model models.SysUser
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateSysUser error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")

	}
	err = e.Orm.Table(model.TableName()).Where("user_id =? ", c.UserId).Updates(c).Error
	if err != nil {
		e.Log.Errorf("Service UpdateSysUser error: %s", err)
		return err
	}
	return nil
}

// UpdateProfile 更新个人资料
func (e *SysUser) UpdateProfile(c *dto.UpdateSysUserProfileReq, p *actions.DataPermission) error {
	var err error
	var model models.SysUser
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateProfile error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	c.Generate(&model)
	err = e.Orm.Table(model.TableName()).
		Where("user_id = ?", c.UserId).
		Updates(map[string]interface{}{"introduction": model.Introduction}).Error
	if err != nil {
		e.Log.Errorf("Service UpdateProfile error: %s", err)
		return err
	}
	return nil
}

// UpdateStatus 更新用户状态
func (e *SysUser) UpdateStatus(c *dto.UpdateSysUserStatusReq, p *actions.DataPermission) error {
	var err error
	var model models.SysUser
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateSysUser error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")

	}
	err = e.Orm.Table(model.TableName()).Where("user_id =? ", c.UserId).Updates(c).Error
	if err != nil {
		e.Log.Errorf("Service UpdateSysUser error: %s", err)
		return err
	}
	return nil
}

// ResetPwd 重置用户密码
func (e *SysUser) ResetPwd(c *dto.ResetSysUserPwdReq, p *actions.DataPermission) error {
	var err error
	var model models.SysUser
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("At Service ResetSysUserPwd error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	var hash []byte
	if hash, err = bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost); err != nil {
		e.Log.Errorf("generate password hash error: %s", err)
		return err
	}
	err = e.Orm.Model(&models.SysUser{}).
		Where("user_id = ?", c.UserId).
		Updates(map[string]interface{}{
			"password":  string(hash),
			"salt":      "",
			"update_by": c.UpdateBy,
		}).Error
	if err != nil {
		e.Log.Errorf("At Service ResetSysUserPwd error: %s", err)
		return err
	}
	return nil
}

// Remove 删除SysUser
func (e *SysUser) Remove(c *dto.SysUserById, p *actions.DataPermission) error {
	var data models.SysUser
	return e.Orm.Transaction(func(tx *gorm.DB) error {
		db := tx.Model(&data).
			Scopes(
				actions.Permission(data.TableName(), p),
			).Delete(&data, c.GetId())
		if err := db.Error; err != nil {
			e.Log.Errorf("Error found in RemoveSysUser : %s", err)
			return err
		}
		if db.RowsAffected == 0 {
			return errors.New("无权删除该数据")
		}
		if err := tx.Unscoped().Where("user_id IN ?", c.GetId()).Delete(&models.SysUserRole{}).Error; err != nil {
			e.Log.Errorf("delete sys_user_role error: %s", err)
			return err
		}
		return nil
	})
}

// UpdatePwd 修改SysUser对象密码
func (e *SysUser) UpdatePwd(id int, oldPassword, newPassword string, p *actions.DataPermission) error {
	var err error

	if newPassword == "" {
		return nil
	}
	c := &models.SysUser{}

	err = e.Orm.Model(c).
		Scopes(
			actions.Permission(c.TableName(), p),
		).Select("UserId", "Password", "Salt").
		First(c, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("无权更新该数据")
		}
		e.Log.Errorf("db error: %s", err)
		return err
	}
	var ok bool
	ok, err = pkg.CompareHashAndPassword(c.Password, oldPassword)
	if err != nil {
		e.Log.Errorf("CompareHashAndPassword error, %s", err.Error())
		return err
	}
	if !ok {
		err = errors.New("旧密码错误")
		e.Log.Warnf("user[%d] %s", id, err.Error())
		return err
	}
	var hash []byte
	if hash, err = bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost); err != nil {
		e.Log.Errorf("generate password hash error: %s", err)
		return err
	}
	db := e.Orm.Model(&models.SysUser{}).Where("user_id = ?", id).
		Updates(map[string]interface{}{
			"password": string(hash),
			"salt":     "",
		})
	if err = db.Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("set password error")
		log.Warnf("db update error")
		return err
	}
	return nil
}

func (e *SysUser) GetProfile(c *dto.SysUserById, user *models.SysUser, roles *[]models.SysRole, posts *[]models.SysPost) error {
	err := e.Orm.Preload("Dept").First(user, c.GetId()).Error
	if err != nil {
		return err
	}
	if err = e.hydrateUser(user); err != nil {
		return err
	}
	if len(user.RoleIds) == 0 && user.RoleId > 0 {
		user.RoleIds = []int{user.RoleId}
	}
	err = e.Orm.Find(roles, user.RoleIds).Error
	if err != nil {
		return err
	}
	err = e.Orm.Find(posts, user.PostIds).Error
	if err != nil {
		return err
	}

	return nil
}

type userRoleBinding struct {
	UserId     int    `gorm:"column:user_id"`
	RoleId     int    `gorm:"column:role_id"`
	IsPrimary  bool   `gorm:"column:is_primary"`
	RoleName   string `gorm:"column:role_name"`
	RoleKey    string `gorm:"column:role_key"`
	RoleStatus string `gorm:"column:status"`
}

func (e *SysUser) hydrateUsers(users *[]models.SysUser) error {
	if users == nil || len(*users) == 0 {
		return nil
	}
	userIDs := make([]int, 0, len(*users))
	for _, item := range *users {
		userIDs = append(userIDs, item.UserId)
	}
	roleMap, err := e.getUserRoleMap(userIDs)
	if err != nil {
		return err
	}
	for index := range *users {
		e.applyUserRoles(&(*users)[index], roleMap[(*users)[index].UserId])
	}
	return nil
}

func (e *SysUser) hydrateUser(user *models.SysUser) error {
	if user == nil || user.UserId == 0 {
		return nil
	}
	roleMap, err := e.getUserRoleMap([]int{user.UserId})
	if err != nil {
		return err
	}
	e.applyUserRoles(user, roleMap[user.UserId])
	return nil
}

func (e *SysUser) getUserRoleMap(userIDs []int) (map[int][]userRoleBinding, error) {
	result := make(map[int][]userRoleBinding)
	if len(userIDs) == 0 {
		return result, nil
	}
	rows := make([]userRoleBinding, 0)
	err := e.Orm.Table("sys_user_role ur").
		Select("ur.user_id, ur.role_id, ur.is_primary, r.role_name, r.role_key, r.status").
		Joins("left join sys_role r on r.role_id = ur.role_id").
		Where("ur.user_id IN ?", userIDs).
		Order("ur.is_primary desc, ur.role_id asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserId] = append(result[row.UserId], row)
	}
	return result, nil
}

func (e *SysUser) applyUserRoles(user *models.SysUser, bindings []userRoleBinding) {
	user.PrimaryRoleId = user.RoleId
	user.RoleIds = []int{}
	user.Roles = []models.SysRole{}
	if len(bindings) == 0 {
		if user.RoleId > 0 {
			user.RoleIds = []int{user.RoleId}
			user.PrimaryRoleId = user.RoleId
		}
		return
	}
	roleIDs := make([]int, 0, len(bindings))
	roles := make([]models.SysRole, 0, len(bindings))
	primaryRoleID := 0
	for _, binding := range bindings {
		roleIDs = append(roleIDs, binding.RoleId)
		roles = append(roles, models.SysRole{
			RoleId:   binding.RoleId,
			RoleName: binding.RoleName,
			RoleKey:  binding.RoleKey,
			Status:   binding.RoleStatus,
		})
		if binding.IsPrimary {
			primaryRoleID = binding.RoleId
		}
	}
	if primaryRoleID == 0 {
		primaryRoleID = roleIDs[0]
	}
	sort.Ints(roleIDs)
	user.RoleId = primaryRoleID
	user.PrimaryRoleId = primaryRoleID
	user.RoleIds = roleIDs
	user.Roles = roles
}

func (e *SysUser) syncUserRoles(tx *gorm.DB, userID int, primaryRoleID int, roleIDs []int) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}
	if len(roleIDs) == 0 {
		return errors.New("至少需要一个角色")
	}
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.SysUserRole{}).Error; err != nil {
		return err
	}
	relations := make([]models.SysUserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		relations = append(relations, models.SysUserRole{
			UserId:    userID,
			RoleId:    roleID,
			IsPrimary: roleID == primaryRoleID,
		})
	}
	return tx.Create(&relations).Error
}
