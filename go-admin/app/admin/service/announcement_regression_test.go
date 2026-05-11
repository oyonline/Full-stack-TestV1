package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	platformModels "go-admin/app/platform/models"
)

func setupRegressionDB(t *testing.T) (*Announcement, func()) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Announcement{},
		&models.AnnouncementScope{},
		&models.AnnouncementReadLog{},
		&platformModels.AttachmentFile{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	s := &Announcement{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s, func() {}
}

// ------------------------------------------------------------------
// 用例 1: rebind 主路径（新建公告 + 插图 + 封面 → 保存后 business_id 指向公告）
// ------------------------------------------------------------------
func TestRegression_RebindMainPath(t *testing.T) {
	s, cleanup := setupRegressionDB(t)
	defer cleanup()

	rows := []platformModels.AttachmentFile{
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/a.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/b.png"},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizCover, BusinessId: AnnouncementTempBusinessId, StoragePath: "static/uploadfile/attachment/202601/c.png"},
	}
	for i := range rows {
		if err := s.Orm.Create(&rows[i]).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	req := &dto.AnnouncementInsertReq{
		Title:         "regression-rebind-main",
		Content:       `<p><img src="/static/uploadfile/attachment/202601/a.png"></p><p><img src="/static/uploadfile/attachment/202601/b.png"></p>`,
		CoverImageUrl: "/static/uploadfile/attachment/202601/c.png",
		Status:        models.AnnouncementStatusPublished,
	}
	req.SetCreateBy(1)

	id, err := s.Insert(req)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id == 0 {
		t.Fatalf("insert returned 0")
	}

	// 断言: 刚上传的 inline/cover 行 business_id = 新公告 ID
	var bound int64
	if err := s.Orm.Model(&platformModels.AttachmentFile{}).
		Where("business_id = ?", fmt.Sprintf("%d", id)).
		Count(&bound).Error; err != nil {
		t.Fatalf("count bound: %v", err)
	}
	if bound != 3 {
		t.Fatalf("want 3 bound rows, got %d", bound)
	}

	// 断言: 没有 business_id='0' 的相关行残留
	var orphan int64
	if err := s.Orm.Model(&platformModels.AttachmentFile{}).
		Where("module_key = ?", AnnouncementModuleKey).
		Where("business_type IN ?", []string{AnnouncementBizInline, AnnouncementBizCover}).
		Where("business_id = ?", AnnouncementTempBusinessId).
		Count(&orphan).Error; err != nil {
		t.Fatalf("count orphan: %v", err)
	}
	if orphan != 0 {
		t.Fatalf("want 0 orphan rows, got %d", orphan)
	}
}

// ------------------------------------------------------------------
// 用例 2: rebind 不抢错图（编辑公告换图 → 旧图仍指 A,新图也指 A,无 business_id='0'）
// ------------------------------------------------------------------
func TestRegression_RebindDoesNotStealOldImage(t *testing.T) {
	s, cleanup := setupRegressionDB(t)
	defer cleanup()

	// 公告 A 已存在,带旧图
	ann := models.Announcement{Title: "A", Content: `<img src="/static/uploadfile/attachment/202601/old.png">`, Status: models.AnnouncementStatusPublished}
	ann.SetCreateBy(1)
	if err := s.Orm.Create(&ann).Error; err != nil {
		t.Fatalf("seed ann: %v", err)
	}

	// 旧图已绑定到 A
	oldAtt := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizInline,
		BusinessId:   fmt.Sprintf("%d", ann.AnnouncementId),
		StoragePath:  "static/uploadfile/attachment/202601/old.png",
	}
	if err := s.Orm.Create(&oldAtt).Error; err != nil {
		t.Fatalf("seed old att: %v", err)
	}

	// 新图为临时附件
	newAtt := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizCover,
		BusinessId:   AnnouncementTempBusinessId,
		StoragePath:  "static/uploadfile/attachment/202601/newcover.png",
	}
	if err := s.Orm.Create(&newAtt).Error; err != nil {
		t.Fatalf("seed new att: %v", err)
	}

	req := &dto.AnnouncementUpdateReq{
		AnnouncementId: ann.AnnouncementId,
		Title:          "A-updated",
		Content:        `<p>new content</p>`,
		CoverImageUrl:  "/static/uploadfile/attachment/202601/newcover.png",
		Status:         models.AnnouncementStatusPublished,
	}
	req.SetUpdateBy(1)
	if err := s.Update(req, nil); err != nil {
		t.Fatalf("update: %v", err)
	}

	// 旧图 business_id 仍指 A
	var oldBid string
	s.Orm.Model(&platformModels.AttachmentFile{}).Where("attachment_id = ?", oldAtt.AttachmentId).Pluck("business_id", &oldBid)
	if oldBid != fmt.Sprintf("%d", ann.AnnouncementId) {
		t.Fatalf("old image business_id changed: want %d, got %s", ann.AnnouncementId, oldBid)
	}

	// 新图 business_id 也指 A
	var newBid string
	s.Orm.Model(&platformModels.AttachmentFile{}).Where("attachment_id = ?", newAtt.AttachmentId).Pluck("business_id", &newBid)
	if newBid != fmt.Sprintf("%d", ann.AnnouncementId) {
		t.Fatalf("new image business_id wrong: want %d, got %s", ann.AnnouncementId, newBid)
	}

	// 不存在 business_id='0' 的本次相关行
	var orphan int64
	s.Orm.Model(&platformModels.AttachmentFile{}).
		Where("module_key = ?", AnnouncementModuleKey).
		Where("business_type IN ?", []string{AnnouncementBizInline, AnnouncementBizCover}).
		Where("business_id = ?", AnnouncementTempBusinessId).
		Count(&orphan)
	if orphan != 0 {
		t.Fatalf("want 0 orphan rows, got %d", orphan)
	}
}

