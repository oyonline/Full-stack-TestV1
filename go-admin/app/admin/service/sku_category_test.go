package service

import (
	"testing"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
)

func newTestSkuCategory(t *testing.T) *SkuCategory {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SkuCategory{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("DELETE FROM sku_category").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}
	s := &SkuCategory{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s
}

// TestSkuCategory_GetTree_Hierarchical 验证 root → 子 → 孙的嵌套关系。
func TestSkuCategory_GetTree_Hierarchical(t *testing.T) {
	s := newTestSkuCategory(t)

	// 构造：root1 -> child1 -> grandchild
	//      root2
	insert := func(name string, parent int64) int64 {
		req := &dto.SkuCategoryInsertReq{
			CategoryName: name,
			ParentId:     parent,
			Status:       models.SkuCategoryStatusEnabled,
		}
		req.SetCreateBy(1)
		if err := s.Insert(req); err != nil {
			t.Fatalf("insert(%s): %v", name, err)
		}
		return req.CategoryId
	}
	root1 := insert("root1", 0)
	child1 := insert("child1", root1)
	insert("grandchild", child1)
	insert("root2", 0)

	var tree []*dto.SkuCategoryTreeNode
	if err := s.GetTree(&dto.SkuCategoryPageReq{}, &tree); err != nil {
		t.Fatalf("GetTree: %v", err)
	}
	if len(tree) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(tree))
	}
	// 找 root1
	var r1 *dto.SkuCategoryTreeNode
	for _, n := range tree {
		if n.CategoryName == "root1" {
			r1 = n
			break
		}
	}
	if r1 == nil {
		t.Fatalf("root1 not found in tree")
	}
	if len(r1.Children) != 1 || r1.Children[0].CategoryName != "child1" {
		t.Fatalf("expected root1 → child1, got %+v", r1.Children)
	}
	if len(r1.Children[0].Children) != 1 || r1.Children[0].Children[0].CategoryName != "grandchild" {
		t.Fatalf("expected child1 → grandchild, got %+v", r1.Children[0].Children)
	}
}

// TestSkuCategory_LevelComputed 验证 level 由 service 推导，不依赖调用方。
func TestSkuCategory_LevelComputed(t *testing.T) {
	s := newTestSkuCategory(t)

	insert := func(name string, parent int64) int64 {
		req := &dto.SkuCategoryInsertReq{CategoryName: name, ParentId: parent}
		req.SetCreateBy(1)
		if err := s.Insert(req); err != nil {
			t.Fatalf("insert: %v", err)
		}
		return req.CategoryId
	}
	root := insert("r", 0)
	child := insert("c", root)
	grand := insert("g", child)

	check := func(id int64, want int) {
		var c models.SkuCategory
		if err := s.Orm.First(&c, id).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if c.Level != want {
			t.Fatalf("category=%d expected level=%d, got %d", id, want, c.Level)
		}
	}
	check(root, 1)
	check(child, 2)
	check(grand, 3)
}

// TestSkuCategory_RemoveBlockedByChildren 父类目存在子节点时删除被拒。
func TestSkuCategory_RemoveBlockedByChildren(t *testing.T) {
	s := newTestSkuCategory(t)
	parent := func() int64 {
		req := &dto.SkuCategoryInsertReq{CategoryName: "p", ParentId: 0}
		req.SetCreateBy(1)
		_ = s.Insert(req)
		return req.CategoryId
	}()
	childReq := &dto.SkuCategoryInsertReq{CategoryName: "c", ParentId: parent}
	childReq.SetCreateBy(1)
	_ = s.Insert(childReq)

	delReq := &dto.SkuCategoryDeleteReq{Ids: []int64{parent}}
	if err := s.Remove(delReq); err == nil {
		t.Fatalf("expected error when deleting parent with children")
	} else if err.Error() != "存在子类目，不能删除" {
		t.Fatalf("expected zh error, got %q", err.Error())
	}
}

// TestSkuCategory_Update_RejectsCycle 不允许把节点挂在自己的子孙下。
func TestSkuCategory_Update_RejectsCycle(t *testing.T) {
	s := newTestSkuCategory(t)
	insert := func(name string, parent int64) int64 {
		req := &dto.SkuCategoryInsertReq{CategoryName: name, ParentId: parent}
		req.SetCreateBy(1)
		_ = s.Insert(req)
		return req.CategoryId
	}
	root := insert("r", 0)
	child := insert("c", root)

	// 把 root 挂到 child 下 → 成环
	updReq := &dto.SkuCategoryUpdateReq{
		CategoryId:   root,
		CategoryName: "r",
		ParentId:     child,
	}
	updReq.SetUpdateBy(1)
	if err := s.Update(updReq); err == nil {
		t.Fatalf("expected error when creating cycle")
	} else if err.Error() != "不能将类目挂在自己的子孙节点下" {
		t.Fatalf("expected cycle error, got %q", err.Error())
	}

	// 把 root 挂到 root 自己下 → 自环
	selfReq := &dto.SkuCategoryUpdateReq{
		CategoryId: root,
		ParentId:   root,
	}
	selfReq.SetUpdateBy(1)
	if err := s.Update(selfReq); err == nil {
		t.Fatalf("expected error for self-parent")
	}
}
