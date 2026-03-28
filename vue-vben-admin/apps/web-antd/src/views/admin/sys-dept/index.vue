<script lang="ts" setup>
/**
 * 系统管理 - 部门管理
 * 树表 + 搜索 + 新增/编辑/删除 + 父级选择排除自身及子节点
 */
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  InputNumber,
  Modal,
  Select,
  Table,
  TreeSelect,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createDeptApi,
  deleteDeptApi,
  getDeptDetailApi,
  getDeptListApi,
  updateDeptApi,
  type SysDeptItem,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import { useAdminTreeList } from '#/composables/use-admin-tree-list';
import { renderAdminEmpty, resolveAdminErrorMessage } from '#/utils/admin-crud';

/** 状态选项 */
const statusOptions = [
  { value: undefined, label: '全部' },
  { value: 1, label: '启用' },
  { value: 0, label: '禁用' },
];

/** 表单状态选项（不含"全部"） */
const formStatusOptions = [
  { value: 1, label: '启用' },
  { value: 0, label: '禁用' },
];

const {
  errorMsg,
  fetchList,
  loading,
  onReset,
  onSearch,
  query,
  treeData,
} = useAdminTreeList<
  SysDeptItem,
  {
    deptName: string;
    status: number | undefined;
  },
  {
    deptName?: string;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    deptName: currentQuery.deptName.trim() || undefined,
    status: currentQuery.status,
  }),
  createQuery: () => ({
    deptName: '',
    status: undefined,
  }),
  fallbackErrorMessage: '获取部门列表失败',
  fetcher: async (params) => getDeptListApi(params),
});

onMounted(() => {
  void fetchList();
});

const baseColumns: TableColumnType[] = [
  { title: '部门名称', dataIndex: 'deptName', key: 'deptName', width: 200 },
  {
    title: '负责人',
    dataIndex: 'leader',
    key: 'leader',
    width: 120,
    customRender: ({ text }) => renderAdminEmpty(text as string),
  },
  {
    title: '手机',
    dataIndex: 'phone',
    key: 'phone',
    width: 140,
    customRender: ({ text }) => renderAdminEmpty(text as string),
  },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }) => (text === 1 ? '启用' : '禁用'),
  },
  { title: '操作', key: 'action', width: 180, fixed: 'right' },
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
  tableId: 'sys-dept-list',
});

/* -------- 新增弹窗 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);

const addForm = reactive({
  parentId: 0,
  deptName: '',
  sort: 0,
  leader: '',
  phone: '',
  email: '',
  status: 1,
});

/* -------- 编辑弹窗 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editDeptId = ref<number | null>(null);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysDeptItem | null>(null);
const editForm = reactive({
  parentId: 0,
  deptName: '',
  sort: 0,
  leader: '',
  phone: '',
  email: '',
  status: 1,
});

/** 将部门树转为 TreeSelect 节点 */
function buildOptions(
  items: SysDeptItem[],
): { title: string; value: number; children?: any[] }[] {
  return items.map((item) => ({
    title: item.deptName,
    value: item.deptId,
    children: item.children?.length ? buildOptions(item.children) : undefined,
  }));
}

/** 上级部门树形选项（新增用） */
const parentTreeOptions = computed(() => [
  { title: '根部门', value: 0 },
  ...buildOptions(treeData.value),
]);

/** 从部门树中移除指定节点及其整棵子树，避免“选自己或后代为父级” */
function filterDeptTreeExcludeNode(
  nodes: SysDeptItem[],
  excludeDeptId: number,
): SysDeptItem[] {
  return nodes
    .filter((n) => n.deptId !== excludeDeptId)
    .map((n) => ({
      ...n,
      children: n.children?.length
        ? filterDeptTreeExcludeNode(n.children, excludeDeptId)
        : undefined,
    }));
}

/** 上级部门树形选项（编辑用）：排除当前节点及其子节点 */
const parentTreeOptionsForEdit = computed(() => {
  const id = editDeptId.value;
  if (id == null) return parentTreeOptions.value;
  const filtered = filterDeptTreeExcludeNode(treeData.value, id);
  return [{ title: '根部门', value: 0 }, ...buildOptions(filtered)];
});

function onAdd() {
  // 重置表单
  addForm.parentId = 0;
  addForm.deptName = '';
  addForm.sort = 0;
  addForm.leader = '';
  addForm.phone = '';
  addForm.email = '';
  addForm.status = 1;
  addVisible.value = true;
}

function onAddCancel() {
  addVisible.value = false;
}

async function onAddOk() {
  // 最小校验：部门名称必填
  if (!addForm.deptName.trim()) {
    message.error('部门名称不能为空');
    return;
  }

  addSubmitting.value = true;
  try {
    await createDeptApi({
      parentId: addForm.parentId,
      deptName: addForm.deptName.trim(),
      sort: addForm.sort,
      leader: addForm.leader.trim(),
      phone: addForm.phone.trim(),
      email: addForm.email.trim(),
      status: addForm.status,
    });
    message.success('新增成功');
    addVisible.value = false;
    // 重置表单
    addForm.parentId = 0;
    addForm.deptName = '';
    addForm.sort = 0;
    addForm.leader = '';
    addForm.phone = '';
    addForm.email = '';
    addForm.status = 1;
    // 刷新列表
    fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
  } finally {
    addSubmitting.value = false;
  }
}

