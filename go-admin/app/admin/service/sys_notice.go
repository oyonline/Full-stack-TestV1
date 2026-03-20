package service

import (
	"errors"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"
	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type SysNotice struct {
	service.Service
}

func (e *SysNotice) GetPage(c *dto.SysNoticeGetPageReq, list *[]models.SysNotice, count *int64) error {
	err := e.Orm.
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("Service GetNoticePage error: %s", err)
		return err
	}
	return nil
}

func (e *SysNotice) Get(d *dto.SysNoticeGetReq, model *models.SysNotice) error {
	err := e.Orm.FirstOrInit(model, d.GetId()).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		_ = e.AddError(err)
		return err
	}
	if model.Id == 0 {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service Get error: %s", err)
		_ = e.AddError(err)
		return err
	}
	return nil
}

func (e *SysNotice) Insert(c *dto.SysNoticeControl) error {
	var err error
	var data models.SysNotice
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("Service Insert error: %s", err)
		return err
	}
	return nil
}

func (e *SysNotice) Update(c *dto.SysNoticeControl) error {
	var err error
	var model = models.SysNotice{}
	e.Orm.First(&model, c.GetId())
	c.Generate(&model)
	db := e.Orm.Save(&model)
	err = db.Error
	if err != nil {
		e.Log.Errorf("Service Update error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

func (e *SysNotice) Remove(d *dto.SysNoticeDeleteReq) error {
	var err error
	var data models.SysNotice
	db := e.Orm.Delete(&data, d.Ids)
	if err = db.Error; err != nil {
		e.Log.Errorf("Service Remove error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("无权删除该数据")
		return err
	}
	return nil
}
