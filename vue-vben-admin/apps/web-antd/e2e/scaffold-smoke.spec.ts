import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

import { getSmokeAccount, loginByUI } from './support/auth';

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
  const clickQueryButton = async (page: Page) => {
    await page
      .locator('#__vben_main_content')
      .getByRole('button', { name: /查\s*询/ })
      .click();
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
      page.locator('#__vben_main_content').getByText('接口管理', { exact: true }),
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
      page.locator('#__vben_main_content').getByText('代码生成', { exact: true }),
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
});