/** 打开编辑弹窗：请求详情并回填 */
async function openEditModal(record: SysDeptItem) {
  editDeptId.value = record.deptId;
  editVisible.value = true;
  editLoading.value = true;
  try {
    const detail = await getDeptDetailApi(record.deptId);
    editForm.parentId = detail.parentId ?? 0;
    editForm.deptName = detail.deptName ?? '';
    editForm.sort = detail.sort ?? 0;
    editForm.leader = detail.leader ?? '';
    editForm.phone = detail.phone ?? '';
    editForm.email = detail.email ?? '';
    editForm.status = detail.status ?? 1;
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取部门详情失败'));
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

async function openDetail(record: SysDeptItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detailRecord.value = await getDeptDetailApi(record.deptId);
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取部门详情失败'));
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

async function onEditOk() {
  if (editDeptId.value == null) return;
  if (!editForm.deptName.trim()) {
    message.error('部门名称不能为空');
    return;
  }
  editSubmitting.value = true;
  try {
    await updateDeptApi(editDeptId.value, {
      parentId: editForm.parentId,
      deptName: editForm.deptName.trim(),
      sort: editForm.sort,
      leader: editForm.leader.trim(),
      phone: editForm.phone.trim(),
      email: editForm.email.trim(),
      status: editForm.status,
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '编辑失败'));
  } finally {
    editSubmitting.value = false;
  }
}

/** 删除：二次确认后调用后端删除并刷新列表 */
function onDelete(record: SysDeptItem) {
  const name = record.deptName || `部门ID:${record.deptId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除部门「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteDeptApi([record.deptId]);
        message.success('删除成功');
        fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>部门管理</template>
    <template #description>
      维护部门层级、负责人和状态。树形表格与 TreeSelect
      弹窗统一使用后台页面样式骨架。
    </template>
    <template #header-extra>
      <AdminActionButton type="primary" codes="admin:sysDept:add" @click="onAdd">
        新增部门
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="部门名称">
          <Input
            v-model:value="query.deptName"
            placeholder="请输入部门名称"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            placeholder="请选择状态"
            allow-clear
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
        <div class="text-base font-semibold text-slate-900">部门树列表</div>
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
      :data-source="treeData"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: scrollX }"
      row-key="deptId"
      size="middle"
      @resizeColumn="handleResizeColumn"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysDept:query"
            @click="openDetail(record as SysDeptItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysDept:edit"
            @click="openEditModal(record as SysDeptItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysDept:remove"
            @click="onDelete(record as SysDeptItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="addVisible"
      title="新增部门"
      :confirm-loading="addSubmitting"
      :width="820"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="上级部门" class="mb-0 md:col-span-2">
          <TreeSelect
            v-model:value="addForm.parentId"
            :tree-data="parentTreeOptions"
            show-search
            allow-clear
            placeholder="选择上级部门"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="部门名称" required class="mb-0">
          <Input
            v-model:value="addForm.deptName"
            placeholder="请输入部门名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="负责人" class="mb-0">
          <Input
            v-model:value="addForm.leader"
            placeholder="请输入负责人"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="手机" class="mb-0">
          <Input
            v-model:value="addForm.phone"
            placeholder="请输入手机"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="邮箱" class="mb-0">
          <Input
            v-model:value="addForm.email"
            placeholder="请输入邮箱"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="状态" class="mb-0">
          <Select
            v-model:value="addForm.status"
            :options="formStatusOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>

    <Modal
      v-model:open="editVisible"
      title="编辑部门"
      :confirm-loading="editSubmitting"
      :width="820"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <Form layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="上级部门" class="mb-0 md:col-span-2">
          <TreeSelect
            v-model:value="editForm.parentId"
            :tree-data="parentTreeOptionsForEdit"
            show-search
            allow-clear
            placeholder="选择上级部门"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="部门名称" required class="mb-0">
          <Input
            v-model:value="editForm.deptName"
            placeholder="请输入部门名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="editForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="负责人" class="mb-0">
          <Input
            v-model:value="editForm.leader"
            placeholder="请输入负责人"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="手机" class="mb-0">
          <Input
            v-model:value="editForm.phone"
            placeholder="请输入手机"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="邮箱" class="mb-0">
          <Input
            v-model:value="editForm.email"
            placeholder="请输入邮箱"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="状态" class="mb-0">
          <Select
            v-model:value="editForm.status"
            :options="formStatusOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="部门详情"
      :loading="detailLoading"
      width="720"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="基础信息" description="查看部门层级、排序和当前状态。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">部门 ID</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ detailRecord.deptId }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">部门名称</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.deptName) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">父级部门 ID</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.parentId) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">部门路径</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.deptPath) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">排序</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.sort) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">状态</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ detailRecord.status === 1 ? '启用' : '禁用' }}</dd>
            </div>
          </dl>
        </AdminDetailSection>

        <AdminDetailSection title="负责人信息" description="用于确认联系人和联系方式。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">负责人</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.leader) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">手机</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.phone) }}</dd>
            </div>
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">邮箱</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.email) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">创建时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.createdAt) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">更新时间</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.updatedAt) }}</dd>
            </div>
          </dl>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
