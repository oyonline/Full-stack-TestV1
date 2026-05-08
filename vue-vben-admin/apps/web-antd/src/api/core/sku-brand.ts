import { requestClient } from '#/api/request';

/** SKU 品牌项（与后端 models.SkuBrand 对齐） */
export interface SkuBrandItem {
  brandId: number;
  brandName: string;
  brandLogoUrl?: string;
  sort: number;
  status: number; // 1=禁用 2=启用
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface SkuBrandPageResult {
  list: SkuBrandItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface GetSkuBrandPageParams {
  pageIndex?: number;
  pageSize?: number;
  brandName?: string;
  status?: number;
}

export interface CreateSkuBrandData {
  brandName: string;
  brandLogoUrl?: string;
  sort?: number;
  status?: number;
}

export type UpdateSkuBrandData = CreateSkuBrandData;

/** GET /api/v1/sku-brand */
export async function getSkuBrandPage(
  params: GetSkuBrandPageParams = {},
): Promise<SkuBrandPageResult> {
  return requestClient.get<SkuBrandPageResult>('/v1/sku-brand', { params });
}

/** GET /api/v1/sku-brand/:id */
export async function getSkuBrandDetail(id: number): Promise<SkuBrandItem> {
  return requestClient.get<SkuBrandItem>(`/v1/sku-brand/${id}`);
}

/** POST /api/v1/sku-brand */
export async function createSkuBrand(
  data: CreateSkuBrandData,
): Promise<number> {
  return requestClient.post('/v1/sku-brand', data);
}

/** PUT /api/v1/sku-brand/:id */
export async function updateSkuBrand(
  id: number,
  data: UpdateSkuBrandData,
): Promise<number> {
  return requestClient.put(`/v1/sku-brand/${id}`, data);
}

/** DELETE /api/v1/sku-brand */
export async function deleteSkuBrand(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sku-brand', { data: { ids } });
}
