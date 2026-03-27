package service

import (
	"errors"
	"strings"

	"github.com/go-admin-team/go-admin-core/sdk/service"

	"go-admin/app/platform/models"
	"go-admin/app/platform/service/dto"
)

type ModuleRegistry struct {
	service.Service
}

func (e *ModuleRegistry) GetPage(c *dto.ModuleRegistryGetPageReq, list *[]models.ModuleRegistry, count *int64) error {
	db := e.Orm.Model(&models.ModuleRegistry{})
	if c.ModuleKey != "" {
		db = db.Where("module_key LIKE ?", "%"+strings.TrimSpace(c.ModuleKey)+"%")
	}
	if c.ModuleName != "" {
		db = db.Where("module_name LIKE ?", "%"+strings.TrimSpace(c.ModuleName)+"%")
	}
	if c.Status != "" {
		db = db.Where("status = ?", c.Status)
	}
	if c.MenuRootCode != "" {
		db = db.Where("menu_root_code LIKE ?", "%"+strings.TrimSpace(c.MenuRootCode)+"%")
	}
	return db.Order("sort ASC, module_id ASC").
		Offset((c.GetPageIndex()-1)*c.GetPageSize()).
		Limit(c.GetPageSize()).
		Find(list).Limit(-1).Offset(-1).Count(count).Error
}

func (e *ModuleRegistry) Get(id int, model *models.ModuleRegistry) error {
	return e.Orm.First(model, id).Error
}

func (e *ModuleRegistry) GetByKey(moduleKey string, model *models.ModuleRegistry) error {
	return e.Orm.Where("module_key = ?", moduleKey).First(model).Error
}

func (e *ModuleRegistry) Insert(c *dto.ModuleRegistryInsertReq) error {
	c.Normalize()
	var count int64
	if err := e.Orm.Model(&models.ModuleRegistry{}).Where("module_key = ? OR route_base = ? OR menu_root_code = ?", c.ModuleKey, c.RouteBase, c.MenuRootCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("模块编码、路由前缀或根菜单编码已存在")
	}
	var model models.ModuleRegistry
	c.Generate(&model)
	return e.Orm.Create(&model).Error
}

func (e *ModuleRegistry) Update(c *dto.ModuleRegistryUpdateReq) error {
	c.Normalize()
	var count int64
	if err := e.Orm.Model(&models.ModuleRegistry{}).
		Where("(module_key = ? OR route_base = ? OR menu_root_code = ?) AND module_id <> ?", c.ModuleKey, c.RouteBase, c.MenuRootCode, c.ModuleId).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("模块编码、路由前缀或根菜单编码已存在")
	}
	var model models.ModuleRegistry
	if err := e.Orm.First(&model, c.ModuleId).Error; err != nil {
		return err
	}
	c.Generate(&model)
	return e.Orm.Save(&model).Error
}

func (e *ModuleRegistry) Delete(id int) error {
	return e.Orm.Delete(&models.ModuleRegistry{}, id).Error
}
