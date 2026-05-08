<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';

import type { SkuBrandItem, SkuBrandPageResult } from '#/api/core';

/**
 * SKU 模块 - 品牌管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 单条/批量删除 + 状态切换
 */
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Table,
} from 'ant-design-vue';

import {
  createSkuBrand,
  deleteSkuBrand,
  getSkuBrandDetail,
  getSkuBrandPage,
  updateSkuBrand,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import {
  formatAdminDateTime,
  resolveAdminErrorMessage,
} from '#/utils/admin-crud';

import LogoUpload from './logo-upload.vue';

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
  SkuBrandItem,
  {
    brandName: string;
    status: '' | number;
  },
  {
    brandName?: string;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    brandName: currentQuery.brandName.trim() || undefined,
    status:
      currentQuery.status === '' ? undefined : Number(currentQuery.status),
  }),
  createQuery: () => ({
    brandName: '',
    status: '' as '' | number,
  }),
  fallbackErrorMessage: '加载品牌列表失败',
  fetcher: async (params) => {
    const res: SkuBrandPageResult = await getSkuBrandPage(params);
    return res;
  },
});

/** 状态下拉选项（含全部） */
const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '禁用' },
  { value: 2, label: '启用' },
];

/** 新增/编辑共用的状态下拉（不含全部） */
const statusEditOptions = [
  { value: 1, label: '禁用' },
  { value: 2, label: '启用' },
];

function renderStatus(status: number): string {
  if (status === 1) return '禁用';
  if (status === 2) return '启用';
  return String(status);
}

const formatDateTime = formatAdminDateTime;

/** 表格列定义 */
const baseColumns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'brandId', key: 'brandId', width: 80 },
  { title: '品牌名称', dataIndex: 'brandName', key: 'brandName', width: 180 },
  {
    title: 'Logo',
    dataIndex: 'brandLogoUrl',
    key: 'brandLogoUrl',
    width: 100,
  },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 90,
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ text }: { text: string }) => formatDateTime(text),
    customCell: () => ({ style: { whiteSpace: 'nowrap' } }),
  },
  {
    title: '操作',
    key: 'action',
    width: 220,
    fixed: 'right',
  },
];

const {
  handleResizeColumn,
  reorderColumns,
  restoreDefaultColumns,
  scrollX,
  setColumnFixed,
  setColumnVisible,
  settingsColumns,
  settingsOpen,
  tableColumns,
} = useAdminTableColumns(baseColumns, {
  systemColumnKeys: ['action'],
  tableId: 'sku-brand-list',
});

/* -------- 批量选择 -------- */
const selectedRowKeys = ref<number[]>([]);

function onSelectChange(keys: (number | string)[]) {
  selectedRowKeys.value = keys.map(Number);
}

