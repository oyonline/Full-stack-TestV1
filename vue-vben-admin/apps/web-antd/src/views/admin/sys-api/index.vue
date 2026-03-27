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

import { getSysApiDetail, getSysApiPage, updateSysApi } from '#/api/core';
import type { SysApiItem, SysApiPageResult } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { renderAdminEmpty } from '#/utils/admin-crud';

const renderEmpty = renderAdminEmpty;

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
  SysApiItem,
  {
    action: string;
    parentId: string;
    path: string;
    title: string;
    type: string;
  },
  {
    action?: string;
    parentId?: string;
    path?: string;
    title?: string;
    type?: string;
  }
>({
  createParams: (currentQuery) => ({
    title: currentQuery.title.trim() || undefined,
    path: currentQuery.path.trim() || undefined,
    action: currentQuery.action.trim() || undefined,
    parentId: currentQuery.parentId.trim() || undefined,
    type: currentQuery.type.trim() || undefined,
  }),
  createQuery: () => ({
    title: '',
    path: '',
    action: '',
    parentId: '',
    type: '',
  }),
  fallbackErrorMessage: '加载列表失败',
  fetcher: async (params) => {
    const res: SysApiPageResult = await getSysApiPage(params);
    return res;
  },
});

const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  {
    title: '标题',
    dataIndex: 'title',
    key: 'title',
    width: 140,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '地址',
    dataIndex: 'path',
    key: 'path',
    width: 180,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '请求方式',
    dataIndex: 'action',
    key: 'action',
    width: 90,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: 'Handle',
    dataIndex: 'handle',
    key: 'handle',
    width: 100,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 80,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
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
const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysApiItem | null>(null);
const editForm = reactive({
  handle: '',
  title: '',
  path: '',
  type: '',
  action: '',
});

async function openDetail(record: SysApiItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detailRecord.value = await getSysApiDetail(record.id);
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

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
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>接口管理</template>
    <template #description>
      维护接口标题、路径和请求方式。筛选区按字段重要性重排，避免多个长输入框横向堆满整行。
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="标题">
          <Input
            v-model:value="query.title"
            placeholder="请输入接口标题"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="地址" :span="2">
          <Input
            v-model:value="query.path"
            placeholder="请输入接口地址"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="请求方式">
          <Input
            v-model:value="query.action"
            placeholder="如 GET"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="父级 ID">
          <Input
            v-model:value="query.parentId"
            placeholder="请输入父级 ID"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="类型">
          <Input
            v-model:value="query.type"
            placeholder="请输入类型"
            allow-clear
            @press-enter="onSearch"
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
        <div class="text-base font-semibold text-slate-900">接口列表</div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysApiItem) => record.id"
      :scroll="{ x: 1080 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'actions'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysApi:query"
            @click="openDetail(record as SysApiItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysApi:edit"
            @click="openEditModal(record as SysApiItem)"
          >
            编辑
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="editVisible"
      title="编辑接口"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="editVisible = false"
    >
      <div v-if="editLoading" class="py-10 text-center text-sm text-slate-400">
        加载详情中…
      </div>
      <Form v-else layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="标题" required class="mb-0">
          <Input
            v-model:value="editForm.title"
            placeholder="请输入接口标题"
            allow-clear
          />
        </FormItem>
        <FormItem label="请求方式" class="mb-0">
          <Input
            v-model:value="editForm.action"
            placeholder="如 GET/POST"
            allow-clear
          />
        </FormItem>
        <FormItem label="地址" required class="mb-0 md:col-span-2">
          <Input
            v-model:value="editForm.path"
            placeholder="请输入接口地址"
            allow-clear
          />
        </FormItem>
        <FormItem label="Handle" class="mb-0">
          <Input
            v-model:value="editForm.handle"
            placeholder="请输入 Handle"
            allow-clear
          />
        </FormItem>
        <FormItem label="类型" class="mb-0">
          <Input
            v-model:value="editForm.type"
            placeholder="请输入类型"
            allow-clear
          />
        </FormItem>
      </Form>
    </Modal>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="接口详情"
      :loading="detailLoading"
      width="680"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="基础信息" description="接口标题、地址和请求方式。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">标题</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.title) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">请求方式</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.action) }}</dd>
            </div>
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">地址</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.path) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">Handle</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.handle) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">类型</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.type) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">创建时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.createdAt) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">更新时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderEmpty(detailRecord.updatedAt) }}</dd>
            </div>
          </dl>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
