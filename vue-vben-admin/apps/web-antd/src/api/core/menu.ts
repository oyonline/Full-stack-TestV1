import type { MenuRecordRaw } from '@vben/types';

import { getApiRaw } from '#/api/request';

/**
 * 获取用户所有菜单
 * 使用 baseRequestClient 显式处理 res.data，确保稳定返回菜单数组
 */
export async function getAllMenusApi(): Promise<MenuRecordRaw[]> {
  const res = await getApiRaw<MenuRecordRaw[]>('/v1/menurole');
  return Array.isArray(res.data.data) ? res.data.data : [];
}
