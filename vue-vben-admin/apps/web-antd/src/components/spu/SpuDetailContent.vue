<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';

import type { SkuItem, SpuItem } from '#/api/core';

import { computed, h, ref, watch } from 'vue';

import {
  Descriptions,
  Spin,
  Table,
  Tag,
  Timeline,
} from 'ant-design-vue';

import {
  getSkuPage,
  getSpuDetail,
  getWorkflowInstanceDetail,
} from '#/api/core';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import { getWorkflowActionLabel } from '#/views/platform/workflow/constants';

const props = defineProps<{
  mode: 'drawer' | 'page';
  readonly: boolean;
  spuId: number;
}>();

const emit = defineEmits<{
  loaded: [item: SpuItem];
}>();

/* -------- 加载状态 -------- */
const loading = ref(false);
const spu = ref<SpuItem | null>(null);
const detailImages = ref<string[]>([]);
const skus = ref<SkuItem[]>([]);
const workflowDetail = ref<null | Awaited<ReturnType<typeof getWorkflowInstanceDetail>>>(null);

/* -------- 审批历史 -------- */
const timelineItems = computed(() => {
  const actions = workflowDetail.value?.actions || [];
  return actions.map((item) => ({
    children: `${formatAdminDateTime(item.createdAt)}  ${renderAdminEmpty(item.operatorName)} · ${getWorkflowActionLabel(item.action)}${item.comment ? ` · ${item.comment}` : ''}`,
    color:
      item.action === 'approve'
        ? 'green'
        : item.action === 'reject'
          ? 'red'
          : item.action === 'withdraw'
            ? 'gray'
            : 'blue',
  }));
});

const rejectAction = computed(() => {
  const actions = workflowDetail.value?.actions || [];
  // 找最后一条 reject
  for (let i = actions.length - 1; i >= 0; i--) {
    const act = actions[i];
    if (act && act.action === 'reject') {
      return act;
    }
  }
  return null;
});

/* -------- SKU 列 -------- */
const skuColumns: TableColumnType[] = [
  { title: 'SKU 编码', dataIndex: 'skuCode', key: 'skuCode', width: 140 },
  { title: 'SKU 名称', dataIndex: 'skuName', key: 'skuName', ellipsis: true },
  { title: '规格', dataIndex: 'spec', key: 'spec', width: 120 },
  { title: '单位', dataIndex: 'unit', key: 'unit', width: 80 },
  { title: '价格', dataIndex: 'price', key: 'price', width: 100 },
];

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

/* -------- 详情图解析 -------- */
function parseDetailImages(raw?: string): string[] {
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw);
    return Array.isArray(arr) ? arr.filter((s) => typeof s === 'string') : [];
  } catch {
    return [];
  }
}

/* -------- 加载数据 -------- */
async function load() {
  if (!props.spuId) return;
  loading.value = true;
  try {
    const d = await getSpuDetail(props.spuId);
    spu.value = d;
    detailImages.value = parseDetailImages(d.detailImages);

    const skuPage = await getSkuPage({
      spuId: props.spuId,
      pageIndex: 1,
      pageSize: 200,
    });
    skus.value = skuPage.list || [];

    // 如果有 workflowInstanceId，加载审批历史
    if (d.workflowInstanceId) {
      try {
        workflowDetail.value = await getWorkflowInstanceDetail(d.workflowInstanceId);
      } catch {
        workflowDetail.value = null;
      }
    } else {
      workflowDetail.value = null;
    }

    emit('loaded', d);
  } catch {
    spu.value = null;
    detailImages.value = [];
    skus.value = [];
    workflowDetail.value = null;
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.spuId,
  () => load(),
  { immediate: true },
);
</script>

<template>
  <div>
    <div v-if="loading" class="py-8 text-center text-sm text-slate-400">
      <Spin />
      <div class="mt-2">加载中…</div>
    </div>

    <div v-else-if="spu" class="space-y-5">
      <!-- Rejected 状态顶部横幅 -->
      <div
        v-if="spu.status === SPU_STATUS.rejected && rejectAction"
        class="rounded-sm border border-red-200 bg-red-50 p-3 text-sm text-red-700"
      >
        <div class="font-semibold">驳回理由</div>
        <div>{{ rejectAction.comment || '无备注' }}</div>
        <div class="mt-1 text-xs text-red-500">
          {{ formatAdminDateTime(rejectAction.createdAt) }} · {{ rejectAction.operatorName || '-' }}
        </div>
      </div>

      <!-- 基础信息 -->
      <AdminDetailSection title="基础信息">
        <Descriptions :column="2" bordered size="small">
          <Descriptions.Item label="SPU 名称">
            {{ renderAdminEmpty(spu.spuName) }}
          </Descriptions.Item>
          <Descriptions.Item label="SPU 编码">
            {{ renderAdminEmpty(spu.spuCode) }}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <component :is="renderStatusTag(spu.status)" />
          </Descriptions.Item>
          <Descriptions.Item label="类目">
            {{ renderAdminEmpty(String(spu.categoryId)) }}
          </Descriptions.Item>
          <Descriptions.Item label="品牌">
            {{ renderAdminEmpty(String(spu.brandId)) }}
          </Descriptions.Item>
          <Descriptions.Item label="创建人">
            {{ renderAdminEmpty(spu.creatorId ? String(spu.creatorId) : '') }}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {{ formatAdminDateTime(spu.updatedAt) }}
          </Descriptions.Item>
        </Descriptions>
      </AdminDetailSection>

      <!-- 主图 -->
      <AdminDetailSection v-if="spu.mainImageUrl" title="主图">
        <img
          :src="spu.mainImageUrl"
          class="max-h-72 rounded-sm object-contain"
          alt="main"
        />
      </AdminDetailSection>

      <!-- 详情图 -->
      <AdminDetailSection v-if="detailImages.length > 0" title="详情图">
        <div class="flex flex-wrap gap-2">
          <img
            v-for="(url, idx) in detailImages"
            :key="`${url}-${idx}`"
            :src="url"
            class="h-28 w-28 rounded-sm object-cover"
            alt="detail"
          />
        </div>
      </AdminDetailSection>

      <!-- 详情 -->
      <AdminDetailSection title="详情">
        <div
          class="prose max-w-none"
          v-html="spu.description || '<p class=\'text-gray-400\'>无详情</p>'"
        ></div>
      </AdminDetailSection>

      <!-- SKU 列表 -->
      <AdminDetailSection :title="`SKU 列表（${skus.length}）`">
        <Table
          :columns="skuColumns"
          :data-source="skus"
          :pagination="false"
          :row-key="(r: SkuItem) => r.skuId"
          size="small"
        />
      </AdminDetailSection>

      <!-- 审批历史 -->
      <AdminDetailSection title="审批历史">
        <Timeline v-if="timelineItems.length" :items="timelineItems" />
        <div v-else class="text-sm text-slate-400">暂无审批记录</div>
      </AdminDetailSection>
    </div>

    <div v-else class="py-8 text-center text-sm text-slate-400">
      未找到该 SPU
    </div>
  </div>
</template>
