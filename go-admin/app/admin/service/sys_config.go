package service

import (
	"encoding/json"
	"errors"
	"strings"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type SysConfig struct {
	service.Service
}

// GetPage 获取SysConfig列表
func (e *SysConfig) GetPage(c *dto.SysConfigGetPageReq, list *[]models.SysConfig, count *int64) error {
	err := e.Orm.
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("Service GetSysConfigPage error:%s", err)
		return err
	}
	return nil
}

// Get 获取SysConfig对象
func (e *SysConfig) Get(d *dto.SysConfigGetReq, model *models.SysConfig) error {
	err := e.Orm.
		FirstOrInit(model, d.GetId()).
		Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		_ = e.AddError(err)
		return err
	}
	if model.Id == 0 {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetSysApi error: %s", err)
		_ = e.AddError(err)
		return err
	}
	return nil
}

// Insert 创建SysConfig对象
func (e *SysConfig) Insert(c *dto.SysConfigControl) error {
	var err error
	if isProtectedConfigKey(c.ConfigKey) {
		return errors.New("关键系统配置请在参数设置中维护")
	}
	var data models.SysConfig
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("Service InsertSysConfig error:%s", err)
		return err
	}
	return nil
}

// Update 修改SysConfig对象
func (e *SysConfig) Update(c *dto.SysConfigControl) error {
	var err error
	var model = models.SysConfig{}
	e.Orm.First(&model, c.GetId())
	if model.Id == 0 {
		return errors.New("查看对象不存在或无权查看")
	}
	if isProtectedConfigKey(model.ConfigKey) || isProtectedConfigKey(c.ConfigKey) {
		return errors.New("关键系统配置请在参数设置中维护")
	}
	c.Generate(&model)
	db := e.Orm.Save(&model)
	err = db.Error
	if err != nil {
		e.Log.Errorf("Service UpdateSysConfig error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")

	}
	return nil
}

// SetSysConfig 修改SysConfig对象
func (e *SysConfig) SetSysConfig(c *[]dto.GetSetSysConfigReq) error {
	var err error
	for _, req := range *c {
		var model = models.SysConfig{}
		e.Orm.Where("config_key = ?", req.ConfigKey).First(&model)
		if model.Id != 0 {
			req.Generate(&model)
			db := e.Orm.Save(&model)
			err = db.Error
			if err != nil {
				e.Log.Errorf("Service SetSysConfig error:%s", err)
				return err
			}
			if db.RowsAffected == 0 {
				return errors.New("无权更新该数据")
			}
		}
	}
	return nil
}

func (e *SysConfig) GetForSet(payload *dto.SystemSettingsPayload) error {
	settings := defaultSystemSettingsPayload()

	if appName, err := e.getSystemConfigValue(systemAppNameKey); err == nil {
		settings.Branding.AppName = appName
	}
	if appLogo, err := e.getSystemConfigValue(systemAppLogoKey); err == nil {
		settings.Branding.AppLogo = appLogo
	}
	if encoded, err := e.getSystemConfigValue(systemUIPreferencesKey); err == nil && encoded != "" {
		uiPreferences, decodeErr := decodeSystemUiPreferences(encoded)
		if decodeErr != nil {
			e.Log.Errorf("Service decodeSystemUiPreferences error:%s", decodeErr)
			return decodeErr
		}
		settings.UIPreferences = uiPreferences
	}

	*payload = settings
	return nil
}

func (e *SysConfig) UpdateForSet(payload *dto.SystemSettingsPayload) error {
	settings := sanitizeSystemSettingsPayload(*payload)
	encoded, err := encodeSystemUiPreferences(settings.UIPreferences)
	if err != nil {
		e.Log.Errorf("Service encodeSystemUiPreferences error:%s", err)
		return err
	}

	if !json.Valid([]byte(encoded)) {
		return errors.New("界面设置格式非法")
	}

	if err := e.upsertSystemConfigValue(
		systemAppNameKey,
		"系统名称",
		settings.Branding.AppName,
		"1",
		"",
	); err != nil {
		e.Log.Errorf("Service UpdateForSet appName error:%s", err)
		return err
	}

	if err := e.upsertSystemConfigValue(
		systemAppLogoKey,
		"系统logo",
		settings.Branding.AppLogo,
		"1",
		"",
	); err != nil {
		e.Log.Errorf("Service UpdateForSet appLogo error:%s", err)
		return err
	}

	if err := e.upsertSystemConfigValue(
		systemUIPreferencesKey,
		systemUIPreferencesConfigName,
		encoded,
		systemUIPreferencesIsFrontend,
		systemUIPreferencesConfigRemark,
	); err != nil {
		e.Log.Errorf("Service UpdateForSet uiPreferences error:%s", err)
		return err
	}

	return nil
}

// Remove 删除SysConfig
func (e *SysConfig) Remove(d *dto.SysConfigDeleteReq) error {
	var err error
	var data models.SysConfig
	records := make([]models.SysConfig, 0)

	if err = e.Orm.Where("id IN ?", d.Ids).Find(&records).Error; err != nil {
		e.Log.Errorf("Error found in RemoveSysConfig preload: %s", err)
		return err
	}
	for _, record := range records {
		if isProtectedConfigKey(record.ConfigKey) {
			return errors.New("关键系统配置不允许在参数管理中删除")
		}
	}

	db := e.Orm.Delete(&data, d.Ids)
	if err = db.Error; err != nil {
		e.Log.Errorf("Service RemoveSysConfig error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("无权删除该数据")
		return err
	}
	return nil
}

// GetWithKey 根据Key获取SysConfig
func (e *SysConfig) GetWithKey(c *dto.SysConfigByKeyReq, resp *dto.GetSysConfigByKEYForServiceResp) error {
	var err error
	var data models.SysConfig
	err = e.Orm.Table(data.TableName()).Where("config_key = ?", c.ConfigKey).First(resp).Error
	if err != nil {
		e.Log.Errorf("At Service GetSysConfigByKEY Error:%s", err)
		return err
	}

	return nil
}

func (e *SysConfig) GetWithKeyList(c *dto.SysConfigGetToSysAppReq, list *[]models.SysConfig) error {
	err := e.Orm.
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
		).
		Find(list).Error
	if err != nil {
		e.Log.Errorf("Service GetSysConfigByKey error:%s", err)
		return err
	}
	return nil
}

func (e *SysConfig) ValidateSystemSettingsPayload(payload *dto.SystemSettingsPayload) error {
	settings := sanitizeSystemSettingsPayload(*payload)
	if settings.Branding.AppName == "" {
		return errors.New("系统名称不能为空")
	}
	if settings.Branding.AppLogo != "" &&
		!strings.HasPrefix(settings.Branding.AppLogo, "http://") &&
		!strings.HasPrefix(settings.Branding.AppLogo, "https://") {
		return errors.New("系统 Logo 仅支持 http/https 图片地址")
	}
	encoded, err := encodeSystemUiPreferences(settings.UIPreferences)
	if err != nil {
		return err
	}
	if !json.Valid([]byte(encoded)) {
		return errors.New("界面设置格式非法")
	}
	return nil
}
