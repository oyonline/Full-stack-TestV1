import { expect, type APIRequestContext, type Page, type Route } from '@playwright/test';

export interface SmokeAccount {
  password: string;
  username: string;
}

export interface LoginApiResult {
  accessToken: string;
}

interface CaptchaPayload {
  answer?: string;
  code?: number;
  data?: string;
  id?: string;
  msg?: string;
}

interface CaptchaFixture {
  payload: CaptchaPayload;
  rawBody: string;
  responseHeaders: Record<string, string>;
  status: number;
}

const DEFAULT_ACCOUNT: SmokeAccount = {
  password: process.env.PLAYWRIGHT_ADMIN_PASSWORD || '123456',
  username: process.env.PLAYWRIGHT_ADMIN_USERNAME || 'admin',
};

const DEFAULT_API_BASE_URL =
  process.env.PLAYWRIGHT_API_URL || 'http://127.0.0.1:10082';

export function getSmokeAccount(): SmokeAccount {
  return DEFAULT_ACCOUNT;
}

export function getApiBaseUrl() {
  return DEFAULT_API_BASE_URL;
}

function attachDebugListeners(page: Page) {
  const key = '__smokeDebugAttached';
  if ((page as Page & { [key]?: boolean })[key]) {
    return;
  }

  page.on('console', (msg) => {
    if (msg.type() === 'error' || msg.type() === 'warning') {
      console.log(`[pw:${msg.type()}] ${msg.text()}`);
    }
  });
  page.on('pageerror', (error) => {
    console.log(`[pw:pageerror] ${error.message}`);
  });
  page.on('requestfailed', (request) => {
    console.log(
      `[pw:requestfailed] ${request.method()} ${request.url()} ${request.failure()?.errorText || ''}`,
    );
  });
  page.on('requestfinished', async (request) => {
    const url = request.url();
    if (
      url.includes('/@vite/client') ||
      url.includes('/src/main.ts') ||
      url.includes('/api/v1/app-config')
    ) {
      console.log(`[pw:requestfinished] ${request.method()} ${url}`);
    }
  });

  (page as Page & { [key]?: boolean })[key] = true;
}

async function attachTestCaptchaRoute(
  page: Page,
  captchaFixture: CaptchaFixture,
) {
  const key = '__smokeCaptchaRouteAttached';
  if ((page as Page & { [key]?: boolean })[key]) {
    return;
  }

  await page.route('**/api/v1/captcha', async (route: Route) => {
    await route.fulfill({
      body: captchaFixture.rawBody,
      contentType: captchaFixture.responseHeaders['content-type'],
      headers: captchaFixture.responseHeaders,
      status: captchaFixture.status,
    });
  });

  (page as Page & { [key]?: boolean })[key] = true;
}

async function fillFirstVisible(
  page: Page,
  selectors: string[],
  value: string,
  timeout = 30_000,
) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    for (const selector of selectors) {
      const locator = page.locator(selector).first();
      if ((await locator.count()) > 0 && (await locator.isVisible())) {
        await locator.fill(value);
        return;
      }
    }

    await page.waitForTimeout(250);
  }

  throw new Error(
    `Unable to find login field for selectors: ${selectors.join(', ')}`,
  );
}

async function clickFirstVisible(page: Page, selectors: string[], timeout = 30_000) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    for (const selector of selectors) {
      const locator = page.locator(selector).first();
      if ((await locator.count()) > 0 && (await locator.isVisible())) {
        await locator.click();
        return;
      }
    }

    await page.waitForTimeout(250);
  }

  throw new Error(
    `Unable to find clickable element for selectors: ${selectors.join(', ')}`,
  );
}

async function fetchCaptcha(
  request: APIRequestContext,
): Promise<CaptchaFixture> {
  const response = await request.get(`${DEFAULT_API_BASE_URL}/api/v1/captcha`, {
    headers: {
      'X-E2E-Test': 'true',
    },
  });
  expect(response.ok()).toBeTruthy();
  const rawBody = await response.text();
  const payload = JSON.parse(rawBody) as CaptchaPayload;
  expect(payload.code).toBe(200);
  expect(payload.id).toBeTruthy();
  expect(payload.answer).toBeTruthy();
  return {
    payload,
    rawBody,
    responseHeaders: response.headers(),
    status: response.status(),
  };
}

export async function loginByApi(
  request: APIRequestContext,
  account: SmokeAccount = DEFAULT_ACCOUNT,
): Promise<LoginApiResult> {
  const captchaFixture = await fetchCaptcha(request);
  const response = await request.post(`${DEFAULT_API_BASE_URL}/api/v1/login`, {
    data: {
      code: String(captchaFixture.payload.answer),
      password: account.password,
      username: account.username,
      uuid: captchaFixture.payload.id,
    },
    headers: {
      'Content-Type': 'application/json',
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    code?: number;
    msg?: string;
    token?: string;
  };
  expect(body.code, body.msg || '登录失败').toBe(200);
  expect(body.token).toBeTruthy();
  return {
    accessToken: body.token!,
  };
}

export async function loginByUI(
  page: Page,
  request: APIRequestContext,
  account: SmokeAccount = DEFAULT_ACCOUNT,
) {
  attachDebugListeners(page);
  const captchaFixture = await fetchCaptcha(request);
  await attachTestCaptchaRoute(page, captchaFixture);

  await page.goto('/auth/login');
  await fillFirstVisible(
    page,
    ['input[placeholder="请输入登录账号"]'],
    account.username,
  );
  await fillFirstVisible(
    page,
    ['input[placeholder="请输入密码"]', 'input[placeholder="密码"]'],
    account.password,
  );
  await fillFirstVisible(
    page,
    ['input[placeholder="请输入验证码"]'],
    String(captchaFixture.payload.answer),
  );
  await clickFirstVisible(page, [
    'button:has-text("登录")',
    'button[aria-label="login"]',
    'button',
  ]);

  await expect(page).not.toHaveURL(/\/auth\/login/);
  await expect(page.getByRole('heading', { name: '首页' })).toBeVisible({
    timeout: 20_000,
  });
}
