package audit

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	TitleKey         = "auditTitle"
	BusinessTypeKey  = "auditBusinessType"
	BusinessTypesKey = "auditBusinessTypes"
	MethodKey        = "auditMethod"
	OperatorTypeKey  = "auditOperatorType"
	RemarkKey        = "auditRemark"
	OperParamKey     = "auditOperParam"
	JSONResultKey    = "auditJSONResult"
)

const (
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionStatus   = "status"
	ActionPassword = "password"
	ActionStart    = "start"
	ActionStop     = "stop"
	ActionRun      = "run"
	ActionApprove  = "approve"
	ActionReject   = "reject"
	ActionWithdraw = "withdraw"
)

const (
	CategorySystemSettings = "system-settings"
	CategoryGenerator      = "generator"
	CategoryRole           = "role"
	CategoryMenu           = "menu"
	CategoryUser           = "user"
	CategoryDept           = "dept"
	CategoryPost           = "post"
	CategoryAPI            = "api"
	CategoryDictType       = "dict-type"
	CategoryDictData       = "dict-data"
	CategoryJob            = "job"
	CategoryWorkflow       = "workflow"
	CategoryModule         = "module"
)

const (
	OperatorManage = "MANAGE"
)

type Meta struct {
	Title         string
	BusinessType  string
	BusinessTypes string
	Method        string
	OperatorType  string
	Remark        string
	OperParam     string
	JSONResult    string
}

func Set(c *gin.Context, meta Meta) {
	if meta.Title != "" {
		c.Set(TitleKey, LimitText(meta.Title, 255))
	}
	if meta.BusinessType != "" {
		c.Set(BusinessTypeKey, LimitText(meta.BusinessType, 128))
	}
	if meta.BusinessTypes != "" {
		c.Set(BusinessTypesKey, LimitText(meta.BusinessTypes, 128))
	}
	if meta.Method != "" {
		c.Set(MethodKey, LimitText(meta.Method, 128))
	}
	if meta.OperatorType != "" {
		c.Set(OperatorTypeKey, LimitText(meta.OperatorType, 128))
	}
	if meta.Remark != "" {
		c.Set(RemarkKey, LimitText(meta.Remark, 255))
	}
	if meta.OperParam != "" {
		c.Set(OperParamKey, LimitText(meta.OperParam, 2048))
	}
	if meta.JSONResult != "" {
		c.Set(JSONResultKey, LimitText(meta.JSONResult, 255))
	}
}

func Summary(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "；")
}

func KV(label string, value interface{}) string {
	valueText := strings.TrimSpace(fmt.Sprint(value))
	if label == "" || valueText == "" || valueText == "<nil>" || valueText == "[]" {
		return ""
	}
	return fmt.Sprintf("%s: %s", label, valueText)
}

func Count(label string, count int) string {
	if count <= 0 {
		return ""
	}
	return fmt.Sprintf("%s: %d", label, count)
}

func LimitText(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
