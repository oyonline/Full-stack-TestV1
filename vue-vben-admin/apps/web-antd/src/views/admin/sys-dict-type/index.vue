<script lang="ts" setup>
/**
 * 系统管理 - 字典类型
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 刷新
 */
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  Table,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDictType,
  deleteDictType,
  getDictTypeDetail,
  getDictTypePage,
  updateDictType,
} from '#/api/core';
import type { SysDictTypeItem, SysDictTypePageResult } from '#/api/core';

/** 表格加载状态 */
const loading = ref(false);
/** 表格数据 */
const tableData = ref<SysDictTypeItem[]>([]);
/** 错误提示 */
const errorMsg = ref('');

/** 分页状态 */
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

/** 搜索：字典名称（模糊） */
const searchDictName = ref('');
/** 搜索：字典类型（模糊） */
const searchDictType = ref('');
/** 搜索：状态 */
const searchStatus = ref<number | ''>('');

/** 状态下拉选项 */
const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

/** 获取字典类型列表 */
async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      dictName?: string;
      dictType?: string;
      status?: number;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchDictName.value.trim()) {
      params.dictName = searchDictName.value.trim();
    }
    if (searchDictType.value.trim()) {
      params.dictType = searchDictType.value.trim();
    }
    if (searchStatus.value !== '') {
      params.status = searchStatus.value;
    }
    const res: SysDictTypePageResult = await getDictTypePage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    errorMsg.value =
      err?.message || err?.response?.data?.msg || '加载字典类型列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

/** 查询 */
function onSearch() {
  pagination.value.current = 1;
  fetchList();
}

/** 重置 */
function onReset() {
  searchDictName.value = '';
  searchDictType.value = '';
  searchStatus.value = '';
  pagination.value.current = 1;
  fetchList();
}

/** 分页变化 */
function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchList();
}

/** 状态渲染 */
function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

/** 空值渲染 */
function renderEmpty(value: string | null | undefined): string {
  return value ?? '-';
}

/** 表格列定义 */
const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '字典名称', dataIndex: 'dictName', key: 'dictName', width: 140 },
  { title: '字典类型', dataIndex: 'dictType', key: 'dictType', width: 140 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: number }) => renderStatus(text),
  },
  {
    title: '备注',
    dataIndex: 'remark',
    key: 'remark',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 160,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
];

/* -------- 新增 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  dictName: '',
  dictType: '',
  status: 2 as number,
  remark: '',
});

function resetAddForm() {
  addForm.dictName = '';
  addForm.dictType = '';
  addForm.status = 2;
  addForm.remark = '';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAddForm(): { ok: boolean; message?: string } {
  const name = addForm.dictName?.trim() ?? '';
  const type = addForm.dictType?.trim() ?? '';
  if (!name) return { ok: false, message: '请输入字典名称' };
  if (!type) return { ok: false, message: '请输入字典类型' };
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
    await createDictType({
      dictName: addForm.dictName.trim(),
      dictType: addForm.dictType.trim(),
      status: addForm.status,
      remark: addForm.remark?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

function onAddCancel() {
  addVisible.value = false;
}

/* -------- 编辑 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<number | null>(null);
const editForm = reactive({
  dictName: '',
  dictType: '',
  status: 2 as number,
  remark: '',
});

async function openEditModal(record: SysDictTypeItem) {
  editId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getDictTypeDetail(record.id);
    editForm.dictName = detail.dictName ?? '';
    editForm.dictType = detail.dictType ?? '';
    editForm.status = detail.status ?? 2;
    editForm.remark = detail.remark ?? '';
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
  const name = editForm.dictName?.trim() ?? '';
  const type = editForm.dictType?.trim() ?? '';
  if (!name) return { ok: false, message: '请输入字典名称' };
  if (!type) return { ok: false, message: '请输入字典类型' };
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
    await updateDictType(editId.value, {
      dictName: editForm.dictName.trim(),
      dictType: editForm.dictType.trim(),
      status: editForm.status,
      remark: editForm.remark?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除 -------- */
function onDelete(record: SysDictTypeItem) {
  const name = record.dictName || record.dictType || `ID:${record.id}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除字典类型「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteDictType([record.id]);
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

const statusEditOptions = [
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">字典类型</h2>
      <div class="flex gap-2">
        <Button @click="fetchList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">字典名称：</span>
      <Input
        v-model:value="searchDictName"
        placeholder="请输入字典名称"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">字典类型：</span>
      <Input
        v-model:value="searchDictType"
        placeholder="请输入字典类型"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
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
      :row-key="(record: SysDictTypeItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysDictTypeItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysDictTypeItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增字典类型"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="字典名称" required>
          <Input
            v-model:value="addForm.dictName"
            placeholder="请输入字典名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="字典类型" required>
          <Input
            v-model:value="addForm.dictType"
            placeholder="请输入字典类型（如 sys_user_sex）"
            allow-clear
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="addForm.status"
            :options="statusEditOptions"
            class="w-full"
          />
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea
            v-model:value="addForm.remark"
            placeholder="请输入备注"
            allow-clear
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑字典类型"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <Form
        v-else
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        class="mt-4"
      >
        <FormItem label="字典名称" required>
          <Input
            v-model:value="editForm.dictName"
            placeholder="请输入字典名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="字典类型" required>
          <Input
            v-model:value="editForm.dictType"
            placeholder="请输入字典类型"
            allow-clear
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="editForm.status"
            :options="statusEditOptions"
            class="w-full"
          />
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea
            v-model:value="editForm.remark"
            placeholder="请输入备注"
            allow-clear
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
