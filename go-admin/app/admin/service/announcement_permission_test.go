package service

import (
	"testing"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	platformModels "go-admin/app/platform/models"
	"go-admin/common/actions"
)

// 测试拓扑：
//
//	dept_path=/1/  dept 1（root）
//	  ├─ dept_path=/1/2/  dept 2
//	  │     ├─ user 11（CreateBy=11，dept=2）
//	  │     └─ user 12（CreateBy=12，dept=2）
//	  └─ dept_path=/1/3/  dept 3
//	        └─ user 21（CreateBy=21，dept=3）
//
// 公告：
//
//	ann1: create_by=11（dept 2）
//	ann2: create_by=12（dept 2）
//	ann3: create_by=21（dept 3）
//
// 当前用户视角：user 11（roleId=10, deptId=2）。
// 角色 10 在 sys_role_dept 中授权 dept 2（dataScope=2 自定义场景）。
//
// 5 个 dataScope 的预期结果（从 user 11 视角）：
//
//	"1" 全部              → ann1/ann2/ann3
//	"2" 自定义（角色绑定）→ ann1/ann2（角色 10 → dept 2 → user 11/12）
//	"3" 本部门            → ann1/ann2（dept 2 内 user 11/12）
//	"4" 本部门及以下      → ann1/ann2（dept_path LIKE '%/2/%' = dept 2 自身，无子部门）
//	"5" 仅本人            → ann1
func setupAnnouncementPermDB(t *testing.T) (*Announcement, func()) {
	t.Helper()

	prevEnable := config.ApplicationConfig.EnableDP
	config.ApplicationConfig.EnableDP = true
	cleanup := func() {
		config.ApplicationConfig.EnableDP = prevEnable
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		cleanup()
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Announcement{},
		&models.AnnouncementScope{},
		&models.AnnouncementReadLog{},
		&models.SysUser{},
		&models.SysDept{},
		&platformModels.AttachmentFile{},
	); err != nil {
		cleanup()
		t.Fatalf("auto migrate: %v", err)
	}
	// sys_role_dept 是 GORM many2many 关联表，没独立 model，手动建。
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_role_dept (role_id INTEGER NOT NULL, dept_id INTEGER NOT NULL)`).Error; err != nil {
		cleanup()
		t.Fatalf("create sys_role_dept: %v", err)
	}

	depts := []models.SysDept{
		{DeptId: 1, DeptName: "root", DeptPath: "/1/", ParentId: 0},
		{DeptId: 2, DeptName: "biz", DeptPath: "/1/2/", ParentId: 1},
		{DeptId: 3, DeptName: "ops", DeptPath: "/1/3/", ParentId: 1},
	}
	for i := range depts {
		if err := db.Create(&depts[i]).Error; err != nil {
			cleanup()
			t.Fatalf("seed dept: %v", err)
		}
	}

	users := []models.SysUser{
		{UserId: 11, NickName: "u11", DeptId: 2, RoleId: 10, Status: "2"},
		{UserId: 12, NickName: "u12", DeptId: 2, RoleId: 10, Status: "2"},
		{UserId: 21, NickName: "u21", DeptId: 3, RoleId: 20, Status: "2"},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			cleanup()
			t.Fatalf("seed user: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO sys_role_dept(role_id, dept_id) VALUES (?, ?)`, 10, 2).Error; err != nil {
		cleanup()
		t.Fatalf("seed sys_role_dept: %v", err)
	}

	now := time.Now()
	anns := []*models.Announcement{
		{Title: "a1", Content: "<p>a1</p>", Status: models.AnnouncementStatusPublished},
		{Title: "a2", Content: "<p>a2</p>", Status: models.AnnouncementStatusPublished},
		{Title: "a3", Content: "<p>a3</p>", Status: models.AnnouncementStatusPublished},
	}
	creators := []int{11, 12, 21}
	for i, a := range anns {
		a.SetCreateBy(creators[i])
		a.PublishAt = &now
		if err := db.Create(a).Error; err != nil {
			cleanup()
			t.Fatalf("seed ann: %v", err)
		}
	}

	s := &Announcement{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s, cleanup
}

func titleSet(items []dto.AnnouncementListItem) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it.Announcement.Title] = true
	}
	return m
}

func TestAnnouncement_GetPage_DataScope(t *testing.T) {
	cases := []struct {
		name      string
		dataScope string
		want      []string
	}{
		{name: "scope=1 全部", dataScope: "1", want: []string{"a1", "a2", "a3"}},
		{name: "scope=2 自定义按角色绑定部门", dataScope: "2", want: []string{"a1", "a2"}},
		{name: "scope=3 本部门", dataScope: "3", want: []string{"a1", "a2"}},
		{name: "scope=4 本部门及以下", dataScope: "4", want: []string{"a1", "a2"}},
		{name: "scope=5 仅本人", dataScope: "5", want: []string{"a1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, cleanup := setupAnnouncementPermDB(t)
			defer cleanup()

			req := &dto.AnnouncementPageReq{}
			req.PageIndex = 1
			req.PageSize = 50
			p := &actions.DataPermission{
				DataScope: tc.dataScope,
				UserId:    11,
				DeptId:    2,
				RoleId:    10,
			}
			list := make([]dto.AnnouncementListItem, 0)
			var count int64
			if err := s.GetPage(req, p, &list, &count, 11); err != nil {
				t.Fatalf("GetPage(scope=%s): %v", tc.dataScope, err)
			}
			if int(count) != len(tc.want) {
				t.Fatalf("scope=%s: count want=%d got=%d titles=%v", tc.dataScope, len(tc.want), count, list)
			}
			got := titleSet(list)
			for _, w := range tc.want {
				if !got[w] {
					t.Fatalf("scope=%s: missing %q in result %v", tc.dataScope, w, got)
				}
			}
		})
	}
}

