package service

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"go-admin/app/admin/models"
	platformModels "go-admin/app/platform/models"
	platformService "go-admin/app/platform/service"
	platformDto "go-admin/app/platform/service/dto"
)

func init() {
	platformService.RegisterTerminalHandler(SpuBusinessType, onSpuWorkflowTerminal)
}

// onSpuWorkflowTerminal 是 SPU 在 workflow 终态（Approved/Rejected/Canceled）时的回调。
//
//   - Approved → SPU.status = SpuStatusApproved (3) + approved_at=now
//   - Rejected → SPU.status = SpuStatusRejected (4)
//   - Canceled → SPU.status = SpuStatusDraft (1)（撤回视为回到草稿，可再次编辑/提交）
//
// 由 platform workflow 在事务内回调；如果 binding.BusinessId 不可解析为 int 或对应 SPU 不存在，
// 函数返回错误以触发外层事务回滚——这避免 wf_instance 已终态但 SPU 状态遗留旧值的脏数据。
func onSpuWorkflowTerminal(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
	if binding == nil {
		return nil
	}
	spuId, err := strconv.ParseInt(binding.BusinessId, 10, 64)
	if err != nil || spuId <= 0 {
		// business_id 非法表示数据脏；不阻塞终态推进
		return nil
	}

	updates := map[string]interface{}{}
	switch terminalStatus {
	case platformDto.WorkflowStatusApproved:
		updates["status"] = models.SpuStatusApproved
		now := time.Now()
		updates["approved_at"] = &now
	case platformDto.WorkflowStatusRejected:
		updates["status"] = models.SpuStatusRejected
	case platformDto.WorkflowStatusCanceled:
		updates["status"] = models.SpuStatusDraft
	default:
		return nil
	}
	return tx.Model(&models.Spu{}).
		Where("spu_id = ?", spuId).
		Updates(updates).Error
}
