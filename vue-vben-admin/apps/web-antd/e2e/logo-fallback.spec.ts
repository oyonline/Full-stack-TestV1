import { expect, test } from '@playwright/test';
import type { APIRequestContext } from '@playwright/test';

import { getApiBaseUrl, loginByApi, loginByUI, getSmokeAccount } from './support/auth';

const API_BASE_URL = getApiBaseUrl();

interface SystemSettings {
  branding: {
    appLogo: string;
    appLogoPlaceholderColor: string;
    appName: string;
  };
  uiPreferences: Record<string, unknown>;
}

async function getSystemSettings(
  request: APIRequestContext,
  token: string,
): Promise<SystemSettings> {
  const res = await request.get(`${API_BASE_URL}/api/v1/set-config`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.ok()).toBeTruthy();
  const body = (await res.json()) as { data: SystemSettings };
  return body.data;
}

async function updateSystemSettings(
  request: APIRequestContext,
  token: string,
  settings: SystemSettings,
): Promise<void> {
  const res = await request.put(`${API_BASE_URL}/api/v1/set-config`, {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    data: settings,
  });
  expect(res.ok(), `PUT /api/v1/set-config failed: ${res.status()}`).toBeTruthy();
}

test.describe.serial('logo fallback — 清空 logo 全场景首字 fallback', () => {
  let adminToken = '';
  let originalSettings: SystemSettings;

  test.beforeAll(async ({ request }) => {
    const { accessToken } = await loginByApi(request);
    adminToken = accessToken;

    originalSettings = await getSystemSettings(request, adminToken);

    if (originalSettings.branding.appLogo) {
      await updateSystemSettings(request, adminToken, {
        ...originalSettings,
        branding: { ...originalSettings.branding, appLogo: '' },
      });
    }
  });

  test.afterAll(async ({ request }) => {
    if (originalSettings?.branding.appLogo) {
      const current = await getSystemSettings(request, adminToken);
      await updateSystemSettings(request, adminToken, {
        ...current,
        branding: { ...current.branding, appLogo: originalSettings.branding.appLogo },
      });
    }
  });

  test('admin 侧边栏 logo 显示首字 fallback（无 img 元素）', async ({ page, request }) => {
    await loginByUI(page, request, getSmokeAccount());

    const settings = await getSystemSettings(request, adminToken);
    const appName = settings.branding.appName || 'S';
    const firstChar = ([...appName.replace(/\s+/g, '')][0] ?? 'S').toUpperCase();

    // After login, the home page should be loaded. Wait for it.
    await expect(page).toHaveURL(/\/home/, { timeout: 15_000 });

    // The VbenLogo fallback div (when logo src is empty) has these unique classes.
    // It renders the first character of the system name.
    const fallbackDiv = page.locator(
      'div.flex.items-center.justify-center.rounded-lg.font-semibold.shadow-sm',
    ).first();
    await expect(fallbackDiv).toBeVisible({ timeout: 10_000 });
    await expect(fallbackDiv).toHaveText(firstChar);

    // No logo img inside the VbenLogo fallback area
    const sidebarLogoImg = page.locator(
      'div.flex.items-center.justify-center.rounded-lg.font-semibold.shadow-sm img',
    );
    await expect(sidebarLogoImg).toHaveCount(0);
  });

  test('登录页 logo 显示首字 fallback（未登录状态访问）', async ({ page, request }) => {
    // Each Playwright test has a fresh browser context (no auth state).
    // Navigate directly to the login page without logging in first.
    const settings = await getSystemSettings(request, adminToken);
    const appName = settings.branding.appName || 'S';
    const firstChar = ([...appName.replace(/\s+/g, '')][0] ?? 'S').toUpperCase();

    await page.goto('/auth/login', { waitUntil: 'networkidle' });

    // bootstrap.ts runs syncAppConfigFromBackend(), which fetches the cleared logo,
    // then syncFavicon() and logo preferences are updated.
    // authentication.vue renders the fallback div when logoSrc is empty and appName is set.
    // Fallback div has classes: text-base font-semibold text-white shadow-sm size-[42px]
    const loginFallback = page.locator(
      'div.absolute.top-0.left-0 div.font-semibold',
    ).first();
    await expect(loginFallback).toBeVisible({ timeout: 10_000 });
    await expect(loginFallback).toHaveText(firstChar);

    // Ensure there's no logo img in the absolute-positioned logo area
    const logoImg = page.locator('div.absolute.top-0.left-0 img');
    await expect(logoImg).toHaveCount(0);
  });

  test('favicon href 含 /api/v1/branding/default-logo.png（无 logo 时回退 T5 接口）', async ({
    page,
    request,
  }) => {
    await loginByUI(page, request, getSmokeAccount());

    // Wait for the home page to be ready
    await expect(page).toHaveURL(/\/home/, { timeout: 15_000 });

    // syncFavicon() runs at bootstrap time (after syncAppConfigFromBackend).
    // With empty logo, it sets favicon to the branding endpoint URL.
    const faviconHref = await page.evaluate(() => {
      const el = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
      return el ? el.href : '';
    });

    expect(
      faviconHref,
      `favicon href should contain /api/v1/branding/default-logo.png, got: ${faviconHref}`,
    ).toContain('/api/v1/branding/default-logo.png');
  });

  test('邮件模板 HTML 含 data:image/png;base64,（T7 base64 内嵌验证）', async ({ request }) => {
    const settings = await getSystemSettings(request, adminToken);
    const appName = settings.branding.appName || 'Eposeidon';
    const bg = encodeURIComponent(settings.branding.appLogoPlaceholderColor || '#1d4ed8');

    const res = await request.get(
      `${API_BASE_URL}/api/v1/branding/email-preview?appName=${encodeURIComponent(appName)}&bg=${bg}`,
    );
    expect(res.ok(), `GET /api/v1/branding/email-preview failed: ${res.status()}`).toBeTruthy();

    const html = await res.text();
    expect(
      html,
      'email HTML should contain data:image/png;base64, (T7 base64 inline logo)',
    ).toContain('data:image/png;base64,');
  });
});
