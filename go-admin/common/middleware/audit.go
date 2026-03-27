package middleware

import (
	"github.com/gin-gonic/gin"

	"go-admin/common/audit"
)

const (
	AuditActionCreate   = audit.ActionCreate
	AuditActionUpdate   = audit.ActionUpdate
	AuditActionDelete   = audit.ActionDelete
	AuditActionStatus   = audit.ActionStatus
	AuditActionPassword = audit.ActionPassword
	AuditActionStart    = audit.ActionStart
	AuditActionStop     = audit.ActionStop
	AuditActionRun      = audit.ActionRun
)

const (
	AuditCategorySystemSettings = audit.CategorySystemSettings
	AuditCategoryGenerator      = audit.CategoryGenerator
	AuditCategoryRole           = audit.CategoryRole
	AuditCategoryMenu           = audit.CategoryMenu
	AuditCategoryUser           = audit.CategoryUser
	AuditCategoryDept           = audit.CategoryDept
	AuditCategoryPost           = audit.CategoryPost
	AuditCategoryAPI            = audit.CategoryAPI
	AuditCategoryDictType       = audit.CategoryDictType
	AuditCategoryDictData       = audit.CategoryDictData
	AuditCategoryJob            = audit.CategoryJob
)

const (
	AuditOperatorManage = audit.OperatorManage
)

type AuditMeta = audit.Meta

func SetAuditMeta(c *gin.Context, meta AuditMeta) {
	audit.Set(c, meta)
}

func AuditSummary(parts ...string) string {
	return audit.Summary(parts...)
}

func AuditKV(label string, value interface{}) string {
	return audit.KV(label, value)
}

func AuditCount(label string, count int) string {
	return audit.Count(label, count)
}
