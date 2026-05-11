package jobs

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	platformModels "go-admin/app/platform/models"
	"go-admin/common/utils"
	"go-admin/config"
)

// AnnouncementAttachmentGC 清理公告附件中的孤儿文件。
// 两阶段：
//   (i)  临时孤儿：business_id='0' 且创建超过 24h
//   (ii) 编辑替换孤儿：business_id 指向有效公告，但 storage_path 不在 content/cover 中
type AnnouncementAttachmentGC struct{}

func (AnnouncementAttachmentGC) Exec(arg interface{}) error {
	startTime := time.Now()
	logPrefix := "[AnnouncementAttachmentGC]"

	db := utils.GetDb()
	if db == nil {
		return fmt.Errorf("%s db not available", logPrefix)
	}

	dryRun := isDryRun()

	// ---- 阶段 (i)：临时孤儿 ----
	var tempOrphans []platformModels.AttachmentFile
	if err := db.Where("module_key = ?", "admin").
		Where("business_type IN ?", []string{"announcement-inline", "announcement-cover"}).
		Where("business_id = ?", "0").
		Where("created_at < ?", time.Now().Add(-24*time.Hour)).
		Find(&tempOrphans).Error; err != nil {
		fmt.Printf("%s [ERROR] stage=temp_orphans query failed: %v\n", logPrefix, err)
		return err
	}

	if dryRun {
		fmt.Printf("%s dry-run stage=temp_orphans matched=%d\n", logPrefix, len(tempOrphans))
	} else {
		if len(tempOrphans) > 0 {
			if err := deleteAttachmentsWithFiles(db, tempOrphans); err != nil {
				fmt.Printf("%s [ERROR] stage=temp_orphans delete failed: %v\n", logPrefix, err)
				return err
			}
		}
		fmt.Printf("%s stage=temp_orphans removed=%d\n", logPrefix, len(tempOrphans))
	}

	// ---- 阶段 (ii)：编辑替换孤儿 ----
	var announcements []models.Announcement
	if err := db.Find(&announcements).Error; err != nil {
		fmt.Printf("%s [ERROR] load announcements failed: %v\n", logPrefix, err)
		return err
	}

	totalReplaced := 0
	for _, ann := range announcements {
		// 收集该公告在 content + cover 中实际引用的 storage_path 集合
		activePaths := extractActivePaths(ann.Content, ann.CoverImageUrl)

		var candidates []platformModels.AttachmentFile
		if err := db.Where("module_key = ?", "admin").
			Where("business_type IN ?", []string{"announcement-inline", "announcement-cover"}).
			Where("business_id = ?", fmt.Sprintf("%d", ann.AnnouncementId)).
			Find(&candidates).Error; err != nil {
			fmt.Printf("%s [WARN] stage=replaced_orphans announcement_id=%d query failed: %v\n",
				logPrefix, ann.AnnouncementId, err)
			continue // 单条失败不阻塞整批
		}

		var toDelete []platformModels.AttachmentFile
		for _, c := range candidates {
			if !contains(activePaths, c.StoragePath) {
				toDelete = append(toDelete, c)
			}
		}

		if len(toDelete) == 0 {
			continue
		}

		if dryRun {
			fmt.Printf("%s dry-run stage=replaced_orphans announcement_id=%d matched=%d\n",
				logPrefix, ann.AnnouncementId, len(toDelete))
			totalReplaced += len(toDelete)
			continue
		}

		if err := deleteAttachmentsWithFiles(db, toDelete); err != nil {
			fmt.Printf("%s [WARN] stage=replaced_orphans announcement_id=%d delete failed: %v\n",
				logPrefix, ann.AnnouncementId, err)
			continue
		}
		totalReplaced += len(toDelete)
	}

	if !dryRun {
		fmt.Printf("%s stage=replaced_orphans removed=%d\n", logPrefix, totalReplaced)
	}

	latency := time.Since(startTime)
	if dryRun {
		fmt.Printf("%s summary total_matched=%d dry_run=true spend=%v\n",
			logPrefix, len(tempOrphans)+totalReplaced, latency)
	} else {
		fmt.Printf("%s summary total_removed=%d dry_run=false spend=%v\n",
			logPrefix, len(tempOrphans)+totalReplaced, latency)
	}

	return nil
}

// isDryRun 读取 dry-run 开关。
// 优先级：环境变量 > config/settings.yml extend.announcement.attachment_gc.dry_run > 默认 true
func isDryRun() bool {
	v := strings.TrimSpace(os.Getenv("ANNOUNCEMENT_GC_DRY_RUN"))
	if v != "" {
		return v != "false" && v != "0" && v != "no" && v != "off"
	}
	return config.ExtConfig.Announcement.AttachmentGC.DryRun
}

// deleteAttachmentsWithFiles 在事务中删除 DB 记录，然后尝试删除物理文件。
// 物理文件删除失败仅 warn，不返回错误（下一轮 cron 兜底）。
func deleteAttachmentsWithFiles(db *gorm.DB, items []platformModels.AttachmentFile) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int, 0, len(items))
	paths := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.AttachmentId)
		if it.StoragePath != "" {
			paths = append(paths, it.StoragePath)
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("attachment_id IN ?", ids).
			Delete(&platformModels.AttachmentFile{}).Error
	})
	if err != nil {
		return err
	}

	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			logger.NewHelper(logger.DefaultLogger).Warnf(
				"[AnnouncementAttachmentGC] remove file failed: %s, err=%v", p, err)
		}
	}
	return nil
}

// imgSrcRe 匹配 <img src="..."> 中的 src 值
var imgSrcRe = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

// extractActivePaths 从公告 content 和 cover_image_url 中提取所有被引用的 storage_path。
func extractActivePaths(content, coverURL string) []string {
	paths := make([]string, 0)
	if coverURL != "" {
		paths = append(paths, coverURL)
	}
	matches := imgSrcRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 && m[1] != "" {
			paths = append(paths, m[1])
		}
	}
	return paths
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
