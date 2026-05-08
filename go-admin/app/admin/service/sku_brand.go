package service

import (
	"go-admin/app/admin/models"
	"go-admin/common/baseservice"
)

// SkuBrand 品牌字典：标准 CRUD，由 BaseService 提供五件套。
type SkuBrand struct {
	baseservice.BaseService[models.SkuBrand]
}
