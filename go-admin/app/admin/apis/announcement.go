package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/middleware"
)

type Announcement struct {
	api.Api
}

// GetPage 公告列表
// @Summary 公告列表
// @Tags 公告
// @Param title query string false "标题"
// @Param status query int false "状态"
// @Param onlyValid query int false "仅生效中"
// @Param onlyVisible query int false "仅当前用户可见"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement [get]
// @Security Bearer
func (e Announcement) GetPage(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	list := make([]dto.AnnouncementListItem, 0)
	var count int64
	if err := s.GetPage(&req, &list, &count, user.GetUserId(c)); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 公告详情
// @Summary 公告详情
// @Tags 公告
// @Param id path int true "公告ID"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement/{id} [get]
// @Security Bearer
func (e Announcement) Get(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementGetReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var item dto.AnnouncementListItem
	if err := s.Get(&req, &item, user.GetUserId(c)); err != nil {
		e.Error(500, err, fmt.Sprintf("公告获取失败：%s", err.Error()))
		return
	}
	e.OK(item, "查询成功")
}

// Insert 新增公告
// @Summary 新增公告
// @Tags 公告
// @Accept application/json
// @Param data body dto.AnnouncementInsertReq true "data"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement [post]
// @Security Bearer
func (e Announcement) Insert(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementInsertReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	id, err := s.Insert(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("新建公告失败：%s", err.Error()))
		return
	}
	middleware.AuditLogCreate(c,
		"公告管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategoryAnnouncement,
			ID:    id,
			Label: req.Title,
		},
		map[string]interface{}{
			"title":     req.Title,
			"status":    req.Status,
			"isTop":     req.IsTop,
			"deptCount": len(req.DeptIds),
		},
		"admin.announcement.insert",
	)
	e.OK(id, "创建成功")
}

// Update 修改公告
// @Summary 修改公告
// @Tags 公告
// @Accept application/json
// @Param id path int true "公告ID"
// @Param data body dto.AnnouncementUpdateReq true "data"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement/{id} [put]
// @Security Bearer
func (e Announcement) Update(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementUpdateReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Update(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("公告更新失败：%s", err.Error()))
		return
	}
	middleware.AuditLogUpdate(c,
		"公告管理",
		middleware.AuditTarget{
			Type:  middleware.AuditCategoryAnnouncement,
			ID:    req.AnnouncementId,
			Label: req.Title,
		},
		nil,
		map[string]interface{}{
			"title":     req.Title,
			"status":    req.Status,
			"isTop":     req.IsTop,
			"deptCount": len(req.DeptIds),
		},
		"admin.announcement.update",
	)
	e.OK(req.AnnouncementId, "更新成功")
}

// Delete 批量删除公告
// @Summary 批量删除公告
// @Tags 公告
// @Param data body dto.AnnouncementDeleteReq true "data"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement [delete]
// @Security Bearer
func (e Announcement) Delete(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementDeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	if err := s.Remove(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("公告删除失败：%s", err.Error()))
		return
	}
	middleware.AuditLogDelete(c,
		"公告管理",
		middleware.AuditTarget{
			Type: middleware.AuditCategoryAnnouncement,
			ID:   req.Ids,
		},
		map[string]interface{}{"ids": req.Ids, "count": len(req.Ids)},
		"admin.announcement.delete",
	)
	e.OK(req.Ids, "删除成功")
}

// MarkRead 标记公告已读
// @Summary 标记公告已读
// @Tags 公告
// @Param id path int true "公告ID"
// @Success 200 {object} response.Response
// @Router /api/v1/announcement/{id}/read [post]
// @Security Bearer
func (e Announcement) MarkRead(c *gin.Context) {
	s := service.Announcement{}
	req := dto.AnnouncementMarkReadReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	uid := user.GetUserId(c)
	if uid == 0 {
		e.Error(401, nil, "未登录")
		return
	}
	if err := s.MarkRead(req.AnnouncementId, uid); err != nil {
		e.Error(500, err, fmt.Sprintf("标记已读失败：%s", err.Error()))
		return
	}
	middleware.AuditLog(c, middleware.AuditEntry{
		Title:  "公告管理",
		Action: middleware.AuditActionUpdate,
		Target: middleware.AuditTarget{
			Type: middleware.AuditCategoryAnnouncement,
			ID:   req.AnnouncementId,
		},
		After:  map[string]interface{}{"action": "mark-read", "userId": uid},
		Method: "admin.announcement.markRead",
	})
	e.OK(req.AnnouncementId, "已读")
}
