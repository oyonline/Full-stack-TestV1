package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SkuCategoryPageReq 列表/搜索请求。
// search:"-" 显式标记非 search 字段，避免 reflect numfield panic（吸取 my-099 教训）。
type SkuCategoryPageReq struct {
	dto.Pagination `search:"-"`
	CategoryName   string `form:"categoryName" search:"type:contains;column:category_name;table:sku_category" comment:"类目名称"`
	ParentId       int64  `form:"parentId" search:"type:exact;column:parent_id;table:sku_category" comment:"父级ID"`
	Status         int    `form:"status" search:"type:exact;column:status;table:sku_category" comment:"状态"`
}

func (m *SkuCategoryPageReq) GetNeedSearch() interface{} {
	return *m
}

// SkuCategoryGetReq 详情请求
type SkuCategoryGetReq struct {
	Id int64 `uri:"id"`
}

func (s *SkuCategoryGetReq) GetId() interface{} {
	return s.Id
}

// SkuCategoryInsertReq 新增请求
type SkuCategoryInsertReq struct {
	CategoryId   int64  `json:"categoryId"`
	CategoryName string `json:"categoryName" binding:"required" comment:"类目名称"`
	ParentId     int64  `json:"parentId" comment:"父级ID（0=根）"`
	Sort         int    `json:"sort" comment:"排序"`
	Status       int    `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SkuCategoryInsertReq) Generate(model *models.SkuCategory) {
	model.CategoryName = s.CategoryName
	model.ParentId = s.ParentId
	model.Sort = s.Sort
	if s.Status == 0 {
		model.Status = models.SkuCategoryStatusEnabled
	} else {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *SkuCategoryInsertReq) GetId() interface{} {
	return s.CategoryId
}

// SkuCategoryUpdateReq 修改请求
type SkuCategoryUpdateReq struct {
	CategoryId   int64  `uri:"id" json:"categoryId"`
	CategoryName string `json:"categoryName" binding:"required" comment:"类目名称"`
	ParentId     int64  `json:"parentId" comment:"父级ID"`
	Sort         int    `json:"sort" comment:"排序"`
	Status       int    `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SkuCategoryUpdateReq) Generate(model *models.SkuCategory) {
	model.CategoryId = s.CategoryId
	model.CategoryName = s.CategoryName
	model.ParentId = s.ParentId
	model.Sort = s.Sort
	if s.Status != 0 {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuCategoryUpdateReq) GetId() interface{} {
	return s.CategoryId
}

// SkuCategoryDeleteReq 批量删除请求
type SkuCategoryDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *SkuCategoryDeleteReq) Generate(model *models.SkuCategory) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuCategoryDeleteReq) GetId() interface{} {
	return s.Ids
}

// SkuCategoryTreeNode 类目树节点：包含 children，便于前端直接渲染。
type SkuCategoryTreeNode struct {
	models.SkuCategory
	Children []*SkuCategoryTreeNode `json:"children"`
}
