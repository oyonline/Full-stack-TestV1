package service

import (
	"errors"
	"fmt"
	"github.com/go-admin-team/go-admin-core/logger"
	"go-admin/common/utils/collectors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/excelUtils"
)

type BudgetFeeCategoryDetails struct {
	service.Service
}

// GetPage 获取BudgetFeeCategoryDetails列表
func (e *BudgetFeeCategoryDetails) GetPage(c *dto.BudgetFeeCategoryDetailsGetPageReq, p *actions.DataPermission, list *[]models.BudgetFeeCategoryDetails, count *int64) error {
	var err error
	var data models.BudgetFeeCategoryDetails
	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("BudgetFeeCategoryDetailsService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取BudgetFeeCategoryDetails对象
func (e *BudgetFeeCategoryDetails) Get(d *dto.BudgetFeeCategoryDetailsGetReq, p *actions.DataPermission, model *models.BudgetFeeCategoryDetails) error {
	var data models.BudgetFeeCategoryDetails

	err := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetBudgetFeeCategoryDetails error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建BudgetFeeCategoryDetails对象
func (e *BudgetFeeCategoryDetails) Insert(c *dto.BudgetFeeCategoryDetailsInsertReq, p *actions.DataPermission) error {
	var err error
	var data models.BudgetFeeCategoryDetails
	c.Generate(&data)
	// 验证数据
	if err = CheckCodeExists(e, &data); err != nil {
		return err
	}
	if err = e.Orm.Scopes(actions.Permission(data.TableName(), p)).Create(&data).Error; err != nil {
		e.Log.Errorf("新增失败:%s", err)
		return err
	}
	return nil
}

// Update 修改BudgetFeeCategoryDetails对象
func (e *BudgetFeeCategoryDetails) Update(c *dto.BudgetFeeCategoryDetailsUpdateReq, p *actions.DataPermission) error {
	var data = models.BudgetFeeCategoryDetails{}
	if err := e.Orm.Scopes(actions.Permission(data.TableName(), p)).First(&data, c.GetId()).Error; err != nil {
		return err
	}
	c.Generate(&data)
	// 验证数据
	if err1 := CheckCodeExists(e, &data); err1 != nil {
		return err1
	}
	if err := e.Orm.Save(&data).Error; err != nil {
		e.Log.Errorf("修改失败:%s", err)
		return err
	}
	return nil
}

// Remove 删除BudgetFeeCategoryDetails
func (e *BudgetFeeCategoryDetails) Remove(d *dto.BudgetFeeCategoryDetailsDeleteReq, p *actions.DataPermission) error {
	var data models.BudgetFeeCategoryDetails
	if err := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Delete(&data, d.GetId()).Error; err != nil {
		e.Log.Errorf("删除失败:%s", err)
		return err
	}
	return nil
}

// CheckCodeExists 检查费用编码是否存在
func CheckCodeExists(e *BudgetFeeCategoryDetails, data *models.BudgetFeeCategoryDetails) error {
	var count int64
	query := e.Orm.Model(&models.BudgetFeeCategoryDetails{}).Where("fee_code = ? and view_type = ?", data.FeeCode, data.ViewType)
	// 如果是修改操作，需要排除当前记录
	if data.Id > 0 {
		query = query.Where("id != ?", data.Id)
	}
	err := query.Count(&count).Error
	if err != nil {
		e.Log.Errorf("CheckCodeExists error:%s \r\n", err)
		return err
	}
	if count > 0 {
		return errors.New(fmt.Sprintf("费用编码【%s】已存在", data.FeeCode))
	}
	return nil
}

func (e *BudgetFeeCategoryDetails) ImportData(req *dto.BudgetFeeCategoryDetailsImportReq, p *actions.DataPermission, userId int) error {
	// 打开上传的文件
	file, fileErr := req.File.Open()
	if fileErr != nil {
		return fileErr
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	// 初始化字典映射
	var dictMappings map[string]map[string]string
	// 导入Excel数据
	detailData, importErr := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategoryDetails](dictMappings).ImportFromExcel(file, "")
	if importErr != nil {
		return importErr
	}
	// 开始事务
	tx := e.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 获取费用类别
	feeCategoryMap, feeCategoryErr := GetFeeCategoryMap(tx)
	if feeCategoryErr != nil {
		tx.Rollback()
		return feeCategoryErr
	}
	//获取客户分组
	customerGroupMap, customerGroupErr := GetCustomerGroupMap(tx, e.Log)
	if customerGroupErr != nil {
		tx.Rollback()
		return customerGroupErr
	}
	// 批量更新
	for i := range detailData {
		detailData[i].CreateBy = userId
		detailData[i].UpdateBy = userId
		feeCategoryKey := detailData[i].PathStr + strconv.Itoa(detailData[i].ViewType)
		if feeCategoryVO, ok := feeCategoryMap[feeCategoryKey]; ok {
			detailData[i].BudgetFeeCategoryId = feeCategoryVO.Id
		}
		if detailData[i].Platform != "" {
			for _, customerGroup := range strings.Split(detailData[i].Platform, ",") {
				customerGroupName := strings.TrimSpace(customerGroup)
				if _, ok := customerGroupMap[customerGroupName]; !ok {
					return errors.New(fmt.Sprintf("客户分组【%s】不存在", customerGroup))
				}
			}
			detailData[i].Platform = "/" + strings.ReplaceAll(detailData[i].Platform, ",", "/") + "/"
		}
	}
	if err := tx.Scopes(actions.Permission(models.BudgetFeeCategoryDetails{}.TableName(), p)).Save(detailData).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (e *BudgetFeeCategoryDetails) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.BudgetFeeCategoryDetails
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategoryDetails](dictMappings)
	// 执行流式导出
	filename := "预算费用明细模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *BudgetFeeCategoryDetails) Export(httpResponseWriter http.ResponseWriter, c *dto.BudgetFeeCategoryDetailsGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.BudgetFeeCategoryDetails
	var templateData []models.BudgetFeeCategoryDetails
	err = e.Orm.Table(data.TableName()).
		Joins("LEFT JOIN budget_fee_category on budget_fee_category.id = budget_fee_category_details.budget_fee_category_id").
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).Select("budget_fee_category_details.*, budget_fee_category.category_name, budget_fee_category.path_str").
		Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("BudgetFeeCategoryDetailsService GetPage error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategoryDetails](dictMappings)
	// 执行流式导出
	filename := "预算费用明细数据"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

func GetCustomerGroupMap(tx *gorm.DB, log *logger.Helper) (map[string]models.KingdeeCustomerGroup, error) {
	//查询费用类别ID
	customerGroupService := KingdeeCustomerGroup{Service: service.Service{Orm: tx, Log: log}}
	var customerGroupList []models.KingdeeCustomerGroup
	if err := customerGroupService.GetKingdeeCustomerGroups(&customerGroupList); err != nil {
		return nil, err
	}
	customerGroupMap := collectors.ToMapDynamic(customerGroupList, func(item models.KingdeeCustomerGroup) string {
		return item.GroupName
	}, func(item models.KingdeeCustomerGroup) models.KingdeeCustomerGroup {
		return item
	})
	return customerGroupMap, nil
}
