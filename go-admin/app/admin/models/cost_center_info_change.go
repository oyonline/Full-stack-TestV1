package models

import (
	"encoding/json"
	"time"

	"go-admin/common/models"
)

type CostCenterInfoChange struct {
	models.Model

	ChangeOrder      string          `json:"changeOrder" gorm:"type:varchar(150);comment:变更单号"`
	ChangeType       string          `json:"changeType" gorm:"type:varchar(150);comment:变更类型"`
	Status           int             `json:"status" gorm:"type:tinyint(1);comment:状态(1=已生效，2=未生效)"`
	EffectiveDate    time.Time       `json:"effectiveDate" gorm:"type:date;comment:生效日期"`
	VersionNumber    string          `json:"versionNumber" gorm:"type:varchar(150);comment:版本号"`
	ChangeDetails    json.RawMessage `json:"changeDetails" gorm:"type:json;comment:变更内容"`
	CostCenterInfoId int64           `json:"costCenterInfoId" gorm:"type:bigint(20);comment:成本中心ID"`
	models.ModelTime
	models.ControlBy
}

func (CostCenterInfoChange) TableName() string {
	return "cost_center_info_change"
}

func (e *CostCenterInfoChange) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *CostCenterInfoChange) GetId() interface{} {
	return e.Id
}
