package models

import common "go-admin/common/models"

type ModuleRegistry struct {
	ModuleId       int    `json:"moduleId" gorm:"primaryKey;autoIncrement;comment:模块ID"`
	ModuleKey      string `json:"moduleKey" gorm:"size:64;uniqueIndex;comment:模块编码"`
	ModuleName     string `json:"moduleName" gorm:"size:128;comment:模块名称"`
	RouteBase      string `json:"routeBase" gorm:"size:128;uniqueIndex;comment:路由前缀"`
	MenuRootCode   string `json:"menuRootCode" gorm:"size:128;uniqueIndex;comment:根菜单编码"`
	Status         string `json:"status" gorm:"size:4;default:2;comment:状态"`
	Sort           int    `json:"sort" gorm:"default:0;comment:排序"`
	Remark         string `json:"remark" gorm:"size:255;comment:备注"`
	PermissionHint string `json:"permissionHint" gorm:"size:128;comment:权限前缀建议"`
	common.ControlBy
	common.ModelTime
}

func (*ModuleRegistry) TableName() string {
	return "module_registry"
}
