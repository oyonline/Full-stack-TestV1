/**
 * 角色相关 API：分页列表、详情、增删改、菜单树/部门树（角色管理页 + 用户页下拉）
 */
import { requestClient } from '#/api/request';
import type { DeptLabel } from '#/api/core/dept';

/** 角色列表项（与后端 models.SysRole 对齐） */
export interface SysRoleItem {
  roleId: number;
  roleName: string;
  roleKey?: string;
  status?: string;
  roleSort?: number;
  flag?: string;
  remark?: string;
  admin?: boolean;
  dataScope?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** 角色分页响应 */
export interface SysRolePageResult {
  list: SysRoleItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 菜单树节点（与后端 dto.MenuLabel 对齐，roleMenuTreeselect 返回） */
export interface MenuLabel {
  id: number;
  label: string;
  children?: MenuLabel[];
}

/** 角色菜单树接口响应 */
export interface RoleMenuTreeResult {
  menus: MenuLabel[];
  checkedKeys: number[];
}

/** 角色部门树接口响应（depts 与 api/core/dept DeptLabel 一致） */
export interface RoleDeptTreeResult {
  depts: DeptLabel[];
  checkedKeys: number[];
}

/**
 * 获取角色分页列表
 * 对接 go-admin GET /api/v1/role
 */
export async function getRolePage(params: {
  pageIndex?: number;
  pageSize?: number;
  roleName?: string;
  roleKey?: string;
  status?: string;
} = {}): Promise<SysRolePageResult> {
  return requestClient.get<SysRolePageResult>('/v1/role', { params });
}

/**
 * 获取角色详情
 * 对接 go-admin GET /api/v1/role/:id
 */
export async function getRoleDetail(id: number): Promise<SysRoleItem> {
  return requestClient.get<SysRoleItem>(`/v1/role/${id}`);
}

/** 新增角色请求体（与后端 SysRoleInsertReq 对齐；后端无 json 标签，使用 form 绑定 JSON 时可能需 PascalCase，此处用 camelCase 先联调验证） */
export interface CreateRoleData {
  roleName: string;
  roleKey: string;
  roleSort?: number;
  status?: string;
  flag?: string;
  remark?: string;
  admin?: boolean;
  dataScope?: string;
  menuIds?: number[];
  deptIds?: number[];
}

/**
 * 新增角色
 * 对接 go-admin POST /api/v1/role
 */
export async function createRole(data: CreateRoleData): Promise<number> {
  return requestClient.post<number>('/v1/role', data);
}

/** 更新角色请求体（与后端 SysRoleUpdateReq 对齐，roleId 放 path） */
export interface UpdateRoleData {
  roleName: string;
  roleKey: string;
  roleSort?: number;
  status?: string;
  flag?: string;
  remark?: string;
  admin?: boolean;
  dataScope?: string;
  menuIds?: number[];
  deptIds?: number[];
}

/**
 * 更新角色
 * 对接 go-admin PUT /api/v1/role/:id
 */
export async function updateRole(
  id: number,
  data: UpdateRoleData,
): Promise<number> {
  return requestClient.put<number>(`/v1/role/${id}`, data);
}

/**
 * 删除角色（支持批量）
 * 对接 go-admin DELETE /api/v1/role，body { ids }
 */
export async function deleteRole(ids: number[]): Promise<number[]> {
  return requestClient.delete<number[]>('/v1/role', { data: { ids } });
}

/**
 * 角色状态修改
 * 对接 go-admin PUT /api/v1/role-status，body { roleId, status }
 */
export async function updateRoleStatus(
  roleId: number,
  status: string,
): Promise<void> {
  return requestClient.put('/v1/role-status', { roleId, status });
}

/**
 * 角色菜单树（用于编辑时勾选菜单权限）
 * 对接 go-admin GET /api/v1/roleMenuTreeselect/:roleId
 * 返回 { menus, checkedKeys }
 */
export async function getRoleMenuTreeselect(
  roleId: number,
): Promise<RoleMenuTreeResult> {
  return requestClient.get<RoleMenuTreeResult>(
    `/v1/roleMenuTreeselect/${roleId}`,
  );
}

/**
 * 角色部门树（用于编辑时勾选数据权限）
 * 对接 go-admin GET /api/v1/roleDeptTreeselect/:roleId
 * 返回 { depts, checkedKeys }
 */
export async function getRoleDeptTreeselect(
  roleId: number,
): Promise<RoleDeptTreeResult> {
  return requestClient.get<RoleDeptTreeResult>(
    `/v1/roleDeptTreeselect/${roleId}`,
  );
}
