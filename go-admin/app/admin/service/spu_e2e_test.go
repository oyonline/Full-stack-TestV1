package service

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http/httptest"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/cmd/migrate/migration"
	migModels "go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
	platformModels "go-admin/app/platform/models"
	platformService "go-admin/app/platform/service"
	platformDto "go-admin/app/platform/service/dto"
	"go-admin/common/actions"
	"go-admin/common/audit"
	"go-admin/common/middleware"
	_ "go-admin/cmd/migrate/migration/version" // registers all migrations via init()
)

// C4-D 端到端验收：覆盖 SPU 完整提交→审批→回写流程 + dataScope 过滤 + 审计契约。
//
// 这套测试不替代单元 spu_test.go / spu_workflow_handler_test.go，而是把
// admin.spu 服务和 platform.workflow 服务串起来跑一遍，验证：
//
//   - SPU.SubmitForReview ↔ Workflow.Start 协同（创建 wf_instance + binding + 第一个 task）
//   - Workflow.Approve 终态触发 onSpuWorkflowTerminal → SPU.status 写回
//   - Workflow.Reject 终态触发 → SPU.status=Rejected
//   - 驳回后重提，新建 wf_instance（旧 binding 被替换）
//   - dataScope 按 create_by 过滤 SPU 列表
//   - apis/spu.go + platform/apis/workflow.go 两条 API 路径产出的审计 Method 名稳定

// e2eSpuSetup 构造一个跑通 admin + platform workflow 联调的 sqlite 实例。
//
// 与 newTestSpu 的区别：
//   - 显式启用 dataScope（test 内通过 cleanup 复原）
//   - 同时迁移 SysDept、确保 dataScope=3/4 的 dept_path 子查询能跑
type e2eSpuFixture struct {
	Spu     *Spu
	WF      *platformService.Workflow
	DB      *gorm.DB
	Cleanup func()
}

func newE2ESpu(t *testing.T) *e2eSpuFixture {
	t.Helper()
	return newE2ESpuWithDP(t, false)
}