// ------------------------------------------------------------------
// 用例 3: cascade 删公告（删除后 att_file / announcement / scope / read_log 全消失,文件也被删）
// ------------------------------------------------------------------
func TestRegression_CascadeDeleteAnnouncement(t *testing.T) {
	s, cleanup := setupRegressionDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.png")
	f2 := filepath.Join(tmpDir, "b.png")
	os.WriteFile(f1, []byte("a"), 0644)
	os.WriteFile(f2, []byte("b"), 0644)

	ann := models.Announcement{Title: "del", Content: "<p>x</p>", Status: models.AnnouncementStatusPublished}
	ann.SetCreateBy(1)
	if err := s.Orm.Create(&ann).Error; err != nil {
		t.Fatalf("seed ann: %v", err)
	}

	// scope
	scope := models.AnnouncementScope{AnnouncementId: ann.AnnouncementId, DeptId: 1}
	if err := s.Orm.Create(&scope).Error; err != nil {
		t.Fatalf("seed scope: %v", err)
	}

	// read_log
	readLog := models.AnnouncementReadLog{UserId: 1, AnnouncementId: ann.AnnouncementId, ReadAt: time.Now()}
	if err := s.Orm.Create(&readLog).Error; err != nil {
		t.Fatalf("seed read_log: %v", err)
	}

	atts := []platformModels.AttachmentFile{
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizInline, BusinessId: fmt.Sprintf("%d", ann.AnnouncementId), StoragePath: f1},
		{ModuleKey: AnnouncementModuleKey, BusinessType: AnnouncementBizCover, BusinessId: fmt.Sprintf("%d", ann.AnnouncementId), StoragePath: f2},
	}
	for i := range atts {
		if err := s.Orm.Create(&atts[i]).Error; err != nil {
			t.Fatalf("seed att %d: %v", i, err)
		}
	}

	req := &dto.AnnouncementDeleteReq{Ids: []int64{ann.AnnouncementId}}
	if err := s.Remove(req, nil); err != nil {
		t.Fatalf("remove: %v", err)
	}

	// DB: att_file 中该公告的全部消失
	var attCnt int64
	s.Orm.Model(&platformModels.AttachmentFile{}).Count(&attCnt)
	if attCnt != 0 {
		t.Fatalf("want 0 att_file rows, got %d", attCnt)
	}

	// 文件系统:物理文件不存在
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Fatalf("f1 should be deleted")
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Fatalf("f2 should be deleted")
	}

	// 公告本身消失
	var annCnt int64
	s.Orm.Model(&models.Announcement{}).Count(&annCnt)
	if annCnt != 0 {
		t.Fatalf("want 0 announcement rows, got %d", annCnt)
	}

	// scope 消失
	var scopeCnt int64
	s.Orm.Model(&models.AnnouncementScope{}).Count(&scopeCnt)
	if scopeCnt != 0 {
		t.Fatalf("want 0 scope rows, got %d", scopeCnt)
	}

	// read_log 消失
	var readCnt int64
	s.Orm.Model(&models.AnnouncementReadLog{}).Count(&readCnt)
	if readCnt != 0 {
		t.Fatalf("want 0 read_log rows, got %d", readCnt)
	}
}

// ------------------------------------------------------------------
// 用例 7: rebind 不抢别公告的图（跨公告复用不被支持）
// ------------------------------------------------------------------
func TestRegression_RebindDoesNotStealFromOtherAnnouncement(t *testing.T) {
	s, cleanup := setupRegressionDB(t)
	defer cleanup()

	// 公告 A 已存在,带图 1
	annA := models.Announcement{Title: "A", Content: `<img src="/static/uploadfile/attachment/202601/shared.png">`, Status: models.AnnouncementStatusPublished}
	annA.SetCreateBy(1)
	if err := s.Orm.Create(&annA).Error; err != nil {
		t.Fatalf("seed annA: %v", err)
	}

	attA := platformModels.AttachmentFile{
		ModuleKey:    AnnouncementModuleKey,
		BusinessType: AnnouncementBizInline,
		BusinessId:   fmt.Sprintf("%d", annA.AnnouncementId),
		StoragePath:  "static/uploadfile/attachment/202601/shared.png",
	}
	if err := s.Orm.Create(&attA).Error; err != nil {
		t.Fatalf("seed attA: %v", err)
	}

	// 公告 B 的富文本里粘贴了 A 的 img src（但 B 没有上传该图为临时附件）
	req := &dto.AnnouncementInsertReq{
		Title:   "B",
		Content: `<p><img src="/static/uploadfile/attachment/202601/shared.png"></p>`,
		Status:  models.AnnouncementStatusPublished,
	}
	req.SetCreateBy(1)
	idB, err := s.Insert(req)
	if err != nil {
		t.Fatalf("insert B: %v", err)
	}

	// 图 1 的 business_id 仍指 A,没有被 rebind 抢到 B
	var bid string
	s.Orm.Model(&platformModels.AttachmentFile{}).Where("attachment_id = ?", attA.AttachmentId).Pluck("business_id", &bid)
	if bid != fmt.Sprintf("%d", annA.AnnouncementId) {
		t.Fatalf("image stolen: want business_id=%d, got %s", annA.AnnouncementId, bid)
	}

	// 公告 B 没有产生新的 attachment 行（因为没有临时附件）
	var bAttCnt int64
	s.Orm.Model(&platformModels.AttachmentFile{}).Where("business_id = ?", fmt.Sprintf("%d", idB)).Count(&bAttCnt)
	if bAttCnt != 0 {
		t.Fatalf("B should have 0 attachments, got %d", bAttCnt)
	}
}
