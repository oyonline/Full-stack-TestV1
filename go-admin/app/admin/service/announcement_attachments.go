package service

import (
	"regexp"
	"strings"

	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
)

const (
	AnnouncementModuleKey      = "admin"
	AnnouncementBizMain        = "announcement"
	AnnouncementBizInline      = "announcement-inline"
	AnnouncementBizCover       = "announcement-cover"
	AnnouncementTempBusinessId = "0"
	attachmentBasePath         = "static/uploadfile/attachment"
)

// imgSrcRegex 匹配 <img src="..."> 或 <img src='...'> 中的 src 值。
var imgSrcRegex = regexp.MustCompile(`<img[^>]+src\s*=\s*["']([^"']+)["']`)

// AnnouncementAttachmentGCConfig 公告附件 GC 配置项。
type AnnouncementAttachmentGCConfig struct {
	DryRun bool `yaml:"dry_run"`
}

// GetAnnouncementAttachmentGCConfig 读取配置；若未配置则默认 dry_run=true。
func GetAnnouncementAttachmentGCConfig() *AnnouncementAttachmentGCConfig {
	// 通过 go-admin-core/sdk/config 的 Extend 映射读取
	// 由于 extend 是 map[string]interface{}，直接返回默认 true 作为兜底。
	// 若项目后续在 config/extend.go 中强类型化，可再改这里。
	return &AnnouncementAttachmentGCConfig{DryRun: true}
}

// ExtractStoragePathsFromContent 解析 HTML 里的 <img src>,
// 返回 storage_path 形态(去掉前导斜杠、仅保留以 'static/uploadfile/attachment/' 开头的项)。
func ExtractStoragePathsFromContent(html string) []string {
	matches := imgSrcRegex.FindAllStringSubmatch(html, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		path := NormalizeStoragePath(m[1])
		if path != "" {
			result = append(result, path)
		}
	}
	return result
}

// NormalizeStoragePath 把 URL/路径统一为 storage_path 形态:
//   - 去掉前导斜杠
//   - 仅保留以 "static/uploadfile/attachment/" 开头的项
func NormalizeStoragePath(raw string) string {
	path := strings.TrimSpace(raw)
	path = strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(path, attachmentBasePath) {
		return ""
	}
	return path
}

// rebindAnnouncementAttachments 把 content 的 <img src> 与 coverUrl 命中的
// 临时附件(business_id='0', business_type IN inline/cover, module='admin')
// 改写为目标 announcementId。
//
// 安全约束:
//   - 仅 rebind 命中 announcement-inline / announcement-cover 的临时行;
//   - 不动 business_id != '0' 的行;
//   - rebind 失败不回滚整个 tx,best-effort + 记 warn。
func rebindAnnouncementAttachments(
	tx *gorm.DB,
	announcementId int64,
	content string,
	coverUrl string,
) (rebinded int, err error) {
	paths := ExtractStoragePathsFromContent(content)
	if coverUrl != "" {
		coverPath := NormalizeStoragePath(coverUrl)
		if coverPath != "" {
			paths = append(paths, coverPath)
		}
	}

	if len(paths) == 0 {
		return 0, nil
	}

	// 去重
	seen := make(map[string]struct{}, len(paths))
	uniq := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		uniq = append(uniq, p)
	}

	// 先查询命中临时附件的 ID 列表,避免 UPDATE 误改非临时行
	var ids []int
	if err := tx.Model(&platformModels.AttachmentFile{}).
		Where("module_key = ?", AnnouncementModuleKey).
		Where("business_type IN ?", []string{AnnouncementBizInline, AnnouncementBizCover}).
		Where("business_id = ?", AnnouncementTempBusinessId).
		Where("storage_path IN ?", uniq).
		Pluck("attachment_id", &ids).Error; err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, nil
	}

	// 批量更新 business_id 为目标 announcementId
	result := tx.Model(&platformModels.AttachmentFile{}).
		Where("attachment_id IN ?", ids).
		Update("business_id", announcementId)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

// removeAnnouncementAttachments 在 Remove tx 内删 announcementIds 关联的所有
// announcement / announcement-inline / announcement-cover 行,
// 返回需要 commit 后 os.Remove 的物理路径列表。
func removeAnnouncementAttachments(
	tx *gorm.DB,
	announcementIds []int64,
) (storagePaths []string, err error) {
	if len(announcementIds) == 0 {
		return nil, nil
	}

	// 收集物理路径
	var paths []string
	if err := tx.Model(&platformModels.AttachmentFile{}).
		Where("module_key = ?", AnnouncementModuleKey).
		Where("business_type IN ?", []string{AnnouncementBizMain, AnnouncementBizInline, AnnouncementBizCover}).
		Where("business_id IN ?", announcementIds).
		Pluck("storage_path", &paths).Error; err != nil {
		return nil, err
	}

	// 删除记录
	if err := tx.Where("module_key = ?", AnnouncementModuleKey).
		Where("business_type IN ?", []string{AnnouncementBizMain, AnnouncementBizInline, AnnouncementBizCover}).
		Where("business_id IN ?", announcementIds).
		Delete(&platformModels.AttachmentFile{}).Error; err != nil {
		return nil, err
	}

	return paths, nil
}
