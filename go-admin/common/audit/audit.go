package audit

import (
	"encoding/json"
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

// Target 标识被操作的业务实体。
type Target struct {
	// Type 实体分类，对应 sys_opera_log.business_types（user/role/post/...）。
	// 通常使用 audit.Category* 常量。
	Type string
	// ID 主键值。允许 int / int64 / string / []int 等 GORM 主键查询支持的类型。
	ID interface{}
	// Label 展示用中文标签（如 username、roleName），便于运营排查。可空。
	Label string
}

// Entry 是业务操作日志的最小契约。每条日志必须能回答 6 个问题：
//
//	actor      谁操作的     —— 由中间件 LoggerToFile 自动从 gin Context 的登录态填充
//	                          (operName / createBy 写入 sys_opera_log；deptName 取决于
//	                          后续中间件补齐进度)
//	action     做了什么      —— Entry.Action（audit.Action* 常量）
//	target     操作了什么    —— Entry.Target.Type/ID/Label
//	before     变更前是什么  —— Entry.Before（update/delete 必填）
//	after      变更后是什么  —— Entry.After（create/update 必填）
//	timestamp  什么时候      —— 由中间件 SetDBOperLog 写入 operTime / createdAt
//
// Log 把 Entry 编码进 gin.Context，最终由 LoggerToFile -> SaveOperaLog 落到 sys_opera_log。
// 不修改 sys_opera_log schema：
//
//	Title         -> Entry.Title
//	BusinessType  -> Entry.Action
//	BusinessTypes -> Entry.Target.Type
//	Method        -> Entry.Method
//	OperatorType  -> Entry.Operator（默认 OperatorManage）
//	Remark        -> 由 target.label / target.id 摘要生成（人类可读）
//	OperParam     -> 完整 JSON（机器可读，含 target/before/after/extra）
type Entry struct {
	Title    string                 // 中文模块标题（如 "岗位管理"），列表页展示用
	Action   string                 // audit.Action* 常量
	Target   Target                 // 操作目标
	Method   string                 // 服务端方法路径，如 admin.sysPost.update
	Operator string                 // 操作员类型，默认 OperatorManage
	Before   interface{}            // 变更前快照（update/delete 必填）
	After    interface{}            // 变更后快照（create/update 必填）
	Extra    map[string]interface{} // 扩展字段（可选）
}

// Log 把 Entry 编码进 gin.Context，等待中间件落库。
// 当 c 为 nil（例如脚本调用）时直接 no-op，不会 panic。
func Log(c *gin.Context, e Entry) {
	if c == nil {
		return
	}
	Set(c, BuildMeta(e))
}

// BuildMeta 把 Entry 转换为底层 Meta（不写入 Context）。便于外部测试 / 自定义编码路径。
func BuildMeta(e Entry) Meta {
	payload := map[string]interface{}{
		"target": map[string]interface{}{
			"type":  e.Target.Type,
			"id":    e.Target.ID,
			"label": e.Target.Label,
		},
	}
	if e.Before != nil {
		payload["before"] = e.Before
	}
	if e.After != nil {
		payload["after"] = e.After
	}
	if len(e.Extra) > 0 {
		payload["extra"] = e.Extra
	}
	operParam := encodeJSON(payload)

	operator := e.Operator
	if operator == "" {
		operator = OperatorManage
	}

	return Meta{
		Title:         e.Title,
		BusinessType:  e.Action,
		BusinessTypes: e.Target.Type,
		Method:        e.Method,
		OperatorType:  operator,
		Remark:        targetRemark(e.Target),
		OperParam:     operParam,
	}
}

// LogCreate 是 create 场景的便捷写法。after 是新建后的实体（可只传必要字段）。
func LogCreate(c *gin.Context, title string, target Target, after interface{}, method string) {
	Log(c, Entry{
		Title:  title,
		Action: ActionCreate,
		Target: target,
		After:  after,
		Method: method,
	})
}

// LogUpdate 是 update 场景的便捷写法。before/after 都应提供变更字段子集即可。
func LogUpdate(c *gin.Context, title string, target Target, before, after interface{}, method string) {
	Log(c, Entry{
		Title:  title,
		Action: ActionUpdate,
		Target: target,
		Before: before,
		After:  after,
		Method: method,
	})
}

// LogDelete 是 delete 场景的便捷写法。before 是删除前的实体（如有）。
func LogDelete(c *gin.Context, title string, target Target, before interface{}, method string) {
	Log(c, Entry{
		Title:  title,
		Action: ActionDelete,
		Target: target,
		Before: before,
		Method: method,
	})
}

func targetRemark(t Target) string {
	parts := make([]string, 0, 2)
	if t.Label != "" {
		parts = append(parts, KV("目标", t.Label))
	}
	if t.ID != nil {
		// 对 nil-interface-with-empty-slice 也能容忍：KV 内部会忽略 "[]"
		parts = append(parts, KV("ID", t.ID))
	}
	return Summary(parts...)
}

func encodeJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
