package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SkuBrandPageReq 列表/搜索请求
type SkuBrandPageReq struct {
	dto.Pagination `search:"-"`
	BrandName      string `form:"brandName" search:"type:contains;column:brand_name;table:sku_brand" comment:"品牌名称"`
	Status         int    `form:"status" search:"type:exact;column:status;table:sku_brand" comment:"状态"`
}

func (m *SkuBrandPageReq) GetNeedSearch() interface{} {
	return *m
}

// SkuBrandGetReq 详情请求
type SkuBrandGetReq struct {
	Id int64 `uri:"id"`
}

func (s *SkuBrandGetReq) GetId() interface{} {
	return s.Id
}

// SkuBrandInsertReq 新增请求
type SkuBrandInsertReq struct {
	BrandId      int64  `json:"brandId"`
	BrandName    string `json:"brandName" binding:"required" comment:"品牌名称"`
	BrandLogoUrl string `json:"brandLogoUrl" comment:"品牌Logo URL"`
	Sort         int    `json:"sort" comment:"排序"`
	Status       int    `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SkuBrandInsertReq) Generate(model *models.SkuBrand) {
	model.BrandName = s.BrandName
	model.BrandLogoUrl = s.BrandLogoUrl
	model.Sort = s.Sort
	if s.Status == 0 {
		model.Status = models.SkuBrandStatusEnabled
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

func (s *SkuBrandInsertReq) GetId() interface{} {
	return s.BrandId
}

// SkuBrandUpdateReq 修改请求
type SkuBrandUpdateReq struct {
	BrandId      int64  `uri:"id" json:"brandId"`
	BrandName    string `json:"brandName" binding:"required" comment:"品牌名称"`
	BrandLogoUrl string `json:"brandLogoUrl" comment:"品牌Logo URL"`
	Sort         int    `json:"sort" comment:"排序"`
	Status       int    `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SkuBrandUpdateReq) Generate(model *models.SkuBrand) {
	model.BrandId = s.BrandId
	model.BrandName = s.BrandName
	model.BrandLogoUrl = s.BrandLogoUrl
	model.Sort = s.Sort
	if s.Status != 0 {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuBrandUpdateReq) GetId() interface{} {
	return s.BrandId
}

// SkuBrandDeleteReq 批量删除请求
type SkuBrandDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *SkuBrandDeleteReq) Generate(model *models.SkuBrand) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuBrandDeleteReq) GetId() interface{} {
	return s.Ids
}
