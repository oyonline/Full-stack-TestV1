package kingdeeUtils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-admin/config"
	"io"
	"net/http"
	"strings"
)

var (
	Cookie         = ""
	defaultHeaders = map[string]string{
		"Accept":         "application/json",
		"Content-Type":   "application/json",
		"Accept-Charset": "utf-8",
		"User-Agent":     "Kingdee/Golang WebApi SDK (compatible: K3Cloud 7.3+)",
	}
)

func Execute(url string, headers map[string]interface{}, postData map[string]interface{}) string {
	jsonData, err := json.Marshal(postData)
	if err != nil {
		fmt.Println(err)
		return "err"
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	for k, v := range defaultHeaders {
		req.Header.Add(k, v)
	}
	for k, v := range headers {
		str := fmt.Sprintf("%v", v)
		req.Header.Add(k, str)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println(err)
		return "err"
	}
	Cookie = strings.Join(resp.Header["Set-Cookie"], ";")
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// 登录: 签名
func GetHeaders() map[string]interface{} {
	// cookie
	if Cookie == "" {
		props := config.ExtConfig.Kingdee
		url := props.Hosturl + LOGIN_API
		postData := map[string]interface{}{
			"acctid":   props.Acctid,
			"username": props.Username,
			"password": props.Password,
			"lcid":     props.Lcid,
		}
		Execute(url, make(map[string]interface{}), postData)
	}
	return map[string]interface{}{"Cookie": Cookie}
}

// 详情
func View(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + VIEW_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 单据查询
func ExecuteBillQuery(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + EXECUTEBILLQUERY_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 单据查询(json) （官方在2023.9.4新增此接口）
func BillQuery(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + BILLQUERY_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 元数据查询
func QueryBusinessInfo(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + QUERYBUSINESSINFO_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 获取数据中心列表
func GetDataCenterList() string {
	url := config.ExtConfig.Kingdee.Hosturl + GETDATACENTERLIST_API
	respStr := Execute(url, GetHeaders(), make(map[string]interface{}))
	return respStr
}

// 保存
func Save(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + SAVE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 批量保存
func BatchSave(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + BATCHSAVE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 审核
func Audit(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + AUDIT_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 反审核
func UnAudit(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + UNAUDIT_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 提交
func Submit(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + SUBMIT_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 操作
func Operation(formId string, opNumber string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + EXCUTEOPERATION_API
	postData := map[string]interface{}{
		"formid":   formId,
		"opNumber": opNumber,
		"data":     data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 下推
func Push(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + PUSH_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 暂存
func Draft(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + DRAFT_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 删除
func Delete(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + DELETE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 分配
func Allocate(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + ALLOCATE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 取消分配
func CancelAllocate(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + CANCEL_ALLOCATE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 弹性域保存
func FlexSave(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + FLEXSAVE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 发送消息
func SendMsg(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + SENDMSG_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 分组保存
func GroupSave(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + GROUPSAVE_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 拆单
func Disassembly(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + DISASSEMBLY_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 工作流审批
func WorkflowAudit(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + WORKFLOWAUDIT_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 查询分组信息
func QueryGroupInfo(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + QUERYGROUPINFO_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 分组删除
func GroupDelete(data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + GROUPDELETE_API
	postData := map[string]interface{}{
		"data": data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}

// 查询报表数据
func GetSysReportData(formId string, data map[string]string) string {
	url := config.ExtConfig.Kingdee.Hosturl + GET_SYS_REPORT_DATA_API
	postData := map[string]interface{}{
		"formid": formId,
		"data":   data,
	}
	respStr := Execute(url, GetHeaders(), postData)
	return respStr
}
