<script lang="ts" setup>
/**
 * 系统管理 - 岗位管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除
 */
import { computed, onMounted, reactive, ref } from 'vue';

import { Button, Input, Modal, Select, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createPost,
  deletePost,
  getPostDetail,
  getPostPage,
  updatePost,
} from '#/api/core';
import type { SysPostItem, SysPostPageResult } from '#/api/core';
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
} from '#/utils/admin-crud';

const {
  errorMsg,
  fetchList: fetchPostList,
  loading,
  onReset,
  onSearch,
  onTableChange,
  pagination,
  query,
  tableData,
} = useAdminTable<
  SysPostItem,
  {
    postName: string;
    status: '' | number;
  },
  {
    postName?: string;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    postName: currentQuery.postName.trim() || undefined,
    status:
      currentQuery.status === '' ? undefined : Number(currentQuery.status),
  }),
  createQuery: () => ({
    postName: '',
    status: '' as '' | number,
  }),
  fallbackErrorMessage: '加载岗位列表失败',
  fetcher: async (params) => {
    const res: SysPostPageResult = await getPostPage(params);
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

/** 空值渲染 */
const renderEmpty = renderAdminEmpty;
const formatDateTime = formatAdminDateTime;

/** 表格列定义 */
const columns: TableColumnType[] = [
  { title: '岗位ID', dataIndex: 'postId', key: 'postId', width: 90 },
  { title: '岗位名称', dataIndex: 'postName', key: 'postName', width: 140 },
  { title: '岗位编码', dataIndex: 'postCode', key: 'postCode', width: 140 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
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
    width: 180,
    customRender: ({ text }: { text: string }) => formatDateTime(text),
    customCell: () => ({ style: { whiteSpace: 'nowrap' } }),
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
];

/* -------- 新增岗位 -------- */

/** 新增弹窗可见性 */
const addVisible = ref(false);
/** 新增提交中状态 */
const addSubmitting = ref(false);

/** 新增表单 */
const addForm = reactive({
  postName: '',
  postCode: '',
  sort: 0,
  status: 2 as number,
  remark: '',
});

/** 重置新增表单为默认值 */
function resetAddForm() {
  addForm.postName = '';
  addForm.postCode = '';
  addForm.sort = 0;
  addForm.status = 2;
  addForm.remark = '';
}

/** 打开新增弹窗 */
function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

/** 新增表单校验 */
function validateAddForm(): { ok: boolean; message?: string } {
  const name = addForm.postName?.trim() ?? '';
  const code = addForm.postCode?.trim() ?? '';
  if (!name) {
    return { ok: false, message: '请输入岗位名称' };
  }
  if (!code) {
    return { ok: false, message: '请输入岗位编码' };
  }
  return { ok: true };
}

/** 确认新增 */
async function onAddOk() {
  const v = validateAddForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createPost({
      postName: addForm.postName.trim(),
      postCode: addForm.postCode.trim(),
      sort: addForm.sort,
      status: addForm.status,
      remark: addForm.remark?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchPostList();
  } catch (e: any) {
    message.error(e?.message || e?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

/** 取消新增 */
function onAddCancel() {
  addVisible.value = false;
}

/* -------- 编辑岗位 -------- */

/** 编辑弹窗可见性 */
const editVisible = ref(false);
/** 编辑提交中状态 */
const editSubmitting = ref(false);
/** 编辑详情加载中状态 */
const editLoading = ref(false);
/** 当前编辑的岗位ID */
const editPostId = ref<number | null>(null);

/** 编辑表单 */
const editForm = reactive({
  postName: '',
  postCode: '',
  sort: 0,
  status: 2 as number,
  remark: '',
});

/** 打开编辑弹窗 */
async function openEditModal(record: SysPostItem) {
  editPostId.value = record.postId;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getPostDetail(record.postId);
    editForm.postName = detail.postName ?? '';
    editForm.postCode = detail.postCode ?? '';
    editForm.sort = detail.sort ?? 0;
    editForm.status = detail.status ?? 2;
    editForm.remark = detail.remark ?? '';
  } catch (e: any) {
    message.error(e?.message || '获取岗位详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

/** 编辑表单校验 */
function validateEditForm(): { ok: boolean; message?: string } {
  const name = editForm.postName?.trim() ?? '';
  const code = editForm.postCode?.trim() ?? '';
  if (!name) {
    return { ok: false, message: '请输入岗位名称' };
  }
  if (!code) {
    return { ok: false, message: '请输入岗位编码' };
  }
  return { ok: true };
}

/** 确认编辑 */
async function onEditOk() {
  if (editPostId.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updatePost(editPostId.value, {
      postName: editForm.postName.trim(),
      postCode: editForm.postCode.trim(),
      sort: editForm.sort,
      status: editForm.status,
      remark: editForm.remark?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchPostList();
  } catch (e: any) {
    message.error(e?.message || e?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

/** 取消编辑 */
function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除岗位 -------- */

/** 删除岗位 */
function onDelete(record: SysPostItem) {
  const postName = record.postName || `岗位ID:${record.postId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除岗位「${postName}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deletePost([record.postId]);
        message.success('删除成功');
        fetchPostList();
      } catch (e: any) {
        message.error(e?.message || e?.response?.data?.msg || '删除失败');
      }
    },
  });
}

/** 新增/编辑共用的状态下拉（不含"全部"） */
const statusEditOptions = [
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

const postFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'postName',
    label: '岗位名称',
    placeholder: '请输入岗位名称',
    required: true,
  },
  {
    component: 'input',
    field: 'postCode',
    label: '岗位编码',
    placeholder: '请输入岗位编码',
    required: true,
  },
  {
    component: 'input-number',
    field: 'sort',
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
  void fetchPostList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>岗位管理</template>
    <template #description>
      统一维护岗位名称、编码和状态。筛选区收口为标准网格，列表与弹窗采用更紧凑的后台样式。
    </template>
    <template #header-extra>
      <AdminActionButton type="primary" codes="admin:sysPost:add" @click="openAddModal">
        新增岗位
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="岗位名称">
          <Input
            v-model:value="query.postName"
            placeholder="请输入岗位名称"
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
        <div class="text-base font-semibold text-slate-900">岗位列表</div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysPostItem) => record.postId"
      :scroll="{ x: 920 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysPost:edit"
            @click="openEditModal(record as SysPostItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysPost:remove"
            @click="onDelete(record as SysPostItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增岗位"
      :confirm-loading="addSubmitting"
      :width="720"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <AdminModalFormFields :model="addForm" :fields="postFormFields" />
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑岗位"
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
        :fields="postFormFields"
      />
    </Modal>
  </AdminPageShell>
</template>
