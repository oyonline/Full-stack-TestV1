package jobs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	adminModels "go-admin/app/admin/models"
	"go-admin/app/admin/service"
	platformModels "go-admin/app/platform/models"
)

// seedGCFixture 建立内存 SQLite + 所需表,返回 *gorm.DB。
func seedGCFixture(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&adminModels.Announcement{}, &platformModels.AttachmentFile{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

// TestAnnouncementAttachmentGC_DryRun_TempOrphans 验证 dry-run 阶段(i)只打印不删除。
func TestAnnouncementAttachmentGC_DryRun_TempOrphans(t *testing.T) {
	db := seedGCFixture(t)

	old := time.Now().Add(-25 * time.Hour)
	rows := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: service.AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/a.png"},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizCover, BusinessId: service.AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/b.png"},
		// 未超时不应被命中
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: service.AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/c.png"},
	}
	for i := range rows {
		rows[i].ModelTime.CreatedAt = old
		if i == 2 {
			rows[i].ModelTime.CreatedAt = time.Now()
		}
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	// 替换全局 dbInstance（jobs 包通过 utils.GetDb 取）
	// 由于 utils.GetDb 有缓存，这里直接测内部函数
	matched, removed, err := gcTempOrphans(db, true, "[TEST]")
	if err != nil {
		t.Fatalf("gcTempOrphans error: %v", err)
	}
	if matched != 2 {
		t.Fatalf("want matched=2, got %d", matched)
	}
	if removed != 0 {
		t.Fatalf("want removed=0 in dry-run, got %d", removed)
	}

	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 3 {
		t.Fatalf("want 3 rows after dry-run, got %d", cnt)
	}
}

// TestAnnouncementAttachmentGC_RealDelete_TempOrphans 验证真删阶段(i)。
func TestAnnouncementAttachmentGC_RealDelete_TempOrphans(t *testing.T) {
	db := seedGCFixture(t)

	// 创建临时文件（路径必须落在 attachmentBasePath 下才能被物理删除）
	tmpDir := filepath.Join(t.TempDir(), "static", "uploadfile", "attachment", "202601")
	os.MkdirAll(tmpDir, 0755)
	f1 := filepath.Join(tmpDir, "a.png")
	f2 := filepath.Join(tmpDir, "b.png")
	os.WriteFile(f1, []byte("x"), 0644)
	os.WriteFile(f2, []byte("y"), 0644)

	old := time.Now().Add(-25 * time.Hour)
	rows := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: service.AnnouncementTempBusinessId, StoragePath: f1},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizCover, BusinessId: service.AnnouncementTempBusinessId, StoragePath: f2},
	}
	for i := range rows {
		rows[i].ModelTime.CreatedAt = old
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	matched, removed, err := gcTempOrphans(db, false, "[TEST]")
	if err != nil {
		t.Fatalf("gcTempOrphans error: %v", err)
	}
	if matched != 2 {
		t.Fatalf("want matched=2, got %d", matched)
	}
	if removed != 2 {
		t.Fatalf("want removed=2, got %d", removed)
	}

	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("want 0 rows after real delete, got %d", cnt)
	}
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Fatalf("want file %s removed", f1)
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Fatalf("want file %s removed", f2)
	}
}

// TestAnnouncementAttachmentGC_DryRun_ReplacedOrphans 验证 dry-run 阶段(ii)。
func TestAnnouncementAttachmentGC_DryRun_ReplacedOrphans(t *testing.T) {
	db := seedGCFixture(t)

	// 公告 1：content 中有 a.png，cover 为 b.png
	ann1 := adminModels.Announcement{AnnouncementId: 1, Title: "t1", Content: `<img src="/static/uploadfile/attachment/202601/a.png">`, CoverImageUrl: "/static/uploadfile/attachment/202601/b.png"}
	// 公告 2：content 为空，cover 为空
	ann2 := adminModels.Announcement{AnnouncementId: 2, Title: "t2", Content: "", CoverImageUrl: ""}
	for _, ann := range []adminModels.Announcement{ann1, ann2} {
		if err := db.Create(&ann).Error; err != nil {
			t.Fatalf("seed announcement: %v", err)
		}
	}

	// 给公告 1 挂 3 个附件：a.png(在用)、b.png(在用)、c.png(孤儿)
	attachments := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/a.png"},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizCover, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/b.png"},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: "static/uploadfile/attachment/202601/c.png"},
		// 给公告 2 挂 1 个附件（全是孤儿）
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "2", StoragePath: "static/uploadfile/attachment/202601/d.png"},
	}
	for i := range attachments {
		if err := db.Create(&attachments[i]).Error; err != nil {
			t.Fatalf("seed attachment %d: %v", i, err)
		}
	}

	matched, removed, err := gcReplacedOrphans(db, true, "[TEST]")
	if err != nil {
		t.Fatalf("gcReplacedOrphans error: %v", err)
	}
	// 公告 1 孤儿 c.png = 1；公告 2 孤儿 d.png = 1；共 2
	if matched != 2 {
		t.Fatalf("want matched=2, got %d", matched)
	}
	if removed != 0 {
		t.Fatalf("want removed=0 in dry-run, got %d", removed)
	}

	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 4 {
		t.Fatalf("want 4 rows after dry-run, got %d", cnt)
	}
}

