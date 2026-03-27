import { getApiRaw, getHttpRaw, postApiRaw, postHttpRaw } from '#/api/request';

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
    data?: string;
    token?: string;
  }

  export interface CaptchaResult {
    code: number;
    data: string;
    id: string;
    msg: string;
    answer?: string;
  }
}

/**
 * 获取验证码
 * 改用 baseRequestClient，返回原始响应 {code, data, id, msg}
 */
export async function getCaptchaApi(): Promise<AuthApi.CaptchaResult> {
  const res = await getHttpRaw<AuthApi.CaptchaResult>('/v1/captcha');
  const body = res.data;
  if (body.code !== 200 || !body.data || !body.id) {
    throw new Error(body.msg || '获取验证码失败');
  }
  return body;
}

/**
 * 登录
 * 改用 baseRequestClient，返回原始响应，手动判断 code 和 token
 */
export async function loginApi(
  data: AuthApi.LoginParams,
): Promise<AuthApi.LoginResult> {
  const res = await postHttpRaw<AuthApi.GoAdminLoginResponse>(
    '/v1/login',
    data,
  );
  const body = res.data;
  if (body.code !== 200 || !body.token) {
    throw new Error(body.msg || '登录失败');
  }
  return { accessToken: body.token };
}

/**
 * 刷新 accessToken
 */
export async function refreshTokenApi(): Promise<
  import('#/api/request').ApiRawResponse<AuthApi.RefreshTokenResult>
> {
  return getApiRaw<AuthApi.RefreshTokenResult>('/v1/refresh_token');
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return postApiRaw<null>('/v1/logout');
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi(): Promise<string[]> {
  return ['*:*:*'];
}
