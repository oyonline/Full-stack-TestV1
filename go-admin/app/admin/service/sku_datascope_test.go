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

type skuDSFixture struct {
	Sku     *Sku
	DB      *gorm.DB
	Cleanup func()
}

func newSkuDSFixture(t *testing.T) *skuDSFixture {
	t.Helper()

	prev := config.ApplicationConfig.EnableDP
	config.ApplicationConfig.EnableDP = true
	cleanup := func() { config.ApplicationConfig.EnableDP = prev }

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		cleanup()
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Sku{},
		&models.Spu{},
		&models.SysUser{},
		&models.SysRole{},
		&models.SysDept{},
	); err != nil {
		cleanup()
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_role_dept (role_id INTEGER NOT NULL, dept_id INTEGER NOT NULL)`).Error; err != nil {
		cleanup()
		t.Fatalf("create sys_role_dept: %v", err)
	}
	for _, tbl := range []string{"sku", "spu", "sys_user", "sys_role", "sys_dept", "sys_role_dept"} {
		if err := db.Exec("DELETE FROM " + tbl).Error; err != nil {
			cleanup()
			t.Fatalf("clean %s: %v", tbl, err)
		}
	}

	s := &Sku{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return &skuDSFixture{Sku: s, DB: db, Cleanup: cleanup}
}

func (f *skuDSFixture) seedSpu(t *testing.T, code string, createBy int) int64 {
	t.Helper()
	spu := models.Spu{SpuCode: code, SpuName: code, Status: models.SpuStatusApproved}
	spu.CreateBy = createBy
	spu.CreatorId = createBy
	if err := f.DB.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu %s: %v", code, err)
	}
	return spu.SpuId
}

func (f *skuDSFixture) seedSku(t *testing.T, code string, spuId int64) {
	t.Helper()
	sku := models.Sku{SkuCode: code, SkuName: code, SpuId: spuId, Status: models.SkuStatusEnabled}
	if err := f.DB.Create(&sku).Error; err != nil {
		t.Fatalf("seed sku %s: %v", code, err)
	}
}

func (f *skuDSFixture) getPage(t *testing.T, dp *actions.DataPermission) []string {
	t.Helper()
	req := &dto.SkuPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	list := make([]dto.SkuListItem, 0)
	var count int64
	if err := f.Sku.GetPage(req, dp, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	codes := make([]string, 0, len(list))
	for _, it := range list {
		codes = append(codes, it.SkuCode)
	}
	sort.Strings(codes)
	return codes
}

// TestSkuDataScope_SelfOnly 验证 dataScope=5：只看自己 SPU 下的 SKU。
func TestSkuDataScope_SelfOnly(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	spuA := f.seedSpu(t, "SPU-A", 10)
	spuB := f.seedSpu(t, "SPU-B", 20)
	f.seedSku(t, "SKU-A1", spuA)
	f.seedSku(t, "SKU-A2", spuA)
	f.seedSku(t, "SKU-B1", spuB)

	// user 10: dataScope=5 → 只看 create_by=10 的 SPU 下的 SKU
	dp10 := &actions.DataPermission{DataScope: "5", UserId: 10}
	got10 := f.getPage(t, dp10)
	if len(got10) != 2 || got10[0] != "SKU-A1" || got10[1] != "SKU-A2" {
		t.Fatalf("dataScope=5 user10: want [SKU-A1, SKU-A2], got %v", got10)
	}

	// user 20: dataScope=5 → 只看 create_by=20 的 SPU 下的 SKU
	dp20 := &actions.DataPermission{DataScope: "5", UserId: 20}
	got20 := f.getPage(t, dp20)
	if len(got20) != 1 || got20[0] != "SKU-B1" {
		t.Fatalf("dataScope=5 user20: want [SKU-B1], got %v", got20)
	}
}

// TestSkuDataScope_AllData 验证 dataScope=1：admin 看到所有 SKU。
func TestSkuDataScope_AllData(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	spuA := f.seedSpu(t, "SPU-X", 11)
	spuB := f.seedSpu(t, "SPU-Y", 22)
	f.seedSku(t, "SKU-X1", spuA)
	f.seedSku(t, "SKU-Y1", spuB)

	dpAdmin := &actions.DataPermission{DataScope: "1", UserId: 1, DeptId: 1, RoleId: 1}
	got := f.getPage(t, dpAdmin)
	if len(got) != 2 {
		t.Fatalf("dataScope=1 admin: want 2 SKUs, got %v", got)
	}
}

// TestSkuDataScope_DeptOnly 验证 dataScope=3：只看本部门用户创建的 SPU 下的 SKU。
func TestSkuDataScope_DeptOnly(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	// 部门 100=dev，200=sales
	depts := []models.SysDept{
		{DeptId: 100, DeptName: "dev", DeptPath: "/100/", ParentId: 0},
		{DeptId: 200, DeptName: "sales", DeptPath: "/200/", ParentId: 0},
	}
	for i := range depts {
		if err := f.DB.Create(&depts[i]).Error; err != nil {
			t.Fatalf("seed dept: %v", err)
		}
	}
	users := []models.SysUser{
		{UserId: 41, NickName: "dev-user", DeptId: 100, RoleId: 60, Status: "2"},
		{UserId: 42, NickName: "sales-user", DeptId: 200, RoleId: 61, Status: "2"},
	}
	for i := range users {
		if err := f.DB.Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	spuDev := f.seedSpu(t, "SPU-DEV", 41)
	spuSales := f.seedSpu(t, "SPU-SALES", 42)
	f.seedSku(t, "SKU-DEV1", spuDev)
	f.seedSku(t, "SKU-SALES1", spuSales)

	// dev user: dataScope=3, dept=100 → 只看 dept 100 用户创建的 SPU 下的 SKU
	dpDev := &actions.DataPermission{DataScope: "3", UserId: 41, DeptId: 100, RoleId: 60}
	gotDev := f.getPage(t, dpDev)
	if len(gotDev) != 1 || gotDev[0] != "SKU-DEV1" {
		t.Fatalf("dataScope=3 dev: want [SKU-DEV1], got %v", gotDev)
	}

	// sales user: dataScope=3, dept=200 → 只看 dept 200 用户创建的 SPU 下的 SKU
	dpSales := &actions.DataPermission{DataScope: "3", UserId: 42, DeptId: 200, RoleId: 61}
	gotSales := f.getPage(t, dpSales)
	if len(gotSales) != 1 || gotSales[0] != "SKU-SALES1" {
		t.Fatalf("dataScope=3 sales: want [SKU-SALES1], got %v", gotSales)
	}
}

// TestSkuDataScope_RoleDept 验证 dataScope=2：按角色关联部门过滤。
func TestSkuDataScope_RoleDept(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	depts := []models.SysDept{
		{DeptId: 300, DeptName: "eng", DeptPath: "/300/", ParentId: 0},
		{DeptId: 400, DeptName: "ops", DeptPath: "/400/", ParentId: 0},
	}
	for i := range depts {
		if err := f.DB.Create(&depts[i]).Error; err != nil {
			t.Fatalf("seed dept: %v", err)
		}
	}
	users := []models.SysUser{
		{UserId: 51, NickName: "eng-user", DeptId: 300, RoleId: 70, Status: "2"},
		{UserId: 52, NickName: "ops-user", DeptId: 400, RoleId: 80, Status: "2"},
	}
	for i := range users {
		if err := f.DB.Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	// role 70 关联 dept 300
	if err := f.DB.Exec("INSERT INTO sys_role_dept (role_id, dept_id) VALUES (?, ?)", 70, 300).Error; err != nil {
		t.Fatalf("seed sys_role_dept: %v", err)
	}

	spuEng := f.seedSpu(t, "SPU-ENG", 51)
	spuOps := f.seedSpu(t, "SPU-OPS", 52)
	f.seedSku(t, "SKU-ENG1", spuEng)
	f.seedSku(t, "SKU-OPS1", spuOps)

	// eng role: dataScope=2, roleId=70 → 只看 role 70 关联部门(300)用户创建的 SPU 下的 SKU
	dpEng := &actions.DataPermission{DataScope: "2", UserId: 51, DeptId: 300, RoleId: 70}
	gotEng := f.getPage(t, dpEng)
	if len(gotEng) != 1 || gotEng[0] != "SKU-ENG1" {
		t.Fatalf("dataScope=2 eng role: want [SKU-ENG1], got %v", gotEng)
	}
}

// TestSkuDataScope_DeptAndChildren 验证 dataScope=4：按部门及其子部门过滤。
func TestSkuDataScope_DeptAndChildren(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	// 拓扑：parent(500) → child(501)
	depts := []models.SysDept{
		{DeptId: 500, DeptName: "parent", DeptPath: "/500/", ParentId: 0},
		{DeptId: 501, DeptName: "child", DeptPath: "/500/501/", ParentId: 500},
	}
	for i := range depts {
		if err := f.DB.Create(&depts[i]).Error; err != nil {
			t.Fatalf("seed dept: %v", err)
		}
	}
	users := []models.SysUser{
		{UserId: 61, NickName: "parent-user", DeptId: 500, RoleId: 90, Status: "2"},
		{UserId: 62, NickName: "child-user", DeptId: 501, RoleId: 91, Status: "2"},
		{UserId: 63, NickName: "other-user", DeptId: 400, RoleId: 92, Status: "2"},
	}
	for i := range users {
		if err := f.DB.Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	spuParent := f.seedSpu(t, "SPU-PARENT", 61)
	spuChild := f.seedSpu(t, "SPU-CHILD", 62)
	spuOther := f.seedSpu(t, "SPU-OTHER", 63)
	f.seedSku(t, "SKU-PARENT1", spuParent)
	f.seedSku(t, "SKU-CHILD1", spuChild)
	f.seedSku(t, "SKU-OTHER1", spuOther)

	// parent user: dataScope=4, dept=500 → 看 dept_path LIKE %/500/% 的用户创建的 SPU 下的 SKU
	dpParent := &actions.DataPermission{DataScope: "4", UserId: 61, DeptId: 500, RoleId: 90}
	gotParent := f.getPage(t, dpParent)
	if len(gotParent) != 2 {
		t.Fatalf("dataScope=4 parent: want 2 (parent+child), got %v", gotParent)
	}
	// SKU-OTHER1 不应出现
	for _, c := range gotParent {
		if c == "SKU-OTHER1" {
			t.Fatalf("dataScope=4 parent: SKU-OTHER1 should not appear, got %v", gotParent)
		}
	}
}

// TestSkuListItem_SpuFieldsPopulated 验证 GetPage 结果中 SpuCode/SpuName/SpuStatus 被正确填充。
func TestSkuListItem_SpuFieldsPopulated(t *testing.T) {
	f := newSkuDSFixture(t)
	defer f.Cleanup()

	spuId := f.seedSpu(t, "P-001", 1)
	f.seedSku(t, "S-001", spuId)

	dpAdmin := &actions.DataPermission{DataScope: "1"}
	list := make([]dto.SkuListItem, 0)
	var count int64
	req := &dto.SkuPageReq{}
	req.PageIndex = 1
	req.PageSize = 50
	if err := f.Sku.GetPage(req, dpAdmin, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	item := list[0]
	if item.SpuCode != "P-001" {
		t.Errorf("SpuCode: want P-001, got %q", item.SpuCode)
	}
	if item.SpuName != "P-001" {
		t.Errorf("SpuName: want P-001, got %q", item.SpuName)
	}
	if item.SpuStatus != models.SpuStatusApproved {
		t.Errorf("SpuStatus: want %d, got %d", models.SpuStatusApproved, item.SpuStatus)
	}
	if item.SkuCode != "S-001" {
		t.Errorf("SkuCode: want S-001, got %q", item.SkuCode)
	}
}