// EnableDP=false 时，scope 短路放行，无论 DataPermission 是什么都返回全部。
// 这是过渡期（C7-1 后、C7-7 前）的关键保护机制。
func TestAnnouncement_GetPage_EnableDPDisabledShortCircuits(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()
	config.ApplicationConfig.EnableDP = false

	req := &dto.AnnouncementPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}
	list := make([]dto.AnnouncementListItem, 0)
	var count int64
	if err := s.GetPage(req, p, &list, &count, 11); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if count != 3 {
		t.Fatalf("EnableDP=false 应该看到全部 3 条，实际 %d", count)
	}
}

// 详情读：scope=5 的 user 11 不能读 user 21 create 的公告（ann3）。
func TestAnnouncement_Get_RespectsDataScope(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()

	// 先用 admin 视角拿到 ann3 的真实 ID
	adminP := &actions.DataPermission{DataScope: "1"}
	req := &dto.AnnouncementPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	list := make([]dto.AnnouncementListItem, 0)
	var count int64
	if err := s.GetPage(req, adminP, &list, &count, 0); err != nil {
		t.Fatalf("GetPage admin: %v", err)
	}
	var ann3Id int64
	for _, it := range list {
		if it.Announcement.Title == "a3" {
			ann3Id = it.Announcement.AnnouncementId
		}
	}
	if ann3Id == 0 {
		t.Fatalf("ann3 not found in admin view")
	}

	// scope=5 的 user 11 读 ann3 应失败
	getReq := &dto.AnnouncementGetReq{Id: ann3Id}
	var item dto.AnnouncementListItem
	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}
	if err := s.Get(getReq, p, &item, 11); err == nil {
		t.Fatalf("scope=5 user 不应能读 user 21 的公告 ann3")
	}

	// 同样的 user，读自己 create 的 ann1 应成功
	var ann1Id int64
	for _, it := range list {
		if it.Announcement.Title == "a1" {
			ann1Id = it.Announcement.AnnouncementId
		}
	}
	getReq.Id = ann1Id
	item = dto.AnnouncementListItem{}
	if err := s.Get(getReq, p, &item, 11); err != nil {
		t.Fatalf("scope=5 读自己 create 的 ann1 不应失败: %v", err)
	}
	if item.Announcement.Title != "a1" {
		t.Fatalf("expected a1, got %s", item.Announcement.Title)
	}
}

// Update：scope=5 的用户不能改他人公告，命中 scope 外返回"公告不存在或已删除"。
func TestAnnouncement_Update_DataScopeBlocksCrossUser(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()

	var ann3 models.Announcement
	if err := s.Orm.Where("title = ?", "a3").First(&ann3).Error; err != nil {
		t.Fatalf("locate ann3: %v", err)
	}

	updReq := &dto.AnnouncementUpdateReq{}
	updReq.AnnouncementId = ann3.AnnouncementId
	updReq.Title = "hijacked"
	updReq.Content = "<p>x</p>"
	updReq.Status = models.AnnouncementStatusPublished
	updReq.SetUpdateBy(11)

	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}
	err := s.Update(updReq, p)
	if err == nil {
		t.Fatalf("scope=5 不应能改他人公告")
	}

	// 验证标题没被改
	var after models.Announcement
	if err := s.Orm.First(&after, ann3.AnnouncementId).Error; err != nil {
		t.Fatalf("read after: %v", err)
	}
	if after.Title != "a3" {
		t.Fatalf("title 被越权改了：%q", after.Title)
	}
}

