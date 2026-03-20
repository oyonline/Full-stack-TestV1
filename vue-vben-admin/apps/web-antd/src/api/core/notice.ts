import { requestClient } from '#/api/request';

/** 公告管理列表项 */
export interface NoticeItem {
  id: number;
  title: string;
  content: string;
  type: string;
  status: string;
  sort: number;
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 公告管理分页响应 */
export interface NoticePageResult {
  list: NoticeItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 列表查询参数 */
interface GetNoticePageParams {
  pageIndex?: number;
  pageSize?: number;
  title?: string;
  type?: string;
  status?: string;
}

/**
 * 获取分页列表
 * 对接 go-admin GET /api/v1/sys-notice
 */
export async function getNoticePage(
  params: GetNoticePageParams = {},
): Promise<NoticePageResult> {
  return requestClient.get<NoticePageResult>('/v1/sys-notice', { params });
}

/**
 * 获取详情
 * 对接 go-admin GET /api/v1/sys-notice/:id
 */
export async function getNoticeDetail(id: number): Promise<NoticeItem> {
  return requestClient.get<NoticeItem>(`/v1/sys-notice/${id}`);
}

/** 新增请求体 */
interface CreateNoticeData {
  title: string;
  content?: string;
  type?: string;
  status?: string;
  sort?: number;
  remark?: string;
}

/**
 * 新增
 * 对接 go-admin POST /api/v1/sys-notice
 */
export async function createNotice(data: CreateNoticeData): Promise<number> {
  return requestClient.post('/v1/sys-notice', data);
}

/** 更新请求体 */
interface UpdateNoticeData {
  title?: string;
  content?: string;
  type?: string;
  status?: string;
  sort?: number;
  remark?: string;
}

/**
 * 更新
 * 对接 go-admin PUT /api/v1/sys-notice/:id
 */
export async function updateNotice(
  id: number,
  data: UpdateNoticeData,
): Promise<number> {
  return requestClient.put(`/v1/sys-notice/${id}`, data);
}

/**
 * 删除（支持批量）
 * 对接 go-admin DELETE /api/v1/sys-notice
 */
export async function deleteNotice(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-notice', { data: { ids } });
}
