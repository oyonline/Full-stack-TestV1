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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000000SkuModule)
}

// _1779000000000SkuModule 落库 SKU 数据层 (C4-A)：
//  1. AutoMigrate 4 张业务表（sku_category / sku_brand / spu / sku）
//  2. INSERT sys_menu：'产品中心'根 + 4 个子菜单（SPU/SKU/类目/品牌）+ 16 个按钮
//  3. INSERT sys_api：19 行（SPU 6 + SKU 5 + Category 4 + Brand 4）
//  4. INSERT sys_menu_api_rule 桥接行
//
// 全部走 FirstOrCreate / 既存检测 + INSERT IGNORE，确保幂等。
func _1779000000000SkuModule(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			new(models.SkuCategory),
			new(models.SkuBrand),
			new(models.Spu),
			new(models.Sku),
		); err != nil {
			return err
		}

		// 1) sys_menu：'产品中心' 顶级菜单（M, parent_id=0）
		rootMenu := models.SysMenu{
			MenuName:  "ProductCenter",
			Title:     "产品中心",
			Icon:      "shopping",
			Path:      "/product",
			MenuType:  "M",
			ParentId:  0,
			Component: "Layout",
			Sort:      60,
			Visible:   "0",
			IsFrame:   "1",
		}
		if err := tx.Where("menu_name = ? AND parent_id = ?", rootMenu.MenuName, rootMenu.ParentId).
			FirstOrCreate(&rootMenu).Error; err != nil {
			return err
		}

		// 4 个 C 子菜单（list 权限挂在 C 上）
		spuMenu := models.SysMenu{
			MenuName: "SpuManage", Title: "产品 SPU 管理", Icon: "shopping-cart",
			Path: "/product/spu", MenuType: "C", Permission: "admin:spu:list",
			ParentId: rootMenu.MenuId, Component: "admin/spu/index",
			Sort: 10, Visible: "0", IsFrame: "1",
		}
		skuMenu := models.SysMenu{
			MenuName: "SkuManage", Title: "产品 SKU 管理", Icon: "tag",
			Path: "/product/sku", MenuType: "C", Permission: "admin:sku:list",
			ParentId: rootMenu.MenuId, Component: "admin/sku/index",
			Sort: 20, Visible: "0", IsFrame: "1",
		}
		categoryMenu := models.SysMenu{
			MenuName: "SkuCategoryManage", Title: "类目管理", Icon: "tree",
			Path: "/product/category", MenuType: "C", Permission: "admin:category:list",
			ParentId: rootMenu.MenuId, Component: "admin/sku-category/index",
			Sort: 30, Visible: "0", IsFrame: "1",
		}
		brandMenu := models.SysMenu{
			MenuName: "SkuBrandManage", Title: "品牌管理", Icon: "star",
			Path: "/product/brand", MenuType: "C", Permission: "admin:brand:list",
			ParentId: rootMenu.MenuId, Component: "admin/sku-brand/index",
			Sort: 40, Visible: "0", IsFrame: "1",
		}
		pageMenus := []*models.SysMenu{&spuMenu, &skuMenu, &categoryMenu, &brandMenu}
		for _, m := range pageMenus {
			if err := tx.Where("menu_name = ?", m.MenuName).FirstOrCreate(m).Error; err != nil {
				return err
			}
		}

		// 4 套按钮（add/edit/remove/query），每套挂在对应 C 菜单下
		type buttonSpec struct {
			ParentId int
			Title    string
			Perm     string
		}
		buttonSpecs := []buttonSpec{
			{spuMenu.MenuId, "新增 SPU", "admin:spu:add"},
			{spuMenu.MenuId, "修改 SPU", "admin:spu:edit"},
			{spuMenu.MenuId, "删除 SPU", "admin:spu:remove"},
			{spuMenu.MenuId, "查询 SPU", "admin:spu:query"},

			{skuMenu.MenuId, "新增 SKU", "admin:sku:add"},
			{skuMenu.MenuId, "修改 SKU", "admin:sku:edit"},
			{skuMenu.MenuId, "删除 SKU", "admin:sku:remove"},
			{skuMenu.MenuId, "查询 SKU", "admin:sku:query"},

			{categoryMenu.MenuId, "新增类目", "admin:category:add"},
			{categoryMenu.MenuId, "修改类目", "admin:category:edit"},
			{categoryMenu.MenuId, "删除类目", "admin:category:remove"},
			{categoryMenu.MenuId, "查询类目", "admin:category:query"},

			{brandMenu.MenuId, "新增品牌", "admin:brand:add"},
			{brandMenu.MenuId, "修改品牌", "admin:brand:edit"},
			{brandMenu.MenuId, "删除品牌", "admin:brand:remove"},
			{brandMenu.MenuId, "查询品牌", "admin:brand:query"},
		}
		buttons := make([]models.SysMenu, len(buttonSpecs))
		for i, spec := range buttonSpecs {
			b := models.SysMenu{
				MenuName: "", Title: spec.Title, Icon: "app-group-fill",
				MenuType: "F", Permission: spec.Perm, ParentId: spec.ParentId,
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

		// 2) sys_api 19 行
		apis := []models.SysApi{
			// SPU 6
			{Handle: "go-admin/app/admin/apis.Spu.GetPage-fm", Title: "SPU 列表", Path: "/api/v1/spu", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Spu.Get-fm", Title: "SPU 详情", Path: "/api/v1/spu/:id", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Spu.Insert-fm", Title: "SPU 创建", Path: "/api/v1/spu", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.Spu.Update-fm", Title: "SPU 更新", Path: "/api/v1/spu/:id", Type: "BUS", Action: "PUT"},
			{Handle: "go-admin/app/admin/apis.Spu.Delete-fm", Title: "SPU 删除", Path: "/api/v1/spu", Type: "BUS", Action: "DELETE"},
			{Handle: "go-admin/app/admin/apis.Spu.Submit-fm", Title: "SPU 提交审核", Path: "/api/v1/spu/:id/submit", Type: "BUS", Action: "POST"},

			// SKU 5
			{Handle: "go-admin/app/admin/apis.Sku.GetPage-fm", Title: "SKU 列表", Path: "/api/v1/sku", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Sku.Get-fm", Title: "SKU 详情", Path: "/api/v1/sku/:id", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Sku.Insert-fm", Title: "SKU 创建", Path: "/api/v1/sku", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.Sku.Update-fm", Title: "SKU 更新", Path: "/api/v1/sku/:id", Type: "BUS", Action: "PUT"},
			{Handle: "go-admin/app/admin/apis.Sku.Delete-fm", Title: "SKU 删除", Path: "/api/v1/sku", Type: "BUS", Action: "DELETE"},

			// Category 4
			{Handle: "go-admin/app/admin/apis.SkuCategory.GetTree-fm", Title: "类目树", Path: "/api/v1/sku-category", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.SkuCategory.Insert-fm", Title: "类目创建", Path: "/api/v1/sku-category", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.SkuCategory.Update-fm", Title: "类目更新", Path: "/api/v1/sku-category/:id", Type: "BUS", Action: "PUT"},
			{Handle: "go-admin/app/admin/apis.SkuCategory.Delete-fm", Title: "类目删除", Path: "/api/v1/sku-category", Type: "BUS", Action: "DELETE"},

			// Brand 4
			{Handle: "go-admin/app/admin/apis.SkuBrand.GetPage-fm", Title: "品牌列表", Path: "/api/v1/sku-brand", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.SkuBrand.Insert-fm", Title: "品牌创建", Path: "/api/v1/sku-brand", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.SkuBrand.Update-fm", Title: "品牌更新", Path: "/api/v1/sku-brand/:id", Type: "BUS", Action: "PUT"},
			{Handle: "go-admin/app/admin/apis.SkuBrand.Delete-fm", Title: "品牌删除", Path: "/api/v1/sku-brand", Type: "BUS", Action: "DELETE"},
		}
		for i := range apis {
			a := &apis[i]
			var existing models.SysApi
			err := tx.Where("path = ? AND action = ?", a.Path, a.Action).
				First(&existing).Error
			if err == nil {
				apis[i] = existing
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(a).Error; err != nil {
				return err
			}
		}

		// 3) sys_menu_api_rule 桥接：
		//   - 列表 / 查询类 GET → C 菜单（list 权限）
		//   - 写动作 → 对应 add/edit/remove 按钮
		// 按钮索引：spu 0..3, sku 4..7, category 8..11, brand 12..15
		const (
			spuAdd, spuEdit, spuRemove, spuQuery                 = 0, 1, 2, 3
			skuAdd, skuEdit, skuRemove, _skuQuery                = 4, 5, 6, 7
			catAdd, catEdit, catRemove, _catQuery                = 8, 9, 10, 11
			brandAdd, brandEdit, brandRemove, _brandQuery        = 12, 13, 14, 15
		)
		_ = _skuQuery
		_ = _catQuery
		_ = _brandQuery

		bridges := []struct {
			MenuId int
			ApiId  int
		}{
			// SPU
			{spuMenu.MenuId, apis[0].Id},          // GetPage → list
			{buttons[spuQuery].MenuId, apis[1].Id}, // Get → query
			{buttons[spuAdd].MenuId, apis[2].Id},
			{buttons[spuEdit].MenuId, apis[3].Id},
			{buttons[spuRemove].MenuId, apis[4].Id},
			{buttons[spuEdit].MenuId, apis[5].Id}, // Submit → edit 权限

			// SKU
			{skuMenu.MenuId, apis[6].Id},
			{skuMenu.MenuId, apis[7].Id},
			{buttons[skuAdd].MenuId, apis[8].Id},
			{buttons[skuEdit].MenuId, apis[9].Id},
			{buttons[skuRemove].MenuId, apis[10].Id},

			// Category
			{categoryMenu.MenuId, apis[11].Id},
			{buttons[catAdd].MenuId, apis[12].Id},
			{buttons[catEdit].MenuId, apis[13].Id},
			{buttons[catRemove].MenuId, apis[14].Id},

			// Brand
			{brandMenu.MenuId, apis[15].Id},
			{buttons[brandAdd].MenuId, apis[16].Id},
			{buttons[brandEdit].MenuId, apis[17].Id},
			{buttons[brandRemove].MenuId, apis[18].Id},
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
