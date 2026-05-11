package service

import (
	"errors"
	"os"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

// Announcement 公告服务，承担非平凡逻辑：scope 维护、mark-read 幂等、派生字段计算。
type Announcement struct {
	service.Service
}

// htmlPolicy 单例：UGCPolicy 允许常见富文本元素（h*/p/em/strong/ul/ol/li/blockquote/img 等），
// 同时阻断 script/iframe/style/on* 事件等危险节点。允许相对路径与 http(s) URL。
var htmlPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// 富文本编辑器常见的 style 与 align 属性需放行
	p.AllowAttrs("style").Globally()
	p.AllowAttrs("align").OnElements("p", "div", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowAttrs("class").Globally()
	// 图片需要支持嵌入附件平台返回的 URL
	p.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	return p
}()

// SanitizeContent 对富文本内容做 XSS 过滤。
func SanitizeContent(html string) string {
	if html == "" {
		return ""
	}
	return htmlPolicy.Sanitize(html)
}

// GetPage 列表查询。支持按部门可见性过滤、按生效时间过滤、补 is_read 与 read_count。
//
// 数据权限说明：
//   - data_scope=1（全部）：admin/默认角色看全部公告（与本接口接入前行为一致）。
//   - data_scope=3/4（本部门 / 本部门及以下）：仅看 create_by 在对应部门的公告。
//   - data_scope=5（仅本人）：仅看自己 create 的公告。
//   - OnlyVisible（按部门可见性，由 announcement_scope 表驱动）与 dataScope 正交，
//     都满足时一条记录才会出现在结果中。
//
// p 由 router 上的 actions.PermissionAction 中间件注入，apis 层用
// actions.GetPermissionFromContext(c) 取出后传入；为 nil 时按"全部"放行。
func (e *Announcement) GetPage(c *dto.AnnouncementPageReq, p *actions.DataPermission, list *[]dto.AnnouncementListItem, count *int64, currentUserId int) error {
	var data models.Announcement
	q := e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		)

	if c.OnlyValid == 1 {
		now := time.Now()
		q = q.Where("publish_at IS NULL OR publish_at <= ?", now).
			Where("expire_at IS NULL OR expire_at > ?", now).
			Where("status = ?", models.AnnouncementStatusPublished)
	}

	if c.OnlyVisible == 1 && currentUserId > 0 {
		// 当前用户的部门集合
		var deptIds []int
		if err := e.Orm.Table("sys_user_depts").
			Where("user_id = ?", currentUserId).
			Pluck("dept_id", &deptIds).Error; err != nil {
			e.Log.Errorf("load user depts: %s", err)
			return err
		}
		if len(deptIds) == 0 {
			*count = 0
			return nil
		}
		q = q.Where("announcement_id IN (?)",
			e.Orm.Table("announcement_scope").
				Select("announcement_id").
				Where("dept_id IN ?", deptIds))
	}

	// 先 Count，再分页 Find
	if err := q.Count(count).Error; err != nil {
		e.Log.Errorf("count announcement: %s", err)
		return err
	}

	pageSize := c.GetPageSize()
	pageIndex := c.GetPageIndex()
	offset := (pageIndex - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	rows := make([]models.Announcement, 0)
	if err := q.
		Order("is_top DESC, top_sort DESC, COALESCE(publish_at, created_at) DESC, announcement_id DESC").
		Limit(pageSize).Offset(offset).
		Find(&rows).Error; err != nil {
		e.Log.Errorf("find announcement: %s", err)
		return err
	}

	if len(rows) == 0 {
		*list = []dto.AnnouncementListItem{}
		return nil
	}

	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.AnnouncementId)
	}

	// 一次性拉 scope
	type scopeRow struct {
		AnnouncementId int64
		DeptId         int
	}
	scopeRows := make([]scopeRow, 0)
	if err := e.Orm.Table("announcement_scope").
		Where("announcement_id IN ?", ids).
		Find(&scopeRows).Error; err != nil {
		e.Log.Errorf("load scope: %s", err)
		return err
	}
	scopeMap := make(map[int64][]int, len(rows))
	for _, sr := range scopeRows {
		scopeMap[sr.AnnouncementId] = append(scopeMap[sr.AnnouncementId], sr.DeptId)
	}

	// 一次性拉 read_count
	type rcRow struct {
		AnnouncementId int64
		Cnt            int64
	}
	rcRows := make([]rcRow, 0)
	if err := e.Orm.Table("announcement_read_log").
		Select("announcement_id AS announcement_id, COUNT(*) AS cnt").
		Where("announcement_id IN ?", ids).
		Group("announcement_id").
		Find(&rcRows).Error; err != nil {
		e.Log.Errorf("load read counts: %s", err)
		return err
	}
	rcMap := make(map[int64]int64, len(rcRows))
	for _, r := range rcRows {
		rcMap[r.AnnouncementId] = r.Cnt
	}

	// 当前用户已读
	readSet := make(map[int64]struct{})
	if currentUserId > 0 {
		var readIds []int64
		if err := e.Orm.Table("announcement_read_log").
			Where("user_id = ?", currentUserId).
			Where("announcement_id IN ?", ids).
			Pluck("announcement_id", &readIds).Error; err != nil {
			e.Log.Errorf("load own reads: %s", err)
			return err
		}
		for _, id := range readIds {
			readSet[id] = struct{}{}
		}
	}

	out := make([]dto.AnnouncementListItem, 0, len(rows))
	for _, r := range rows {
		_, isRead := readSet[r.AnnouncementId]
		out = append(out, dto.AnnouncementListItem{
			Announcement: r,
			DeptIds:      scopeMap[r.AnnouncementId],
			IsRead:       isRead,
			ReadCount:    rcMap[r.AnnouncementId],
		})
	}
	*list = out
	return nil
}

