<script lang="ts" setup>
/**
 * 系统管理 - 用户管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 重置密码 + 删除 + 部门/角色/岗位关联
 */
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { Button, Input, Modal, Select, Table, Tag, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createSysUser,
  deleteSysUser,
  getDeptTreeApi,
  getPostPage,
  getRolePage,
  getSysUserDetail,
  getSysUserPage,
  resetSysUserPassword,
  updateSysUser,
} from '#/api/core';
import type { DeptLabel, DeptTreeOption, SysUserItem } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import UserAvatar from '#/components/user-avatar.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

const statusOptions = [
  { value: '', label: '全部' },
  { value: '2', label: '启用' },
  { value: '1', label: '停用' },
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
  SysUserItem,
  {
    nickName: string;
    phone: string;
    roleIds: number[];
    status: string;
    username: string;
  },
  {
    nickName?: string;
    phone?: string;
    roleIds?: string;
    status?: string;
    username?: string;
  }
>({
  createParams: (currentQuery) => ({
    username: currentQuery.username.trim() || undefined,
    nickName: currentQuery.nickName.trim() || undefined,
    phone: currentQuery.phone.trim() || undefined,
    roleIds: currentQuery.roleIds.length
      ? currentQuery.roleIds.join(',')
      : undefined,
    status: currentQuery.status || undefined,
  }),
  createQuery: () => ({
    username: '',
    nickName: '',
    phone: '',
    roleIds: [] as number[],
    status: '',
  }),
  fallbackErrorMessage: '加载用户列表失败',
  fetcher: async (params) => getSysUserPage(params),
});

const fetchUserList = fetchList;

/** 部门树（用于下拉） */
const deptTreeRaw = ref<DeptLabel[]>([]);

function buildDeptTreeOptions(nodes: DeptLabel[]): DeptTreeOption[] {
  return (nodes || []).map(
    (n): DeptTreeOption => ({
      value: n.id,
      label: n.label,
      children: n.children?.length
        ? buildDeptTreeOptions(n.children)
        : undefined,
    }),
  );
}

const deptTreeOptions = computed(() => buildDeptTreeOptions(deptTreeRaw.value));

/** 角色选项 */
const roleOptions = ref<{ value: number; label: string }[]>([]);
/** 岗位选项 */
const postOptions = ref<{ value: number; label: string }[]>([]);

const roleLabelMap = computed(() =>
  new Map(roleOptions.value.map((item) => [item.value, item.label])),
);

const roleFilterOptions = computed(() =>
  [...roleOptions.value].sort((a, b) => a.label.localeCompare(b.label, 'zh-CN')),
);

function getRoleLabels(record: Partial<SysUserItem> | null | undefined): string[] {
  if (!record) {
    return [];
  }
  if (record.roles?.length) {
    return record.roles
      .map((role) => role.roleName || roleLabelMap.value.get(role.roleId) || `角色${role.roleId}`)
      .filter(Boolean);
  }
  if (record.roleIds?.length) {
    return record.roleIds
      .map((roleId) => roleLabelMap.value.get(roleId) || `角色${roleId}`)
      .filter(Boolean);
  }
  if (record.roleId) {
    return [roleLabelMap.value.get(record.roleId) || `角色${record.roleId}`];
  }
  return [];
}

function getPrimaryRoleLabel(record: Partial<SysUserItem> | null | undefined): string {
  if (!record) {
    return '-';
  }
  const primaryRoleId = record.primaryRoleId ?? record.roleId;
  if (!primaryRoleId) {
    return '-';
  }
  const fromRoleList = record.roles?.find((role) => role.roleId === primaryRoleId)?.roleName;
  return fromRoleList || roleLabelMap.value.get(primaryRoleId) || `角色${primaryRoleId}`;
}

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
    roleOptions.value = (res.list || [])
      .filter((r) => r.status === '2')
      .map((r) => ({
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

function renderStatus(s: string): string {
  if (s === '2') return '启用';
  if (s === '1' || s === '0') return '停用';
  return s || '-';
}

const baseColumns: TableColumnType[] = [
  { title: '用户ID', dataIndex: 'userId', key: 'userId', width: 80 },
  { title: '头像', key: 'avatar', width: 88 },
  { title: '登录账号', dataIndex: 'username', key: 'username', width: 110 },
  { title: '姓名', dataIndex: 'nickName', key: 'nickName', width: 110 },
  {
    title: '部门',
    key: 'dept',
    width: 120,
    customRender: ({ record }: { record: SysUserItem }) =>
      renderAdminEmpty(record.dept?.deptName ?? record.deptId),
  },
  { title: '岗位ID', dataIndex: 'postId', key: 'postId', width: 80 },
  { title: '主角色', key: 'primaryRole', width: 140 },
  { title: '角色', key: 'roles', width: 220 },
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
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  { title: '操作', key: 'action', width: 220, fixed: 'right' },
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
  tableId: 'sys-user-list',
});

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
  roleIds: [] as number[],
  primaryRoleId: undefined as number | undefined,
  postId: undefined as number | undefined,
  sex: '',
  remark: '',
  status: '2',
});

function resetAddForm() {
  addForm.username = '';
  addForm.password = '';
  addForm.nickName = '';
  addForm.phone = '';
  addForm.email = '';
  addForm.deptId = undefined;
  addForm.roleIds = [];
  addForm.primaryRoleId = undefined;
  addForm.postId = undefined;
  addForm.sex = '';
  addForm.remark = '';
  addForm.status = '1';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

// 暴露给自动化测试使用
(window as any).openAddModalForTest = openAddModal;

function validateAdd(): { ok: boolean; message?: string } {
  if (!addForm.username?.trim()) return { ok: false, message: '请输入登录账号' };
  if (!addForm.password?.trim()) return { ok: false, message: '请输入密码' };
  if (!addForm.nickName?.trim()) return { ok: false, message: '请输入姓名' };
  if (!addForm.phone?.trim()) return { ok: false, message: '请输入手机号' };
  if (!addForm.email?.trim()) return { ok: false, message: '请输入邮箱' };
  if (addForm.deptId == null || addForm.deptId === 0)
    return { ok: false, message: '请选择部门' };
  if (!addForm.roleIds.length) return { ok: false, message: '请至少选择一个角色' };
  if (addForm.primaryRoleId == null || !addForm.roleIds.includes(addForm.primaryRoleId)) {
    return { ok: false, message: '请选择主角色' };
  }
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
      primaryRoleId: addForm.primaryRoleId!,
      roleIds: addForm.roleIds,
      roleId: addForm.primaryRoleId,
      postId: addForm.postId,
      sex: addForm.sex || undefined,
      remark: addForm.remark?.trim() || undefined,
      status: addForm.status,
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchUserList();
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
  roleIds: [] as number[],
  primaryRoleId: undefined as number | undefined,
  postId: undefined as number | undefined,
  sex: '',
  remark: '',
  status: '2',
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
    editForm.roleIds =
      detail.roleIds?.length ? [...detail.roleIds] : detail.roleId ? [detail.roleId] : [];
    editForm.primaryRoleId = detail.primaryRoleId ?? detail.roleId ?? undefined;
    editForm.postId = detail.postId ?? undefined;
    editForm.sex = detail.sex ?? '';
    editForm.remark = detail.remark ?? '';
    editForm.status = detail.status ?? '2';
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取用户详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEdit(): { ok: boolean; message?: string } {
  if (!editForm.username?.trim()) return { ok: false, message: '请输入登录账号' };
  if (!editForm.nickName?.trim()) return { ok: false, message: '请输入姓名' };
  if (!editForm.phone?.trim()) return { ok: false, message: '请输入手机号' };
  if (!editForm.email?.trim()) return { ok: false, message: '请输入邮箱' };
  if (editForm.deptId == null || editForm.deptId === 0)
    return { ok: false, message: '请选择部门' };
  if (!editForm.roleIds.length) return { ok: false, message: '请至少选择一个角色' };
  if (editForm.primaryRoleId == null || !editForm.roleIds.includes(editForm.primaryRoleId)) {
    return { ok: false, message: '请选择主角色' };
  }
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
      primaryRoleId: editForm.primaryRoleId!,
      roleIds: editForm.roleIds,
      roleId: editForm.primaryRoleId,
      postId: editForm.postId,
      sex: editForm.sex || undefined,
      remark: editForm.remark?.trim() || undefined,
      status: editForm.status,
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchUserList();
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

// --------- 重置密码 ---------
const resetPwdVisible = ref(false);
const resetPwdSubmitting = ref(false);
const resetPwdUserId = ref<number | null>(null);
const resetPwdTargetLabel = ref('');
const resetPwdForm = reactive({
  confirmPassword: '',
  password: '',
});

function resetResetPwdForm() {
  resetPwdForm.password = '';
  resetPwdForm.confirmPassword = '';
}

function openResetPwdModal(record: SysUserItem) {
  resetResetPwdForm();
  resetPwdUserId.value = record.userId;
  resetPwdTargetLabel.value =
    record.nickName || record.username || `用户ID:${record.userId}`;
  resetPwdVisible.value = true;
}

function validateResetPwd(): { ok: boolean; message?: string } {
  if (!resetPwdForm.password) {
    return { ok: false, message: '请输入新密码' };
  }
  if (!resetPwdForm.confirmPassword) {
    return { ok: false, message: '请再次输入新密码' };
  }
  if (resetPwdForm.password !== resetPwdForm.confirmPassword) {
    return { ok: false, message: '两次输入的密码不一致' };
  }
  return { ok: true };
}

async function onResetPwdOk() {
  if (resetPwdUserId.value == null) {
    return;
  }
  const validation = validateResetPwd();
  if (!validation.ok) {
    message.error(validation.message);
    return;
  }
  resetPwdSubmitting.value = true;
  try {
    await resetSysUserPassword({
      password: resetPwdForm.password,
      userId: resetPwdUserId.value,
    });
    message.success('密码已重置');
    resetPwdVisible.value = false;
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '重置密码失败');
  } finally {
    resetPwdSubmitting.value = false;
  }
}

function onResetPwdCancel() {
  resetPwdVisible.value = false;
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
  { value: '2', label: '启用' },
  { value: '1', label: '停用' },
];

const addPrimaryRoleOptions = computed(() =>
  roleOptions.value.filter((option) => addForm.roleIds.includes(option.value)),
);

const editPrimaryRoleOptions = computed(() =>
  roleOptions.value.filter((option) => editForm.roleIds.includes(option.value)),
);

watch(
  () => addForm.roleIds,
  (value) => {
    if (!value.length) {
      addForm.primaryRoleId = undefined;
      return;
    }
    if (
      addForm.primaryRoleId == null ||
      !value.includes(addForm.primaryRoleId)
    ) {
      addForm.primaryRoleId = value[0];
    }
  },
  { deep: true },
);

watch(
  () => editForm.roleIds,
  (value) => {
    if (!value.length) {
      editForm.primaryRoleId = undefined;
      return;
    }
    if (
      editForm.primaryRoleId == null ||
      !value.includes(editForm.primaryRoleId)
    ) {
      editForm.primaryRoleId = value[0];
    }
  },
  { deep: true },
);

const userAddFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'username',
    label: '登录账号',
    placeholder: '登录账号',
    props: { 'data-testid': 'input-username' },
    required: true,
  },
  {
    component: 'input',
    field: 'password',
    label: '密码',
    placeholder: '密码',
    props: { type: 'password', 'data-testid': 'input-password' },
    required: true,
  },
  {
    component: 'input',
    field: 'nickName',
    label: '姓名',
    placeholder: '姓名',
    props: { 'data-testid': 'input-nickname' },
    required: true,
  },
  {
    component: 'input',
    field: 'phone',
    label: '手机号',
    placeholder: '手机号',
    props: { 'data-testid': 'input-phone' },
    required: true,
  },
  {
    component: 'input',
    field: 'email',
    label: '邮箱',
    placeholder: '邮箱',
    props: { 'data-testid': 'input-email' },
    required: true,
  },
  {
    allowClear: true,
    component: 'tree-select',
    field: 'deptId',
    label: '部门',
    placeholder: '请选择部门',
    props: {
      treeData: deptTreeOptions.value,
      treeDefaultExpandAll: true,
    },
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'roleIds',
    label: '角色列表',
    options: roleOptions.value,
    placeholder: '请选择角色',
    props: {
      mode: 'multiple',
    },
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'primaryRoleId',
    label: '主角色',
    options: addPrimaryRoleOptions.value,
    placeholder: '请选择主角色',
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'postId',
    label: '岗位',
    options: postOptions.value,
    placeholder: '请选择岗位',
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
    placeholder: '备注',
    span: 2,
  },
]);

const userEditFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'username',
    label: '登录账号',
    placeholder: '登录账号',
    required: true,
  },
  {
    component: 'input',
    field: 'nickName',
    label: '姓名',
    placeholder: '姓名',
    required: true,
  },
  {
    component: 'input',
    field: 'phone',
    label: '手机号',
    placeholder: '手机号',
    required: true,
  },
  {
    component: 'input',
    field: 'email',
    label: '邮箱',
    placeholder: '邮箱',
    required: true,
  },
  {
    allowClear: true,
    component: 'tree-select',
    field: 'deptId',
    label: '部门',
    placeholder: '请选择部门',
    props: {
      treeData: deptTreeOptions.value,
      treeDefaultExpandAll: true,
    },
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'roleIds',
    label: '角色列表',
    options: roleOptions.value,
    placeholder: '请选择角色',
    props: {
      mode: 'multiple',
    },
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'primaryRoleId',
    label: '主角色',
    options: editPrimaryRoleOptions.value,
    placeholder: '请选择主角色',
    required: true,
  },
  {
    allowClear: true,
    component: 'select',
    field: 'postId',
    label: '岗位',
    options: postOptions.value,
    placeholder: '请选择岗位',
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
    placeholder: '备注',
    span: 2,
  },
]);

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysUserItem | null>(null);

