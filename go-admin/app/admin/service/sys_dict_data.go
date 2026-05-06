package service

import (
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/baseservice"
	cDto "go-admin/common/dto"
)

// SysDictData 嵌入 BaseService[models.SysDictData] 提供 GetPage/Get/Insert/Update/Remove。
// GetAll 是非分页全量查询，BaseService 不内置，保留自定义实现。
type SysDictData struct {
	baseservice.BaseService[models.SysDictData]
}

// GetAll 获取所有（不分页，仅依据 search 条件过滤）。
func (e *SysDictData) GetAll(c *dto.SysDictDataGetPageReq, list *[]models.SysDictData) error {
	var data models.SysDictData
	if err := e.Orm.Model(&data).
		Scopes(cDto.MakeCondition(c.GetNeedSearch())).
		Find(list).Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}
