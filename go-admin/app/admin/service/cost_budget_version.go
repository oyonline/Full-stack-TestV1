package service

import (
	"errors"
	"fmt"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"go-admin/common/utils/collectors"
	"go-admin/common/utils/treeUtils"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/excelUtils"
)

type CostBudgetVersion struct {
	service.Service
}

// GetPage 获取CostBudgetVersion列表
func (e *CostBudgetVersion) GetPage(c *dto.CostBudgetVersionGetPageReq, p *actions.DataPermission, list *[]models.CostBudgetVersion, count *int64) error {
	var err error
	var data models.CostBudgetVersion

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("CostBudgetVersionService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取CostBudgetVersion对象
func (e *CostBudgetVersion) Get(d *dto.CostBudgetVersionGetReq, p *actions.DataPermission) ([]dto.GroupResult, error) {
	var data models.CostBudgetVersion
	if err := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).First(&data, d.GetId()).Error; err != nil {
		e.Log.Errorf("db error:%s", err)
		return nil, err
	}
	var centerInfo models.CostCenterInfo
	if err := e.Orm.Model(&centerInfo).First(&centerInfo, data.CostCenterInfoId).Error; err != nil {
		e.Log.Errorf("查询关联成本中心数据失败:%s", err)
		return nil, err
	}
	data.CostCenterName = centerInfo.CostCenterName
	data.CostCenterCode = centerInfo.CostCenterCode
	// 第一步：按版本分组并转换为月份数据
	monthDataMap := e.Step1GroupAndTransform(data.Id)
	// 第二步：构建父子结构树并统计
	treeDataMap := e.Step2BuildTree(monthDataMap)
	// 第三步：按版本分组生成最终结果
	groupResults := e.Step3GroupToResult(treeDataMap, data)
	return groupResults, nil
}

// Insert 创建CostBudgetVersion对象
func (e *CostBudgetVersion) Insert(c *dto.CostBudgetVersionInsertReq) error {
	var data models.CostBudgetVersion
	c.Generate(&data)
	data.CostBudgetCode = fmt.Sprintf("BAV-%d-%d", data.Years, data.CostCenterInfoId)
	if c.File == nil {
		return errors.New("请选择导入文件")
	}
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	file, err := c.File.Open()
	if err != nil {
		fmt.Println("打开文件失败: ", err.Error())
		tx.Rollback()
		return err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	totalAmount, versionDetailDataList, err := e.dealImportData(file, tx, data)
	if err != nil {
		tx.Rollback()
		return err
	}
	data.BudgetAmount = totalAmount
	if err = tx.Create(&data).Error; err != nil {
		fmt.Println("创建失败原因: ", err.Error())
		tx.Rollback()
		return err
	}
	for i := range versionDetailDataList {
		versionDetailDataList[i].CostBudgetVersionId = data.Id
	}
	if err = tx.Save(&versionDetailDataList).Error; err != nil {
		fmt.Println("预算版本明细创建失败: ", err.Error())
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Update 修改CostBudgetVersion对象
func (e *CostBudgetVersion) Update(c *dto.CostBudgetVersionUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.CostBudgetVersion{}
	e.Orm.Scopes(actions.Permission(data.TableName(), p)).First(&data, c.GetId())
	//c.Generate(&data)
	if data.Status > 1 {
		return errors.New("草稿状态的预算版本才能修改")
	}

	if c.File == nil {
		return errors.New("请选择导入文件")
	}
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	file, err := c.File.Open()
	if err != nil {
		fmt.Println("打开文件失败: ", err.Error())
		tx.Rollback()
		return err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	totalAmount, versionDetailDataList, err := e.dealImportData(file, tx, data)
	if err != nil {
		tx.Rollback()
		return err
	}
	data.BudgetAmount = totalAmount
	if err = tx.Save(&data).Error; err != nil {
		fmt.Println("预算版本编辑失败原因: ", err.Error())
		tx.Rollback()
		return err
	}
	for i := range versionDetailDataList {
		versionDetailDataList[i].CostBudgetVersionId = data.Id
	}
	if err = tx.Unscoped().Where("cost_budget_version_id = ?", data.Id).Delete(&models.CostBudgetVersionDetail{}).Error; err != nil {
		fmt.Println("预算版本明细删除失败: ", err.Error())
		tx.Rollback()
		return err
	}
	if err = tx.Save(&versionDetailDataList).Error; err != nil {
		fmt.Println("预算版本明细创建失败: ", err.Error())
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Remove 删除CostBudgetVersion
func (e *CostBudgetVersion) Remove(d *dto.CostBudgetVersionDeleteReq, p *actions.DataPermission) error {
	var dataList []models.CostBudgetVersion
	var data models.CostBudgetVersion

	e.Orm.Scopes(actions.Permission(data.TableName(), p)).Find(&dataList, d.GetId())
	for i := range dataList {
		if dataList[i].Status > 1 {
			return errors.New("草稿状态的预算版本才能删除")
		}
	}

	if err := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Delete(&data, d.GetId()).Error; err != nil {
		e.Log.Errorf("删除CostBudgetVersion: %s", err)
		return err
	}
	return nil
}

func (e *CostBudgetVersion) ImportData(req *dto.CostBudgetVersionImportReq, p *actions.DataPermission, userId int) error {
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
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostBudgetVersion](dictMappings)
	// 导入Excel数据
	detailData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return err
	}
	var data11 models.CostBudgetVersion
	// 开始事务
	tx := e.Orm.Debug().Begin()
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
	if err = tx.Scopes(actions.Permission(data11.TableName(), p)).Save(detailData).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (e *CostBudgetVersion) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.CostBudgetVersion
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostBudgetVersion](dictMappings)
	// 执行流式导出
	filename := "预算版本管理模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *CostBudgetVersion) Export(httpResponseWriter http.ResponseWriter, c *dto.CostBudgetVersionGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.CostBudgetVersion
	var templateData []dto.CostBudgetVersionExport
	err = e.Orm.Table(data.TableName()).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).Find(&templateData).Error
	if err != nil {
		e.Log.Errorf("CostBudgetVersionService Export error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostBudgetVersionExport](dictMappings)
	// 执行流式导出
	filename := "预算版本管理"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

func (e *CostBudgetVersion) dealImportData(file multipart.File, tx *gorm.DB, data models.CostBudgetVersion) (float64, []models.CostBudgetVersionDetail, error) {
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[dto.CostBudgetVersionDetailExport](dictMappings)
	// 导入Excel数据
	importData, err := mapper.ImportFromExcel(file, "")
	if err != nil {
		return 0, nil, err
	}
	if importData == nil || len(importData) == 0 {
		return 0, nil, errors.New("导入数据为空")
	}
	feeCategoryMap, costCenterMap, err2 := getMapByImportData(tx)
	if err2 != nil {
		return 0, nil, err2
	}
	var versionDetailDataList []models.CostBudgetVersionDetail
	var totalAmount float64
	for i := range importData {
		// 1. 验证费用类别
		if _, ok := feeCategoryMap[importData[i].CategoryName]; !ok {
			return 0, nil, fmt.Errorf("第%d行：费用类别【%s】不存在", i+2, importData[i].CategoryName)
		}
		// 2. 验证成本中心
		if _, ok := costCenterMap[importData[i].CostCenterName]; !ok {
			return 0, nil, fmt.Errorf("第%d行：成本中心【%s】不存在", i+2, importData[i].CostCenterName)
		}
		if data.CostCenterInfoId != costCenterMap[importData[i].CostCenterName].Id {
			return 0, nil, fmt.Errorf("第%d行：选择的成本中心和导入不一致【%s】", i+2, importData[i].CostCenterName)
		}
		totalAmount += importData[i].TotalAmount()
	}
	for i := range importData {
		for monthOffset := 1; monthOffset <= 12; monthOffset++ {
			var detailDataVO models.CostBudgetVersionDetail
			detailDataVO.BudgetFeeCategoryId = feeCategoryMap[importData[i].CategoryName].Id
			detailDataVO.BudgetAmount = importData[i].GetMonthAmount(monthOffset)
			detailDataVO.YearsMonth = fmt.Sprintf("%d-%02d", data.Years, monthOffset)
			detailDataVO.CostBudgetVersionId = data.Id
			detailDataVO.CreateBy = data.CreateBy
			detailDataVO.UpdateBy = data.CreateBy
			versionDetailDataList = append(versionDetailDataList, detailDataVO)
		}
	}
	return totalAmount, versionDetailDataList, nil
}

func getMapByImportData(tx *gorm.DB) (map[string]models.BudgetFeeCategory, map[string]models.CostCenterInfo, error) {
	//费用类别
	var feeCategoryList []models.BudgetFeeCategory
	if err := tx.Model(&models.BudgetFeeCategory{}).Find(&feeCategoryList).Error; err != nil {
		return nil, nil, err
	}
	feeCategoryMap := collectors.ToMapDynamic(feeCategoryList, func(v models.BudgetFeeCategory) string {
		return v.CategoryName
	}, func(v models.BudgetFeeCategory) models.BudgetFeeCategory { return v })
	var costCenterList []models.CostCenterInfo
	if err := tx.Model(&models.CostCenterInfo{}).Where("status = ?", 2).Find(&costCenterList).Error; err != nil {
		return feeCategoryMap, nil, err
	}
	costCenterMap := collectors.ToMapDynamic(costCenterList, func(v models.CostCenterInfo) string {
		return v.CostCenterName
	}, func(v models.CostCenterInfo) models.CostCenterInfo { return v })
	return feeCategoryMap, costCenterMap, nil
}

func (e *CostBudgetVersion) Step1GroupAndTransform(id int64) map[int64][]dto.MonthData {
	var resultMap = make(map[int64][]dto.MonthData)
	// 查询预算版本明细
	querySql := `SELECT
	t1.id,
	t1.parent_id,
	t1.category_name,
	t2.cost_budget_version_id,
	t2.years_month,
	t2.budget_amount 
FROM
	budget_fee_category t1
	LEFT JOIN cost_budget_version_detail t2 ON t2.budget_fee_category_id = t1.id 
WHERE
	t2.cost_budget_version_id = ?`
	var versionDetailDataList []dto.OriginalData
	if err := e.Orm.Raw(querySql, id).Scan(&versionDetailDataList).Error; err != nil {
		e.Log.Errorf("查询预算版本明细失败:%s", err)
		return resultMap
	}
	//先按 CostBudgetVersionId 分组
	var versionDetailDataMap = collectors.GroupBy(versionDetailDataList, func(item dto.OriginalData) int64 {
		return item.CostBudgetVersionId
	})
	// 转换每个分组为 MonthData
	for versionID, items := range versionDetailDataMap {
		resultMap[versionID] = e.transformToMonthData(items)
	}
	return resultMap
}

// transformToMonthData 将原始数据转换为月份数据（核心转换逻辑）
func (e *CostBudgetVersion) transformToMonthData(items []dto.OriginalData) []dto.MonthData {
	// 使用 map 按 ID 聚合，避免重复
	monthDataMap := make(map[int64]*dto.MonthData)

	for _, item := range items {
		// 如果不存在则创建新记录
		if _, exists := monthDataMap[item.Id]; !exists {
			monthDataMap[item.Id] = &dto.MonthData{
				Id:                  item.Id,
				ParentId:            item.ParentId,
				CategoryName:        item.CategoryName,
				CostBudgetVersionId: item.CostBudgetVersionId,
				Children:            make([]dto.MonthData, 0),
			}
		}

		// 解析月份并设置金额
		month := e.parseMonth(item.YearsMonth)
		if month >= 1 && month <= 12 {
			monthDataMap[item.Id].SetMonthValue(month, item.BudgetAmount)
		}
	}
	result := make([]dto.MonthData, 0, len(monthDataMap))
	for _, data := range monthDataMap {
		result = append(result, *data)
	}
	return result
}

func (e *CostBudgetVersion) parseMonth(yearsMonth string) int {
	if yearsMonth == "" {
		return 0
	}
	parts := strings.Split(yearsMonth, "-")
	if len(parts) != 2 {
		return 0
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return 0
	}
	return month
}

// Step2BuildTree 第二步：构建父子结构树并统计各节点数据
func (e *CostBudgetVersion) Step2BuildTree(monthDataMap map[int64][]dto.MonthData) map[int64][]dto.MonthData {
	resultMap := make(map[int64][]dto.MonthData)

	for versionID, monthDataList := range monthDataMap {

		builder := treeUtils.NewSimpleTreeBuilder[dto.MonthData]("Id", "ParentId", "Children")
		rootNodes := builder.BuildTree(monthDataList)
		for i := range rootNodes {
			e.aggregateNode(&rootNodes[i])
		}
		resultMap[versionID] = rootNodes
	}

	return resultMap
}

// aggregateNode 递归聚合单个节点的数据
func (e *CostBudgetVersion) aggregateNode(node *dto.MonthData) {
	if len(node.Children) == 0 {
		return
	}

	// 先递归处理所有子节点
	for i := range node.Children {
		e.aggregateNode(&node.Children[i])
	}

	// 累加子节点数据到父节点
	for _, child := range node.Children {
		node.Month1 += child.Month1
		node.Month2 += child.Month2
		node.Month3 += child.Month3
		node.Month4 += child.Month4
		node.Month5 += child.Month5
		node.Month6 += child.Month6
		node.Month7 += child.Month7
		node.Month8 += child.Month8
		node.Month9 += child.Month9
		node.Month10 += child.Month10
		node.Month11 += child.Month11
		node.Month12 += child.Month12
		node.MonthTotal += child.MonthTotal
	}
}

// Step3GroupToResult 第三步：转换为 GroupResult 格式并按版本统计
func (e *CostBudgetVersion) Step3GroupToResult(treeDataMap map[int64][]dto.MonthData, dataVO models.CostBudgetVersion) []dto.GroupResult {
	result := make([]dto.GroupResult, 0, len(treeDataMap))

	for _, treeData1 := range treeDataMap {
		group := dto.GroupResult{
			CostBudgetVersion: dataVO,
		}
		// 过滤掉没有子节点的要素
		filteredTreeData := make([]dto.MonthData, 0)
		for _, node := range treeData1 {
			if len(node.Children) > 0 {
				filteredTreeData = append(filteredTreeData, node)
			}
		}

		for _, node := range filteredTreeData {
			group.Month1 += node.Month1
			group.Month2 += node.Month2
			group.Month3 += node.Month3
			group.Month4 += node.Month4
			group.Month5 += node.Month5
			group.Month6 += node.Month6
			group.Month7 += node.Month7
			group.Month8 += node.Month8
			group.Month9 += node.Month9
			group.Month10 += node.Month10
			group.Month11 += node.Month11
			group.Month12 += node.Month12
			group.MonthTotal += node.MonthTotal
		}
		group.TreeData = filteredTreeData
		result = append(result, group)
	}

	return result
}
