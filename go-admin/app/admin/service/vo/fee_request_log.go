package vo

import (
	"go-admin/app/admin/models"
	"sort"
	"time"

	"gorm.io/gorm"
)

type FeeRequestLog struct {
	Id              int64             `json:"id" gorm:"primaryKey;autoIncrement;column:id;type:bigint"`
	InstanceCode    string            `json:"instanceCode" gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	ReqUserOpenid   string            `json:"-" gorm:"column:req_user_openid;type:varchar(64)" comment:"申请人openid"`
	Status          string            `json:"status" gorm:"column:status;type:varchar(20)"`
	OrgCode         string            `json:"orgCode" gorm:"column:org_code;type:varchar(64)" comment:"付款主体"`
	Currency        string            `json:"currency" gorm:"column:currency;type:varchar(10)"`
	YearsMonth      string            `json:"yearsMonth" gorm:"column:budget_years_month;type:varchar(10)" comment:"预算年月"`
	BudgetAmount    float64           `json:"budgetAmount" gorm:"column:budget_amount;type:decimal(10,2)" comment:"预算总额"`
	BudgetUsed      float64           `json:"budgetUsed" gorm:"column:budget_used;type:decimal(10,2)" comment:"已使用预算"`
	RequestAmount   float64           `json:"requestAmount" gorm:"column:request_amount;type:decimal(10,2)" comment:"申请金额"`
	CostCenterId    int64             `json:"costCenterId" gorm:"column:cost_center_info_id;type:bigint" comment:"成本中心ID"`
	CostCenterName  string            `json:"costCenterName" gorm:"column:cost_center_name;type:varchar(200)" comment:"成本中心名称"`
	BudgetVersionId int64             `json:"budgetVersionId" gorm:"column:budget_version_id;type:bigint" comment:"预算版本ID"`
	GroupName       string            `json:"groupName" gorm:"column:group_name;type:varchar(200)" comment:"客户分组名称"`
	GroupCode       string            `json:"groupCode" gorm:"column:group_number;type:varchar(100)" comment:"客服分组编码"`
	BudgetDetailId  int64             `json:"budgetDetailId" gorm:"column:budget_detail_id;type:bigint" comment:"预算详情ID"`
	FeeCode         string            `json:"feeCode" gorm:"column:fee_code;type:varchar(50)" comment:"费用编码"`
	FeeName         string            `json:"feeName" gorm:"column:fee_name;type:varchar(200)" comment:"费用名称"`
	UserDeptId      int               `json:"-" orm:"column:user_dept_id;type:bigint" comment:"申请部门ID"`
	DepartmentId    int               `json:"-" gorm:"column:dept_id;type:bigint" comment:"承担部门ID"`
	RequestTime     time.Time         `json:"-" gorm:"column:request_time;type:datetime" comment:"费用申请时间"`
	RequestTimeStr  string            `json:"requestTime" gorm:"-" comment:"费用申请时间"`
	BearDept        models.SimpleDept `json:"bearDept" gorm:"foreignKey:DepartmentId;references:DeptId"`
	ReqUser         models.SimpleUser `json:"reqUser" gorm:"foreignKey:ReqUserOpenid;references:OpenId"`
	ReqDept         models.SimpleDept `json:"reqDept" gorm:"foreignKey:UserDeptId;references:DeptId"`
}

func (FeeRequestLog) TableName() string {
	return "fee_request_log"
}
func (e *FeeRequestLog) BeforeCreate(_ *gorm.DB) error {
	e.RequestTime = time.Now()
	return nil
}

func (e *FeeRequestLog) AfterFind(db *gorm.DB) error {
	e.RequestTimeStr = e.RequestTime.Format(time.DateTime)
	return nil
}

type FeeRequestLogDetail struct {
	Id                    int64                       `json:"id" gorm:"primaryKey;autoIncrement;column:id;type:bigint"`
	InstanceCode          string                      `json:"instanceCode" gorm:"column:instance_code;type:varchar(64);comment:实例Code" swagger:"string" description:"实例Code" example:"5505CBBB-DFC3-44DB-A292-6FDFB328745E"`
	ReqUserOpenid         string                      `json:"-" gorm:"column:req_user_openid;type:varchar(64)" swagger:"string" description:"申请人openid" example:"ou_0058d139b05b555270ffe9a3e42b0961"`
	Status                string                      `json:"status" gorm:"column:status;type:varchar(20)"  swagger:"string" description:"单据状态" example:"PENDING,CANCELED,REJECTED,APPROVED"`
	OrgCode               string                      `json:"orgCode" gorm:"column:org_code;type:varchar(64)" comment:"付款主体编码"`
	OrgName               string                      `json:"orgName" gorm:"column:fname" comment:"付款主体"`
	CategoryName          string                      `json:"categoryName" gorm:"->" comment:"费用类别"`
	CategoryCode          string                      `json:"categoryCode" gorm:"->" comment:"费用类别编码"`
	KingdeeDepartmentCode string                      `json:"kingdeeDepartmentCode" gorm:"->" comment:"金蝶部门编码"`
	Currency              string                      `json:"currency" gorm:"column:currency;type:varchar(10)"`
	YearsMonth            string                      `json:"yearsMonth" gorm:"column:budget_years_month;type:varchar(10)" comment:"预算年月"`
	BudgetAmount          float64                     `json:"budgetAmount" gorm:"column:budget_amount;type:decimal(10,2)" comment:"预算总额"`
	BudgetUsed            float64                     `json:"budgetUsed" gorm:"column:budget_used;type:decimal(10,2)" comment:"已使用预算"`
	RequestAmount         float64                     `json:"requestAmount" gorm:"column:request_amount;type:decimal(10,2)" comment:"申请金额"`
	CostCenterId          int64                       `json:"costCenterId" gorm:"column:cost_center_info_id;type:bigint" comment:"成本中心ID"`
	CostCenterName        string                      `json:"costCenterName" gorm:"column:cost_center_name;type:varchar(200)" comment:"成本中心名称"`
	BudgetVersionId       int64                       `json:"-" gorm:"column:budget_version_id;type:bigint" comment:"预算版本ID"`
	GroupName             string                      `json:"groupName" gorm:"column:group_name;type:varchar(200)" comment:"客户分组名称"`
	GroupCode             string                      `json:"groupCode" gorm:"column:group_number;type:varchar(100)" comment:"客服分组编码"`
	BudgetDetailId        int64                       `json:"-" gorm:"column:budget_detail_id;type:bigint" comment:"预算详情ID"`
	FeeCode               string                      `json:"feeCode" gorm:"column:fee_code;type:varchar(50)" comment:"费用编码"`
	FeeName               string                      `json:"feeName" gorm:"column:fee_name;type:varchar(200)" comment:"费用名称"`
	UserDeptId            int                         `json:"-" orm:"column:user_dept_id;type:bigint" comment:"申请部门ID"`
	DepartmentId          int                         `json:"-" gorm:"column:dept_id;type:bigint" comment:"承担部门ID"`
	RequestTime           time.Time                   `json:"-" gorm:"column:request_time;type:datetime" comment:"费用申请时间"`
	RequestTimeStr        string                      `json:"requestTime" gorm:"-" comment:"费用申请时间"`
	BearDept              models.SimpleDept           `json:"bearDept" gorm:"foreignKey:DepartmentId;references:DeptId"`
	ReqUser               models.SimpleUser           `json:"reqUser" gorm:"foreignKey:ReqUserOpenid;references:OpenId"`
	ReqDept               models.SimpleDept           `json:"reqDept" gorm:"foreignKey:UserDeptId;references:DeptId"`
	MainRecord            models.FeishuApprovalRecord `json:"-" gorm:"foreignKey:InstanceCode;references:InstanceCode"`
	Timeline              models.FeishuTimeLines      `json:"timeline" gorm:"foreignKey:InstanceCode;references:InstanceCode"`
	Attachment            models.FeishuAttachmentList `json:"attachments" gorm:"foreignKey:InstanceCode;references:InstanceCode"`
}

func (FeeRequestLogDetail) TableName() string {
	return "fee_request_log"
}

func (e *FeeRequestLogDetail) AfterFind(db *gorm.DB) error {
	e.RequestTimeStr = e.RequestTime.Format(time.DateTime)
	if len(e.MainRecord.TaskList) > 0 && len(e.Timeline) > 0 {
		timeLines := e.Timeline
		sort.Slice(timeLines, func(i, j int) bool {
			return timeLines[i].ID < timeLines[j].ID
		})
		lastTimeline := timeLines[len(e.Timeline)-1]
		for _, t := range e.MainRecord.TaskList {
			if t.OpenId != "" {
				var user models.SimpleUser
				db.Model(&user).Where("open_id=?", t.OpenId).First(&user)
				nowTask := models.FeishuTimeline{
					ID:           0,
					InstanceCode: e.InstanceCode,
					Type:         t.Type,
					CreateDate:   lastTimeline.CreateDate,
					UserId:       t.UserId,
					OpenId:       t.OpenId,
					TaskId:       t.Id,
					OptUser:      user,
				}
				timeLines = append(timeLines, nowTask)
			}
		}
		e.Timeline = timeLines
	}

	return nil
}
