package service

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/collectors"
	"go-admin/common/utils/compareUtils"
	"go-admin/common/utils/excelUtils"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CostCenterInfo struct {
	service.Service
}

// GetPage 获取CostCenterInfo列表
func (e *CostCenterInfo) GetPage(c *dto.CostCenterInfoGetPageReq, p *actions.DataPermission, count *int64) ([]models.CostCenterInfo, error) {
	var err error
	var data models.CostCenterInfo
	var list []models.CostCenterInfo
	db := e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		)
	if c.GroupId != 0 {
		db = db.Where("id IN (SELECT cost_center_info_id FROM cost_center_related_customer WHERE group_id = ?)", c.GroupId)
	}
	if c.DeptId != 0 {
		db.Where("dept_id in (SELECT dept_id from sys_dept WHERE dept_path LIKE ?)", "%/"+strconv.FormatInt(c.DeptId, 10)+"/%")
	}
	if err = db.Find(&list).Limit(-1).Offset(-1).Count(count).Error; err != nil {
		e.Log.Errorf("查询分页结果失败:%s", err)
		return list, err
	}
	//查询客户分组信息
	list, err = e.getCustomerInfoData(list)
	if err != nil {
		return list, err
	}
	return list, nil
}

// Get 获取CostCenterInfo对象
func (e *CostCenterInfo) Get(d *dto.CostCenterInfoGetReq, p *actions.DataPermission) (models.CostCenterInfo, error) {
	var data models.CostCenterInfo
	var list []models.CostCenterInfo
	err := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Find(&list, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetCostCenterInfo error:%s \r\n", err)
		return data, err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return data, err
	}
	list, err = e.getCustomerInfoData(list)
	if err != nil {
		return data, err
	}
	if len(list) > 0 {
		data = list[0]
	}
	return data, nil
}

