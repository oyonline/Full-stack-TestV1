import { requestClient } from '#/api/request';

/** 定时任务列表项（与后端 dto.SysJobItem 对齐） */
export interface SysJobItem {
  jobId: number;
  jobName: string;
  jobGroup: string;
  jobType: number;
  cronExpression: string;
  invokeTarget: string;
  args: string;
  misfirePolicy: number;
  concurrent: number;
  status: number;
  entryId: number;
}

export interface SysJobPageResult {
  list: SysJobItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

interface GetSysJobPageParams {
  pageIndex?: number;
  pageSize?: number;
  jobId?: number;
  jobName?: string;
  jobGroup?: string;
  cronExpression?: string;
  invokeTarget?: string;
  status?: number;
}

export interface SysJobPayload {
  jobId?: number;
  jobName: string;
  jobGroup?: string;
  jobType?: number;
  cronExpression?: string;
  invokeTarget?: string;
  args?: string;
  misfirePolicy?: number;
  concurrent?: number;
  status?: number;
  entryId?: number;
}

/**
 * 定时任务分页列表
 * 对接 go-admin GET /api/v1/sysjob
 */
export async function getSysJobPage(
  params: GetSysJobPageParams = {},
): Promise<SysJobPageResult> {
  return requestClient.get('/v1/sysjob', { params });
}

/**
 * 定时任务详情
 * 对接 go-admin GET /api/v1/sysjob/:id
 */
export async function getSysJobDetail(id: number): Promise<SysJobItem> {
  return requestClient.get(`/v1/sysjob/${id}`);
}

/**
 * 新增定时任务
 * 对接 go-admin POST /api/v1/sysjob
 */
export async function createSysJob(data: SysJobPayload): Promise<number> {
  return requestClient.post('/v1/sysjob', data);
}

/**
 * 更新定时任务
 * 对接 go-admin PUT /api/v1/sysjob
 */
export async function updateSysJob(data: SysJobPayload): Promise<number> {
  return requestClient.put('/v1/sysjob', data);
}

/**
 * 删除定时任务（支持批量）
 * 对接 go-admin DELETE /api/v1/sysjob
 */
export async function deleteSysJob(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sysjob', {
    data: { ids },
  });
}

/**
 * 启动定时任务
 * 对接 go-admin GET /api/v1/job/start/:id
 */
export async function startSysJob(id: number): Promise<unknown> {
  return requestClient.get(`/v1/job/start/${id}`);
}

/**
 * 停止/移除定时任务
 * 对接 go-admin GET /api/v1/job/remove/:id
 */
export async function removeSysJob(id: number): Promise<unknown> {
  return requestClient.get(`/v1/job/remove/${id}`);
}
