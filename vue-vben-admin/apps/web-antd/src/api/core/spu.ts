import { requestClient } from '#/api/request';

/**
 * SPU 列表/详情视图模型（与后端 dto.SpuListItem 对齐）。
 *
 * workflowStatus / workflowTitle 来自 wf_business_binding LEFT JOIN，未提交审核时为空字符串。
 * detailImages 是后端 JSON 字段的字符串原文 — 前端读写时需自行 JSON.parse / JSON.stringify。
 */
export interface SpuItem {
  spuId: number;
  spuCode: string;
  spuName: string;
  categoryId: number;
  brandId: number;
  description?: string;
  mainImageUrl?: string;
  /** 详情图 JSON 字符串，前端解析为 string[] */
  detailImages?: string;
  status: number; // 1=草稿 2=审核中 3=已通过 4=已驳回 5=已下线
  workflowInstanceId: number;
  submittedAt?: null | string;
  approvedAt?: null | string;
  creatorId?: number;
  deptId?: number;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
  // 派生字段（来自 wf_business_binding LEFT JOIN）
  workflowStatus?: string;
  workflowTitle?: string;
}

export interface SpuPageResult {
  list: SpuItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface GetSpuPageParams {
  pageIndex?: number;
  pageSize?: number;
  spuCode?: string;
  spuName?: string;
  categoryId?: number;
  brandId?: number;
  status?: number;
}

export interface CreateSpuData {
  spuCode: string;
  spuName: string;
  categoryId?: number;
  brandId?: number;
  description?: string;
  mainImageUrl?: string;
  /** 详情图 JSON 字符串（如 JSON.stringify(string[])） */
  detailImages?: string;
  status?: number;
}

export type UpdateSpuData = CreateSpuData;

export interface SubmitSpuData {
  /** 流程定义ID（可选；为空则按 definition_key='spu_create_review' 查找启用版本） */
  definitionId?: number;
  remark?: string;
}

export interface SubmitSpuResult {
  spuId: number;
  instanceId: number;
}

/** GET /api/v1/spu */
export async function getSpuPage(
  params: GetSpuPageParams = {},
): Promise<SpuPageResult> {
  return requestClient.get<SpuPageResult>('/v1/spu', { params });
}

/** GET /api/v1/spu/:id */
export async function getSpuDetail(id: number): Promise<SpuItem> {
  return requestClient.get<SpuItem>(`/v1/spu/${id}`);
}

/** POST /api/v1/spu */
export async function createSpu(data: CreateSpuData): Promise<number> {
  return requestClient.post('/v1/spu', data);
}

/** PUT /api/v1/spu/:id */
export async function updateSpu(
  id: number,
  data: UpdateSpuData,
): Promise<number> {
  return requestClient.put(`/v1/spu/${id}`, data);
}

/** DELETE /api/v1/spu */
export async function deleteSpu(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/spu', { data: { ids } });
}

/** POST /api/v1/spu/:id/submit — 提交审核 */
export async function submitSpu(
  id: number,
  data: SubmitSpuData = {},
): Promise<SubmitSpuResult> {
  return requestClient.post<SubmitSpuResult>(`/v1/spu/${id}/submit`, data);
}
