package models

import "go-admin/common/models"

type SysNotice struct {
	models.Model
	Title   string `json:"title" gorm:"size:256;comment:标题"`
	Content string `json:"content" gorm:"type:text;comment:内容"`
	Type    string `json:"type" gorm:"size:64;comment:类型"`
	Status  string `json:"status" gorm:"size:4;comment:状态"`
	Sort    int    `json:"sort" gorm:"comment:排序"`
	Remark  string `json:"remark" gorm:"size:256;comment:备注"`
	models.ControlBy
	models.ModelTime
}

func (*SysNotice) TableName() string {
	return "sys_notice"
}

func (e *SysNotice) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SysNotice) GetId() interface{} {
	return e.Id
}
