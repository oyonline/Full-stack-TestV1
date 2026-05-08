package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"
)

// SkuCategory 类目树服务。
//
// 树形语义：parent_id=0 表示根节点；level 由 service 计算（root=1，子=父+1），
// 调用方不应自己传 level，避免不一致。
type SkuCategory struct {
	service.Service
}

// GetTree 返回完整的类目树（所有可见节点 + children 嵌套）。
// 不分页：类目数量在业务上有限，且前端常需要一次性渲染完整树。
func (e *SkuCategory) GetTree(c *dto.SkuCategoryPageReq, out *[]*dto.SkuCategoryTreeNode) error {
	var rows []models.SkuCategory
	q := e.Orm.Model(&models.SkuCategory{}).
		Scopes(cDto.MakeCondition(c.GetNeedSearch())).
		Order("parent_id ASC, sort ASC, category_id ASC")
	if err := q.Find(&rows).Error; err != nil {
		e.Log.Errorf("SkuCategory GetTree: %s", err)
		return err
	}

	nodes := make(map[int64]*dto.SkuCategoryTreeNode, len(rows))
	for i := range rows {
		n := &dto.SkuCategoryTreeNode{SkuCategory: rows[i]}
		nodes[rows[i].CategoryId] = n
	}

	roots := make([]*dto.SkuCategoryTreeNode, 0)
	for _, n := range nodes {
		if parent, ok := nodes[n.ParentId]; ok && n.ParentId != 0 {
			parent.Children = append(parent.Children, n)
		} else {
			roots = append(roots, n)
		}
	}
	*out = roots
	return nil
}

// GetPage 平铺列表（如树渲染不便，前端可走该接口拿扁平结构 + parent_id 自行渲染）。
func (e *SkuCategory) GetPage(c *dto.SkuCategoryPageReq, list *[]models.SkuCategory, count *int64) error {
	var data models.SkuCategory
	err := e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Order("parent_id ASC, sort ASC").
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("SkuCategory GetPage: %s", err)
		return err
	}
	return nil
}

// Get 单条查询。
func (e *SkuCategory) Get(c *dto.SkuCategoryGetReq, model *models.SkuCategory) error {
	db := e.Orm.Model(&models.SkuCategory{}).First(model, c.GetId())
	err := db.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("查看对象不存在或无权查看")
	}
	if err != nil {
		return err
	}
	return nil
}

// Insert 新增类目。Level 由 parent 决定，避免调用方传错。
func (e *SkuCategory) Insert(c *dto.SkuCategoryInsertReq) error {
	level, err := e.computeLevel(c.ParentId)
	if err != nil {
		return err
	}
	var data models.SkuCategory
	c.Generate(&data)
	data.Level = level
	if err := e.Orm.Create(&data).Error; err != nil {
		e.Log.Errorf("SkuCategory Insert: %s", err)
		return err
	}
	c.CategoryId = data.CategoryId
	return nil
}

// Update 修改类目。禁止把节点挂在自身或自己的子孙下，避免成环。
func (e *SkuCategory) Update(c *dto.SkuCategoryUpdateReq) error {
	if c.ParentId == c.CategoryId {
		return errors.New("不能将类目设为自身的父级")
	}
	var existing models.SkuCategory
	if err := e.Orm.First(&existing, c.CategoryId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("类目不存在")
		}
		return err
	}
	if c.ParentId != existing.ParentId {
		// 父级变更：检查不能挂在自己的子孙下
		descendant, err := e.isDescendant(c.CategoryId, c.ParentId)
		if err != nil {
			return err
		}
		if descendant {
			return errors.New("不能将类目挂在自己的子孙节点下")
		}
		level, err := e.computeLevel(c.ParentId)
		if err != nil {
			return err
		}
		existing.Level = level
	}
	c.Generate(&existing)
	if err := e.Orm.Save(&existing).Error; err != nil {
		e.Log.Errorf("SkuCategory Update: %s", err)
		return err
	}
	return nil
}

// Remove 批量删除。父类目存在子类目时拒绝删除，避免悬挂。
func (e *SkuCategory) Remove(c *dto.SkuCategoryDeleteReq) error {
	if len(c.Ids) == 0 {
		return errors.New("ids 不能为空")
	}
	var childCount int64
	if err := e.Orm.Model(&models.SkuCategory{}).
		Where("parent_id IN ?", c.Ids).
		Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("存在子类目，不能删除")
	}
	var data models.SkuCategory
	db := e.Orm.Model(&data).Delete(&data, c.Ids)
	if err := db.Error; err != nil {
		e.Log.Errorf("SkuCategory Remove: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}

// computeLevel 由 parent_id 推导 level：parent=0 → level=1；否则 = parent.level + 1。
func (e *SkuCategory) computeLevel(parentId int64) (int, error) {
	if parentId == 0 {
		return 1, nil
	}
	var parent models.SkuCategory
	if err := e.Orm.First(&parent, parentId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("父级类目不存在")
		}
		return 0, err
	}
	return parent.Level + 1, nil
}

// isDescendant 判定 candidate 是否为 root 的子孙（用于 Update 时防成环）。
// 通过沿 parent 链上溯：若上溯过程中遇到 root，则 candidate 是 root 的子孙。
func (e *SkuCategory) isDescendant(rootId, candidateId int64) (bool, error) {
	if candidateId == 0 || candidateId == rootId {
		return false, nil
	}
	current := candidateId
	for i := 0; i < 32 && current != 0; i++ { // 最多 32 层兜底，避免脏数据成环死循环
		var c models.SkuCategory
		if err := e.Orm.Select("category_id, parent_id").First(&c, current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		if c.ParentId == rootId {
			return true, nil
		}
		current = c.ParentId
	}
	return false, nil
}