// Get 详情查询，含 scope 与 read 派生字段。
//
// 数据权限：详情读也走 actions.Permission，防止 data_scope=5 的用户直接拿 ID
// 越权读他人公告（"找不到"对越权请求与"真没有"返回一致）。
func (e *Announcement) Get(c *dto.AnnouncementGetReq, p *actions.DataPermission, item *dto.AnnouncementListItem, currentUserId int) error {
	var ann models.Announcement
	if err := e.Orm.Model(&models.Announcement{}).
		Scopes(actions.Permission((&models.Announcement{}).TableName(), p)).
		First(&ann, c.GetId()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("公告不存在或已删除")
		}
		e.Log.Errorf("get announcement: %s", err)
		return err
	}

	var deptIds []int
	if err := e.Orm.Table("announcement_scope").
		Where("announcement_id = ?", ann.AnnouncementId).
		Pluck("dept_id", &deptIds).Error; err != nil {
		e.Log.Errorf("load scope: %s", err)
		return err
	}

	var readCount int64
	if err := e.Orm.Table("announcement_read_log").
		Where("announcement_id = ?", ann.AnnouncementId).
		Count(&readCount).Error; err != nil {
		e.Log.Errorf("count reads: %s", err)
		return err
	}

	isRead := false
	if currentUserId > 0 {
		var c int64
		if err := e.Orm.Table("announcement_read_log").
			Where("announcement_id = ? AND user_id = ?", ann.AnnouncementId, currentUserId).
			Count(&c).Error; err != nil {
			e.Log.Errorf("check read: %s", err)
			return err
		}
		isRead = c > 0
	}

	*item = dto.AnnouncementListItem{
		Announcement: ann,
		DeptIds:      deptIds,
		IsRead:       isRead,
		ReadCount:    readCount,
	}
	return nil
}

// Insert 新增公告并写入 scope。Content 入库前做 XSS 过滤。
func (e *Announcement) Insert(c *dto.AnnouncementInsertReq) (int64, error) {
	c.Content = SanitizeContent(c.Content)

	var ann models.Announcement
	c.Generate(&ann)

	err := e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ann).Error; err != nil {
			return err
		}
		if len(c.DeptIds) > 0 {
			scopes := make([]models.AnnouncementScope, 0, len(c.DeptIds))
			seen := make(map[int]struct{}, len(c.DeptIds))
			for _, did := range c.DeptIds {
				if did == 0 {
					continue
				}
				if _, ok := seen[did]; ok {
					continue
				}
				seen[did] = struct{}{}
				scopes = append(scopes, models.AnnouncementScope{
					AnnouncementId: ann.AnnouncementId,
					DeptId:         did,
				})
			}
			if len(scopes) > 0 {
				if err := tx.Create(&scopes).Error; err != nil {
					return err
				}
			}
		}
		if _, err := rebindAnnouncementAttachments(tx, ann.AnnouncementId, ann.Content, ann.CoverImageUrl); err != nil {
			e.Log.Warnf("rebind attachments after insert: %s", err)
		}
		return nil
	})
	if err != nil {
		e.Log.Errorf("insert announcement: %s", err)
		return 0, err
	}
	c.AnnouncementId = ann.AnnouncementId
	return ann.AnnouncementId, nil
}

