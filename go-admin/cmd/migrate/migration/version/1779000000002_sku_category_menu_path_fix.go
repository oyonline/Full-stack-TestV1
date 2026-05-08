package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000002SkuCategoryMenuPathFix)
}

// _1779000000002SkuCategoryMenuPathFix 修正 SkuCategoryManage 菜单的 component 路径。
//
// C4-A 将 SkuCategoryManage.component 写为 "admin/sku-category/index"，与本仓库
// views/admin/sys-* 命名约定不一致；前端 router/access.ts 的 mapComponent 找不到
// 对应 .vue 文件后会回退到 not-found，导致'类目管理'页无法进入。
//
// 此处仅幂等更新已存在的菜单行：仅当 component 仍为旧值时才覆盖，避免误改用户已手工
// 调整的路径。SPU/SKU/SkuBrand 三个兄弟菜单同样存在该不一致，但属于其各自 bead 的
// 处置范围（C4-C-C 品牌、C4-C-D SPU），不在本迁移内顺手修。
func _1779000000002SkuCategoryMenuPathFix(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"UPDATE sys_menu SET component = ? WHERE menu_name = ? AND component = ?",
			"admin/sys-sku-category/index",
			"SkuCategoryManage",
			"admin/sku-category/index",
		).Error; err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
