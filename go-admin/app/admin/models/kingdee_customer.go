package models

import (
	"go-admin/common/models"
	"go-admin/common/utils/dateUtils"

	"gorm.io/gorm"
)

type KingdeeCustomer struct {
	models.Model
	CustId         int64  `gorm:"size:20;" json:"custId"`                                                    // FCustId 金蝶客户ID
	CustomerNumber string `gorm:"size:55;" json:"customerNumber" excel:"客户编码,sort:1,width:20,required:true"` // FNumber 客户编码
	CustomerName   string `gorm:"size:55;" json:"customerName" excel:"客户名称,sort:2,width:20,required:true"`   // FName 客户名称
	Country        string `gorm:"size:25;" json:"country"`                                                   // FCountry 国家
	CreateOrgId    int64  `gorm:"size:20;" json:"createOrgId"`                                               // FCreateOrgId 创建组织ID
	UseOrgId       int64  `gorm:"size:20;" json:"useOrgId"`                                                  // FUseOrgId 使用组织ID
	GroupId        int64  `gorm:"size:20;" json:"groupId"`                                                   // FGroup 客户分组ID
	GroupName      string `gorm:"size:55;" json:"groupName"`                                                 // FName 客户分组名称
	GroupNumber    string `gorm:"size:55;" json:"groupNumber"`                                               // FNumber 客户分组编码
	CustomerStatus string `gorm:"size:25;" json:"customerStatus"`                                            // FDocumentStatus 单据状态(A:创建,B:审核中,C:已审核,D:重新审核,Z:暂存)
	ForbidStatus   string `gorm:"size:25;" json:"forbidStatus"`                                              // FForbidStatus 禁用状态(A:启用,B:禁用)
	CreateDate     string `gorm:"size:25;" json:"createDate"`                                                // FCreateDate 金蝶创建日期
	ModifyDate     string `gorm:"size:25;" json:"modifyDate"`                                                // FModifyDate 金蝶修改日期
	Remark         string `gorm:"size:555;" json:"remark" excel:"备注,sort:4,width:20,required:true"`          // 备注
	DeptId         int64  `gorm:"size:20;" json:"deptId"`                                                    // 归属部门ID，sys_dept.dept_id
	CostId         int64  `gorm:"size:20;" json:"costId"`                                                    // 成本中心ID，cost_center_info.id
	models.ControlBy
	models.ModelTime
	UseOrgName string `gorm:"-" json:"useOrgName" excel:"使用组织名称,sort:3,width:20,required:true"` // 使用组织名称
}

func (*KingdeeCustomer) TableName() string {
	return "kingdee_customer"
}

func (e *KingdeeCustomer) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *KingdeeCustomer) GetId() interface{} {
	return e.Id
}

// 时间转换
func (e *KingdeeCustomer) ParseTime() (err error) {
	e.CreateDate = dateUtils.ParseDate(e.CreateDate, dateUtils.PossibleLayouts[26])
	e.ModifyDate = dateUtils.ParseDate(e.ModifyDate, dateUtils.PossibleLayouts[26])
	return
}

func (e *KingdeeCustomer) BeforeCreate(_ *gorm.DB) error {
	return e.ParseTime()
}

func (e *KingdeeCustomer) BeforeUpdate(_ *gorm.DB) error {
	return e.ParseTime()
}
