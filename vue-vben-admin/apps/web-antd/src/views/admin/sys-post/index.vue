<script lang="ts" setup>
/**
 * 系统管理 - 岗位管理
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
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createPost,
  deletePost,
  getPostDetail,
  getPostPage,
  updatePost,
} from '#/api/core';
import type { SysPostItem, SysPostPageResult } from '#/api/core';

/** 表格加载状态 */
const loading = ref(false);
/** 表格数据 */
const tableData = ref<SysPostItem[]>([]);
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

/** 搜索：岗位名称（模糊） */
const searchPostName = ref('');
/** 搜索：状态（1=停用，2=启用） */
const searchStatus = ref<number | ''>('');

/** 状态下拉选项 */
const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '停用' },
  { value: 2, label: '启用' },
];

/** 获取岗位列表 */
async function fetchPostList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      postName?: string;
      status?: number;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchPostName.value.trim()) {
      params.postName = searchPostName.value.trim();
    }
    if (searchStatus.value !== '') {
      params.status = searchStatus.value;
    }
    const res: SysPostPageResult = await getPostPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: any) {
    errorMsg.value = e?.message || e?.response?.data?.msg || '加载岗位列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

/** 查询按钮 */
function onSearch() {
  pagination.value.current = 1;
  fetchPostList();
}

/** 重置按钮 */
function onReset() {
  searchPostName.value = '';
  searchStatus.value = '';
  pagination.value.current = 1;
  fetchPostList();
}

/** 分页变化 */
function onTableChange(
  pag: { current?: number; pageSize?: number },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  _filters: unknown,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchPostList();
}

/** 状态渲染 */
function renderStatus(status: number): string {
  if (status === 1) return '停用';
  if (status === 2) return '启用';
  return String(status);
}

/** 空值渲染 */
function renderEmpty(value: string | null | undefined): string {
  return value || '-';
}

/** ISO 时间格式化为 YYYY-MM-DD HH:mm:ss，便于表格单行展示 */
function formatDateTime(isoStr: string | null | undefined): string {
  if (!isoStr) return '-';
  try {
    const d = new Date(isoStr);
    if (Number.isNaN(d.getTime())) return isoStr;
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  } catch {
    return isoStr;
  }
}

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

onMounted(() => {
  fetchPostList();
});
</script>

<template>
  <div class="p-4">
    <!-- 页面标题 -->
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">岗位管理</h2>
      <Button type="primary" @click="openAddModal">新增岗位</Button>
    </div>

    <!-- 搜索区 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">岗位名称：</span>
      <Input
        v-model:value="searchPostName"
        placeholder="请输入岗位名称"
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

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="mb-4 text-red-600">{{ errorMsg }}</div>

    <!-- 表格区 -->
    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysPostItem) => record.postId"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysPostItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysPostItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增岗位"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="岗位名称" required>
          <Input
            v-model:value="addForm.postName"
            placeholder="请输入岗位名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="岗位编码" required>
          <Input
            v-model:value="addForm.postCode"
            placeholder="请输入岗位编码"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="addForm.sort"
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

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑岗位"
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
        <FormItem label="岗位名称" required>
          <Input
            v-model:value="editForm.postName"
            placeholder="请输入岗位名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="岗位编码" required>
          <Input
            v-model:value="editForm.postCode"
            placeholder="请输入岗位编码"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="editForm.sort"
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
