<script lang="ts" setup>
/**
 * 系统管理 - 接口管理
 * 列表 + 搜索 + 分页 + 刷新 + 编辑（无新增/删除，后端仅支持 GET/PUT）
 */
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  Modal,
  Table,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  getSysApiDetail,
  getSysApiPage,
  updateSysApi,
} from '#/api/core';
import type { SysApiItem, SysApiPageResult } from '#/api/core';

const loading = ref(false);
const tableData = ref<SysApiItem[]>([]);
const errorMsg = ref('');

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

const searchTitle = ref('');
const searchPath = ref('');
const searchAction = ref('');
const searchParentId = ref('');
const searchType = ref('');

async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
      title: searchTitle.value.trim() || undefined,
      path: searchPath.value.trim() || undefined,
      action: searchAction.value.trim() || undefined,
      parentId: searchParentId.value.trim() || undefined,
      type: searchType.value.trim() || undefined,
    };
    const res: SysApiPageResult = await getSysApiPage(params);
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
  searchPath.value = '';
  searchAction.value = '';
  searchParentId.value = '';
  searchType.value = '';
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
  { title: '标题', dataIndex: 'title', key: 'title', width: 140, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '地址', dataIndex: 'path', key: 'path', width: 180, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '请求方式', dataIndex: 'action', key: 'action', width: 90, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: 'Handle', dataIndex: 'handle', key: 'handle', width: 100, ellipsis: true, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  { title: '类型', dataIndex: 'type', key: 'type', width: 80, customRender: ({ text }: { text: string }) => renderEmpty(text) },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 160,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  { title: '操作', key: 'actions', width: 90, fixed: 'right' },
];

const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<number | null>(null);
const editForm = reactive({
  handle: '',
  title: '',
  path: '',
  type: '',
  action: '',
});

async function openEditModal(record: SysApiItem) {
  editId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getSysApiDetail(record.id);
    editForm.handle = detail.handle ?? '';
    editForm.title = detail.title ?? '';
    editForm.path = detail.path ?? '';
    editForm.type = detail.type ?? '';
    editForm.action = detail.action ?? '';
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
  if (!editForm.title?.trim()) return { ok: false, message: '请输入标题' };
  if (!editForm.path?.trim()) return { ok: false, message: '请输入地址' };
  return { ok: true };
}

async function onEditOk() {
  if (editId.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateSysApi(editId.value, {
      handle: editForm.handle?.trim() ?? '',
      title: editForm.title.trim(),
      path: editForm.path.trim(),
      type: editForm.type?.trim() ?? '',
      action: editForm.action?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    message.error(err?.message || err?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">接口管理</h2>
      <Button @click="fetchList">刷新</Button>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">标题：</span>
      <Input
        v-model:value="searchTitle"
        placeholder="请输入"
        allow-clear
        class="w-44"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">地址：</span>
      <Input
        v-model:value="searchPath"
        placeholder="请输入"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">请求方式：</span>
      <Input
        v-model:value="searchAction"
        placeholder="如 GET"
        allow-clear
        class="w-24"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">父级ID：</span>
      <Input
        v-model:value="searchParentId"
        placeholder="请输入"
        allow-clear
        class="w-24"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">类型：</span>
      <Input
        v-model:value="searchType"
        placeholder="请输入"
        allow-clear
        class="w-24"
        @press-enter="onSearch"
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
      :row-key="(record: SysApiItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'actions'">
          <Button type="link" size="small" @click="openEditModal(record as SysApiItem)">
            编辑
          </Button>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="editVisible"
      title="编辑接口"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="editVisible = false"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">加载详情中…</div>
      <Form
        v-else
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        class="mt-4"
      >
        <FormItem label="标题" required>
          <Input v-model:value="editForm.title" placeholder="请输入" allow-clear />
        </FormItem>
        <FormItem label="地址" required>
          <Input v-model:value="editForm.path" placeholder="请输入" allow-clear />
        </FormItem>
        <FormItem label="请求方式">
          <Input v-model:value="editForm.action" placeholder="如 GET/POST" allow-clear />
        </FormItem>
        <FormItem label="Handle">
          <Input v-model:value="editForm.handle" placeholder="请输入" allow-clear />
        </FormItem>
        <FormItem label="类型">
          <Input v-model:value="editForm.type" placeholder="请输入" allow-clear />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
