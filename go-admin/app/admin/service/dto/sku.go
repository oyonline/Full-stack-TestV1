package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SkuListItem 列表/详情视图模型，含 SPU 关键字段（通过 LEFT JOIN 填充）。
type SkuListItem struct {
	models.Sku
	SpuCode   string `json:"spuCode"`
	SpuName   string `json:"spuName"`
	SpuStatus int    `json:"spuStatus"`
}

// SkuPageReq SKU 列表/搜索请求
type SkuPageReq struct {
	dto.Pagination `search:"-"`
	SpuId          int64  `form:"spuId" search:"type:exact;column:spu_id;table:sku" comment:"SPU ID"`
	SkuCode        string `form:"skuCode" search:"type:contains;column:sku_code;table:sku" comment:"SKU 编码"`
	SkuName        string `form:"skuName" search:"type:contains;column:sku_name;table:sku" comment:"SKU 名称"`
	Status         int    `form:"status" search:"type:exact;column:status;table:sku" comment:"状态"`
}

func (m *SkuPageReq) GetNeedSearch() interface{} {
	return *m
}

// SkuGetReq 详情请求
type SkuGetReq struct {
	Id int64 `uri:"id"`
}

func (s *SkuGetReq) GetId() interface{} {
	return s.Id
}

// SkuInsertReq 新增请求
type SkuInsertReq struct {
	SkuId   int64   `json:"skuId"`
	SpuId   int64   `json:"spuId" binding:"required" comment:"所属 SPU"`
	SkuCode string  `json:"skuCode" binding:"required" comment:"SKU 编码"`
	SkuName string  `json:"skuName" comment:"SKU 名称"`
	Spec    string  `json:"spec" comment:"规格"`
	Unit    string  `json:"unit" comment:"单位"`
	Price   float64 `json:"price" comment:"价格"`
	Status  int     `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SkuInsertReq) Generate(model *models.Sku) {
	model.SpuId = s.SpuId
	model.SkuCode = s.SkuCode
	model.SkuName = s.SkuName
	model.Spec = s.Spec
	model.Unit = s.Unit
	model.Price = s.Price
	if s.Status == 0 {
		model.Status = models.SkuStatusDisabled
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

func (s *SkuInsertReq) GetId() interface{} {
	return s.SkuId
}

// SkuUpdateReq 修改请求
type SkuUpdateReq struct {
	SkuId   int64   `uri:"id" json:"skuId"`
	SpuId   int64   `json:"spuId" binding:"required"`
	SkuCode string  `json:"skuCode" binding:"required"`
	SkuName string  `json:"skuName"`
	Spec    string  `json:"spec"`
	Unit    string  `json:"unit"`
	Price   float64 `json:"price"`
	Status  int     `json:"status"`
	common.ControlBy
}

func (s *SkuUpdateReq) Generate(model *models.Sku) {
	model.SkuId = s.SkuId
	model.SpuId = s.SpuId
	model.SkuCode = s.SkuCode
	model.SkuName = s.SkuName
	model.Spec = s.Spec
	model.Unit = s.Unit
	model.Price = s.Price
	if s.Status != 0 {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuUpdateReq) GetId() interface{} {
	return s.SkuId
}

// SkuDeleteReq 批量删除请求
type SkuDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *SkuDeleteReq) Generate(model *models.Sku) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SkuDeleteReq) GetId() interface{} {
	return s.Ids
}
