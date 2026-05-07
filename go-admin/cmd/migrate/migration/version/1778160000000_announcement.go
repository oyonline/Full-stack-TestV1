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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1778160000000Announcement)
}

// _1778160000000Announcement 落库公告管理 MVP：
//  1. AutoMigrate 三张业务表
//  2. INSERT 一父菜单 + 四按钮
//  3. INSERT 六个 sys_api 行
//  4. INSERT 六个 sys_menu_api_rule 桥接行
//
// 全部走 OnConflict DoNothing / 既存检测，确保幂等。
func _1778160000000Announcement(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			new(models.Announcement),
			new(models.AnnouncementScope),
			new(models.AnnouncementReadLog),
		); err != nil {
			return err
		}

		// 1) sys_menu：先建父菜单（非 ROOT 的父也按 component=admin/sys-announcement/index 落库），
		//    再建 4 个按钮（list/add/edit/remove）
		parentMenu := models.SysMenu{
			MenuName:   "AnnouncementManage",
			Title:      "公告管理",
			Icon:       "ant-design:notification-outlined",
			Path:       "/admin/sys-announcement",
			MenuType:   "C",
			Permission: "admin:announcement:list",
			ParentId:   2, // 系统管理菜单（与 sys-post 同级，参考 db.sql sys_menu (57)）
			Component:  "admin/sys-announcement/index",
			Sort:       55,
			Visible:    "0",
			IsFrame:    "1",
		}
		if err := tx.Where("menu_name = ?", parentMenu.MenuName).
			FirstOrCreate(&parentMenu).Error; err != nil {
			return err
		}
		// 兜底：FirstOrCreate 命中已存在记录时不会更新；保留现状即可。

		buttons := []models.SysMenu{
			{MenuName: "", Title: "新增公告", Icon: "app-group-fill", MenuType: "F",
				Permission: "admin:announcement:add", ParentId: parentMenu.MenuId, Visible: "0", IsFrame: "1"},
			{MenuName: "", Title: "修改公告", Icon: "app-group-fill", MenuType: "F",
				Permission: "admin:announcement:edit", ParentId: parentMenu.MenuId, Visible: "0", IsFrame: "1"},
			{MenuName: "", Title: "删除公告", Icon: "app-group-fill", MenuType: "F",
				Permission: "admin:announcement:remove", ParentId: parentMenu.MenuId, Visible: "0", IsFrame: "1"},
			{MenuName: "", Title: "标记公告已读", Icon: "app-group-fill", MenuType: "F",
				Permission: "admin:announcement:read", ParentId: parentMenu.MenuId, Visible: "0", IsFrame: "1"},
		}
		for i := range buttons {
			b := &buttons[i]
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
			if err := tx.Create(b).Error; err != nil {
				return err
			}
		}

		// 2) sys_api 六行
		apis := []models.SysApi{
			{Handle: "go-admin/app/admin/apis.Announcement.GetPage-fm", Title: "公告列表", Path: "/api/v1/announcement", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Announcement.Get-fm", Title: "公告详情", Path: "/api/v1/announcement/:id", Type: "BUS", Action: "GET"},
			{Handle: "go-admin/app/admin/apis.Announcement.Insert-fm", Title: "公告创建", Path: "/api/v1/announcement", Type: "BUS", Action: "POST"},
			{Handle: "go-admin/app/admin/apis.Announcement.Update-fm", Title: "公告更新", Path: "/api/v1/announcement/:id", Type: "BUS", Action: "PUT"},
			{Handle: "go-admin/app/admin/apis.Announcement.Delete-fm", Title: "公告删除", Path: "/api/v1/announcement", Type: "BUS", Action: "DELETE"},
			{Handle: "go-admin/app/admin/apis.Announcement.MarkRead-fm", Title: "公告标记已读", Path: "/api/v1/announcement/:id/read", Type: "BUS", Action: "POST"},
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

		// 3) sys_menu_api_rule 桥接：父菜单可见对应 list/get；写动作（add/edit/remove）走对应权限按钮；mark-read 挂在 list 权限上（前台进入页面即可调）
		bridges := []struct {
			MenuId int
			ApiId  int
		}{
			{parentMenu.MenuId, apis[0].Id}, // GetPage 列表 -> 父菜单
			{parentMenu.MenuId, apis[1].Id}, // Get 详情 -> 父菜单
			{buttons[0].MenuId, apis[2].Id}, // Insert -> add 按钮
			{buttons[1].MenuId, apis[3].Id}, // Update -> edit 按钮
			{buttons[2].MenuId, apis[4].Id}, // Delete -> remove 按钮
			{parentMenu.MenuId, apis[5].Id}, // MarkRead -> 父菜单（任何进入详情的用户都需要）
		}
		for _, br := range bridges {
			if err := tx.Exec(
				"INSERT IGNORE INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
				br.MenuId, br.ApiId,
			).Error; err != nil {
				return err
			}
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
