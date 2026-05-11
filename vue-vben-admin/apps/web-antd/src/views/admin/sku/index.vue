<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';

import type { SkuItem, SkuPageResult } from '#/api/core';

import { onMounted } from 'vue';

import { Button, Input, InputNumber, Select, Table, Tag } from 'ant-design-vue';

import { getSkuPage } from '#/api/core';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import { formatAdminDateTime } from '#/utils/admin-crud';

const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '禁用' },
  { value: 2, label: '启用' },
];

const {
  errorMsg,
  fetchList,
  loading,
  onReset,
  onSearch,
  onTableChange,
  pagination,
  query,
  tableData,
} = useAdminTable<
  SkuItem,
  {
    skuCode: string;
    skuName: string;
    spuId: number | '';
    status: '' | number;
  },
  {
    skuCode?: string;
    skuName?: string;
    spuId?: number;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    skuCode: currentQuery.skuCode.trim() || undefined,
    skuName: currentQuery.skuName.trim() || undefined,
    spuId: currentQuery.spuId === '' ? undefined : Number(currentQuery.spuId),
    status:
      currentQuery.status === '' ? undefined : Number(currentQuery.status),
  }),
  createQuery: () => ({
    skuCode: '',
    skuName: '',
    spuId: '' as '',
    status: '' as '' | number,
  }),
  fallbackErrorMessage: '加载 SKU 列表失败',
  fetcher: async (params) => {
    const res: SkuPageResult = await getSkuPage(params);
    return res;
  },
});

const formatDateTime = formatAdminDateTime;

const baseColumns: TableColumnType[] = [
  { title: 'SKU ID', dataIndex: 'skuId', key: 'skuId', width: 90 },
  { title: 'SPU ID', dataIndex: 'spuId', key: 'spuId', width: 90 },
  {
    title: 'SKU 编码',
    dataIndex: 'skuCode',
    key: 'skuCode',
    width: 160,
  },
  { title: 'SKU 名称', dataIndex: 'skuName', key: 'skuName', width: 200 },
  { title: '规格', dataIndex: 'spec', key: 'spec', width: 160 },
  { title: '单位', dataIndex: 'unit', key: 'unit', width: 80 },
  {
    title: '售价',
    dataIndex: 'price',
    key: 'price',
    width: 100,
    customRender: ({ text }: { text: number }) =>
      text != null ? `¥${(text / 100).toFixed(2)}` : '—',
  },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ text }: { text: string }) => formatDateTime(text),
    customCell: () => ({ style: { whiteSpace: 'nowrap' } }),
  },
];

const {
  handleResizeColumn,
  reorderColumns,
  restoreDefaultColumns,
  scrollX,
  setColumnFixed,
  setColumnVisible,
  settingsColumns,
  settingsOpen,
  tableColumns,
} = useAdminTableColumns(baseColumns, {
  systemColumnKeys: [],
  tableId: 'sku-manage-list',
});

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>Product Center</template>
    <template #title>SKU 管理</template>
    <template #description>
      查看和筛选所有 SKU 条目。SKU 的新增与编辑在 SPU 详情页的子表中进行。
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="SKU 编码">
          <Input
            v-model:value="query.skuCode"
            placeholder="请输入 SKU 编码"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="SKU 名称">
          <Input
            v-model:value="query.skuName"
            placeholder="请输入 SKU 名称"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="SPU ID">
          <InputNumber
            v-model:value="query.spuId"
            placeholder="请输入 SPU ID"
            class="w-full"
            :min="1"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            placeholder="请选择状态"
          />
        </AdminFilterField>
      </div>
    </template>
    <template #filter-actions>
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </template>
    <template #toolbar>
      <div>
        <div class="text-base font-semibold text-slate-900">SKU 列表</div>
      </div>
    </template>
    <template #toolbar-extra>
      <AdminTableColumnSettings
        v-model:open="settingsOpen"
        :columns="settingsColumns"
        @change-fixed="({ key, fixed }) => setColumnFixed(key, fixed)"
        @reorder="
          ({ oldIndex, newIndex }) => reorderColumns(oldIndex, newIndex)
        "
        @reset="restoreDefaultColumns"
        @toggle-visible="({ key, visible }) => setColumnVisible(key, visible)"
      />
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SkuItem) => record.skuId"
      :scroll="{ x: scrollX }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
      @resize-column="handleResizeColumn"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <Tag
            :color="(record as SkuItem).status === 2 ? 'green' : 'default'"
          >
            {{ (record as SkuItem).status === 2 ? '启用' : '禁用' }}
          </Tag>
        </template>
      </template>
    </Table>
  </AdminPageShell>
</template>
