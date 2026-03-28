<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button, Input, Modal, Select, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDictType,
  deleteDictType,
  getDictDataPage,
  getDictTypePage,
} from '#/api/core';
import type { SysDictTypeItem, SysDictTypePageResult } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, resolveAdminErrorMessage } from '#/utils/admin-crud';

const router = useRouter();

const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

const statusEditOptions = [
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

const {
  errorMsg,
  fetchList,
  loading,
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
    const result: SysDictTypePageResult = await getDictTypePage(params);
    return result;
  },
});

function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

const columns: TableColumnType[] = [
  {
    title: '字典名称',
    dataIndex: 'dictName',
    key: 'dictName',
    width: 160,
  },
  {
    title: '字典类型',
    dataIndex: 'dictType',
    key: 'dictType',
    width: 180,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 90,
    customRender: ({ text }: { text: number }) => renderStatus(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
    fixed: 'right',
  },
];

const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  dictName: '',
  dictType: '',
  status: 2 as number,
  remark: '',
});

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

function openDetail(record: SysDictTypeItem) {
  void router.push({
    path: '/admin/sys-dict-type/detail',
    query: { dictId: String(record.id) },
  });
}

function onSearch() {
  pagination.current = 1;
  void fetchList();
}

function onReset() {
  Object.assign(query, {
    dictName: '',
    dictType: '',
    status: '',
  });
  pagination.current = 1;
  void fetchList();
}

function validateAddForm(): { message?: string; ok: boolean } {
  const dictName = addForm.dictName.trim();
  const dictType = addForm.dictType.trim();
  if (!dictName) return { ok: false, message: '请输入字典名称' };
  if (!dictType) return { ok: false, message: '请输入字典类型' };
  return { ok: true };
}

async function resolveCreatedDictTypeId(
  dictType: string,
  dictName: string,
): Promise<number> {
  const result = await getDictTypePage({
    dictName,
    dictType,
    pageIndex: 1,
    pageSize: 10,
  });

  return (
    result.list.find((item) => item.dictType === dictType)?.id ??
    result.list[0]?.id ??
    0
  );
}

async function onAddOk() {
  const validation = validateAddForm();
  if (!validation.ok) {
    message.error(validation.message);
    return;
  }

  const payload = {
    dictName: addForm.dictName.trim(),
    dictType: addForm.dictType.trim(),
    status: addForm.status,
    remark: addForm.remark.trim(),
  };

  addSubmitting.value = true;
  try {
    const created = await createDictType(payload);
    let targetId = Number(created ?? 0);

    if (!targetId) {
      targetId = await resolveCreatedDictTypeId(payload.dictType, payload.dictName);
    }

    addVisible.value = false;
    message.success('新增成功');

    if (targetId) {
      void router.push({
        path: '/admin/sys-dict-type/detail',
        query: { dictId: String(targetId) },
      });
      return;
    }

    await fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
  } finally {
    addSubmitting.value = false;
  }
}

async function handleDelete(record: SysDictTypeItem) {
  try {
    const result = await getDictDataPage({
      dictType: record.dictType,
      pageIndex: 1,
      pageSize: 1,
    });
    if ((result.count || 0) > 0) {
      message.warning('请先删除该类型下的字典数据');
      return;
    }
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '校验字典数据失败'));
    return;
  }

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
        await fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
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
    <template #eyebrow>System Admin</template>
    <template #title>字典管理</template>
    <template #description>
      这里是字典类型目录页，先定位到目标类型，再进入详情页维护该类型下的字典数据。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:sysDictType:add"
        @click="openAddModal"
      >
        新增字典类型
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
        <div class="text-base font-semibold text-slate-900">字典类型目录</div>
        <div class="mt-1 text-sm text-slate-500">
          当前页只负责浏览、筛选和进入类型详情，不直接承担字典数据维护。
        </div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysDictTypeItem) => record.id"
      :scroll="{ x: 920 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            class="px-0"
            @click="openDetail(record as SysDictTypeItem)"
          >
            进入详情
          </Button>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="system:sysdicttype:remove"
            @click="handleDelete(record as SysDictTypeItem)"
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
      @cancel="addVisible = false"
    >
      <AdminModalFormFields :model="addForm" :fields="dictTypeFormFields" />
    </Modal>
  </AdminPageShell>
</template>
