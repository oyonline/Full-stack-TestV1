<script lang="ts" setup>
/**
 * 系统管理 - 角色管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 菜单树/部门树最小接入
 * 复用岗位管理母版，树选择参考 roleMenuTreeselect / roleDeptTreeselect
 */
import { computed, onMounted, reactive, ref } from 'vue';

import { Button, Input, Modal, Select, Table, Tree, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';
import type { TreeProps } from 'ant-design-vue';

import {
  createRole,
  deleteRole,
  getRoleDeptTreeselect,
  getRoleDetail,
  getRoleMenuTreeselect,
  getRolePage,
  updateRole,
  updateRoleStatus,
} from '#/api/core/role';
import type {
  CreateRoleData,
  MenuLabel,
  SysRoleItem,
  UpdateRoleData,
} from '#/api/core/role';
import type { DeptLabel } from '#/api/core/dept';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
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

type TreeNode = NonNullable<TreeProps['treeData']>[number];
type TreeKey = string | number;

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
  SysRoleItem,
  {
    roleKey: string;
    roleName: string;
    status: string;
  },
  {
    roleKey?: string;
    roleName?: string;
    status?: string;
  }
>({
  createParams: (currentQuery) => ({
    roleName: currentQuery.roleName.trim() || undefined,
    roleKey: currentQuery.roleKey.trim() || undefined,
    status: currentQuery.status || undefined,
  }),
  createQuery: () => ({
    roleName: '',
    roleKey: '',
    status: '',
  }),
  fallbackErrorMessage: '加载角色列表失败',
  fetcher: async (params) => getRolePage(params),
});

const fetchRoleList = fetchList;

function renderStatus(s: string | undefined): string {
  if (s === '2') return '启用';
  if (s === '1') return '停用';
  return s || '-';
}

const columns: TableColumnType[] = [
  { title: '角色ID', dataIndex: 'roleId', key: 'roleId', width: 90 },
  { title: '角色名称', dataIndex: 'roleName', key: 'roleName', width: 140 },
  { title: '权限字符', dataIndex: 'roleKey', key: 'roleKey', width: 120 },
  { title: '排序', dataIndex: 'roleSort', key: 'roleSort', width: 80 },
  {
    title: '数据范围',
    dataIndex: 'dataScope',
    key: 'dataScope',
    width: 100,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
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
  { title: '操作', key: 'action', width: 200, fixed: 'right' },
];

/** 菜单树 -> Ant Design Tree 数据 */
function menuLabelToTreeData(nodes: MenuLabel[]): TreeNode[] {
  return (nodes || []).map((n) => ({
    key: n.id,
    title: n.label || '',
    children: n.children?.length ? menuLabelToTreeData(n.children) : undefined,
  }));
}

/** 部门树 -> Ant Design Tree 数据 */
function deptLabelToTreeData(nodes: DeptLabel[]): TreeNode[] {
  return (nodes || []).map((n) => ({
    key: n.id,
    title: n.label || '',
    children: n.children?.length ? deptLabelToTreeData(n.children) : undefined,
  }));
}

function normalizeTreeKeys(keys: TreeKey[]): number[] {
  return keys.map((key) => Number(key)).filter((key) => Number.isFinite(key));
}

/* -------- 新增 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  roleName: '',
  roleKey: '',
  roleSort: 0,
  status: '2',
  remark: '',
  dataScope: '',
});
const addMenuTreeData = ref<TreeNode[]>([]);
const addMenuCheckedKeys = ref<TreeKey[]>([]);
const addMenuHalfCheckedKeys = ref<TreeKey[]>([]);
const addDeptTreeData = ref<TreeNode[]>([]);
const addDeptCheckedKeys = ref<TreeKey[]>([]);
const addDeptHalfCheckedKeys = ref<TreeKey[]>([]);

function resetAddForm() {
  addForm.roleName = '';
  addForm.roleKey = '';
  addForm.roleSort = 0;
  addForm.status = '2';
  addForm.remark = '';
  addForm.dataScope = '';
  addMenuCheckedKeys.value = [];
  addMenuHalfCheckedKeys.value = [];
  addDeptCheckedKeys.value = [];
  addDeptHalfCheckedKeys.value = [];
}

async function openAddModal() {
  resetAddForm();
  addVisible.value = true;
  try {
    const [menuRes, deptRes] = await Promise.all([
      getRoleMenuTreeselect(0),
      getRoleDeptTreeselect(0),
    ]);
    addMenuTreeData.value = menuLabelToTreeData(menuRes?.menus || []);
    addDeptTreeData.value = deptLabelToTreeData(deptRes?.depts || []);
  } catch {
    addMenuTreeData.value = [];
    addDeptTreeData.value = [];
  }
}

function validateAddForm(): { ok: boolean; message?: string } {
  const name = addForm.roleName?.trim() ?? '';
  const key = addForm.roleKey?.trim() ?? '';
  if (!name) return { ok: false, message: '请输入角色名称' };
  if (!key) return { ok: false, message: '请输入权限字符' };
  return { ok: true };
}

const onAddMenuCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  const c = checkedKeys as TreeKey[];
  addMenuCheckedKeys.value = c;
  addMenuHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

const onAddDeptCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  const c = checkedKeys as TreeKey[];
  addDeptCheckedKeys.value = c;
  addDeptHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

async function onAddOk() {
  const v = validateAddForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    const menuIds = normalizeTreeKeys([
      ...new Set([
        ...addMenuCheckedKeys.value,
        ...addMenuHalfCheckedKeys.value,
      ]),
    ]);
    const deptIds = normalizeTreeKeys([
      ...new Set([
        ...addDeptCheckedKeys.value,
        ...addDeptHalfCheckedKeys.value,
      ]),
    ]);
    const data: CreateRoleData = {
      roleName: addForm.roleName.trim(),
      roleKey: addForm.roleKey.trim(),
      roleSort: addForm.roleSort,
      status: addForm.status,
      remark: addForm.remark?.trim() || undefined,
      dataScope: addForm.dataScope?.trim() || undefined,
      menuIds: menuIds.length ? menuIds : undefined,
      deptIds: deptIds.length ? deptIds : undefined,
    };
    await createRole(data);
    message.success('新增成功');
    addVisible.value = false;
    fetchRoleList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
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
const editRoleId = ref<number | null>(null);
const editForm = reactive({
  roleName: '',
  roleKey: '',
  roleSort: 0,
  status: '2',
  remark: '',
  dataScope: '',
});
const editMenuTreeData = ref<TreeNode[]>([]);
const editMenuCheckedKeys = ref<TreeKey[]>([]);
const editMenuHalfCheckedKeys = ref<TreeKey[]>([]);
const editDeptTreeData = ref<TreeNode[]>([]);
const editDeptCheckedKeys = ref<TreeKey[]>([]);
const editDeptHalfCheckedKeys = ref<TreeKey[]>([]);

const onEditMenuCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  const c = checkedKeys as TreeKey[];
  editMenuCheckedKeys.value = c;
  editMenuHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

const onEditDeptCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  const c = checkedKeys as TreeKey[];
  editDeptCheckedKeys.value = c;
  editDeptHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

async function openEditModal(record: SysRoleItem) {
  editRoleId.value = record.roleId;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const [detail, menuRes, deptRes] = await Promise.all([
      getRoleDetail(record.roleId),
      getRoleMenuTreeselect(record.roleId),
      getRoleDeptTreeselect(record.roleId),
    ]);
    editForm.roleName = detail.roleName ?? '';
    editForm.roleKey = detail.roleKey ?? '';
    editForm.roleSort = detail.roleSort ?? 0;
    editForm.status = detail.status ?? '2';
    editForm.remark = detail.remark ?? '';
    editForm.dataScope = detail.dataScope ?? '';
    editMenuTreeData.value = menuLabelToTreeData(menuRes?.menus || []);
    editMenuCheckedKeys.value = menuRes?.checkedKeys || [];
    editMenuHalfCheckedKeys.value = [];
    editDeptTreeData.value = deptLabelToTreeData(deptRes?.depts || []);
    editDeptCheckedKeys.value = deptRes?.checkedKeys || [];
    editDeptHalfCheckedKeys.value = [];
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取角色详情失败'));
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
  const name = editForm.roleName?.trim() ?? '';
  const key = editForm.roleKey?.trim() ?? '';
  if (!name) return { ok: false, message: '请输入角色名称' };
  if (!key) return { ok: false, message: '请输入权限字符' };
  return { ok: true };
}

async function onEditOk() {
  if (editRoleId.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    const menuIds = normalizeTreeKeys([
      ...new Set([
        ...editMenuCheckedKeys.value,
        ...editMenuHalfCheckedKeys.value,
      ]),
    ]);
    const deptIds = normalizeTreeKeys([
      ...new Set([
        ...editDeptCheckedKeys.value,
        ...editDeptHalfCheckedKeys.value,
      ]),
    ]);
    const data: UpdateRoleData = {
      roleName: editForm.roleName.trim(),
      roleKey: editForm.roleKey.trim(),
      roleSort: editForm.roleSort,
      status: editForm.status,
      remark: editForm.remark?.trim() || undefined,
      dataScope: editForm.dataScope?.trim() || undefined,
      menuIds: menuIds.length ? menuIds : undefined,
      deptIds: deptIds.length ? deptIds : undefined,
    };
    await updateRole(editRoleId.value, data);
    message.success('编辑成功');
    editVisible.value = false;
    fetchRoleList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '编辑失败'));
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除 -------- */
function onDelete(record: SysRoleItem) {
  const name = record.roleName || `角色ID:${record.roleId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除角色「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteRole([record.roleId]);
        message.success('删除成功');
        fetchRoleList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

/* -------- 状态切换 -------- */
async function onToggleStatus(record: SysRoleItem) {
  const next = record.status === '2' ? '1' : '2';
  const label = next === '2' ? '启用' : '停用';
  try {
    await updateRoleStatus(record.roleId, next);
    message.success(`${label}成功`);
    fetchRoleList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, `${label}失败`));
  }
}

const statusEditOptions = [
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
];

const roleFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'roleName',
    label: '角色名称',
    placeholder: '请输入角色名称',
    required: true,
  },
  {
    component: 'input',
    field: 'roleKey',
    label: '权限字符',
    placeholder: '请输入权限字符',
    required: true,
  },
  {
    component: 'input-number',
    field: 'roleSort',
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
    component: 'input',
    field: 'dataScope',
    label: '数据范围',
    placeholder: '选填',
    span: 2,
  },
  {
    component: 'textarea',
    field: 'remark',
    label: '备注',
    placeholder: '请输入备注',
    span: 2,
  },
]);

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysRoleItem | null>(null);

async function openDetail(record: SysRoleItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detailRecord.value = await getRoleDetail(record.roleId);
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取角色详情失败'));
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

onMounted(() => {
  void fetchRoleList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>角色管理</template>
    <template #description>
      管理角色名称、权限字符与数据范围，树权限表单与列表页统一沿用后台样式骨架。
    </template>
    <template #header-extra>
      <Button @click="fetchRoleList">刷新</Button>
      <AdminActionButton type="primary" codes="admin:sysRole:add" @click="openAddModal">
        新增角色
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="角色名称">
          <Input
            v-model:value="query.roleName"
            placeholder="请输入角色名称"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="权限字符">
          <Input
            v-model:value="query.roleKey"
            placeholder="请输入权限字符"
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
        <div class="text-base font-semibold text-slate-900">角色列表</div>
        <p class="mt-1 text-sm text-slate-500">
          支持按角色名称、权限字符和状态筛选，并维护菜单树与部门树权限。
        </p>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysRoleItem) => record.roleId"
      :scroll="{ x: 1080 }"
      size="middle"
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysRole:query"
            @click="openDetail(record as SysRoleItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysRole:update"
            @click="openEditModal(record as SysRoleItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysRole:update"
            @click="onToggleStatus(record as SysRoleItem)"
          >
            {{ (record as SysRoleItem).status === '2' ? '停用' : '启用' }}
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysRole:remove"
            @click="onDelete(record as SysRoleItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增角色"
      width="820"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <AdminModalFormFields :model="addForm" :fields="roleFormFields" />
      <div class="mt-4 grid gap-4">
        <section>
          <div class="mb-2 text-sm font-medium text-slate-700">菜单权限</div>
          <div
            class="app-radius-box max-h-48 overflow-auto border border-slate-200 bg-slate-50 p-3"
          >
            <Tree
              v-if="addMenuTreeData.length"
              checkable
              :tree-data="addMenuTreeData"
              :checked-keys="addMenuCheckedKeys"
              @check="onAddMenuCheck"
            />
            <span v-else class="text-gray-400">加载中…</span>
          </div>
        </section>
        <section>
          <div class="mb-2 text-sm font-medium text-slate-700">数据权限（部门）</div>
          <div
            class="app-radius-box max-h-48 overflow-auto border border-slate-200 bg-slate-50 p-3"
          >
            <Tree
              v-if="addDeptTreeData.length"
              checkable
              :tree-data="addDeptTreeData"
              :checked-keys="addDeptCheckedKeys"
              @check="onAddDeptCheck"
            />
            <span v-else class="text-gray-400">加载中…</span>
          </div>
        </section>
      </div>
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑角色"
      width="820"
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
      <AdminModalFormFields
        v-else
        :model="editForm"
        :fields="roleFormFields"
      />
      <div v-if="!editLoading" class="mt-4 grid gap-4">
        <section>
          <div class="mb-2 text-sm font-medium text-slate-700">菜单权限</div>
          <div
            class="app-radius-box max-h-48 overflow-auto border border-slate-200 bg-slate-50 p-3"
          >
            <Tree
              v-if="editMenuTreeData.length"
              checkable
              :tree-data="editMenuTreeData"
              :checked-keys="editMenuCheckedKeys"
              @check="onEditMenuCheck"
            />
            <span v-else class="text-gray-400">暂无数据</span>
          </div>
        </section>
        <section>
          <div class="mb-2 text-sm font-medium text-slate-700">数据权限（部门）</div>
          <div
            class="app-radius-box max-h-48 overflow-auto border border-slate-200 bg-slate-50 p-3"
          >
            <Tree
              v-if="editDeptTreeData.length"
              checkable
              :tree-data="editDeptTreeData"
              :checked-keys="editDeptCheckedKeys"
              @check="onEditDeptCheck"
            />
            <span v-else class="text-gray-400">暂无数据</span>
          </div>
        </section>
      </div>
    </Modal>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="角色详情"
      :loading="detailLoading"
      width="720"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="基础信息" description="角色标识、排序和当前状态。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">角色名称</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.roleName) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">权限字符</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.roleKey) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">排序</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.roleSort) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">状态</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderStatus(detailRecord.status) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">创建时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ formatAdminDateTime(detailRecord.createdAt) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">数据范围</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.dataScope) }}</dd>
            </div>
          </dl>
        </AdminDetailSection>

        <AdminDetailSection title="备注" description="保留角色的业务补充说明。">
          <p class="text-sm leading-6 text-slate-700">
            {{ renderAdminEmpty(detailRecord.remark) }}
          </p>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
