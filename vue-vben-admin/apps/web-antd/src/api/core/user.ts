import type { UserInfo } from '@vben/types';

import { useAccessStore } from '@vben/stores';

import { requestClient } from '#/api/request';
import { getApiRaw } from '#/api/request';

/** go-admin GET /api/v1/getinfo 原始 data 结构 */
interface GoAdminGetInfoData {
  avatar?: string;
  buttons?: string[];
  code?: number;
  deptId?: number;
  introduction?: string;
  name?: string;
  permissions?: string[];
  primaryRoleId?: number;
  primaryRoleKey?: string;
  primaryRoleName?: string;
  roleIds?: number[];
  roleKeys?: string[];
  roleNames?: string[];
  roles?: string[];
  userId?: number;
  userName?: string;
}

/**
 * 获取当前用户信息（对接 go-admin GET /api/v1/getinfo）
 * 成功条件：code === 200，业务数据在 data 中；在 api 层做字段映射，保持 store 不变。
 */
export async function getUserInfoApi(): Promise<UserInfo> {
  const accessStore = useAccessStore();
  const token = accessStore.accessToken;
  const res = await getApiRaw<GoAdminGetInfoData>('/v1/getinfo', {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  const body = res.data;
  if (body.code !== 200 || !body.data) {
    throw new Error(body.msg ?? '获取用户信息失败');
  }
  const d = body.data;
  return {
    username: d.userName ?? '',
    realName: d.name ?? '',
    avatar: d.avatar ?? '',
    userId: d.userId != null ? String(d.userId) : '',
    roles: d.primaryRoleName ? [d.primaryRoleName] : (d.roles ?? []),
    desc: d.introduction ?? '',
    homePath: '/home',
    token: '',
    ...(d.primaryRoleId != null && { primaryRoleId: d.primaryRoleId }),
    ...(d.primaryRoleName && { primaryRoleName: d.primaryRoleName }),
    ...(d.roleIds && { roleIds: d.roleIds }),
    ...(d.roleNames && { roleNames: d.roleNames }),
    ...(d.permissions && { permissions: d.permissions }),
    ...(d.buttons && { buttons: d.buttons }),
  } as UserInfo;
}

// --------------- 用户管理 CRUD（/v1/sys-user） ---------------

/** 部门简要（列表中的 dept 关联） */
export interface SysUserDeptRef {
  deptId?: number;
  deptName?: string;
}

export interface SysUserRoleRef {
  roleId: number;
  roleKey?: string;
  roleName?: string;
}

/** 用户列表项（与后端 models.SysUser 对齐） */
export interface SysUserItem {
  userId: number;
  username: string;
  nickName: string;
  phone: string;
  roleId: number;
  primaryRoleId?: number;
  roleIds?: number[];
  roles?: SysUserRoleRef[];
  avatar?: string;
  sex?: string;
  email?: string;
  deptId: number;
  postId: number;
  remark?: string;
  status: string; // "2" 启用/正常 "1" 停用/关闭
  dept?: SysUserDeptRef;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 用户分页响应 */
export interface SysUserPageResult {
  list: SysUserItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 用户列表查询参数 */
export interface GetSysUserPageParams {
  pageIndex?: number;
  pageSize?: number;
  username?: string;
  nickName?: string;
  phone?: string;
  status?: string;
  roleId?: string;
  roleIds?: string;
  postId?: string;
  deptId?: string;
}

/**
 * 获取用户分页列表
 * 对接 go-admin GET /api/v1/sys-user
 */
export async function getSysUserPage(
  params: GetSysUserPageParams = {},
): Promise<SysUserPageResult> {
  return requestClient.get<SysUserPageResult>('/v1/sys-user', { params });
}

/**
 * 获取用户详情
 * 对接 go-admin GET /api/v1/sys-user/:id
 */
export async function getSysUserDetail(userId: number): Promise<SysUserItem> {
  return requestClient.get<SysUserItem>(`/v1/sys-user/${userId}`);
}

/** 新增用户请求体（与 dto.SysUserInsertReq 对齐） */
export interface CreateSysUserData {
  username: string;
  password: string;
  nickName: string;
  phone: string;
  email: string;
  deptId: number;
  primaryRoleId: number;
  roleIds: number[];
  roleId?: number;
  postId?: number;
  sex?: string;
  remark?: string;
  status?: string;
}

/**
 * 新增用户
 * 对接 go-admin POST /api/v1/sys-user
 */
export async function createSysUser(data: CreateSysUserData): Promise<number> {
  return requestClient.post('/v1/sys-user', data);
}

/** 更新用户请求体（与 dto.SysUserUpdateReq 对齐，PUT 无 path :id，body 带 userId） */
export interface UpdateSysUserData {
  userId: number;
  username: string;
  nickName: string;
  phone: string;
  email: string;
  deptId: number;
  primaryRoleId: number;
  roleIds: number[];
  roleId?: number;
  postId?: number;
  sex?: string;
  remark?: string;
  status?: string;
}

/**
 * 更新用户
 * 对接 go-admin PUT /api/v1/sys-user（无 path :id，body 含 userId）
 */
export async function updateSysUser(data: UpdateSysUserData): Promise<number> {
  return requestClient.put('/v1/sys-user', data);
}

/**
 * 删除用户（支持批量）
 * 对接 go-admin DELETE /api/v1/sys-user，body { ids: number[] }
 */
export async function deleteSysUser(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-user', { data: { ids } });
}

/**
 * 更新用户状态
 * 对接 go-admin PUT /api/v1/user/status
 */
export async function updateSysUserStatus(
  userId: number,
  status: string,
): Promise<void> {
  return requestClient.put('/v1/user/status', { userId, status });
}
