package dto

import (
	"strings"

	cDto "go-admin/common/dto"
)

const AttachmentStorageLocal = "local"

type AttachmentGetPageReq struct {
	cDto.Pagination `search:"-"`

	ModuleKey    string `form:"moduleKey" search:"-"`
	BusinessType string `form:"businessType" search:"-"`
	BusinessId   string `form:"businessId" search:"-"`
}

func (m *AttachmentGetPageReq) Normalize() {
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.BusinessType = strings.TrimSpace(m.BusinessType)
	m.BusinessId = strings.TrimSpace(m.BusinessId)
}

type AttachmentUploadReq struct {
	ModuleKey    string `form:"moduleKey" binding:"required"`
	BusinessType string `form:"businessType" binding:"required"`
	BusinessId   string `form:"businessId" binding:"required"`
	BusinessNo   string `form:"businessNo"`
}

func (m *AttachmentUploadReq) Normalize() {
	m.ModuleKey = strings.TrimSpace(m.ModuleKey)
	m.BusinessType = strings.TrimSpace(m.BusinessType)
	m.BusinessId = strings.TrimSpace(m.BusinessId)
	m.BusinessNo = strings.TrimSpace(m.BusinessNo)
}

type AttachmentGetReq struct {
	Id int `uri:"id" binding:"required"`
}