// newE2ESpuWithDP 与 newE2ESpu 同构，但会按 enableDP 开关 EnableDP 配置。
// 默认（dataScope-disabled）模式让其他 e2e 测试可以传 nil DataPermission，
// 避免每个 test 都构造一份伪 DP；仅 dataScope 专项测试启用 DP。
func newE2ESpuWithDP(t *testing.T, enableDP bool) *e2eSpuFixture {
	t.Helper()

	prevEnable := config.ApplicationConfig.EnableDP
	config.ApplicationConfig.EnableDP = enableDP
	cleanup := func() {
		config.ApplicationConfig.EnableDP = prevEnable
	}

	// 用 file::memory:?cache=shared 让 gorm 连接池里所有 connection 共享同一份 in-memory DB。
	// :memory: 每个连接独占一份，会导致 AutoMigrate 写入的表只对发起迁移那条 connection 可见，
	// 后续 Workflow.Approve 走另一条 connection 时报 "no such table"（newTestSpu 的同样选择）。
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		cleanup()
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Spu{},
		&models.Sku{},
		&models.SysUser{},
		&models.SysRole{},
		&models.SysDept{},
		&platformModels.ModuleRegistry{},
		&platformModels.WorkflowDefinition{},
		&platformModels.WorkflowDefinitionNode{},
		&platformModels.WorkflowInstance{},
		&platformModels.WorkflowTask{},
		&platformModels.WorkflowActionLog{},
		&platformModels.WorkflowBusinessBinding{},
	); err != nil {
		cleanup()
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_role_dept (role_id INTEGER NOT NULL, dept_id INTEGER NOT NULL)`).Error; err != nil {
		cleanup()
		t.Fatalf("create sys_role_dept: %v", err)
	}
	// shared-cache 让所有连接看见同一份 in-memory DB，但也意味着 *跨测试* 共享。
	// 清空所有相关表，避免上一轮测试残留干扰。
	for _, tbl := range []string{
		"spu", "sku", "sys_user", "sys_role", "sys_dept", "sys_role_dept",
		"module_registry", "wf_definition", "wf_definition_node", "wf_instance",
		"wf_task", "wf_action_log", "wf_business_binding",
	} {
		if err := db.Exec("DELETE FROM " + tbl).Error; err != nil {
			cleanup()
			t.Fatalf("clean %s: %v", tbl, err)
		}
	}

	s := &Spu{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	wf := &platformService.Workflow{Service: s.Service}

	return &e2eSpuFixture{Spu: s, WF: wf, DB: db, Cleanup: cleanup}
}

// seedRoleDefinition 写入 product_admin 角色 + 'spu_create_review' 流程定义 + role 类型审批节点。
// 返回 role_id（供调用方在 makeRoleAuthCtx 注入 claimRoleIDs 使用）和定义 ID。
func seedRoleDefinition(t *testing.T, db *gorm.DB) (roleID int, definitionID int) {
	t.Helper()
	role := models.SysRole{RoleId: 99, RoleName: "产品管理员", RoleKey: "product_admin", Status: "2"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	def := platformModels.WorkflowDefinition{
		DefinitionKey:  SpuDefaultDefinitionKey,
		DefinitionName: "SPU 创建审核",
		ModuleKey:      SpuModuleKey,
		BusinessType:   SpuBusinessType,
		Status:         "2",
		Version:        1,
	}
	if err := db.Create(&def).Error; err != nil {
		t.Fatalf("seed def: %v", err)
	}
	node := platformModels.WorkflowDefinitionNode{
		DefinitionId:  def.DefinitionId,
		NodeKey:       "approve_1",
		NodeName:      "产品管理员审批",
		NodeType:      platformDto.WorkflowNodeTypeApprove,
		Sort:          1,
		ApproverType:  platformDto.WorkflowApproverRole,
		ApproverValue: "99",
		ApproverName:  "产品管理员",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	return role.RoleId, def.DefinitionId
}

// makeRoleAuthCtx 是 makeAuthCtx 的扩展：附带 roleIds claim，使得
// canProcessTask(role) 成立。
func makeRoleAuthCtx(userID int, userName string, roleIDs []int) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)
	intf := make([]interface{}, len(roleIDs))
	for i, r := range roleIDs {
		intf[i] = float64(r)
	}
	claims := jwt.MapClaims{
		"identity": float64(userID),
		"nice":     userName,
		"roleIds":  intf,
	}
	c.Set(jwt.JwtPayloadKey, claims)
	return c
}

// findPendingTask 取 instance 上唯一的 pending task（驱动 Approve/Reject）。
func findPendingTask(t *testing.T, db *gorm.DB, instanceID int) platformModels.WorkflowTask {
	t.Helper()
	var task platformModels.WorkflowTask
	if err := db.Where("instance_id = ? AND status = ?", instanceID, platformDto.WorkflowTaskPending).
		First(&task).Error; err != nil {
		t.Fatalf("find pending task on instance %d: %v", instanceID, err)
	}
	return task
}

// TestE2E_Spu_Submit_Approve_Writeback 完整跑通：
//
//	插入 SPU → SubmitForReview → Workflow.Approve → SPU.status=Approved + approved_at 已设
//
// 同时验证：
//   - wf_business_binding.workflow_status 推进到 approved
//   - wf_action_log 至少有 'start' 和 'approve' 两条
func TestE2E_Spu_Submit_Approve_Writeback(t *testing.T) {
	f := newE2ESpu(t)
	defer f.Cleanup()
	roleID, _ := seedRoleDefinition(t, f.DB)

	// 提交者
	starter := models.SysUser{UserId: 11, NickName: "operator", Status: "2", RoleId: 50}
	if err := f.DB.Create(&starter).Error; err != nil {
		t.Fatalf("seed starter: %v", err)
	}
	// 审批人（属于 product_admin 角色）
	approver := models.SysUser{UserId: 12, NickName: "approver", Status: "2", RoleId: roleID}
	if err := f.DB.Create(&approver).Error; err != nil {
		t.Fatalf("seed approver: %v", err)
	}

	// 1) 插入 SPU
	insReq := &dto.SpuInsertReq{SpuCode: "E2E-001", SpuName: "e2e product"}
	insReq.SetCreateBy(starter.UserId)
	spuId, err := f.Spu.Insert(insReq)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// 2) 提交审核
	starterCtx := makeRoleAuthCtx(starter.UserId, starter.NickName, []int{starter.RoleId})
	subReq := &dto.SpuSubmitReq{SpuId: spuId, Remark: "go review"}
	instanceID, err := f.Spu.SubmitForReview(starterCtx, nil, subReq)
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}

	// 3) 审批人取 pending task → Approve
	task := findPendingTask(t, f.DB, instanceID)
	approverCtx := makeRoleAuthCtx(approver.UserId, approver.NickName, []int{roleID})
	if _, err := f.WF.Approve(approverCtx, task.TaskId, "ok"); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	// 4) 断言 SPU 状态写回
	var post models.Spu
	if err := f.DB.First(&post, spuId).Error; err != nil {
		t.Fatalf("read spu: %v", err)
	}
	if post.Status != models.SpuStatusApproved {
		t.Fatalf("expected SPU status=Approved(3), got %d", post.Status)
	}
	if post.ApprovedAt == nil {
		t.Fatalf("expected approved_at set after approve")
	}

	// binding 推进到 approved
	var binding platformModels.WorkflowBusinessBinding
	if err := f.DB.Where("instance_id = ?", instanceID).First(&binding).Error; err != nil {
		t.Fatalf("binding: %v", err)
	}
	if binding.WorkflowStatus != platformDto.WorkflowStatusApproved {
		t.Fatalf("binding.WorkflowStatus=%q, want %q", binding.WorkflowStatus, platformDto.WorkflowStatusApproved)
	}

	// action log 至少有 start + approve
	var actionLogs []platformModels.WorkflowActionLog
	if err := f.DB.Where("instance_id = ?", instanceID).Order("log_id ASC").Find(&actionLogs).Error; err != nil {
		t.Fatalf("actions: %v", err)
	}
	gotActions := make([]string, 0, len(actionLogs))
	for _, a := range actionLogs {
		gotActions = append(gotActions, a.Action)
	}
	if len(gotActions) < 2 || gotActions[0] != platformDto.WorkflowActionStart || gotActions[1] != platformDto.WorkflowActionApprove {
		t.Fatalf("expected actions=[start, approve, ...], got %v", gotActions)
	}
}

// TestE2E_Spu_Submit_Reject_Writeback 驳回路径。
func TestE2E_Spu_Submit_Reject_Writeback(t *testing.T) {
	f := newE2ESpu(t)
	defer f.Cleanup()
	roleID, _ := seedRoleDefinition(t, f.DB)

	starter := models.SysUser{UserId: 21, NickName: "operator-r", Status: "2", RoleId: 50}
	if err := f.DB.Create(&starter).Error; err != nil {
		t.Fatalf("seed starter: %v", err)
	}
	approver := models.SysUser{UserId: 22, NickName: "approver-r", Status: "2", RoleId: roleID}
	if err := f.DB.Create(&approver).Error; err != nil {
		t.Fatalf("seed approver: %v", err)
	}

	insReq := &dto.SpuInsertReq{SpuCode: "E2E-RJ", SpuName: "reject me"}
	insReq.SetCreateBy(starter.UserId)
	spuId, err := f.Spu.Insert(insReq)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	starterCtx := makeRoleAuthCtx(starter.UserId, starter.NickName, []int{starter.RoleId})
	instanceID, err := f.Spu.SubmitForReview(starterCtx, nil, &dto.SpuSubmitReq{SpuId: spuId})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	task := findPendingTask(t, f.DB, instanceID)
	approverCtx := makeRoleAuthCtx(approver.UserId, approver.NickName, []int{roleID})
	if _, err := f.WF.Reject(approverCtx, task.TaskId, "信息不全"); err != nil {
		t.Fatalf("Reject: %v", err)
	}

	var post models.Spu
	if err := f.DB.First(&post, spuId).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	if post.Status != models.SpuStatusRejected {
		t.Fatalf("expected SPU status=Rejected(4), got %d", post.Status)
	}
	if post.ApprovedAt != nil {
		t.Fatalf("approved_at should remain nil on reject, got %v", post.ApprovedAt)
	}
}

// TestE2E_Spu_Resubmit_After_Reject 驳回后允许再提交，新 wf_instance + 旧 binding 被替换。
func TestE2E_Spu_Resubmit_After_Reject(t *testing.T) {
	f := newE2ESpu(t)
	defer f.Cleanup()
	roleID, _ := seedRoleDefinition(t, f.DB)

	starter := models.SysUser{UserId: 31, NickName: "operator-rs", Status: "2", RoleId: 50}
	if err := f.DB.Create(&starter).Error; err != nil {
		t.Fatalf("seed starter: %v", err)
	}
	approver := models.SysUser{UserId: 32, NickName: "approver-rs", Status: "2", RoleId: roleID}
	if err := f.DB.Create(&approver).Error; err != nil {
		t.Fatalf("seed approver: %v", err)
	}

	insReq := &dto.SpuInsertReq{SpuCode: "E2E-RS", SpuName: "resubmit me"}
	insReq.SetCreateBy(starter.UserId)
	spuId, err := f.Spu.Insert(insReq)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	starterCtx := makeRoleAuthCtx(starter.UserId, starter.NickName, []int{starter.RoleId})
	approverCtx := makeRoleAuthCtx(approver.UserId, approver.NickName, []int{roleID})

	// 第一次：提交 → 驳回
	firstID, err := f.Spu.SubmitForReview(starterCtx, nil, &dto.SpuSubmitReq{SpuId: spuId})
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	task1 := findPendingTask(t, f.DB, firstID)
	if _, err := f.WF.Reject(approverCtx, task1.TaskId, "缺图"); err != nil {
		t.Fatalf("first reject: %v", err)
	}

	// 第二次：状态应为 Rejected，可以再提交
	var afterReject models.Spu
	if err := f.DB.First(&afterReject, spuId).Error; err != nil {
		t.Fatalf("read after reject: %v", err)
	}
	if afterReject.Status != models.SpuStatusRejected {
		t.Fatalf("expected Rejected before re-submit, got %d", afterReject.Status)
	}

	secondID, err := f.Spu.SubmitForReview(starterCtx, nil, &dto.SpuSubmitReq{SpuId: spuId, Remark: "v2"})
	if err != nil {
		t.Fatalf("re-submit: %v", err)
	}
	if secondID == firstID {
		t.Fatalf("expected new wf_instance on re-submit, got same id %d", secondID)
	}

	// SPU 推回 Reviewing
	var afterResubmit models.Spu
	if err := f.DB.First(&afterResubmit, spuId).Error; err != nil {
		t.Fatalf("read after resubmit: %v", err)
	}
	if afterResubmit.Status != models.SpuStatusReviewing {
		t.Fatalf("expected Reviewing after re-submit, got %d", afterResubmit.Status)
	}
	if afterResubmit.WorkflowInstanceId != int64(secondID) {
		t.Fatalf("expected WorkflowInstanceId=%d, got %d", secondID, afterResubmit.WorkflowInstanceId)
	}

	// binding 表对该 SPU 应只有 1 条（旧的被替换）
	var bindingCount int64
	if err := f.DB.Model(&platformModels.WorkflowBusinessBinding{}).
		Where("module_key = ? AND business_type = ? AND business_id = ?",
			SpuModuleKey, SpuBusinessType, int64ToString(spuId)).
		Count(&bindingCount).Error; err != nil {
		t.Fatalf("count binding: %v", err)
	}
	if bindingCount != 1 {
		t.Fatalf("expected 1 active binding for SPU %d, got %d", spuId, bindingCount)
	}
}

// TestE2E_Spu_DataScope_DeptOnly 验证 dataScope=3 仅看本部门 SPU。
//
// 拓扑：
//
//	dept 100 dev   user A (UserId=41, role data_scope=3)
//	dept 200 sales user B (UserId=42, role data_scope=3)
//
// A 和 B 各自创建一个 SPU。A 列表（dataScope=3, deptId=100）只看自己的 SPU。
func TestE2E_Spu_DataScope_DeptOnly(t *testing.T) {
	f := newE2ESpuWithDP(t, true)
	defer f.Cleanup()

	// 部门
	depts := []models.SysDept{
		{DeptId: 100, DeptName: "dev", DeptPath: "/100/", ParentId: 0},
		{DeptId: 200, DeptName: "sales", DeptPath: "/200/", ParentId: 0},
	}
	for i := range depts {
		if err := f.DB.Create(&depts[i]).Error; err != nil {
			t.Fatalf("seed dept: %v", err)
		}
	}
	// 用户 A 在 dev、用户 B 在 sales
	users := []models.SysUser{
		{UserId: 41, NickName: "a-dev", DeptId: 100, RoleId: 60, Status: "2"},
		{UserId: 42, NickName: "b-sales", DeptId: 200, RoleId: 61, Status: "2"},
	}
	for i := range users {
		if err := f.DB.Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	insA := &dto.SpuInsertReq{SpuCode: "DEV-1", SpuName: "dev product"}
	insA.SetCreateBy(41)
	if _, err := f.Spu.Insert(insA); err != nil {
		t.Fatalf("insert A: %v", err)
	}
	insB := &dto.SpuInsertReq{SpuCode: "SALES-1", SpuName: "sales product"}
	insB.SetCreateBy(42)
	if _, err := f.Spu.Insert(insB); err != nil {
		t.Fatalf("insert B: %v", err)
	}

	// 用户 A：dataScope=3（本部门）→ 只看 dept 100 用户（即自己）创建的 SPU
	dpA := &actions.DataPermission{DataScope: "3", UserId: 41, DeptId: 100, RoleId: 60}
	listA := make([]dto.SpuListItem, 0)
	var countA int64
	pageReq := &dto.SpuPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 50
	if err := f.Spu.GetPage(pageReq, dpA, &listA, &countA); err != nil {
		t.Fatalf("GetPage A: %v", err)
	}

	gotCodes := make([]string, 0, len(listA))
	for _, it := range listA {
		gotCodes = append(gotCodes, it.SpuCode)
	}
	sort.Strings(gotCodes)
	if len(gotCodes) != 1 || gotCodes[0] != "DEV-1" {
		t.Fatalf("dataScope=3 dept dev expected only DEV-1, got %v", gotCodes)
	}
	if countA != 1 {
		t.Fatalf("count expected 1, got %d", countA)
	}

	// 用户 B：dataScope=3 → 只看 dept 200 → SALES-1
	dpB := &actions.DataPermission{DataScope: "3", UserId: 42, DeptId: 200, RoleId: 61}
	listB := make([]dto.SpuListItem, 0)
	var countB int64
	pageReq2 := &dto.SpuPageReq{}
	pageReq2.PageIndex = 1
	pageReq2.PageSize = 50
	if err := f.Spu.GetPage(pageReq2, dpB, &listB, &countB); err != nil {
		t.Fatalf("GetPage B: %v", err)
	}
	if len(listB) != 1 || listB[0].SpuCode != "SALES-1" {
		t.Fatalf("dataScope=3 dept sales expected only SALES-1, got %+v", listB)
	}

	// admin（dataScope=1）应看到全部
	dpAdmin := &actions.DataPermission{DataScope: "1", UserId: 1, DeptId: 1, RoleId: 1}
	listAll := make([]dto.SpuListItem, 0)
	var countAll int64
	pageReq3 := &dto.SpuPageReq{}
	pageReq3.PageIndex = 1
	pageReq3.PageSize = 50
	if err := f.Spu.GetPage(pageReq3, dpAdmin, &listAll, &countAll); err != nil {
		t.Fatalf("GetPage admin: %v", err)
	}
	if countAll != 2 {
		t.Fatalf("admin expected 2 SPUs, got %d", countAll)
	}
}

// TestE2E_Spu_AuditMethod_Contract 是契约测试：apis/spu.go 与 platform/apis/workflow.go
// 真正落盘的审计 Method 字符串，是后续日志/告警/排查链路的稳定 key——
// 一旦改名就会让历史日志查询失效。这里钉死。
//
// 同步给文档：sku-module-guide.md 列了同一份契约。
func TestE2E_Spu_AuditMethod_Contract(t *testing.T) {
	cases := []struct {
		entry      audit.Entry
		wantMethod string
		wantBT     string // BusinessType（即 audit.Action）
	}{
		{
			entry: audit.Entry{
				Title:  "SPU 管理",
				Action: audit.ActionCreate,
				Target: audit.Target{Type: audit.CategorySpu, ID: int64(1), Label: "p"},
				Method: "admin.spu.insert",
			},
			wantMethod: "admin.spu.insert",
			wantBT:     audit.ActionCreate,
		},
		{
			entry: audit.Entry{
				Title:  "SPU 管理",
				Action: audit.ActionStart,
				Target: audit.Target{Type: audit.CategorySpu, ID: int64(1)},
				Method: "admin.spu.submit",
			},
			wantMethod: "admin.spu.submit",
			wantBT:     audit.ActionStart,
		},
		{
			entry: audit.Entry{
				Title:  "审批流",
				Action: audit.ActionApprove,
				Target: audit.Target{Type: audit.CategoryWorkflow, ID: 1},
				Method: "platform.workflow.task.approve",
			},
			wantMethod: "platform.workflow.task.approve",
			wantBT:     audit.ActionApprove,
		},
		{
			entry: audit.Entry{
				Title:  "审批流",
				Action: audit.ActionReject,
				Target: audit.Target{Type: audit.CategoryWorkflow, ID: 1},
				Method: "platform.workflow.task.reject",
			},
			wantMethod: "platform.workflow.task.reject",
			wantBT:     audit.ActionReject,
		},
	}
	for _, tc := range cases {
		meta := audit.BuildMeta(tc.entry)
		if meta.Method != tc.wantMethod {
			t.Errorf("Method mismatch: want %q got %q", tc.wantMethod, meta.Method)
		}
		if meta.BusinessType != tc.wantBT {
			t.Errorf("BusinessType mismatch for %q: want %q got %q", tc.wantMethod, tc.wantBT, meta.BusinessType)
		}
		if meta.BusinessTypes != tc.entry.Target.Type {
			t.Errorf("BusinessTypes mismatch for %q: want %q got %q", tc.wantMethod, tc.entry.Target.Type, meta.BusinessTypes)
		}
		if meta.Title == "" {
			t.Errorf("Title empty for %q", tc.wantMethod)
		}
	}
}

// TestE2E_Spu_AuditEmit_OnSubmit 验证：apis 层在 c (gin.Context) 上注入审计 meta 后，
// 中间件可读出本次 Method / Title。这模拟 SaveOperaLog 之前的链路接入点。
//
// 不依赖中间件实际写库——直接断言 c 上的 audit key 已被设置。
func TestE2E_Spu_AuditEmit_OnSubmit(t *testing.T) {
	c := makeRoleAuthCtx(11, "tester", []int{50})
	middleware.AuditLog(c, middleware.AuditEntry{
		Title:  "SPU 管理",
		Action: middleware.AuditActionStart,
		Target: middleware.AuditTarget{Type: middleware.AuditCategorySpu, ID: int64(99)},
		After: map[string]interface{}{
			"action":     "submit-for-review",
			"instanceId": 1234,
		},
		Method: "admin.spu.submit",
	})

	if v, ok := c.Get(audit.MethodKey); !ok || v.(string) != "admin.spu.submit" {
		t.Fatalf("expected audit method=admin.spu.submit on context, got %v ok=%v", v, ok)
	}
	if v, ok := c.Get(audit.TitleKey); !ok || v.(string) != "SPU 管理" {
		t.Fatalf("expected audit title set, got %v ok=%v", v, ok)
	}
	if v, ok := c.Get(audit.BusinessTypeKey); !ok || v.(string) != audit.ActionStart {
		t.Fatalf("expected audit BusinessType=start, got %v ok=%v", v, ok)
	}
	if v, ok := c.Get(audit.BusinessTypesKey); !ok || v.(string) != audit.CategorySpu {
		t.Fatalf("expected audit BusinessTypes=spu, got %v ok=%v", v, ok)
	}
}

// TestSpuCreateReview_FromFreshMigration 从空库跑全量 migration，再验证完整审批路径。
//
// 测试矩阵：
//   - 空 DB → AutoMigrate base schema（替代 1599190683659_tables，该步依赖磁盘 config/db.sql）
//   - 运行其余所有 migration，含 1779000000001_spu_workflow_seed 和
//     1779000000003_product_role_seed（EPO-54）
//   - 从 sys_role 捞 product_operator（data_scope=5）/ product_admin（data_scope=1）
//   - product_operator 创 SPU → 提交审核 → product_admin 通过 → SPU.status=3
//   - product_operator data_scope=5 列表过滤验证
//
// 前置：EPO-54 migration 1779000000003_product_role_seed.go 必须已 merge；
// 否则 product_operator 不在 sys_role，测试在步骤 "look up product_operator" 处 Fatalf（正常红灯）。
func TestSpuCreateReview_FromFreshMigration(t *testing.T) {
	// 独立命名的 in-memory DB，避免与其他 e2e 测试的 file::memory:?cache=shared 共享
	dsn := fmt.Sprintf("file:epo55_fresh_%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// AutoMigrate sys_migration + 替代 1599190683659_tables 建立的基础表
	// （1599190683659 读取磁盘 config/db.sql，无法在 test 环境运行）
	// CasbinRule 也需要提前建，供 1774900000000_drop_finance_subsystem 的 DELETE 语句使用。
	if err := db.AutoMigrate(
		&common.Migration{},
		&migModels.SysDept{},
		&migModels.SysConfig{},
		&migModels.SysTables{},
		&migModels.SysColumns{},
		&migModels.SysApi{},
		&migModels.SysMenu{},  // many2many 会同时建 sys_menu_api_rule
		&migModels.SysLoginLog{},
		&migModels.SysOperaLog{},
		&migModels.SysRoleDept{},
		&migModels.SysUser{},
		&migModels.SysRole{},  // many2many 会同时建 sys_role_menu
		&migModels.SysPost{},
		&migModels.DictData{},
		&migModels.DictType{},
		&migModels.SysJob{},
		&migModels.TbDemo{},
		&migModels.CasbinRule{}, // sys_casbin_rule，供 finance cleanup migration 删策略
	); err != nil {
		t.Fatalf("auto migrate base tables: %v", err)
	}

	// 预标记无法在 SQLite test 环境运行的 migration 为已完成：
	//   1599190683659 — 依赖磁盘 config/db.sql（InitDb 读文件）
	//   1778160000000 — 使用 "INSERT IGNORE" MySQL 方言（SQLite 不支持）
	skipVersions := []string{"1599190683659", "1778160000000"}
	for _, v := range skipVersions {
		if err := db.Create(&common.Migration{Version: v}).Error; err != nil {
			t.Fatalf("pre-seed %s: %v", v, err)
		}
	}

	// 运行其余所有已注册的 migration（_ "go-admin/cmd/migrate/migration/version" 在 init() 里注册）
	// 包括：1779000000001_spu_workflow_seed + 1779000000003_product_role_seed（EPO-54）
	migration.Migrate.SetDb(db)
	migration.Migrate.Migrate()

	// 从 sys_role 捞出 product_operator — migration 1779000000003 应已种好
	var opRole models.SysRole
	if err := db.Table("sys_role").Where("role_key = ?", "product_operator").First(&opRole).Error; err != nil {
		t.Fatalf("product_operator 角色缺失（EPO-54 未 merge？）: %v", err)
	}
	if opRole.DataScope != "5" {
		t.Fatalf("product_operator.data_scope=%q, want \"5\"", opRole.DataScope)
	}

	var adminRole models.SysRole
	if err := db.Table("sys_role").Where("role_key = ?", "product_admin").First(&adminRole).Error; err != nil {
		t.Fatalf("product_admin 角色缺失: %v", err)
	}
	if adminRole.DataScope != "1" {
		t.Fatalf("product_admin.data_scope=%q, want \"1\"", adminRole.DataScope)
	}

	// wf_definition + approve_1 节点应由 migration 建好
	var wfDef platformModels.WorkflowDefinition
	if err := db.Where("definition_key = ?", SpuDefaultDefinitionKey).First(&wfDef).Error; err != nil {
		t.Fatalf("wf_definition spu_create_review 缺失: %v", err)
	}
	var wfNode platformModels.WorkflowDefinitionNode
	if err := db.Where("definition_id = ? AND node_key = ?", wfDef.DefinitionId, "approve_1").First(&wfNode).Error; err != nil {
		t.Fatalf("wf_definition_node approve_1 缺失（product_admin 需先存在）: %v", err)
	}
	if wfNode.ApproverValue != strconv.Itoa(adminRole.RoleId) {
		t.Fatalf("approve_1.approver_value=%q, want %q (product_admin.role_id)",
			wfNode.ApproverValue, strconv.Itoa(adminRole.RoleId))
	}

	// 开启 dataScope 以验证 product_operator data_scope=5 的列表过滤
	prevEnable := config.ApplicationConfig.EnableDP
	config.ApplicationConfig.EnableDP = true
	t.Cleanup(func() { config.ApplicationConfig.EnableDP = prevEnable })

	s := &Spu{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	wf := &platformService.Workflow{Service: s.Service}

	// 测试用户：product_operator（UserId=201）
	operator := models.SysUser{UserId: 201, NickName: "产品操作员", Status: "2", RoleId: opRole.RoleId, DeptId: 10}
	if err := db.Create(&operator).Error; err != nil {
		t.Fatalf("seed operator: %v", err)
	}
	// 测试用户：product_admin（UserId=202）
	approver := models.SysUser{UserId: 202, NickName: "产品管理员", Status: "2", RoleId: adminRole.RoleId, DeptId: 10}
	if err := db.Create(&approver).Error; err != nil {
		t.Fatalf("seed approver: %v", err)
	}

	// product_operator 创建 SPU
	insReq := &dto.SpuInsertReq{SpuCode: "FM-E2E-001", SpuName: "fresh migration e2e product"}
	insReq.SetCreateBy(operator.UserId)
	spuId, err := s.Insert(insReq)
	if err != nil {
		t.Fatalf("Insert SPU: %v", err)
	}

	// 提交审核（dataScope 已开启，需传 DataPermission；nil 会在 Permission() 里触发 panic）
	operatorCtx := makeRoleAuthCtx(operator.UserId, operator.NickName, []int{opRole.RoleId})
	dpOp := &actions.DataPermission{DataScope: "5", UserId: operator.UserId, DeptId: operator.DeptId, RoleId: opRole.RoleId}
	instanceID, err := s.SubmitForReview(operatorCtx, dpOp, &dto.SpuSubmitReq{SpuId: spuId, Remark: "fresh install e2e"})
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}

	// product_admin 取 pending task → approve → SPU.status=3
	task := findPendingTask(t, db, instanceID)
	approverCtx := makeRoleAuthCtx(approver.UserId, approver.NickName, []int{adminRole.RoleId})
	if _, err := wf.Approve(approverCtx, task.TaskId, "LGTM"); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	var postSpu models.Spu
	if err := db.First(&postSpu, spuId).Error; err != nil {
		t.Fatalf("read SPU after approve: %v", err)
	}
	if postSpu.Status != models.SpuStatusApproved {
		t.Fatalf("SPU.status=%d, want %d (Approved/3)", postSpu.Status, models.SpuStatusApproved)
	}
	if postSpu.ApprovedAt == nil {
		t.Fatalf("SPU.approved_at should be set after approve")
	}

	// 验证 product_operator data_scope=5：只能看自己创建的 SPU
	// 再建一个由其他 operator 创建的 SPU
	otherOp := models.SysUser{UserId: 203, NickName: "另一操作员", Status: "2", RoleId: opRole.RoleId, DeptId: 20}
	if err := db.Create(&otherOp).Error; err != nil {
		t.Fatalf("seed other operator: %v", err)
	}
	insOther := &dto.SpuInsertReq{SpuCode: "FM-E2E-002", SpuName: "other operator product"}
	insOther.SetCreateBy(otherOp.UserId)
	if _, err := s.Insert(insOther); err != nil {
		t.Fatalf("insert other SPU: %v", err)
	}

	dpOpList := &actions.DataPermission{DataScope: "5", UserId: operator.UserId, DeptId: operator.DeptId, RoleId: opRole.RoleId}
	listOp := make([]dto.SpuListItem, 0)
	var countOp int64
	pageReq := &dto.SpuPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 50
	if err := s.GetPage(pageReq, dpOpList, &listOp, &countOp); err != nil {
		t.Fatalf("GetPage for operator: %v", err)
	}
	if countOp != 1 || listOp[0].SpuCode != "FM-E2E-001" {
		t.Fatalf("product_operator(data_scope=5) expected only FM-E2E-001, got count=%d codes=%v",
			countOp, func() []string {
				var codes []string
				for _, it := range listOp {
					codes = append(codes, it.SpuCode)
				}
				return codes
			}())
	}

	// product_admin data_scope=1 应看到全部 SPU
	dpAdmin := &actions.DataPermission{DataScope: "1", UserId: approver.UserId, DeptId: approver.DeptId, RoleId: adminRole.RoleId}
	listAdmin := make([]dto.SpuListItem, 0)
	var countAdmin int64
	pageReq2 := &dto.SpuPageReq{}
	pageReq2.PageIndex = 1
	pageReq2.PageSize = 50
	if err := s.GetPage(pageReq2, dpAdmin, &listAdmin, &countAdmin); err != nil {
		t.Fatalf("GetPage for admin: %v", err)
	}
	if countAdmin != 2 {
		t.Fatalf("product_admin(data_scope=1) expected 2 SPUs, got %d", countAdmin)
	}
}
