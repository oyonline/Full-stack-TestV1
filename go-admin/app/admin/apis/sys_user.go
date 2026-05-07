package apis

import (
	"github.com/gin-gonic/gin/binding"
	"go-admin/app/admin/models"
	"go-admin/common/authctx"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"
	"github.com/google/uuid"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

type SysUser struct {
	api.Api
}

// GetPage
// @Summary 列表用户信息数据
// @Description 获取JSON
// @Tags 用户
// @Param username query string false "username"
// @Success 200 {string} {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-user [get]
// @Security Bearer
func (e SysUser) GetPage(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	//数据权限检查
	p := actions.GetPermissionFromContext(c)

	list := make([]models.SysUser, 0)
	var count int64

	err = s.GetPage(&req, p, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get
// @Summary 获取用户
// @Description 获取JSON
// @Tags 用户
// @Param userId path int true "用户编码"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-user/{userId} [get]
// @Security Bearer
func (e SysUser) Get(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserById{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var object models.SysUser
	//数据权限检查
	p := actions.GetPermissionFromContext(c)
	err = s.Get(&req, p, &object)
	if err != nil {
		e.Error(http.StatusUnprocessableEntity, err, "查询失败")
		return
	}
	e.OK(object, "查询成功")
}

// Insert
// @Summary 创建用户
// @Description 获取JSON
// @Tags 用户
// @Accept  application/json
// @Product application/json
// @Param data body dto.SysUserInsertReq true "用户数据"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-user [post]
// @Security Bearer
func (e SysUser) Insert(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserInsertReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	// 设置创建人
	req.SetCreateBy(user.GetUserId(c))
	err = s.Insert(&req)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	middleware.AuditLogCreate(c,
		"用户管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategoryUser,
			ID:    req.UserId,
			Label: req.Username,
		},
		map[string]interface{}{
			"username":      req.Username,
			"nickName":      req.NickName,
			"deptId":        req.DeptId,
			"primaryRoleId": req.PrimaryRoleId,
			"postId":        req.PostId,
		},
		"admin.sysUser.insert",
	)

	e.OK(req.GetId(), "创建成功")
}

// Update
// @Summary 修改用户数据
// @Description 获取JSON
// @Tags 用户
// @Accept  application/json
// @Product application/json
// @Param data body dto.SysUserUpdateReq true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-user/{userId} [put]
// @Security Bearer
func (e SysUser) Update(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserUpdateReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	req.SetUpdateBy(user.GetUserId(c))

	//数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.Update(&req, p)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	middleware.AuditLogUpdate(c,
		"用户管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategoryUser,
			ID:    req.UserId,
			Label: req.Username,
		},
		nil,
		map[string]interface{}{
			"username":      req.Username,
			"nickName":      req.NickName,
			"deptId":        req.DeptId,
			"primaryRoleId": req.PrimaryRoleId,
			"postId":        req.PostId,
		},
		"admin.sysUser.update",
	)
	e.OK(req.GetId(), "更新成功")
}

// Delete
// @Summary 删除用户数据
// @Description 删除数据
// @Tags 用户
// @Param userId path int true "userId"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-user/{userId} [delete]
// @Security Bearer
func (e SysUser) Delete(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserById{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	// 设置编辑人
	req.SetUpdateBy(user.GetUserId(c))

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.Remove(&req, p)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	middleware.AuditLogDelete(c,
		"用户管理",
		middleware.AuditTarget{
			Type: middleware.AuditCategoryUser,
			ID:   req.Ids,
		},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.sysUser.delete",
	)
	e.OK(req.GetId(), "删除成功")
}

// InsetAvatar
// @Summary 修改头像
// @Description 获取JSON
// @Tags 个人中心
// @Accept multipart/form-data
// @Param file formData file true "file"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/avatar [post]
// @Security Bearer
func (e SysUser) InsetAvatar(c *gin.Context) {
	s := service.SysUser{}
	req := dto.UpdateSysUserAvatarReq{}
	err := e.MakeContext(c).
		MakeOrm().
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	// 数据权限检查
	p := actions.GetPermissionFromContext(c)
	form, _ := c.MultipartForm()
	files := form.File["upload[]"]
	guid := uuid.New().String()
	filPath := "static/uploadfile/" + guid + ".jpg"
	for _, file := range files {
		e.Logger.Debugf("upload avatar file: %s", file.Filename)
		// 上传文件至指定目录
		err = c.SaveUploadedFile(file, filPath)
		if err != nil {
			e.Logger.Errorf("save file error, %s", err.Error())
			e.Error(500, err, "")
			return
		}
	}
	req.UserId = p.UserId
	req.Avatar = "/" + filPath

	updated, err := s.UpdateAvatar(&req, p)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	// 响应 shape 必须匹配前端 UploadUserAvatarResult { avatar, avatarType, avatarColor }
	// (vue-vben-admin/apps/web-antd/src/api/core/user.ts)；avatar 必须与 DB
	// 持久化值一致（带前导 /），否则刷新后渲染地址错乱。avatarColor 回传上传前的
	// 原值，便于前端 store 在切换回 letter 模式时仍能复用用户配置过的背景色。
	e.OK(map[string]string{
		"avatar":      updated.Avatar,
		"avatarType":  updated.AvatarType,
		"avatarColor": updated.AvatarColor,
	}, "修改成功")
}

// UpdateProfile
// @Summary 修改个人资料
// @Description 获取JSON
// @Tags 个人中心
// @Accept  application/json
// @Product application/json
// @Param data body dto.UpdateSysUserProfileReq true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/profile [put]
// @Security Bearer
func (e SysUser) UpdateProfile(c *gin.Context) {
	s := service.SysUser{}
	req := dto.UpdateSysUserProfileReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	p := actions.GetPermissionFromContext(c)
	req.UserId = user.GetUserId(c)
	req.SetUpdateBy(req.UserId)

	err = s.UpdateProfile(&req, p)
	if err != nil {
		e.Logger.Error(err)
		e.Error(http.StatusForbidden, err, "个人资料保存失败")
		return
	}
	middleware.AuditLogUpdate(c,
		"个人中心",
		middleware.AuditTarget{
			Type: middleware.AuditCategoryUser,
			ID:   req.UserId,
		},
		nil,
		map[string]interface{}{"introduction": req.Introduction},
		"admin.sysUser.updateProfile",
	)
	e.OK(nil, "保存成功")
}

// UpdateStatus 修改用户状态
// @Summary 修改用户状态
// @Description 获取JSON
// @Tags 用户
// @Accept  application/json
// @Product application/json
// @Param data body dto.UpdateSysUserStatusReq true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/status [put]
// @Security Bearer
func (e SysUser) UpdateStatus(c *gin.Context) {
	s := service.SysUser{}
	req := dto.UpdateSysUserStatusReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	req.SetUpdateBy(user.GetUserId(c))

	//数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.UpdateStatus(&req, p)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	middleware.AuditLog(c, middleware.AuditEntry{
		Title:  "用户管理",
		Action: middleware.AuditActionStatus,
		Target: middleware.AuditTarget{
			Type: middleware.AuditCategoryUser,
			ID:   req.UserId,
		},
		After:  map[string]interface{}{"status": req.Status},
		Method: "admin.sysUser.updateStatus",
	})
	e.OK(req.GetId(), "更新成功")
}

