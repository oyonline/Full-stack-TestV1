package dto

import (
	"strings"

	"go-admin/app/platform/models"
	common "go-admin/common/models"
	cDto "go-admin/common/dto"
)

type ModuleRegistryGetPageReq struct {
	cDto.Pagination `search:"-"`

	ModuleKey    string `form:"moduleKey"`
	ModuleName   string `form:"moduleName"`
	Status       string `form:"status"`
	MenuRootCode string `form:"menuRootCode"`
}

type ModuleRegistryGetReq struct {
	Id int `uri:"id"`
}

func (m *ModuleRegistryGetReq) GetId() interface{} {
	return m.Id
}

type ModuleRegistryInsertReq struct {
	ModuleKey    string `json:"moduleKey" binding:"required"`
	ModuleName   string `json:"moduleName" binding:"required"`
	RouteBase    string `json:"routeBase" binding:"required"`
	MenuRootCode string `json:"menuRootCode" binding:"required"`
	Status       string `json:"status"`
	Sort         int    `json:"sort"`
	Remark       string `json:"remark"`
	common.ControlBy
}

func (m *ModuleRegistryInsertReq) Normalize() {
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.ModuleName = strings.TrimSpace(m.ModuleName)
	m.RouteBase = strings.TrimSpace(m.RouteBase)
	m.MenuRootCode = strings.TrimSpace(m.MenuRootCode)
	m.Remark = strings.TrimSpace(m.Remark)
	if m.Status == "" {
		m.Status = "2"
	}
}

func (m *ModuleRegistryInsertReq) Generate(model *models.ModuleRegistry) {
	model.ModuleKey = m.ModuleKey
	model.ModuleName = m.ModuleName
	model.RouteBase = m.RouteBase
	model.MenuRootCode = m.MenuRootCode
	model.Status = m.Status
	model.Sort = m.Sort
	model.Remark = m.Remark
	model.PermissionHint = m.ModuleKey
}

type ModuleRegistryUpdateReq struct {
	ModuleId      int    `json:"moduleId" binding:"required"`
	ModuleKey    string `json:"moduleKey" binding:"required"`
	ModuleName   string `json:"moduleName" binding:"required"`
	RouteBase    string `json:"routeBase" binding:"required"`
	MenuRootCode string `json:"menuRootCode" binding:"required"`
	Status       string `json:"status"`
	Sort         int    `json:"sort"`
	Remark       string `json:"remark"`
	common.ControlBy
}

func (m *ModuleRegistryUpdateReq) Normalize() {
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.ModuleName = strings.TrimSpace(m.ModuleName)
	m.RouteBase = strings.TrimSpace(m.RouteBase)
	m.MenuRootCode = strings.TrimSpace(m.MenuRootCode)
	m.Remark = strings.TrimSpace(m.Remark)
	if m.Status == "" {
		m.Status = "2"
	}
}

func (m *ModuleRegistryUpdateReq) Generate(model *models.ModuleRegistry) {
	model.ModuleId = m.ModuleId
	model.ModuleKey = m.ModuleKey
	model.ModuleName = m.ModuleName
	model.RouteBase = m.RouteBase
	model.MenuRootCode = m.MenuRootCode
	model.Status = m.Status
	model.Sort = m.Sort
	model.Remark = m.Remark
	model.PermissionHint = m.ModuleKey
}

type ModuleRegistryDeleteReq struct {
	Id int `uri:"id" binding:"required"`
}
