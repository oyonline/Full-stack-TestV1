package service

import (
	"testing"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
)

// newTestSysPost 用 in-memory sqlite 装好 SysPost service，便于走完 BaseService 的 CRUD 全链路。
func newTestSysPost(t *testing.T) *SysPost {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SysPost{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	// SysPost 默认是软删除（DeletedAt），用一张干净表保证测试隔离。
	if err := db.Exec("DELETE FROM sys_post").Error; err != nil {
		t.Fatalf("clean table: %v", err)
	}

	s := &SysPost{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)
	return s
}

func TestSysPost_BaseServiceCRUD(t *testing.T) {
	s := newTestSysPost(t)

	// Insert 两条记录
	for i, name := range []string{"运维", "研发"} {
		req := &dto.SysPostInsertReq{
			PostName: name,
			PostCode: name + "_code",
			Sort:     i + 1,
			Status:   2,
			Remark:   "demo",
		}
		req.SetCreateBy(1)
		if err := s.Insert(req); err != nil {
			t.Fatalf("Insert(%s) failed: %v", name, err)
		}
	}

	// GetPage 应返回 2 条
	pageReq := &dto.SysPostPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 10
	var list []models.SysPost
	var count int64
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count=2, got %d", count)
	}
	if len(list) != 2 {
		t.Fatalf("expected list len=2, got %d", len(list))
	}

	// 根据已插入数据取出第 1 条（id 顺序由自增决定，不假设具体值）
	target := list[0]

	// Get 单条
	var got models.SysPost
	getReq := &dto.SysPostGetReq{Id: target.PostId}
	if err := s.Get(getReq, &got); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.PostId != target.PostId || got.PostName != target.PostName {
		t.Fatalf("Get returned wrong row: %+v vs %+v", got, target)
	}

	// Get 不存在的 id → 应返回中文 "查看对象不存在或无权查看" 错误
	missingReq := &dto.SysPostGetReq{Id: 999_999}
	var missing models.SysPost
	if err := s.Get(missingReq, &missing); err == nil {
		t.Fatalf("expected error for missing id, got nil")
	} else if err.Error() != "查看对象不存在或无权查看" {
		t.Fatalf("expected zh error, got %q", err.Error())
	}

	// Update：把 Status 从 2 改成 1
	updReq := &dto.SysPostUpdateReq{
		PostId:   target.PostId,
		PostName: target.PostName,
		PostCode: target.PostCode,
		Sort:     target.Sort,
		Status:   1,
		Remark:   "updated",
	}
	updReq.SetUpdateBy(2)
	if err := s.Update(updReq); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	var afterUpdate models.SysPost
	if err := s.Orm.First(&afterUpdate, target.PostId).Error; err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if afterUpdate.Status != 1 || afterUpdate.Remark != "updated" {
		t.Fatalf("Update did not persist: %+v", afterUpdate)
	}

	// Remove：删除单条 → 软删除生效（DeletedAt 设置；Find/First 看不到）
	delReq := &dto.SysPostDeleteReq{Ids: []int{target.PostId}}
	if err := s.Remove(delReq); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Get 已删除的 id → 软删除后应返回 not found 错误
	var afterDelete models.SysPost
	if err := s.Get(getReq, &afterDelete); err == nil {
		t.Fatalf("expected not-found after soft delete, got nil")
	}

	// 列表应只剩 1 条
	list = list[:0]
	count = 0
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage after delete failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count=1 after delete, got %d", count)
	}

	// Remove 一个不存在的 id → 应返回 "无权删除该数据"
	noopReq := &dto.SysPostDeleteReq{Ids: []int{999_999}}
	if err := s.Remove(noopReq); err == nil || err.Error() != "无权删除该数据" {
		t.Fatalf("expected zh error for delete-missing, got %v", err)
	}
}

func TestSysPost_InsertReturnPopulatesID(t *testing.T) {
	s := newTestSysPost(t)

	req := &dto.SysPostInsertReq{
		PostName: "财务",
		PostCode: "fin",
		Sort:     1,
		Status:   2,
	}
	req.SetCreateBy(1)
	var saved models.SysPost
	if err := s.InsertReturn(req, &saved); err != nil {
		t.Fatalf("InsertReturn failed: %v", err)
	}
	if saved.PostId == 0 {
		t.Fatalf("expected auto-increment PostId on saved model, got 0")
	}
	if saved.PostName != "财务" {
		t.Fatalf("expected PostName='财务', got %q", saved.PostName)
	}
}

func TestSysPost_GetPageRespectsPagination(t *testing.T) {
	s := newTestSysPost(t)

	// 插入 5 条
	for i := 0; i < 5; i++ {
		req := &dto.SysPostInsertReq{
			PostName: "p" + string(rune('A'+i)),
			PostCode: "code" + string(rune('A'+i)),
			Sort:     i,
			Status:   2,
		}
		req.SetCreateBy(1)
		if err := s.Insert(req); err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	pageReq := &dto.SysPostPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 2
	var list []models.SysPost
	var count int64
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected total count=5, got %d", count)
	}
	if len(list) != 2 {
		t.Fatalf("expected page size=2, got %d", len(list))
	}

	// 第 3 页应返回 1 条
	pageReq.PageIndex = 3
	list = list[:0]
	count = 0
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage page=3: %v", err)
	}
	if count != 5 || len(list) != 1 {
		t.Fatalf("expected page3 count=5 len=1, got count=%d len=%d", count, len(list))
	}
}
