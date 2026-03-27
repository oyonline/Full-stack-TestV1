package service

import (
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/app/admin/service/vo"
	cDto "go-admin/common/dto"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type FeeRequest struct {
	service.Service
}

func (e *FeeRequest) GetPage(c *dto.FeeRequestPageReq, count *int64) ([]vo.FeeRequestLog, error) {
	var err error
	var logList []vo.FeeRequestLog
	err = e.Orm.Model(&models.FeeRequestLog{}).Preload("BearDept").Preload("ReqUser").Preload("ReqDept").Scopes(
		cDto.MakeCondition(c.GetNeedSearch()),
		cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
	).Find(&logList).Offset(-1).Limit(-1).Count(count).Error
	return logList, err
}

func (e *FeeRequest) Get(id int64) (vo.FeeRequestLogDetail, error) {
	var err error
	var logList vo.FeeRequestLogDetail
	err = e.Orm.Table("fee_request_log l").
		Preload("MainRecord").Preload("BearDept").Preload("ReqUser").Preload("ReqDept").Preload("Timeline.OptUser").Preload("Attachment").
		Joins("INNER JOIN budget_fee_category_details bd ON l.budget_detail_id = bd.id").
		Joins("INNER JOIN budget_fee_category bc ON bd.budget_fee_category_id = bc.id").
		Joins("INNER JOIN kingdee_organize_info oi ON oi.forgid = l.org_code").
		Joins("INNER JOIN feishu_approval_form af ON af.instance_code = l.instance_code").
		Where("l.id = ?", id).
		Select("l.*, bc.category_code, bc.category_name,oi.fname,af.kingdee_department_code").
		First(&logList).Error

	return logList, err
}
