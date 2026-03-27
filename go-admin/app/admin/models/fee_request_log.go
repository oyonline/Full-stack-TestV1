package models

import (
	"go-admin/cmd/migrate/migration/models"
	"time"

	"gorm.io/gorm"
)

type FeeRequestLog struct {
	models.Model
	InstanceCode    string    `json:"instanceCode" gorm:"column:instance_code;type:varchar(64);comment:实例Code"`
	ReqUserOpenid   string    `json:"-" gorm:"column:req_user_openid;type:varchar(64)" comment:"申请人openid"`
	Status          string    `json:"status" gorm:"column:status;type:varchar(20)"`
	OrgCode         string    `json:"orgCode" gorm:"column:org_code;type:varchar(64)" comment:"付款主体"`
	Currency        string    `json:"currency" gorm:"column:currency;type:varchar(10)"`
	YearsMonth      string    `json:"yearsMonth" gorm:"column:budget_years_month;type:varchar(10)" comment:"预算年月"`
	BudgetAmount    float64   `json:"budgetAmount" gorm:"column:budget_amount;type:decimal(10,2)" comment:"预算总额"`
	BudgetUsed      float64   `json:"budgetUsed" gorm:"column:budget_used;type:decimal(10,2)" comment:"已使用预算"`
	RequestAmount   float64   `json:"requestAmount" gorm:"column:request_amount;type:decimal(10,2)" comment:"申请金额"`
	CostCenterId    int64     `json:"costCenterId" gorm:"column:cost_center_info_id;type:bigint" comment:"成本中心ID"`
	CostCenterName  string    `json:"costCenterName" gorm:"column:cost_center_name;type:varchar(200)" comment:"成本中心名称"`
	BudgetVersionId int64     `json:"budgetVersionId" gorm:"column:budget_version_id;type:bigint" comment:"预算版本ID"`
	GroupName       string    `json:"groupName" gorm:"column:group_name;type:varchar(200)" comment:"客户分组名称"`
	GroupCode       string    `json:"groupCode" gorm:"column:group_number;type:varchar(100)" comment:"客服分组编码"`
	BudgetDetailId  int64     `json:"budgetDetailId" gorm:"column:budget_detail_id;type:bigint" comment:"预算详情ID"`
	FeeCode         string    `json:"feeCode" gorm:"column:fee_code;type:varchar(50)" comment:"费用编码"`
	FeeName         string    `json:"feeName" gorm:"column:fee_name;type:varchar(200)" comment:"费用名称"`
	UserDeptId      int       `json:"-" orm:"column:user_dept_id;type:bigint" comment:"申请部门ID"`
	DepartmentId    int       `json:"-" gorm:"column:dept_id;type:bigint" comment:"承担部门ID"`
	RequestTime     time.Time `json:"-" gorm:"column:request_time;type:datetime" comment:"费用申请时间"`
	models.ControlBy
	models.ModelTime
}

func (FeeRequestLog) TableName() string {
	return "fee_request_log"
}
func (e *FeeRequestLog) BeforeCreate(_ *gorm.DB) error {
	e.RequestTime = time.Now()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return nil
}

func (e *FeeRequestLog) BeforeSave(_ *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}
