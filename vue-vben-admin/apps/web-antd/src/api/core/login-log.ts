import { requestClient } from '#/api/request';

/** 登录日志列表项（与后端 models.SysLoginLog 对齐） */
export interface SysLoginLogItem {
  id: number;
  username: string;
  status: string;
  ipaddr: string;
  loginLocation: string;
  browser: string;
  os: string;
  platform: string;
  loginTime: string;
  remark: string;
  msg: string;
  createdAt: string;
  updatedAt: string;
}

export interface SysLoginLogPageResult {
  list: SysLoginLogItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

interface GetLoginLogPageParams {
  pageIndex?: number;
  pageSize?: number;
  username?: string;
  status?: string;
  ipaddr?: string;
  loginLocation?: string;
  beginTime?: string;
  endTime?: string;
}

/**
 * 登录日志分页列表
 * 对接 go-admin GET /api/v1/sys-login-log
 */
export async function getLoginLogPage(
  params: GetLoginLogPageParams = {},
): Promise<SysLoginLogPageResult> {
  return requestClient.get<SysLoginLogPageResult>('/v1/sys-login-log', {
    params,
  });
}

/**
 * 登录日志详情
 * 对接 go-admin GET /api/v1/sys-login-log/:id
 */
export async function getLoginLogDetail(
  id: number,
): Promise<SysLoginLogItem> {
  return requestClient.get<SysLoginLogItem>(`/v1/sys-login-log/${id}`);
}

/**
 * 删除登录日志（支持批量）
 * 对接 go-admin DELETE /api/v1/sys-login-log
 */
export async function deleteLoginLog(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-login-log', { data: { ids } });
}
