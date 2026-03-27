package dto

import "go-admin/app/admin/models"

// FeishuRequest 飞书自定义字段筛选
type FeishuRequest struct {
	EmployeeId    string        `json:"employeeId"`
	EmployeeID    string        `json:"employee_id"`
	LinkageParams LinkageParams `json:"linkage_params"`
	Locale        string        `json:"locale"`
	OpenId        string        `json:"openId"`
	OpenID        string        `json:"open_id"`
	TenantKey     string        `json:"tenantKey"`
	TenantKeyAlt  string        `json:"tenant_key"` // 注意：Go 字段名不能以数字开头且不能重复，这里做了区分
	Token         string        `json:"token"`
	UserId        string        `json:"userId"`
	UserID        string        `json:"user_id"`
}

// LinkageParams 对应 linkage_params 嵌套对象
type LinkageParams struct {
	Platform       string `json:"platform"`
	DepartmentName string `json:"department"`
	OrgName        string `json:"org"`
}

// FeishuEventCallback 飞书事件回调
type FeishuEventCallback struct {
	Event EventDetail `json:"event"`
	Token string      `json:"token"`
	Ts    string      `json:"ts"` // 时间戳字符串，例如 "1773735602.781551"
	Type  string      `json:"type"`
	UUID  string      `json:"uuid"`
}

// EventDetail 对应 event 字段内部的结构体
type EventDetail struct {
	AppID        string `json:"app_id"`
	ApprovalCode string `json:"approval_code"`
	CustomKey    string `json:"custom_key"`
	DefKey       string `json:"def_key"`
	GenerateType string `json:"generate_type"`
	InstanceCode string `json:"instance_code"`
	OpenID       string `json:"open_id"`
	OperateTime  string `json:"operate_time"` // 毫秒级时间戳字符串
	Status       string `json:"status"`
	TaskID       string `json:"task_id"`
	TenantKey    string `json:"tenant_key"`
	Type         string `json:"type"` // 注意：这里 event 内部也有一个 type 字段
	UserID       string `json:"user_id"`
}

// FeishuApiResponse 飞书审批详情
type FeishuApiResponse struct {
	Code  int          `json:"code"`
	Msg   string       `json:"msg"`
	Error interface{}  `json:"error"`
	Data  ApprovalData `json:"data"`
}

// --- 核心业务数据 (Data) ---

// ApprovalData 对应 Data 部分
type ApprovalData struct {
	ApprovalName string                `json:"approval_name"`
	StartTime    string                `json:"start_time"` // 毫秒时间戳字符串
	EndTime      string                `json:"end_time"`
	UserId       string                `json:"user_id"`
	OpenId       string                `json:"open_id"`
	SerialNumber string                `json:"serial_number"`
	DepartmentId string                `json:"department_id"`
	Status       string                `json:"status"` // 例如 "REJECTED", "APPROVED"
	Uuid         string                `json:"uuid"`
	FormStr      string                `json:"form"`
	Form         []FormWidget          `json:"-"`
	Timeline     []TimelineEvent       `json:"timeline"`
	TaskList     models.FeishuTaskList `json:"task_list"`
	ApprovalCode string                `json:"approval_code"`
	Reverted     bool                  `json:"reverted"`
	InstanceCode string                `json:"instance_code"`
}

// TimelineEvent 对应 Timeline 中的元素
type TimelineEvent struct {
	Type       string `json:"type"` // 例如 "START", "REJECT", "APPROVE"
	CreateTime string `json:"create_time"`
	UserId     string `json:"user_id"`
	OpenId     string `json:"open_id"`
	TaskId     string `json:"task_id,omitempty"` // 可选字段
	Comment    string `json:"comment,omitempty"`
	Ext        string `json:"ext"` // 通常是 JSON 字符串 "{}"
	NodeKey    string `json:"node_key"`
}

// FormWidget 代表表单中的一个控件
type FormWidget struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Ext    interface{} `json:"ext"` // 结构复杂且多变，建议先用 interface{} 或自定义结构
	Value  interface{} `json:"value"`
	Option *FormOption `json:"option,omitempty"`
}

type FormOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}
