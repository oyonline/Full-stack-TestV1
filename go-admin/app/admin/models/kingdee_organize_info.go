package models

import (
	"go-admin/common/models"
	"go-admin/common/utils/dateUtils"

	"gorm.io/gorm"
)

type KingdeeOrganizeInfo struct {
	models.Model
	Forgid          int64  `gorm:"size:20;" json:"FOrgID"`          // FOrgID 金蝶实体主键
	Fdocumentstatus string `gorm:"size:25;" json:"FDocumentStatus"` // FDocumentStatus 单据状态(A:创建,B:审核中,C:已审核,D:重新审核,Z:暂存)
	Fforbidstatus   string `gorm:"size:25;" json:"FForbidStatus"`   // FForbidStatus 禁用状态(A:启用,B:禁用)
	Fname           string `gorm:"size:255;" json:"FName"`          // FName 名称
	Fnumber         string `gorm:"size:255;" json:"FNumber"`        // FNumber 编码
	Fdescription    string `gorm:"size:255;" json:"FDescription"`   // FDescription 描述
	Fcreatedate     string `gorm:"size:25;" json:"FCreateDate"`     // FCreateDate 创建日期
	Fmodifydate     string `gorm:"size:25;" json:"FModifyDate"`     // FModifyDate 修改日期
	Fcontact        string `gorm:"size:255;" json:"FContact"`       // FContact 联系人
	Forgformid      string `gorm:"size:255;" json:"FOrgFormID"`     // FOrgFormID 形态
	Faddress        string `gorm:"size:555;" json:"FAddress"`       // FAddress 地址
	Ftel            string `gorm:"size:25;" json:"FTel"`            // FTel 联系电话
	Facctorgtype    string `gorm:"size:25;" json:"FAcctOrgType"`    // FAcctOrgType 核算组织类型
	Fparentid       int64  `gorm:"size:20;" json:"FParentID"`       // FParentID 所属法人
	Fisbusinessorg  bool   `gorm:"size:1;" json:"FIsBusinessOrg"`   // FIsBusinessOrg 业务组织
	Fisaccountorg   bool   `gorm:"size:1;" json:"FIsAccountOrg"`    // FIsAccountOrg 核算组织
	FPxzoScc        string `gorm:"size:255;" json:"F.PXZO.SCC"`     // F_PXZO_SCC 社会信用代码
	FPxzoEn         string `gorm:"size:255;" json:"F.PXZO.EN"`      // F_PXZO_EN 英文名称
	FPxzoEd         string `gorm:"size:255;" json:"F.PXZO.ED"`      // F_PXZO_ED 英文地址
	models.ModelTime
}

func (*KingdeeOrganizeInfo) TableName() string {
	return "kingdee_organize_info"
}

func (e *KingdeeOrganizeInfo) GetId() interface{} {
	return e.Id
}

// 时间转换
func (e *KingdeeOrganizeInfo) ParseTime() (err error) {
	e.Fcreatedate = dateUtils.ParseDate(e.Fcreatedate, dateUtils.PossibleLayouts[26])
	e.Fmodifydate = dateUtils.ParseDate(e.Fmodifydate, dateUtils.PossibleLayouts[26])
	return
}

func (e *KingdeeOrganizeInfo) BeforeCreate(_ *gorm.DB) error {
	return e.ParseTime()
}

func (e *KingdeeOrganizeInfo) BeforeUpdate(_ *gorm.DB) error {
	return e.ParseTime()
}
