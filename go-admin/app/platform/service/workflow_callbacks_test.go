package service

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
	"go-admin/app/platform/service/dto"
)

func newCallbackDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&platformModels.WorkflowBusinessBinding{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("DELETE FROM wf_business_binding").Error; err != nil {
		t.Fatalf("clean table: %v", err)
	}
	return db
}

func seedBinding(t *testing.T, db *gorm.DB, instanceID int, businessType string) {
	t.Helper()
	row := platformModels.WorkflowBusinessBinding{
		ModuleKey:      "spu",
		BusinessType:   businessType,
		BusinessId:     "1001",
		BusinessNo:     "SPU-1001",
		Title:          "test binding",
		InstanceId:     instanceID,
		WorkflowStatus: dto.WorkflowStatusReview,
		BusinessStatus: dto.WorkflowStatusReview,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed binding: %v", err)
	}
}

func TestRegisterAndDispatch_Approved(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 100, "spu")

	called := 0
	var seenStatus string
	var seenBusinessType string
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		called++
		seenStatus = terminalStatus
		seenBusinessType = binding.BusinessType
		return nil
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 100}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusApproved)
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected handler called 1x, got %d", called)
	}
	if seenStatus != dto.WorkflowStatusApproved {
		t.Fatalf("expected status %q, got %q", dto.WorkflowStatusApproved, seenStatus)
	}
	if seenBusinessType != "spu" {
		t.Fatalf("expected businessType %q, got %q", "spu", seenBusinessType)
	}
}

func TestRegisterAndDispatch_Rejected(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 200, "spu")

	called := 0
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		if terminalStatus != dto.WorkflowStatusRejected {
			t.Errorf("expected rejected, got %q", terminalStatus)
		}
		called++
		return nil
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 200}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusRejected)
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected handler called 1x, got %d", called)
	}
}

func TestRegisterAndDispatch_Canceled(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 300, "spu")

	called := 0
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		if terminalStatus != dto.WorkflowStatusCanceled {
			t.Errorf("expected cancelled, got %q", terminalStatus)
		}
		called++
		return nil
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 300}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusCanceled)
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected handler called 1x, got %d", called)
	}
}

func TestNoHandlerRegistered_NoOp(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 400, "unregistered-type")

	instance := &platformModels.WorkflowInstance{InstanceId: 400}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusApproved)
	})
	if err != nil {
		t.Fatalf("expected no error when no handler registered, got %v", err)
	}
}

func TestNoBindingFound_NoOp(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	// 不 seed binding：模拟旁路 workflow（没绑业务）。

	called := 0
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		called++
		return nil
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 999}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusApproved)
	})
	if err != nil {
		t.Fatalf("expected no error when binding missing, got %v", err)
	}
	if called != 0 {
		t.Fatalf("expected handler NOT called when binding missing, got %d", called)
	}
}

func TestHandlerError_RollsBackTransaction(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 500, "spu")

	handlerErr := errors.New("business handler failed")
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		// 在 handler 内部对 binding 做改动，验证 handler error 会回滚这些改动。
		if err := tx.Model(&platformModels.WorkflowBusinessBinding{}).
			Where("instance_id = ?", binding.InstanceId).
			Update("business_status", "handler-mutated").Error; err != nil {
			return err
		}
		return handlerErr
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 500}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusApproved)
	})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error to bubble up, got %v", err)
	}

	// 验证事务回滚：binding 的 business_status 不应被改成 "handler-mutated"。
	var post platformModels.WorkflowBusinessBinding
	if err := db.Where("instance_id = ?", 500).First(&post).Error; err != nil {
		t.Fatalf("read back binding: %v", err)
	}
	if post.BusinessStatus == "handler-mutated" {
		t.Fatalf("expected transaction rollback to discard handler mutation, but business_status=%q", post.BusinessStatus)
	}
	if post.BusinessStatus != dto.WorkflowStatusReview {
		t.Fatalf("expected business_status to remain seeded value %q, got %q", dto.WorkflowStatusReview, post.BusinessStatus)
	}
}

func TestNonTerminalStatus_NoOp(t *testing.T) {
	resetTerminalHandlersForTest()
	t.Cleanup(resetTerminalHandlersForTest)

	db := newCallbackDB(t)
	seedBinding(t, db, 600, "spu")

	called := 0
	RegisterTerminalHandler("spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
		called++
		return nil
	})

	instance := &platformModels.WorkflowInstance{InstanceId: 600}
	err := db.Transaction(func(tx *gorm.DB) error {
		return dispatchTerminalHandler(tx, instance, dto.WorkflowStatusReview)
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if called != 0 {
		t.Fatalf("expected handler NOT called for non-terminal status, got %d", called)
	}
}

func TestIsTerminalWorkflowStatus(t *testing.T) {
	cases := map[string]bool{
		dto.WorkflowStatusApproved: true,
		dto.WorkflowStatusRejected: true,
		dto.WorkflowStatusCanceled: true,
		dto.WorkflowStatusReview:   false,
		dto.WorkflowStatusDraft:    false,
		"":                         false,
		"unknown":                  false,
	}
	for status, want := range cases {
		if got := isTerminalWorkflowStatus(status); got != want {
			t.Errorf("isTerminalWorkflowStatus(%q)=%v, want %v", status, got, want)
		}
	}
}
