import type { UserInfo } from '@vben/types';

import { useAccessStore } from '@vben/stores';

import { baseRequestClient } from '#/api/request';

/** go-admin GET /api/v1/getinfo 原始 data 结构 */
interface GoAdminGetInfoData {
  avatar?: string;
  buttons?: string[];
  code?: number;
  deptId?: number;
  introduction?: string;
  name?: string;
  permissions?: string[];
  roles?: string[];
  userId?: number;
  userName?: string;
}

/** go-admin getinfo 响应体 */
interface GoAdminGetInfoResponse {
  code: number;
  data?: GoAdminGetInfoData;
  msg?: string;
}

/**
 * 获取当前用户信息（对接 go-admin GET /api/v1/getinfo）
 * 成功条件：code === 200，业务数据在 data 中；在 api 层做字段映射，保持 store 不变。
 */
export async function getUserInfoApi(): Promise<UserInfo> {
  const accessStore = useAccessStore();
  const token = accessStore.accessToken;
  const res = await baseRequestClient.get<GoAdminGetInfoResponse>('/v1/getinfo', {
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
    roles: d.roles ?? [],
    desc: d.introduction ?? '',
    homePath: '/',
    token: '',
    ...(d.permissions && { permissions: d.permissions }),
    ...(d.buttons && { buttons: d.buttons }),
  } as UserInfo;
}
