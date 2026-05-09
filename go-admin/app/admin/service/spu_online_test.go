package service

import (
	"testing"

	"go-admin/app/admin/models"
)

// TestSpu_GoOffline_HappyPath 验证 status=3 && is_online=true 的 SPU 可正常下架，
// 且级联把所有 SKU 置为 SkuStatusDisabled(1)。
func TestSpu_GoOffline_HappyPath(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "OFF-1", SpuName: "online spu", Status: models.SpuStatusApproved, IsOnline: true}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	sku1 := models.Sku{SpuId: spu.SpuId, SkuCode: "SK-1", SkuName: "sku1", Status: models.SkuStatusEnabled}
	sku2 := models.Sku{SpuId: spu.SpuId, SkuCode: "SK-2", SkuName: "sku2", Status: models.SkuStatusEnabled}
	if err := s.Orm.Create(&sku1).Error; err != nil {
		t.Fatalf("seed sku1: %v", err)
	}
	if err := s.Orm.Create(&sku2).Error; err != nil {
		t.Fatalf("seed sku2: %v", err)
	}

	if err := s.GoOffline(spu.SpuId, 1, nil); err != nil {
		t.Fatalf("GoOffline: %v", err)
	}

	var updSpu models.Spu
	s.Orm.First(&updSpu, spu.SpuId)
	if updSpu.IsOnline {
		t.Fatalf("expected is_online=false after GoOffline")
	}

	var count int64
	s.Orm.Model(&models.Sku{}).Where("spu_id = ? AND status = ?", spu.SpuId, models.SkuStatusEnabled).Count(&count)
	if count != 0 {
		t.Fatalf("expected all SKU disabled after GoOffline, still %d enabled", count)
	}
}

// TestSpu_GoOffline_RejectsNonApproved 验证非 status=3 的 SPU 不能下架。
func TestSpu_GoOffline_RejectsNonApproved(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "OFF-2", SpuName: "draft spu", Status: models.SpuStatusDraft, IsOnline: false}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	if err := s.GoOffline(spu.SpuId, 1, nil); err == nil {
		t.Fatalf("expected error for non-approved SPU GoOffline")
	}
}

// TestSpu_GoOffline_RejectsAlreadyOffline 验证已下架（is_online=false）的 SPU 不能重复下架。
func TestSpu_GoOffline_RejectsAlreadyOffline(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "OFF-3", SpuName: "offline spu", Status: models.SpuStatusApproved, IsOnline: false}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	if err := s.GoOffline(spu.SpuId, 1, nil); err == nil {
		t.Fatalf("expected error for already-offline SPU GoOffline")
	}
}

// TestSpu_GoOnline_HappyPath 验证 status=3 && is_online=false 的 SPU 可正常上架，
// 且不恢复 SKU 状态（架构师 §8.3）。
func TestSpu_GoOnline_HappyPath(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "ON-1", SpuName: "offline spu", Status: models.SpuStatusApproved, IsOnline: false}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	// SKU 处于 disabled 状态，GoOnline 不应改它
	sku := models.Sku{SpuId: spu.SpuId, SkuCode: "SK-ON-1", SkuName: "sku", Status: models.SkuStatusDisabled}
	if err := s.Orm.Create(&sku).Error; err != nil {
		t.Fatalf("seed sku: %v", err)
	}

	if err := s.GoOnline(spu.SpuId, 1, nil); err != nil {
		t.Fatalf("GoOnline: %v", err)
	}

	var updSpu models.Spu
	s.Orm.First(&updSpu, spu.SpuId)
	if !updSpu.IsOnline {
		t.Fatalf("expected is_online=true after GoOnline")
	}

	// SKU 应保持 disabled（不反向恢复）
	var updSku models.Sku
	s.Orm.First(&updSku, sku.SkuId)
	if updSku.Status != models.SkuStatusDisabled {
		t.Fatalf("expected SKU status unchanged (disabled), got %d", updSku.Status)
	}
}

// TestSpu_GoOnline_RejectsNonApproved 验证非 status=3 的 SPU 不能上架。
func TestSpu_GoOnline_RejectsNonApproved(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "ON-2", SpuName: "draft spu", Status: models.SpuStatusDraft, IsOnline: false}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	if err := s.GoOnline(spu.SpuId, 1, nil); err == nil {
		t.Fatalf("expected error for non-approved SPU GoOnline")
	}
}

// TestSpu_GoOnline_RejectsAlreadyOnline 验证已上架（is_online=true）的 SPU 不能重复上架。
func TestSpu_GoOnline_RejectsAlreadyOnline(t *testing.T) {
	s := newTestSpu(t)

	spu := models.Spu{SpuCode: "ON-3", SpuName: "online spu", Status: models.SpuStatusApproved, IsOnline: true}
	if err := s.Orm.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu: %v", err)
	}
	if err := s.GoOnline(spu.SpuId, 1, nil); err == nil {
		t.Fatalf("expected error for already-online SPU GoOnline")
	}
}
