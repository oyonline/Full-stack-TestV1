<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import {
  Alert,
  Button,
  Descriptions,
  Spin,
  Table,
  Tabs,
  Tag,
} from 'ant-design-vue';
import { h } from 'vue';
import type { TableColumnType } from 'ant-design-vue';

import type {
  SkuItem,
  SpuItem,
  WorkflowInstanceAction,
} from '#/api/core';
import {
  getSkuPage,
  getSpuDetail,
  getWorkflowInstanceActions,
} from '#/api/core';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import SpuAuditTimeline from './SpuAuditTimeline.vue';
import SpuRejectBanner from './SpuRejectBanner.vue';

const props = withDefaults(
  defineProps<{
    mode?: 'drawer' | 'page';
    readonly?: boolean;
    spuId: number;
  }>(),
  {
    mode: 'drawer',
    readonly: false,
  },
);

/* -------- 数据加载 -------- */
const loading = ref(false);
const errorMsg = ref('');
const spu = ref<null | SpuItem>(null);
const detailImages = ref<string[]>([]);
const detailSkus = ref<SkuItem[]>([]);
const activeTab = ref('basic');

/* -------- 审批历史 -------- */
const approvalActions = ref<WorkflowInstanceAction[]>([]);
const approvalLoading = ref(false);
const approvalError = ref('');

const rejectAction = computed(() =>
  approvalActions.value.find((a) => a.action === 'reject'),
);

const detailSkuColumns: TableColumnType[] = [
  { title: 'SKU 编码', dataIndex: 'skuCode', key: 'skuCode', width: 140 },
  { title: 'SKU 名称', dataIndex: 'skuName', key: 'skuName', ellipsis: true },
  { title: '规格', dataIndex: 'spec', key: 'spec', width: 120 },
  { title: '单位', dataIndex: 'unit', key: 'unit', width: 80 },
  { title: '价格', dataIndex: 'price', key: 'price', width: 100 },
];

function parseDetailImages(raw?: string): string[] {
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw);
    return Array.isArray(arr) ? arr.filter((s) => typeof s === 'string') : [];
  } catch {
    return [];
  }
}

async function loadApprovalHistory(instanceId: number) {
  approvalLoading.value = true;
  approvalError.value = '';
  try {
    const result = await getWorkflowInstanceActions(instanceId);
    approvalActions.value = result.list ?? [];
  } catch (error: any) {
    approvalError.value = error?.message || '加载审批历史失败';
  } finally {
    approvalLoading.value = false;
  }
}

async function loadData() {
  if (!props.spuId) {
    errorMsg.value = '缺少 SPU ID';
    return;
  }
  loading.value = true;
  errorMsg.value = '';
  spu.value = null;
  detailImages.value = [];
  detailSkus.value = [];
  approvalActions.value = [];
  approvalError.value = '';
  try {
    const d = await getSpuDetail(props.spuId);
    spu.value = d;
    detailImages.value = parseDetailImages(d.detailImages);
    const skuPage = await getSkuPage({
      spuId: props.spuId,
      pageIndex: 1,
      pageSize: 200,
    });
    detailSkus.value = skuPage.list || [];
  } catch (error: any) {
    errorMsg.value = error?.message || '获取 SPU 详情失败';
  } finally {
    loading.value = false;
  }
}

function onTabChange(key: number | string) {
  activeTab.value = String(key);
  if (
    key === 'approval_history' &&
    spu.value?.workflowInstanceId &&
    !approvalActions.value.length &&
    !approvalLoading.value &&
    !approvalError.value
  ) {
    void loadApprovalHistory(spu.value.workflowInstanceId);
  }
}

watch(
  () => props.spuId,
  () => {
    activeTab.value = 'basic';
    void loadData();
  },
  { immediate: true },
);

/* -------- 状态标签 -------- */
const SPU_STATUS = {
  approved: 3,
  draft: 1,
  offline: 5,
  rejected: 4,
  reviewing: 2,
} as const;

function renderStatusTag(status: number) {
  if (status === SPU_STATUS.draft)
    return h(Tag, { color: 'default' }, () => '草稿');
  if (status === SPU_STATUS.reviewing)
    return h(Tag, { color: 'blue' }, () => '待审核');
  if (status === SPU_STATUS.approved)
    return h(Tag, { color: 'green' }, () => '已通过');
  if (status === SPU_STATUS.rejected)
    return h(Tag, { color: 'red' }, () => '已驳回');
  if (status === SPU_STATUS.offline)
    return h(Tag, { color: 'orange' }, () => '已下架');
  return String(status);
}
</script>