// ResetPwd 重置用户密码
// @Summary 重置用户密码
// @Description 获取JSON
// @Tags 用户
// @Accept  application/json
// @Product application/json
// @Param data body dto.ResetSysUserPwdReq true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/pwd/reset [put]
// @Security Bearer
func (e SysUser) ResetPwd(c *gin.Context) {
	s := service.SysUser{}
	req := dto.ResetSysUserPwdReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	req.SetUpdateBy(user.GetUserId(c))

	//数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.ResetPwd(&req, p)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	middleware.AuditLog(c, middleware.AuditEntry{
		Title:  "用户管理",
		Action: middleware.AuditActionPassword,
		Target: middleware.AuditTarget{
			Type: middleware.AuditCategoryUser,
			ID:   req.UserId,
		},
		Method: "admin.sysUser.resetPwd",
	})
	e.OK(req.GetId(), "更新成功")
}

// UpdatePwd
// @Summary 修改密码
// @Description 获取JSON
// @Tags 用户
// @Accept  application/json
// @Product application/json
// @Param data body dto.PassWord true "body"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/pwd/set [put]
// @Security Bearer
func (e SysUser) UpdatePwd(c *gin.Context) {
	s := service.SysUser{}
	req := dto.PassWord{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	// 数据权限检查
	p := actions.GetPermissionFromContext(c)

	err = s.UpdatePwd(user.GetUserId(c), req.OldPassword, req.NewPassword, p)
	if err != nil {
		e.Logger.Error(err)
		if err.Error() == "旧密码错误" {
			e.Error(http.StatusForbidden, err, err.Error())
			return
		}
		e.Error(http.StatusForbidden, err, "密码修改失败")
		return
	}

	e.OK(nil, "密码修改成功")
}

// GetProfile
// @Summary 获取个人中心用户
// @Description 获取JSON
// @Tags 个人中心
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/user/profile [get]
// @Security Bearer
func (e SysUser) GetProfile(c *gin.Context) {
	s := service.SysUser{}
	req := dto.SysUserById{}
	err := e.MakeContext(c).
		MakeOrm().
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	req.Id = user.GetUserId(c)

	sysUser := models.SysUser{}
	roles := make([]models.SysRole, 0)
	posts := make([]models.SysPost, 0)
	err = s.GetProfile(&req, &sysUser, &roles, &posts)
	if err != nil {
		e.Logger.Errorf("get user profile error, %s", err.Error())
		e.Error(500, err, "获取用户信息失败")
		return
	}
	e.OK(gin.H{
		"user":  sysUser,
		"roles": roles,
		"posts": posts,
	}, "查询成功")
}

// GetInfo
// @Summary 获取个人信息
// @Description 获取JSON
// @Tags 个人中心
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/getinfo [get]
// @Security Bearer
func (e SysUser) GetInfo(c *gin.Context) {
	req := dto.SysUserById{}
	s := service.SysUser{}
	r := service.SysRole{}
	err := e.MakeContext(c).
		MakeOrm().
		MakeService(&r.Service).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	primaryRoleID := authctx.GetPrimaryRoleID(c)
	primaryRoleName := authctx.GetPrimaryRoleName(c)
	primaryRoleKey := authctx.GetPrimaryRoleKey(c)
	roleIDs := authctx.GetRoleIDs(c)
	roleNames := authctx.GetRoleNames(c)
	roleKeys := authctx.GetRoleKeys(c)

	var mp = make(map[string]interface{})
	mp["roles"] = []string{primaryRoleName}
	mp["primaryRoleId"] = primaryRoleID
	mp["primaryRoleName"] = primaryRoleName
	mp["primaryRoleKey"] = primaryRoleKey
	mp["roleIds"] = roleIDs
	mp["roleNames"] = roleNames
	mp["roleKeys"] = roleKeys
	if containsAdminRole(roleKeys, roleNames) {
		mp["permissions"] = []string{"*:*:*"}
		mp["buttons"] = []string{"*:*:*"}
	} else {
		list, _ := r.GetByIds(roleIDs)
		mp["permissions"] = list
		mp["buttons"] = list
	}
	sysUser := models.SysUser{}
	req.Id = user.GetUserId(c)
	err = s.Get(&req, p, &sysUser)
	if err != nil {
		e.Error(http.StatusUnauthorized, err, "登录失败")
		return
	}
	mp["introduction"] = sysUser.Introduction
	mp["avatar"] = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
	if sysUser.Avatar != "" {
		mp["avatar"] = sysUser.Avatar
	}
	// 头像扩展字段：letter 头像（首字母+背景色）模式下，前端 store 需要这两个
	// 字段才能正确渲染；image 模式下也带上，便于回显当前类型。
	mp["avatarType"] = sysUser.AvatarType
	mp["avatarColor"] = sysUser.AvatarColor
	mp["userName"] = sysUser.Username
	mp["userId"] = sysUser.UserId
	mp["deptId"] = sysUser.DeptId
	mp["name"] = sysUser.NickName
	mp["code"] = 200
	e.OK(mp, "")
}

func containsAdminRole(roleKeys []string, roleNames []string) bool {
	for _, roleKey := range roleKeys {
		if strings.EqualFold(roleKey, "admin") {
			return true
		}
	}
	for _, roleName := range roleNames {
		if roleName == "系统管理员" || strings.EqualFold(roleName, "admin") {
			return true
		}
	}
	return false
}
