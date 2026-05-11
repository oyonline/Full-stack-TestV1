package service

import (
	"sort"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
)

// seedAttachmentFixture 建立内存 SQLite + att_file 表,返回 *gorm.DB。
func seedAttachmentFixture(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&platformModels.AttachmentFile{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

// TestExtractStoragePathsFromContent_Basic 验证 HTML 中 <img src> 提取与过滤。
func TestExtractStoragePathsFromContent_Basic(t *testing.T) {
	cases := []struct {
		name string
		html string
		want []string
	}{
		{
			name: "empty",
			html: "",
			want: nil,
		},
		{
			name: "no img",
			html: "<p>hello</p>",
			want: nil,
		},
		{
			name: "single valid img",
			html: `<img src="/static/uploadfile/attachment/202601/abc.png" alt="x">`,
			want: []string{"static/uploadfile/attachment/202601/abc.png"},
		},
		{
			name: "multiple valid imgs",
			html: `<p><img src="/static/uploadfile/attachment/202601/a.png"></p>` +
				`<p><img src="static/uploadfile/attachment/202602/b.jpg"></p>`,
			want: []string{
				"static/uploadfile/attachment/202601/a.png",
				"static/uploadfile/attachment/202602/b.jpg",
			},
		},
		{
			name: "mixed valid and invalid src",
			html: `<img src="/static/uploadfile/attachment/202601/good.png">` +
				`<img src="https://cdn.example.com/bad.png">` +
				`<img src="/other/path/file.png">`,
			want: []string{"static/uploadfile/attachment/202601/good.png"},
		},
		{
			name: "single quotes",
			html: `<img src='static/uploadfile/attachment/202603/c.gif'>`,
			want: []string{"static/uploadfile/attachment/202603/c.gif"},
		},
		{
			name: "img without src",
			html: `<img alt="no src">`,
			want: nil,
		},
		{
			name: "duplicate src deduped by caller",
			html: `<img src="/static/uploadfile/attachment/202601/d.png">` +
				`<img src="/static/uploadfile/attachment/202601/d.png">`,
			want: []string{
				"static/uploadfile/attachment/202601/d.png",
				"static/uploadfile/attachment/202601/d.png",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractStoragePathsFromContent(tc.html)
			if len(got) != len(tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("idx %d want %q got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

// TestRebindAnnouncementAttachments_Basic 验证 rebind 把临时附件绑定到目标公告。
func TestRebindAnnouncementAttachments_Basic(t *testing.T) {
	db := seedAttachmentFixture(t)

	// 准备临时附件
	rows := []platformModels.AttachmentFile{
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/a.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/b.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizCover, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/c.png"},
		// 非临时行,不应被改
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: "99", StoragePath: "static/uploadfile/attachment/202601/d.png"},
		// 其他 module,不应被改
		{ModuleKey: "other", BusinessType: AnnouncementBizInline, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/e.png"},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	content := `<p><img src="/static/uploadfile/attachment/202601/a.png"></p>` +
		`<p><img src="/static/uploadfile/attachment/202601/b.png"></p>`
	coverUrl := "/static/uploadfile/attachment/202601/c.png"

	rebinded, err := rebindAnnouncementAttachments(db, 42, content, coverUrl)
	if err != nil {
		t.Fatalf("rebind error: %v", err)
	}
	if rebinded != 3 {
		t.Fatalf("want rebinded=3, got %d", rebinded)
	}

	// 验证 business_id 已更新
	var ids []int
	if err := db.Model(&platformModels.AttachmentFile{}).
		Where("business_id = ?", "42").
		Pluck("attachment_id", &ids).Error; err != nil {
		t.Fatalf("query rebinded: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("want 3 rebinded rows, got %d", len(ids))
	}

	// 验证未改行
	var untouched int64
	db.Model(&platformModels.AttachmentFile{}).Where("business_id = ?", "99").Count(&untouched)
	if untouched != 1 {
		t.Fatalf("want untouched business_id=99 count=1, got %d", untouched)
	}
	db.Model(&platformModels.AttachmentFile{}).Where("module_key = ?", "other").Where("business_id = ?", AnnouncementTempBusinessId).Count(&untouched)
	if untouched != 1 {
		t.Fatalf("want untouched other-module count=1, got %d", untouched)
	}
}

// TestRebindAnnouncementAttachments_NoMatch 验证无命中时返回 0。
func TestRebindAnnouncementAttachments_NoMatch(t *testing.T) {
	db := seedAttachmentFixture(t)
	rebinded, err := rebindAnnouncementAttachments(db, 42, "<p>no img</p>", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rebinded != 0 {
		t.Fatalf("want 0, got %d", rebinded)
	}
}

// TestRebindAnnouncementAttachments_CoverOnly 验证仅 coverUrl 命中时也能 rebind。
func TestRebindAnnouncementAttachments_CoverOnly(t *testing.T) {
	db := seedAttachmentFixture(t)
	row := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizCover,
		BusinessId:   AnnouncementTempBusinessId,
		StoragePath:  "static/uploadfile/attachment/202601/cover.png",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	rebinded, err := rebindAnnouncementAttachments(db, 7, "", "/static/uploadfile/attachment/202601/cover.png")
	if err != nil {
		t.Fatalf("rebind error: %v", err)
	}
	if rebinded != 1 {
		t.Fatalf("want 1, got %d", rebinded)
	}

	var bid string
	db.Model(&platformModels.AttachmentFile{}).Where("attachment_id = ?", row.AttachmentId).Pluck("business_id", &bid)
	if bid != "7" {
		t.Fatalf("want business_id=7, got %s", bid)
	}
}

// TestRemoveAnnouncementAttachments_Basic 验证删除并返回物理路径。
func TestRemoveAnnouncementAttachments_Basic(t *testing.T) {
	db := seedAttachmentFixture(t)

	rows := []platformModels.AttachmentFile{
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizMain, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/x.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/y.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizCover, BusinessId: "2", StoragePath: "static/uploadfile/attachment/202601/z.png"},
		// 其他公告,不应被删
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizMain, BusinessId: "3", StoragePath: "static/uploadfile/attachment/202601/w.png"},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	paths, err := removeAnnouncementAttachments(db, []int64{1, 2})
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}
	sort.Strings(paths)
	want := []string{
		"static/uploadfile/attachment/202601/x.png",
		"static/uploadfile/attachment/202601/y.png",
		"static/uploadfile/attachment/202601/z.png",
	}
	if len(paths) != len(want) {
		t.Fatalf("want paths %v, got %v", want, paths)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("idx %d want %q got %q", i, want[i], paths[i])
		}
	}

	var remain int64
	db.Model(&platformModels.AttachmentFile{}).Count(&remain)
	if remain != 1 {
		t.Fatalf("want remain=1, got %d", remain)
	}
}

// TestRemoveAnnouncementAttachments_EmptyIds 验证空 ID 列表安全返回。
func TestRemoveAnnouncementAttachments_EmptyIds(t *testing.T) {
	db := seedAttachmentFixture(t)
	paths, err := removeAnnouncementAttachments(db, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("want empty paths, got %v", paths)
	}
}

// TestNormalizeStoragePath 验证路径规范化。
func TestNormalizeStoragePath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/static/uploadfile/attachment/202601/a.png", "static/uploadfile/attachment/202601/a.png"},
		{"static/uploadfile/attachment/202601/a.png", "static/uploadfile/attachment/202601/a.png"},
		{"https://cdn.example.com/a.png", ""},
		{"/other/path/file.png", ""},
		{"", ""},
		{"  /static/uploadfile/attachment/202601/b.png  ", "static/uploadfile/attachment/202601/b.png"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := NormalizeStoragePath(tc.input)
			if got != tc.want {
				t.Fatalf("input %q want %q got %q", tc.input, tc.want, got)
			}
		})
	}
}

// TestRebindAnnouncementAttachments_Dedup 验证 content 中重复 path 只 rebind 一次。
func TestRebindAnnouncementAttachments_Dedup(t *testing.T) {
	db := seedAttachmentFixture(t)
	row := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizInline,
		BusinessId:   AnnouncementTempBusinessId,
		StoragePath:  "static/uploadfile/attachment/202601/dup.png",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	content := `<img src="/static/uploadfile/attachment/202601/dup.png">` +
		`<img src="/static/uploadfile/attachment/202601/dup.png">`
	rebinded, err := rebindAnnouncementAttachments(db, 10, content, "")
	if err != nil {
		t.Fatalf("rebind error: %v", err)
	}
	if rebinded != 1 {
		t.Fatalf("want 1, got %d", rebinded)
	}
}

// TestRebindAnnouncementAttachments_NonTempNotChanged 验证 business_id != '0' 的行不会被改。
func TestRebindAnnouncementAttachments_NonTempNotChanged(t *testing.T) {
	db := seedAttachmentFixture(t)
	row := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizInline,
		BusinessId:   "77",
		StoragePath:  "static/uploadfile/attachment/202601/existing.png",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	content := `<img src="/static/uploadfile/attachment/202601/existing.png">`
	rebinded, err := rebindAnnouncementAttachments(db, 20, content, "")
	if err != nil {
		t.Fatalf("rebind error: %v", err)
	}
	if rebinded != 0 {
		t.Fatalf("want 0, got %d", rebinded)
	}

	var bid string
	db.Model(&platformModels.AttachmentFile{}).Where("attachment_id = ?", row.AttachmentId).Pluck("business_id", &bid)
	if bid != "77" {
		t.Fatalf("want business_id=77, got %s", bid)
	}
}

// TestRemoveAnnouncementAttachments_OtherModuleNotDeleted 验证其他 module_key 不会被删。
func TestRemoveAnnouncementAttachments_OtherModuleNotDeleted(t *testing.T) {
	db := seedAttachmentFixture(t)
	rows := []platformModels.AttachmentFile{
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizMain, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/x.png"},
		{ModuleKey: "other", BusinessType: AnnouncementBizMain, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/y.png"},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	paths, err := removeAnnouncementAttachments(db, []int64{1})
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "static/uploadfile/attachment/202601/x.png" {
		t.Fatalf("want 1 path, got %v", paths)
	}

	var remain int64
	db.Model(&platformModels.AttachmentFile{}).Count(&remain)
	if remain != 1 {
		t.Fatalf("want remain=1, got %d", remain)
	}
}

// TestExtractStoragePathsFromContent_MixedQuotes 验证混合引号场景。
func TestExtractStoragePathsFromContent_MixedQuotes(t *testing.T) {
	html := `<img src="/static/uploadfile/attachment/202601/a.png">` +
		`<img src='/static/uploadfile/attachment/202601/b.png'>` +
		`<img src="https://example.com/c.png">`
	got := ExtractStoragePathsFromContent(html)
	want := []string{
		"static/uploadfile/attachment/202601/a.png",
		"static/uploadfile/attachment/202601/b.png",
	}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d want %q got %q", i, want[i], got[i])
		}
	}
}

// TestRebindAnnouncementAttachments_WithLogger 验证 rebind 在 tx 中可正常执行(不 panic)。
func TestRebindAnnouncementAttachments_WithLogger(t *testing.T) {
	db := seedAttachmentFixture(t)
	// 仅验证不 panic
	_, _ = rebindAnnouncementAttachments(db, 1, "", "")
}

// TestRemoveAnnouncementAttachments_ReturnsPathsForOsRemove 验证返回的路径可直接用于 os.Remove。
func TestRemoveAnnouncementAttachments_ReturnsPathsForOsRemove(t *testing.T) {
	db := seedAttachmentFixture(t)
	row := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizMain,
		BusinessId:   "5",
		StoragePath:  "static/uploadfile/attachment/202601/test.png",
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	paths, err := removeAnnouncementAttachments(db, []int64{5})
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "static/uploadfile/attachment/202601/test.png" {
		t.Fatalf("unexpected paths: %v", paths)
	}
}
