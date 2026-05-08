import { requestClient } from '#/api/request';

/** SKU 项（与后端 models.Sku 对齐） */
export interface SkuItem {
  skuId: number;
  spuId: number;
  skuCode: string;
  skuName?: string;
  spec?: string;
  unit?: string;
  price: number;
  status: number; // 1=禁用 2=启用
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface SkuPageResult {
  list: SkuItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface GetSkuPageParams {
  pageIndex?: number;
  pageSize?: number;
  spuId?: number;
  skuCode?: string;
  skuName?: string;
  status?: number;
}

export interface CreateSkuData {
  spuId: number;
  skuCode: string;
  skuName?: string;
  spec?: string;
  unit?: string;
  price?: number;
  status?: number;
}

export type UpdateSkuData = CreateSkuData;

/** GET /api/v1/sku */
export async function getSkuPage(
  params: GetSkuPageParams = {},
): Promise<SkuPageResult> {
  return requestClient.get<SkuPageResult>('/v1/sku', { params });
}

/** GET /api/v1/sku/:id */
export async function getSkuDetail(id: number): Promise<SkuItem> {
  return requestClient.get<SkuItem>(`/v1/sku/${id}`);
}

/** POST /api/v1/sku */
export async function createSku(data: CreateSkuData): Promise<number> {
  return requestClient.post('/v1/sku', data);
}

/** PUT /api/v1/sku/:id */
export async function updateSku(
  id: number,
  data: UpdateSkuData,
): Promise<number> {
  return requestClient.put(`/v1/sku/${id}`, data);
}

/** DELETE /api/v1/sku */
export async function deleteSku(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sku', { data: { ids } });
}
