/**
 * 统一请求客户端配置
 * 
 * 当前状态（2026-03-23 核实）：
 * 1. baseRequestClient: ✅ 可用
 *    - 返回: axios 完整响应对象
 *    - 使用: res.data 获取 {code, data, requestId}
 *    - 适用: 所有 go-admin 接口
 * 
 * 2. requestClient: ❌ 配置混乱，不建议使用
 *    - responseReturn: 'data' 与拦截器逻辑冲突
 *    - 返回结构不确定，可能返回 {code, data} 或 data.data 或 undefined
 *    - 待修复后再启用
 * 
 * 推荐做法: 统一使用 baseRequestClient，手动处理 res.data
 */
import type { RequestClientOptions } from '@vben/request';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  defaultResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';

import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';

import { refreshTokenApi } from './core';

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

/**
 * 创建 requestClient（自动提取 data）
 */
function createRequestClient(baseURL: string, options?: RequestClientOptions) {
  const client = new RequestClient({
    ...options,
    baseURL,
  });

  async function doReAuthenticate() {
    console.warn('Access token or refresh token is invalid or expired. ');
    const { useAccessStore } = await import('@vben/stores');
    const accessStore = useAccessStore();
    const authStore = useAuthStore();
    accessStore.setAccessToken(null);
    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      await authStore.logout();
    }
  }

  async function doRefreshToken() {
    const { useAccessStore } = await import('@vben/stores');
    const accessStore = useAccessStore();
    const resp = await refreshTokenApi();
    const body = resp.data as { data?: string; token?: string };
    const newToken =
      (body && (body.token ?? body.data)) || (typeof body === 'string' ? body : '');
    accessStore.setAccessToken(newToken);
    return newToken;
  }

  function formatToken(token: null | string) {
    return token ? `Bearer ${token}` : null;
  }

  // 认证拦截器 - 动态获取 token
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const { useAccessStore } = await import('@vben/stores');
      const accessStore = useAccessStore();
      config.headers.Authorization = formatToken(accessStore.accessToken);
      config.headers['Accept-Language'] = preferences.app.locale;
      return config;
    },
  });

  // 响应拦截器 - 处理业务响应体 {code, data, msg, requestId}
  // 注意：responseReturn: 'data' 已提取，这里直接处理业务对象
  client.addResponseInterceptor({
    fulfilled: (businessData) => {
      // businessData 已经是 {code, data, msg, requestId}
      if (businessData && typeof businessData === 'object' && 'code' in businessData) {
        if (businessData.code === 200) {
          return businessData.data;  // 返回业务 data
        }
        // code 不为 200，抛出错误
        throw new Error(businessData.msg || `请求失败: code ${businessData.code}`);
      }
      // 非标准格式，原样返回
      return businessData;
    },
  });

  // token 过期处理
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: preferences.app.enableRefreshToken,
      formatToken,
    }),
  );

  // 错误处理
  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      const responseData = error?.response?.data ?? {};
      const errorMessage = responseData?.error ?? responseData?.message ?? '';
      message.error(errorMessage || msg);
    }),
  );

  return client;
}

// 自动提取 data 的客户端（推荐用于标准接口）
// 自动提取 data 的客户端（已修复，可用）
// 流程: axios response → responseReturn 提取 response.data → 拦截器处理 {code, data} → 返回 data
// 使用: const menus = await requestClient.get('/v1/menurole'); // 直接返回菜单数组
export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

// 原始响应客户端（手动处理 code/data，用于特殊场景）
// 返回: axios 完整响应对象，需通过 res.data 获取 {code, data, requestId}
// 使用: const res = await baseRequestClient.get('/v1/xxx'); const data = res.data;
export const baseRequestClient = new RequestClient({ baseURL: apiURL });

// 给 baseRequestClient 也加上认证拦截器
baseRequestClient.addRequestInterceptor({
  fulfilled: async (config) => {
    const { useAccessStore } = await import('@vben/stores');
    const accessStore = useAccessStore();
    const token = accessStore.accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
});