package service

import (
	"sort"
	"testing"

	"github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

// C7-6 端到端验收：用真实组织拓扑跑通 SPU 5 路 dataScope。
//
// 拓扑与 announcement_data_scope_test.go 完全对齐：
//
//	dept_path=/1/        dept 1   A 公司 (root)
//	  ├─ dept_path=/1/2/   dept 2   销售部
//	  │     ├─ dept_path=/1/2/21/  dept 21  销售一组
//	  │     └─ dept_path=/1/2/22/  dept 22  销售二组
//	  ├─ dept_path=/1/3/   dept 3   技术部
//	  └─ dept_path=/1/4/   dept 4   财务部
//
// 用户 / 角色（与 announcement fixture 1:1 对应，复用 personaSeed 定义）：
//
//	user 1  admin          dept=1  role=1  data_scope=1   全部
//	user 2  sales_lead     dept=2  role=2  data_scope=4   本部门及以下
//	user 3  sales_member   dept=21 role=3  data_scope=5   仅本人（销售一组成员）
//	user 4  finance_member dept=4  role=4  data_scope=3   本部门
//	user 5  cross_dept     dept=3  role=5  data_scope=2   自定义（角色绑定 dept 3, 4）
//
// SPU 样本（每位用户建一条 SPU）：
//
//	SPU-ADMIN    create_by=1
//	SPU-LEAD     create_by=2
//	SPU-MEMBER   create_by=3
//	SPU-FINANCE  create_by=4
//	SPU-CROSS    create_by=5

// seedSpuDataScopeFixture 建立部门树 + 用户 + SPU + 角色绑定；每次返回全新 :memory: SQLite。
// 重复调用幂等：每个 :memory: DB 是独立实例，测试之间互不干扰。
func seedSpuDataScopeFixture(t *testing.T) (*Spu, func()) {
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
		&models.Spu{},
		&models.SysUser{},
		&models.SysDept{},
	); err != nil {
		cleanup()
		t.Fatalf("auto migrate: %v", err)
	}
	// wf_business_binding は GetPage 内の binding lookup で必要（エラーは warning 扱いだが table を用意しておく）
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS wf_business_binding (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		module_key TEXT,
		business_type TEXT,
		business_id TEXT,
		workflow_status TEXT,
		title TEXT
	)`).Error; err != nil {
		cleanup()
		t.Fatalf("create wf_business_binding: %v", err)
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

	// 复用 announcement_data_scope_test.go 中定义的 personaSeed 变量
	personas := []personaSeed{personaAdmin, personaSalesLd, personaSalesMb, personaFinance, personaCrossDpt}
	for _, p := range personas {
		u := models.SysUser{UserId: p.UserId, NickName: p.NickName, DeptId: p.DeptId, RoleId: p.RoleId, Status: "2"}
		if err := db.Create(&u).Error; err != nil {
			cleanup()
			t.Fatalf("seed user %s: %v", p.NickName, err)
		}
	}

	// 角色 5（cross_dept）自定义范围：绑定 dept 3（技术部）+ dept 4（财务部）
	for _, did := range []int{3, 4} {
		if err := db.Exec(`INSERT INTO sys_role_dept(role_id, dept_id) VALUES (?, ?)`, personaCrossDpt.RoleId, did).Error; err != nil {
			cleanup()
			t.Fatalf("seed sys_role_dept: %v", err)
		}
	}

	// SPU 样本：每位用户建一条，SpuCode 作为断言 key
	type spuSample struct {
		code    string
		creator int
	}
	samples := []spuSample{
		{"SPU-ADMIN", personaAdmin.UserId},
		{"SPU-LEAD", personaSalesLd.UserId},
		{"SPU-MEMBER", personaSalesMb.UserId},
		{"SPU-FINANCE", personaFinance.UserId},
		{"SPU-CROSS", personaCrossDpt.UserId},
	}
	for _, s := range samples {
		sp := models.Spu{SpuCode: s.code, SpuName: s.code, Status: models.SpuStatusDraft}
		sp.SetCreateBy(s.creator)
		if err := db.Create(&sp).Error; err != nil {
			cleanup()
			t.Fatalf("seed spu %q: %v", s.code, err)
		}
	}

	svc := &Spu{}
	svc.Orm = db
	svc.Log = logger.NewHelper(logger.DefaultLogger)
	return svc, cleanup
}

// runSpuScopeCase 跑一遍 GetPage，按 SpuCode 集合断言，口径与 runScopeCase 对齐。
func runSpuScopeCase(t *testing.T, p personaSeed, wantCodes []string) {
	t.Helper()
	svc, cleanup := seedSpuDataScopeFixture(t)
	defer cleanup()

	req := &dto.SpuPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	dp := &actions.DataPermission{
		DataScope: p.DataScope,
		UserId:    p.UserId,
		DeptId:    p.DeptId,
		RoleId:    p.RoleId,
	}
	list := make([]dto.SpuListItem, 0)
	var count int64
	if err := svc.GetPage(req, dp, &list, &count); err != nil {
		t.Fatalf("%s GetPage: %v", p.NickName, err)
	}

	gotCodes := make([]string, 0, len(list))
	for _, it := range list {
		gotCodes = append(gotCodes, it.SpuCode)
	}
	sort.Strings(gotCodes)
	wantSorted := append([]string{}, wantCodes...)
	sort.Strings(wantSorted)

	if int(count) != len(wantCodes) {
		t.Fatalf("%s: count want=%d got=%d codes=%v", p.NickName, len(wantCodes), count, gotCodes)
	}
	if len(gotCodes) != len(wantSorted) {
		t.Fatalf("%s: codes want=%v got=%v", p.NickName, wantSorted, gotCodes)
	}
	for i := range wantSorted {
		if gotCodes[i] != wantSorted[i] {
			t.Fatalf("%s: codes want=%v got=%v", p.NickName, wantSorted, gotCodes)
		}
	}
}

// TestSpu_DataScope_All scope=1 全部：admin 看到全部 5 条 SPU。
func TestSpu_DataScope_All(t *testing.T) {
	runSpuScopeCase(t, personaAdmin,
		[]string{"SPU-ADMIN", "SPU-LEAD", "SPU-MEMBER", "SPU-FINANCE", "SPU-CROSS"})
}

// TestSpu_DataScope_Custom scope=2 自定义：cross_dept 角色绑定 {dept 3, dept 4}，
// 这些部门内用户 = {user 5（自己）, user 4（finance_member）}，
// 可见 SPU = SPU-CROSS + SPU-FINANCE。
// 反向验证：SPU-ADMIN / SPU-LEAD / SPU-MEMBER 均不可见。
func TestSpu_DataScope_Custom(t *testing.T) {
	runSpuScopeCase(t, personaCrossDpt, []string{"SPU-CROSS", "SPU-FINANCE"})
}

// TestSpu_DataScope_Dept scope=3 本部门：finance_member 在 dept 4（财务部），
// 本部门用户 = {user 4}，可见 SPU = SPU-FINANCE。
// 反向验证：其余 4 条 SPU 均不可见。
func TestSpu_DataScope_Dept(t *testing.T) {
	runSpuScopeCase(t, personaFinance, []string{"SPU-FINANCE"})
}

// TestSpu_DataScope_DeptAndBelow scope=4 本部门及以下：sales_lead 在 dept 2（销售部），
// dept_path LIKE '%/2/%' 命中 dept 2 / 21 / 22，
// 这些部门内用户 = {user 2, user 3}，可见 SPU = SPU-LEAD + SPU-MEMBER。
// 反向验证：SPU-ADMIN / SPU-FINANCE / SPU-CROSS 均不可见。
func TestSpu_DataScope_DeptAndBelow(t *testing.T) {
	runSpuScopeCase(t, personaSalesLd, []string{"SPU-LEAD", "SPU-MEMBER"})
}

// TestSpu_DataScope_Self scope=5 仅本人：sales_member（user 3）只看到自己的 SPU-MEMBER。
// 反向验证：其余 4 条 SPU（含同部门的其他用户）均不可见。
func TestSpu_DataScope_Self(t *testing.T) {
	runSpuScopeCase(t, personaSalesMb, []string{"SPU-MEMBER"})
}
