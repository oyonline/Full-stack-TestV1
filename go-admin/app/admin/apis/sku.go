package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"

	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
)

type Sku struct {
	api.Api
}

// GetPage SKU 列表（只读，dataScope 继承 SPU）
func (e Sku) GetPage(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuPageReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	list := make([]dto.SkuListItem, 0)
	var count int64
	p := actions.GetPermissionFromContext(c)
	if err := s.GetPage(&req, p, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get SKU 详情（只读，dataScope 继承 SPU）
func (e Sku) Get(c *gin.Context) {
	s := service.Sku{}
	req := dto.SkuGetReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors; err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var item dto.SkuListItem
	p := actions.GetPermissionFromContext(c)
	if err := s.Get(&req, p, &item); err != nil {
		e.Error(500, err, fmt.Sprintf("SKU 获取失败：%s", err.Error()))
		return
	}
	e.OK(item, "查询成功")
}
