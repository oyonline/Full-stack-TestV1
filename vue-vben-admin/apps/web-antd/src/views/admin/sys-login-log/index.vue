<script lang="ts" setup>
/**
 * 系统管理 - 登录日志
 * 列表 + 搜索 + 分页 + 删除 + 可选详情（只读，无新增/编辑）
 */
import { onMounted, ref } from 'vue';

import {
  Button,
  Descriptions,
  Input,
  Modal,
  Table,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import { deleteLoginLog, getLoginLogDetail, getLoginLogPage } from '#/api/core';
import type { SysLoginLogItem, SysLoginLogPageResult } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

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
  SysLoginLogItem,
  {
    beginTime: string;
    endTime: string;
    ipaddr: string;
    status: string;
    username: string;
  },
  {
    beginTime?: string;
    endTime?: string;
    ipaddr?: string;
    status?: string;
    username?: string;
  }
>({
  createParams: (currentQuery) => ({
    username: currentQuery.username.trim() || undefined,
    status: currentQuery.status.trim() || undefined,
    ipaddr: currentQuery.ipaddr.trim() || undefined,
    beginTime: currentQuery.beginTime.trim() || undefined,
    endTime: currentQuery.endTime.trim() || undefined,
  }),
  createQuery: () => ({
    username: '',
    status: '',
    ipaddr: '',
    beginTime: '',
    endTime: '',
  }),
  fallbackErrorMessage: '加载列表失败',
  fetcher: async (params) => {
    const res: SysLoginLogPageResult = await getLoginLogPage(params);
    return res;
  },
});

const renderEmpty = renderAdminEmpty;

const baseColumns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '登录账号', dataIndex: 'username', key: 'username', width: 100 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: 'IP',
    dataIndex: 'ipaddr',
    key: 'ipaddr',
    width: 120,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '归属地',
    dataIndex: 'loginLocation',
    key: 'loginLocation',
    width: 120,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '浏览器',
    dataIndex: 'browser',
    key: 'browser',
    width: 100,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '系统',
    dataIndex: 'os',
    key: 'os',
    width: 100,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '登录时间',
    dataIndex: 'loginTime',
    key: 'loginTime',
    width: 165,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '信息',
    dataIndex: 'msg',
    key: 'msg',
    width: 100,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
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
  tableId: 'sys-login-log-list',
});

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysLoginLogItem | null>(null);

async function openDetail(record: SysLoginLogItem) {
  detailRecord.value = null;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const d = await getLoginLogDetail(record.id);
    detailRecord.value = d;
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

function onDelete(record: SysLoginLogItem) {
  const name = record.username || `ID:${record.id}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除登录日志「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteLoginLog([record.id]);
        message.success('删除成功');
        fetchList();
      } catch (e: unknown) {
        const err = e as {
          message?: string;
          response?: { data?: { msg?: string } };
        };
        message.error(err?.message || err?.response?.data?.msg || '删除失败');
      }
    },
  });
}

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Audit</template>
    <template #title>登录日志</template>
    <template #description>
      查看登录状态、IP 和客户端信息，详情抽屉采用统一的审计信息展示密度。
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="登录账号">
          <Input
            v-model:value="query.username"
            placeholder="请输入登录账号"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Input
            v-model:value="query.status"
            placeholder="请输入状态"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="IP">
          <Input
            v-model:value="query.ipaddr"
            placeholder="请输入 IP"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="开始时间">
          <Input
            v-model:value="query.beginTime"
            placeholder="如 2024-01-01"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="结束时间">
          <Input
            v-model:value="query.endTime"
            placeholder="如 2024-12-31"
            allow-clear
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
        <div class="text-base font-semibold text-slate-900">登录日志列表</div>
        <p class="mt-1 text-sm text-slate-500">
          重点查看登录状态、设备信息与时间区间，可快速进入详情抽屉定位问题。
        </p>
      </div>
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

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysLoginLogItem) => record.id"
      :scroll="{ x: scrollX }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
      @resizeColumn="handleResizeColumn"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysLoginLog:query"
            @click="openDetail(record as SysLoginLogItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysLoginLog:remove"
            @click="onDelete(record as SysLoginLogItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="登录日志详情"
      width="560"
      :loading="detailLoading"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="登录概览" description="账号、状态和终端环境信息。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="ID">{{ detailRecord.id }}</Descriptions.Item>
            <Descriptions.Item label="登录账号">{{
              renderEmpty(detailRecord.username)
            }}</Descriptions.Item>
            <Descriptions.Item label="状态">{{
              renderEmpty(detailRecord.status)
            }}</Descriptions.Item>
            <Descriptions.Item label="信息">{{
              renderEmpty(detailRecord.msg)
            }}</Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>

        <AdminDetailSection title="终端信息" description="用于快速定位访问来源与设备环境。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="IP">{{
              renderEmpty(detailRecord.ipaddr)
            }}</Descriptions.Item>
            <Descriptions.Item label="归属地">{{
              renderEmpty(detailRecord.loginLocation)
            }}</Descriptions.Item>
            <Descriptions.Item label="浏览器">{{
              renderEmpty(detailRecord.browser)
            }}</Descriptions.Item>
            <Descriptions.Item label="系统">{{
              renderEmpty(detailRecord.os)
            }}</Descriptions.Item>
            <Descriptions.Item label="固件">{{
              renderEmpty(detailRecord.platform)
            }}</Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>

        <AdminDetailSection title="时间与备注" description="保留登录发生时间和审计附加说明。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="登录时间">{{
              formatAdminDateTime(detailRecord.loginTime)
            }}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{{
              formatAdminDateTime(detailRecord.createdAt)
            }}</Descriptions.Item>
            <Descriptions.Item label="备注">{{
              renderEmpty(detailRecord.remark)
            }}</Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
