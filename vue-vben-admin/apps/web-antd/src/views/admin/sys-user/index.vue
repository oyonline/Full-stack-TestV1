<script lang="ts" setup>
/**
 * 系统管理 - 用户管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 部门/角色/岗位关联
 */
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  Table,
  TreeSelect,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createSysUser,
  deleteSysUser,
  getDeptTreeApi,
  getPostPage,
  getRolePage,
  getSysUserDetail,
  getSysUserPage,
  updateSysUser,
} from '#/api/core';
import type {
  DeptLabel,
  SysUserItem,
  SysUserPageResult,
} from '#/api/core';

/** 表格加载状态 */
const loading = ref(false);
const tableData = ref<SysUserItem[]>([]);
const errorMsg = ref('');

/** 分页 */
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

/** 搜索 */
const searchUsername = ref('');
const searchNickName = ref('');
const searchPhone = ref('');
const searchStatus = ref<string>('');

const statusOptions = [
  { value: '', label: '全部' },
  { value: '0', label: '停用' },
  { value: '1', label: '启用' },
];

/** 部门树（用于下拉） */
const deptTreeRaw = ref<DeptLabel[]>([]);

function buildDeptTreeOptions(nodes: DeptLabel[]): { value: number; label: string; children?: { value: number; label: string; children?: unknown[] }[] }[] {
  return (nodes || []).map((n) => ({
    value: n.id,
    label: n.label,
    children: n.children?.length ? buildDeptTreeOptions(n.children as DeptLabel[]) : undefined,
  }));
}

const deptTreeOptions = computed(() => buildDeptTreeOptions(deptTreeRaw.value || []));

/** 角色选项 */
const roleOptions = ref<{ value: number; label: string }[]>([]);
/** 岗位选项 */
const postOptions = ref<{ value: number; label: string }[]>([]);

async function loadDeptTree() {
  try {
    deptTreeRaw.value = await getDeptTreeApi();
  } catch {
    deptTreeRaw.value = [];
  }
}

async function loadRoleOptions() {
  try {
    const res = await getRolePage({ pageIndex: 1, pageSize: 500 });
    roleOptions.value = (res.list || []).map((r) => ({
      value: r.roleId,
      label: r.roleName || `角色${r.roleId}`,
    }));
  } catch {
    roleOptions.value = [];
  }
}

async function loadPostOptions() {
  try {
    const res = await getPostPage({ pageIndex: 1, pageSize: 500 });
    postOptions.value = (res.list || []).map((p) => ({
      value: p.postId,
      label: p.postName || `岗位${p.postId}`,
    }));
  } catch {
    postOptions.value = [];
  }
}

/** 获取用户列表 */
async function fetchUserList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: Record<string, unknown> = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchUsername.value.trim()) params.username = searchUsername.value.trim();
    if (searchNickName.value.trim()) params.nickName = searchNickName.value.trim();
    if (searchPhone.value.trim()) params.phone = searchPhone.value.trim();
    if (searchStatus.value !== '') params.status = searchStatus.value;
    const res: SysUserPageResult = await getSysUserPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value = err?.message || err?.response?.data?.msg || '加载用户列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  pagination.value.current = 1;
  fetchUserList();
}

function onReset() {
  searchUsername.value = '';
  searchNickName.value = '';
  searchPhone.value = '';
  searchStatus.value = '';
  pagination.value.current = 1;
  fetchUserList();
}

function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchUserList();
}

function renderStatus(s: string): string {
  if (s === '1') return '启用';
  if (s === '0') return '停用';
  return s || '-';
}

function renderEmpty(v: string | number | null | undefined): string {
  return v != null && v !== '' ? String(v) : '-';
}

