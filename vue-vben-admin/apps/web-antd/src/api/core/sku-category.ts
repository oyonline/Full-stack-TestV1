import { requestClient } from '#/api/request';

/** SKU 类目项（与后端 models.SkuCategory 对齐） */
export interface SkuCategoryItem {
  categoryId: number;
  categoryName: string;
  parentId: number;
  level: number;
  sort: number;
  status: number; // 1=禁用 2=启用
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 类目树节点（包含 children，便于前端直接渲染） */
export interface SkuCategoryTreeNode extends SkuCategoryItem {
  children: SkuCategoryTreeNode[];
}

export interface GetSkuCategoryTreeParams {
  categoryName?: string;
  parentId?: number;
  status?: number;
}

export interface CreateSkuCategoryData {
  categoryName: string;
  parentId?: number;
  sort?: number;
  status?: number;
}

export type UpdateSkuCategoryData = CreateSkuCategoryData;

/** GET /api/v1/sku-category — 返回类目树 */
export async function getSkuCategoryTree(
  params: GetSkuCategoryTreeParams = {},
): Promise<SkuCategoryTreeNode[]> {
  return requestClient.get<SkuCategoryTreeNode[]>('/v1/sku-category', {
    params,
  });
}

/** POST /api/v1/sku-category */
export async function createSkuCategory(
  data: CreateSkuCategoryData,
): Promise<number> {
  return requestClient.post('/v1/sku-category', data);
}

/** PUT /api/v1/sku-category/:id */
export async function updateSkuCategory(
  id: number,
  data: UpdateSkuCategoryData,
): Promise<number> {
  return requestClient.put(`/v1/sku-category/${id}`, data);
}

/** DELETE /api/v1/sku-category */
export async function deleteSkuCategory(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sku-category', { data: { ids } });
}
