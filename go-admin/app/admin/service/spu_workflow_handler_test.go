package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	platformModels "go-admin/app/platform/models"
	platformDto "go-admin/app/platform/service/dto"
)

// newSpuHandlerDB 提供一个空的 sqlite + 自动迁移 SPU 表，模拟 platform workflow 的事务上下文。
func newSpuHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Spu{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("DELETE FROM spu").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}
	return db
}

func seedSpu(t *testing.T, db *gorm.DB, status int) int64 {
	t.Helper()
	spu := models.Spu{
		SpuCode: "TEST-CODE",
		SpuName: "test spu",
		Status:  status,
	}
	if err := db.Create(&spu).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return spu.SpuId
}

func TestSpuTerminalHandler_Approved(t *testing.T) {
	db := newSpuHandlerDB(t)
	spuId := seedSpu(t, db, models.SpuStatusReviewing)

	binding := &platformModels.WorkflowBusinessBinding{
		ModuleKey:    SpuModuleKey,
		BusinessType: SpuBusinessType,
		BusinessId:   strconvI64(spuId),
	}
	if err := onSpuWorkflowTerminal(db, binding, platformDto.WorkflowStatusApproved); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var post models.Spu
	if err := db.First(&post, spuId).Error; err != nil {
		t.Fatalf("readback: %v", err)
	}
	if post.Status != models.SpuStatusApproved {
		t.Fatalf("expected status=Approved(3), got %d", post.Status)
	}
	if post.ApprovedAt == nil || post.ApprovedAt.IsZero() {
		t.Fatalf("expected approved_at to be set")
	}
	if time.Since(*post.ApprovedAt) > time.Minute {
		t.Fatalf("approved_at should be recent, got %v", post.ApprovedAt)
	}
}

func TestSpuTerminalHandler_Rejected(t *testing.T) {
	db := newSpuHandlerDB(t)
	spuId := seedSpu(t, db, models.SpuStatusReviewing)

	binding := &platformModels.WorkflowBusinessBinding{BusinessId: strconvI64(spuId)}
	if err := onSpuWorkflowTerminal(db, binding, platformDto.WorkflowStatusRejected); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var post models.Spu
	if err := db.First(&post, spuId).Error; err != nil {
		t.Fatalf("readback: %v", err)
	}
	if post.Status != models.SpuStatusRejected {
		t.Fatalf("expected status=Rejected(4), got %d", post.Status)
	}
	if post.ApprovedAt != nil {
		t.Fatalf("approved_at should remain nil on reject, got %v", post.ApprovedAt)
	}
}

func TestSpuTerminalHandler_Canceled(t *testing.T) {
	db := newSpuHandlerDB(t)
	spuId := seedSpu(t, db, models.SpuStatusReviewing)

	binding := &platformModels.WorkflowBusinessBinding{BusinessId: strconvI64(spuId)}
	if err := onSpuWorkflowTerminal(db, binding, platformDto.WorkflowStatusCanceled); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var post models.Spu
	if err := db.First(&post, spuId).Error; err != nil {
		t.Fatalf("readback: %v", err)
	}
	if post.Status != models.SpuStatusDraft {
		t.Fatalf("expected status=Draft(1) after cancel, got %d", post.Status)
	}
}

func TestSpuTerminalHandler_UnknownStatus_NoOp(t *testing.T) {
	db := newSpuHandlerDB(t)
	spuId := seedSpu(t, db, models.SpuStatusReviewing)

	binding := &platformModels.WorkflowBusinessBinding{BusinessId: strconvI64(spuId)}
	if err := onSpuWorkflowTerminal(db, binding, "weird"); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var post models.Spu
	if err := db.First(&post, spuId).Error; err != nil {
		t.Fatalf("readback: %v", err)
	}
	if post.Status != models.SpuStatusReviewing {
		t.Fatalf("status should not change on unknown terminal, got %d", post.Status)
	}
}

func TestSpuTerminalHandler_InvalidBusinessId_NoOp(t *testing.T) {
	db := newSpuHandlerDB(t)
	_ = seedSpu(t, db, models.SpuStatusReviewing)

	binding := &platformModels.WorkflowBusinessBinding{BusinessId: "not-a-number"}
	if err := onSpuWorkflowTerminal(db, binding, platformDto.WorkflowStatusApproved); err != nil {
		t.Fatalf("handler should swallow invalid id, got %v", err)
	}
}

func TestSpuTerminalHandler_NilBinding(t *testing.T) {
	db := newSpuHandlerDB(t)
	if err := onSpuWorkflowTerminal(db, nil, platformDto.WorkflowStatusApproved); err != nil {
		t.Fatalf("nil binding should be no-op, got %v", err)
	}
}

// strconvI64 复用 spu.go 的 int64ToString，避免引入第三方 helper。
func strconvI64(v int64) string { return int64ToString(v) }
