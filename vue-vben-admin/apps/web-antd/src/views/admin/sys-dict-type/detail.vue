<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  Button,
  Drawer,
  Empty,
  Input,
  Modal,
  Select,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDictData,
  getDictDataDetail,
  getDictDataPage,
  getDictTypeDetail,
  updateDictData,
  updateDictType,
  deleteDictData,
} from '#/api/core';
import type {
  SysDictDataItem,
  SysDictDataPageResult,
  SysDictTypeItem,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import {
  formatAdminDateTime,
  renderAdminEmpty,
  resolveAdminErrorMessage,
} from '#/utils/admin-crud';

const route = useRoute();
const router = useRouter();

const dictId = computed(() => {
  const raw = route.query.dictId;
  const value = Array.isArray(raw) ? raw[0] : raw;
  const normalized = Number(value ?? 0);
  return Number.isFinite(normalized) ? normalized : 0;
});

const renderEmpty = renderAdminEmpty;

const statusOptions = [
  { value: '', label: '全部' },
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
];

const statusEditOptions = [
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

const detailLoading = ref(false);
const detailErrorMsg = ref('');
const detailRecord = ref<null | SysDictTypeItem>(null);
const totalDataCount = ref(0);

const {
  errorMsg: dataErrorMsg,
  fetchList: fetchDataList,
  loading: dataLoading,
  onTableChange: onDataTableChangeRaw,
  pagination: dataPagination,
  query: dataQuery,
  tableData: dataTableData,
} = useAdminTable<
  SysDictDataItem,
  {
    dictLabel: string;
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
    dictType: detailRecord.value?.dictType || undefined,
    status: currentQuery.status || undefined,
  }),
  createQuery: () => ({
    dictLabel: '',
    dictValue: '',
    status: '',
  }),
  fallbackErrorMessage: '加载字典数据失败',
  fetcher: async (params) => {
    if (!detailRecord.value?.dictType) {
      return {
        list: [],
        count: 0,
        pageIndex: params.pageIndex,
        pageSize: params.pageSize,
      } as SysDictDataPageResult;
    }
    return getDictDataPage(params);
  },
});

function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

function getStatusColor(status: number) {
  return status === 2 ? 'success' : 'default';
}

const dataColumns: TableColumnType[] = [
  {
    title: '字典标签',
    dataIndex: 'dictLabel',
    key: 'dictLabel',
    width: 140,
  },
  {
    title: '字典键值',
    dataIndex: 'dictValue',
    key: 'dictValue',
    width: 160,
  },
  {
    title: '排序',
    dataIndex: 'dictSort',
    key: 'dictSort',
    width: 80,
  },
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
    width: 168,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
];

const summaryItems = computed(() => {
  if (!detailRecord.value) return [];
  return [
    {
      label: '字典类型',
      value: detailRecord.value.dictType,
    },
    {
      label: '状态',
      value: renderStatus(detailRecord.value.status),
    },
    {
      label: '当前数据总数',
      value: `${totalDataCount.value} 条`,
    },
    {
      label: '备注',
      value: renderEmpty(detailRecord.value.remark),
    },
  ];
});

const pageTitle = computed(
  () => detailRecord.value?.dictName || '字典类型详情',
);

const pageDescription = computed(() => {
  if (!detailRecord.value) {
    return '查看字典类型详情，并在当前详情页下维护该类型对应的字典数据。';
  }
  return `${detailRecord.value.dictType} · 当前类型下共 ${totalDataCount.value} 条字典数据`;
});

function resetDataQuery() {
  Object.assign(dataQuery, {
    dictLabel: '',
    dictValue: '',
    status: '',
  });
  dataPagination.current = 1;
}

function clearDetailState(messageText = '') {
  detailRecord.value = null;
  detailErrorMsg.value = messageText;
  totalDataCount.value = 0;
  dataTableData.value = [];
  dataPagination.total = 0;
}

async function loadDataCount(dictType: string) {
  const result = await getDictDataPage({
    dictType,
    pageIndex: 1,
    pageSize: 1,
  });
  totalDataCount.value = result.count || 0;
}

