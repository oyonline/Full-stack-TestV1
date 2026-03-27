package models

import "go-admin/common/models"

type KingdeeDept struct {
	Id           int64  `gorm:"size:20;" json:"FDeptId"`         // FDeptId 部门ID
	ParentId     int64  `gorm:"size:20;" json:"FParentId"`       // FParentId 上级部门ID
	DeptName     string `gorm:"size:255;" json:"FName"`          // FName 部门名称
	DeptNumber   string `gorm:"size:25;" json:"FNumber"`         // FNumber 部门编码
	DeptStatus   string `gorm:"size:25;" json:"FDocumentStatus"` // FDocumentStatus 单据状态(A:创建,B:审核中,C:已审核,D:重新审核,Z:暂存)
	ForbidStatus string `gorm:"size:25;" json:"FForbidStatus"`   // FForbidStatus 禁用状态(A:启用,B:禁用)
	UseOrgId     int64  `gorm:"size:20;" json:"FUseOrgId"`       // FUseOrgId 使用组织
	models.ModelTime
}

func (*KingdeeDept) TableName() string {
	return "kingdee_dept"
}

func (e *KingdeeDept) GetId() interface{} {
	return e.Id
}
