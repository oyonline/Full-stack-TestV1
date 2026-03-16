<script lang="ts" setup>
/**
 * 系统管理 - 操作日志
 * 列表 + 搜索 + 分页 + 删除 + 可选详情（只读，无新增/编辑）
 */
import { onMounted, ref } from 'vue';

import { Button, Descriptions, Drawer, Input, Modal, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  deleteOperaLog,
  getOperaLogDetail,
  getOperaLogPage,
} from '#/api/core';
import type {
  SysOperaLogItem,
  SysOperaLogPageResult,
} from '#/api/core';

const loading = ref(false);
const tableData = ref<SysOperaLogItem[]>([]);
const errorMsg = ref('');

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

const searchTitle = ref('');
const searchMethod = ref('');
const searchRequestMethod = ref('');
const searchOperUrl = ref('');
const searchOperIp = ref('');
const searchStatus = ref<number | ''>('');
const searchBeginTime = ref('');
const searchEndTime = ref('');

async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      title?: string;
      method?: string;
      requestMethod?: string;
      operUrl?: string;
      operIp?: string;
      status?: number;
      beginTime?: string;
      endTime?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchTitle.value.trim()) params.title = searchTitle.value.trim();
    if (searchMethod.value.trim()) params.method = searchMethod.value.trim();
    if (searchRequestMethod.value.trim()) params.requestMethod = searchRequestMethod.value.trim();
    if (searchOperUrl.value.trim()) params.operUrl = searchOperUrl.value.trim();
    if (searchOperIp.value.trim()) params.operIp = searchOperIp.value.trim();
    if (searchStatus.value !== '') params.status = searchStatus.value;
    if (searchBeginTime.value.trim()) params.beginTime = searchBeginTime.value.trim();
    if (searchEndTime.value.trim()) params.endTime = searchEndTime.value.trim();
    const res: SysOperaLogPageResult = await getOperaLogPage(params);
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
  searchTitle.value = '';
  searchMethod.value = '';
  searchRequestMethod.value = '';
  searchOperUrl.value = '';
  searchOperIp.value = '';
  searchStatus.value = '';
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
  { title: '操作模块', dataIndex: 'title', key: 'title', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '请求方式', dataIndex: 'requestMethod', key: 'requestMethod', width: 90, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '操作者', dataIndex: 'operName', key: 'operName', width: 90, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '部门', dataIndex: 'deptName', key: 'deptName', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '访问地址', dataIndex: 'operUrl', key: 'operUrl', width: 140, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '客户端IP', dataIndex: 'operIp', key: 'operIp', width: 110, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '操作时间', dataIndex: 'operTime', key: 'operTime', width: 165, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '状态', dataIndex: 'status', key: 'status', width: 70, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
];

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysOperaLogItem | null>(null);

async function openDetail(record: SysOperaLogItem) {
  detailRecord.value = null;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const d = await getOperaLogDetail(record.id);
    detailRecord.value = d;
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

function onDelete(record: SysOperaLogItem) {
  const name = record.title || record.operName || `ID:${record.id}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除操作日志「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteOperaLog([record.id]);
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
      <h2 class="text-lg font-medium">操作日志</h2>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">操作模块：</span>
      <Input
        v-model:value="searchTitle"
        placeholder="请输入"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">请求方式：</span>
      <Input
        v-model:value="searchRequestMethod"
        placeholder="GET/POST等"
        allow-clear
        class="w-28"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">函数：</span>
      <Input
        v-model:value="searchMethod"
        placeholder="请输入"
        allow-clear
        class="w-32"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">访问地址：</span>
      <Input
        v-model:value="searchOperUrl"
        placeholder="请输入"
        allow-clear
        class="w-44"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">IP：</span>
      <Input
        v-model:value="searchOperIp"
        placeholder="请输入"
        allow-clear
        class="w-32"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">状态：</span>
      <Input
        v-model:value="searchStatus"
        placeholder="1/2"
        allow-clear
        class="w-20"
      />
      <span class="ml-2 text-sm text-gray-600">开始时间：</span>
      <Input
        v-model:value="searchBeginTime"
        placeholder="如 2024-01-01"
        allow-clear
        class="w-32"
      />
      <span class="ml-2 text-sm text-gray-600">结束时间：</span>
      <Input
        v-model:value="searchEndTime"
        placeholder="如 2024-12-31"
        allow-clear
        class="w-32"
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
      :row-key="(record: SysOperaLogItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="openDetail(record as SysOperaLogItem)">
            详情
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysOperaLogItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <Drawer
      v-model:open="detailVisible"
      title="操作日志详情"
      width="560"
      :footer-style="{ textAlign: 'right' }"
    >
      <div v-if="detailLoading" class="py-8 text-center text-gray-400">加载中…</div>
      <Descriptions v-else-if="detailRecord" :column="1" bordered size="small">
        <Descriptions.Item label="ID">{{ detailRecord.id }}</Descriptions.Item>
        <Descriptions.Item label="操作模块">{{ detailRecord.title }}</Descriptions.Item>
        <Descriptions.Item label="请求方式">{{ detailRecord.requestMethod }}</Descriptions.Item>
        <Descriptions.Item label="操作者">{{ detailRecord.operName }}</Descriptions.Item>
        <Descriptions.Item label="部门">{{ detailRecord.deptName }}</Descriptions.Item>
        <Descriptions.Item label="访问地址">{{ detailRecord.operUrl }}</Descriptions.Item>
        <Descriptions.Item label="客户端IP">{{ detailRecord.operIp }}</Descriptions.Item>
        <Descriptions.Item label="访问位置">{{ detailRecord.operLocation }}</Descriptions.Item>
        <Descriptions.Item label="操作时间">{{ detailRecord.operTime }}</Descriptions.Item>
        <Descriptions.Item label="状态">{{ detailRecord.status }}</Descriptions.Item>
        <Descriptions.Item label="耗时">{{ detailRecord.latencyTime }}</Descriptions.Item>
        <Descriptions.Item label="请求参数"><pre class="max-h-32 overflow-auto text-xs">{{ detailRecord.operParam }}</pre></Descriptions.Item>
        <Descriptions.Item label="返回数据"><pre class="max-h-32 overflow-auto text-xs">{{ detailRecord.jsonResult }}</pre></Descriptions.Item>
        <Descriptions.Item label="备注">{{ detailRecord.remark }}</Descriptions.Item>
        <Descriptions.Item label="创建时间">{{ detailRecord.createdAt }}</Descriptions.Item>
      </Descriptions>
    </Drawer>
  </div>
</template>
