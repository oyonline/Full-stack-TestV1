package version

import (
	"strconv"
	"testing"

	"go-admin/cmd/migrate/migration/models"
)

// TestMigration_ProductDirectorManagerRoles_AfterProductRoleSeed 模拟 0003 已跑完：
// spu_create_review 的 approve_1 指向 product_admin；0005 应新增 director/manager
// 并把该节点重定向到 product_director。
func TestMigration_ProductDirectorManagerRoles_AfterProductRoleSeed(t *testing.T) {
	db, _, _ := setupProductRoleSeedDB(t)

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

	const ver3 = "1779000000003_product_role_seed.go"
	if err := _1779000000003ProductRoleSeed(db, ver3); err != nil {
		t.Fatalf("0003 migration run: %v", err)
	}

	var adminRole models.SysRole
	if err := db.Where("role_key = ?", "product_admin").First(&adminRole).Error; err != nil {
		t.Fatalf("product_admin: %v", err)
	}
	var nodeBefore models.WorkflowDefinitionNode
	if err := db.Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").First(&nodeBefore).Error; err != nil {
		t.Fatalf("approve_1 node before 0005: %v", err)
	}
	if nodeBefore.ApproverValue != strconv.Itoa(adminRole.RoleId) {
		t.Fatalf("pre-0005 node want approver_value=%d (product_admin), got %q",
			adminRole.RoleId, nodeBefore.ApproverValue)
	}

	const ver5 = "1779000000005_product_director_manager_roles.go"
	if err := _1779000000005ProductDirectorManagerRoles(db, ver5); err != nil {
		t.Fatalf("0005 migration run: %v", err)
	}

	var directorRole, managerRole models.SysRole
	if err := db.Where("role_key = ?", "product_director").First(&directorRole).Error; err != nil {
		t.Fatalf("product_director: %v", err)
	}
	if directorRole.DataScope != "1" {
		t.Fatalf("product_director DataScope want=1, got=%s", directorRole.DataScope)
	}
	if err := db.Where("role_key = ?", "product_manager").First(&managerRole).Error; err != nil {
		t.Fatalf("product_manager: %v", err)
	}
	if managerRole.DataScope != "5" {
		t.Fatalf("product_manager DataScope want=5, got=%s", managerRole.DataScope)
	}

	var nodeAfter models.WorkflowDefinitionNode
	if err := db.Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").First(&nodeAfter).Error; err != nil {
		t.Fatalf("approve_1 node after 0005: %v", err)
	}
	wantVal := strconv.Itoa(directorRole.RoleId)
	if nodeAfter.ApproverValue != wantVal {
		t.Fatalf("post-0005 node approver_value want=%s (product_director), got %q", wantVal, nodeAfter.ApproverValue)
	}

	var dirPlatform int64
	if err := db.Table("sys_casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "product_director", "/api/v1/platform/workflow%").
		Count(&dirPlatform).Error; err != nil {
		t.Fatalf("count director platform casbin: %v", err)
	}
	if dirPlatform != 3 {
		t.Fatalf("product_director platform casbin want=3, got=%d", dirPlatform)
	}
	var mgrPlatform int64
	if err := db.Table("sys_casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "product_manager", "/api/v1/platform/workflow%").
		Count(&mgrPlatform).Error; err != nil {
		t.Fatalf("count manager platform casbin: %v", err)
	}
	if mgrPlatform != 2 {
		t.Fatalf("product_manager platform casbin want=2, got=%d", mgrPlatform)
	}
}
