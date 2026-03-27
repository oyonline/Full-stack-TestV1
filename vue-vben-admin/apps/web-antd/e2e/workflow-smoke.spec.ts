import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

import { getApiBaseUrl, getSmokeAccount, loginByApi, loginByUI } from './support/auth';

const unique = Date.now();

const createWorkflowDefinition = async (request: Parameters<typeof loginByApi>[0], token: string) => {
  const apiBaseUrl = getApiBaseUrl();
  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  const definitionRes = await request.post(`${apiBaseUrl}/api/v1/platform/workflow/definitions`, {
    data: {
      businessType: `budget-request-e2e-${unique}`,
      definitionKey: `budget-workflow-e2e-${unique}`,
      definitionName: `预算审批E2E-${unique}`,
      moduleKey: 'finance-budget',
      remark: 'workflow step2 smoke',
      status: '2',
      version: 1,
    },
    headers,
  });
  expect(definitionRes.ok()).toBeTruthy();
  const definitionBody = await definitionRes.json();
  expect(definitionBody.code).toBe(200);
  const definitionId = definitionBody?.data?.definitionId as number;
  expect(definitionId).toBeTruthy();

  const nodesRes = await request.put(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions/${definitionId}/nodes`,
    {
      data: {
        nodes: [
          {
            nodeKey: 'start',
            nodeName: '开始',
            nodeType: 'start',
            sort: 1,
          },
          {
            approverName: 'admin',
            approverType: 'user',
            approverValue: '1',
            nodeKey: 'approve_1',
            nodeName: '管理员审批',
            nodeType: 'approve',
            sort: 2,
          },
          {
            nodeKey: 'end',
            nodeName: '结束',
            nodeType: 'end',
            sort: 3,
          },
        ],
      },
      headers,
    },
  );
  expect(nodesRes.ok()).toBeTruthy();
  const nodesBody = await nodesRes.json();
  expect(nodesBody.code).toBe(200);

  return {
    businessType: `budget-request-e2e-${unique}`,
    definitionId,
    headers,
  };
};

const startWorkflowInstance = async (
  request: Parameters<typeof loginByApi>[0],
  headers: Record<string, string>,
  definitionId: number,
  businessType: string,
  suffix: string,
) => {
  const apiBaseUrl = getApiBaseUrl();
  const res = await request.post(`${apiBaseUrl}/api/v1/platform/workflow/instances/start`, {
    data: {
      businessId: `${unique}-${suffix}`,
      businessNo: `BUDGET-${unique}-${suffix.toUpperCase()}`,
      businessType,
      definitionId,
      moduleKey: 'finance-budget',
      title: `预算审批-${suffix}-${unique}`,
    },
    headers,
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.code).toBe(200);
  return body.data as {
    instance: { instanceId: number; title: string };
    tasks: Array<{ taskId: number }>;
  };
};

const clickQueryButton = async (page: Page) => {
  await page
    .locator('#__vben_main_content')
    .getByRole('button', { name: /查\s*询/ })
    .click();
};

const openDetailByTitle = async (page: Page, title: string) => {
  const row = page.locator('.ant-table-tbody tr', { hasText: title }).first();
  await expect(row).toBeVisible();
  await row.getByRole('button', { name: '详情' }).click();
};

const getVisibleDrawer = (page: Page) => page.locator('.ant-drawer-content').last();

const closeVisibleDrawer = async (page: Page) => {
  const drawer = getVisibleDrawer(page);
  await drawer.getByRole('button', { name: 'Close' }).click();
  await expect(drawer).toBeHidden();
};

const clickVisibleActionButton = async (
  page: Page,
  action: 'approve' | 'reject' | 'withdraw',
) => {
  const testIds = {
    approve: 'workflow-approve-action',
    reject: 'workflow-reject-action',
    withdraw: 'workflow-withdraw-action',
  } as const;
  const button = page.locator(`[data-testid="${testIds[action]}"]`).last();
  await expect(button).toBeVisible();
  await button.evaluate((element) => {
    element.scrollIntoView({ block: 'center', inline: 'center' });
    (element as HTMLButtonElement).click();
  });
};

const confirmPrimaryModal = async (page: Page) => {
  const modal = page.locator('.ant-modal-confirm').last();
  await expect(modal).toBeVisible();
  await modal.locator('.ant-modal-confirm-btns .ant-btn-primary').click({ force: true });
};

test.describe.serial('workflow step2 smoke', () => {
  test('todo and started pages validate workflow actions in UI', async ({ page, request }) => {
    const loginResult = await loginByApi(request, getSmokeAccount());
    const { businessType, definitionId, headers } = await createWorkflowDefinition(
      request,
      loginResult.accessToken,
    );

    const approveFlow = await startWorkflowInstance(
      request,
      headers,
      definitionId,
      businessType,
      'approve',
    );
    const rejectFlow = await startWorkflowInstance(
      request,
      headers,
      definitionId,
      businessType,
      'reject',
    );
    const withdrawFlow = await startWorkflowInstance(
      request,
      headers,
      definitionId,
      businessType,
      'withdraw',
    );

    await loginByUI(page, request, getSmokeAccount());

    await page.goto('/platform/workflow/todo');
    await expect(page.getByPlaceholder('请输入标题')).toBeVisible();
    await expect(
      page.locator('#__vben_main_content').getByText('我的待办', { exact: true }),
    ).toBeVisible();

    await page.getByPlaceholder('请输入标题').fill(approveFlow.instance.title);
    await clickQueryButton(page);
    await openDetailByTitle(page, approveFlow.instance.title);
    const approveDrawer = getVisibleDrawer(page);
    await expect(approveDrawer.locator('.ant-drawer-title')).toHaveText(approveFlow.instance.title);
    await expect(approveDrawer.getByText('审批记录时间线', { exact: true })).toBeVisible();
    await expect(approveDrawer.getByText('关联任务', { exact: true })).toBeVisible();
    await clickVisibleActionButton(page, 'approve');
    const approveResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        response.url().includes(`/api/v1/platform/workflow/tasks/${approveFlow.tasks[0]?.taskId || approveFlow.instance.instanceId}/approve`),
    );
    await confirmPrimaryModal(page);
    await approveResponse;
    await expect(page.getByText('通过成功')).toBeVisible();
    await closeVisibleDrawer(page);

    await page.getByPlaceholder('请输入标题').fill(rejectFlow.instance.title);
    await clickQueryButton(page);
    await openDetailByTitle(page, rejectFlow.instance.title);
    const rejectDrawer = getVisibleDrawer(page);
    await expect(rejectDrawer.locator('.ant-drawer-title')).toHaveText(rejectFlow.instance.title);
    await clickVisibleActionButton(page, 'reject');
    const rejectResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        response.url().includes(`/api/v1/platform/workflow/tasks/${rejectFlow.tasks[0]?.taskId || rejectFlow.instance.instanceId}/reject`),
    );
    await confirmPrimaryModal(page);
    await rejectResponse;
    await expect(page.getByText('驳回成功')).toBeVisible();
    await closeVisibleDrawer(page);

    await page.goto('/platform/workflow/started');
    await expect(page.getByPlaceholder('请输入标题')).toBeVisible();
    await expect(
      page.locator('#__vben_main_content').getByText('我发起的', { exact: true }),
    ).toBeVisible();

    await page.getByPlaceholder('请输入标题').fill(withdrawFlow.instance.title);
    await clickQueryButton(page);
    await openDetailByTitle(page, withdrawFlow.instance.title);
    const withdrawDrawer = getVisibleDrawer(page);
    await expect(withdrawDrawer.locator('.ant-drawer-title')).toHaveText(withdrawFlow.instance.title);
    await clickVisibleActionButton(page, 'withdraw');
    const withdrawResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        response.url().includes(`/api/v1/platform/workflow/instances/${withdrawFlow.instance.instanceId}/withdraw`),
    );
    await confirmPrimaryModal(page);
    await withdrawResponse;
    await expect(page.getByText('撤回成功')).toBeVisible();
    await closeVisibleDrawer(page);

    await page.getByPlaceholder('请输入标题').fill(approveFlow.instance.title);
    await clickQueryButton(page);
    await expect(page.locator('.ant-table-tbody tr', { hasText: '已通过' }).first()).toBeVisible();

    await page.getByPlaceholder('请输入标题').fill(rejectFlow.instance.title);
    await clickQueryButton(page);
    await expect(page.locator('.ant-table-tbody tr', { hasText: '已驳回' }).first()).toBeVisible();

    await page.getByPlaceholder('请输入标题').fill(withdrawFlow.instance.title);
    await clickQueryButton(page);
    await expect(page.locator('.ant-table-tbody tr', { hasText: '已撤回' }).first()).toBeVisible();
  });
});
