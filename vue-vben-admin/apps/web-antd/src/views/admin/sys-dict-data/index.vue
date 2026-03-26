<script lang="ts" setup>
/**
 * 系统管理 - 字典数据
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 刷新
 * dictType 筛选用字典类型下拉，来自 getDictTypeOptionSelect
 */
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  type SelectProps,
  Input,
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
import type { DictTypeOption, SysDictDataItem } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

const renderEmpty = renderAdminEmpty;

/** 字典类型下拉（供搜索与表单） */
const dictTypeOptions = ref<DictTypeOption[]>([]);
const dictTypeSelectOptions = computed(() =>
  dictTypeOptions.value.map((item) => ({
    label: `${item.dictName} (${item.dictType})`,
    value: item.dictType,
  })),
);

/** 状态下拉 */
const statusOptions = [
  { value: '', label: '全部' },
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
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
  SysDictDataItem,
  {
    dictLabel: string;
    dictType: string;
    dictValue: string;
    status: string;
  },
  {
    dictLabel?: string;
    dictType?: string;
    dictValue?: string;
    status?: string;
  }
>({
  createParams: (currentQuery) => ({
    dictLabel: currentQuery.dictLabel.trim() || undefined,
    dictValue: currentQuery.dictValue.trim() || undefined,
    dictType: currentQuery.dictType.trim() || undefined,
    status: currentQuery.status || undefined,
  }),
  createQuery: () => ({
    dictLabel: '',
    dictValue: '',
    dictType: '',
    status: '',
  }),
  fallbackErrorMessage: '加载字典数据列表失败',
  fetcher: async (params) => getDictDataPage(params),
});

/** 加载字典类型选项 */
async function loadDictTypeOptions() {
  try {
    dictTypeOptions.value = await getDictTypeOptionSelect();
  } catch {
    dictTypeOptions.value = [];
  }
}

/** 状态渲染 */
function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

const filterDictTypeOption: SelectProps['filterOption'] = (input, option) => {
  const keyword = input.toLowerCase();
  return String(option?.value ?? '')
    .toLowerCase()
    .includes(keyword);
};

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
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
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
function onDelete(record: SysDictDataItem) {
  const name =
    record.dictLabel || record.dictValue || `编码:${record.dictCode}`;
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

const dictDataFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'select',
    field: 'dictType',
    label: '字典类型',
    options: dictTypeSelectOptions.value,
    placeholder: '请选择字典类型',
    required: true,
    showSearch: true,
    filterOption: filterDictTypeOption,
    span: 2,
  },
  {
    component: 'input',
    field: 'dictLabel',
    label: '字典标签',
    placeholder: '请输入字典标签',
    required: true,
  },
  {
    component: 'input',
    field: 'dictValue',
    label: '字典键值',
    placeholder: '请输入字典键值',
    required: true,
  },
  {
    component: 'input-number',
    field: 'dictSort',
    label: '排序',
    min: 0,
  },
  {
    component: 'select',
    field: 'status',
    label: '状态',
    options: statusEditOptions,
  },
  {
    component: 'textarea',
    field: 'remark',
    label: '备注',
    placeholder: '请输入备注',
    span: 2,
  },
]);

onMounted(() => {
  loadDictTypeOptions();
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>字典数据</template>
    <template #description>
      维护字典标签、键值与状态，筛选区与表格区采用统一后台布局，降低列表页的横向拥挤感。
    </template>
    <template #header-extra>
      <Button @click="fetchList">刷新</Button>
      <AdminActionButton
        type="primary"
        codes="admin:sysDictData:add"
        @click="openAddModal"
      >
        新增
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="字典标签">
          <Input
            v-model:value="query.dictLabel"
            placeholder="请输入字典标签"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="字典键值">
          <Input
            v-model:value="query.dictValue"
            placeholder="请输入字典键值"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="字典类型">
          <Select
            v-model:value="query.dictType"
            :options="dictTypeSelectOptions"
            placeholder="请选择字典类型"
            allow-clear
            show-search
            :filter-option="filterDictTypeOption"
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
        <div class="text-base font-semibold text-slate-900">字典数据列表</div>
        <p class="mt-1 text-sm text-slate-500">
          支持按标签、键值和字典类型快速定位数据项，列表与弹窗布局保持统一。
        </p>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysDictDataItem) => record.dictCode"
      :scroll="{ x: 1040 }"
      size="middle"
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysDictData:edit"
            @click="openEditModal(record as SysDictDataItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysDictData:remove"
            @click="onDelete(record as SysDictDataItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增字典数据"
      :confirm-loading="addSubmitting"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <AdminModalFormFields :model="addForm" :fields="dictDataFormFields" />
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑字典数据"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <AdminModalFormFields
        v-else
        :model="editForm"
        :fields="dictDataFormFields"
      />
    </Modal>
  </AdminPageShell>
</template>
