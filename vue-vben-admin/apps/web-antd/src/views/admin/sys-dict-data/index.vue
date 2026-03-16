<script lang="ts" setup>
/**
 * 系统管理 - 字典数据
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 刷新
 * dictType 筛选用字典类型下拉，来自 getDictTypeOptionSelect
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
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDictData,
  deleteDictData,
  getDictDataDetail,
  getDictDataPage,
  getDictTypeOptionSelect,
  updateDictData,
} from '#/api/core';
import type {
  DictTypeOption,
  SysDictDataItem,
  SysDictDataPageResult,
} from '#/api/core';

/** 表格加载状态 */
const loading = ref(false);
/** 表格数据 */
const tableData = ref<SysDictDataItem[]>([]);
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

/** 搜索 */
const searchDictLabel = ref('');
const searchDictValue = ref('');
const searchDictType = ref('');
const searchStatus = ref<string>('');

/** 字典类型下拉（供搜索与表单） */
const dictTypeOptions = ref<DictTypeOption[]>([]);

/** 状态下拉 */
const statusOptions = [
  { value: '', label: '全部' },
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
];

/** 加载字典类型选项 */
async function loadDictTypeOptions() {
  try {
    dictTypeOptions.value = await getDictTypeOptionSelect();
  } catch {
    dictTypeOptions.value = [];
  }
}

/** 获取字典数据列表 */
async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      dictLabel?: string;
      dictValue?: string;
      dictType?: string;
      status?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchDictLabel.value.trim()) {
      params.dictLabel = searchDictLabel.value.trim();
    }
    if (searchDictValue.value.trim()) {
      params.dictValue = searchDictValue.value.trim();
    }
    if (searchDictType.value.trim()) {
      params.dictType = searchDictType.value.trim();
    }
    if (searchStatus.value !== '') {
      params.status = searchStatus.value;
    }
    const res: SysDictDataPageResult = await getDictDataPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value =
      err?.message || err?.response?.data?.msg || '加载字典数据列表失败';
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
  searchDictLabel.value = '';
  searchDictValue.value = '';
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
function renderEmpty(value: string | number | null | undefined): string {
  if (value === null || value === undefined) return '-';
  return String(value);
}

/** 表格列定义 */
const columns: TableColumnType[] = [
  { title: '字典编码', dataIndex: 'dictCode', key: 'dictCode', width: 90 },
  { title: '字典标签', dataIndex: 'dictLabel', key: 'dictLabel', width: 120 },
  { title: '字典键值', dataIndex: 'dictValue', key: 'dictValue', width: 120 },
  { title: '字典类型', dataIndex: 'dictType', key: 'dictType', width: 120 },
  { title: '排序', dataIndex: 'dictSort', key: 'dictSort', width: 70 },
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
  dictLabel: '',
  dictValue: '',
  dictType: '',
  dictSort: 0,
  status: 2 as number,
  remark: '',
});

