package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
)

func newModuleGateDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&platformModels.ModuleRegistry{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("DELETE FROM module_registry").Error; err != nil {
		t.Fatalf("clean table: %v", err)
	}
	return db
}

func TestEnsureModuleEnabled_EmptyTablePassesThrough(t *testing.T) {
	db := newModuleGateDB(t)
	if err := EnsureModuleEnabled(db, "anything-goes"); err != nil {
		t.Fatalf("expected nil (degraded passthrough on empty table), got %v", err)
	}
}

func TestEnsureModuleEnabled_HitEnabledRow(t *testing.T) {
	db := newModuleGateDB(t)
	row := platformModels.ModuleRegistry{
		ModuleKey:    "foo",
		ModuleName:   "Foo",
		RouteBase:    "/foo",
		MenuRootCode: "foo-root",
		Status:       "2",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed row: %v", err)
	}
	if err := EnsureModuleEnabled(db, "foo"); err != nil {
		t.Fatalf("expected nil for enabled module, got %v", err)
	}
}

func TestEnsureModuleEnabled_MissOnNonEmptyEnabledTableRejects(t *testing.T) {
	db := newModuleGateDB(t)
	row := platformModels.ModuleRegistry{
		ModuleKey:    "foo",
		ModuleName:   "Foo",
		RouteBase:    "/foo",
		MenuRootCode: "foo-root",
		Status:       "2",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed row: %v", err)
	}
	err := EnsureModuleEnabled(db, "bar")
	if err == nil {
		t.Fatalf("expected rejection for unknown moduleKey, got nil")
	}
	if err.Error() != "模块未注册或未启用" {
		t.Fatalf("expected zh error, got %q", err.Error())
	}
}

// 路径 4：表里只有 status=1 的禁用记录 → status=2 的计数为 0 → 视同"空表"语义放行。
func TestEnsureModuleEnabled_OnlyDisabledRowDegrades(t *testing.T) {
	db := newModuleGateDB(t)
	row := platformModels.ModuleRegistry{
		ModuleKey:    "foo",
		ModuleName:   "Foo",
		RouteBase:    "/foo",
		MenuRootCode: "foo-root",
		Status:       "1",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed disabled row: %v", err)
	}
	if err := EnsureModuleEnabled(db, "foo"); err != nil {
		t.Fatalf("expected degraded passthrough (no enabled rows in table), got %v", err)
	}
	if err := EnsureModuleEnabled(db, "anything-else"); err != nil {
		t.Fatalf("expected degraded passthrough for any key when no enabled rows exist, got %v", err)
	}
}
