<script lang="ts" setup>
/**
 * 系统管理 - 角色管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 菜单树/部门树最小接入
 * 复用岗位管理母版，树选择参考 roleMenuTreeselect / roleDeptTreeselect
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
  Tree,
  message,
} from 'ant-design-vue';
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
  SysRolePageResult,
  UpdateRoleData,
} from '#/api/core/role';
import type { DeptLabel } from '#/api/core/dept';

/** 表格加载状态 */
const loading = ref(false);
const tableData = ref<SysRoleItem[]>([]);
const errorMsg = ref('');
type TreeNode = NonNullable<TreeProps['treeData']>[number];
type TreeKey = string | number;

/** 分页 */
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

/** 搜索 */
const searchRoleName = ref('');
const searchRoleKey = ref('');
const searchStatus = ref<string>('');

const statusOptions = [
  { value: '', label: '全部' },
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
];

/** 获取角色列表 */
async function fetchRoleList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: Record<string, unknown> = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchRoleName.value.trim())
      params.roleName = searchRoleName.value.trim();
    if (searchRoleKey.value.trim()) params.roleKey = searchRoleKey.value.trim();
    if (searchStatus.value !== '') params.status = searchStatus.value;
    const res: SysRolePageResult = await getRolePage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    errorMsg.value =
      err?.message || err?.response?.data?.msg || '加载角色列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  pagination.value.current = 1;
  fetchRoleList();
}

function onReset() {
  searchRoleName.value = '';
  searchRoleKey.value = '';
  searchStatus.value = '';
  pagination.value.current = 1;
  fetchRoleList();
}

function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchRoleList();
}

function renderStatus(s: string | undefined): string {
  if (s === '2') return '启用';
  if (s === '1') return '停用';
  return s || '-';
}

function renderEmpty(v: string | number | null | undefined): string {
  return v != null && v !== '' ? String(v) : '-';
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
    customRender: ({ text }: { text: string }) => renderEmpty(text),
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
    customRender: ({ text }: { text: string }) => renderEmpty(text),
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
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取角色详情失败');
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

/* -------- 状态切换 -------- */
async function onToggleStatus(record: SysRoleItem) {
  const next = record.status === '2' ? '1' : '2';
  const label = next === '2' ? '启用' : '停用';
  try {
    await updateRoleStatus(record.roleId, next);
    message.success(`${label}成功`);
    fetchRoleList();
  } catch (e: unknown) {
    const err = e as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || `${label}失败`);
  }
}

const statusEditOptions = [
  { value: '1', label: '停用' },
  { value: '2', label: '启用' },
];

onMounted(() => {
  fetchRoleList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">角色管理</h2>
      <div class="flex gap-2">
        <Button @click="fetchRoleList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增角色</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">角色名称：</span>
      <Input
        v-model:value="searchRoleName"
        placeholder="请输入角色名称"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="text-sm text-gray-600">权限字符：</span>
      <Input
        v-model:value="searchRoleKey"
        placeholder="请输入权限字符"
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
      :row-key="(record: SysRoleItem) => record.roleId"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysRoleItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            @click="onToggleStatus(record as SysRoleItem)"
          >
            {{ (record as SysRoleItem).status === '2' ? '停用' : '启用' }}
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysRoleItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增角色"
      width="720"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 5 }" :wrapper-col="{ span: 18 }" class="mt-4">
        <FormItem label="角色名称" required>
          <Input
            v-model:value="addForm.roleName"
            placeholder="请输入角色名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="权限字符" required>
          <Input
            v-model:value="addForm.roleKey"
            placeholder="请输入权限字符"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="addForm.roleSort"
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
        <FormItem label="数据范围">
          <Input
            v-model:value="addForm.dataScope"
            placeholder="选填"
            allow-clear
          />
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea
            v-model:value="addForm.remark"
            placeholder="请输入备注"
            allow-clear
            :rows="2"
          />
        </FormItem>
        <FormItem label="菜单权限">
          <div class="max-h-48 overflow-auto rounded border p-2">
            <Tree
              v-if="addMenuTreeData.length"
              checkable
              :tree-data="addMenuTreeData"
              :checked-keys="addMenuCheckedKeys"
              @check="onAddMenuCheck"
            />
            <span v-else class="text-gray-400">加载中…</span>
          </div>
        </FormItem>
        <FormItem label="数据权限（部门）">
          <div class="max-h-48 overflow-auto rounded border p-2">
            <Tree
              v-if="addDeptTreeData.length"
              checkable
              :tree-data="addDeptTreeData"
              :checked-keys="addDeptCheckedKeys"
              @check="onAddDeptCheck"
            />
            <span v-else class="text-gray-400">加载中…</span>
          </div>
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑角色"
      width="720"
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
        :label-col="{ span: 5 }"
        :wrapper-col="{ span: 18 }"
        class="mt-4"
      >
        <FormItem label="角色名称" required>
          <Input
            v-model:value="editForm.roleName"
            placeholder="请输入角色名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="权限字符" required>
          <Input
            v-model:value="editForm.roleKey"
            placeholder="请输入权限字符"
            allow-clear
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber
            v-model:value="editForm.roleSort"
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
        <FormItem label="数据范围">
          <Input
            v-model:value="editForm.dataScope"
            placeholder="选填"
            allow-clear
          />
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea
            v-model:value="editForm.remark"
            placeholder="请输入备注"
            allow-clear
            :rows="2"
          />
        </FormItem>
        <FormItem label="菜单权限">
          <div class="max-h-48 overflow-auto rounded border p-2">
            <Tree
              v-if="editMenuTreeData.length"
              checkable
              :tree-data="editMenuTreeData"
              :checked-keys="editMenuCheckedKeys"
              @check="onEditMenuCheck"
            />
            <span v-else class="text-gray-400">暂无数据</span>
          </div>
        </FormItem>
        <FormItem label="数据权限（部门）">
          <div class="max-h-48 overflow-auto rounded border p-2">
            <Tree
              v-if="editDeptTreeData.length"
              checkable
              :tree-data="editDeptTreeData"
              :checked-keys="editDeptCheckedKeys"
              @check="onEditDeptCheck"
            />
            <span v-else class="text-gray-400">暂无数据</span>
          </div>
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
