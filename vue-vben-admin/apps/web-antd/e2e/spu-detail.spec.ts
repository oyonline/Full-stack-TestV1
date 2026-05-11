import { expect, test } from '@playwright/test';
import type { APIRequestContext, Page } from '@playwright/test';

import { getApiBaseUrl, getSmokeAccount, loginByApi, loginByUI } from './support/auth';

type CategoryTreeNode = {
  categoryId: number;
  children?: CategoryTreeNode[];
};

function firstLeafCategoryId(nodes: CategoryTreeNode[]): number | null {
  for (const n of nodes) {
    if (!n.children?.length) {
      return n.categoryId;
    }
    const nested = firstLeafCategoryId(n.children);
    if (nested !== null) {
      return nested;
    }
  }
  return null;
}

const getVisibleDrawer = (page: Page) => page.locator('.ant-drawer-content').last();

async function clickSpuQuery(page: Page) {
  await page
    .locator('#__vben_main_content')
    .getByRole('button', { name: /查\s*询/ })
    .click();
}

async function openSpuDetailFromList(page: Page, spuName: string) {
  await page.getByPlaceholder('请输入名称关键词').fill(spuName);
  await clickSpuQuery(page);
  const row = page.locator('.ant-table-tbody tr', { hasText: spuName }).first();
  await expect(row).toBeVisible();
  await row.getByRole('link', { name: '查看' }).or(row.getByRole('button', { name: '查看' })).first().click();
}

