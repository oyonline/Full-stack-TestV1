import { requestClient } from '#/api/request';

/** 部门列表项（与后端 models.SysDept 对齐） */
export interface SysDeptItem {
  deptId: number;
  parentId: number;
  deptPath?: string;
  deptName: string;
  sort: number;
  leader?: string;
  phone?: string;
  email?: string;
  status: number; // 0=禁用, 1=启用
  children?: SysDeptItem[];
  createdAt?: string;
  updatedAt?: string;
}

/** 部门列表查询参数 */
export interface SysDeptQuery {
  deptName?: string;
  status?: number;
  parentId?: number;
  deptId?: number;
}

/** 部门表单数据（新增/修改） */
export interface SysDeptForm {
  parentId: number;
  deptName: string;
  sort: number;
  leader?: string;
  phone?: string;
  email?: string;
  status: number;
}

/** 部门树形选择器节点（与后端 dto.DeptLabel 对齐） */
export interface DeptLabel {
  id: number;
  label: string;
  children?: DeptLabel[];
}

/**
 * 获取部门列表（树形结构）
 * 对接 go-admin GET /api/v1/dept
 */
export async function getDeptListApi(
  params?: SysDeptQuery,
): Promise<SysDeptItem[]> {
  return requestClient.get<SysDeptItem[]>('/v1/dept', { params });
}

/**
 * 获取部门详情
 * 对接 go-admin GET /api/v1/dept/:id
 */
export async function getDeptDetailApi(id: number): Promise<SysDeptItem> {
  return requestClient.get<SysDeptItem>(`/v1/dept/${id}`);
}

/**
 * 新增部门
 * 对接 go-admin POST /api/v1/dept
 */
export async function createDeptApi(data: SysDeptForm): Promise<number> {
  return requestClient.post('/v1/dept', data);
}

/**
 * 更新部门
 * 对接 go-admin PUT /api/v1/dept/:id
 */
export async function updateDeptApi(
  id: number,
  data: SysDeptForm,
): Promise<number> {
  return requestClient.put(`/v1/dept/${id}`, data);
}

/**
 * 删除部门（支持批量）
 * 对接 go-admin DELETE /api/v1/dept
 */
export async function deleteDeptApi(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/dept', { data: { ids } });
}

/**
 * 获取部门树（用于选择器）
 * 对接 go-admin GET /api/v1/deptTree
 */
export async function getDeptTreeApi(): Promise<DeptLabel[]> {
  return requestClient.get<DeptLabel[]>('/v1/deptTree');
}
