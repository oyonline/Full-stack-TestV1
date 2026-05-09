package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_SpuIsOnline_DataMigration 验证已有 status=3 的 SPU 在迁移后 is_online=true，
// 其他 status 的 SPU 保持 is_online=false。
func TestMigration_SpuIsOnline_DataMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:spuisonline?mode=memory&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Spu{}, &common.Migration{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	// 清空以便重入
	db.Exec("DELETE FROM spu")
	db.Exec("DELETE FROM sys_migration")

	// Seed: status=3(Approved) SPU + status=1(Draft) SPU
	approved := models.Spu{SpuCode: "A1", SpuName: "Approved SPU", Status: 3}
	draft := models.Spu{SpuCode: "D1", SpuName: "Draft SPU", Status: 1}
	if err := db.Create(&approved).Error; err != nil {
		t.Fatalf("seed approved: %v", err)
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("seed draft: %v", err)
	}

	const ver = "1779000000004_spu_is_online.go"
	if err := _1779000000004SpuIsOnline(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	var gotApproved models.Spu
	if err := db.Where("spu_code = ?", "A1").First(&gotApproved).Error; err != nil {
		t.Fatalf("read approved spu: %v", err)
	}
	if !gotApproved.IsOnline {
		t.Fatalf("approved SPU should have is_online=true after migration")
	}

	var gotDraft models.Spu
	if err := db.Where("spu_code = ?", "D1").First(&gotDraft).Error; err != nil {
		t.Fatalf("read draft spu: %v", err)
	}
	if gotDraft.IsOnline {
		t.Fatalf("draft SPU should have is_online=false after migration")
	}
}

// TestMigration_SpuIsOnline_Idempotent 验证重跑迁移不产生重复数据。
func TestMigration_SpuIsOnline_Idempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:spuisonlineidp?mode=memory&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Spu{},
		&models.SysMenu{},
		&models.SysApi{},
		&common.Migration{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS sys_menu_api_rule (sys_menu_menu_id INTEGER NOT NULL, sys_api_id INTEGER NOT NULL)`)
	db.Exec("DELETE FROM spu")
	db.Exec("DELETE FROM sys_menu")
	db.Exec("DELETE FROM sys_api")
	db.Exec("DELETE FROM sys_menu_api_rule")
	db.Exec("DELETE FROM sys_migration")

	// Seed SpuManage menu so buttons can be created
	spuMenu := models.SysMenu{
		MenuName: "SpuManage", Title: "产品 SPU 管理",
		Path: "/product/spu", MenuType: "C", Permission: "admin:spu:list",
		ParentId: 1, Component: "admin/sys-spu/index",
	}
	if err := db.Create(&spuMenu).Error; err != nil {
		t.Fatalf("seed spuMenu: %v", err)
	}

	const ver = "1779000000004_spu_is_online.go"

	// First run
	if err := _1779000000004SpuIsOnline(db, ver); err != nil {
		t.Fatalf("migration first run: %v", err)
	}

	// Clear sys_migration and re-run (simulate idempotency)
	db.Exec("DELETE FROM sys_migration WHERE version = ?", ver)
	if err := _1779000000004SpuIsOnline(db, ver); err != nil {
		t.Fatalf("migration rerun: %v", err)
	}

	var btnCount int64
	if err := db.Table("sys_menu").
		Where("permission IN ?", []string{"admin:spu:online", "admin:spu:offline"}).
		Count(&btnCount).Error; err != nil {
		t.Fatalf("count buttons: %v", err)
	}
	if btnCount != 2 {
		t.Fatalf("expected 2 button rows after rerun, got %d", btnCount)
	}

	var apiCount int64
	if err := db.Table("sys_api").
		Where("path IN ?", []string{"/api/v1/spu/:id/online", "/api/v1/spu/:id/offline"}).
		Count(&apiCount).Error; err != nil {
		t.Fatalf("count apis: %v", err)
	}
	if apiCount != 2 {
		t.Fatalf("expected 2 api rows after rerun, got %d", apiCount)
	}
}
