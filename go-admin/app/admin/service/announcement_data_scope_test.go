package service

import (
	"sort"
	"testing"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

// C7-5 端到端验收：用真实组织拓扑跑通 5 路 dataScope。
//
// 拓扑（A 公司 → 销售部 / 技术部 / 财务部，销售部 → 销售一组 / 销售二组）：
//
//	dept_path=/1/        dept 1   A 公司 (root)
//	  ├─ dept_path=/1/2/   dept 2   销售部
//	  │     ├─ dept_path=/1/2/21/  dept 21  销售一组
//	  │     └─ dept_path=/1/2/22/  dept 22  销售二组
//	  ├─ dept_path=/1/3/   dept 3   技术部
//	  └─ dept_path=/1/4/   dept 4   财务部
//
// 用户 / 角色：
//
//	user 1  admin          dept=1  role=1  data_scope=1   全部
//	user 2  sales_lead     dept=2  role=2  data_scope=4   本部门及以下
//	user 3  sales_member   dept=21 role=3  data_scope=5   仅本人（销售一组成员）
//	user 4  finance_member dept=4  role=4  data_scope=3   本部门
//	user 5  cross_dept     dept=3  role=5  data_scope=2   自定义（角色绑定 dept 3, 4）
//
// 公告样本（每条 create_by 一一对应上述用户）：
//
//	ann_admin    create_by=1  全员公告
//	ann_lead     create_by=2  销售部公告
//	ann_member   create_by=3  个人公告
//	ann_finance  create_by=4  财务公告
type personaSeed struct {
	UserId    int
	DeptId    int
	RoleId    int
	NickName  string
	DataScope string
}

var (
	personaAdmin    = personaSeed{UserId: 1, DeptId: 1, RoleId: 1, NickName: "admin", DataScope: "1"}
	personaSalesLd  = personaSeed{UserId: 2, DeptId: 2, RoleId: 2, NickName: "sales_lead", DataScope: "4"}
	personaSalesMb  = personaSeed{UserId: 3, DeptId: 21, RoleId: 3, NickName: "sales_member", DataScope: "5"}
	personaFinance  = personaSeed{UserId: 4, DeptId: 4, RoleId: 4, NickName: "finance_member", DataScope: "3"}
	personaCrossDpt = personaSeed{UserId: 5, DeptId: 3, RoleId: 5, NickName: "cross_dept", DataScope: "2"}
)

// seedDataScopeFixture 建立 C7-5 用例的部门树 + 用户 + 公告 + 角色绑定。
// 重复调用幂等：用 :memory: SQLite，每次返回新 DB。
func seedDataScopeFixture(t *testing.T) (*Announcement, func()) {
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
	); err != nil {
		cleanup()
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_role_dept (role_id INTEGER NOT NULL, dept_id INTEGER NOT NULL)`).Error; err != nil {
		cleanup()
		t.Fatalf("create sys_role_dept: %v", err)
	}

	depts := []models.SysDept{
		{DeptId: 1, DeptName: "A 公司", DeptPath: "/1/", ParentId: 0},
		{DeptId: 2, DeptName: "销售部", DeptPath: "/1/2/", ParentId: 1},
		{DeptId: 21, DeptName: "销售一组", DeptPath: "/1/2/21/", ParentId: 2},
		{DeptId: 22, DeptName: "销售二组", DeptPath: "/1/2/22/", ParentId: 2},
		{DeptId: 3, DeptName: "技术部", DeptPath: "/1/3/", ParentId: 1},
		{DeptId: 4, DeptName: "财务部", DeptPath: "/1/4/", ParentId: 1},
	}
	for i := range depts {
		if err := db.Create(&depts[i]).Error; err != nil {
			cleanup()
			t.Fatalf("seed dept %d: %v", depts[i].DeptId, err)
		}
	}

	personas := []personaSeed{personaAdmin, personaSalesLd, personaSalesMb, personaFinance, personaCrossDpt}
	for _, p := range personas {
		u := models.SysUser{UserId: p.UserId, NickName: p.NickName, DeptId: p.DeptId, RoleId: p.RoleId, Status: "2"}
		if err := db.Create(&u).Error; err != nil {
			cleanup()
			t.Fatalf("seed user %s: %v", p.NickName, err)
		}
	}

	// 角色 5（cross_dept）自定义范围：绑定 dept 3（技术部）+ dept 4（财务部）。
	for _, did := range []int{3, 4} {
		if err := db.Exec(`INSERT INTO sys_role_dept(role_id, dept_id) VALUES (?, ?)`, personaCrossDpt.RoleId, did).Error; err != nil {
			cleanup()
			t.Fatalf("seed sys_role_dept: %v", err)
		}
	}

	now := time.Now()
	samples := []struct {
		title    string
		creator  int
	}{
		{"全员公告", personaAdmin.UserId},
		{"销售部公告", personaSalesLd.UserId},
		{"个人公告", personaSalesMb.UserId},
		{"财务公告", personaFinance.UserId},
	}
	for _, s := range samples {
		ann := models.Announcement{
			Title:     s.title,
			Content:   "<p>" + s.title + "</p>",
			Status:    models.AnnouncementStatusPublished,
			PublishAt: &now,
		}
		ann.SetCreateBy(s.creator)
		if err := db.Create(&ann).Error; err != nil {
			cleanup()
			t.Fatalf("seed ann %q: %v", s.title, err)
		}
	}

	s := &Announcement{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s, cleanup
}

// runScopeCase 跑一遍 GetPage，按标题集合断言。
func runScopeCase(t *testing.T, p personaSeed, want []string) {
	t.Helper()
	s, cleanup := seedDataScopeFixture(t)
	defer cleanup()

	req := &dto.AnnouncementPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	dp := &actions.DataPermission{
		DataScope: p.DataScope,
		UserId:    p.UserId,
		DeptId:    p.DeptId,
		RoleId:    p.RoleId,
	}
	list := make([]dto.AnnouncementListItem, 0)
	var count int64
	if err := s.GetPage(req, dp, &list, &count, p.UserId); err != nil {
		t.Fatalf("%s GetPage: %v", p.NickName, err)
	}

	gotTitles := make([]string, 0, len(list))
	for _, it := range list {
		gotTitles = append(gotTitles, it.Announcement.Title)
	}
	sort.Strings(gotTitles)
	wantSorted := append([]string{}, want...)
	sort.Strings(wantSorted)

	if int(count) != len(want) {
		t.Fatalf("%s: count want=%d got=%d titles=%v", p.NickName, len(want), count, gotTitles)
	}
	if len(gotTitles) != len(wantSorted) {
		t.Fatalf("%s: titles want=%v got=%v", p.NickName, wantSorted, gotTitles)
	}
	for i := range wantSorted {
		if gotTitles[i] != wantSorted[i] {
			t.Fatalf("%s: titles want=%v got=%v", p.NickName, wantSorted, gotTitles)
		}
	}
}

// scope=1 全部：admin 看到 4 条全部公告。
func TestDataScope_All(t *testing.T) {
	runScopeCase(t, personaAdmin, []string{"全员公告", "销售部公告", "个人公告", "财务公告"})
}

// scope=2 自定义：cross_dept 角色绑定 {dept 3, dept 4}，
// 这些部门内的用户是 {user 5（自己）, user 4（finance_member）}，
// 公告 create_by 落在该集合的只有"财务公告"（cross_dept 自己没建公告）。
func TestDataScope_Custom(t *testing.T) {
	runScopeCase(t, personaCrossDpt, []string{"财务公告"})
}

// scope=3 本部门：finance_member 在 dept 4（财务部），
// 本部门用户 = {user 4}，可见公告 = "财务公告"。
func TestDataScope_Dept(t *testing.T) {
	runScopeCase(t, personaFinance, []string{"财务公告"})
}

// scope=4 本部门及以下：sales_lead 在 dept 2（销售部），
// dept_path LIKE '%/2/%' 命中 dept 2 / 21 / 22，
// 这些部门内用户 = {user 2, user 3}，可见公告 = "销售部公告" + "个人公告"。
func TestDataScope_DeptAndBelow(t *testing.T) {
	runScopeCase(t, personaSalesLd, []string{"销售部公告", "个人公告"})
}

// scope=5 仅本人：sales_member（user 3）只看到自己 create 的"个人公告"。
func TestDataScope_Self(t *testing.T) {
	runScopeCase(t, personaSalesMb, []string{"个人公告"})
}
