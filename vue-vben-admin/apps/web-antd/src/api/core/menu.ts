import type { MenuRecordRaw } from '@vben-core/typings';

import { baseRequestClient } from '#/api/request';

/**
 * 获取用户所有菜单
 * 使用 baseRequestClient 显式处理 res.data，确保稳定返回菜单数组
 */
export async function getAllMenusApi(): Promise<MenuRecordRaw[]> {
  const res = await baseRequestClient.get('/v1/menurole');
  const responseData = res.data as any;
  
  if (responseData?.code === 200 && Array.isArray(responseData.data)) {
    return responseData.data;
  }
  
  return [];
}