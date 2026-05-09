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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000003ProductRoleSeed)
}

// _1779000000003ProductRoleSeed 落地 product_admin / product_operator 两个角色的
// sys_role / sys_role_menu / sys_casbin_rule 种子，并回灌 spu_create_review 审批节点。
//
// 依赖 1779000000000_sku_module.go 已落库的 sys_menu / sys_api / sys_menu_api_rule。
// 1779000000001_spu_workflow_seed.go 在 product_admin 缺失时跳过的节点由本迁移补回。
//
// 全程单事务，幂等：FirstOrCreate for 角色/API/casbin；count-check for sys_role_menu；
// 仅当 approve_1 节点确实缺失时才 Create。
func _1779000000003ProductRoleSeed(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等建两个角色
		adminRole := models.SysRole{
			RoleName:  "产品管理员",
			RoleKey:   "product_admin",
			RoleSort:  10,
			Status:    "2",
			DataScope: "1",
			Remark:    "产品中心管理员/审批人",
		}
		if err := tx.Where("role_key = ?", adminRole.RoleKey).FirstOrCreate(&adminRole).Error; err != nil {
			return err
		}

		operatorRole := models.SysRole{
			RoleName:  "产品操作员",
			RoleKey:   "product_operator",
			RoleSort:  11,
			Status:    "2",
			DataScope: "5",
			Remark:    "产品中心操作员/SPU 提交人",
		}
		if err := tx.Where("role_key = ?", operatorRole.RoleKey).FirstOrCreate(&operatorRole).Error; err != nil {
			return err
		}

		// 2. 查询需要绑定的菜单
		// product_admin：ProductCenter + 4 C 菜单 + 全部 16 个按钮
		// product_operator：ProductCenter + 4 C 菜单 + spu/sku 8 个按钮（category/brand 只读）
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

		adminMenuIDs := append(pageMenuIDs, allButtonIDs...)
		opMenuIDs := append(pageMenuIDs, opButtonIDs...)

		// 3. 绑定 sys_role_menu（count-check，幂等，跨方言）
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
		if err := bindRoleMenus(adminRole.RoleId, adminMenuIDs); err != nil {
			return err
		}
		if err := bindRoleMenus(operatorRole.RoleId, opMenuIDs); err != nil {
			return err
		}

		// 4. 按 sys_menu_api_rule 联查，为绑定菜单生成 sys_casbin_rule
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
		if err := writeCasbinForMenus("product_admin", adminMenuIDs); err != nil {
			return err
		}
		if err := writeCasbinForMenus("product_operator", opMenuIDs); err != nil {
			return err
		}

		// 5. 落库 platform workflow API（5行 FirstOrCreate）+ 直接写 casbin
		// 这些 API 在 1779000000000 未落库；无须挂菜单，直接补 casbin 策略即可。
		type platformApiSpec struct {
			api     models.SysApi
			roleKey string
		}
		platformSpecs := []platformApiSpec{
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.ListTodo-fm", Title: "待办任务列表", Path: "/api/v1/platform/workflow/tasks/todo", Type: "BUS", Action: "GET"}, "product_admin"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.Approve-fm", Title: "审批通过", Path: "/api/v1/platform/workflow/tasks/:id/approve", Type: "BUS", Action: "POST"}, "product_admin"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowTask.Reject-fm", Title: "审批驳回", Path: "/api/v1/platform/workflow/tasks/:id/reject", Type: "BUS", Action: "POST"}, "product_admin"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowInstance.ListStarted-fm", Title: "我发起的流程", Path: "/api/v1/platform/workflow/instances/started", Type: "BUS", Action: "GET"}, "product_operator"},
			{models.SysApi{Handle: "go-admin/app/platform/apis.WorkflowInstance.Withdraw-fm", Title: "撤回流程", Path: "/api/v1/platform/workflow/instances/:id/withdraw", Type: "BUS", Action: "POST"}, "product_operator"},
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

		// 6. 回灌 wf_definition_node：若 spu_create_review 存在但缺 approve_1 则补种
		// 1779000000001 在 product_admin 不存在时跳过了此节点；本迁移负责补回。
		var def models.WorkflowDefinition
		if err := tx.Where("definition_key = ?", "spu_create_review").First(&def).Error; err == nil {
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
					NodeName:      "产品管理员审批",
					NodeType:      "approve",
					Sort:          1,
					ApproverType:  "role",
					ApproverValue: strconv.Itoa(adminRole.RoleId),
					ApproverName:  "产品管理员",
				}
				if err := tx.Create(&node).Error; err != nil {
					return err
				}
			}
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		// wf_definition 不存在时跳过节点回灌（整机全量 migrate 时 1779000000001 会跑到；此处不卡）

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
