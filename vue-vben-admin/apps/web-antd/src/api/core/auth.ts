import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
    code?: string;
    uuid?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  /** go-admin 登录接口原始响应 */
  export interface GoAdminLoginResponse {
    code: number;
    token?: string;
    expire?: string;
    msg?: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }

  export interface CaptchaResult {
    code: number;
    data: string;
    id: string;
    msg: string;
  }
}

/**
 * 获取验证码
 * 改用 baseRequestClient，返回原始响应 {code, data, id, msg}
 */
export async function getCaptchaApi(): Promise<AuthApi.CaptchaResult> {
  const res = await baseRequestClient.get<AuthApi.CaptchaResult>('/v1/captcha');
  return res.data;
}

/**
 * 登录
 * 改用 baseRequestClient，返回原始响应，手动判断 code 和 token
 */
export async function loginApi(data: AuthApi.LoginParams): Promise<AuthApi.LoginResult> {
  // 拿到完整原始响应
  const res = await baseRequestClient.post<AuthApi.GoAdminLoginResponse>('/v1/login', data);
  const body = res.data;

  // 手动判断 code 和 token
  if (body.code !== 200 || !body.token) {
    throw new Error(body.msg || '登录失败');
  }

  // 按项目要求的格式返回
  return { accessToken: body.token };
}

/**
 * 刷新 accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.get<AuthApi.RefreshTokenResult>('/v1/refresh_token');
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/v1/logout');
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi(): Promise<string[]> {
  return ['*:*:*'];
}