async function loadDetailPage(resetFilters = true) {
  if (!dictId.value) {
    clearDetailState('缺少 dictId 参数');
    return;
  }

  detailLoading.value = true;
  detailErrorMsg.value = '';

  try {
    const detail = await getDictTypeDetail(dictId.value);
    detailRecord.value = detail;
    if (resetFilters) {
      resetDataQuery();
    }
    await Promise.all([loadDataCount(detail.dictType), fetchDataList()]);
  } catch (error) {
    clearDetailState(resolveAdminErrorMessage(error, '加载字典类型详情失败'));
  } finally {
    detailLoading.value = false;
  }
}

watch(
  () => dictId.value,
  () => {
    void loadDetailPage(true);
  },
  { immediate: true },
);

function goBackToDirectory() {
  void router.push('/admin/sys-dict-type');
}

function onDataSearch() {
  dataPagination.current = 1;
  void fetchDataList();
}

function onDataReset() {
  resetDataQuery();
  void fetchDataList();
}

function onDataTableChange(pag: { current?: number; pageSize?: number }) {
  onDataTableChangeRaw(pag);
}

const typeDrawerOpen = ref(false);
const typeDrawerSubmitting = ref(false);
const typeDictTypeLocked = ref(false);
const typeForm = reactive({
  dictName: '',
  dictType: '',
  status: 2 as number,
  remark: '',
});

const typeFormFields = computed<AdminFormFieldSchema[]>(() => [
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
    disabled: typeDictTypeLocked.value,
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

function openTypeDrawer() {
  if (!detailRecord.value) return;
  typeForm.dictName = detailRecord.value.dictName ?? '';
  typeForm.dictType = detailRecord.value.dictType ?? '';
  typeForm.status = detailRecord.value.status ?? 2;
  typeForm.remark = detailRecord.value.remark ?? '';
  typeDictTypeLocked.value = totalDataCount.value > 0;
  typeDrawerOpen.value = true;
}

function validateTypeForm(): { message?: string; ok: boolean } {
  const dictName = typeForm.dictName.trim();
  const dictType = typeForm.dictType.trim();
  if (!dictName) return { ok: false, message: '请输入字典名称' };
  if (!dictType) return { ok: false, message: '请输入字典类型' };
  return { ok: true };
}

async function onTypeDrawerOk() {
  if (!detailRecord.value) return;
  const validation = validateTypeForm();
  if (!validation.ok) {
    message.error(validation.message);
    return;
  }

  typeDrawerSubmitting.value = true;
  try {
    await updateDictType(detailRecord.value.id, {
      dictName: typeForm.dictName.trim(),
      dictType: typeForm.dictType.trim(),
      status: typeForm.status,
      remark: typeForm.remark.trim(),
    });
    typeDrawerOpen.value = false;
    message.success('保存成功');
    await loadDetailPage(false);
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '保存失败'));
  } finally {
    typeDrawerSubmitting.value = false;
  }
}

const dataModalOpen = ref(false);
const dataModalLoading = ref(false);
const dataModalSubmitting = ref(false);
const dataModalMode = ref<'create' | 'edit'>('create');
const editingDictCode = ref<null | number>(null);
const dataForm = reactive({
  dictType: '',
  dictLabel: '',
  dictValue: '',
  dictSort: 0,
  status: 2 as number,
  remark: '',
});

const dataFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'dictType',
    label: '字典类型',
    disabled: true,
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

function resetDataForm() {
  dataForm.dictType = detailRecord.value?.dictType ?? '';
  dataForm.dictLabel = '';
  dataForm.dictValue = '';
  dataForm.dictSort = 0;
  dataForm.status = 2;
  dataForm.remark = '';
  editingDictCode.value = null;
}

function openCreateDataModal() {
  if (!detailRecord.value) return;
  dataModalMode.value = 'create';
  resetDataForm();
  dataModalLoading.value = false;
  dataModalOpen.value = true;
}

