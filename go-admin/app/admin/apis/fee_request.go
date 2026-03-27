package apis

import (
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
)

type FeeRequest struct {
	api.Api
}

// GetPage 飞书费用申请列表
// @Summary 飞书费用申请
// @Description 飞书费用申请列表
// @Tags 飞书费用申请列表
// @Param data body dto.FeeRequestPageReq true "data"
// @Success 200 {object} response.Response{data=response.Page{list=[]vo.FeeRequestLog}} "{"code": 200, "data": [...]}"
// @Router /api/v1/fee-request/list [post]
// @Security Bearer
func (e FeeRequest) GetPage(c *gin.Context) {
	req := dto.FeeRequestPageReq{}
	s := service.FeeRequest{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	var count int64
	dataList, err := s.GetPage(&req, &count)
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	e.PageOK(dataList, int(count), req.PageIndex, req.PageSize, "查询成功")
}

// Get 飞书费用申请详情
// @Summary 飞书费用申请
// @Description 飞书费用申请详情
// @Tags 飞书费用申请详情
// @Param id path int true "id"
// @Success 200 {object} response.Response{data=models.FeeRequestLog} "{"code": 200, "data": [...]}"
// @Router /api/v1/fee-request/{id} [get]
// @Security Bearer
func (e FeeRequest) Get(c *gin.Context) {
	req := dto.FeeRequestGet{}
	s := service.FeeRequest{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	data, err := s.Get(req.Id)
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	e.OK(data, "查询成功")
}
