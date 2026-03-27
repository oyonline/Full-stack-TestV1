import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { expect, test } from '@playwright/test';

import { getApiBaseUrl, getSmokeAccount, loginByApi, loginByUI } from './support/auth';

const unique = Date.now();
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const fixturePath = path.resolve(
  __dirname,
  './fixtures/workflow-attachment-smoke.txt',
);

async function createDefinitionAndInstance(
  request: Parameters<typeof loginByApi>[0],
  accessToken: string,
) {
  const apiBaseUrl = getApiBaseUrl();
  const headers = {
    Authorization: `Bearer ${accessToken}`,
    'Content-Type': 'application/json',
  };

  const businessType = `budget-attachment-${unique}`;
  const definitionRes = await request.post(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions`,
    {
      data: {
        businessType,
        definitionKey: `budget-attachment-${unique}`,
        definitionName: `预算附件验证-${unique}`,
        moduleKey: 'finance-budget',
        remark: 'workflow attachment smoke',
        status: '2',
        version: 1,
      },
      headers,
    },
  );
  expect(definitionRes.ok()).toBeTruthy();
  const definitionBody = await definitionRes.json();
  expect(definitionBody.code).toBe(200);
  const definitionId = definitionBody.data.definitionId as number;

  const nodesRes = await request.put(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions/${definitionId}/nodes`,
    {
      data: {
        nodes: [
          { nodeKey: 'start', nodeName: '开始', nodeType: 'start', sort: 1 },
          {
            approverName: 'admin',
            approverType: 'user',
            approverValue: '1',
            nodeKey: 'approve_1',
            nodeName: '管理员审批',
            nodeType: 'approve',
            sort: 2,
          },
          { nodeKey: 'end', nodeName: '结束', nodeType: 'end', sort: 3 },
        ],
      },
      headers,
    },
  );
  expect(nodesRes.ok()).toBeTruthy();
  const nodesBody = await nodesRes.json();
  expect(nodesBody.code).toBe(200);

  const startRes = await request.post(
    `${apiBaseUrl}/api/v1/platform/workflow/instances/start`,
    {
      data: {
        businessId: `${unique}`,
        businessNo: `ATTACH-${unique}`,
        businessType,
        definitionId,
        moduleKey: 'finance-budget',
        title: `附件验证-${unique}`,
      },
      headers,
    },
  );
  expect(startRes.ok()).toBeTruthy();
  const startBody = await startRes.json();
  expect(startBody.code).toBe(200);
  return startBody.data.instance as { instanceId: number; title: string };
}

test.describe.serial('workflow step3 attachment smoke', () => {
  test('started detail supports upload, download and delete attachment', async ({
    page,
    request,
  }) => {
    const loginResult = await loginByApi(request, getSmokeAccount());
    const instance = await createDefinitionAndInstance(request, loginResult.accessToken);

    await loginByUI(page, request, getSmokeAccount());

    await page.goto('/platform/workflow/started');
    await expect(page.getByPlaceholder('请输入标题')).toBeVisible();
    await page.getByPlaceholder('请输入标题').fill(instance.title);
    await page
      .locator('#__vben_main_content')
      .getByRole('button', { name: /查\s*询/ })
      .click();

    const row = page.locator('.ant-table-tbody tr', { hasText: instance.title }).first();
    await expect(row).toBeVisible();
    await row.getByRole('button', { name: '详情' }).click();

    const drawer = page.locator('.ant-drawer-content').last();
    await expect(drawer.locator('.ant-drawer-title')).toHaveText(instance.title);
    await expect(drawer.getByText('关联附件', { exact: true })).toBeVisible();

    const uploadResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        response.url().includes('/api/v1/platform/attachments/upload'),
    );
    await drawer
      .locator('[data-testid="workflow-attachment-input"]')
      .setInputFiles(fixturePath);
    await uploadResponse;
    await expect(page.getByText('附件上传成功')).toBeVisible();

    const attachmentRow = drawer
      .locator('.ant-table-tbody tr', { hasText: 'workflow-attachment-smoke.txt' })
      .first();
    await expect(attachmentRow).toBeVisible();

    await attachmentRow
      .locator('[data-testid="workflow-attachment-download"]')
      .click();
    await expect(page.getByText('附件下载已开始')).toBeVisible();

    await attachmentRow
      .locator('[data-testid="workflow-attachment-delete"]')
      .click();
    const confirmPopover = page.locator('.ant-popover').last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover.locator('.ant-btn-primary').click();
    await expect(page.getByText('附件删除成功')).toBeVisible();
    await expect(
      drawer.locator('.ant-table-tbody tr', { hasText: 'workflow-attachment-smoke.txt' }),
    ).toHaveCount(0);
  });
});
