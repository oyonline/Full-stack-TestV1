export const workflowStatusLabels: Record<string, string> = {
  approved: '已通过',
  cancelled: '已撤回',
  draft: '草稿',
  in_review: '审批中',
  rejected: '已驳回',
};

export const workflowStatusColors: Record<string, string> = {
  approved: 'success',
  cancelled: 'default',
  draft: 'default',
  in_review: 'processing',
  rejected: 'error',
};

export const workflowActionLabels: Record<string, string> = {
  approve: '通过',
  reject: '驳回',
  start: '发起',
  withdraw: '撤回',
};

export const workflowTaskStatusLabels: Record<string, string> = {
  approved: '已通过',
  cancelled: '已取消',
  pending: '待处理',
  rejected: '已驳回',
};

export const workflowApproverTypeLabels: Record<string, string> = {
  role: '角色',
  user: '用户',
};

export function getWorkflowStatusLabel(value?: string) {
  return workflowStatusLabels[value || ''] || value || '-';
}

export function getWorkflowActionLabel(value?: string) {
  return workflowActionLabels[value || ''] || value || '-';
}

export function getWorkflowTaskStatusLabel(value?: string) {
  return workflowTaskStatusLabels[value || ''] || value || '-';
}

export function getWorkflowApproverTypeLabel(value?: string) {
  return workflowApproverTypeLabels[value || ''] || value || '-';
}
