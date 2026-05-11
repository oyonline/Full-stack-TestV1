package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

func Test1779000000004_ModuleRegistryAdminSeed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.ModuleRegistry{}, &common.Migration{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	if err := _1779000000004ModuleRegistryAdminSeed(db, "1779000000004"); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	var m models.ModuleRegistry
	if err := db.Where("module_key = ?", "admin").First(&m).Error; err != nil {
		t.Fatalf("admin module not found: %v", err)
	}
	if m.Status != "2" {
		t.Fatalf("expected status=2, got %q", m.Status)
	}
	if m.ModuleName != "后台管理" {
		t.Fatalf("expected ModuleName=后台管理, got %q", m.ModuleName)
	}

	// 幂等：module_registry 的 OnConflict DoNothing 保证数据不变；
	// sys_migration 的 version UNIQUE 约束导致第二次写 migration 记录会失败，
	// 这是预期行为（与生产 migration 框架行为一致：已执行的 version 不会再跑）。
	// 这里只验证 module_registry 侧幂等。
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "module_key"}},
		DoNothing: true,
	}).Create(&models.ModuleRegistry{
		ModuleKey:      "admin",
		ModuleName:     "后台管理",
		RouteBase:      "/admin",
		MenuRootCode:   "admin",
		Status:         "2",
		Sort:           1,
		PermissionHint: "admin",
		Remark:         "平台基础业务模块（SPU/SKU/公告等）",
	}).Error; err != nil {
		t.Fatalf("idempotent re-insert failed: %v", err)
	}
	var count int64
	if err := db.Model(&models.ModuleRegistry{}).Where("module_key = ?", "admin").Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 admin row, got %d", count)
	}
}