function resetAddForm() {
  addForm.dictLabel = '';
  addForm.dictValue = '';
  addForm.dictType = '';
  addForm.dictSort = 0;
  addForm.status = 2;
  addForm.remark = '';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAddForm(): { ok: boolean; message?: string } {
  const label = addForm.dictLabel?.trim() ?? '';
  const val = addForm.dictValue?.trim() ?? '';
  const type = addForm.dictType?.trim() ?? '';
  if (!label) return { ok: false, message: '请输入字典标签' };
  if (!val) return { ok: false, message: '请输入字典键值' };
  if (!type) return { ok: false, message: '请选择字典类型' };
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
    await createDictData({
      dictLabel: addForm.dictLabel.trim(),
      dictValue: addForm.dictValue.trim(),
      dictType: addForm.dictType.trim(),
      dictSort: addForm.dictSort,
      status: addForm.status,
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

/* -------- 编辑 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editDictCode = ref<number | null>(null);
const editForm = reactive({
  dictLabel: '',
  dictValue: '',
  dictType: '',
  dictSort: 0,
  status: 2 as number,
  remark: '',
});

async function openEditModal(record: SysDictDataItem) {
  editDictCode.value = record.dictCode;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getDictDataDetail(record.dictCode);
    editForm.dictLabel = detail.dictLabel ?? '';
    editForm.dictValue = detail.dictValue ?? '';
    editForm.dictType = detail.dictType ?? '';
    editForm.dictSort = detail.dictSort ?? 0;
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
  const label = editForm.dictLabel?.trim() ?? '';
  const val = editForm.dictValue?.trim() ?? '';
  const type = editForm.dictType?.trim() ?? '';
  if (!label) return { ok: false, message: '请输入字典标签' };
  if (!val) return { ok: false, message: '请输入字典键值' };
  if (!type) return { ok: false, message: '请选择字典类型' };
  return { ok: true };
}

async function onEditOk() {
  if (editDictCode.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateDictData(editDictCode.value, {
      dictLabel: editForm.dictLabel.trim(),
      dictValue: editForm.dictValue.trim(),
      dictType: editForm.dictType.trim(),
      dictSort: editForm.dictSort,
      status: editForm.status,
      remark: editForm.remark?.trim() ?? '',
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

function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除 -------- */
function onDelete(record: SysDictDataItem) {
  const name = record.dictLabel || record.dictValue || `编码:${record.dictCode}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除字典数据「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteDictData([record.dictCode]);
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
  loadDictTypeOptions();
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">字典数据</h2>
      <div class="flex gap-2">
        <Button @click="fetchList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">字典标签：</span>
      <Input
        v-model:value="searchDictLabel"
        placeholder="请输入字典标签"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="text-sm text-gray-600">字典键值：</span>
      <Input
        v-model:value="searchDictValue"
        placeholder="请输入字典键值"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="text-sm text-gray-600">字典类型：</span>
      <Select
        v-model:value="searchDictType"
        :options="dictTypeOptions"
        placeholder="请选择字典类型"
        allow-clear
        class="w-48"
        show-search
        :filter-option="
          (input: string, opt: { value: string }) =>
            opt?.value?.toLowerCase().includes(input?.toLowerCase())
        "
      />
      <span class="text-sm text-gray-600">状态：</span>
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
      :row-key="(record: SysDictDataItem) => record.dictCode"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysDictDataItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysDictDataItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增字典数据"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="字典类型" required>
          <Select
            v-model:value="addForm.dictType"
            :options="dictTypeOptions"
            placeholder="请选择字典类型"
            class="w-full"
            show-search
            :filter-option="
              (input: string, opt: { value: string }) =>
                opt?.value?.toLowerCase().includes(input?.toLowerCase())
            "
          />
        </FormItem>
        <FormItem label="字典标签" required>
          <Input
            v-model:value="addForm.dictLabel"
            placeholder="请输入字典标签"
            allow-clear
          />
        </FormItem>
        <FormItem label="字典键值" required>
          <Input
            v-model:value="addForm.dictValue"
            placeholder="请输入字典键值"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="addForm.dictSort"
            :min="0"
            class="w-full"
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
          <Input
            v-model:value="addForm.remark"
            placeholder="请输入备注"
            allow-clear
            type="textarea"
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑字典数据"
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
        <FormItem label="字典类型" required>
          <Select
            v-model:value="editForm.dictType"
            :options="dictTypeOptions"
            placeholder="请选择字典类型"
            class="w-full"
            show-search
            :filter-option="
              (input: string, opt: { value: string }) =>
                opt?.value?.toLowerCase().includes(input?.toLowerCase())
            "
          />
        </FormItem>
        <FormItem label="字典标签" required>
          <Input
            v-model:value="editForm.dictLabel"
            placeholder="请输入字典标签"
            allow-clear
          />
        </FormItem>
        <FormItem label="字典键值" required>
          <Input
            v-model:value="editForm.dictValue"
            placeholder="请输入字典键值"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="editForm.dictSort"
            :min="0"
            class="w-full"
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
          <Input
            v-model:value="editForm.remark"
            placeholder="请输入备注"
            allow-clear
            type="textarea"
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
