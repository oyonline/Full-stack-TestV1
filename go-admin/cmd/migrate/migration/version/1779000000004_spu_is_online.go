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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000004SpuIsOnline)
}

// _1779000000004SpuIsOnline 为 SPU 落地上下架功能（EPO-88 方案 B）：
//
//  1. AutoMigrate spu 表，加 is_online 列（TINYINT(1) NOT NULL DEFAULT 0）
//  2. 已有数据迁移：status=3(Approved) 的 SPU → is_online=true
//  3. sys_menu 新增两个按钮：'SPU 上架'(admin:spu:online) / 'SPU 下架'(admin:spu:offline)
//  4. sys_api 新增两条：POST /api/v1/spu/:id/online 和 /offline
//  5. sys_menu_api_rule 桥接
//
// 全部走 FirstOrCreate / 既存检测，确保幂等。
func _1779000000004SpuIsOnline(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1) AutoMigrate：给 spu 表加 is_online 列
		if err := tx.AutoMigrate(new(models.Spu)); err != nil {
			return err
		}

		// 2) 已有数据迁移：status=3 的存量 SPU → is_online=true
		if err := tx.Exec(
			"UPDATE spu SET is_online = 1 WHERE status = 3 AND is_online = 0",
		).Error; err != nil {
			return err
		}

		// 3) 找到 SPU 菜单（SpuManage），新增两个按钮
		var spuMenu models.SysMenu
		if err := tx.Where("menu_name = ?", "SpuManage").First(&spuMenu).Error; err != nil {
			// SPU 菜单不存在时跳过按钮和 API 绑定（迁移基线未注入），防止卡死
			return tx.Create(&common.Migration{Version: version}).Error
		}

		type buttonSpec struct {
			Title string
			Perm  string
		}
		newButtons := []buttonSpec{
			{"SPU 上架", "admin:spu:online"},
			{"SPU 下架", "admin:spu:offline"},
		}
		buttons := make([]models.SysMenu, len(newButtons))
		for i, spec := range newButtons {
			b := models.SysMenu{
				Title: spec.Title, Icon: "app-group-fill",
				MenuType: "F", Permission: spec.Perm, ParentId: spuMenu.MenuId,
				Visible: "0", IsFrame: "1",
			}
			var existing models.SysMenu
			err := tx.Where("permission = ? AND parent_id = ?", b.Permission, b.ParentId).
				First(&existing).Error
			if err == nil {
				buttons[i] = existing
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(&b).Error; err != nil {
				return err
			}
			buttons[i] = b
		}

		// 4) sys_api：两条新接口
		newApis := []models.SysApi{
			{Handle: "go-admin/app/admin/apis.Spu.GoOnline-fm", Title: "SPU 上架", Path: "/api/v1/spu/:id/online", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.Spu.GoOffline-fm", Title: "SPU 下架", Path: "/api/v1/spu/:id/offline", Type: "BUS", Action: "POST"},
		}
		for i := range newApis {
			a := &newApis[i]
			var existing models.SysApi
			err := tx.Where("path = ? AND action = ?", a.Path, a.Action).First(&existing).Error
			if err == nil {
				newApis[i] = existing
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(a).Error; err != nil {
				return err
			}
		}

		// 5) sys_menu_api_rule 桥接：在线/下线按钮 → 对应 API
		bridges := []struct {
			MenuId int
			ApiId  int
		}{
			{buttons[0].MenuId, newApis[0].Id}, // online button → GoOnline API
			{buttons[1].MenuId, newApis[1].Id}, // offline button → GoOffline API
		}
		for _, br := range bridges {
			var n int64
			if err := tx.Table("sys_menu_api_rule").
				Where("sys_menu_menu_id = ? AND sys_api_id = ?", br.MenuId, br.ApiId).
				Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			if err := tx.Exec(
				"INSERT INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
				br.MenuId, br.ApiId,
			).Error; err != nil {
				return err
			}
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
