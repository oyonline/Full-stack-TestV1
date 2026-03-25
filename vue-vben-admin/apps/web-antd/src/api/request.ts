import type { RequestClientConfig, RequestResponse } from '@vben/request';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';

import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';

import { refreshTokenApi } from './core';

export interface ApiResponse<T> {
  code: number;
  data: T;
  msg?: string;
  requestId?: string;
}

export type ApiRawResponse<T> = RequestResponse<ApiResponse<T>>;
export type HttpRawResponse<T> = RequestResponse<T>;

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

function formatToken(token: null | string) {
  return token ? `Bearer ${token}` : null;
}

function withDefaultHeaders(config: RequestClientConfig = {}) {
  return config;
}

async function doReAuthenticate() {
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
  const response = await refreshTokenApi();
  const tokenPayload = response.data.data;
  const newToken = tokenPayload?.token ?? tokenPayload?.data;
  accessStore.setAccessToken(newToken ?? null);
  return newToken ?? '';
}

function applyRequestInterceptors(client: RequestClient) {
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const { useAccessStore } = await import('@vben/stores');
      const accessStore = useAccessStore();
      config.headers = config.headers ?? {};
      config.headers.Authorization = formatToken(accessStore.accessToken);
      config.headers['Accept-Language'] = preferences.app.locale;
      return config;
    },
  });
}

function applySharedResponseInterceptors(client: RequestClient) {
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: preferences.app.enableRefreshToken,
      formatToken,
    }),
  );

  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      const responseData = error?.response?.data ?? {};
      const errorMessage = responseData?.error ?? responseData?.message ?? '';
      message.error(errorMessage || msg);
    }),
  );
}

function unwrapApiResponse<T>(response: ApiRawResponse<T>): T {
  const body = response.data;
  if (body.code === 200) {
    return body.data;
  }
  throw new Error(body.msg || `请求失败: code ${body.code}`);
}

const baseRequestClient = new RequestClient({ baseURL: apiURL });
applyRequestInterceptors(baseRequestClient);
applySharedResponseInterceptors(baseRequestClient);

const requestClient = {
  async delete<T>(url: string, config?: RequestClientConfig): Promise<T> {
    const response = await baseRequestClient.delete<ApiRawResponse<T>>(
      url,
      withDefaultHeaders(config),
    );
    return unwrapApiResponse(response);
  },
  async get<T>(url: string, config?: RequestClientConfig): Promise<T> {
    const response = await baseRequestClient.get<ApiRawResponse<T>>(
      url,
      withDefaultHeaders(config),
    );
    return unwrapApiResponse(response);
  },
  async post<T>(
    url: string,
    data?: any,
    config?: RequestClientConfig,
  ): Promise<T> {
    const response = await baseRequestClient.post<ApiRawResponse<T>>(
      url,
      data,
      withDefaultHeaders(config),
    );
    return unwrapApiResponse(response);
  },
  async put<T>(
    url: string,
    data?: any,
    config?: RequestClientConfig,
  ): Promise<T> {
    const response = await baseRequestClient.put<ApiRawResponse<T>>(
      url,
      data,
      withDefaultHeaders(config),
    );
    return unwrapApiResponse(response);
  },
};

async function getApiRaw<T>(
  url: string,
  config?: RequestClientConfig,
): Promise<ApiRawResponse<T>> {
  return baseRequestClient.get<ApiRawResponse<T>>(
    url,
    withDefaultHeaders(config),
  );
}

async function postApiRaw<T>(
  url: string,
  data?: any,
  config?: RequestClientConfig,
): Promise<ApiRawResponse<T>> {
  return baseRequestClient.post<ApiRawResponse<T>>(
    url,
    data,
    withDefaultHeaders(config),
  );
}

async function getHttpRaw<T>(
  url: string,
  config?: RequestClientConfig,
): Promise<HttpRawResponse<T>> {
  return baseRequestClient.get<HttpRawResponse<T>>(
    url,
    withDefaultHeaders(config),
  );
}

async function postHttpRaw<T>(
  url: string,
  data?: any,
  config?: RequestClientConfig,
): Promise<HttpRawResponse<T>> {
  return baseRequestClient.post<HttpRawResponse<T>>(
    url,
    data,
    withDefaultHeaders(config),
  );
}

export {
  baseRequestClient,
  getApiRaw,
  getHttpRaw,
  postApiRaw,
  postHttpRaw,
  requestClient,
};
