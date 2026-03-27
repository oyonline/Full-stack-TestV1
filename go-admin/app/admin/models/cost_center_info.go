package models

import (
	"gorm.io/gorm"
	"time"

	"go-admin/common/models"
)

type CostCenterInfo struct {
	models.Model

	CostCenterName string    `json:"costCenterName" gorm:"type:varchar(150);comment:成本中心名称" excel:"成本中心名称(必填),sort:1,width:20,required:true" compare:"成本中心名称"`
	CostCenterCode string    `json:"costCenterCode" gorm:"type:varchar(150);comment:成本中心编码" excel:"成本中心编码(必填),sort:2,width:20,required:true" compare:"成本中心编码"`
	CostCenterType int       `json:"costCenterType" gorm:"type:tinyint(1);comment:成本中心类型(1=事业部|2=成本中心|3=费用类别)"`
	DeptId         int64     `json:"deptId" gorm:"type:bigint(20);comment:上级部门" compare:"上级部门ID"`
	Description    string    `json:"description" gorm:"type:varchar(500);comment:描述备注" excel:"描述备注,sort:6,width:20" compare:"描述备注"`
	Status         int       `json:"status" gorm:"type:tinyint(1);comment:状态(1=停用|2=启用|3=待启用)" compare:"状态"`
	EffectiveDate  time.Time `json:"effectiveDate" gorm:"type:date;comment:生效日期" excel:"生效日期(必填),sort:3,width:20,required:true" compare:"生效日期"`
	StopDate       time.Time `json:"stopDate" gorm:"type:date;comment:停用日期"`
	models.ModelTime
	models.ControlBy
	//业务用字段
	GroupNameList    []GroupNameInfoData `json:"groupNameList" gorm:"-"`
	CustomerInfoList []CustomerInfoData  `json:"customerInfoList" gorm:"-"`
	EffectiveDateStr string              `json:"effectiveDateStr" gorm:"-"`
	StopDateStr      string              `json:"stopDateStr" gorm:"-"`
	CreatedAtStr     string              `json:"createdAtStr" gorm:"-"`
	UpdatedAtStr     string              `json:"updatedAtStr" gorm:"-"`
	//比较数据用
	GroupIds         string `json:"groupIds" gorm:"-" compare:"客户分组IDs"`
	GroupNameStrList string `json:"groupNameStrList" gorm:"-" excel:"客户分组名称,sort:5,width:20" compare:"客户分组名称"`
	DeptPathName     string `json:"deptPathName" gorm:"-" excel:"上级部门全路径,sort:4,width:20" compare:"上级部门全路径"`
}

func (CostCenterInfo) TableName() string {
	return "cost_center_info"
}

func (e *CostCenterInfo) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostCenterInfo) GetId() interface{} {
	return e.Id
}

type CustomerInfoData struct {
	CostCenterInfoId int64  `json:"costCenterInfoId"`
	GroupId          int64  `json:"groupId"`
	GroupName        string `json:"groupName"`
	GroupNumber      string `json:"groupNumber"`
	CustomerName     string `json:"customerName"`
}

type GroupNameInfoData struct {
	CostCenterInfoId int64  `json:"costCenterInfoId"`
	GroupId          int64  `json:"groupId"`
	GroupName        string `json:"groupName"`
}

func (e *CostCenterInfo) AfterFind(_ *gorm.DB) error {
	e.EffectiveDateStr = e.EffectiveDate.Format(time.DateOnly)
	e.StopDateStr = e.StopDate.Format(time.DateOnly)
	e.CreatedAtStr = e.CreatedAt.Format(time.DateTime)
	e.UpdatedAtStr = e.UpdatedAt.Format(time.DateTime)
	return nil
}
