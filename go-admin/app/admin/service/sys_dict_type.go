package service

import (
	"fmt"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/baseservice"
	cDto "go-admin/common/dto"
)

// SysDictType 嵌入 BaseService[models.SysDictType] 提供 GetPage/Get/Update/Remove。
// Insert 因含唯一性校验保留自定义；GetAll 是不分页全量查询，BaseService 不内置。
type SysDictType struct {
	baseservice.BaseService[models.SysDictType]
}

// Insert 创建对象（含 dict_type 唯一性校验）。
func (e *SysDictType) Insert(c *dto.SysDictTypeInsertReq) error {
	var data models.SysDictType
	c.Generate(&data)
	var count int64
	e.Orm.Model(&data).Where("dict_type = ?", data.DictType).Count(&count)
	if count > 0 {
		return fmt.Errorf("当前字典类型[%s]已经存在！", data.DictType)
	}
	if err := e.Orm.Create(&data).Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// GetAll 获取所有（不分页，仅依据 search 条件过滤）。
func (e *SysDictType) GetAll(c *dto.SysDictTypeGetPageReq, list *[]models.SysDictType) error {
	var data models.SysDictType
	if err := e.Orm.Model(&data).
		Scopes(cDto.MakeCondition(c.GetNeedSearch())).
		Find(list).Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}
