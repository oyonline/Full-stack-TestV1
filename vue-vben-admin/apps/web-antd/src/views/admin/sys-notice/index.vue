<script lang="ts" setup>
/**
 * 系统管理 - 公告管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除
 */
import { onMounted, reactive, ref } from 'vue';
import {
  Button,
  Form,
  FormItem,
  Input,
  InputNumber,
  Modal,
  Select,
  Table,
  Textarea,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createNotice,
  deleteNotice,
  getNoticeDetail,
  getNoticePage,
  updateNotice,
} from '#/api/core';
import type { NoticeItem, NoticePageResult } from '#/api/core';

const loading = ref(false);
const tableData = ref<NoticeItem[]>([]);
const errorMsg = ref('');

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

// 搜索条件
const searchTitle = ref('');
const searchType = ref('');
const searchStatus = ref('');

// 状态下拉选项
const statusOptions = [
  { value: '', label: '全部' },
  { value: '0', label: '禁用' },
  { value: '1', label: '启用' },
];

// 类型下拉选项
const typeOptions = [
  { value: '', label: '全部' },
  { value: '1', label: '通知' },
  { value: '2', label: '公告' },
];

async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      title?: string;
      type?: string;
      status?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchTitle.value.trim()) {
      params.title = searchTitle.value.trim();
    }
    if (searchType.value) {
      params.type = searchType.value;
    }
    if (searchStatus.value) {
      params.status = searchStatus.value;
    }
    const res: NoticePageResult = await getNoticePage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value = err?.message || err?.response?.data?.msg || '加载失败';
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
  searchType.value = '';
  searchStatus.value = '';
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

function renderStatus(val: string): string {
  if (val === '1') return '启用';
  if (val === '0') return '禁用';
  return val || '-';
}

function renderType(val: string): string {
  if (val === '1') return '通知';
  if (val === '2') return '公告';
  return val || '-';
}

function renderEmpty(value: string | null | undefined): string {
  return value ?? '-';
}

const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '标题', dataIndex: 'title', key: 'title', width: 200, ellipsis: true },
  { 
    title: '类型', 
    dataIndex: 'type', 
    key: 'type', 
    width: 100,
    customRender: ({ text }: { text: string }) => renderType(text),
  },
  { 
    title: '状态', 
    dataIndex: 'status', 
    key: 'status', 
    width: 100,
    customRender: ({ text }: { text: string }) => renderStatus(text),
  },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
  { 
    title: '备注', 
    dataIndex: 'remark', 
    key: 'remark', 
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

// 新增
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  title: '',
  content: '',
  type: '1',
  status: '1',
  sort: 0,
  remark: '',
});

function resetAddForm() {
  addForm.title = '';
  addForm.content = '';
  addForm.type = '1';
  addForm.status = '1';
  addForm.sort = 0;
  addForm.remark = '';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAddForm(): { ok: boolean; message?: string } {
  if (!addForm.title?.trim()) {
    return { ok: false, message: '请输入标题' };
  }
  return { ok: true };
}

async function onAddOk() {
  const v = validateAddForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createNotice({
      title: addForm.title.trim(),
      content: addForm.content?.trim() ?? '',
      type: addForm.type,
      status: addForm.status,
      sort: addForm.sort,
      remark: addForm.remark?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    message.error(err?.message || err?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

function onAddCancel() {
  addVisible.value = false;
}

// 编辑
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<number | null>(null);
const editForm = reactive({
  title: '',
  content: '',
  type: '1',
  status: '1',
  sort: 0,
  remark: '',
});

async function openEditModal(record: NoticeItem) {
  editId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getNoticeDetail(record.id);
    editForm.title = detail.title ?? '';
    editForm.content = detail.content ?? '';
    editForm.type = detail.type ?? '1';
    editForm.status = detail.status ?? '1';
    editForm.sort = detail.sort ?? 0;
    editForm.remark = detail.remark ?? '';
  } catch (e: unknown) {
    message.error('获取详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
  if (!editForm.title?.trim()) {
    return { ok: false, message: '请输入标题' };
  }
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
    await updateNotice(editId.value, {
      title: editForm.title.trim(),
      content: editForm.content?.trim() ?? '',
      type: editForm.type,
      status: editForm.status,
      sort: editForm.sort,
      remark: editForm.remark?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    message.error('编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

// 删除
function onDelete(record: NoticeItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除公告「${record.title}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteNotice([record.id]);
        message.success('删除成功');
        fetchList();
      } catch (e: unknown) {
        message.error('删除失败');
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
      <h2 class="text-lg font-medium">公告管理</h2>
      <div class="flex gap-2">
        <Button @click="fetchList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增公告</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">标题：</span>
      <Input
        v-model:value="searchTitle"
        placeholder="请输入标题"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">类型：</span>
      <Select
        v-model:value="searchType"
        :options="typeOptions"
        class="w-28"
        placeholder="请选择"
      />
      <span class="ml-2 text-sm text-gray-600">状态：</span>
      <Select
        v-model:value="searchStatus"
        :options="statusOptions"
        class="w-28"
        placeholder="请选择"
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
      :row-key="(record: NoticeItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="openEditModal(record as NoticeItem)">编辑</Button>
          <Button type="link" size="small" danger @click="onDelete(record as NoticeItem)">删除</Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal v-model:open="addVisible" title="新增公告" :confirm-loading="addSubmitting" @ok="onAddOk" @cancel="onAddCancel">
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="标题" required>
          <Input v-model:value="addForm.title" placeholder="请输入标题" allow-clear />
        </FormItem>
        <FormItem label="内容">
          <Textarea v-model:value="addForm.content" placeholder="请输入内容" :rows="4" />
        </FormItem>
        <FormItem label="类型">
          <Select v-model:value="addForm.type" :options="typeOptions.filter(o => o.value)" class="w-full" />
        </FormItem>
        <FormItem label="状态">
          <Select v-model:value="addForm.status" :options="statusOptions.filter(o => o.value)" class="w-full" />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="addForm.sort" class="w-full" />
        </FormItem>
        <FormItem label="备注">
          <Input v-model:value="addForm.remark" placeholder="请输入备注" allow-clear />
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑公告" :confirm-loading="editSubmitting" @ok="onEditOk" @cancel="onEditCancel">
      <div v-if="editLoading" class="py-8 text-center text-gray-400">加载中...</div>
      <Form v-else :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="标题" required>
          <Input v-model:value="editForm.title" placeholder="请输入标题" allow-clear />
        </FormItem>
        <FormItem label="内容">
          <Textarea v-model:value="editForm.content" placeholder="请输入内容" :rows="4" />
        </FormItem>
        <FormItem label="类型">
          <Select v-model:value="editForm.type" :options="typeOptions.filter(o => o.value)" class="w-full" />
        </FormItem>
        <FormItem label="状态">
          <Select v-model:value="editForm.status" :options="statusOptions.filter(o => o.value)" class="w-full" />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="editForm.sort" class="w-full" />
        </FormItem>
        <FormItem label="备注">
          <Input v-model:value="editForm.remark" placeholder="请输入备注" allow-clear />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
