import { requestClient } from '#/api/request';

export interface WorkflowTaskItem {
  taskId: number;
  instanceId: number;
  definitionId: number;
  definitionName: string;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo: string;
  title: string;
  starterId: number;
  starterName: string;
  status: string;
  currentNodeId: number;
  currentNodeKey: string;
  currentNodeName: string;
  taskStatus: string;
  assigneeType: string;
  assigneeId: number;
  assigneeName: string;
  taskCreatedAt: string;
  startedAt: string;
  finishedAt?: string;
}

export interface WorkflowStartedItem {
  instanceId: number;
  definitionId: number;
  definitionName: string;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo: string;
  title: string;
  starterId: number;
  starterName: string;
  status: string;
  currentNodeId: number;
  currentNodeKey: string;
  currentNodeName: string;
  lastAction: string;
  lastActionRemark: string;
  startedAt: string;
  finishedAt?: string;
}

export interface WorkflowDefinitionNode {
  nodeId: number;
  definitionId: number;
  nodeKey: string;
  nodeName: string;
  nodeType: string;
  sort: number;
  approverType: string;
  approverValue: string;
  approverName: string;
  remark: string;
}

export interface WorkflowInstance {
  instanceId: number;
  definitionId: number;
  definitionKey: string;
  definitionName: string;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo: string;
  title: string;
  status: string;
  currentNodeId: number;
  currentNodeKey: string;
  currentNodeName: string;
  starterId: number;
  starterName: string;
  startedAt: string;
  finishedAt?: string;
  lastAction: string;
  lastActionRemark: string;
}

export interface WorkflowBusinessBinding {
  bindingId: number;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo: string;
  title: string;
  instanceId: number;
  workflowStatus: string;
  businessStatus: string;
  lastAction: string;
  lastActionRemark: string;
}

export interface WorkflowTaskDetail {
  taskId: number;
  instanceId: number;
  definitionId: number;
  nodeId: number;
  nodeKey: string;
  nodeName: string;
  assigneeType: string;
  assigneeId: number;
  assigneeName: string;
  status: string;
  action: string;
  comment: string;
  actionBy: number;
  actionByName: string;
  createdAt: string;
  processedAt?: string;
  cancelledReason: string;
}

export interface WorkflowActionLog {
  logId: number;
  instanceId: number;
  taskId: number;
  action: string;
  fromStatus: string;
  toStatus: string;
  fromNodeKey: string;
  fromNodeName: string;
  toNodeKey: string;
  toNodeName: string;
  operatorId: number;
  operatorName: string;
  comment: string;
  createdAt: string;
}

export interface WorkflowInstanceDetailResult {
  instance: WorkflowInstance;
  binding: WorkflowBusinessBinding;
  tasks: WorkflowTaskDetail[];
  actions: WorkflowActionLog[];
  nodes: WorkflowDefinitionNode[];
}

export interface WorkflowPageResult<T> {
  list: T[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

interface GetWorkflowTodoPageParams {
  pageIndex: number;
  pageSize: number;
  title?: string;
  businessType?: string;
  businessNo?: string;
  status?: string;
}

interface GetWorkflowStartedPageParams {
  pageIndex: number;
  pageSize: number;
  title?: string;
  businessType?: string;
  businessNo?: string;
  status?: string;
}

export async function getWorkflowTodoPage(
  params: GetWorkflowTodoPageParams,
): Promise<WorkflowPageResult<WorkflowTaskItem>> {
  return requestClient.get<WorkflowPageResult<WorkflowTaskItem>>(
    '/v1/platform/workflow/tasks/todo',
    { params },
  );
}

export async function getWorkflowStartedPage(
  params: GetWorkflowStartedPageParams,
): Promise<WorkflowPageResult<WorkflowStartedItem>> {
  return requestClient.get<WorkflowPageResult<WorkflowStartedItem>>(
    '/v1/platform/workflow/instances/started',
    { params },
  );
}

export async function getWorkflowInstanceDetail(
  instanceId: number,
): Promise<WorkflowInstanceDetailResult> {
  return requestClient.get<WorkflowInstanceDetailResult>(
    `/v1/platform/workflow/instances/${instanceId}`,
  );
}

export async function approveWorkflowTask(
  taskId: number,
  comment?: string,
): Promise<WorkflowInstanceDetailResult> {
  return requestClient.post<WorkflowInstanceDetailResult>(
    `/v1/platform/workflow/tasks/${taskId}/approve`,
    { comment: comment?.trim() || '' },
  );
}

export async function rejectWorkflowTask(
  taskId: number,
  comment?: string,
): Promise<WorkflowInstanceDetailResult> {
  return requestClient.post<WorkflowInstanceDetailResult>(
    `/v1/platform/workflow/tasks/${taskId}/reject`,
    { comment: comment?.trim() || '' },
  );
}

export async function withdrawWorkflowInstance(
  instanceId: number,
  comment?: string,
): Promise<WorkflowInstanceDetailResult> {
  return requestClient.post<WorkflowInstanceDetailResult>(
    `/v1/platform/workflow/instances/${instanceId}/withdraw`,
    { comment: comment?.trim() || '' },
  );
}
