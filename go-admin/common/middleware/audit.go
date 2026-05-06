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
	AuditActionApprove  = audit.ActionApprove
	AuditActionReject   = audit.ActionReject
	AuditActionWithdraw = audit.ActionWithdraw
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
	AuditCategoryWorkflow       = audit.CategoryWorkflow
	AuditCategoryModule         = audit.CategoryModule
)

const (
	AuditOperatorManage = audit.OperatorManage
)

type AuditMeta = audit.Meta

// AuditEntry 与 AuditTarget 是业务操作日志最小契约的入口类型，详见 common/audit.Entry。
type (
	AuditEntry  = audit.Entry
	AuditTarget = audit.Target
)

func SetAuditMeta(c *gin.Context, meta AuditMeta) {
	audit.Set(c, meta)
}

// AuditLog 按最小契约（actor/action/target/before/after/timestamp）写一条业务操作日志。
// actor / timestamp 由中间件后续填充，调用方只需提供 entry 中的业务字段。
func AuditLog(c *gin.Context, entry AuditEntry) {
	audit.Log(c, entry)
}

// AuditLogCreate / AuditLogUpdate / AuditLogDelete 是常见动作的便捷写法。
func AuditLogCreate(c *gin.Context, title string, target AuditTarget, after interface{}, method string) {
	audit.LogCreate(c, title, target, after, method)
}

func AuditLogUpdate(c *gin.Context, title string, target AuditTarget, before, after interface{}, method string) {
	audit.LogUpdate(c, title, target, before, after, method)
}

func AuditLogDelete(c *gin.Context, title string, target AuditTarget, before interface{}, method string) {
	audit.LogDelete(c, title, target, before, method)
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