<template>
  <div>
    <!-- 加载态 -->
    <div v-if="loading" class="py-8 text-center">
      <Spin />
    </div>

    <!-- 错误态 -->
    <Alert
      v-else-if="errorMsg"
      type="error"
      show-icon
      class="m-4"
      :message="errorMsg"
    >
      <template #action>
        <Button size="small" @click="loadData">重试</Button>
      </template>
    </Alert>

    <div v-else-if="spu">
      <!-- Rejected 横幅 -->
      <SpuRejectBanner v-if="rejectAction" :reject-action="rejectAction" />

      <Tabs :active-key="activeTab" @change="onTabChange">
        <!-- 基本信息 -->
        <Tabs.TabPane key="basic" tab="基本信息">
          <div class="space-y-5 pt-2">
            <Descriptions :column="2" bordered size="small">
              <Descriptions.Item label="SPU 名称">
                {{ spu.spuName }}
              </Descriptions.Item>
              <Descriptions.Item label="SPU 编码">
                {{ spu.spuCode }}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <component :is="renderStatusTag(spu.status)" />
              </Descriptions.Item>
              <Descriptions.Item label="类目">
                {{ renderAdminEmpty(String(spu.categoryId || '-')) }}
              </Descriptions.Item>
              <Descriptions.Item label="品牌">
                {{ renderAdminEmpty(String(spu.brandId || '-')) }}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {{ formatAdminDateTime(spu.createdAt) }}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {{ formatAdminDateTime(spu.updatedAt) }}
              </Descriptions.Item>
            </Descriptions>

            <div v-if="spu.mainImageUrl">
              <div class="mb-1 text-sm font-medium">主图</div>
              <img
                :src="spu.mainImageUrl"
                class="max-h-72 rounded-sm object-contain"
                alt="main"
              />
            </div>

            <div v-if="detailImages.length > 0">
              <div class="mb-1 text-sm font-medium">详情图</div>
              <div class="flex flex-wrap gap-2">
                <img
                  v-for="(url, idx) in detailImages"
                  :key="`${url}-${idx}`"
                  :src="url"
                  class="h-28 w-28 rounded-sm object-cover"
                  alt="detail"
                />
              </div>
            </div>

            <div>
              <div class="mb-1 text-sm font-medium">详情</div>
              <div
                class="prose max-w-none"
                v-html="spu.description || '<p class=\'text-gray-400\'>无详情</p>'"
              ></div>
            </div>
          </div>
        </Tabs.TabPane>

        <!-- SKU -->
        <Tabs.TabPane key="sku" tab="SKU">
          <div class="pt-2">
            <div class="mb-2 text-sm font-medium">
              SKU 列表（{{ detailSkus.length }}）
            </div>
            <Table
              :columns="detailSkuColumns"
              :data-source="detailSkus"
              :pagination="false"
              :row-key="(r: SkuItem) => r.skuId"
              size="small"
            />
          </div>
        </Tabs.TabPane>

        <!-- 详情图 -->
        <Tabs.TabPane key="detail_images" tab="详情图">
          <div class="pt-2">
            <div v-if="detailImages.length > 0" class="flex flex-wrap gap-3">
              <img
                v-for="(url, idx) in detailImages"
                :key="`${url}-${idx}`"
                :src="url"
                class="h-40 w-40 rounded-sm object-cover"
                alt="detail"
              />
            </div>
            <div v-else class="py-8 text-center text-sm text-slate-400">
              暂无详情图
            </div>
          </div>
        </Tabs.TabPane>

        <!-- 审批历史 -->
        <Tabs.TabPane key="approval_history" tab="审批历史">
          <div class="pt-2">
            <SpuAuditTimeline
              :actions="approvalActions"
              :compact="mode === 'drawer'"
              :error="approvalError"
              :loading="approvalLoading"
              @retry="
                spu?.workflowInstanceId
                  ? loadApprovalHistory(spu.workflowInstanceId)
                  : undefined
              "
            />
          </div>
        </Tabs.TabPane>
      </Tabs>
    </div>
  </div>
</template>
