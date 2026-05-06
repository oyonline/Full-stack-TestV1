package service

import (
	"testing"

	"github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
)

// migrationSmoke 是 fs-2m0.2 批量迁移的烟雾测试。验证目标：
//
//  1. 嵌入了 baseservice.BaseService[T] 的 service 在装好 Orm/Log 之后，BaseService
//     提供的 5 件套方法可被调用、签名匹配、执行不 panic、SQL 行为与原实现一致。
//
//  2. 自定义方法（如 SysOperaLog.Insert(*model)、SysDictType.Insert (含唯一性校验)、
//     SysDictData.GetAll）仍然存在且工作。
//
// 不同于 sys_post 的全链路 CRUD 测试（已覆盖 BaseService 内部行为），这里每个迁移
// service 只跑一两条最小路径，确认嵌入 + 自定义方法共存正确。

func newSmokeORM(t *testing.T, schemas ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(schemas...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestSysLoginLog_BaseServiceMethodsReachable(t *testing.T) {
	db := newSmokeORM(t, &models.SysLoginLog{})
	if err := db.Exec("DELETE FROM sys_login_log").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}

	s := &SysLoginLog{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)

	// 直接通过 ORM 写一条记录，再走 Get / GetPage / Remove 验证 BaseService 的实现链路。
	row := models.SysLoginLog{Username: "alice", Status: "1"}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("insert seed: %v", err)
	}

	// GetPage
	pageReq := &dto.SysLoginLogGetPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 10
	var list []models.SysLoginLog
	var count int64
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if count != 1 || len(list) != 1 {
		t.Fatalf("expected 1 row, got count=%d len=%d", count, len(list))
	}

	// Get
	got := models.SysLoginLog{}
	if err := s.Get(&dto.SysLoginLogGetReq{Id: int(row.Id)}, &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("Username: %q", got.Username)
	}

	// Remove
	if err := s.Remove(&dto.SysLoginLogDeleteReq{Ids: []int{int(row.Id)}}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	count = 0
	list = list[:0]
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage after remove: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count=0 after remove, got %d", count)
	}
}

func TestSysOperaLog_KeepsCustomInsertSignature(t *testing.T) {
	db := newSmokeORM(t, &models.SysOperaLog{})
	if err := db.Exec("DELETE FROM sys_opera_log").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}

	s := &SysOperaLog{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)

	// 自定义 Insert 接受 *models.SysOperaLog（不是 DTO），仍然能被调用。
	if err := s.Insert(&models.SysOperaLog{Title: "test", BusinessType: "create"}); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	// BaseService 提供的 GetPage 仍然工作。
	pageReq := &dto.SysOperaLogGetPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 10
	var list []models.SysOperaLog
	var count int64
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count=1 after Insert, got %d", count)
	}
}

func TestSysDictType_CustomInsertEnforcesUniqueness(t *testing.T) {
	db := newSmokeORM(t, &models.SysDictType{})
	if err := db.Exec("DELETE FROM sys_dict_type").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}

	s := &SysDictType{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)

	first := &dto.SysDictTypeInsertReq{DictName: "状态", DictType: "sys_status"}
	if err := s.Insert(first); err != nil {
		t.Fatalf("Insert first: %v", err)
	}

	// 自定义 Insert 的唯一性校验保留在迁移后。
	dup := &dto.SysDictTypeInsertReq{DictName: "状态-dup", DictType: "sys_status"}
	if err := s.Insert(dup); err == nil {
		t.Fatalf("expected uniqueness error on duplicate dict_type, got nil")
	}

	// GetAll 是 BaseService 不内置的自定义方法，也保留。
	var all []models.SysDictType
	pageReq := &dto.SysDictTypeGetPageReq{}
	if err := s.GetAll(pageReq, &all); err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 dict type, got %d", len(all))
	}
}

func TestSysDictData_BaseServiceCRUD(t *testing.T) {
	db := newSmokeORM(t, &models.SysDictData{})
	if err := db.Exec("DELETE FROM sys_dict_data").Error; err != nil {
		t.Fatalf("clean: %v", err)
	}

	s := &SysDictData{}
	s.Orm = db
	s.Log = logger.NewHelper(logger.DefaultLogger)

	// Insert 走 BaseService 默认实现。
	if err := s.Insert(&dto.SysDictDataInsertReq{DictLabel: "启用", DictValue: "1", DictType: "sys_status"}); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	// GetPage
	pageReq := &dto.SysDictDataGetPageReq{}
	pageReq.PageIndex = 1
	pageReq.PageSize = 10
	var list []models.SysDictData
	var count int64
	if err := s.GetPage(pageReq, &list, &count); err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count=1, got %d", count)
	}

	// GetAll 是自定义保留方法。
	var all []models.SysDictData
	if err := s.GetAll(pageReq, &all); err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 row, got %d", len(all))
	}
}
