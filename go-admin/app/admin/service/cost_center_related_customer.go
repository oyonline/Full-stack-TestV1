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

type CostCenterRelatedCustomer struct {
	service.Service
}

// GetPage 获取CostCenterRelatedCustomer列表
func (e *CostCenterRelatedCustomer) GetPage(c *dto.CostCenterRelatedCustomerGetPageReq, p *actions.DataPermission, list *[]models.CostCenterRelatedCustomer, count *int64) error {
	var err error
	var data models.CostCenterRelatedCustomer

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("CostCenterRelatedCustomerService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取CostCenterRelatedCustomer对象
func (e *CostCenterRelatedCustomer) Get(d *dto.CostCenterRelatedCustomerGetReq, p *actions.DataPermission, model *models.CostCenterRelatedCustomer) error {
	var data models.CostCenterRelatedCustomer

	err := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetCostCenterRelatedCustomer error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建CostCenterRelatedCustomer对象
func (e *CostCenterRelatedCustomer) Insert(c *dto.CostCenterRelatedCustomerInsertReq) error {
	var err error
	var data models.CostCenterRelatedCustomer
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("CostCenterRelatedCustomerService Insert error:%s \r\n", err)
		return err
	}
	return nil
}

// Update 修改CostCenterRelatedCustomer对象
func (e *CostCenterRelatedCustomer) Update(c *dto.CostCenterRelatedCustomerUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.CostCenterRelatedCustomer{}
	e.Orm.Scopes(
		actions.Permission(data.TableName(), p),
	).First(&data, c.GetId())
	c.Generate(&data)

	db := e.Orm.Save(&data)
	if err = db.Error; err != nil {
		e.Log.Errorf("CostCenterRelatedCustomerService Save error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 删除CostCenterRelatedCustomer
func (e *CostCenterRelatedCustomer) Remove(d *dto.CostCenterRelatedCustomerDeleteReq, p *actions.DataPermission) error {
	var data models.CostCenterRelatedCustomer

	db := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).Delete(&data, d.GetId())
	if err := db.Error; err != nil {
		e.Log.Errorf("Service RemoveCostCenterRelatedCustomer error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

func (e *CostCenterRelatedCustomer) ImportData(req *dto.CostCenterRelatedCustomerImportReq, p *actions.DataPermission, userId int) error {
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
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterRelatedCustomer](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	var data11 models.CostCenterRelatedCustomer
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

func (e *CostCenterRelatedCustomer) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.CostCenterRelatedCustomer
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterRelatedCustomer](dictMappings)
	// 执行流式导出
	filename := "成本中心关联客户分组模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *CostCenterRelatedCustomer) Export(httpResponseWriter http.ResponseWriter, c *dto.CostCenterRelatedCustomerGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.CostCenterRelatedCustomer
	var templateData []dto.CostCenterRelatedCustomerExport
	err = e.Orm.Table(data.TableName()).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("CostCenterRelatedCustomerService Export error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostCenterRelatedCustomerExport](dictMappings)
	// 执行流式导出
	filename := "成本中心关联客户分组"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}