async function openEditDataModal(record: SysDictDataItem) {
  dataModalMode.value = 'edit';
  editingDictCode.value = record.dictCode;
  dataModalLoading.value = true;
  dataModalOpen.value = true;
  try {
    const detail = await getDictDataDetail(record.dictCode);
    dataForm.dictType = detail.dictType ?? detailRecord.value?.dictType ?? '';
    dataForm.dictLabel = detail.dictLabel ?? '';
    dataForm.dictValue = detail.dictValue ?? '';
    dataForm.dictSort = detail.dictSort ?? 0;
    dataForm.status = detail.status ?? 2;
    dataForm.remark = detail.remark ?? '';
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取字典数据详情失败'));
    dataModalOpen.value = false;
  } finally {
    dataModalLoading.value = false;
  }
}

function validateDataForm(): { message?: string; ok: boolean } {
  const dictLabel = dataForm.dictLabel.trim();
  const dictValue = dataForm.dictValue.trim();
  if (!dictLabel) return { ok: false, message: '请输入字典标签' };
  if (!dictValue) return { ok: false, message: '请输入字典键值' };
  return { ok: true };
}

async function refreshAfterDataMutation() {
  if (!detailRecord.value) return;
  await Promise.all([loadDataCount(detailRecord.value.dictType), fetchDataList()]);
}

async function onDataModalOk() {
  if (!detailRecord.value) return;
  const validation = validateDataForm();
  if (!validation.ok) {
    message.error(validation.message);
    return;
  }

  dataModalSubmitting.value = true;
  const payload = {
    dictLabel: dataForm.dictLabel.trim(),
    dictType: detailRecord.value.dictType,
    dictValue: dataForm.dictValue.trim(),
    dictSort: dataForm.dictSort,
    status: dataForm.status,
    remark: dataForm.remark.trim(),
  };

  try {
    if (dataModalMode.value === 'create') {
      await createDictData(payload);
      message.success('新增成功');
    } else if (editingDictCode.value !== null) {
      await updateDictData(editingDictCode.value, payload);
      message.success('编辑成功');
    }
    dataModalOpen.value = false;
    await refreshAfterDataMutation();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '保存失败'));
  } finally {
    dataModalSubmitting.value = false;
  }
}

function handleDeleteData(record: SysDictDataItem) {
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
        await refreshAfterDataMutation();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}
</script>

