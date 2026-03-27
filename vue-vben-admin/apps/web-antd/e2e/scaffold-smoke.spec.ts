import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

import { getApiBaseUrl, getSmokeAccount, loginByUI } from './support/auth';

const E2E_BASE_URL = process.env.PLAYWRIGHT_BASE_URL || 'http://127.0.0.1:5666';

const schemaPayload = JSON.stringify(
  {
    drawingItems: [],
    formConf: {},
    idGlobal: '100',
    treeNodeId: '100',
    version: '1.1',
  },
  null,
  2,
);

test.describe.serial('scaffold smoke', () => {
  const apiBaseUrl = getApiBaseUrl();

  const clickQueryButton = async (page: Page) => {
    await page
      .locator('#__vben_main_content')
      .getByRole('button', { name: /查\s*询/ })
      .click();
  };

  const readAccessToken = async (page: Page) => {
    const token = await page.evaluate(() => {
      const visited = new Set<unknown>();
      const findToken = (value: unknown): null | string => {
        if (!value || typeof value !== 'object') {
          return null;
        }
        if (visited.has(value)) {
          return null;
        }
        visited.add(value);

        const candidate = value as Record<string, unknown>;
        if (typeof candidate.accessToken === 'string' && candidate.accessToken) {
          return candidate.accessToken;
        }

        for (const nested of Object.values(candidate)) {
          const nestedToken = findToken(nested);
          if (nestedToken) {
            return nestedToken;
          }
        }

        return null;
      };

      for (const key of Object.keys(localStorage)) {
        const raw = localStorage.getItem(key);
        if (!raw) {
          continue;
        }
        try {
          const parsed = JSON.parse(raw);
          const token = findToken(parsed);
          if (token) {
            return token;
          }
        } catch {}
      }
      return null;
    });

    expect(token, '无法从本地持久化状态中读取 accessToken').toBeTruthy();
    return token!;
  };

  test('login and menu navigation work', async ({ page, request }) => {
    await loginByUI(page, request, getSmokeAccount());
    await page
      .getByRole('button', { name: /参数设置 维护系统动态参数与运行开关/ })
      .click();
    await expect(page).toHaveURL(/\/admin\/sys-config\/set/);
    await expect(page.getByRole('heading', { name: '参数设置' })).toBeVisible();
  });

  test('system settings save and login copy update immediately', async ({
    browser,
    page,
    request,
  }) => {
    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/admin/sys-config/set');

    const loginTitleInput = page.getByPlaceholder('欢迎回到管理系统');
    const originalTitle = (await loginTitleInput.inputValue()).trim();
    const nextTitle = `E2E 登录页标题 ${Date.now()}`;

    await loginTitleInput.fill(nextTitle);
    await page.getByRole('button', { name: '保存设置' }).click();
    await expect(
      page.getByText('参数设置已保存，系统展示已同步刷新'),
    ).toBeVisible();

    const guestContext = await browser.newContext();
    const guestPage = await guestContext.newPage();
    await guestPage.goto(`${E2E_BASE_URL}/auth/login`);
    await expect(guestPage.getByText(nextTitle)).toBeVisible();
    await guestContext.close();

    await page.goto('/admin/sys-config/set');
    await loginTitleInput.fill(originalTitle);
    await page.getByRole('button', { name: '保存设置' }).click();
    await expect(
      page.getByText('参数设置已保存，系统展示已同步刷新'),
    ).toBeVisible();
  });

  test('sys-api list loads', async ({ page, request }) => {
    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/admin/sys-api');
    await expect(
      page.getByPlaceholder('请输入接口标题'),
    ).toBeVisible();
    await expect(
      page.locator('#__vben_main_content').getByText('接口列表', { exact: true }),
    ).toBeVisible();
    await expect(
      page.locator('.ant-table-tbody tr:not(.ant-table-measure-row)').first(),
    ).toBeVisible();
  });

  test('form builder opens and supports schema import/export', async ({
    page,
    request,
  }) => {
    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/dev-tools/build');
    await expect(
      page.locator('#__vben_main_content').getByText('表单构建', { exact: true }),
    ).toBeVisible();
    await expect(page.frameLocator('iframe').locator('body')).toBeVisible();

    await page.getByRole('button', { name: '导入 Schema' }).click();
    await page
      .getByPlaceholder('请粘贴完整的 FormSchemaJson')
      .fill(schemaPayload);
    await page
      .locator('.ant-modal-footer .ant-btn-primary')
      .evaluate((button) => (button as HTMLButtonElement).click());
    await expect(
      page.getByPlaceholder('请粘贴完整的 FormSchemaJson'),
    ).toBeHidden();

    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: '下载 Schema' }).click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe('form-schema.json');
  });

  test('code generator supports import, edit, save and preview', async ({
    page,
    request,
  }) => {
    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/dev-tools/gen');
    await expect(
      page.getByPlaceholder('请输入数据库表名'),
    ).toBeVisible();
    await expect(
      page.getByRole('tab', { name: '数据库表' }),
    ).toBeVisible();

    await page.getByPlaceholder('请输入数据库表名').fill('sys_config');
    await clickQueryButton(page);

    const dbRow = page
      .locator('.ant-tabs-tabpane-active .ant-table-tbody tr', { hasText: 'sys_config' })
      .first();
    if ((await dbRow.count()) > 0) {
      await dbRow.getByRole('button', { name: '导入到生成器' }).click();
      await expect(page.getByText('已加入代码生成配置')).toBeVisible();
    }

    await page.getByRole('tab', { name: '生成配置' }).click();
    await page.getByPlaceholder('请输入生成表名').fill('sys_config');
    await clickQueryButton(page);

    const sysRow = page
      .locator('.ant-tabs-tabpane-active .ant-table-tbody tr', { hasText: 'sys_config' })
      .first();
    await expect(sysRow).toBeVisible();
    await sysRow.getByRole('button', { name: '编辑配置' }).click();

    await expect(page.getByRole('button', { name: '保存配置' })).toBeVisible();
    await page.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible();

    await page.getByRole('button', { name: '保存并预览' }).click();
    await expect(page.getByText('模板预览').first()).toBeVisible();
  });

  test('multi-role user uses primary-role menus and merged permissions', async ({
    page,
    request,
  }) => {
    await loginByUI(page, request, getSmokeAccount());
    const adminAccessToken = await readAccessToken(page);
    const adminHeaders = {
      Authorization: `Bearer ${adminAccessToken}`,
      'Content-Type': 'application/json',
    };
    const username = `e2e_multi_${Date.now()}`;
    const password = '123456';
    const phoneSuffix = String(Date.now()).slice(-8);

    let createdUserId: number | undefined;

    try {
      const createResponse = await request.post(`${apiBaseUrl}/api/v1/sys-user`, {
        data: {
          deptId: 1,
          email: `${username}@example.com`,
          nickName: 'E2E 多角色',
          phone: `139${phoneSuffix}`,
          primaryRoleId: 3,
          roleId: 3,
          roleIds: [3, 1],
          status: '2',
          username,
          password,
        },
        headers: adminHeaders,
      });
      expect(createResponse.ok()).toBeTruthy();
      const createBody = await createResponse.json();
      expect(createBody.code).toBe(200);
      createdUserId =
        typeof createBody.data === 'number' && Number.isFinite(createBody.data)
          ? createBody.data
          : undefined;

      if (!createdUserId) {
        const userPageResponse = await request.get(
          `${apiBaseUrl}/api/v1/sys-user?pageIndex=1&pageSize=20&username=${username}`,
          {
            headers: {
              Authorization: `Bearer ${adminAccessToken}`,
            },
          },
        );
        expect(userPageResponse.ok()).toBeTruthy();
        const userPageBody = await userPageResponse.json();
        const createdUser = userPageBody?.data?.list?.find(
          (item: { username?: string; userId?: number }) => item.username === username,
        );
        createdUserId = createdUser?.userId;
      }
      expect(createdUserId).toBeTruthy();

      const userContext = await page.context().browser()!.newContext();
      const userPage = await userContext.newPage();
      await loginByUI(userPage, request, { password, username });
      const multiRoleAccessToken = await readAccessToken(userPage);

      const getInfoResponse = await request.get(`${apiBaseUrl}/api/v1/getinfo`, {
        headers: {
          Authorization: `Bearer ${multiRoleAccessToken}`,
        },
      });
      expect(getInfoResponse.ok()).toBeTruthy();
      const getInfoBody = await getInfoResponse.json();
      expect(getInfoBody.code).toBe(200);
      expect(getInfoBody.data.primaryRoleId).toBe(3);
      expect(getInfoBody.data.roleIds).toEqual(expect.arrayContaining([1, 3]));
      expect(getInfoBody.data.permissions || []).toContain('*:*:*');

      const menuRoleResponse = await request.get(`${apiBaseUrl}/api/v1/menurole`, {
        headers: {
          Authorization: `Bearer ${multiRoleAccessToken}`,
        },
      });
      expect(menuRoleResponse.ok()).toBeTruthy();
      const menuRoleBody = await menuRoleResponse.json();
      expect(menuRoleBody.code).toBe(200);
      const menuPayload = JSON.stringify(menuRoleBody.data || []);
      expect(menuPayload).toContain('/admin/sys-api');
      expect(menuPayload).not.toContain('/admin/sys-server-monitor');
      expect(menuPayload).not.toContain('/admin/sys-job');

      await userPage.goto('/admin/sys-api');
      await expect(
        userPage.getByPlaceholder('请输入接口标题'),
      ).toBeVisible();
      await expect(
        userPage.locator('#__vben_main_content').getByText('接口列表', { exact: true }),
      ).toBeVisible();
      await expect(userPage.getByText('服务监控', { exact: true })).toHaveCount(0);
      await expect(userPage.getByText('定时任务', { exact: true })).toHaveCount(0);
      await userContext.close();
    } finally {
      if (createdUserId) {
        await request.delete(`${apiBaseUrl}/api/v1/sys-user`, {
          data: { ids: [createdUserId] },
          headers: adminHeaders,
        });
      }
    }
  });
});
