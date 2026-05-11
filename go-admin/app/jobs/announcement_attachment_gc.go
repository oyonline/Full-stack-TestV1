package jobs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	platformModels "go-admin/app/platform/models"
	"go-admin/common/utils"
)

// AnnouncementAttachmentGC 公告附件垃圾回收 Job。
// 两阶段：
//   1) 清理 business_id='0' 且超过 24h 的临时孤儿附件；
//   2) 逐公告检查，删除 storage_path 已不在 content/cover_image_url 中的替换孤儿附件。
type AnnouncementAttachmentGC struct{}

func (AnnouncementAttachmentGC) Exec(arg interface{}) error {
	start := time.Now()
	db := utils.GetDb()
	if db == nil {
		return fmt.Errorf("[AnnouncementAttachmentGC] db not available")
	}

	dryRun := gcDryRunEnabled()
	logPrefix := "[AnnouncementAttachmentGC]"

	var totalMatched, totalRemoved int64

	// ---------- 阶段一：临时孤儿 ----------
	stage1Matched, stage1Removed, err := gcTempOrphans(db, dryRun, logPrefix)
	if err != nil {
		logger.Errorf("%s stage=temp_orphans error: %v", logPrefix, err)
	} else {
		totalMatched += stage1Matched
		totalRemoved += stage1Removed
	}

	// ---------- 阶段二：编辑替换孤儿 ----------
	stage2Matched, stage2Removed, err := gcReplacedOrphans(db, dryRun, logPrefix)
	if err != nil {
		logger.Errorf("%s stage=replaced_orphans error: %v", logPrefix, err)
	} else {
		totalMatched += stage2Matched
		totalRemoved += stage2Removed
	}

	spend := time.Since(start)
	if dryRun {
		logger.Infof("%s summary total_matched=%d dry_run=true spend=%v", logPrefix, totalMatched, spend)
	} else {
		logger.Infof("%s summary total_removed=%d dry_run=false spend=%v", logPrefix, totalRemoved, spend)
	}
	return nil
}

// gcTempOrphans 删除 business_id='0' 且 created_at < NOW()-24h 的附件。
func gcTempOrphans(db *gorm.DB, dryRun bool, logPrefix string) (matched, removed int64, err error) {
	cutoff := time.Now().Add(-24 * time.Hour)

	var rows []platformModels.AttachmentFile
	if err := db.Where("module_key = ?", service.AnnouncementModuleKey).
		Where("business_type IN ?", []string{service.AnnouncementBizInline, service.AnnouncementBizCover}).
		Where("business_id = ?", service.AnnouncementTempBusinessId).
		Where("created_at < ?", cutoff).
		Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	matched = int64(len(rows))
	if matched == 0 {
		return 0, 0, nil
	}

	if dryRun {
		logger.Infof("%s dry-run stage=temp_orphans matched=%d", logPrefix, matched)
		return matched, 0, nil
	}

	paths := make([]string, 0, len(rows))
	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		paths = append(paths, r.StoragePath)
		ids = append(ids, r.AttachmentId)
	}

	if err := db.Where("attachment_id IN ?", ids).
		Delete(&platformModels.AttachmentFile{}).Error; err != nil {
		return matched, 0, err
	}
	removed = int64(len(ids))
	logger.Infof("%s stage=temp_orphans removed=%d", logPrefix, removed)

	for _, p := range paths {
		if p == "" {
			continue
		}
		if rmErr := os.Remove(p); rmErr != nil && !os.IsNotExist(rmErr) {
			logger.Warnf("%s remove file warn: %s err=%v", logPrefix, p, rmErr)
		}
	}
	return matched, removed, nil
}

// gcReplacedOrphans 逐公告检查，删除已不在 content/cover_image_url 中的附件。
func gcReplacedOrphans(db *gorm.DB, dryRun bool, logPrefix string) (matched, removed int64, err error) {
	var announcements []models.Announcement
	if err := db.Select("announcement_id, content, cover_image_url").
		Find(&announcements).Error; err != nil {
		return 0, 0, err
	}

	for _, ann := range announcements {
		validPaths := make(map[string]struct{})
		if ann.CoverImageUrl != "" {
			if p := service.NormalizeStoragePath(ann.CoverImageUrl); p != "" {
				validPaths[p] = struct{}{}
			}
		}
		for _, p := range service.ExtractStoragePathsFromContent(ann.Content) {
			validPaths[p] = struct{}{}
		}

		var orphans []platformModels.AttachmentFile
		if err := db.Where("module_key = ?", service.AnnouncementModuleKey).
			Where("business_type IN ?", []string{service.AnnouncementBizInline, service.AnnouncementBizCover}).
			Where("business_id = ?", fmt.Sprintf("%d", ann.AnnouncementId)).
			Find(&orphans).Error; err != nil {
			logger.Warnf("%s stage=replaced_orphans announcement_id=%d query err=%v", logPrefix, ann.AnnouncementId, err)
			continue
		}

		var toDelete []platformModels.AttachmentFile
		for _, o := range orphans {
			if _, ok := validPaths[o.StoragePath]; !ok {
				toDelete = append(toDelete, o)
			}
		}

		if len(toDelete) == 0 {
			continue
		}

		matched += int64(len(toDelete))

		if dryRun {
			logger.Infof("%s dry-run stage=replaced_orphans announcement_id=%d matched=%d", logPrefix, ann.AnnouncementId, len(toDelete))
			continue
		}

		// 单公告单事务
		if err := db.Transaction(func(tx *gorm.DB) error {
			ids := make([]int, 0, len(toDelete))
			for _, o := range toDelete {
				ids = append(ids, o.AttachmentId)
			}
			if err := tx.Where("attachment_id IN ?", ids).
				Delete(&platformModels.AttachmentFile{}).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			logger.Warnf("%s stage=replaced_orphans announcement_id=%d tx err=%v", logPrefix, ann.AnnouncementId, err)
			continue
		}

		removed += int64(len(toDelete))
		logger.Infof("%s stage=replaced_orphans announcement_id=%d removed=%d", logPrefix, ann.AnnouncementId, len(toDelete))

		for _, o := range toDelete {
			if o.StoragePath == "" {
				continue
			}
			if rmErr := os.Remove(o.StoragePath); rmErr != nil && !os.IsNotExist(rmErr) {
				logger.Warnf("%s remove file warn: %s err=%v", logPrefix, o.StoragePath, rmErr)
			}
		}
	}

	return matched, removed, nil
}

// gcDryRunEnabled 默认 true；通过环境变量 GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN=false 关闭。
func gcDryRunEnabled() bool {
	v := strings.TrimSpace(os.Getenv("GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN"))
	if v == "" {
		return true
	}
	return strings.ToLower(v) != "false"
}
