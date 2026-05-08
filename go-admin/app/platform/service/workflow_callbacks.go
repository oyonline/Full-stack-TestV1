package service

import (
	"errors"
	"sync"

	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
	"go-admin/app/platform/service/dto"
)

// WorkflowTerminalHandler 业务模块在 workflow 终态（已通过/已驳回/已撤销）时被调用。
// 在同一事务中执行，业务模块可以更新自己的 status 字段。
type WorkflowTerminalHandler func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error

var (
	terminalHandlers      = map[string]WorkflowTerminalHandler{}
	terminalHandlersMutex sync.RWMutex
)

// RegisterTerminalHandler 业务模块在 init() 时注册 handler。
// businessType 与 wf_business_binding.business_type 一致。
func RegisterTerminalHandler(businessType string, h WorkflowTerminalHandler) {
	terminalHandlersMutex.Lock()
	defer terminalHandlersMutex.Unlock()
	terminalHandlers[businessType] = h
}

// resetTerminalHandlersForTest 仅用于测试隔离 — 清空注册表。
func resetTerminalHandlersForTest() {
	terminalHandlersMutex.Lock()
	defer terminalHandlersMutex.Unlock()
	terminalHandlers = map[string]WorkflowTerminalHandler{}
}

// dispatchTerminalHandler 在 workflow.updateBusinessBinding 后调用。仅终态触发。
// 没有 binding 视为旁路 workflow，跳过；没有 handler 也跳过（旁路接入）。
func dispatchTerminalHandler(tx *gorm.DB, instance *platformModels.WorkflowInstance, terminalStatus string) error {
	if !isTerminalWorkflowStatus(terminalStatus) {
		return nil
	}
	var binding platformModels.WorkflowBusinessBinding
	if err := tx.Where("instance_id = ?", instance.InstanceId).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	terminalHandlersMutex.RLock()
	h := terminalHandlers[binding.BusinessType]
	terminalHandlersMutex.RUnlock()
	if h == nil {
		return nil
	}
	return h(tx, &binding, terminalStatus)
}

// isTerminalWorkflowStatus 判定是否为 workflow 终态。
func isTerminalWorkflowStatus(status string) bool {
	return status == dto.WorkflowStatusApproved ||
		status == dto.WorkflowStatusRejected ||
		status == dto.WorkflowStatusCanceled
}