// Update 修改公告并重建 scope。
//
// 数据权限：更新前的"找到这条记录"走 actions.Permission，防止 data_scope=5 的
// 用户直接 PUT 越权改他人公告（命中 scope 外则视为不存在）。
func (e *Announcement) Update(c *dto.AnnouncementUpdateReq, p *actions.DataPermission) error {
	c.Content = SanitizeContent(c.Content)

	return e.Orm.Transaction(func(tx *gorm.DB) error {
		var existing models.Announcement
		if err := tx.Model(&models.Announcement{}).
			Scopes(actions.Permission((&models.Announcement{}).TableName(), p)).
			First(&existing, c.AnnouncementId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("公告不存在或已删除")
			}
			return err
		}
		c.Generate(&existing)
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}

		// 仅当请求显式提供 DeptIds 时才重建 scope
		if c.DeptIds != nil {
			if err := tx.Where("announcement_id = ?", c.AnnouncementId).
				Delete(&models.AnnouncementScope{}).Error; err != nil {
				return err
			}
			if len(c.DeptIds) > 0 {
				scopes := make([]models.AnnouncementScope, 0, len(c.DeptIds))
				seen := make(map[int]struct{}, len(c.DeptIds))
				for _, did := range c.DeptIds {
					if did == 0 {
						continue
					}
					if _, ok := seen[did]; ok {
						continue
					}
					seen[did] = struct{}{}
					scopes = append(scopes, models.AnnouncementScope{
						AnnouncementId: c.AnnouncementId,
						DeptId:         did,
					})
				}
				if len(scopes) > 0 {
					if err := tx.Create(&scopes).Error; err != nil {
						return err
					}
				}
			}
		}
		if _, err := rebindAnnouncementAttachments(tx, existing.AnnouncementId, existing.Content, existing.CoverImageUrl); err != nil {
			e.Log.Warnf("rebind attachments after update: %s", err)
		}
		return nil
	})
}

// Remove 批量删除公告，级联清理 scope 与 read_log。
//
// 数据权限：先按 actions.Permission 过滤 c.Ids 得到当前用户实际可删除的子集，
// 再级联删除。这样避免 data_scope=5 的用户传一组 ID 把别人的公告也带删掉。
// 若过滤后子集为空，返回"没有可删除的公告"语义错误。
func (e *Announcement) Remove(c *dto.AnnouncementDeleteReq, p *actions.DataPermission) error {
	if len(c.Ids) == 0 {
		return errors.New("ids 不能为空")
	}
	tableName := (&models.Announcement{}).TableName()
	var allowed []int64
	if err := e.Orm.Model(&models.Announcement{}).
		Scopes(actions.Permission(tableName, p)).
		Where(tableName+".announcement_id IN ?", c.Ids).
		Pluck(tableName+".announcement_id", &allowed).Error; err != nil {
		return err
	}
	if len(allowed) == 0 {
		return errors.New("公告不存在或无权删除")
	}
	var pendingPaths []string
	err := e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("announcement_id IN ?", allowed).
			Delete(&models.AnnouncementScope{}).Error; err != nil {
			return err
		}
		if err := tx.Where("announcement_id IN ?", allowed).
			Delete(&models.AnnouncementReadLog{}).Error; err != nil {
			return err
		}
		var data models.Announcement
		if err := tx.Delete(&data, allowed).Error; err != nil {
			return err
		}
		paths, err := removeAnnouncementAttachments(tx, allowed)
		if err != nil {
			return err
		}
		pendingPaths = paths
		return nil
	})
	if err == nil {
		for _, p := range pendingPaths {
			if rmErr := os.Remove(p); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
				e.Log.Warnf("remove file %s: %s", p, rmErr)
			}
		}
	}
	return err
}

// MarkRead 幂等记录已读：首次写入读时间，重复调用不报错。
//
// 数据权限：标记前的"公告存在"校验也走 scope。data_scope=5 的用户标记 scope 外
// 公告时返回"公告不存在"，与读侧一致。
func (e *Announcement) MarkRead(announcementId int64, userId int, p *actions.DataPermission) error {
	if announcementId <= 0 || userId <= 0 {
		return errors.New("invalid params")
	}
	// 确认公告存在且在当前用户的 scope 内（避免任意 ID 注入读日志）
	var cnt int64
	if err := e.Orm.Model(&models.Announcement{}).
		Scopes(actions.Permission((&models.Announcement{}).TableName(), p)).
		Where("announcement_id = ?", announcementId).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("公告不存在")
	}

	// 已存在则跳过
	var rc int64
	if err := e.Orm.Model(&models.AnnouncementReadLog{}).
		Where("announcement_id = ? AND user_id = ?", announcementId, userId).
		Count(&rc).Error; err != nil {
		return err
	}
	if rc > 0 {
		return nil
	}
	row := models.AnnouncementReadLog{
		UserId:         userId,
		AnnouncementId: announcementId,
		ReadAt:         time.Now(),
	}
	if err := e.Orm.Create(&row).Error; err != nil {
		// 并发场景下可能命中复合主键冲突——视为幂等成功
		e.Log.Warnf("mark-read create (likely race): %s", err)
		return nil
	}
	return nil
}
