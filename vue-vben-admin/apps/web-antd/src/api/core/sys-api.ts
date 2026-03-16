import { requestClient } from '#/api/request';

/** go-admin 列表项（与后端 models.SysApi 对齐） */
export interface SysApiItem {
  id: number;
  handle?: string;
  title?: string;
  path?: string;
  action?: string;
  type?: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 分页响应：requestClient 解包后为 data 内容 */
export interface SysApiPageResult {
  list: SysApiItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 分页请求参数（与 dto.SysApiGetPageReq 对齐） */
export interface GetSysApiPageParams {
  pageIndex: number;
  pageSize: number;
  title?: string;
  path?: string;
  action?: string;
  parentId?: string;
  type?: string;
}

/**
 * 获取系统接口分页列表
 * 对接 go-admin GET /api/v1/sys-api
 */
export async function getSysApiPage(
  params: GetSysApiPageParams,
): Promise<SysApiPageResult> {
  const data = await requestClient.get<SysApiPageResult>('/v1/sys-api', {
    params,
  });
  return {
    list: data?.list ?? [],
    count: data?.count ?? 0,
    pageIndex: data?.pageIndex ?? params.pageIndex,
    pageSize: data?.pageSize ?? params.pageSize,
  };
}

/**
 * 获取系统接口详情
 * 对接 go-admin GET /api/v1/sys-api/:id
 */
export async function getSysApiDetail(id: number): Promise<SysApiItem> {
  return requestClient.get<SysApiItem>(`/v1/sys-api/${id}`);
}

/** 更新接口请求体（与 dto.SysApiUpdateReq 对齐，不含 id） */
export interface UpdateSysApiData {
  handle?: string;
  title?: string;
  path?: string;
  type?: string;
  action?: string;
}

/**
 * 更新系统接口
 * 对接 go-admin PUT /api/v1/sys-api/:id
 */
export async function updateSysApi(
  id: number,
  data: UpdateSysApiData,
): Promise<unknown> {
  return requestClient.put(`/v1/sys-api/${id}`, data);
}

/** 分页响应（内部用，getSysApiList 解包） */
interface SysApiPageData {
  list?: SysApiItem[];
  count?: number;
  pageIndex?: number;
  pageSize?: number;
}

/**
 * 获取系统接口列表（用于菜单关联接口多选）
 * 对接 go-admin GET /api/v1/sys-api
 */
export async function getSysApiList(): Promise<SysApiItem[]> {
  const data = await requestClient.get<SysApiPageData>('/v1/sys-api', {
    params: { pageIndex: 1, pageSize: 999 },
  });
  return Array.isArray(data?.list) ? data.list : [];
}