const columns: TableColumnType[] = [
  { title: '用户ID', dataIndex: 'userId', key: 'userId', width: 80 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 110 },
  { title: '昵称', dataIndex: 'nickName', key: 'nickName', width: 110 },
  {
    title: '部门',
    key: 'dept',
    width: 120,
    customRender: ({ record }: { record: SysUserItem }) =>
      record.dept?.deptName ?? record.deptId ?? '-',
  },
  { title: '岗位ID', dataIndex: 'postId', key: 'postId', width: 80 },
  { title: '角色ID', dataIndex: 'roleId', key: 'roleId', width: 80 },
  { title: '手机号', dataIndex: 'phone', key: 'phone', width: 120 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: string }) => renderStatus(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 160,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

// --------- 新增 ---------
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  username: '',
  password: '',
  nickName: '',
  phone: '',
  email: '',
  deptId: undefined as number | undefined,
  roleId: undefined as number | undefined,
  postId: undefined as number | undefined,
  sex: '',
  remark: '',
  status: '1',
});

function resetAddForm() {
  addForm.username = '';
  addForm.password = '';
  addForm.nickName = '';
  addForm.phone = '';
  addForm.email = '';
  addForm.deptId = undefined;
  addForm.roleId = undefined;
  addForm.postId = undefined;
  addForm.sex = '';
  addForm.remark = '';
  addForm.status = '1';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAdd(): { ok: boolean; message?: string } {
  if (!addForm.username?.trim()) return { ok: false, message: '请输入用户名' };
  if (!addForm.password?.trim()) return { ok: false, message: '请输入密码' };
  if (!addForm.nickName?.trim()) return { ok: false, message: '请输入昵称' };
  if (!addForm.phone?.trim()) return { ok: false, message: '请输入手机号' };
  if (!addForm.email?.trim()) return { ok: false, message: '请输入邮箱' };
  if (addForm.deptId == null || addForm.deptId === 0)
    return { ok: false, message: '请选择部门' };
  return { ok: true };
}

async function onAddOk() {
  const v = validateAdd();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createSysUser({
      username: addForm.username.trim(),
      password: addForm.password.trim(),
      nickName: addForm.nickName.trim(),
      phone: addForm.phone.trim(),
      email: addForm.email.trim(),
      deptId: addForm.deptId!,
      roleId: addForm.roleId,
      postId: addForm.postId,
      sex: addForm.sex || undefined,
      remark: addForm.remark?.trim() || undefined,
      status: addForm.status,
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchUserList();
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

// --------- 编辑 ---------
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editUserId = ref<number | null>(null);
const editForm = reactive({
  username: '',
  nickName: '',
  phone: '',
  email: '',
  deptId: undefined as number | undefined,
  roleId: undefined as number | undefined,
  postId: undefined as number | undefined,
  sex: '',
  remark: '',
  status: '1',
});

async function openEditModal(record: SysUserItem) {
  editUserId.value = record.userId;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getSysUserDetail(record.userId);
    editForm.username = detail.username ?? '';
    editForm.nickName = detail.nickName ?? '';
    editForm.phone = detail.phone ?? '';
    editForm.email = detail.email ?? '';
    editForm.deptId = detail.deptId ?? undefined;
    editForm.roleId = detail.roleId ?? undefined;
    editForm.postId = detail.postId ?? undefined;
    editForm.sex = detail.sex ?? '';
    editForm.remark = detail.remark ?? '';
    editForm.status = detail.status ?? '1';
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取用户详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEdit(): { ok: boolean; message?: string } {
  if (!editForm.username?.trim()) return { ok: false, message: '请输入用户名' };
  if (!editForm.nickName?.trim()) return { ok: false, message: '请输入昵称' };
  if (!editForm.phone?.trim()) return { ok: false, message: '请输入手机号' };
  if (!editForm.email?.trim()) return { ok: false, message: '请输入邮箱' };
  if (editForm.deptId == null || editForm.deptId === 0)
    return { ok: false, message: '请选择部门' };
  return { ok: true };
}

async function onEditOk() {
  if (editUserId.value == null) return;
  const v = validateEdit();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateSysUser({
      userId: editUserId.value,
      username: editForm.username.trim(),
      nickName: editForm.nickName.trim(),
      phone: editForm.phone.trim(),
      email: editForm.email.trim(),
      deptId: editForm.deptId!,
      roleId: editForm.roleId,
      postId: editForm.postId,
      sex: editForm.sex || undefined,
      remark: editForm.remark?.trim() || undefined,
      status: editForm.status,
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchUserList();
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

// --------- 删除 ---------
function onDelete(record: SysUserItem) {
  const name = record.nickName || record.username || `用户ID:${record.userId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSysUser([record.userId]);
        message.success('删除成功');
        fetchUserList();
      } catch (e: unknown) {
        const err = e as { message?: string; response?: { data?: { msg?: string } } };
        message.error(err?.message || err?.response?.data?.msg || '删除失败');
      }
    },
  });
}

const statusEditOptions = [
  { value: '0', label: '停用' },
  { value: '1', label: '启用' },
];

onMounted(() => {
  loadDeptTree();
  loadRoleOptions();
  loadPostOptions();
  fetchUserList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">用户管理</h2>
      <div class="flex gap-2">
        <Button @click="fetchUserList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增用户</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">用户名：</span>
      <Input
        v-model:value="searchUsername"
        placeholder="用户名"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="text-sm text-gray-600">昵称：</span>
      <Input
        v-model:value="searchNickName"
        placeholder="昵称"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="text-sm text-gray-600">手机号：</span>
      <Input
        v-model:value="searchPhone"
        placeholder="手机号"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
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
      :row-key="(r: SysUserItem) => r.userId"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="openEditModal(record as SysUserItem)">
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysUserItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增用户"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="用户名" required>
          <Input v-model:value="addForm.username" placeholder="用户名" allow-clear />
        </FormItem>
        <FormItem label="密码" required>
          <Input
            v-model:value="addForm.password"
            type="password"
            placeholder="密码"
            allow-clear
          />
        </FormItem>
        <FormItem label="昵称" required>
          <Input v-model:value="addForm.nickName" placeholder="昵称" allow-clear />
        </FormItem>
        <FormItem label="手机号" required>
          <Input v-model:value="addForm.phone" placeholder="手机号" allow-clear />
        </FormItem>
        <FormItem label="邮箱" required>
          <Input v-model:value="addForm.email" placeholder="邮箱" allow-clear />
        </FormItem>
        <FormItem label="部门" required>
          <TreeSelect
            v-model:value="addForm.deptId"
            :tree-data="deptTreeOptions"
            placeholder="请选择部门"
            tree-default-expand-all
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="角色">
          <Select
            v-model:value="addForm.roleId"
            :options="roleOptions"
            placeholder="请选择角色"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="岗位">
          <Select
            v-model:value="addForm.postId"
            :options="postOptions"
            placeholder="请选择岗位"
            allow-clear
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
            placeholder="备注"
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
      title="编辑用户"
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
        <FormItem label="用户名" required>
          <Input v-model:value="editForm.username" placeholder="用户名" allow-clear />
        </FormItem>
        <FormItem label="昵称" required>
          <Input v-model:value="editForm.nickName" placeholder="昵称" allow-clear />
        </FormItem>
        <FormItem label="手机号" required>
          <Input v-model:value="editForm.phone" placeholder="手机号" allow-clear />
        </FormItem>
        <FormItem label="邮箱" required>
          <Input v-model:value="editForm.email" placeholder="邮箱" allow-clear />
        </FormItem>
        <FormItem label="部门" required>
          <TreeSelect
            v-model:value="editForm.deptId"
            :tree-data="deptTreeOptions"
            placeholder="请选择部门"
            tree-default-expand-all
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="角色">
          <Select
            v-model:value="editForm.roleId"
            :options="roleOptions"
            placeholder="请选择角色"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="岗位">
          <Select
            v-model:value="editForm.postId"
            :options="postOptions"
            placeholder="请选择岗位"
            allow-clear
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
            placeholder="备注"
            allow-clear
            type="textarea"
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
