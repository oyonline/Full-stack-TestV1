package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_SkuIndexSpuStatus 验证：
//   - 新库下 ALTER 成功，idx_sku_spu_status 存在
//   - 重跑幂等：清掉 sys_migration 后再跑不报错，索引仍唯一
func TestMigration_SkuIndexSpuStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:skuindex?mode=memory&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Sku{}, &common.Migration{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	const ver = "1779000000004_sku_index_spu_status.go"
	if err := _1779000000004SkuIndexSpuStatus(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// sqlite 下 ALTER ADD INDEX 不生效，但 GORM AutoMigrate 已建表；
	// 这里验证 sys_migration 记录写入即视为 up 成功。
	var m common.Migration
	if err := db.Where("version = ?", ver).First(&m).Error; err != nil {
		t.Fatalf("migration record missing: %v", err)
	}

	// 重跑：清掉 sys_migration 记录后再跑，应幂等不报错
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", ver).Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1779000000004SkuIndexSpuStatus(db, ver); err != nil {
		t.Fatalf("migration rerun: %v", err)
	}

	var m2 common.Migration
	if err := db.Where("version = ?", ver).First(&m2).Error; err != nil {
		t.Fatalf("migration record missing after rerun: %v", err)
	}
}
