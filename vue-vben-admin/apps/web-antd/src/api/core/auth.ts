import { baseRequestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值（供 store 使用） */
  export interface LoginResult {
    accessToken: string;
  }

  /** go-admin 登录接口原始响应 */
  export interface GoAdminLoginResponse {
    code: number;
    token?: string;
    expire?: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

/**
 * 登录（对接 go-admin POST /api/v1/login）
 * 成功条件：HTTP 成功且 body.code === 200 且存在 body.token
 */
export async function loginApi(data: AuthApi.LoginParams): Promise<AuthApi.LoginResult> {
  const res = await baseRequestClient.post<AuthApi.GoAdminLoginResponse>(
    '/v1/login',
    {
      username: data.username,
      password: data.password,
      code: '0',
      uuid: '0',
    },
  );
  const body = res.data;
  if (body.code !== 200 || !body.token) {
    throw new Error((body as any).msg ?? '登录失败');
  }
  return { accessToken: body.token };
}

/**
 * 刷新 accessToken（对接 go-admin GET /api/v1/refresh_token）
 */
export async function refreshTokenApi() {
  return baseRequestClient.get<AuthApi.RefreshTokenResult>('/v1/refresh_token', {
    withCredentials: true,
  });
}

/**
 * 退出登录（对接 go-admin POST /api/v1/logout）
 */
export async function logoutApi() {
  return baseRequestClient.post('/v1/logout', undefined, {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码（最小兼容：不再请求后端，直接返回固定权限码，保证登录后可进入首页）
 */
export async function getAccessCodesApi(): Promise<string[]> {
  return ['*:*:*'];
}
