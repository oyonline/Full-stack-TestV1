<script lang="ts" setup>
/**
 * 系统管理 - 字典类型
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 刷新
 */
import { computed, onMounted, reactive, ref } from 'vue';

import { Button, Input, Modal, Select, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDictType,
  deleteDictType,
  getDictTypeDetail,
  getDictTypePage,
  updateDictType,
} from '#/api/core';
import type { SysDictTypeItem, SysDictTypePageResult } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { renderAdminEmpty } from '#/utils/admin-crud';

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
  SysDictTypeItem,
  {
    dictName: string;
    dictType: string;
    status: '' | number;
  },
  {
    dictName?: string;
    dictType?: string;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    dictName: currentQuery.dictName.trim() || undefined,
    dictType: currentQuery.dictType.trim() || undefined,
    status:
      currentQuery.status === '' ? undefined : Number(currentQuery.status),
  }),
  createQuery: () => ({
    dictName: '',
    dictType: '',
    status: '' as '' | number,
  }),
  fallbackErrorMessage: '加载字典类型列表失败',
  fetcher: async (params) => {
    const res: SysDictTypePageResult = await getDictTypePage(params);
    return res;
  },
});

/** 状态下拉选项 */
const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

/** 状态渲染 */
function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

const renderEmpty = renderAdminEmpty;

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

const dictTypeFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'dictName',
    label: '字典名称',
    placeholder: '请输入字典名称',
    required: true,
  },
  {
    component: 'input',
    field: 'dictType',
    label: '字典类型',
    placeholder: '请输入字典类型（如 sys_user_sex）',
    required: true,
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
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>字典类型</template>
    <template #description>
      维护系统字典类型定义，统一收口搜索区、操作区和表格区的视觉层级。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:sysDictType:add"
        @click="openAddModal"
      >
        新增
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="字典名称">
          <Input
            v-model:value="query.dictName"
            placeholder="请输入字典名称"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="字典类型">
          <Input
            v-model:value="query.dictType"
            placeholder="请输入字典类型"
            allow-clear
            @press-enter="onSearch"
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
        <div class="text-base font-semibold text-slate-900">字典类型列表</div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysDictTypeItem) => record.id"
      :scroll="{ x: 980 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysDictType:edit"
            @click="openEditModal(record as SysDictTypeItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="system:sysdicttype:remove"
            @click="onDelete(record as SysDictTypeItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增字典类型"
      :confirm-loading="addSubmitting"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <AdminModalFormFields :model="addForm" :fields="dictTypeFormFields" />
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑字典类型"
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
        :fields="dictTypeFormFields"
      />
    </Modal>
  </AdminPageShell>
</template>
