import { requestClient } from '#/api/request';

/** 操作日志列表项（与后端 models.SysOperaLog 对齐） */
export interface SysOperaLogItem {
  id: number;
  title: string;
  businessType: string;
  businessTypes: string;
  method: string;
  requestMethod: string;
  operatorType: string;
  operName: string;
  deptName: string;
  operUrl: string;
  operIp: string;
  operLocation: string;
  operParam: string;
  status: string;
  operTime: string;
  jsonResult: string;
  remark: string;
  latencyTime: string;
  userAgent: string;
  createdAt: string;
  updatedAt: string;
}

export interface SysOperaLogPageResult {
  list: SysOperaLogItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

interface GetOperaLogPageParams {
  pageIndex?: number;
  pageSize?: number;
  title?: string;
  method?: string;
  requestMethod?: string;
  operUrl?: string;
  operIp?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
}

/**
 * 操作日志分页列表
 * 对接 go-admin GET /api/v1/sys-opera-log
 */
export async function getOperaLogPage(
  params: GetOperaLogPageParams = {},
): Promise<SysOperaLogPageResult> {
  return requestClient.get<SysOperaLogPageResult>('/v1/sys-opera-log', {
    params,
  });
}

/**
 * 操作日志详情
 * 对接 go-admin GET /api/v1/sys-opera-log/:id
 */
export async function getOperaLogDetail(
  id: number,
): Promise<SysOperaLogItem> {
  return requestClient.get<SysOperaLogItem>(`/v1/sys-opera-log/${id}`);
}

/**
 * 删除操作日志（支持批量）
 * 对接 go-admin DELETE /api/v1/sys-opera-log
 */
export async function deleteOperaLog(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-opera-log', { data: { ids } });
}
