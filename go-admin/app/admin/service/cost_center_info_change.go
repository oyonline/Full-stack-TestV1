package service

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/excelUtils"
)

type CostCenterInfoChange struct {
	service.Service
}

// GetPage 获取CostCenterInfoChange列表
func (e *CostCenterInfoChange) GetPage(c *dto.CostCenterInfoChangeGetPageReq, p *actions.DataPermission, list *[]models.CostCenterInfoChange, count *int64) error {
	var err error
	var data models.CostCenterInfoChange

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("CostCenterInfoChangeService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取CostCenterInfoChange对象
func (e *CostCenterInfoChange) Get(d *dto.CostCenterInfoChangeGetReq, p *actions.DataPermission, model *models.CostCenterInfoChange) error {
	var data models.CostCenterInfoChange

	err := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetCostCenterInfoChange error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建CostCenterInfoChange对象
func (e *CostCenterInfoChange) Insert(c *dto.CostCenterInfoChangeInsertReq) error {
	var err error
	var data models.CostCenterInfoChange
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("CostCenterInfoChangeService Insert error:%s \r\n", err)
		return err
	}
	return nil
}

// Update 修改CostCenterInfoChange对象
func (e *CostCenterInfoChange) Update(c *dto.CostCenterInfoChangeUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.CostCenterInfoChange{}
	e.Orm.Scopes(
		actions.Permission(data.TableName(), p),
	).First(&data, c.GetId())
	c.Generate(&data)

	db := e.Orm.Save(&data)
	if err = db.Error; err != nil {
		e.Log.Errorf("CostCenterInfoChangeService Save error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 删除CostCenterInfoChange
func (e *CostCenterInfoChange) Remove(d *dto.CostCenterInfoChangeDeleteReq, p *actions.DataPermission) error {
	var data models.CostCenterInfoChange

	db := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).Delete(&data, d.GetId())
	if err := db.Error; err != nil {
		e.Log.Errorf("Service RemoveCostCenterInfoChange error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

func (e *CostCenterInfoChange) ImportData(req *dto.CostCenterInfoChangeImportReq, p *actions.DataPermission, userId int) error {
	var err error
	// 打开上传的文件
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	// 初始化字典映射
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterInfoChange](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	var data11 models.CostCenterInfoChange
	// 开始事务
	tx := e.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 批量更新
	for i := range detailData {
		detailData[i].CreateBy = userId
		detailData[i].UpdateBy = userId
	}
	if err := tx.Scopes(actions.Permission(data11.TableName(), p)).Save(detailData).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (e *CostCenterInfoChange) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.CostCenterInfoChange
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterInfoChange](dictMappings)
	// 执行流式导出
	filename := "成本中心变更数据模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *CostCenterInfoChange) Export(httpResponseWriter http.ResponseWriter, c *dto.CostCenterInfoChangeGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.CostCenterInfoChange
	var templateData []dto.CostCenterInfoChangeExport
	err = e.Orm.Table(data.TableName()).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("CostCenterInfoChangeService Export error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostCenterInfoChangeExport](dictMappings)
	// 执行流式导出
	filename := "成本中心变更数据"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}
