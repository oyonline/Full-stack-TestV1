package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/baseservice"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/excelUtils"
	"go-admin/common/utils/kingdeeUtils"
)

// KingdeeCustomer 嵌入 BaseService[models.KingdeeCustomer] 提供 GetPage/Get/Insert/Update/Remove。
// 模板下载、Excel 导入导出、金蝶拉取四个方法是 KD 集成专属逻辑，保留自定义实现。
type KingdeeCustomer struct {
	baseservice.BaseService[models.KingdeeCustomer]
}

func (e *KingdeeCustomer) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	var templateData []models.KingdeeCustomer
	var dictMappings map[string]map[string]string
	mapper := excelUtils.NewExcelSingleStructureMapper[models.KingdeeCustomer](dictMappings)
	filename := "金蝶店铺模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *KingdeeCustomer) ImportData(req *dto.KingdeeCustomerImport, userId int) error {
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	var dictMappings map[string]map[string]string
	mapper := excelUtils.NewExcelSingleStructureMapper[models.KingdeeCustomer](dictMappings)
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	tx := e.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for i := range detailData {
		detailData[i].UseOrgId = 100076
		detailData[i].UpdateBy = userId
	}
	if err = tx.Save(detailData).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Export 导出 KingdeeCustomer。
func (e *KingdeeCustomer) Export(httpResponseWriter http.ResponseWriter, c *dto.KingdeeCustomerPageReq) error {
	var templateData []dto.KingdeeCustomerExport
	err := e.Orm.Select("customer_number, customer_name, group_name, b.fname AS use_org_name, forbid_status, country, remark").
		Table("kingdee_customer").
		Joins("LEFT JOIN kingdee_organize_info AS b ON use_org_id = b.forgid").
		Scopes(cDto.MakeCondition(c.GetNeedSearch())).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("Export GetList error:%s \r\n", err)
		return err
	}
	var dictMappings map[string]map[string]string
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.KingdeeCustomerExport](dictMappings)
	filename := "店铺信息"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

// PullKingdeeCustomers 拉取 KingdeeCustomer 对象。
func (e *KingdeeCustomer) PullKingdeeCustomers(createBy *int) error {
	var customer models.KingdeeCustomer
	e.Orm.Order("modify_date DESC").Take(&customer)

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
	if err := json.Unmarshal([]byte(respStr), &listPull); err != nil {
		e.Log.Errorf("JsonUnmarshal error:%s", err)
		return err
	}

	list := make([]models.KingdeeCustomer, 0, len(listPull))
	for i := range listPull {
		c := models.KingdeeCustomer{}
		c.CustId = listPull[i].CustId
		c.CustomerName = listPull[i].CustomerName
		c.CustomerNumber = listPull[i].CustomerNumber
		c.Country = listPull[i].Country
		c.CreateOrgId = listPull[i].CreateOrgId
		c.UseOrgId = listPull[i].UseOrgId
		c.GroupId = listPull[i].GroupId
		c.GroupName = listPull[i].GroupName
		c.GroupNumber = listPull[i].GroupNumber
		c.CustomerStatus = listPull[i].CustomerStatus
		c.ForbidStatus = listPull[i].ForbidStatus
		c.CreateDate = listPull[i].CreateDate
		c.ModifyDate = listPull[i].ModifyDate
		c.CreateBy = *createBy
		c.UpdateBy = *createBy
		list = append(list, c)
	}

	if err := e.Orm.Omit("remark", "dept_id", "cost_id").Save(&list).Error; err != nil {
		e.Log.Errorf("CreateInBatches error:%s", err)
		return err
	}
	return nil
}