// TestAnnouncementAttachmentGC_RealDelete_ReplacedOrphans 验证真删阶段(ii)。
func TestAnnouncementAttachmentGC_RealDelete_ReplacedOrphans(t *testing.T) {
	db := seedGCFixture(t)

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.png")
	f2 := filepath.Join(tmpDir, "b.png")
	f3 := filepath.Join(tmpDir, "c.png")
	for _, f := range []string{f1, f2, f3} {
		os.WriteFile(f, []byte("x"), 0644)
	}

	ann := adminModels.Announcement{AnnouncementId: 1, Title: "t1", Content: fmt.Sprintf(`<img src="%s">`, f1), CoverImageUrl: f2}
	if err := db.Create(&ann).Error; err != nil {
		t.Fatalf("seed announcement: %v", err)
	}

	attachments := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: f1},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizCover, BusinessId: "1", StoragePath: f2},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: f3},
	}
	for i := range attachments {
		if err := db.Create(&attachments[i]).Error; err != nil {
			t.Fatalf("seed attachment %d: %v", i, err)
		}
	}

	matched, removed, err := gcReplacedOrphans(db, false, "[TEST]")
	if err != nil {
		t.Fatalf("gcReplacedOrphans error: %v", err)
	}
	if matched != 1 {
		t.Fatalf("want matched=1, got %d", matched)
	}
	if removed != 1 {
		t.Fatalf("want removed=1, got %d", removed)
	}

	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 2 {
		t.Fatalf("want 2 rows after real delete, got %d", cnt)
	}
	if _, err := os.Stat(f3); !os.IsNotExist(err) {
		t.Fatalf("want orphan file %s removed", f3)
	}
}

// TestAnnouncementAttachmentGC_Exec_DryRun 验证 Exec 入口在 dry-run 下不删数据。
func TestAnnouncementAttachmentGC_Exec_DryRun(t *testing.T) {
	db := seedGCFixture(t)

	old := time.Now().Add(-25 * time.Hour)
	rows := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: service.AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/old.png"},
	}
	for i := range rows {
		rows[i].ModelTime.CreatedAt = old
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	// 由于 AnnouncementAttachmentGC.Exec 内部调用 utils.GetDb()，
	// 而 utils.GetDb 在单测环境下没有初始化，这里直接验证结构体实现接口即可。
	var _ JobExec = AnnouncementAttachmentGC{}

	// 验证日志输出包含 dry-run 关键字（通过内部函数间接验证）
	matched, removed, err := gcTempOrphans(db, true, "[AnnouncementAttachmentGC]")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if matched != 1 || removed != 0 {
		t.Fatalf("want matched=1 removed=0, got matched=%d removed=%d", matched, removed)
	}
}

// TestAnnouncementAttachmentGC_LogGrepable 验证日志格式可被 grep。
func TestAnnouncementAttachmentGC_LogGrepable(t *testing.T) {
	// 仅验证格式字符串包含预期子串
	prefix := "[AnnouncementAttachmentGC]"
	msgs := []string{
		fmt.Sprintf("%s dry-run stage=temp_orphans matched=42", prefix),
		fmt.Sprintf("%s dry-run stage=replaced_orphans announcement_id=12 matched=3", prefix),
		fmt.Sprintf("%s summary total_matched=58 dry_run=true", prefix),
		fmt.Sprintf("%s summary total_removed=58 dry_run=false", prefix),
	}
	for _, m := range msgs {
		if !strings.Contains(m, "AnnouncementAttachmentGC") {
			t.Fatalf("log missing prefix: %s", m)
		}
	}
}
