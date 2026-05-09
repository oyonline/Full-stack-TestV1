package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// sysRoleMenuRow is a minimal join-table struct for AutoMigrate in tests.
type sysRoleMenuRow struct {
	RoleId int `gorm:"primaryKey;column:role_id"`
	MenuId int `gorm:"primaryKey;column:menu_id"`
}

func (sysRoleMenuRow) TableName() string { return "sys_role_menu" }

// sysMenuApiRuleRow is a minimal join-table struct for AutoMigrate in tests.
type sysMenuApiRuleRow struct {
	SysMenuMenuId int `gorm:"primaryKey;column:sys_menu_menu_id"`
	SysApiId      int `gorm:"primaryKey;column:sys_api_id"`
}

func (sysMenuApiRuleRow) TableName() string { return "sys_menu_api_rule" }

// setupProductRoleSeedDB creates an in-memory SQLite DB with all tables needed
// by _1779000000003ProductRoleSeed and seeds representative menus + APIs +
// sys_menu_api_rule bridges (simulating a post-1779000000000 environment).
// Returns (db, spuMenuId, spuAddBtnId) for callers that need to verify bindings.
func setupProductRoleSeedDB(t *testing.T) (db *gorm.DB, spuMenuId int, spuAddBtnId int) {
	t.Helper()
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.SysMenu{},
		&models.SysApi{},
		&models.SysRole{},
		&models.CasbinRule{},
		&models.WorkflowDefinition{},
		&models.WorkflowDefinitionNode{},
		&common.Migration{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.AutoMigrate(&sysRoleMenuRow{}); err != nil {
		t.Fatalf("auto migrate sys_role_menu: %v", err)
	}
	if err := db.AutoMigrate(&sysMenuApiRuleRow{}); err != nil {
		t.Fatalf("auto migrate sys_menu_api_rule: %v", err)
	}

	// Seed menus: ProductCenter (M) + SpuManage (C) + spu:add button (F).
	// Minimal representative subset; enough to verify casbin derivation.
	rootMenu := models.SysMenu{MenuName: "ProductCenter", Title: "产品中心", MenuType: "M", ParentId: 0}
	if err := db.Create(&rootMenu).Error; err != nil {
		t.Fatalf("seed root menu: %v", err)
	}
	spuC := models.SysMenu{MenuName: "SpuManage", Title: "产品 SPU 管理", MenuType: "C",
		Permission: "admin:spu:list", ParentId: rootMenu.MenuId}
	if err := db.Create(&spuC).Error; err != nil {
		t.Fatalf("seed spu C menu: %v", err)
	}
	spuMenuId = spuC.MenuId

	spuAddBtn := models.SysMenu{MenuName: "", Title: "新增 SPU", MenuType: "F",
		Permission: "admin:spu:add", ParentId: spuC.MenuId}
	if err := db.Create(&spuAddBtn).Error; err != nil {
		t.Fatalf("seed spu add button: %v", err)
	}
	spuAddBtnId = spuAddBtn.MenuId

	// Seed sys_api: GET /api/v1/spu (list) + POST /api/v1/spu (add)
	listApi := models.SysApi{Title: "SPU 列表", Path: "/api/v1/spu", Action: "GET", Type: "BUS"}
	if err := db.Create(&listApi).Error; err != nil {
		t.Fatalf("seed list api: %v", err)
	}
	addApi := models.SysApi{Title: "SPU 创建", Path: "/api/v1/spu", Action: "POST", Type: "BUS"}
	if err := db.Create(&addApi).Error; err != nil {
		t.Fatalf("seed add api: %v", err)
	}

	// Seed sys_menu_api_rule bridges: SpuManage -> GET, spuAdd -> POST
	if err := db.Exec("INSERT INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
		spuC.MenuId, listApi.Id).Error; err != nil {
		t.Fatalf("seed bridge list: %v", err)
	}
	if err := db.Exec("INSERT INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
		spuAddBtn.MenuId, addApi.Id).Error; err != nil {
		t.Fatalf("seed bridge add: %v", err)
	}

	return db, spuMenuId, spuAddBtnId
}

// TestMigration_ProductRoleSeed_Fresh verifies a fresh-install run where neither
// product_admin nor product_operator exist yet, and spu_create_review has no node.
func TestMigration_ProductRoleSeed_Fresh(t *testing.T) {
	db, spuMenuId, spuAddBtnId := setupProductRoleSeedDB(t)

	// Seed wf_definition without a node (simulating 1779000000001 having skipped the node).
	def := models.WorkflowDefinition{
		DefinitionKey:  "spu_create_review",
		DefinitionName: "SPU 创建审核",
		ModuleKey:      "admin",
		BusinessType:   "spu",
		Status:         "2",
		Version:        1,
	}
	if err := db.Create(&def).Error; err != nil {
		t.Fatalf("seed wf_definition: %v", err)
	}

	const ver = "1779000000003_product_role_seed.go"
	if err := _1779000000003ProductRoleSeed(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// sys_role: two new roles
	var adminRole, operatorRole models.SysRole
	if err := db.Where("role_key = ?", "product_admin").First(&adminRole).Error; err != nil {
		t.Fatalf("product_admin role not found: %v", err)
	}
	if adminRole.DataScope != "1" {
		t.Fatalf("product_admin DataScope want=1, got=%s", adminRole.DataScope)
	}
	if err := db.Where("role_key = ?", "product_operator").First(&operatorRole).Error; err != nil {
		t.Fatalf("product_operator role not found: %v", err)
	}
	if operatorRole.DataScope != "5" {
		t.Fatalf("product_operator DataScope want=5, got=%s", operatorRole.DataScope)
	}

	// sys_role_menu: admin and operator are bound to SpuManage and spu:add button
	for _, roleId := range []int{adminRole.RoleId, operatorRole.RoleId} {
		var n int64
		if err := db.Table("sys_role_menu").Where("role_id = ? AND menu_id = ?", roleId, spuMenuId).Count(&n).Error; err != nil {
			t.Fatalf("count role_menu: %v", err)
		}
		if n != 1 {
			t.Fatalf("role=%d not bound to SpuManage (menu_id=%d), count=%d", roleId, spuMenuId, n)
		}
		if err := db.Table("sys_role_menu").Where("role_id = ? AND menu_id = ?", roleId, spuAddBtnId).Count(&n).Error; err != nil {
			t.Fatalf("count role_menu: %v", err)
		}
		if n != 1 {
			t.Fatalf("role=%d not bound to spu:add button (menu_id=%d), count=%d", roleId, spuAddBtnId, n)
		}
	}

	// sys_casbin_rule: each role gets at least the 2 API paths + 3 platform workflow paths
	for _, roleKey := range []string{"product_admin", "product_operator"} {
		var n int64
		if err := db.Table("sys_casbin_rule").
			Where("ptype = ? AND v0 = ?", "p", roleKey).
			Count(&n).Error; err != nil {
			t.Fatalf("count casbin for %s: %v", roleKey, err)
		}
		if n == 0 {
			t.Fatalf("no casbin rules for %s", roleKey)
		}
	}
	// product_admin gets platform workflow casbin rules
	var adminPlatformCount int64
	if err := db.Table("sys_casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "product_admin", "/api/v1/platform/workflow%").
		Count(&adminPlatformCount).Error; err != nil {
		t.Fatalf("count admin platform casbin: %v", err)
	}
	if adminPlatformCount != 3 {
		t.Fatalf("product_admin platform casbin rules want=3, got=%d", adminPlatformCount)
	}
	var opPlatformCount int64
	if err := db.Table("sys_casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "product_operator", "/api/v1/platform/workflow%").
		Count(&opPlatformCount).Error; err != nil {
		t.Fatalf("count operator platform casbin: %v", err)
	}
	if opPlatformCount != 2 {
		t.Fatalf("product_operator platform casbin rules want=2, got=%d", opPlatformCount)
	}

	// wf_definition_node: approve_1 node created with product_admin role_id
	var nodeCount int64
	if err := db.Table("wf_definition_node").
		Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").
		Count(&nodeCount).Error; err != nil {
		t.Fatalf("count node: %v", err)
	}
	if nodeCount != 1 {
		t.Fatalf("approve_1 node want=1, got=%d", nodeCount)
	}

	// sys_migration entry created
	var migCount int64
	if err := db.Table("sys_migration").Where("version = ?", ver).Count(&migCount).Error; err != nil {
		t.Fatalf("count migration: %v", err)
	}
	if migCount != 1 {
		t.Fatalf("sys_migration entry want=1, got=%d", migCount)
	}
}

// TestMigration_ProductRoleSeed_RolesExist verifies the idempotency case where
// an operator has manually created product_admin and product_operator before the
// migration runs. Existing roles must be reused, not duplicated.
func TestMigration_ProductRoleSeed_RolesExist(t *testing.T) {
	db, _, _ := setupProductRoleSeedDB(t)

	// Pre-create roles (simulating manual setup)
	preAdmin := models.SysRole{RoleName: "产品管理员", RoleKey: "product_admin", Status: "2", DataScope: "1"}
	if err := db.Create(&preAdmin).Error; err != nil {
		t.Fatalf("pre-create admin role: %v", err)
	}
	preOp := models.SysRole{RoleName: "产品操作员", RoleKey: "product_operator", Status: "2", DataScope: "5"}
	if err := db.Create(&preOp).Error; err != nil {
		t.Fatalf("pre-create operator role: %v", err)
	}

	const ver = "1779000000003_product_role_seed.go"
	if err := _1779000000003ProductRoleSeed(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// sys_role: still exactly 1 of each, not duplicated
	var adminCount, opCount int64
	if err := db.Table("sys_role").Where("role_key = ?", "product_admin").Count(&adminCount).Error; err != nil {
		t.Fatalf("count admin role: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("product_admin should be 1, got %d", adminCount)
	}
	if err := db.Table("sys_role").Where("role_key = ?", "product_operator").Count(&opCount).Error; err != nil {
		t.Fatalf("count operator role: %v", err)
	}
	if opCount != 1 {
		t.Fatalf("product_operator should be 1, got %d", opCount)
	}

	// sys_casbin_rule: platform workflow entries exist
	var adminPlatform int64
	if err := db.Table("sys_casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "product_admin", "/api/v1/platform/workflow%").
		Count(&adminPlatform).Error; err != nil {
		t.Fatalf("count admin platform casbin: %v", err)
	}
	if adminPlatform != 3 {
		t.Fatalf("product_admin platform casbin want=3, got=%d", adminPlatform)
	}
}

// TestMigration_ProductRoleSeed_NodeExists verifies that when spu_create_review
// already has an approve_1 node (simulating 1779000000001 ran with product_admin
// already in place), the migration does not duplicate the node.
func TestMigration_ProductRoleSeed_NodeExists(t *testing.T) {
	db, _, _ := setupProductRoleSeedDB(t)

	// Pre-create product_admin role
	adminRole := models.SysRole{RoleName: "产品管理员", RoleKey: "product_admin", Status: "2", DataScope: "1"}
	if err := db.Create(&adminRole).Error; err != nil {
		t.Fatalf("pre-create admin role: %v", err)
	}

	// Seed wf_definition AND the approve_1 node (simulating 1779000000001 ran cleanly)
	def := models.WorkflowDefinition{
		DefinitionKey: "spu_create_review", DefinitionName: "SPU 创建审核",
		ModuleKey: "admin", BusinessType: "spu", Status: "2", Version: 1,
	}
	if err := db.Create(&def).Error; err != nil {
		t.Fatalf("seed wf_definition: %v", err)
	}
	existingNode := models.WorkflowDefinitionNode{
		DefinitionId:  def.DefinitionId,
		NodeKey:       "approve_1",
		NodeName:      "产品管理员审批",
		NodeType:      "approve",
		Sort:          1,
		ApproverType:  "role",
		ApproverValue: "99",
		ApproverName:  "产品管理员",
	}
	if err := db.Create(&existingNode).Error; err != nil {
		t.Fatalf("seed existing node: %v", err)
	}

	const ver = "1779000000003_product_role_seed.go"
	if err := _1779000000003ProductRoleSeed(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// wf_definition_node: still exactly 1, not duplicated
	var nodeCount int64
	if err := db.Table("wf_definition_node").
		Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").
		Count(&nodeCount).Error; err != nil {
		t.Fatalf("count node: %v", err)
	}
	if nodeCount != 1 {
		t.Fatalf("approve_1 node should be 1 (not duplicated), got %d", nodeCount)
	}

	// The existing approver_value must be preserved (not overwritten)
	var node models.WorkflowDefinitionNode
	if err := db.Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").
		First(&node).Error; err != nil {
		t.Fatalf("read node: %v", err)
	}
	if node.ApproverValue != "99" {
		t.Fatalf("existing node approver_value must be preserved, got %q", node.ApproverValue)
	}
}
