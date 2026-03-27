package handler

import (
	"go-admin/common"
	"go-admin/common/authctx"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/response"
	"github.com/mssola/user_agent"
	"go-admin/common/global"
)

func PayloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		u, _ := v["user"].(SysUser)
		r, _ := v["role"].(SysRole)
		roles, _ := v["roles"].([]SysRole)
		roleIDs := make([]int, 0, len(roles))
		roleKeys := make([]string, 0, len(roles))
		roleNames := make([]string, 0, len(roles))
		for _, role := range roles {
			roleIDs = append(roleIDs, role.RoleId)
			roleKeys = append(roleKeys, role.RoleKey)
			roleNames = append(roleNames, role.RoleName)
		}
		if len(roleIDs) == 0 && r.RoleId > 0 {
			roleIDs = []int{r.RoleId}
		}
		if len(roleKeys) == 0 && r.RoleKey != "" {
			roleKeys = []string{r.RoleKey}
		}
		if len(roleNames) == 0 && r.RoleName != "" {
			roleNames = []string{r.RoleName}
		}
		return jwt.MapClaims{
			jwt.IdentityKey:   u.UserId,
			jwt.RoleIdKey:     r.RoleId,
			jwt.RoleKey:       r.RoleKey,
			jwt.NiceKey:       u.Username,
			jwt.DataScopeKey:  r.DataScope,
			jwt.RoleNameKey:   r.RoleName,
			"primaryRoleId":   r.RoleId,
			"primaryRoleKey":  r.RoleKey,
			"primaryRoleName": r.RoleName,
			"roleIds":         roleIDs,
			"roleKeys":        roleKeys,
			"roleNames":       roleNames,
		}
	}
	return jwt.MapClaims{}
}

func IdentityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	return map[string]interface{}{
		"IdentityKey":     claims["identity"],
		"UserName":        claims["nice"],
		"RoleKey":         claims["rolekey"],
		"UserId":          claims["identity"],
		"PrimaryRoleId":   claims["primaryRoleId"],
		"PrimaryRoleKey":  claims["primaryRoleKey"],
		"PrimaryRoleName": claims["primaryRoleName"],
		"RoleIds":         claims["roleIds"],
		"RoleKeys":        claims["roleKeys"],
		"RoleNames":       claims["roleNames"],
		"DataScope":       claims["datascope"],
	}
}

// Authenticator 获取token
// @Summary 登陆
// @Description 获取token
// @Description LoginHandler can be used by clients to get a jwt token.
// @Description Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// @Description Reply will be of the form {"token": "TOKEN"}.
// @Description dev mode：It should be noted that all fields cannot be empty, and a value of 0 can be passed in addition to the account password
// @Description 注意：开发模式：需要注意全部字段不能为空，账号密码外可以传入0值
// @Tags 登陆
// @Accept  application/json
// @Product application/json
// @Param account body Login  true "account"
// @Success 200 {string} string "{"code": 200, "expire": "2019-08-07T12:45:48+08:00", "token": ".eyJleHAiOjE1NjUxNTMxNDgsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTU2NTE0OTU0OH0.-zvzHvbg0A" }"
// @Router /api/v1/login [post]
func Authenticator(c *gin.Context) (interface{}, error) {
	log := api.GetRequestLogger(c)
	db, err := pkg.GetOrm(c)
	if err != nil {
		log.Errorf("get db error, %s", err.Error())
		response.Error(c, 500, err, "数据库连接获取失败")
		return nil, jwt.ErrFailedAuthentication
	}

	var loginVals Login
	var status = "2"
	var msg = "登录成功"
	var username = ""
	defer func() {
		LoginLogToDB(c, status, msg, username)
	}()

	if err = c.ShouldBind(&loginVals); err != nil {
		username = loginVals.Username
		msg = "数据解析失败"
		status = "1"

		return nil, jwt.ErrMissingLoginValues
	}
	if !captcha.Verify(loginVals.UUID, loginVals.Code, true) {
		username = loginVals.Username
		msg = "验证码错误"
		status = "1"

		return nil, jwt.ErrInvalidVerificationode
	}
	sysUser, role, roles, e := loginVals.GetUser(db)
	if e == nil {
		username = loginVals.Username

		return map[string]interface{}{"user": sysUser, "role": role, "roles": roles}, nil
	} else {
		msg = "登录失败"
		status = "1"
		log.Warnf("%s login failed!", loginVals.Username)
	}
	return nil, jwt.ErrFailedAuthentication
}

