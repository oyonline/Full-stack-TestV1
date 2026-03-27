package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// FeishuApprovalRecord 飞书审批实例数据
type FeishuApprovalRecord struct {
	ID               int64              `gorm:"primaryKey;autoIncrement" json:"id"` // 数据库自增主键
	ApprovalName     string             `gorm:"column:approval_name;type:varchar(255);comment:审批名称" json:"approvalName"`
	StartDate        string             `gorm:"column:start_time;type:varchar(20);index:idx_start_time;comment:开始时间(毫秒时间戳)" json:"startDate"`
	EndDate          string             `gorm:"column:end_time;type:varchar(20);index:idx_end_time;comment:结束时间(毫秒时间戳)" json:"endDate"`
	UserId           string             `gorm:"column:user_id;type:varchar(64);index:idx_user_id;comment:用户ID" json:"userId"`
	OpenId           string             `gorm:"column:open_id;type:varchar(64);index:idx_open_id;comment:用户OpenID" json:"openId"`
	SerialNumber     string             `gorm:"column:serial_number;type:varchar(64);uniqueIndex:idx_serial;comment:流水号" json:"serialNumber"`
	DepartmentId     string             `gorm:"column:department_id;type:varchar(64);comment:申请部门ID" json:"departmentId"`
	BearDepartmentId string             `gorm:"column:bear_department_id;type:varchar(64);comment:承担部门ID" json:"bearDepartmentId"`
	Status           string             `gorm:"column:status;type:varchar(32);index:idx_status;comment:审批状态" json:"status"`
	Uuid             string             `gorm:"column:uuid;type:varchar(64);uniqueIndex:idx_uuid;comment:业务UUID" json:"uuid"`
	ApprovalCode     string             `gorm:"column:approval_code;type:varchar(64);uniqueIndex:idx_approval_code;comment:审批实例Code" json:"approvalCode"`
	Reverted         bool               `gorm:"column:reverted;type:tinyint(1);default:0;comment:是否回滚" json:"reverted"`
	InstanceCode     string             `gorm:"column:instance_code;type:varchar(64);uniqueIndex:idx_instance;comment:实例Code" json:"instanceCode"`
	FormStr          string             `gorm:"column:form_json" json:"-"`
	Form             FeishuApprovalForm `gorm:"foreignKey:InstanceCode;references:InstanceCode" json:"formData"`
	Timeline         FeishuTimeLines    `gorm:"foreignKey:InstanceCode;references:InstanceCode" json:"timeline"`
	TaskList         FeishuTaskList     `json:"taskList" gorm:"column:task_list;type:longtext"`
}

// TableName 指定表名 (可选，默认会是 approval_records)
func (FeishuApprovalRecord) TableName() string {
	return "feishu_approval_records"
}

type FeishuApprovalForm struct {
	ID                    int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceCode          string             `gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	DepartmentId          string             `gorm:"column:department_id;type:varchar(100);comment:承担部门Id"`
	KingdeeDepartmentCode string             `gorm:"column:kingdee_department_code;type:varchar(32);comment:金蝶部门编码"`
	Platform              string             `json:"appName" gorm:"column:platform;type:varchar(64);comment:平台"`
	OrgCode               string             `json:"orgCode" gorm:"column:org_code;type:varchar(64);comment:付款主体"`
	Currency              string             `json:"currency" gorm:"column:currency;type:varchar(10);comment:货币单位"`
	TotalAmount           float64            `json:"totalAmount" gorm:"column:total_amount;type:decimal(10,2);comment:总金额"`
	Capital               string             `json:"capital" gorm:"column:capital;type:varchar(100)"`
	Attachments           []FeishuAttachment `json:"attachments" gorm:"foreignKey:InstanceCode;references:InstanceCode"`
	Details               []FeeDetail        `json:"details" gorm:"foreignKey:InstanceCode;references:InstanceCode"`
}

func (FeishuApprovalForm) TableName() string {
	return "feishu_approval_form"
}

type FeeDetail struct {
	ID           int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceCode string  `json:"-" gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	FeeCode      string  `json:"feeCode" gorm:"column:fee_code;type:varchar(32);comment:费用编码" json:"fee_code"`
	Reason       string  `json:"reason" gorm:"column:reason;type:text;comment:报销事由"`
	Currency     string  `json:"currency" gorm:"column:currency;type:varchar(10);comment:货币单位"`
	FeeDate      string  `json:"feeDate" gorm:"column:fee_date;type:varchar(10);comment:费用日期"`
	FeeAmount    float64 `json:"feeAmount" gorm:"column:fee_amount;type:decimal(10,2);comment:金额"`
}

