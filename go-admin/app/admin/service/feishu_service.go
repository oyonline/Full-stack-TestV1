package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-admin/app/admin/models"
	"go-admin/common/utils/feishuUtils"
	"strings"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type FeishuService struct {
	service.Service
}

// SendMessage 发送飞书消息
func (e *FeishuService) SendMessage(userOpenid, tempCode string, params map[string]string) error {
	template := e.getTemplate(tempCode, params)
	if template.ID > 0 {
		feishuClient, err := feishuUtils.NewFeishuClient()
		if err != nil {
			return nil
		}
		resp, err := feishuClient.SendMessage(userOpenid, template.MsgType, template.Template)
		if err != nil {
			return nil
		}
		var respEnt models.FeishuMessageResponse
		jsonResp, err := json.Marshal(resp)
		jsonStr := string(jsonResp)
		err = json.Unmarshal([]byte(jsonStr), &respEnt)
		if err != nil {
			return err
		}
		if respEnt.Body.Content != nil {
			respEnt.BodyStr = *respEnt.Body.Content
		}
		respEnt.TemplateId = template.ID
		return e.Orm.Create(&respEnt).Error
	}
	return errors.New(tempCode + " 消息模板未找到")
}

func (e *FeishuService) getTemplate(tempCode string, params map[string]string) models.FeishuMessageTemplate {
	var temp models.FeishuMessageTemplate
	err := e.Orm.Model(&temp).Where("template_code = ?", tempCode).First(&temp).Error
	if err == nil || temp.ID > 0 {
		paramFields := strings.Split(temp.Params, ",")
		templateJson := temp.Template
		for _, v := range paramFields {
			if _, ok := params[v]; !ok {
				params[v] = ""
			}
			param := fmt.Sprintf("{%s}", v)
			templateJson = strings.ReplaceAll(templateJson, param, params[v])
		}
		temp.Template = templateJson
	}
	return temp
}