async function createSpuDraft(
  request: APIRequestContext,
  headers: Record<string, string>,
  params: { spuName: string; spuCode: string; categoryId: number; brandId: number },
): Promise<number> {
  const apiBaseUrl = getApiBaseUrl();
  const res = await request.post(`${apiBaseUrl}/api/v1/spu`, {
    data: {
      spuCode: params.spuCode,
      spuName: params.spuName,
      categoryId: params.categoryId,
      brandId: params.brandId,
      description: '<p>e2e spu detail</p>',
      detailImages: '[]',
      status: 1,
    },
    headers,
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.code).toBe(200);
  const id = typeof body.data === 'number' ? body.data : Number(body.data);
  expect(id).toBeGreaterThan(0);
  return id;
}

async function submitSpuForReview(
  request: APIRequestContext,
  headers: Record<string, string>,
  spuId: number,
): Promise<{ instanceId: number }> {
  const apiBaseUrl = getApiBaseUrl();
  const res = await request.post(`${apiBaseUrl}/api/v1/spu/${spuId}/submit`, {
    data: {},
    headers,
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.code).toBe(200);
  expect(body.data?.instanceId).toBeTruthy();
  return { instanceId: body.data.instanceId as number };
}

async function findTodoTaskIdForSpuTitle(
  request: APIRequestContext,
  headers: Record<string, string>,
  title: string,
): Promise<number> {
  const apiBaseUrl = getApiBaseUrl();
  const qs = new URLSearchParams({
    pageIndex: '1',
    pageSize: '20',
    businessType: 'spu',
    title,
  });
  const res = await request.get(`${apiBaseUrl}/api/v1/platform/workflow/tasks/todo?${qs.toString()}`, {
    headers,
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.code).toBe(200);
  const list = body.data?.list as Array<{ taskId: number }> | undefined;
  expect(list?.length).toBeTruthy();
  return list![0]!.taskId;
}

async function approveWorkflowTask(
  request: APIRequestContext,
  headers: Record<string, string>,
  taskId: number,
) {
  const apiBaseUrl = getApiBaseUrl();
  const res = await request.post(`${apiBaseUrl}/api/v1/platform/workflow/tasks/${taskId}/approve`, {
    data: { comment: 'e2e approve' },
    headers,
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.code).toBe(200);
}

/** 审核中不可删：必要时先审批再删 */
/**
 * 迁移在缺少 product_admin 角色时可能未写入审批节点，导致 Submit 报「流程定义未配置审批节点」。
 * 这里将 spu_create_review 的审批节点固定为 user=1（与 workflow-smoke 一致），保证 Playwright 环境可提交。
 */
async function ensureSpuCreateReviewWorkflowHasApproveNode(
  request: APIRequestContext,
  headers: Record<string, string>,
) {
  const apiBaseUrl = getApiBaseUrl();
  const listRes = await request.get(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions?pageIndex=1&pageSize=50&moduleKey=admin&businessType=spu&status=2`,
    { headers },
  );
  expect(listRes.ok()).toBeTruthy();
  const listBody = await listRes.json();
  expect(listBody.code).toBe(200);
  const rows = (listBody.data?.list || []) as Array<{ definitionId: number; definitionKey: string }>;
  const def = rows.find((r) => r.definitionKey === 'spu_create_review');
  expect(def?.definitionId, '需要已迁移的 spu_create_review 流程定义（moduleKey=admin, businessType=spu）').toBeTruthy();

  const detailRes = await request.get(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions/${def!.definitionId}`,
    { headers },
  );
  expect(detailRes.ok()).toBeTruthy();
  const detailBody = await detailRes.json();
  expect(detailBody.code).toBe(200);
  const nodes = (detailBody.data?.nodes || []) as Array<{ nodeType?: string }>;
  const hasApprove = nodes.some((n) => n.nodeType === 'approve');
  if (hasApprove) {
    return;
  }

  const nodesRes = await request.put(
    `${apiBaseUrl}/api/v1/platform/workflow/definitions/${def!.definitionId}/nodes`,
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
  expect(nodesBody.code, JSON.stringify(nodesBody)).toBe(200);
}

async function deleteSpuSafe(
  request: APIRequestContext,
  headers: Record<string, string>,
  spuId: number,
) {
  const apiBaseUrl = getApiBaseUrl();
  const detailRes = await request.get(`${apiBaseUrl}/api/v1/spu/${spuId}`, { headers });
  if (!detailRes.ok()) {
    return;
  }
  const detailBody = await detailRes.json();
  const status = detailBody?.data?.status as number | undefined;
  const spuName = String(detailBody?.data?.spuName ?? '');
  if (status === 2 && spuName) {
    try {
      const taskId = await findTodoTaskIdForSpuTitle(request, headers, spuName);
      await approveWorkflowTask(request, headers, taskId);
    } catch {
      // 若待办不在当前用户下，跳过后续删除
      return;
    }
  }
  const delRes = await request.delete(`${apiBaseUrl}/api/v1/spu`, {
    data: { ids: [spuId] },
    headers,
  });
  if (!delRes.ok()) {
    const t = await delRes.text();
    console.warn(`deleteSpuSafe failed spuId=${spuId}: ${delRes.status()} ${t}`);
  }
}

test.describe.serial('SPU detail scenarios (drawer / page / approval / status / mobile)', () => {
  const suffix = Date.now();
  let apiHeaders: Record<string, string>;
  let categoryId: number;
  let brandId: number;
  const toCleanup: number[] = [];

  test.beforeAll(async ({ request }) => {
    const { accessToken } = await loginByApi(request, getSmokeAccount());
    apiHeaders = {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    };
    const apiBaseUrl = getApiBaseUrl();

    await ensureSpuCreateReviewWorkflowHasApproveNode(request, apiHeaders);

    const catRes = await request.get(`${apiBaseUrl}/api/v1/sku-category?status=2`, { headers: apiHeaders });
    expect(catRes.ok()).toBeTruthy();
    const catBody = await catRes.json();
    expect(catBody.code).toBe(200);
    const leaf = firstLeafCategoryId((catBody.data || []) as CategoryTreeNode[]);
    expect(leaf, '需要至少一个启用类目的叶子节点以创建 SPU').toBeTruthy();
    categoryId = leaf!;

    const brandRes = await request.get(`${apiBaseUrl}/api/v1/sku-brand?pageIndex=1&pageSize=5&status=2`, {
      headers: apiHeaders,
    });
    expect(brandRes.ok()).toBeTruthy();
    const brandBody = await brandRes.json();
    expect(brandBody.code).toBe(200);
    const firstBrand = brandBody.data?.list?.[0] as { brandId?: number } | undefined;
    expect(firstBrand?.brandId, '需要至少一个启用品牌').toBeTruthy();
    brandId = firstBrand!.brandId!;
  });

  test.afterEach(async ({ request }) => {
    while (toCleanup.length) {
      const id = toCleanup.pop()!;
      await deleteSpuSafe(request, apiHeaders, id);
    }
  });

  test('draft: drawer open/close shows SPU detail content', async ({ page, request }) => {
    const spuName = `E2E Drawer ${suffix}`;
    const spuCode = `E2E-DW-${suffix}`;
    const spuId = await createSpuDraft(request, apiHeaders, {
      spuName,
      spuCode,
      categoryId,
      brandId,
    });
    toCleanup.push(spuId);

    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/product/spu');
    await expect(page.getByPlaceholder('请输入名称关键词')).toBeVisible();

    await page.getByPlaceholder('请输入名称关键词').fill(spuName);
    await clickSpuQuery(page);
    const row = page.locator('.ant-table-tbody tr', { hasText: spuName }).first();
    await expect(row).toBeVisible();
    await row.getByRole('button', { name: '查看' }).click();

    const drawer = getVisibleDrawer(page);
    await expect(drawer.getByText('SPU 详情', { exact: true }).first()).toBeVisible();
    await expect(drawer.getByRole('tab', { name: '基本信息' })).toBeVisible();
    await expect(drawer.getByText(spuName)).toBeVisible();

    await drawer.getByRole('button', { name: 'Close' }).click();
    await expect(drawer).toBeHidden();
  });

  test('reviewing: standalone route loads detail and shows 待审核', async ({ page, request }) => {
    const spuName = `E2E Page ${suffix}`;
    const spuCode = `E2E-PG-${suffix}`;
    const spuId = await createSpuDraft(request, apiHeaders, {
      spuName,
      spuCode,
      categoryId,
      brandId,
    });
    toCleanup.push(spuId);
    await submitSpuForReview(request, apiHeaders, spuId);

    await loginByUI(page, request, getSmokeAccount());
    await page.goto('/product/spu');
    await page.getByPlaceholder('请输入名称关键词').fill(spuName);
    await clickSpuQuery(page);

    const row = page.locator('.ant-table-tbody tr', { hasText: spuName }).first();
    await expect(row).toBeVisible();
    await row.getByRole('link', { name: '查看' }).click();

    await expect(page).toHaveURL(new RegExp(`/product/spu/${spuId}`));
    await expect(page.getByText('SPU 详情', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('待审核').first()).toBeVisible();
    await expect(page.getByText(spuName).first()).toBeVisible();
  });

  test('approval history tab shows workflow start entry after submit', async ({ page, request }) => {
    const spuName = `E2E ApproveTab ${suffix}`;
    const spuCode = `E2E-AT-${suffix}`;
    const spuId = await createSpuDraft(request, apiHeaders, {
      spuName,
      spuCode,
      categoryId,
      brandId,
    });
    toCleanup.push(spuId);
    await submitSpuForReview(request, apiHeaders, spuId);

    await loginByUI(page, request, getSmokeAccount());
    await page.goto(`/product/spu/${spuId}`);
    await expect(page.getByText('SPU 详情', { exact: true }).first()).toBeVisible();

    const actionsResp = page.waitForResponse(
      (r) =>
        r.request().method() === 'GET' &&
        r.url().includes('/api/v1/platform/workflow/instances/') &&
        r.url().includes('/actions') &&
        r.ok(),
      { timeout: 25_000 },
    );
    await page.getByRole('tab', { name: '审批历史' }).click();
    const actionsResponse = await actionsResp;
    const actionsJson = (await actionsResponse.json()) as {
      code?: number;
      data?: { list?: Array<{ action?: string }> };
    };
    expect(actionsJson.code).toBe(200);
    const logList = actionsJson.data?.list ?? [];
    expect(
      logList.length,
      '点击「审批历史」应请求实例 actions 且至少含 submit 产生的 start 日志',
    ).toBeGreaterThan(0);
    expect(logList.some((row) => row.action === 'start')).toBeTruthy();
  });

  test('after approve API reload shows 已通过 status tag', async ({ page, request }) => {
    const spuName = `E2E Status ${suffix}`;
    const spuCode = `E2E-ST-${suffix}`;
    const spuId = await createSpuDraft(request, apiHeaders, {
      spuName,
      spuCode,
      categoryId,
      brandId,
    });
    toCleanup.push(spuId);
    await submitSpuForReview(request, apiHeaders, spuId);
    const taskId = await findTodoTaskIdForSpuTitle(request, apiHeaders, spuName);
    await approveWorkflowTask(request, apiHeaders, taskId);

    await loginByUI(page, request, getSmokeAccount());
    await page.goto(`/product/spu/${spuId}`);
    await expect(page.getByText('已通过').first()).toBeVisible({ timeout: 20_000 });

    await page.reload();
    await expect(page.getByText('已通过').first()).toBeVisible();
    await expect(page.getByRole('tab', { name: '审批历史' })).toBeVisible();
  });

  test('mobile viewport: detail page tabs and basic info remain usable', async ({ page, request }) => {
    const spuName = `E2E Mobile ${suffix}`;
    const spuCode = `E2E-MB-${suffix}`;
    const spuId = await createSpuDraft(request, apiHeaders, {
      spuName,
      spuCode,
      categoryId,
      brandId,
    });
    toCleanup.push(spuId);
    await submitSpuForReview(request, apiHeaders, spuId);
    const taskId = await findTodoTaskIdForSpuTitle(request, apiHeaders, spuName);
    await approveWorkflowTask(request, apiHeaders, taskId);

    await page.setViewportSize({ width: 390, height: 844 });
    await loginByUI(page, request, getSmokeAccount());
    await page.goto(`/product/spu/${spuId}`);

    await expect(page).toHaveURL(new RegExp(`/product/spu/${spuId}`));
    // 小屏下页头/面包屑可能折叠，「SPU 详情」首条命中常为隐藏节点；以主区 Tab 为准
    const main = page.getByRole('main');
    await expect(main.getByRole('tab', { name: '基本信息' })).toBeVisible();
    await main.getByRole('tab', { name: 'SKU' }).click();
    await expect(main.getByText(/SKU 列表/)).toBeVisible();
  });
});
