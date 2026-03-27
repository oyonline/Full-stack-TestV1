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

type CostBudgetVersionDetail struct {
	service.Service
}

// GetPage 获取CostBudgetVersionDetail列表
func (e *CostBudgetVersionDetail) GetPage(c *dto.CostBudgetVersionDetailGetPageReq, p *actions.DataPermission, list *[]models.CostBudgetVersionDetail, count *int64) error {
	var err error
	var data models.CostBudgetVersionDetail

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("CostBudgetVersionDetailService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取CostBudgetVersionDetail对象
func (e *CostBudgetVersionDetail) Get(d *dto.CostBudgetVersionDetailGetReq, p *actions.DataPermission, model *models.CostBudgetVersionDetail) error {
	var data models.CostBudgetVersionDetail

	err := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetCostBudgetVersionDetail error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建CostBudgetVersionDetail对象
func (e *CostBudgetVersionDetail) Insert(c *dto.CostBudgetVersionDetailInsertReq) error {
	var err error
	var data models.CostBudgetVersionDetail
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("CostBudgetVersionDetailService Insert error:%s \r\n", err)
		return err
	}
	return nil
}

// Update 修改CostBudgetVersionDetail对象
func (e *CostBudgetVersionDetail) Update(c *dto.CostBudgetVersionDetailUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.CostBudgetVersionDetail{}
	e.Orm.Scopes(
		actions.Permission(data.TableName(), p),
	).First(&data, c.GetId())
	c.Generate(&data)

	db := e.Orm.Save(&data)
	if err = db.Error; err != nil {
		e.Log.Errorf("CostBudgetVersionDetailService Save error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 删除CostBudgetVersionDetail
func (e *CostBudgetVersionDetail) Remove(d *dto.CostBudgetVersionDetailDeleteReq, p *actions.DataPermission) error {
	var data models.CostBudgetVersionDetail

	db := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).Delete(&data, d.GetId())
	if err := db.Error; err != nil {
		e.Log.Errorf("Service RemoveCostBudgetVersionDetail error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

func (e *CostBudgetVersionDetail) ImportData(req *dto.CostBudgetVersionDetailImportReq, p *actions.DataPermission, userId int) error {
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
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostBudgetVersionDetail](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	var data11 models.CostBudgetVersionDetail
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

func (e *CostBudgetVersionDetail) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []dto.CostBudgetVersionDetailExport
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostBudgetVersionDetailExport](dictMappings)
	// 执行流式导出
	filename := "预算版本管理详情模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *CostBudgetVersionDetail) Export(httpResponseWriter http.ResponseWriter, c *dto.CostBudgetVersionDetailGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.CostBudgetVersionDetail
	var templateData []dto.CostBudgetVersionDetailExport
	err = e.Orm.Table(data.TableName()).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("CostBudgetVersionDetailService Export error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostBudgetVersionDetailExport](dictMappings)
	// 执行流式导出
	filename := "预算版本管理详情"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}
