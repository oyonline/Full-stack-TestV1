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

const loading = ref(false);
const treeData = ref<SysDeptItem[]>([]);

/** 搜索条件 */
const searchDeptName = ref('');
const searchStatus = ref<number | undefined>(undefined);

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

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, any> = {};
    if (searchDeptName.value.trim()) {
      params.deptName = searchDeptName.value.trim();
    }
    if (searchStatus.value !== undefined) {
      params.status = searchStatus.value;
    }
    const data = await getDeptListApi(params);
    treeData.value = Array.isArray(data) ? data : [];
  } catch (e: any) {
    message.error(e?.message || '获取部门列表失败');
    treeData.value = [];
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  fetchList();
}

function onReset() {
  searchDeptName.value = '';
  searchStatus.value = undefined;
  fetchList();
}

function onRefresh() {
  fetchList();
}

onMounted(() => {
  fetchList();
});

const columns: TableColumnType[] = [
  { title: '部门名称', dataIndex: 'deptName', key: 'deptName', width: 200 },
  { title: '负责人', dataIndex: 'leader', key: 'leader', width: 120 },
  { title: '手机', dataIndex: 'phone', key: 'phone', width: 140 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }) => (text === 1 ? '启用' : '禁用'),
  },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
];

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
function buildOptions(items: SysDeptItem[]): { title: string; value: number; children?: any[] }[] {
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
      children:
        n.children?.length
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
  } catch (e: any) {
    message.error(e?.message || '新增失败');
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
  } catch (e: any) {
    message.error(e?.message || '获取部门详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
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
  } catch (e: any) {
    message.error(e?.message || '编辑失败');
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
      } catch (e: any) {
        message.error(e?.message || '删除失败');
      }
    },
  });
}
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">部门管理</h2>
      <div class="flex gap-2">
        <Button type="primary" @click="onAdd">新增部门</Button>
        <Button @click="onRefresh">刷新</Button>
      </div>
    </div>

    <!-- 搜索区 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">部门名称：</span>
      <Input
        v-model:value="searchDeptName"
        placeholder="请输入部门名称"
        allow-clear
        class="w-48"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">状态：</span>
      <Select
        v-model:value="searchStatus"
        :options="statusOptions"
        placeholder="请选择状态"
        allow-clear
        class="w-28"
      />
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </div>

    <Table
      :columns="columns"
      :data-source="treeData"
      :loading="loading"
      :pagination="false"
      row-key="deptId"
      size="small"
      bordered
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysDeptItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysDeptItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增部门"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="上级部门">
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
        <FormItem label="部门名称" required>
          <Input
            v-model:value="addForm.deptName"
            placeholder="请输入部门名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="负责人">
          <Input
            v-model:value="addForm.leader"
            placeholder="请输入负责人"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="手机">
          <Input
            v-model:value="addForm.phone"
            placeholder="请输入手机"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="邮箱">
          <Input
            v-model:value="addForm.email"
            placeholder="请输入邮箱"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="addForm.status"
            :options="formStatusOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑部门"
      :confirm-loading="editSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="上级部门">
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
        <FormItem label="部门名称" required>
          <Input
            v-model:value="editForm.deptName"
            placeholder="请输入部门名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="editForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="负责人">
          <Input
            v-model:value="editForm.leader"
            placeholder="请输入负责人"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="手机">
          <Input
            v-model:value="editForm.phone"
            placeholder="请输入手机"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="邮箱">
          <Input
            v-model:value="editForm.email"
            placeholder="请输入邮箱"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="editForm.status"
            :options="formStatusOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
