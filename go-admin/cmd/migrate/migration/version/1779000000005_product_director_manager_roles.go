package version

import (
	"fmt"
	"runtime"
	"strconv"

	"go-admin/cmd/migrate/migration"
	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000005ProductDirectorManagerRoles)
}

// _1779000000005ProductDirectorManagerRoles 落地 product_director（产品总监）/
// product_manager（产品经理）角色：sys_role / sys_role_menu / sys_casbin_rule，
// 并将 spu_create_review 的 approve_1 节点从 product_admin 重定向到 product_director
// （仅当节点仍为 role 且 approver_value 等于当前 product_admin.role_id 时），
// 以便与组织职责命名对齐；保留 product_admin / product_operator 不变。
//
// 依赖 1779000000000_sku_module 与 1779000000003_product_role_seed 已执行。
// 幂等：角色与 casbin FirstOrCreate；节点更新仅匹配 admin 时执行一次。
func _1779000000005ProductDirectorManagerRoles(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		directorRole := models.SysRole{
			RoleName:  "产品总监",
			RoleKey:   "product_director",
			RoleSort:  12,
			Status:    "2",
			DataScope: "1",
			Remark:    "产品中心审批人（SPU 创建审核等）",
		}
		if err := tx.Where("role_key = ?", directorRole.RoleKey).FirstOrCreate(&directorRole).Error; err != nil {
			return err
		}

		managerRole := models.SysRole{
			RoleName:  "产品经理",
			RoleKey:   "product_manager",
			RoleSort:  13,
			Status:    "2",
			DataScope: "5",
			Remark:    "产品中心创建人/提交人",
		}
		if err := tx.Where("role_key = ?", managerRole.RoleKey).FirstOrCreate(&managerRole).Error; err != nil {
			return err
		}

		pageMenuNames := []string{"ProductCenter", "SpuManage", "SkuManage", "SkuCategoryManage", "SkuBrandManage"}
		var pageMenus []models.SysMenu
		if err := tx.Where("menu_name IN ?", pageMenuNames).Find(&pageMenus).Error; err != nil {
			return err
		}
		pageMenuIDs := make([]int, 0, len(pageMenus))
		for _, m := range pageMenus {
			pageMenuIDs = append(pageMenuIDs, m.MenuId)
		}

		allButtonPerms := []string{
			"admin:spu:add", "admin:spu:edit", "admin:spu:remove", "admin:spu:query",
			"admin:sku:add", "admin:sku:edit", "admin:sku:remove", "admin:sku:query",
			"admin:category:add", "admin:category:edit", "admin:category:remove", "admin:category:query",
			"admin:brand:add", "admin:brand:edit", "admin:brand:remove", "admin:brand:query",
		}
		var allButtons []models.SysMenu
		if err := tx.Where("permission IN ? AND menu_type = ?", allButtonPerms, "F").Find(&allButtons).Error; err != nil {
			return err
		}
		allButtonIDs := make([]int, 0, len(allButtons))
		for _, b := range allButtons {
			allButtonIDs = append(allButtonIDs, b.MenuId)
		}

		opButtonPerms := []string{
			"admin:spu:add", "admin:spu:edit", "admin:spu:remove", "admin:spu:query",
			"admin:sku:add", "admin:sku:edit", "admin:sku:remove", "admin:sku:query",
		}
		var opButtons []models.SysMenu
		if err := tx.Where("permission IN ? AND menu_type = ?", opButtonPerms, "F").Find(&opButtons).Error; err != nil {
			return err
		}
		opButtonIDs := make([]int, 0, len(opButtons))
		for _, b := range opButtons {
			opButtonIDs = append(opButtonIDs, b.MenuId)
		}

		directorMenuIDs := append(pageMenuIDs, allButtonIDs...)
		managerMenuIDs := append(pageMenuIDs, opButtonIDs...)

		bindRoleMenus := func(roleId int, menuIds []int) error {
			for _, menuId := range menuIds {
				var n int64
				if err := tx.Table("sys_role_menu").
					Where("role_id = ? AND menu_id = ?", roleId, menuId).
					Count(&n).Error; err != nil {
					return err
				}
				if n > 0 {
					continue
				}
				if err := tx.Exec(
					"INSERT INTO sys_role_menu (role_id, menu_id) VALUES (?, ?)",
					roleId, menuId,
				).Error; err != nil {
					return fmt.Errorf("insert sys_role_menu role=%d menu=%d: %w", roleId, menuId, err)
				}
			}
			return nil
		}
		if err := bindRoleMenus(directorRole.RoleId, directorMenuIDs); err != nil {
			return err
		}
		if err := bindRoleMenus(managerRole.RoleId, managerMenuIDs); err != nil {
			return err
		}

		type apiEntry struct {
			Path   string
			Action string
		}
		writeCasbinForMenus := func(roleKey string, menuIds []int) error {
			if len(menuIds) == 0 {
				return nil
			}
			var entries []apiEntry
			if err := tx.Table("sys_menu_api_rule").
				Select("sys_api.path AS path, sys_api.action AS action").
				Joins("JOIN sys_api ON sys_api.id = sys_menu_api_rule.sys_api_id").
				Where("sys_menu_api_rule.sys_menu_menu_id IN ?", menuIds).
				Scan(&entries).Error; err != nil {
				return err
			}
			for _, e := range entries {
				rule := models.CasbinRule{Ptype: "p", V0: roleKey, V1: e.Path, V2: e.Action, V3: "", V4: "", V5: ""}
				if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", roleKey, e.Path, e.Action).
					FirstOrCreate(&rule).Error; err != nil {
					return fmt.Errorf("casbin rule (%s, %s, %s): %w", roleKey, e.Path, e.Action, err)
				}
			}
			return nil
		}
		if err := writeCasbinForMenus("product_director", directorMenuIDs); err != nil {
			return err
		}
		if err := writeCasbinForMenus("product_manager", managerMenuIDs); err != nil {
			return err
		}

		type platformApiSpec struct {
			api     models.SysApi
			roleKey string
		}
		platformSpecs := []platformApiSpec{
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.ListTodo-fm", Title: "待办任务列表", Path: "/api/v1/platform/workflow/tasks/todo", Type: "BUS", Action: "GET"}, "product_director"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.Approve-fm", Title: "审批通过", Path: "/api/v1/platform/workflow/tasks/:id/approve", Type: "BUS", Action: "POST"}, "product_director"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.Reject-fm", Title: "审批驳回", Path: "/api/v1/platform/workflow/tasks/:id/reject", Type: "BUS", Action: "POST"}, "product_director"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowInstance.ListStarted-fm", Title: "我发起的流程", Path: "/api/v1/platform/workflow/instances/started", Type: "BUS", Action: "GET"}, "product_manager"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowInstance.Withdraw-fm", Title: "撤回流程", Path: "/api/v1/platform/workflow/instances/:id/withdraw", Type: "BUS", Action: "POST"}, "product_manager"},
		}
		for i := range platformSpecs {
			a := &platformSpecs[i].api
			var existing models.SysApi
			err := tx.Where("path = ? AND action = ?", a.Path, a.Action).First(&existing).Error
			if err == nil {
				platformSpecs[i].api = existing
			} else if err != gorm.ErrRecordNotFound {
				return err
			} else {
				if err := tx.Create(a).Error; err != nil {
					return err
				}
			}
		}
		for _, ps := range platformSpecs {
			rule := models.CasbinRule{Ptype: "p", V0: ps.roleKey, V1: ps.api.Path, V2: ps.api.Action, V3: "", V4: "", V5: ""}
			if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", ps.roleKey, ps.api.Path, ps.api.Action).
				FirstOrCreate(&rule).Error; err != nil {
				return err
			}
		}

		var adminRole models.SysRole
		adminErr := tx.Where("role_key = ?", "product_admin").First(&adminRole).Error

		var def models.WorkflowDefinition
		if err := tx.Where("definition_key = ?", "spu_create_review").First(&def).Error; err == nil {
			if adminErr == nil {
				res := tx.Model(&models.WorkflowDefinitionNode{}).
					Where("definition_id = ? AND node_key = ? AND approver_type = ? AND approver_value = ?",
						def.DefinitionId, "approve_1", "role", strconv.Itoa(adminRole.RoleId)).
					Updates(map[string]interface{}{
						"approver_value": strconv.Itoa(directorRole.RoleId),
						"approver_name":  "产品总监审批",
					})
				if res.Error != nil {
					return res.Error
				}
			}
			var nodeCount int64
			if err := tx.Table("wf_definition_node").
				Where("definition_id = ? AND node_key = ?", def.DefinitionId, "approve_1").
				Count(&nodeCount).Error; err != nil {
				return err
			}
			if nodeCount == 0 {
				node := models.WorkflowDefinitionNode{
					DefinitionId:  def.DefinitionId,
					NodeKey:       "approve_1",
					NodeName:      "产品总监审批",
					NodeType:      "approve",
					Sort:          1,
					ApproverType:  "role",
					ApproverValue: strconv.Itoa(directorRole.RoleId),
					ApproverName:  "产品总监",
				}
				if err := tx.Create(&node).Error; err != nil {
					return err
				}
			}
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
