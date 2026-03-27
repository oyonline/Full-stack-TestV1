package service

import (
	"errors"
	"fmt"
	"go-admin/common/utils/excelUtils"
	"mime/multipart"
	"net/http"

	"encoding/json"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/kingdeeUtils"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"
)

type KingdeeCustomer struct {
	service.Service
}

// GetPage 获取KingdeeCustomer列表
func (e *KingdeeCustomer) GetPage(c *dto.KingdeeCustomerPageReq, list *[]models.KingdeeCustomer, count *int64) error {
	var err error
	var data models.KingdeeCustomer

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("db error:%s \r", err)
		return err
	}
	return nil
}

// Get 获取KingdeeCustomer对象
func (e *KingdeeCustomer) Get(d *dto.KingdeeCustomerGetReq, model *models.KingdeeCustomer) error {
	var err error
	var data models.KingdeeCustomer

	db := e.Orm.Model(&data).
		First(model, d.GetId())
	err = db.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建KingdeeCustomer对象
func (e *KingdeeCustomer) Insert(c *dto.KingdeeCustomerInsertReq) error {
	var err error
	var data models.KingdeeCustomer
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Update 修改KingdeeCustomer对象
func (e *KingdeeCustomer) Update(c *dto.KingdeeCustomerUpdateReq) error {
	var err error
	var model = models.KingdeeCustomer{}
	e.Orm.First(&model, c.GetId())
	c.Generate(&model)

	db := e.Orm.Save(&model)
	if err = db.Error; err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")

	}
	return nil
}

// Remove 删除KingdeeCustomer
func (e *KingdeeCustomer) Remove(d *dto.KingdeeCustomerDeleteReq) error {
	var err error
	var data models.KingdeeCustomer

	db := e.Orm.Model(&data).Delete(&data, d.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Delete error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("无权删除该数据")
		return err
	}
	return nil
}

func (e *KingdeeCustomer) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.KingdeeCustomer
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.KingdeeCustomer](dictMappings)
	// 执行流式导出
	filename := "金蝶店铺模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *KingdeeCustomer) ImportData(req *dto.KingdeeCustomerImport, userId int) error {
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
	mapper := excelUtils.NewExcelSingleStructureMapper[models.KingdeeCustomer](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	// 开始事务
	tx := e.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 批量更新
	for i := range detailData {
		// 查询使用组织ID
		detailData[i].UseOrgId = 100076

		// 根据店铺名称和使用组织ID更新备注
		detailData[i].UpdateBy = userId
	}
	err = tx.Save(detailData).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Export 导出KingdeeCustomer
func (e *KingdeeCustomer) Export(httpResponseWriter http.ResponseWriter, c *dto.KingdeeCustomerPageReq) error {
	var err error
	var templateData []dto.KingdeeCustomerExport
	err = e.Orm.Select("customer_number, customer_name, group_name, b.fname AS use_org_name, forbid_status, country, remark").
		Table("kingdee_customer").
		Joins("LEFT JOIN kingdee_organize_info AS b ON use_org_id = b.forgid").
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
		).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("Export GetList error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.KingdeeCustomerExport](dictMappings)
	// 执行流式导出
	filename := "店铺信息"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

// PullKingdeeCustomers 拉取KingdeeCustomer对象
func (e *KingdeeCustomer) PullKingdeeCustomers(createBy *int) error {
	var err error
	var customer models.KingdeeCustomer
	e.Orm.Order("modify_date DESC").Take(&customer)

	// 拉取金蝶客户信息
	postData := map[string]string{
		"FormId":       "BD_Customer",
		"FieldKeys":    "FCustId,FName,FNumber,FCountry.FDataValue,FCreateOrgId,FUseOrgId,FGroup,FGroup.FName,FGroup.FNumber,FDocumentStatus,FForbidStatus,FCreateDate,FModifyDate",
		"FilterString": "",
		"StartRow":     "0",
		"Limit":        "0",
	}
	if customer.ModifyDate != "" {
		postData["FilterString"] = fmt.Sprintf("FModifyDate > '%s'", customer.ModifyDate)
	}
	respStr := kingdeeUtils.BillQuery(postData)
	if respStr == "[]" {
		return nil
	}

	listPull := make([]dto.KingdeeCustomerPull, 0)
	err = json.Unmarshal([]byte(respStr), &listPull)
	if err != nil {
		e.Log.Errorf("JsonUnmarshal error:%s", err)
		return err
	}

	list := make([]models.KingdeeCustomer, 0)
	for i := range listPull {
		customer := models.KingdeeCustomer{}
		customer.CustId = listPull[i].CustId
		customer.CustomerName = listPull[i].CustomerName
		customer.CustomerNumber = listPull[i].CustomerNumber
		customer.Country = listPull[i].Country
		customer.CreateOrgId = listPull[i].CreateOrgId
		customer.UseOrgId = listPull[i].UseOrgId
		customer.GroupId = listPull[i].GroupId
		customer.GroupName = listPull[i].GroupName
		customer.GroupNumber = listPull[i].GroupNumber
		customer.CustomerStatus = listPull[i].CustomerStatus
		customer.ForbidStatus = listPull[i].ForbidStatus
		customer.CreateDate = listPull[i].CreateDate
		customer.ModifyDate = listPull[i].ModifyDate
		customer.CreateBy = *createBy
		customer.UpdateBy = *createBy
		list = append(list, customer)
	}

	err = e.Orm.Omit("remark", "dept_id", "cost_id").Save(&list).Error
	if err != nil {
		e.Log.Errorf("CreateInBatches error:%s", err)
		return err
	}
	return nil
}
