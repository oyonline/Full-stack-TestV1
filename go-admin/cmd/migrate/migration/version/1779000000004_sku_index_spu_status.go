package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000004SkuIndexSpuStatus)
}

// _1779000000004SkuIndexSpuStatus 给 sku 表加 (spu_id, status) 复合索引，
// 优化带状态过滤的 SKU 列表查询。
//
// 与 models/sku.go 中 GORM tag `index:idx_sku_spu_status,priority:1/2` 保持一致。
func _1779000000004SkuIndexSpuStatus(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 使用 GORM Migrator 建索引，跨库兼容（MySQL / SQLite / Postgres）。
		// 模型里已声明 index:idx_sku_spu_status，但 AutoMigrate 只在 fresh DB 时生效；
		// 对已存在表显式 CreateIndex 保证所有环境都能加上。
		if !tx.Migrator().HasIndex(&models.Sku{}, "idx_sku_spu_status") {
			if err := tx.Migrator().CreateIndex(&models.Sku{}, "idx_sku_spu_status"); err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}


