package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SpuPageReq SPU 列表/搜索请求。
// 注意：所有非 search 字段（OnlyMine 等）显式 search:"-" 避免 reflect numfield panic（my-099 教训）。
type SpuPageReq struct {
	dto.Pagination `search:"-"`
	SpuCode        string `form:"spuCode" search:"type:contains;column:spu_code;table:spu" comment:"SPU 编码"`
	SpuName        string `form:"spuName" search:"type:contains;column:spu_name;table:spu" comment:"SPU 名称"`
	CategoryId     int64  `form:"categoryId" search:"type:exact;column:category_id;table:spu" comment:"类目"`
	BrandId        int64  `form:"brandId" search:"type:exact;column:brand_id;table:spu" comment:"品牌"`
	Status         int    `form:"status" search:"type:exact;column:status;table:spu" comment:"状态"`
}

func (m *SpuPageReq) GetNeedSearch() interface{} {
	return *m
}

// SpuGetReq 详情请求
type SpuGetReq struct {
	Id int64 `uri:"id"`
}

func (s *SpuGetReq) GetId() interface{} {
	return s.Id
}

// SpuInsertReq 新增请求
type SpuInsertReq struct {
	SpuId        int64  `json:"spuId"`
	SpuCode      string `json:"spuCode" binding:"required" comment:"SPU 编码"`
	SpuName      string `json:"spuName" binding:"required" comment:"SPU 名称"`
	CategoryId   int64  `json:"categoryId" comment:"叶子类目ID"`
	BrandId      int64  `json:"brandId" comment:"品牌ID"`
	Description  string `json:"description" comment:"富文本描述"`
	MainImageUrl string `json:"mainImageUrl" comment:"主图 URL"`
	DetailImages string `json:"detailImages" comment:"详情图 JSON"`
	Status       int    `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *SpuInsertReq) Generate(model *models.Spu) {
	model.SpuCode = s.SpuCode
	model.SpuName = s.SpuName
	model.CategoryId = s.CategoryId
	model.BrandId = s.BrandId
	model.Description = s.Description
	model.MainImageUrl = s.MainImageUrl
	model.DetailImages = s.DetailImages
	if s.Status == 0 {
		model.Status = models.SpuStatusDraft
	} else {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
		model.CreatorId = s.CreateBy
	}
}

func (s *SpuInsertReq) GetId() interface{} {
	return s.SpuId
}

// SpuUpdateReq 修改请求
type SpuUpdateReq struct {
	SpuId        int64  `uri:"id" json:"spuId"`
	SpuCode      string `json:"spuCode" binding:"required"`
	SpuName      string `json:"spuName" binding:"required"`
	CategoryId   int64  `json:"categoryId"`
	BrandId      int64  `json:"brandId"`
	Description  string `json:"description"`
	MainImageUrl string `json:"mainImageUrl"`
	DetailImages string `json:"detailImages"`
	Status       int    `json:"status"`
	common.ControlBy
}

func (s *SpuUpdateReq) Generate(model *models.Spu) {
	model.SpuId = s.SpuId
	model.SpuCode = s.SpuCode
	model.SpuName = s.SpuName
	model.CategoryId = s.CategoryId
	model.BrandId = s.BrandId
	model.Description = s.Description
	model.MainImageUrl = s.MainImageUrl
	model.DetailImages = s.DetailImages
	if s.Status != 0 {
		model.Status = s.Status
	}
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SpuUpdateReq) GetId() interface{} {
	return s.SpuId
}

// SpuDeleteReq 批量删除请求
type SpuDeleteReq struct {
	Ids []int64 `json:"ids"`
	common.ControlBy
}

func (s *SpuDeleteReq) Generate(model *models.Spu) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
}

func (s *SpuDeleteReq) GetId() interface{} {
	return s.Ids
}

// SpuSubmitReq 提交审核请求
type SpuSubmitReq struct {
	SpuId        int64  `uri:"id"`
	DefinitionId int    `json:"definitionId" comment:"流程定义ID（可选；为空则按 definition_key='spu_create_review' 查找启用版本）"`
	Remark       string `json:"remark" comment:"提交说明"`
}

func (s *SpuSubmitReq) GetId() interface{} {
	return s.SpuId
}

// SpuOnlineReq 上架请求
type SpuOnlineReq struct {
	SpuId int64 `uri:"id"`
}

func (s *SpuOnlineReq) GetId() interface{} {
	return s.SpuId
}

// SpuOfflineReq 下架请求
type SpuOfflineReq struct {
	SpuId int64 `uri:"id"`
}

func (s *SpuOfflineReq) GetId() interface{} {
	return s.SpuId
}

// SpuListItem 列表/详情视图模型，含 workflow 派生字段。
// WorkflowStatus / WorkflowTitle 来自 workflow_business_binding JOIN（可能为空）。
type SpuListItem struct {
	models.Spu
	WorkflowStatus string `json:"workflowStatus"`
	WorkflowTitle  string `json:"workflowTitle"`
}
