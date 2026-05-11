package jobs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	adminModels "go-admin/app/admin/models"
	"go-admin/app/admin/service"
	platformModels "go-admin/app/platform/models"
)

func setupGCRegressionDB(t *testing.T) *gorm.DB {
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

// ------------------------------------------------------------------
// 用例 4: GC 临时孤儿（business_id='0' + 25h 前 → 真删后 DB 行和物理文件都消失）
// ------------------------------------------------------------------
func TestRegression_GC_TempOrphan_RealDelete(t *testing.T) {
	db := setupGCRegressionDB(t)

	tmpDir := filepath.Join(t.TempDir(), "static", "uploadfile", "attachment", "202601")
	os.MkdirAll(tmpDir, 0755)
	f1 := filepath.Join(tmpDir, "old.png")
	os.WriteFile(f1, []byte("x"), 0644)

	old := time.Now().Add(-25 * time.Hour)
	row := platformModels.AttachmentFile{
		ModuleKey:    service.AnnouncementModuleKey,
		BusinessType: service.AnnouncementBizInline,
		BusinessId:   service.AnnouncementTempBusinessId,
		StoragePath:  f1,
	}
	row.ModelTime.CreatedAt = old
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	matched, removed, err := gcTempOrphans(db, false, "[AnnouncementAttachmentGC]")
	if err != nil {
		t.Fatalf("gcTempOrphans: %v", err)
	}
	if matched != 1 {
		t.Fatalf("want matched=1, got %d", matched)
	}
	if removed != 1 {
		t.Fatalf("want removed=1, got %d", removed)
	}

	// DB 行消失
	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("want 0 rows, got %d", cnt)
	}

	// 物理文件消失
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Fatalf("file should be removed")
	}
}

// ------------------------------------------------------------------
// 用例 5: GC 编辑替换孤儿（公告换图后旧图被 GC,新图不动）
// ------------------------------------------------------------------
func TestRegression_GC_ReplacedOrphan(t *testing.T) {
	db := setupGCRegressionDB(t)

	spOld := "static/uploadfile/attachment/202601/old.png"
	spNew := "static/uploadfile/attachment/202601/new.png"

	ann := adminModels.Announcement{
		AnnouncementId: 1,
		Title:          "edit",
		Content:        fmt.Sprintf(`<img src="/%s">`, spNew),
		CoverImageUrl:  "",
	}
	if err := db.Create(&ann).Error; err != nil {
		t.Fatalf("seed ann: %v", err)
	}

	atts := []platformModels.AttachmentFile{
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: spOld},
		{ModuleKey: service.AnnouncementModuleKey, BusinessType: service.AnnouncementBizInline, BusinessId: "1", StoragePath: spNew},
	}
	for i := range atts {
		if err := db.Create(&atts[i]).Error; err != nil {
			t.Fatalf("seed att %d: %v", i, err)
		}
	}

	matched, removed, err := gcReplacedOrphans(db, false, "[AnnouncementAttachmentGC]")
	if err != nil {
		t.Fatalf("gcReplacedOrphans: %v", err)
	}
	if matched != 1 {
		t.Fatalf("want matched=1, got %d", matched)
	}
	if removed != 1 {
		t.Fatalf("want removed=1, got %d", removed)
	}

	// 旧图行消失
	var oldCnt int64
	db.Model(&platformModels.AttachmentFile{}).Where("storage_path = ?", spOld).Count(&oldCnt)
	if oldCnt != 0 {
		t.Fatalf("old image should be removed, got %d rows", oldCnt)
	}

	// 新图行不动
	var newCnt int64
	db.Model(&platformModels.AttachmentFile{}).Where("storage_path = ?", spNew).Count(&newCnt)
	if newCnt != 1 {
		t.Fatalf("new image should remain, got %d rows", newCnt)
	}
}

// ------------------------------------------------------------------
// 用例 6: GC dry-run 不动数据（dry_run=true 重复用例 4 场景）
// ------------------------------------------------------------------
func TestRegression_GC_DryRun_NoDelete(t *testing.T) {
	db := setupGCRegressionDB(t)

	old := time.Now().Add(-25 * time.Hour)
	row := platformModels.AttachmentFile{
		ModuleKey:    service.AnnouncementModuleKey,
		BusinessType: service.AnnouncementBizInline,
		BusinessId:   service.AnnouncementTempBusinessId,
		StoragePath:  "static/uploadfile/attachment/202601/old.png",
	}
	row.ModelTime.CreatedAt = old
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	matched, removed, err := gcTempOrphans(db, true, "[AnnouncementAttachmentGC]")
	if err != nil {
		t.Fatalf("gcTempOrphans: %v", err)
	}
	if matched != 1 {
		t.Fatalf("want matched=1, got %d", matched)
	}
	if removed != 0 {
		t.Fatalf("want removed=0 in dry-run, got %d", removed)
	}

	// DB 行依然在
	var cnt int64
	db.Model(&platformModels.AttachmentFile{}).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("want 1 row after dry-run, got %d", cnt)
	}
}
