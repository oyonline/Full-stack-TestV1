package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_SpuWorkflowSeed 验证：
//   - 创建 wf_definition (definition_key=spu_create_review, status=2)
//   - 当 sys_role 中存在 product_admin 时同时创建一个审批节点
//   - 当 sys_role 中没有 product_admin 时只创建定义本身（不卡迁移）
//   - 重跑幂等：行数不变
func TestMigration_SpuWorkflowSeed_WithRole(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(
		&models.WorkflowDefinition{},
		&models.WorkflowDefinitionNode{},
		&common.Migration{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	type sysRole struct {
		RoleId  int    `gorm:"primaryKey"`
		RoleKey string `gorm:"size:64"`
	}
	if err := db.Table("sys_role").AutoMigrate(&sysRole{}); err != nil {
		t.Fatalf("create sys_role: %v", err)
	}
	if err := db.Table("sys_role").Create(&sysRole{RoleId: 42, RoleKey: "product_admin"}).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	const ver = "1779000000001_spu_workflow_seed.go"
	if err := _1779000000001SpuWorkflowSeed(db, ver); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var def models.WorkflowDefinition
	if err := db.Where("definition_key = ?", "spu_create_review").First(&def).Error; err != nil {
		t.Fatalf("def not created: %v", err)
	}
	if def.Status != "2" {
		t.Fatalf("def.Status=%q, want 2", def.Status)
	}

	var nodeCount int64
	if err := db.Model(&models.WorkflowDefinitionNode{}).
		Where("definition_id = ?", def.DefinitionId).
		Count(&nodeCount).Error; err != nil {
		t.Fatalf("count nodes: %v", err)
	}
	if nodeCount != 1 {
		t.Fatalf("expected 1 node, got %d", nodeCount)
	}

	// 重跑：再清掉 migration 记录后跑一次，应该幂等
	if err := db.Where("version = ?", ver).Delete(&common.Migration{}).Error; err != nil {
		t.Fatalf("delete migration: %v", err)
	}
	if err := _1779000000001SpuWorkflowSeed(db, ver); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	var defCount int64
	db.Model(&models.WorkflowDefinition{}).Where("definition_key = ?", "spu_create_review").Count(&defCount)
	if defCount != 1 {
		t.Fatalf("expected 1 def after rerun, got %d", defCount)
	}
}

func TestMigration_SpuWorkflowSeed_WithoutRole(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(
		&models.WorkflowDefinition{},
		&models.WorkflowDefinitionNode{},
		&common.Migration{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	type sysRole struct {
		RoleId  int    `gorm:"primaryKey"`
		RoleKey string `gorm:"size:64"`
	}
	if err := db.Table("sys_role").AutoMigrate(&sysRole{}); err != nil {
		t.Fatalf("create sys_role: %v", err)
	}
	// 不 seed product_admin

	const ver = "1779000000001_spu_workflow_seed_norole.go"
	if err := _1779000000001SpuWorkflowSeed(db, ver); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var defCount, nodeCount int64
	db.Model(&models.WorkflowDefinition{}).Count(&defCount)
	db.Model(&models.WorkflowDefinitionNode{}).Count(&nodeCount)
	if defCount != 1 {
		t.Fatalf("expected 1 def, got %d", defCount)
	}
	if nodeCount != 0 {
		t.Fatalf("expected 0 nodes when role missing, got %d", nodeCount)
	}
}
