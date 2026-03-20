package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type SysNoticeGetPageReq struct {
	dto.Pagination `search:"-"`
	Title          string `form:"title" search:"type:contains;column:title;table:sys_notice"`
	Type           string `form:"type" search:"type:exact;column:type;table:sys_notice"`
	Status         string `form:"status" search:"type:exact;column:status;table:sys_notice"`
}

func (m *SysNoticeGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type SysNoticeControl struct {
	Id      int    `uri:"Id" comment:"编码"`
	Title   string `json:"title" comment:"标题"`
	Content string `json:"content" comment:"内容"`
	Type    string `json:"type" comment:"类型"`
	Status  string `json:"status" comment:"状态"`
	Sort    int    `json:"sort" comment:"排序"`
	Remark  string `json:"remark" comment:"备注"`
	common.ControlBy
}

func (s *SysNoticeControl) Generate(model *models.SysNotice) {
	if s.Id != 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.Title = s.Title
	model.Content = s.Content
	model.Type = s.Type
	model.Status = s.Status
	model.Sort = s.Sort
	model.Remark = s.Remark
}

func (s *SysNoticeControl) GetId() interface{} {
	return s.Id
}

type SysNoticeGetReq struct {
	Id int `uri:"id"`
}

func (s *SysNoticeGetReq) GetId() interface{} {
	return s.Id
}

type SysNoticeDeleteReq struct {
	Ids []int `json:"ids"`
	common.ControlBy
}

func (s *SysNoticeDeleteReq) GetId() interface{} {
	return s.Ids
}