// LoginLogToDB Write log to database
func LoginLogToDB(c *gin.Context, status string, msg string, username string) {
	if !config.LoggerConfig.EnabledDB {
		return
	}
	log := api.GetRequestLogger(c)
	l := make(map[string]interface{})

	ua := user_agent.New(c.Request.UserAgent())
	l["ipaddr"] = common.GetClientIP(c)
	l["loginLocation"] = "" // pkg.GetLocation(common.GetClientIP(c),gaConfig.ExtConfig.AMap.Key)
	l["loginTime"] = pkg.GetCurrentTime()
	l["status"] = status
	l["remark"] = c.Request.UserAgent()
	browserName, browserVersion := ua.Browser()
	l["browser"] = browserName + " " + browserVersion
	l["os"] = ua.OS()
	l["platform"] = ua.Platform()
	l["username"] = username
	l["msg"] = msg

	q := sdk.Runtime.GetMemoryQueue("")
	message, err := sdk.Runtime.GetStreamMessage(c.Request.Host, global.LoginLog, l)
	if err != nil {
		log.Errorf("GetStreamMessage error, %s", err.Error())
		//日志报错错误，不中断请求
	} else {
		err = q.Append(message)
		if err != nil {
			log.Errorf("Append message error, %s", err.Error())
		}
	}
}

// LogOut
// @Summary 退出登录
// @Description 获取token
// LoginHandler can be used by clients to get a jwt token.
// Reply will be of the form {"token": "TOKEN"}.
// @Accept  application/json
// @Product application/json
// @Success 200 {string} string "{"code": 200, "msg": "成功退出系统" }"
// @Router /logout [post]
// @Security Bearer
func LogOut(c *gin.Context) {
	LoginLogToDB(c, "2", "退出成功", user.GetUserName(c))
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "退出成功",
	})

}

func Authorizator(data interface{}, c *gin.Context) bool {

	if v, ok := data.(map[string]interface{}); ok {
		if primaryRoleName, ok := v["PrimaryRoleName"].(string); ok {
			c.Set("primaryRoleName", primaryRoleName)
			c.Set("role", primaryRoleName)
		}
		if primaryRoleKey, ok := v["PrimaryRoleKey"].(string); ok {
			c.Set("primaryRoleKey", primaryRoleKey)
		}
		if primaryRoleID, ok := v["PrimaryRoleId"].(float64); ok {
			c.Set("primaryRoleId", int(primaryRoleID))
		}
		switch roleNames := v["RoleNames"].(type) {
		case []string:
			c.Set("roleNames", roleNames)
		case []interface{}:
			names := make([]string, 0, len(roleNames))
			for _, item := range roleNames {
				if value, ok := item.(string); ok && value != "" {
					names = append(names, value)
				}
			}
			c.Set("roleNames", names)
		}
		switch roleKeys := v["RoleKeys"].(type) {
		case []string:
			c.Set("roleKeys", roleKeys)
		case []interface{}:
			keys := make([]string, 0, len(roleKeys))
			for _, item := range roleKeys {
				if value, ok := item.(string); ok && value != "" {
					keys = append(keys, value)
				}
			}
			c.Set("roleKeys", keys)
		}
		switch roleIDs := v["RoleIds"].(type) {
		case []int:
			c.Set("roleIds", roleIDs)
		case []interface{}:
			ids := make([]int, 0, len(roleIDs))
			for _, item := range roleIDs {
				switch value := item.(type) {
				case float64:
					ids = append(ids, int(value))
				case int:
					ids = append(ids, value)
				}
			}
			c.Set("roleIds", ids)
		case float64:
			c.Set("roleIds", []int{int(roleIDs)})
		}
		if userID, ok := v["UserId"].(float64); ok {
			c.Set("userId", int(userID))
		}
		if userName, ok := v["UserName"].(string); ok {
			c.Set("userName", userName)
		}
		if dataScope, ok := v["DataScope"].(string); ok {
			c.Set("dataScope", dataScope)
		}
		if len(authctx.GetRoleIDs(c)) == 0 {
			return false
		}
		return true
	}
	return false
}

func Unauthorized(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  message,
	})
}