async function openDetail(record: SysUserItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detailRecord.value = await getSysUserDetail(record.userId);
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取用户详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

onMounted(() => {
  loadDeptTree();
  loadRoleOptions();
  loadPostOptions();
  void fetchUserList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>用户管理</template>
    <template #description>
      管理用户、部门、角色和岗位绑定关系。复杂列表页也统一采用后台筛选网格和卡片式表格结构。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        data-testid="btn-add-user"
        codes="admin:sysUser:add"
        @click="openAddModal"
      >
        新增用户
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="登录账号">
          <Input
            v-model:value="query.username"
            placeholder="登录账号"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="姓名">
          <Input
            v-model:value="query.nickName"
            placeholder="姓名"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="手机号">
          <Input
            v-model:value="query.phone"
            placeholder="手机号"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            class="w-full"
            :options="statusOptions"
            placeholder="请选择状态"
          />
        </AdminFilterField>
        <AdminFilterField label="角色">
          <Select
            v-model:value="query.roleIds"
            class="w-full"
            :max-tag-count="2"
            :options="roleFilterOptions"
            mode="multiple"
            option-filter-prop="label"
            placeholder="请选择角色（命中任一）"
            show-search
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
        <div class="text-base font-semibold text-slate-900">用户列表</div>
      </div>
    </template>
    <template #toolbar-extra>
      <AdminTableColumnSettings
        v-model:open="settingsOpen"
        :columns="settingsColumns"
        @change-fixed="({ key, fixed }) => setColumnFixed(key, fixed)"
        @reorder="({ oldIndex, newIndex }) => reorderColumns(oldIndex, newIndex)"
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
      :row-key="(r: SysUserItem) => r.userId"
      :scroll="{ x: scrollX }"
      size="middle"
      @change="onTableChange"
      @resizeColumn="handleResizeColumn"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'avatar'">
          <UserAvatar
            :avatar="(record as SysUserItem).avatar"
            :avatar-color="(record as SysUserItem).avatarColor"
            :avatar-type="(record as SysUserItem).avatarType"
            :real-name="(record as SysUserItem).nickName"
            :username="(record as SysUserItem).username"
            class="size-9"
          />
        </template>
        <template v-else-if="column.key === 'primaryRole'">
          <span>{{ getPrimaryRoleLabel(record as SysUserItem) }}</span>
        </template>
        <template v-else-if="column.key === 'roles'">
          <div class="flex flex-wrap gap-1">
            <Tag
              v-for="roleLabel in getRoleLabels(record as SysUserItem)"
              :key="roleLabel"
              color="blue"
            >
              {{ roleLabel }}
            </Tag>
            <span v-if="!getRoleLabels(record as SysUserItem).length">
              {{ renderAdminEmpty('-') }}
            </span>
          </div>
        </template>
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysUser:query"
            @click="openDetail(record as SysUserItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysUser:edit"
            @click="openEditModal(record as SysUserItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysUser:resetPwd"
            @click="openResetPwdModal(record as SysUserItem)"
          >
            重置密码
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysUser:remove"
            @click="onDelete(record as SysUserItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增用户"
      :confirm-loading="addSubmitting"
      :width="860"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <AdminModalFormFields :model="addForm" :fields="userAddFormFields" />
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑用户"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      :width="860"
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
        :fields="userEditFormFields"
      />
    </Modal>

    <Modal
      v-model:open="resetPwdVisible"
      title="重置密码"
      :confirm-loading="resetPwdSubmitting"
      ok-text="确认重置"
      cancel-text="取消"
      @ok="onResetPwdOk"
      @cancel="onResetPwdCancel"
    >
      <div class="space-y-4">
        <div class="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
          正在为用户
          <span class="font-medium text-slate-900">{{ resetPwdTargetLabel }}</span>
          设置新密码。此操作不需要输入旧密码，保存后新密码立即生效。
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700">
            新密码
          </label>
          <Input.Password
            v-model:value="resetPwdForm.password"
            autocomplete="new-password"
            placeholder="请输入新密码"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700">
            确认新密码
          </label>
          <Input.Password
            v-model:value="resetPwdForm.confirmPassword"
            autocomplete="new-password"
            placeholder="请再次输入新密码"
          />
        </div>
      </div>
    </Modal>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="用户详情"
      :loading="detailLoading"
      width="720"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="基础信息" description="账号、姓名和当前启用状态。">
          <div class="mb-4 flex items-center gap-4 rounded-2xl bg-slate-50 p-4">
            <UserAvatar
              :avatar="detailRecord.avatar"
              :avatar-color="detailRecord.avatarColor"
              :avatar-type="detailRecord.avatarType"
              :real-name="detailRecord.nickName"
              :username="detailRecord.username"
              class="size-14"
            />
            <div>
              <div class="text-sm font-medium text-slate-900">
                {{ renderAdminEmpty(detailRecord.nickName || detailRecord.username) }}
              </div>
              <div class="mt-1 text-xs text-slate-500">
                头像模式：{{ detailRecord.avatarType === 'image' ? '图片头像' : '字母头像' }}
              </div>
              <div class="mt-1 text-xs text-slate-500">
                背景色：{{ renderAdminEmpty(detailRecord.avatarColor || '自动映射') }}
              </div>
            </div>
          </div>
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">登录账号</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.username) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">姓名</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.nickName) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">状态</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderStatus(detailRecord.status) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">创建时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ formatAdminDateTime(detailRecord.createdAt) }}</dd>
            </div>
          </dl>
        </AdminDetailSection>

        <AdminDetailSection title="联系与组织" description="查看联系方式、部门和岗位绑定。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">手机号</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.phone) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">邮箱</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.email) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">部门</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.dept?.deptName ?? detailRecord.deptId) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">岗位 ID</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.postId) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">主角色</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(getPrimaryRoleLabel(detailRecord)) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">性别</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.sex) }}</dd>
            </div>
          </dl>
          <div class="mt-4">
            <dt class="text-xs text-slate-500">全部角色</dt>
            <dd class="mt-2 flex flex-wrap gap-2">
              <Tag
                v-for="roleLabel in getRoleLabels(detailRecord)"
                :key="roleLabel"
                color="blue"
              >
                {{ roleLabel }}
              </Tag>
              <span v-if="!getRoleLabels(detailRecord).length" class="text-sm text-slate-900">
                {{ renderAdminEmpty('-') }}
              </span>
            </dd>
          </div>
        </AdminDetailSection>

        <AdminDetailSection title="备注" description="保留业务补充说明。">
          <p class="text-sm leading-6 text-slate-700">
            {{ renderAdminEmpty(detailRecord.remark) }}
          </p>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
