package service

import (
	"errors"
	"fmt"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
	"go-admin/common/utils/collectors"
	"go-admin/common/utils/excelUtils"
	"go-admin/common/utils/treeUtils"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type BudgetFeeCategory struct {
	service.Service
}

// GetPage 获取BudgetFeeCategory列表
func (e *BudgetFeeCategory) GetPage(c *dto.BudgetFeeCategoryGetPageReq, p *actions.DataPermission, list *[]models.BudgetFeeCategory, count *int64) error {
	var err error
	var data models.BudgetFeeCategory

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("BudgetFeeCategoryService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

// Get 获取BudgetFeeCategory对象
func (e *BudgetFeeCategory) Get(d *dto.BudgetFeeCategoryGetReq, p *actions.DataPermission, model *models.BudgetFeeCategory) error {
	var data models.BudgetFeeCategory

	err := e.Orm.Model(&data).
		Scopes(
			actions.Permission(data.TableName(), p),
		).
		First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetBudgetFeeCategory error:%s \r\n", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// Insert 创建BudgetFeeCategory对象
func (e *BudgetFeeCategory) Insert(c *dto.BudgetFeeCategoryInsertReq, p *actions.DataPermission) error {
	var err error
	var data models.BudgetFeeCategory
	c.Generate(&data)
	//处理父级数据
	if err = dealPathStr(e, p, &data); err != nil {
		return err
	}
	err = e.Orm.Scopes(actions.Permission(data.TableName(), p)).Create(&data).Error
	if err != nil {
		e.Log.Errorf("BudgetFeeCategoryService Insert error:%s \r\n", err)
		return err
	}
	return nil
}

// Update 修改BudgetFeeCategory对象
func (e *BudgetFeeCategory) Update(c *dto.BudgetFeeCategoryUpdateReq, p *actions.DataPermission) error {
	var err error
	var data = models.BudgetFeeCategory{}
	e.Orm.Scopes(actions.Permission(data.TableName(), p)).First(&data, c.GetId())
	// 保存原始父ID，用于判断是否发生了父级变更
	originalParentId := data.ParentId
	originalCategory := data.CategoryName
	//赋值操作
	c.Generate(&data)
	if err = dealPathStr(e, p, &data); err != nil {
		return err
	}
	// 开始事务
	tx := e.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	//更新当前节点
	db := tx.Save(&data)
	if err = db.Error; err != nil {
		tx.Rollback()
		e.Log.Errorf("BudgetFeeCategoryService Save error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("无权更新该数据")
	}
	// 如果父级发生了变化，批量更新所有子级的路径信息
	if originalParentId != data.ParentId || originalCategory != data.CategoryName {
		err := batchUpdateChildrenPath(tx, data)
		if err != nil {
			tx.Rollback()
			e.Log.Errorf("BudgetFeeCategoryService batchUpdateChildrenPath error:%s \r\n", err)
			return err
		}
	}
	// 提交事务
	return tx.Commit().Error
}

// Remove 删除BudgetFeeCategory
func (e *BudgetFeeCategory) Remove(d *dto.BudgetFeeCategoryDeleteReq, p *actions.DataPermission) error {
	var data models.BudgetFeeCategory
	//判断是否有子级
	var children []models.BudgetFeeCategory
	if err1 := e.Orm.Where("parent_id in ?", d.GetId()).Find(&children).Error; err1 != nil {
		return err1
	}
	if len(children) > 0 {
		names := collectors.DistinctField(children, func(v models.BudgetFeeCategory) string { return v.CategoryName })
		return errors.New(fmt.Sprintf("已关联子级：【%s】，不能删除", strings.Join(names, ",")))
	}
	//判断预算版本明细中是否存在费用编码
	var count int64
	if err := e.Orm.Model(&models.CostBudgetVersionDetail{}).Where("budget_fee_category_id in ?", d.GetId()).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New(fmt.Sprintf("已关联预算版本明细，不能删除"))
	}
	//判断是否关联费用明细
	var count1 int64
	if err := e.Orm.Model(&models.BudgetFeeCategoryDetails{}).Where("budget_fee_category_id in ?", d.GetId()).Count(&count1).Error; err != nil {
		return err
	}
	if count1 > 0 {
		return errors.New(fmt.Sprintf("已关联费用明细，不能删除"))
	}
	//删除
	db := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Delete(&data, d.GetId())
	if err := db.Error; err != nil {
		e.Log.Errorf("Service RemoveBudgetFeeCategory error:%s \r\n", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

func (e *BudgetFeeCategory) GetAll(c *dto.BudgetFeeCategoryGetPageReq, p *actions.DataPermission, list *[]models.BudgetFeeCategory) error {
	var err error
	var data models.BudgetFeeCategory

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Error
	if err != nil {
		e.Log.Errorf("BudgetFeeCategoryService GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

func (e *BudgetFeeCategory) SetBudgetFeeCategoryListTree(c *dto.BudgetFeeCategoryGetPageReq, p *actions.DataPermission) ([]dto.BudgetFeeCategoryListTree, error) {
	var list []models.BudgetFeeCategory
	err := e.GetAll(c, p, &list)
	treeData := make([]dto.BudgetFeeCategoryListTree, len(list))
	for i := 0; i < len(list); i++ {
		e1 := dto.BudgetFeeCategoryListTree{}
		e1.Id = list[i].Id
		e1.ParentId = list[i].ParentId
		e1.CategoryName = list[i].CategoryName
		e1.CategoryNameEn = list[i].CategoryNameEn
		e1.CategoryCode = list[i].CategoryCode
		e1.ViewType = list[i].ViewType
		treeData[i] = e1
	}
	builder := treeUtils.NewSimpleTreeBuilder[dto.BudgetFeeCategoryListTree]("Id", "ParentId", "Children")
	treeData = builder.BuildTree(treeData)
	return treeData, err
}

func dealPathStr(e *BudgetFeeCategory, p *actions.DataPermission, data *models.BudgetFeeCategory) error {
	//是否存在校验,如果是更新操作，需要排除自身
	query := e.Orm.Model(&models.BudgetFeeCategory{}).Where("category_name = ? and view_type = ? and parent_id = ?", data.CategoryName, data.ViewType, data.ParentId)
	if data.Id > 0 {
		query = query.Where("id != ?", data.Id)
	}
	var count11 int64
	if err11 := query.Count(&count11).Error; err11 != nil {
		e.Log.Errorf("BudgetFeeCategoryService dealPathStr error:%s \r\n", err11)
		return err11
	}
	if count11 > 0 {
		return errors.New(fmt.Sprintf("费用类别【%s】已存在", data.CategoryName))
	}
	//查询父节点
	var dataParent models.BudgetFeeCategory
	e.Orm.Model(&dataParent).Scopes(actions.Permission(dataParent.TableName(), p)).First(&dataParent, data.ParentId)
	//处理路径信息
	if data.ParentId == 0 {
		data.Level = 1
		data.PathStr = data.CategoryName
		data.PathStrId = "/0/" + pkg.Int64ToString(data.Id) + "/"
	} else {
		data.Level = dataParent.Level + 1
		data.PathStr = dataParent.PathStr + ">" + data.CategoryName
		data.PathStrId = dataParent.PathStrId + pkg.Int64ToString(data.Id) + "/"
	}
	return nil
}

// 批量更新指定节点的所有子节点路径
func batchUpdateChildrenPath(tx *gorm.DB, data1 models.BudgetFeeCategory) error {
	// 收集所有需要更新的节点
	var nodesToUpdate []models.BudgetFeeCategory
	err := collectChildrenForUpdate(tx, data1, &nodesToUpdate)
	if err != nil {
		return err
	}
	// 批量更新
	if err := tx.Save(&nodesToUpdate).Error; err != nil {
		return err
	}
	return nil
}

// 递归收集需要更新的子节点
func collectChildrenForUpdate(tx *gorm.DB, data1 models.BudgetFeeCategory, nodes *[]models.BudgetFeeCategory) error {
	var children []models.BudgetFeeCategory
	err := tx.Where("parent_id = ?", data1.Id).Find(&children).Error
	if err != nil {
		return err
	}
	for _, child := range children {
		// 计算新的路径值
		child.Level = data1.Level + 1
		child.PathStr = data1.PathStr + ">" + child.CategoryName
		child.PathStrId = data1.PathStrId + pkg.Int64ToString(child.Id) + "/"
		*nodes = append(*nodes, child)
		// 递归收集子节点的子节点
		err = collectChildrenForUpdate(tx, child, nodes)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetFeeCategoryMap(tx *gorm.DB) (map[string]models.BudgetFeeCategory, error) {
	//查询费用类别ID
	var feeCategoryMap map[string]models.BudgetFeeCategory
	var feeCategoryList []models.BudgetFeeCategory
	if err := tx.Model(&models.BudgetFeeCategory{}).Find(&feeCategoryList).Error; err != nil {
		return feeCategoryMap, err
	}
	feeCategoryMap = collectors.ToMapDynamic(feeCategoryList, func(item models.BudgetFeeCategory) string {
		return item.PathStr + strconv.Itoa(item.ViewType)
	}, func(item models.BudgetFeeCategory) models.BudgetFeeCategory {
		return item
	})
	return feeCategoryMap, nil
}

func (e *BudgetFeeCategory) ImportData(req *dto.BudgetFeeCategoryImportReq, p *actions.DataPermission, userId int) error {
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
	detailData, importErr := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategory](dictMappings).ImportFromExcel(file, "")
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
	builder := treeUtils.NewSimpleTreeBuilder[models.BudgetFeeCategory]("CategoryName", "ParentId", "Children")
	detailData = builder.BuildTree(detailData)
	// 批量更新
	return tx.Commit().Error
}

func (e *BudgetFeeCategory) DownloadTemplate(httpResponseWriter http.ResponseWriter) error {
	// 创建空模板数据
	var templateData []models.BudgetFeeCategory
	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategory](dictMappings)
	// 执行流式导出
	filename := "费用类别模板"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationImport)
}

func (e *BudgetFeeCategory) Export(httpResponseWriter http.ResponseWriter, c *dto.BudgetFeeCategoryGetPageReq, p *actions.DataPermission) error {
	var err error
	var data models.BudgetFeeCategory
	var templateData []models.BudgetFeeCategory
	if err = e.Orm.Model(&data).
		Scopes(cDto.MakeCondition(c.GetNeedSearch()), actions.Permission(data.TableName(), p)).Find(&templateData).Error; err != nil {
		e.Log.Errorf("BudgetFeeCategoryDetailsService GetPage error:%s \r\n", err)
		return err
	}

	var dictMappings map[string]map[string]string
	// 创建结构体映射器
	mapper := excelUtils.NewExcelSingleStructureMapper[models.BudgetFeeCategory](dictMappings)
	// 执行流式导出
	filename := "费用类别数据"
	return mapper.ExportToExcel(templateData, httpResponseWriter, filename, filename, excelUtils.OperationExport)
}
