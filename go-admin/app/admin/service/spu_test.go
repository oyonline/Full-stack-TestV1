package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/logger"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	platformModels "go-admin/app/platform/models"
	platformDto "go-admin/app/platform/service/dto"
)

func newTestSpu(t *testing.T) *Spu {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Spu{},
		&models.Sku{},
		&models.SysUser{},
		&models.SysRole{},
		&platformModels.ModuleRegistry{},
		&platformModels.WorkflowDefinition{},
		&platformModels.WorkflowDefinitionNode{},
		&platformModels.WorkflowInstance{},
		&platformModels.WorkflowTask{},
		&platformModels.WorkflowActionLog{},
		&platformModels.WorkflowBusinessBinding{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	for _, tbl := range []string{"spu", "sku", "sys_user", "sys_role", "module_registry",
		"wf_definition", "wf_definition_node", "wf_instance", "wf_task",
		"wf_action_log", "wf_business_binding"} {
		if err := db.Exec("DELETE FROM " + tbl).Error; err != nil {
			t.Fatalf("clean %s: %v", tbl, err)
		}
	}
	// 显式注入 admin 模块到 module_registry，确保 EnsureModuleEnabled 严格校验通过
	if err := db.Create(&platformModels.ModuleRegistry{
		ModuleKey:      "admin",
		ModuleName:     "后台管理",
		RouteBase:      "/admin",
		MenuRootCode:   "admin",
		Status:         "2",
		Sort:           1,
		PermissionHint: "admin",
		Remark:         "test seed",
	}).Error; err != nil {
		t.Fatalf("seed admin module_registry: %v", err)
	}

	s := &Spu{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s
}

// makeAuthCtx 构造一个带 jwt claims 的 gin.Context，便于走到 user.GetUserId / GetUserName。
func makeAuthCtx(userId int, userName string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/test", nil)
	c.Request = req
	claims := jwt.MapClaims{
		"identity": float64(userId),
		"nice":     userName,
	}
	c.Set(jwt.JwtPayloadKey, claims)
	return c
}

func TestSpu_Insert_RejectsEmptyCode(t *testing.T) {
	s := newTestSpu(t)
	req := &dto.SpuInsertReq{SpuCode: "  ", SpuName: "x"}
	req.SetCreateBy(1)
	if _, err := s.Insert(req); err == nil {
		t.Fatalf("expected error for blank code")
	}
}

func TestSpu_Insert_DefaultsToDraft(t *testing.T) {
	s := newTestSpu(t)
	req := &dto.SpuInsertReq{SpuCode: "C1", SpuName: "p1"}
	req.SetCreateBy(7)
	id, err := s.Insert(req)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	var spu models.Spu
	if err := s.Orm.First(&spu, id).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	if spu.Status != models.SpuStatusDraft {
		t.Fatalf("expected Status=Draft(1), got %d", spu.Status)
	}
	if spu.CreatorId != 7 {
		t.Fatalf("expected CreatorId=7, got %d", spu.CreatorId)
	}
}

func TestSpu_Update_RejectsReviewing(t *testing.T) {
	s := newTestSpu(t)
	spu := models.Spu{SpuCode: "C2", SpuName: "p2", Status: models.SpuStatusReviewing}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	updReq := &dto.SpuUpdateReq{
		SpuId:   spu.SpuId,
		SpuCode: "C2",
		SpuName: "p2-edited",
	}
	updReq.SetUpdateBy(1)
	if err := s.Update(updReq, nil); err == nil {
		t.Fatalf("expected reviewing block")
	}
}

// TestSpu_SubmitForReview_HappyPath：插入 SPU，提交审核，验证：
//   - SPU.status 推进到 Reviewing
//   - workflow_instance_id 写回
//   - submitted_at 已设
//   - wf_instance + wf_business_binding 已创建
func TestSpu_SubmitForReview_HappyPath(t *testing.T) {
	s := newTestSpu(t)

	// 角色 + 流程定义 + 节点 seed
	role := models.SysRole{RoleId: 99, RoleName: "产品管理员", RoleKey: "product_admin", Status: "2"}
	if err := s.Orm.Create(&role).Error; err != nil {
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
	if err := s.Orm.Create(&def).Error; err != nil {
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
	if err := s.Orm.Create(&node).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}

	// 用户作为提交者
	starter := models.SysUser{UserId: 11, NickName: "tester", Status: "2"}
	if err := s.Orm.Create(&starter).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// 插入 SPU
	insReq := &dto.SpuInsertReq{SpuCode: "WF-001", SpuName: "wf product"}
	insReq.SetCreateBy(11)
	spuId, err := s.Insert(insReq)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// SubmitForReview
	c := makeAuthCtx(11, "tester")
	subReq := &dto.SpuSubmitReq{SpuId: spuId, Remark: "go"}
	instanceId, err := s.SubmitForReview(c, nil, subReq)
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}
	if instanceId <= 0 {
		t.Fatalf("expected instanceId>0, got %d", instanceId)
	}

	var post models.Spu
	if err := s.Orm.First(&post, spuId).Error; err != nil {
		t.Fatalf("read spu: %v", err)
	}
	if post.Status != models.SpuStatusReviewing {
		t.Fatalf("expected Status=Reviewing(2), got %d", post.Status)
	}
	if post.WorkflowInstanceId != int64(instanceId) {
		t.Fatalf("expected WorkflowInstanceId=%d, got %d", instanceId, post.WorkflowInstanceId)
	}
	if post.SubmittedAt == nil {
		t.Fatalf("expected SubmittedAt set")
	}

	var binding platformModels.WorkflowBusinessBinding
	if err := s.Orm.Where("instance_id = ?", instanceId).First(&binding).Error; err != nil {
		t.Fatalf("binding: %v", err)
	}
	if binding.BusinessType != SpuBusinessType {
		t.Fatalf("binding.BusinessType=%q, want %q", binding.BusinessType, SpuBusinessType)
	}
	if binding.BusinessId != int64ToString(spuId) {
		t.Fatalf("binding.BusinessId=%q, want %q", binding.BusinessId, int64ToString(spuId))
	}
}

func TestSpu_SubmitForReview_RejectsNonDraftStatus(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{
		SpuCode: "WF-2",
		SpuName: "wf2",
		Status:  models.SpuStatusReviewing, // 已经在审核中
	}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	c := makeAuthCtx(1, "x")
	subReq := &dto.SpuSubmitReq{SpuId: spu.SpuId}
	if _, err := s.SubmitForReview(c, nil, subReq); err == nil {
		t.Fatalf("expected error for non-draft submit")
	}
}

func TestSpu_SubmitForReview_NoDefinition(t *testing.T) {
	s := newTestSpu(t)

	insReq := &dto.SpuInsertReq{SpuCode: "WF-3", SpuName: "wf3"}
	insReq.SetCreateBy(1)
	spuId, err := s.Insert(insReq)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	c := makeAuthCtx(1, "x")
	subReq := &dto.SpuSubmitReq{SpuId: spuId}
	if _, err := s.SubmitForReview(c, nil, subReq); err == nil {
		t.Fatalf("expected error when no definition seeded")
	}
}
