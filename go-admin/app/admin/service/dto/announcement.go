package dto

import (
	"time"

	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// AnnouncementPageReq 列表/搜索请求
type AnnouncementPageReq struct {
	dto.Pagination `search:"-"`
	Title          string `form:"title" search:"type:contains;column:title;table:announcement" comment:"标题"`
	Status         int    `form:"status" search:"type:exact;column:status;table:announcement" comment:"状态"`
	IsTop          int    `form:"isTop" search:"type:exact;column:is_top;table:announcement" comment:"是否置顶"`
	OnlyValid      int    `form:"onlyValid" comment:"是否仅返回当前生效的（按 publish_at/expire_at 过滤）"`
	OnlyVisible    int    `form:"onlyVisible" comment:"是否仅返回当前用户可见的（按部门 scope 过滤）"`
}

func (m *AnnouncementPageReq) GetNeedSearch() interface{} {
	return *m
}

// AnnouncementGetReq 详情请求
type AnnouncementGetReq struct {
	Id int64 `uri:"id"`
}

func (s *AnnouncementGetReq) GetId() interface{} {
	return s.Id
}

// AnnouncementInsertReq 新增请求
type AnnouncementInsertReq struct {
	AnnouncementId int64      `json:"announcementId" comment:"公告ID"`
	Title          string     `json:"title" binding:"required" comment:"标题"`
	Content        string     `json:"content" comment:"正文（富文本HTML）"`
	CoverImageUrl  string     `json:"coverImageUrl" comment:"封面图URL"`
	Status         int        `json:"status" comment:"状态"`
	IsTop          int        `json:"isTop" comment:"是否置顶"`
	TopSort        int        `json:"topSort" comment:"置顶排序值"`
	PublishAt      *time.Time `json:"publishAt" comment:"生效起始"`
	ExpireAt       *time.Time `json:"expireAt" comment:"失效时间"`
	DeptIds        []int      `json:"deptIds" comment:"可见部门ID集合"`
	Remark         string     `json:"remark" comment:"备注"`
	common.ControlBy
}

func (s *AnnouncementInsertReq) Generate(model *models.Announcement) {
	model.Title = s.Title
	model.Content = s.Content
	model.CoverImageUrl = s.CoverImageUrl
	if s.Status == 0 {
		model.Status = models.AnnouncementStatusDraft
	} else {
		model.Status = s.Status
	}
	model.IsTop = s.IsTop
	model.TopSort = s.TopSort
	model.PublishAt = s.PublishAt
	model.ExpireAt = s.ExpireAt
	model.Remark = s.Remark
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
		model.CreatorId = s.CreateBy
	}
}

func (s *AnnouncementInsertReq) GetId() interface{} {
	return s.AnnouncementId
}

// AnnouncementUpdateReq 修改请求
type AnnouncementUpdateReq struct {
	AnnouncementId int64      `uri:"id" json:"announcementId" comment:"公告ID"`
	Title          string     `json:"title" binding:"required" comment:"标题"`
	Content        string     `json:"content" comment:"正文"`
	CoverImageUrl  string     `json:"coverImageUrl" comment:"封面图URL"`
	Status         int        `json:"status" comment:"状态"`
	IsTop          int        `json:"isTop" comment:"是否置顶"`
	TopSort        int        `json:"topSort" comment:"置顶排序值"`
	PublishAt      *time.Time `json:"publishAt" comment:"生效起始"`
	ExpireAt       *time.Time `json:"expireAt" comment:"失效时间"`
	DeptIds        []int      `json:"deptIds" comment:"可见部门ID集合"`
	Remark         string     `json:"remark" comment:"备注"`
	common.ControlBy
}

func (s *AnnouncementUpdateReq) Generate(model *models.Announcement) {
	model.AnnouncementId = s.AnnouncementId
	model.Title = s.Title
	model.Content = s.Content
	model.CoverImageUrl = s.CoverImageUrl
	if s.Status != 0 {
		model.Status = s.Status
	}
	model.IsTop = s.IsTop
	model.TopSort = s.TopSort
	model.PublishAt = s.PublishAt
	model.ExpireAt = s.ExpireAt
	model.Remark = s.Remark
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *AnnouncementUpdateReq) GetId() interface{} {
	return s.AnnouncementId
}

// AnnouncementDeleteReq 批量删除请求
type AnnouncementDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *AnnouncementDeleteReq) Generate(model *models.Announcement) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *AnnouncementDeleteReq) GetId() interface{} {
	return s.Ids
}

// AnnouncementMarkReadReq mark-read 请求
type AnnouncementMarkReadReq struct {
	AnnouncementId int64 `uri:"id"`
}

func (s *AnnouncementMarkReadReq) GetId() interface{} {
	return s.AnnouncementId
}

// AnnouncementListItem 列表/详情返回的视图模型，含派生字段
type AnnouncementListItem struct {
	models.Announcement
	DeptIds   []int `json:"deptIds"`
	IsRead    bool  `json:"isRead"`
	ReadCount int64 `json:"readCount"`
}