// Remove：scope=5 的用户传一组 ID（包含别人的），只能删自己的。
func TestAnnouncement_Remove_DataScopeFiltersIds(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()

	var anns []models.Announcement
	if err := s.Orm.Order("announcement_id ASC").Find(&anns).Error; err != nil {
		t.Fatalf("load anns: %v", err)
	}
	if len(anns) != 3 {
		t.Fatalf("expected 3 anns, got %d", len(anns))
	}

	// user 11 (scope=5) 传 [ann1.Id, ann3.Id]，应只删 ann1
	delReq := &dto.AnnouncementDeleteReq{
		Ids: []int64{anns[0].AnnouncementId, anns[2].AnnouncementId},
	}
	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}
	if err := s.Remove(delReq, p); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	var remaining []models.Announcement
	if err := s.Orm.Order("announcement_id ASC").Find(&remaining).Error; err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(remaining))
	}
	for _, r := range remaining {
		if r.Title == "a1" {
			t.Fatalf("a1 应被删除，但仍存在")
		}
	}

	// 再传一组完全在 scope 外的 ID（ann3），应返回错误而不是默删
	delReq2 := &dto.AnnouncementDeleteReq{Ids: []int64{anns[2].AnnouncementId}}
	if err := s.Remove(delReq2, p); err == nil {
		t.Fatalf("scope 外的 Id 应返回错误")
	}

	// ann3 仍存在
	var ann3After models.Announcement
	if err := s.Orm.First(&ann3After, anns[2].AnnouncementId).Error; err != nil {
		t.Fatalf("ann3 should still exist: %v", err)
	}
}

// MarkRead：scope=5 的用户对 scope 外公告标记已读应失败（不写入 read_log）。
func TestAnnouncement_MarkRead_DataScopeBlocked(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()

	var ann3 models.Announcement
	if err := s.Orm.Where("title = ?", "a3").First(&ann3).Error; err != nil {
		t.Fatalf("locate ann3: %v", err)
	}

	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}
	if err := s.MarkRead(ann3.AnnouncementId, 11, p); err == nil {
		t.Fatalf("scope=5 user 不应能 mark 他人公告 ann3 已读")
	}

	var cnt int64
	if err := s.Orm.Model(&models.AnnouncementReadLog{}).
		Where("announcement_id = ? AND user_id = ?", ann3.AnnouncementId, 11).
		Count(&cnt).Error; err != nil {
		t.Fatalf("count reads: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("read_log 不应被写入，实际 %d 条", cnt)
	}

	// 自己的 ann1 应能 mark 已读
	var ann1 models.Announcement
	if err := s.Orm.Where("title = ?", "a1").First(&ann1).Error; err != nil {
		t.Fatalf("locate ann1: %v", err)
	}
	if err := s.MarkRead(ann1.AnnouncementId, 11, p); err != nil {
		t.Fatalf("mark own ann should succeed: %v", err)
	}
}

// OnlyVisible（按 announcement_scope 部门可见性）与 dataScope 正交：两者都满足才出现。
//
// 场景：ann4 由 user 21 create（dept 3），但 announcement_scope 显式可见到 dept 2。
// user 11（dept 2，dataScope=5 仅本人）：
//   - 默认 OnlyVisible=0：dataScope=5 → 看不到 ann4（因为不是 user 11 create）
//   - OnlyVisible=1：announcement_scope 包含 dept 2，但 dataScope=5 仍生效 → 还是看不到
//
// 这条用例同时证明：dataScope 不是 announcement_scope 的子集替代关系，是叠加 AND。
func TestAnnouncement_OnlyVisibleAndDataScope_Orthogonal(t *testing.T) {
	s, cleanup := setupAnnouncementPermDB(t)
	defer cleanup()

	now := time.Now()
	ann4 := models.Announcement{
		Title:     "a4",
		Content:   "<p>a4</p>",
		Status:    models.AnnouncementStatusPublished,
		PublishAt: &now,
	}
	ann4.SetCreateBy(21)
	if err := s.Orm.Create(&ann4).Error; err != nil {
		t.Fatalf("seed ann4: %v", err)
	}
	if err := s.Orm.Create(&models.AnnouncementScope{
		AnnouncementId: ann4.AnnouncementId,
		DeptId:         2,
	}).Error; err != nil {
		t.Fatalf("seed scope: %v", err)
	}

	// 给 user 11 准备 sys_user_depts（OnlyVisible 路径要查的关联表）
	if err := s.Orm.Exec(`CREATE TABLE IF NOT EXISTS sys_user_depts (user_id INTEGER, dept_id INTEGER)`).Error; err != nil {
		t.Fatalf("create sys_user_depts: %v", err)
	}
	if err := s.Orm.Exec(`INSERT INTO sys_user_depts(user_id, dept_id) VALUES (?, ?)`, 11, 2).Error; err != nil {
		t.Fatalf("seed sys_user_depts: %v", err)
	}

	p := &actions.DataPermission{DataScope: "5", UserId: 11, DeptId: 2, RoleId: 10}

	req := &dto.AnnouncementPageReq{OnlyVisible: 1}
	req.PageIndex = 1
	req.PageSize = 50
	list := make([]dto.AnnouncementListItem, 0)
	var count int64
	if err := s.GetPage(req, p, &list, &count, 11); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	got := titleSet(list)
	if got["a4"] {
		t.Fatalf("OnlyVisible 命中但 dataScope=5 不命中，ann4 不应出现，实际：%v", got)
	}
	// ann1 由 user 11 自己 create，但默认没在 announcement_scope 里 → OnlyVisible=1 又会过滤掉 ann1
	// 所以这个 case 下结果应该为空集。
	if len(list) != 0 {
		t.Fatalf("OnlyVisible=1 ∧ dataScope=5（无交集）应为空集，实际 %v", got)
	}
}
