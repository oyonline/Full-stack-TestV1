package service

import (
	"errors"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

type SysJob struct {
	service.Service
}

// GetPage 获取SysJob列表
func (e *SysJob) GetPage(c *dto.SysJobGetPageReq, p *actions.DataPermission, list *[]models.SysJob, count *int64) error {
	var err error
	var data models.SysJob

	err = e.Orm.Debug().
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Get 获取SysJob对象
func (e *SysJob) Get(d *dto.SysJobById, p *actions.DataPermission, model *models.SysJob) error {
	var data models.SysJob

	err := e.Orm.Model(&data).Debug().
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Insert 创建SysJob对象
func (e *SysJob) Insert(c *dto.SysJobInsertReq) error {
	var err error
	var data models.SysJob
	var i int64
	err = e.Orm.Model(&data).Where("job_name = ?", c.JobName).Count(&i).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if i > 0 {
		err := errors.New("任务名称已存在！")
		e.Log.Errorf("db error: %s", err)
		return err
	}
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Update 修改SysJob对象
func (e *SysJob) Update(c *dto.SysJobUpdateReq, p *actions.DataPermission) error {
	var err error
	var model models.SysJob
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateSysJob error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	c.Generate(&model)
	update := e.Orm.Model(&model).Where("job_id = ?", &model.JobId).Updates(&model)
	if err = update.Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if update.RowsAffected == 0 {
		err = errors.New("update job error")
		log.Warnf("db update error")
		return err
	}
	return nil
}

// Remove 删除SysJob
func (e *SysJob) Remove(c *dto.SysJobById, p *actions.DataPermission) error {
	var err error
	var data models.SysJob

	db := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).Delete(&data, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Error found in RemoveSysJob: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

// RemoveBatch 批量删除SysJob
func (e *SysJob) RemoveBatch(c *dto.SysJobByIds, p *actions.DataPermission) error {
	var err error
	var data models.SysJob

	db := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).Delete(&data, c.GetIds())
	if err = db.Error; err != nil {
		e.Log.Errorf("Error found in RemoveBatchSysJob: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

// Start 启动定时任务
func (e *SysJob) Start(c *dto.SysJobById, p *actions.DataPermission) error {
	var err error
	var model models.SysJob

	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service StartSysJob error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权操作该数据")
	}

	// 更新状态为启动
	err = e.Orm.Model(&model).Where("job_id = ?", model.JobId).Update("status", 1).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}

	// TODO: 调用定时任务调度器启动任务
	// 这里需要根据具体的定时任务调度框架实现
	// 例如：使用 cron 或 quartz 等

	return nil
}

// RemoveJob 停止/移除定时任务
func (e *SysJob) RemoveJob(c *dto.SysJobById, p *actions.DataPermission) error {
	var err error
	var model models.SysJob

	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service RemoveJob error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权操作该数据")
	}

	// 更新状态为停止
	err = e.Orm.Model(&model).Where("job_id = ?", model.JobId).Update("status", 0).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}

	// TODO: 调用定时任务调度器停止任务
	// 这里需要根据具体的定时任务调度框架实现
	// 例如：使用 cron 或 quartz 等

	return nil
}

// UpdateStatus 更新任务状态
func (e *SysJob) UpdateStatus(c *dto.UpdateSysJobStatusReq, p *actions.DataPermission) error {
	var err error
	var model models.SysJob
	db := e.Orm.Scopes(
		actions.Permission(model.TableName(), p),
	).First(&model, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Service UpdateSysJobStatus error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	err = e.Orm.Table(model.TableName()).Where("job_id = ?", c.JobId).Updates(c).Error
	if err != nil {
		e.Log.Errorf("Service UpdateSysJobStatus error: %s", err)
		return err
	}
	return nil
}
