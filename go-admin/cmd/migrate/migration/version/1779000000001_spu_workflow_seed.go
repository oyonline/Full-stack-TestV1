package version

import (
	"runtime"
	"strconv"

	"go-admin/cmd/migrate/migration"
	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000001SpuWorkflowSeed)
}

// _1779000000001SpuWorkflowSeed 落地 'SPU 创建审核' 默认流程定义 + 一个审批节点。
//
//   - definition_key: spu_create_review
//   - module_key:     admin
//   - business_type:  spu（与 service.SpuBusinessType 一致）
//   - 节点 1: approver_type=role, approver_value=<product_admin role_id>
//
// product_admin 角色不存在时（迁移基线尚未注入该角色），跳过节点种子，仅创建定义本身。
// 后台可通过 wf-definition 管理页补全节点；这避免硬卡迁移。
//
// 全部走 FirstOrCreate，幂等。
func _1779000000001SpuWorkflowSeed(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		def := models.WorkflowDefinition{
			DefinitionKey:  "spu_create_review",
			DefinitionName: "SPU 创建审核",
			ModuleKey:      "admin",
			BusinessType:   "spu",
			Status:         "2",
			Version:        1,
			Remark:         "C4 SPU 提交审核流程",
		}
		if err := tx.Where("definition_key = ?", def.DefinitionKey).
			FirstOrCreate(&def).Error; err != nil {
			return err
		}

		// 解析 product_admin 角色（如不存在，跳过节点种子）
		var roleId int
		row := struct {
			RoleId int `gorm:"column:role_id"`
		}{}
		err := tx.Table("sys_role").
			Select("role_id").
			Where("role_key = ?", "product_admin").
			Take(&row).Error
		if err == nil {
			roleId = row.RoleId
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		if roleId > 0 {
			node := models.WorkflowDefinitionNode{
				DefinitionId:  def.DefinitionId,
				NodeKey:       "approve_1",
				NodeName:      "产品管理员审批",
				NodeType:      "approve",
				Sort:          1,
				ApproverType:  "role",
				ApproverValue: strconv.Itoa(roleId),
				ApproverName:  "产品管理员",
			}
			if err := tx.Where("definition_id = ? AND node_key = ?", node.DefinitionId, node.NodeKey).
				FirstOrCreate(&node).Error; err != nil {
				return err
			}
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
