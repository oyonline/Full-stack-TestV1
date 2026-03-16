import { requestClient } from '#/api/request';

/** 岗位列表项（与后端 models.SysPost 对齐） */
export interface SysPostItem {
  postId: number;
  postName: string;
  postCode: string;
  sort: number;
  status: number; // 1=停用, 2=启用
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 岗位分页响应（requestClient 解包后为 data 内容） */
export interface SysPostPageResult {
  list: SysPostItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 岗位列表查询参数 */
interface GetPostPageParams {
  pageIndex?: number;
  pageSize?: number;
  postName?: string;
  postCode?: string;
  status?: number;
}

/**
 * 获取岗位分页列表
 * 对接 go-admin GET /api/v1/post
 */
export async function getPostPage(
  params: GetPostPageParams = {},
): Promise<SysPostPageResult> {
  return requestClient.get<SysPostPageResult>('/v1/post', { params });
}

/**
 * 获取岗位详情
 * 对接 go-admin GET /api/v1/post/:id
 */
export async function getPostDetail(id: number): Promise<SysPostItem> {
  return requestClient.get<SysPostItem>(`/v1/post/${id}`);
}

/** 新增岗位请求体 */
interface CreatePostData {
  postName: string;
  postCode: string;
  sort?: number;
  status?: number;
  remark?: string;
}

/**
 * 新增岗位
 * 对接 go-admin POST /api/v1/post
 */
export async function createPost(data: CreatePostData): Promise<number> {
  return requestClient.post('/v1/post', data);
}

/** 更新岗位请求体 */
interface UpdatePostData {
  postName?: string;
  postCode?: string;
  sort?: number;
  status?: number;
  remark?: string;
}

/**
 * 更新岗位
 * 对接 go-admin PUT /api/v1/post/:id
 */
export async function updatePost(
  id: number,
  data: UpdatePostData,
): Promise<number> {
  return requestClient.put(`/v1/post/${id}`, data);
}

/**
 * 删除岗位（支持批量）
 * 对接 go-admin DELETE /api/v1/post
 */
export async function deletePost(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/post', { data: { ids } });
}
