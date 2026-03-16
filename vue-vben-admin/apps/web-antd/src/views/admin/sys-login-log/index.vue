<script lang="ts" setup>
/**
 * 系统管理 - 登录日志
 * 列表 + 搜索 + 分页 + 删除 + 可选详情（只读，无新增/编辑）
 */
import { onMounted, ref } from 'vue';

import { Button, Descriptions, Drawer, Input, Modal, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  deleteLoginLog,
  getLoginLogDetail,
  getLoginLogPage,
} from '#/api/core';
import type {
  SysLoginLogItem,
  SysLoginLogPageResult,
} from '#/api/core';

const loading = ref(false);
const tableData = ref<SysLoginLogItem[]>([]);
const errorMsg = ref('');

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

const searchUsername = ref('');
const searchStatus = ref('');
const searchIpaddr = ref('');
const searchBeginTime = ref('');
const searchEndTime = ref('');

async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      username?: string;
      status?: string;
      ipaddr?: string;
      beginTime?: string;
      endTime?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchUsername.value.trim()) params.username = searchUsername.value.trim();
    if (searchStatus.value.trim()) params.status = searchStatus.value.trim();
    if (searchIpaddr.value.trim()) params.ipaddr = searchIpaddr.value.trim();
    if (searchBeginTime.value.trim()) params.beginTime = searchBeginTime.value.trim();
    if (searchEndTime.value.trim()) params.endTime = searchEndTime.value.trim();
    const res: SysLoginLogPageResult = await getLoginLogPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value = err?.message || err?.response?.data?.msg || '加载列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  pagination.value.current = 1;
  fetchList();
}

function onReset() {
  searchUsername.value = '';
  searchStatus.value = '';
  searchIpaddr.value = '';
  searchBeginTime.value = '';
  searchEndTime.value = '';
  pagination.value.current = 1;
  fetchList();
}

function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchList();
}

function renderEmpty(value: string | null | undefined): string {
  return value !== undefined && value !== null ? String(value) : '-';
}

const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 100 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: 'IP', dataIndex: 'ipaddr', key: 'ipaddr', width: 120, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '归属地', dataIndex: 'loginLocation', key: 'loginLocation', width: 120, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '浏览器', dataIndex: 'browser', key: 'browser', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '系统', dataIndex: 'os', key: 'os', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '登录时间', dataIndex: 'loginTime', key: 'loginTime', width: 165, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '信息', dataIndex: 'msg', key: 'msg', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
];

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
        const err = e as { message?: string; response?: { data?: { msg?: string } } };
        message.error(err?.message || err?.response?.data?.msg || '删除失败');
      }
    },
  });
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">登录日志</h2>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">用户名：</span>
      <Input
        v-model:value="searchUsername"
        placeholder="请输入"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">状态：</span>
      <Input
        v-model:value="searchStatus"
        placeholder="请输入"
        allow-clear
        class="w-32"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">IP：</span>
      <Input
        v-model:value="searchIpaddr"
        placeholder="请输入"
        allow-clear
        class="w-36"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">开始时间：</span>
      <Input
        v-model:value="searchBeginTime"
        placeholder="如 2024-01-01"
        allow-clear
        class="w-36"
      />
      <span class="ml-2 text-sm text-gray-600">结束时间：</span>
      <Input
        v-model:value="searchEndTime"
        placeholder="如 2024-12-31"
        allow-clear
        class="w-36"
      />
      <Button type="primary" size="small" @click="onSearch">查询</Button>
      <Button size="small" @click="onReset">重置</Button>
    </div>

    <div v-if="errorMsg" class="mb-4 text-red-600">{{ errorMsg }}</div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysLoginLogItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="openDetail(record as SysLoginLogItem)">
            详情
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysLoginLogItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <Drawer
      v-model:open="detailVisible"
      title="登录日志详情"
      width="480"
      :footer-style="{ textAlign: 'right' }"
    >
      <div v-if="detailLoading" class="py-8 text-center text-gray-400">加载中…</div>
      <Descriptions v-else-if="detailRecord" :column="1" bordered size="small">
        <Descriptions.Item label="ID">{{ detailRecord.id }}</Descriptions.Item>
        <Descriptions.Item label="用户名">{{ detailRecord.username }}</Descriptions.Item>
        <Descriptions.Item label="状态">{{ detailRecord.status }}</Descriptions.Item>
        <Descriptions.Item label="IP">{{ detailRecord.ipaddr }}</Descriptions.Item>
        <Descriptions.Item label="归属地">{{ detailRecord.loginLocation }}</Descriptions.Item>
        <Descriptions.Item label="浏览器">{{ detailRecord.browser }}</Descriptions.Item>
        <Descriptions.Item label="系统">{{ detailRecord.os }}</Descriptions.Item>
        <Descriptions.Item label="固件">{{ detailRecord.platform }}</Descriptions.Item>
        <Descriptions.Item label="登录时间">{{ detailRecord.loginTime }}</Descriptions.Item>
        <Descriptions.Item label="信息">{{ detailRecord.msg }}</Descriptions.Item>
        <Descriptions.Item label="备注">{{ detailRecord.remark }}</Descriptions.Item>
        <Descriptions.Item label="创建时间">{{ detailRecord.createdAt }}</Descriptions.Item>
      </Descriptions>
    </Drawer>
  </div>
</template>
