<script lang="ts" setup>
import { h, onMounted, ref } from 'vue';

import { Button, Input, Select, Table, Tag } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import { getWorkflowStartedPage, type WorkflowStartedItem } from '#/api/core';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import WorkflowInstanceDetailDrawer from '../components/workflow-instance-detail-drawer.vue';
import {
  getWorkflowStatusLabel,
  workflowStatusColors,
} from '../constants';

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'in_review' },
  { label: '已通过', value: 'approved' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'cancelled' },
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
  WorkflowStartedItem,
  {
    businessNo: string;
    businessType: string;
    status: string;
    title: string;
  },
  {
    businessNo?: string;
    businessType?: string;
    status?: string;
    title?: string;
  }
>({
  createParams: (currentQuery) => ({
    title: currentQuery.title.trim() || undefined,
    businessType: currentQuery.businessType.trim() || undefined,
    businessNo: currentQuery.businessNo.trim() || undefined,
    status: currentQuery.status || undefined,
  }),
  createQuery: () => ({
    title: '',
    businessType: '',
    businessNo: '',
    status: '',
  }),
  fallbackErrorMessage: '加载我发起的流程失败',
  fetcher: async (params) => getWorkflowStartedPage(params),
});

const detailOpen = ref(false);
const currentInstanceId = ref<null | number>(null);
const canWithdraw = ref(false);

const baseColumns: TableColumnType[] = [
  {
    title: '标题',
    dataIndex: 'title',
    key: 'title',
    width: 220,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '业务类型',
    dataIndex: 'businessType',
    key: 'businessType',
    width: 140,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '业务编号',
    dataIndex: 'businessNo',
    key: 'businessNo',
    width: 160,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '当前状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: string }) =>
      text
        ? h(Tag, { color: workflowStatusColors[text] || 'default' }, () =>
            getWorkflowStatusLabel(text),
          )
        : renderAdminEmpty(text),
  },
  {
    title: '当前节点',
    dataIndex: 'currentNodeName',
    key: 'currentNodeName',
    width: 130,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '发起时间',
    dataIndex: 'startedAt',
    key: 'startedAt',
    width: 165,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '完成时间',
    dataIndex: 'finishedAt',
    key: 'finishedAt',
    width: 165,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 100,
    fixed: 'right',
    customRender: ({ record }: { record: WorkflowStartedItem }) =>
      h(
        Button,
        {
          size: 'small',
          type: 'link',
          onClick: () => openDetail(record),
        },
        () => '详情',
      ),
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
  systemColumnKeys: ['action'],
  tableId: 'workflow-started-list',
});

function openDetail(record: WorkflowStartedItem) {
  currentInstanceId.value = record.instanceId;
  canWithdraw.value = record.status === 'in_review';
  detailOpen.value = true;
}

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="标题">
          <Input
            v-model:value="query.title"
            allow-clear
            placeholder="请输入标题"
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="业务类型">
          <Input
            v-model:value="query.businessType"
            allow-clear
            placeholder="请输入业务类型"
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="业务编号">
          <Input
            v-model:value="query.businessNo"
            allow-clear
            placeholder="请输入业务编号"
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            class="w-full"
          />
        </AdminFilterField>
      </div>
    </template>

    <template #filter-actions>
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </template>

    <template #toolbar>
      <div class="text-sm font-medium text-slate-700">我发起的</div>
    </template>
    <template #toolbar-extra>
      <AdminTableColumnSettings
        v-model:open="settingsOpen"
        :columns="settingsColumns"
        @change-fixed="({ key, fixed }) => setColumnFixed(key, fixed)"
        @reorder="({ oldIndex, newIndex }) => reorderColumns(oldIndex, newIndex)"
        @reset="restoreDefaultColumns"
        @toggle-visible="({ key, visible }) => setColumnVisible(key, visible)"
      />
    </template>

    <AdminErrorAlert v-if="errorMsg" :message="errorMsg" />

    <Table
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ x: scrollX }"
      row-key="instanceId"
      size="middle"
      @change="onTableChange"
      @resizeColumn="handleResizeColumn"
    />

    <WorkflowInstanceDetailDrawer
      v-model:open="detailOpen"
      :can-withdraw="canWithdraw"
      :instance-id="currentInstanceId"
      @acted="fetchList"
    />
  </AdminPageShell>
</template>