// Insert 创建CostCenterInfo对象
func (e *CostCenterInfo) Insert(c *dto.CostCenterInfoInsertReq) error {
	var data models.CostCenterInfo
	c.Generate(&data)
	//开启事务
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 统一处理保存逻辑
	var dateStr = time.Now().Format("20060102150405") + "01"
	if err := saveCostCenter(tx, data, "新增", []byte("{}"), dateStr); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Update 修改CostCenterInfo对象
func (e *CostCenterInfo) Update(c *dto.CostCenterInfoUpdateReq, p *actions.DataPermission) error {
	//查询修改前旧数据
	var oldData []models.CostCenterInfo
	if err := e.Orm.Model(&models.CostCenterInfo{}).Scopes(actions.Permission(models.CostCenterInfo{}.TableName(), p)).Find(&oldData, c.GetId()).Error; err != nil {
		return err
	}
	if oldData[0].Status != 3 {
		return errors.New("待启用的数据才能修改")
	}
	oldData, err := e.getCustomerInfoData(oldData)
	if err != nil {
		return err
	}
	//新修改数据
	var newData models.CostCenterInfo
	collectors.CopyFieldsWithIgnore(&oldData[0], &newData)
	c.Generate(&newData)

	deptDataMap, err := e.getDeptDataByDeptIds([]int64{})
	if err != nil {
		return err
	}
	//开启事务
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 判断是否有重要变更
	var dateStr = time.Now().Format("20060102150405") + "01"
	//正常更新
	if err = normalUpdate(tx, oldData[0], newData, deptDataMap, dateStr, "编辑"); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// Remove 删除CostCenterInfo
func (e *CostCenterInfo) Remove(d *dto.CostCenterInfoDeleteReq, p *actions.DataPermission) error {
	//是否关联预算版本
	var count int64
	e.Orm.Table("cost_budget_version").Where("cost_center_info_id in ?", d.GetId()).Count(&count)
	if count > 0 {
		return errors.New("关联预算版本数据，不能删除")
	}
	//草稿状态才能删除
	var dataList []models.CostCenterInfo
	if err := e.Orm.Model(&models.CostCenterInfo{}).Where("id in ?", d.GetId()).Find(&dataList).Error; err != nil {
		return err
	}
	for _, v := range dataList {
		if v.Status != 3 {
			return errors.New("待启用的数据才能删除")
		}
	}
	//删除操作
	if err := e.Orm.Model(&models.CostCenterInfo{}).Scopes(actions.Permission(models.CostCenterInfo{}.TableName(), p)).Delete(&models.CostCenterInfo{}, d.GetId()).Error; err != nil {
		e.Log.Errorf("删除失败:%s", err)
		return err
	}
	return nil
}

func (e *CostCenterInfo) ImportData(req *dto.CostCenterInfoImportReq, p *actions.DataPermission, userId int) error {
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
	detailData, importErr := excelUtils.NewExcelSingleStructureMapper[models.CostCenterInfo](dictMappings).ImportFromExcel(file, "")
	if importErr != nil {
		return importErr
	}
	// 查询所有旧数据建立索引
	oldDataMap, errIndex := e.buildOldDataIndex(detailData)
	if errIndex != nil {
		return errIndex
	}
	// 查询辅助数据
	deptDataMap, groupNameMap, dataErr := e.getDeptAndCustomerInfoByAll()
	if dataErr != nil {
		return dataErr
	}
	// 开始事务
	tx := e.Orm.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 批量处理
	for i := range detailData {
		detailData[i].CreateBy = userId
		detailData[i].UpdateBy = userId
		// 转换部门 ID
		if deptVO, exists := deptDataMap[detailData[i].DeptPathName]; exists {
			detailData[i].DeptId = int64(deptVO.DeptId)
		} else {
			return errors.New(fmt.Sprintf("部门不存在：%s", detailData[i].DeptPathName))
		}
		//关联客户分组
		if detailData[i].GroupNameStrList != "" {
			var groupNameList1 []models.GroupNameInfoData
			groupNameList := strings.Split(detailData[i].GroupNameStrList, ",")
			for _, groupName := range groupNameList {
				// 去除前后空格
				groupName = strings.TrimSpace(groupName)
				if groupVO, exists := groupNameMap[groupName]; exists {
					groupNameList1 = append(groupNameList1, models.GroupNameInfoData{
						GroupId:   groupVO.GroupId,
						GroupName: groupVO.GroupName,
					})
				} else {
					return errors.New(fmt.Sprintf("客户分组不存在：%s", groupName))
				}
			}
			detailData[i].GroupNameList = groupNameList1
		}
		// 查找是否存在（按名称或编码）
		var dateStr = time.Now().Format("20060102150405") + "0" + strconv.Itoa(i+1)
		if oldDataVO, exists := oldDataMap[detailData[i].CostCenterName]; exists {
			// 修改
			if oldDataVO.Status != 3 {
				return errors.New(fmt.Sprintf("【%s】状态不是待启用，不能修改", oldDataVO.CostCenterName))
			}
			detailData[i].Id = oldDataVO.Id
			if err := normalUpdate(tx, oldDataVO, detailData[i], map[int64]models.SysDept{}, dateStr, "导入编辑"); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// 新增
			if err := saveCostCenter(tx, detailData[i], "导入新增", []byte("{}"), dateStr); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func (e *CostCenterInfo) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.CostCenterInfo
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterInfo](dictMappings)
	// 执行流式导出
	filename := "成本中心数据模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *CostCenterInfo) Export(httpResponseWriter http.ResponseWriter, c *dto.CostCenterInfoGetPageReq, p *actions.DataPermission) error {
	var err error
	var tableName = models.CostCenterInfo{}.TableName()
	var exportData []models.CostCenterInfo
	err = e.Orm.Table(tableName).Scopes(cDto.MakeCondition(c.GetNeedSearch()), actions.Permission(tableName, p)).Find(&exportData).Error
	if err != nil {
		e.Log.Errorf("导出失败:%s", err)
		return err
	}
	exportData, err = e.getCustomerInfoData(exportData)
	if err != nil {
		e.Log.Errorf("导出-组装数据失败:%s", err)
		return err
	}
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.CostCenterInfo](dictMappings)
	// 执行流式导出
	filename := "成本中心数据"
	return mapper.ExportToExcel(exportData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}

func (e *CostCenterInfo) getCustomerInfoData(list []models.CostCenterInfo) ([]models.CostCenterInfo, error) {
	var costCenterIDs []int64
	var deptIDs []int64
	for _, v := range list {
		costCenterIDs = append(costCenterIDs, v.Id)
		deptIDs = append(deptIDs, v.DeptId)
	}
	//查询客户分组信息
	err2, customerInfoDataMap, groupNameMap := e.getCustomerInfoDataByCostCenterIds(costCenterIDs)
	if err2 != nil {
		return list, err2
	}
	//查询部门信息
	deptDataMap, err2 := e.getDeptDataByDeptIds(deptIDs)
	if err2 != nil {
		return list, err2
	}
	//组装数据
	for i := range list {
		v := &list[i]
		if customerInfos, exists := customerInfoDataMap[v.Id]; exists && len(customerInfos) > 0 {
			v.CustomerInfoList = customerInfos
		}
		if groupNames, exists := groupNameMap[v.Id]; exists && len(groupNames) > 0 {
			v.GroupNameList = groupNames
			var groupNameStrList []string
			var groupIds []string
			for _, group := range v.GroupNameList {
				groupNameStrList = append(groupNameStrList, group.GroupName)
				groupIds = append(groupIds, strconv.FormatInt(group.GroupId, 10))
			}
			v.GroupNameStrList = strings.Join(groupNameStrList, ",")
			v.GroupIds = strings.Join(groupIds, ",")
		}
		if dept, exists := deptDataMap[v.DeptId]; exists {
			v.DeptPathName = dept.DeptPathName
		}
	}
	return list, nil
}

func (e *CostCenterInfo) getDeptDataByDeptIds(deptIDs []int64) (map[int64]models.SysDept, error) {
	var deptList []models.SysDept
	var deptDataMap map[int64]models.SysDept
	db := e.Orm.Table("sys_dept")

	if len(deptIDs) > 0 {
		db = db.Where("dept_id in ?", deptIDs)
	}

	if err := db.Find(&deptList).Error; err != nil {
		e.Log.Errorf("查询部门信息失败:%s ", err)
		return deptDataMap, err
	}
	deptDataMap = collectors.ToMapDynamic(deptList, func(v models.SysDept) int64 { return int64(v.DeptId) }, func(v models.SysDept) models.SysDept { return v })
	return deptDataMap, nil
}

func (e *CostCenterInfo) getCustomerInfoDataByCostCenterIds(costCenterIDs []int64) (error, map[int64][]models.CustomerInfoData, map[int64][]models.GroupNameInfoData) {
	var customerInfoList []models.CustomerInfoData
	var customerInfoDataMap map[int64][]models.CustomerInfoData
	var groupNameMap map[int64][]models.GroupNameInfoData

	db := e.Orm.Table("cost_center_related_customer t1").
		Joins("JOIN kingdee_customer t2 ON t2.group_id = t1.group_id").
		Select("t1.cost_center_info_id,\n\tt2.group_id,\n\tt2.group_name,\n\tt2.group_number,\n\tt2.customer_name ")

	if len(costCenterIDs) > 0 {
		db = db.Where("t1.cost_center_info_id in ?", costCenterIDs)
	}
	if err := db.Find(&customerInfoList).Error; err != nil {
		e.Log.Errorf("查询客户分组信息失败:%s ", err)
		return err, customerInfoDataMap, groupNameMap
	}
	customerInfoDataMap = collectors.GroupBy(customerInfoList, func(v models.CustomerInfoData) int64 { return v.CostCenterInfoId })
	groupNameMap = collectors.GroupByAndDistinct(customerInfoList, func(v models.CustomerInfoData) int64 { return v.CostCenterInfoId }, func(v models.CustomerInfoData) models.GroupNameInfoData {
		return models.GroupNameInfoData{CostCenterInfoId: v.CostCenterInfoId, GroupId: v.GroupId, GroupName: v.GroupName}
	})
	return nil, customerInfoDataMap, groupNameMap
}

func (e *CostCenterInfo) getDeptAndCustomerInfoByAll() (map[string]models.SysDept, map[string]models.KingdeeCustomer, error) {
	var deptList []models.SysDept
	var deptDataMap map[string]models.SysDept

	var groupNameList []models.KingdeeCustomer
	var groupNameMap map[string]models.KingdeeCustomer

	err := e.Orm.Table("sys_dept").Find(&deptList).Error
	if err != nil {
		e.Log.Errorf("查询部门信息失败:%s ", err)
		return deptDataMap, groupNameMap, err
	}
	deptDataMap = collectors.ToMapDynamic(deptList, func(v models.SysDept) string { return v.DeptPathName }, func(v models.SysDept) models.SysDept { return v })

	err = e.Orm.Table("kingdee_customer").Find(&groupNameList).Error
	if err != nil {
		e.Log.Errorf("查询客户分组信息失败:%s ", err)
		return deptDataMap, groupNameMap, err
	}
	groupNameMap = collectors.ToMapDynamic(groupNameList, func(v models.KingdeeCustomer) string { return v.GroupName }, func(v models.KingdeeCustomer) models.KingdeeCustomer { return v })
	return deptDataMap, groupNameMap, err
}

func autoSetStatus(inputEffectiveDate time.Time) int {
	now := time.Now()
	if !inputEffectiveDate.IsZero() {
		if !inputEffectiveDate.After(now) {
			return 2 // 启用
		}
		return 3 // 待启用
	}
	// 默认：待启用，时间为当前
	return 3
}

func (e *CostCenterInfo) hasImportantChange(oldData, newData models.CostCenterInfo) bool {
	// 部门变化
	if oldData.DeptId != newData.DeptId {
		return true
	}

	// 客户分组数量变化
	if len(oldData.GroupNameList) != len(newData.GroupNameList) {
		return true
	}

	// 客户分组内容变化（使用 map 高效比较）
	oldGroupIds := make(map[int64]bool, len(oldData.GroupNameList))
	for _, g := range oldData.GroupNameList {
		oldGroupIds[g.GroupId] = true
	}

	for _, g := range newData.GroupNameList {
		if !oldGroupIds[g.GroupId] {
			return true
		}
	}

	return false
}

func saveCostCenter(tx *gorm.DB, data models.CostCenterInfo, changeType string, changeDetails json.RawMessage, dateStr string) error {
	// 智能设置状态和启用时间
	data.Status = autoSetStatus(data.EffectiveDate)
	// 处理停用时间
	if err := dealStopDateByInsert(tx, data); err != nil {
		return err
	}
	// 1. 插入主表
	if err := tx.Create(data).Error; err != nil {
		log.Errorf("新增失败:%s", err)
		return err
	}
	// 2. 保存关联客户分组
	if err := saveRelatedCustomers(tx, data); err != nil {
		return err
	}
	// 3. 生成变更记录
	if err := createChangeLog(tx, changeType, changeDetails, data, dateStr); err != nil {
		return err
	}

	return nil
}

func saveRelatedCustomers(tx *gorm.DB, data models.CostCenterInfo) error {
	if len(data.GroupNameList) == 0 {
		return nil
	}
	// 删除旧的关联
	if err := tx.Unscoped().Where("cost_center_info_id = ?", data.Id).Delete(&models.CostCenterRelatedCustomer{}).Error; err != nil {
		return err
	}
	//关联客户分组
	var relatedCustomerList []models.CostCenterRelatedCustomer
	for _, v := range data.GroupNameList {
		var relatedCustomer = models.CostCenterRelatedCustomer{
			CostCenterInfoId: data.Id,
			GroupId:          v.GroupId,
		}
		relatedCustomer.CreateBy = data.CreateBy
		relatedCustomer.UpdateBy = data.UpdateBy
		relatedCustomerList = append(relatedCustomerList, relatedCustomer)
	}
	// 批量插入新关联
	if err := tx.Create(&relatedCustomerList).Error; err != nil {
		log.Errorf("保存关联失败:%s", err)
		return err
	}
	return nil
}

func createChangeLog(tx *gorm.DB, changeType string, changeDetails json.RawMessage,
	data models.CostCenterInfo, dateStr string) error {
	//生成变更记录
	var changeLog = models.CostCenterInfoChange{
		CostCenterInfoId: data.Id,
		ChangeOrder:      "C" + dateStr,
		ChangeType:       changeType,
		Status:           data.Status,
		EffectiveDate:    data.EffectiveDate,
		ChangeDetails:    changeDetails,
		VersionNumber:    "V" + dateStr,
	}
	changeLog.CreateBy = data.CreateBy
	changeLog.UpdateBy = data.UpdateBy
	if err := tx.Create(&changeLog).Error; err != nil {
		log.Errorf("生成变更记录失败:%s", err)
		return err
	}

	return nil
}

func normalUpdate(tx *gorm.DB, oldData, newData models.CostCenterInfo, deptDataMap map[int64]models.SysDept, dateStr string, changeType string) error {
	newData.Status = autoSetStatus(newData.EffectiveDate)
	//处理停用时间
	if err := dealStopDateByInsert(tx, newData); err != nil {
		return err
	}
	// 比较变更
	changeDetails, err := getChangeDetails(&oldData, &newData, deptDataMap)
	if err != nil {
		return err
	}
	//更新主表
	if err = tx.Model(&models.CostCenterInfo{}).Where("id = ?", newData.Id).Updates(newData).Error; err != nil {
		log.Errorf("正常更新失败:%s", err)
		return err
	}
	// 2. 保存关联客户分组
	if err := saveRelatedCustomers(tx, newData); err != nil {
		return err
	}
	// 生成变更记录
	if err = createChangeLog(tx, changeType, changeDetails, newData, dateStr); err != nil {
		return err
	}
	return nil
}

func getChangeDetails(oldData1, newData1 *models.CostCenterInfo, deptDataMap map[int64]models.SysDept) (json.RawMessage, error) {
	//获取部门信息
	if deptVO, exist := deptDataMap[newData1.DeptId]; exist {
		newData1.DeptPathName = deptVO.DeptPathName
	}
	//关联客户分组
	var groupNameStr []string
	var groupIds []string
	for _, v := range newData1.GroupNameList {
		groupNameStr = append(groupNameStr, v.GroupName)
		groupIds = append(groupIds, strconv.FormatInt(v.GroupId, 10))
	}
	newData1.GroupNameStrList = strings.Join(groupNameStr, ",")
	newData1.GroupIds = strings.Join(groupIds, ",")
	return compareUtils.Compare(oldData1, newData1)
}

func (e *CostCenterInfo) buildOldDataIndex(detailData []models.CostCenterInfo) (map[string]models.CostCenterInfo, error) {
	//查询旧数据
	var tableName = models.CostCenterInfo{}.TableName()
	var oldDataList []models.CostCenterInfo
	costCenterNameList := collectors.DistinctField(detailData, func(v models.CostCenterInfo) string { return v.CostCenterName })
	if err := e.Orm.Table(tableName).Where("cost_center_name in ?", costCenterNameList).Find(&oldDataList).Error; err != nil {
		return nil, err
	}
	oldDataList, err := e.getCustomerInfoData(oldDataList)
	if err != nil {
		return nil, err
	}
	oldDataMap := collectors.ToMapDynamic(oldDataList, func(v models.CostCenterInfo) string {
		return v.CostCenterName
	}, func(v models.CostCenterInfo) models.CostCenterInfo { return v })
	return oldDataMap, nil
}

func (e *CostCenterInfo) CostCenterTimeTask() error {
	var list []models.CostCenterInfo
	if err := e.Orm.Table("cost_center_info t1").
		Joins("INNER JOIN ( SELECT cost_center_name, MAX( id ) AS max_id FROM cost_center_info WHERE `status` = 3 GROUP BY cost_center_name ) t2 ON t2.cost_center_name = t1.cost_center_name AND t2.max_id = t1.id ").
		Select("t1.*").Where("t1.`status`= ?", 3).Find(&list).Error; err != nil {
		e.Log.Errorf("查询失败:%s", err)
		return err
	}

	tx := e.Orm.Debug().Begin()
	for i, v := range list {
		if !v.EffectiveDate.After(time.Now()) {
			// 包含两种情况：
			// 1. EffectiveDate < Now  → 昨天及以前，启用
			// 2. EffectiveDate == Now → 今天，启用
			v.Status = 2
			//处理停用时间
			if err := dealStopDateByInsert(tx, v); err != nil {
				e.Log.Errorf("CostCenterTimeTask-处理停用时间更新失败:%s", err)
				tx.Rollback()
				return err
			}
			//更新主表
			if err := tx.Model(&models.CostCenterInfo{}).Where("id = ?", v.Id).Updates(v).Error; err != nil {
				e.Log.Errorf("CostCenterTimeTask-更新失败:%s", err)
				tx.Rollback()
				return err
			}
			// 生成变更记录
			var dateStr = time.Now().Format("20060102150405") + "0" + strconv.Itoa(i+1)
			var changeDetails = []byte("{}")
			if err := createChangeLog(tx, "定时任务更新状态", changeDetails, v, dateStr); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func dealStopDateByInsert(tx *gorm.DB, data models.CostCenterInfo) error {
	if data.Status == 2 {
		//找到相同的成本中心停用旧的成本中心
		var oldDataList []models.CostCenterInfo
		if err := tx.Model(&models.CostCenterInfo{}).Where("(cost_center_name = ?)", data.CostCenterName).Find(&oldDataList).Error; err != nil {
			return err
		}
		oldIds := collectors.DistinctField(oldDataList, func(v models.CostCenterInfo) int64 { return v.Id })

		if err := tx.Model(&models.CostCenterInfo{}).Where("id in ?", oldIds).
			Updates(map[string]interface{}{"status": 1, "update_by": data.UpdateBy, "updated_at": time.Now(), "stop_date": time.Now()}).Error; err != nil {
			return err
		}
		//停用预算版本关联
		if err := tx.Model(&models.CostBudgetVersion{}).Where("cost_center_info_id in ?", oldIds).
			Updates(map[string]interface{}{"status": 3, "update_by": data.UpdateBy, "updated_at": time.Now(), "stop_date": time.Now()}).Error; err != nil {
			return err
		}
	}
	return nil
}