func (FeeDetail) TableName() string {
	return "feishu_approval_fee_detail"
}

type FeishuAttachment struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceCode string `json:"-" gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	FileName     string `json:"fileName" gorm:"column:file_name;type:varchar(255);"`
	FileUrl      string `json:"fileUrl" gorm:"column:file_url;type:text;"`
}

func (FeishuAttachment) TableName() string {
	return "feishu_approval_attachment"
}

type FeishuTimeline struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceCode string     `json:"-" gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	Type         string     `json:"type" gorm:"column:type;varchar(20);comment:状态 START,REJECT,APPROVE"`
	CreateDate   string     `json:"createDate" gorm:"column:create_date;varchar(20);comment:时间"`
	UserId       string     `json:"userId" gorm:"column:user_id;type:varchar(64);comment:用户ID"`
	OpenId       string     `json:"openId" gorm:"column:open_id;type:varchar(64);comment:用户OpenID"`
	TaskId       string     `json:"taskId" gorm:"column:task_id;type:varchar(64);comment:TaskId"` // 可选字段
	Comment      string     `json:"comment" gorm:"column:comment;type:text;comment:评论"`
	Ext          string     `json:"ext" gorm:"column:ext;type:text;comment:扩展信息json"`
	NodeKey      string     `json:"nodeKey" gorm:"column:nodeKey;type:varchar(64);comment:扩展信息json"`
	OptUser      SimpleUser `json:"optUser" gorm:"foreignKey:OpenId;references:OpenId"`
}

func (FeishuTimeline) TableName() string {
	return "feishu_approval_timeline"
}

type FeishuTask struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	OpenId    string `json:"OpenId"`
	Status    string `json:"Status"`
	NodeId    string `json:"NodeId"`
	NodeName  string `json:"NodeName"`
	Type      string `json:"Type"`
	StartTime string `json:"StartTime"`
	EndTime   string `json:"EndTime"`
}

type FeishuAttachmentList []FeishuAttachment

func (ql FeishuAttachmentList) Value() (driver.Value, error) {
	if ql == nil {
		return nil, nil // 如果数据库字段允许 NULL，返回 nil
		// 或者返回 json.Marshal([]Quote{}) 如果你想存空数组 "[]"
	}
	return json.Marshal(ql)
}

func (ql *FeishuAttachmentList) Scan(value interface{}) error {
	if value == nil {
		*ql = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into FeishuAttachmentList", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*ql = nil
		return nil
	}

	err := json.Unmarshal(bytes, ql)
	if err != nil {
		return fmt.Errorf("error unmarshalling FeishuAttachmentList JSON: %w", err)
	}

	return nil
}

type FeishuTimeLines []FeishuTimeline

func (ql FeishuTimeLines) Value() (driver.Value, error) {
	if ql == nil {
		return nil, nil // 如果数据库字段允许 NULL，返回 nil
		// 或者返回 json.Marshal([]Quote{}) 如果你想存空数组 "[]"
	}
	return json.Marshal(ql)
}

func (ql *FeishuTimeLines) Scan(value interface{}) error {
	if value == nil {
		*ql = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into FeishuTaskList", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*ql = nil
		return nil
	}

	err := json.Unmarshal(bytes, ql)
	if err != nil {
		return fmt.Errorf("error unmarshalling FeishuTaskList JSON: %w", err)
	}

	return nil
}

type FeishuTaskList []FeishuTask

func (ql FeishuTaskList) Value() (driver.Value, error) {
	if ql == nil {
		return nil, nil // 如果数据库字段允许 NULL，返回 nil
		// 或者返回 json.Marshal([]Quote{}) 如果你想存空数组 "[]"
	}
	return json.Marshal(ql)
}

func (ql *FeishuTaskList) Scan(value interface{}) error {
	if value == nil {
		*ql = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into FeishuTaskList", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*ql = nil
		return nil
	}

	err := json.Unmarshal(bytes, ql)
	if err != nil {
		return fmt.Errorf("error unmarshalling FeishuTaskList JSON: %w", err)
	}

	return nil
}
