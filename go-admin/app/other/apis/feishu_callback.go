package apis

import (
	"encoding/json"
	"errors"
	"go-admin/app/other/service"
	"go-admin/app/other/service/dto"
	"go-admin/common/utils"
	"go-admin/common/utils/feishuUtils"
	"go-admin/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
)

type FeishuCallback struct {
	api.Api
}

func (e FeishuCallback) Subscript(c *gin.Context) {
	client, err := feishuUtils.NewFeishuClient()
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	appCode := c.Query("appCode")
	if appCode == "" {
		err = errors.New("appCode is required")
		e.Error(500, err, err.Error())
		return
	}
	resp, err := client.Subscript(appCode)
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	e.OK(resp, "订阅成功")
}

func (e FeishuCallback) OrgList(c *gin.Context) {
	s := service.FeishuOptions{}
	err := e.MakeContext(c).
		MakeOrm().
		MakeService(&s.Service).
		Errors
	if err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	response := s.OrgList()
	c.JSON(http.StatusOK, response)
	return
}

func (e FeishuCallback) DepartmentList(c *gin.Context) {
	req := dto.FeishuRequest{}
	s := service.FeishuOptions{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	response := s.DepartmentList(req.LinkageParams.OrgName, req.LinkageParams.DepartmentName)
	c.JSON(http.StatusOK, response)
	return
}

func (e FeishuCallback) PlatformList(c *gin.Context) {
	req := dto.FeishuRequest{}
	s := service.FeishuOptions{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	response := s.PlatformList(req.LinkageParams.DepartmentName)
	c.JSON(http.StatusOK, response)
	return
}

func (e FeishuCallback) FeeCode(c *gin.Context) {
	req := dto.FeishuRequest{}
	s := service.FeishuOptions{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	response := s.FeeCodeList(req.LinkageParams.Platform)
	c.JSON(http.StatusOK, response)
	return
}

func (e FeishuCallback) Callback(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	// 2. 通用解析为 map (用于判断类型和记录日志)
	var content map[string]interface{}
	if err = json.Unmarshal(body, &content); err != nil {
		c.String(http.StatusOK, "ok")
		return
	}
	// 飞书配置回调地址时会发送 type: "url_verification" 或包含 challenge 字段
	if challenge, ok := content["challenge"]; ok {
		c.JSON(http.StatusOK, gin.H{
			"challenge": challenge,
		})
		return
	}
	feishuConfig := config.ExtConfig.Feishu
	feishuToken := feishuConfig.Token
	if feishuToken != "" && content["token"] != feishuToken {
		e.Logger.Errorf("Feishu Invalid token %s != %s", content["token"], feishuToken)
		c.String(http.StatusOK, "ok")
		return
	}
	feishuAppCode := feishuConfig.Approvalcode
	// 尝试解析为强类型结构体
	var eventEnt dto.FeishuEventCallback
	if err = json.Unmarshal(body, &eventEnt); err != nil {
		// 如果不是标准事件结构，但也不是 challenge，通常也返回 ok 避免飞书重试，或者返回错误
		c.String(http.StatusOK, "ok")
		return
	}
	if !utils.InTypeArray(eventEnt.Event.ApprovalCode, feishuAppCode) {
		c.String(http.StatusOK, "ok")
		return
	}
	s := service.FeishuRequest{}
	err = e.MakeContext(c).
		MakeOrm().
		MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Errorf("Failed to make service: %s", err.Error())
		c.String(http.StatusOK, "ok")
		return
	}
	// 检查必要字段是否存在 (对应 if eventEnt.Event.UserID != "")
	if eventEnt.Event.UserID != "" && eventEnt.Event.InstanceCode != "" {
		feishuClient, err := feishuUtils.NewFeishuClient()
		if err != nil {
			e.Logger.Errorf("Failed to create Feishu client: %s", err.Error())
		}
		// 根据飞书事件ID 和 飞书用户ID 获取详情
		resp, err := feishuClient.GetDetail(eventEnt.Event.InstanceCode, eventEnt.Event.UserID)
		if err != nil {
			e.Logger.Errorf("FeishuClient: %s", err.Error())
			c.String(http.StatusOK, "ok")
			return
		}
		err = s.ProcessCallback(&resp)
		if err != nil {
			e.Logger.Errorf("CallBack Struct: %s", err.Error())
			c.String(http.StatusOK, "ok")
			return
		}
	} else {
		e.Logger.Error("Event received but missing UserID or InstanceCode")
	}
	c.String(http.StatusOK, "ok")
}