<template>
  <AdminPageShell header-mode="compact">
    <template #eyebrow>System Admin</template>
    <template #title>{{ pageTitle }}</template>
    <template #description>{{ pageDescription }}</template>
    <template #header-extra>
      <Button @click="goBackToDirectory">返回目录</Button>
      <AdminActionButton
        type="default"
        codes="admin:sysDictType:edit"
        @click="openTypeDrawer"
      >
        编辑类型
      </AdminActionButton>
      <AdminActionButton
        type="primary"
        codes="admin:sysDictData:add"
        @click="openCreateDataModal"
      >
        新增字典数据
      </AdminActionButton>
    </template>

    <div v-if="detailLoading" class="py-16 text-center text-sm text-slate-400">
      加载字典类型详情中…
    </div>

    <template v-else-if="detailRecord">
      <section class="app-radius-panel border border-slate-200 bg-white p-4 shadow-sm md:p-5">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div class="space-y-1">
            <div class="text-lg font-semibold tracking-tight text-slate-900">
              {{ detailRecord.dictName }}
            </div>
            <div class="text-sm leading-6 text-slate-500">
              当前字典类型：{{ detailRecord.dictType }}
            </div>
          </div>
          <Tag :color="getStatusColor(detailRecord.status)">
            {{ renderStatus(detailRecord.status) }}
          </Tag>
        </div>

        <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="item in summaryItems"
            :key="item.label"
            class="rounded-2xl border border-slate-200 bg-slate-50/70 px-4 py-3"
          >
            <div class="text-xs leading-5 text-slate-500">{{ item.label }}</div>
            <div class="mt-1 text-sm font-medium leading-6 text-slate-900">
              {{ item.value }}
            </div>
          </div>
        </div>
      </section>

      <section class="app-radius-panel mt-4 border border-slate-200 bg-white p-4 shadow-sm md:p-5">
        <div class="mb-4 flex flex-col gap-2 border-b border-slate-100 pb-4 md:flex-row md:items-center md:justify-between">
          <div>
            <div class="text-base font-semibold text-slate-900">
              当前类型下的字典数据
            </div>
            <div class="mt-1 text-sm text-slate-500">
              当前类型下共 {{ totalDataCount }} 条字典数据
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
          <div class="grid gap-4 md:grid-cols-3">
            <AdminFilterField label="字典标签">
              <Input
                v-model:value="dataQuery.dictLabel"
                placeholder="请输入字典标签"
                allow-clear
                @press-enter="onDataSearch"
              />
            </AdminFilterField>
            <AdminFilterField label="字典键值">
              <Input
                v-model:value="dataQuery.dictValue"
                placeholder="请输入字典键值"
                allow-clear
                @press-enter="onDataSearch"
              />
            </AdminFilterField>
            <AdminFilterField label="状态">
              <Select
                v-model:value="dataQuery.status"
                :options="statusOptions"
                placeholder="请选择状态"
              />
            </AdminFilterField>
          </div>
          <div class="mt-4 flex flex-wrap items-center justify-end gap-2">
            <Button type="primary" @click="onDataSearch">查询</Button>
            <Button @click="onDataReset">重置</Button>
          </div>
        </div>

        <div class="mt-4 text-sm text-slate-500">
          当前筛选结果 {{ dataPagination.total }} 条
        </div>

        <AdminErrorAlert :message="detailErrorMsg || dataErrorMsg" />

        <Table
          class="mt-4"
          :columns="dataColumns"
          :data-source="dataTableData"
          :loading="dataLoading"
          :pagination="dataPagination"
          :row-key="(record: SysDictDataItem) => record.dictCode"
          :scroll="{ x: 980 }"
          size="middle"
          @change="(pag) => onDataTableChange(pag)"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'action'">
              <AdminActionButton
                type="link"
                size="small"
                codes="admin:sysDictData:edit"
                @click="openEditDataModal(record as SysDictDataItem)"
              >
                编辑
              </AdminActionButton>
              <AdminActionButton
                type="link"
                size="small"
                danger
                codes="admin:sysDictData:remove"
                @click="handleDeleteData(record as SysDictDataItem)"
              >
                删除
              </AdminActionButton>
            </template>
          </template>
        </Table>
      </section>
    </template>

    <section
      v-else
      class="app-radius-panel border border-slate-200 bg-white p-6 shadow-sm"
    >
      <AdminErrorAlert :message="detailErrorMsg" />
      <div class="flex min-h-[360px] items-center justify-center">
        <Empty description="未找到对应的字典类型">
          <Button type="primary" @click="goBackToDirectory">返回目录</Button>
        </Empty>
      </div>
    </section>

    <Drawer
      :open="typeDrawerOpen"
      title="编辑字典类型"
      :width="520"
      :footer-style="{ textAlign: 'right' }"
      @update:open="typeDrawerOpen = $event"
    >
      <div
        v-if="typeDictTypeLocked"
        class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-700"
      >
        当前类型下已有字典数据，编码已锁定，避免子数据仍指向旧的
        <code>dictType</code>。
      </div>
      <AdminModalFormFields :model="typeForm" :fields="typeFormFields" />
      <template #footer>
        <div class="flex justify-end gap-2">
          <Button @click="typeDrawerOpen = false">取消</Button>
          <Button
            type="primary"
            :loading="typeDrawerSubmitting"
            @click="onTypeDrawerOk"
          >
            保存
          </Button>
        </div>
      </template>
    </Drawer>

    <Modal
      v-model:open="dataModalOpen"
      :title="dataModalMode === 'create' ? '新增字典数据' : '编辑字典数据'"
      :confirm-loading="dataModalSubmitting"
      :ok-button-props="{ disabled: dataModalLoading }"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onDataModalOk"
      @cancel="dataModalOpen = false"
    >
      <div v-if="dataModalLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <AdminModalFormFields
        v-else
        :model="dataForm"
        :fields="dataFormFields"
      />
    </Modal>
  </AdminPageShell>
</template>