async function onBatchDelete() {
  const ids = selectedRowKeys.value;
  if (ids.length === 0) {
    message.warning('请先选择要删除的品牌');
    return;
  }
  Modal.confirm({
    title: '确认批量删除',
    content: `确定要删除选中的 ${ids.length} 条品牌吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSkuBrand([...ids]);
        message.success('删除成功');
        selectedRowKeys.value = [];
        await fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

/* -------- 新增品牌 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  brandName: '',
  brandLogoUrl: '',
  sort: 0,
  status: 2 as number,
});

function resetAddForm() {
  addForm.brandName = '';
  addForm.brandLogoUrl = '';
  addForm.sort = 0;
  addForm.status = 2;
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateBrandForm(form: { brandName: string }): {
  message?: string;
  ok: boolean;
} {
  if (!form.brandName.trim()) {
    return { ok: false, message: '请输入品牌名称' };
  }
  return { ok: true };
}

async function onAddOk() {
  const v = validateBrandForm(addForm);
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createSkuBrand({
      brandName: addForm.brandName.trim(),
      brandLogoUrl: addForm.brandLogoUrl || undefined,
      sort: addForm.sort,
      status: addForm.status,
    });
    message.success('新增成功');
    addVisible.value = false;
    await fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
  } finally {
    addSubmitting.value = false;
  }
}

function onAddCancel() {
  addVisible.value = false;
}

/* -------- 编辑品牌 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editBrandId = ref<null | number>(null);
const editForm = reactive({
  brandName: '',
  brandLogoUrl: '',
  sort: 0,
  status: 2 as number,
});

async function openEditModal(record: SkuBrandItem) {
  editBrandId.value = record.brandId;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getSkuBrandDetail(record.brandId);
    editForm.brandName = detail.brandName ?? '';
    editForm.brandLogoUrl = detail.brandLogoUrl ?? '';
    editForm.sort = detail.sort ?? 0;
    editForm.status = detail.status ?? 2;
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取品牌详情失败'));
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

async function onEditOk() {
  if (editBrandId.value === null) return;
  const v = validateBrandForm(editForm);
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateSkuBrand(editBrandId.value, {
      brandName: editForm.brandName.trim(),
      brandLogoUrl: editForm.brandLogoUrl || undefined,
      sort: editForm.sort,
      status: editForm.status,
    });
    message.success('编辑成功');
    editVisible.value = false;
    await fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '编辑失败'));
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除品牌（单条） -------- */
function onDelete(record: SkuBrandItem) {
  const name = record.brandName || `品牌ID:${record.brandId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除品牌「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSkuBrand([record.brandId]);
        message.success('删除成功');
        await fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

/* -------- 状态切换 -------- */
async function onToggleStatus(record: SkuBrandItem) {
  const nextStatus = record.status === 2 ? 1 : 2;
  const label = nextStatus === 2 ? '启用' : '禁用';
  try {
    await updateSkuBrand(record.brandId, {
      brandName: record.brandName,
      brandLogoUrl: record.brandLogoUrl || undefined,
      sort: record.sort,
      status: nextStatus,
    });
    message.success(`${label}成功`);
    await fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, `${label}失败`));
  }
}

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>SKU Center</template>
    <template #title>品牌管理</template>
    <template #description>
      统一维护商品品牌字典，包含品牌名称、Logo、排序与启用状态。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:brand:add"
        @click="openAddModal"
      >
        新增品牌
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="品牌名称">
          <Input
            v-model:value="query.brandName"
            placeholder="请输入品牌名称"
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
        <div class="text-base font-semibold text-slate-900">品牌列表</div>
      </div>
    </template>
    <template #toolbar-extra>
      <AdminActionButton
        danger
        codes="admin:brand:remove"
        :disabled="selectedRowKeys.length === 0"
        @click="onBatchDelete"
      >
        批量删除
      </AdminActionButton>
      <AdminTableColumnSettings
        v-model:open="settingsOpen"
        :columns="settingsColumns"
        @change-fixed="({ key, fixed }) => setColumnFixed(key, fixed)"
        @reorder="
          ({ oldIndex, newIndex }) => reorderColumns(oldIndex, newIndex)
        "
        @reset="restoreDefaultColumns"
        @toggle-visible="({ key, visible }) => setColumnVisible(key, visible)"
      />
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SkuBrandItem) => record.brandId"
      :row-selection="{
        selectedRowKeys,
        onChange: onSelectChange,
      }"
      :scroll="{ x: scrollX }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
      @resize-column="handleResizeColumn"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'brandLogoUrl'">
          <img
            v-if="(record as SkuBrandItem).brandLogoUrl"
            :src="(record as SkuBrandItem).brandLogoUrl"
            class="size-12 rounded-sm border border-slate-200 object-contain"
            alt="logo"
          />
          <span v-else class="text-xs text-slate-400">—</span>
        </template>
        <template v-else-if="column.key === 'status'">
          <span
            :class="
              (record as SkuBrandItem).status === 2
                ? 'text-green-600'
                : 'text-slate-400'
            "
          >
            {{ renderStatus((record as SkuBrandItem).status) }}
          </span>
        </template>
        <template v-else-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:brand:edit"
            @click="openEditModal(record as SkuBrandItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:brand:edit"
            @click="onToggleStatus(record as SkuBrandItem)"
          >
            {{ (record as SkuBrandItem).status === 2 ? '禁用' : '启用' }}
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:brand:remove"
            @click="onDelete(record as SkuBrandItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增品牌"
      :confirm-loading="addSubmitting"
      :width="640"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <div class="mt-4 grid gap-4">
        <div>
          <div class="mb-1 text-sm">
            品牌名称<span class="text-red-500"> *</span>
          </div>
          <Input
            v-model:value="addForm.brandName"
            placeholder="请输入品牌名称"
            allow-clear
          />
        </div>
        <div>
          <div class="mb-1 text-sm">Logo</div>
          <LogoUpload v-model="addForm.brandLogoUrl" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <div class="mb-1 text-sm">排序</div>
            <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
          </div>
          <div>
            <div class="mb-1 text-sm">状态</div>
            <Select
              v-model:value="addForm.status"
              :options="statusEditOptions"
              class="w-full"
            />
          </div>
        </div>
      </div>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑品牌"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      :width="640"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <div v-else class="mt-4 grid gap-4">
        <div>
          <div class="mb-1 text-sm">
            品牌名称<span class="text-red-500"> *</span>
          </div>
          <Input
            v-model:value="editForm.brandName"
            placeholder="请输入品牌名称"
            allow-clear
          />
        </div>
        <div>
          <div class="mb-1 text-sm">Logo</div>
          <LogoUpload
            v-model="editForm.brandLogoUrl"
            :business-id="editBrandId ?? undefined"
          />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <div class="mb-1 text-sm">排序</div>
            <InputNumber
              v-model:value="editForm.sort"
              :min="0"
              class="w-full"
            />
          </div>
          <div>
            <div class="mb-1 text-sm">状态</div>
            <Select
              v-model:value="editForm.status"
              :options="statusEditOptions"
              class="w-full"
            />
          </div>
        </div>
      </div>
    </Modal>
  </AdminPageShell>
</template>
