<script lang="ts" setup>
import type { SkuCategoryTreeNode } from '#/api/core';

/**
 * 产品中心 - SKU 类目管理（树形）
 * 列表（Tree 递归）+ 新增 / 编辑 / 删除（有子节点禁止）
 * 拖拽排序留 phase3，不在本期实现。
 */
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Tree,
  TreeSelect,
} from 'ant-design-vue';

import {
  createSkuCategory,
  deleteSkuCategory,
  getSkuCategoryTree,
  updateSkuCategory,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTreeList } from '#/composables/use-admin-tree-list';
import { resolveAdminErrorMessage } from '#/utils/admin-crud';

/** 状态选项（后端：1=禁用 2=启用） */
const statusOptions = [
  { value: undefined, label: '全部' },
  { value: 2, label: '启用' },
  { value: 1, label: '禁用' },
];

const formStatusOptions = [
  { value: 2, label: '启用' },
  { value: 1, label: '禁用' },
];

const { errorMsg, fetchList, loading, onReset, onSearch, query, treeData } =
  useAdminTreeList<
    SkuCategoryTreeNode,
    {
      categoryName: string;
      status: number | undefined;
    },
    {
      categoryName?: string;
      status?: number;
    }
  >({
    createParams: (currentQuery) => ({
      categoryName: currentQuery.categoryName.trim() || undefined,
      status: currentQuery.status,
    }),
    createQuery: () => ({
      categoryName: '',
      status: undefined,
    }),
    fallbackErrorMessage: '获取类目列表失败',
    fetcher: async (params) => getSkuCategoryTree(params),
  });

onMounted(() => {
  void fetchList();
});

/** 默认展开所有节点 */
const expandedKeys = ref<number[]>([]);

function collectKeys(nodes: SkuCategoryTreeNode[], out: number[]) {
  for (const node of nodes) {
    out.push(node.categoryId);
    if (node.children?.length) collectKeys(node.children, out);
  }
}

function refreshExpandedKeys() {
  const keys: number[] = [];
  collectKeys(treeData.value, keys);
  expandedKeys.value = keys;
}

async function reload() {
  await fetchList();
  refreshExpandedKeys();
}

/** TreeSelect 节点（用于父级选择） */
function buildOptions(
  items: SkuCategoryTreeNode[],
): { children?: any[]; title: string; value: number }[] {
  return items.map((item) => ({
    title: item.categoryName,
    value: item.categoryId,
    children: item.children?.length ? buildOptions(item.children) : undefined,
  }));
}

const parentTreeOptions = computed(() => [
  { title: '根类目', value: 0 },
  ...buildOptions(treeData.value),
]);

/** 编辑时排除自身及子树，避免选自己或后代为父级 */
function filterTreeExcludeNode(
  nodes: SkuCategoryTreeNode[],
  excludeId: number,
): SkuCategoryTreeNode[] {
  return nodes
    .filter((n) => n.categoryId !== excludeId)
    .map((n) => ({
      ...n,
      children: n.children?.length
        ? filterTreeExcludeNode(n.children, excludeId)
        : [],
    }));
}

const parentTreeOptionsForEdit = computed(() => {
  const id = editCategoryId.value;
  if (id === null) return parentTreeOptions.value;
  const filtered = filterTreeExcludeNode(treeData.value, id);
  return [{ title: '根类目', value: 0 }, ...buildOptions(filtered)];
});

/** Tree 数据：把 SkuCategoryTreeNode 适配成 antd Tree 的 fieldNames 形式 */
const treeFieldNames = {
  children: 'children',
  title: 'categoryName',
  key: 'categoryId',
};

/* -------- 新增弹窗 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);

const addForm = reactive({
  parentId: 0,
  categoryName: '',
  sort: 0,
  status: 2,
});

function resetAddForm(parentId = 0) {
  addForm.parentId = parentId;
  addForm.categoryName = '';
  addForm.sort = 0;
  addForm.status = 2;
}

function onAdd(parentId = 0) {
  resetAddForm(parentId);
  addVisible.value = true;
}

function onAddCancel() {
  addVisible.value = false;
}

async function onAddOk() {
  if (!addForm.categoryName.trim()) {
    message.error('类目名称不能为空');
    return;
  }
  addSubmitting.value = true;
  try {
    await createSkuCategory({
      parentId: addForm.parentId,
      categoryName: addForm.categoryName.trim(),
      sort: addForm.sort,
      status: addForm.status,
    });
    message.success('新增成功');
    addVisible.value = false;
    resetAddForm();
    await reload();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
  } finally {
    addSubmitting.value = false;
  }
}

/* -------- 编辑弹窗 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editCategoryId = ref<null | number>(null);
const editForm = reactive({
  parentId: 0,
  categoryName: '',
  sort: 0,
  status: 2,
});

function openEditModal(record: SkuCategoryTreeNode) {
  editCategoryId.value = record.categoryId;
  editForm.parentId = record.parentId ?? 0;
  editForm.categoryName = record.categoryName ?? '';
  editForm.sort = record.sort ?? 0;
  editForm.status = record.status ?? 2;
  editVisible.value = true;
}

function onEditCancel() {
  editVisible.value = false;
}

async function onEditOk() {
  if (editCategoryId.value === null) return;
  if (!editForm.categoryName.trim()) {
    message.error('类目名称不能为空');
    return;
  }
  editSubmitting.value = true;
  try {
    await updateSkuCategory(editCategoryId.value, {
      parentId: editForm.parentId,
      categoryName: editForm.categoryName.trim(),
      sort: editForm.sort,
      status: editForm.status,
    });
    message.success('编辑成功');
    editVisible.value = false;
    await reload();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '编辑失败'));
  } finally {
    editSubmitting.value = false;
  }
}

/** 删除：有子节点则禁止；二次确认后调用后端 */
function onDelete(record: SkuCategoryTreeNode) {
  if (record.children && record.children.length > 0) {
    message.warning('该类目存在子节点，请先删除子类目');
    return;
  }
  const name = record.categoryName || `类目ID:${record.categoryId}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除类目「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSkuCategory([record.categoryId]);
        message.success('删除成功');
        await reload();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>Product Center</template>
    <template #title>类目管理</template>
    <template #description>
      维护 SKU
      类目树层级、排序和启用状态。新增子类与编辑、删除均在节点上原地操作。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:category:add"
        @click="onAdd(0)"
      >
        新增根类目
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="类目名称">
          <Input
            v-model:value="query.categoryName"
            placeholder="请输入类目名称"
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
        <div class="text-base font-semibold text-slate-900">类目树</div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <div v-if="loading" class="py-10 text-center text-sm text-slate-400">
      加载中…
    </div>
    <div
      v-else-if="treeData.length === 0"
      class="py-10 text-center text-sm text-slate-400"
    >
      暂无数据
    </div>
    <Tree
      v-else
      :tree-data="treeData as any"
      :field-names="treeFieldNames"
      :expanded-keys="expandedKeys"
      :selectable="false"
      block-node
      @expand="(keys) => (expandedKeys = keys as number[])"
    >
      <template #title="node">
        <div class="flex w-full items-center gap-3">
          <span class="font-medium text-slate-900">{{
            node.categoryName
          }}</span>
          <span
            class="rounded-sm bg-slate-100 px-1.5 py-0.5 text-xs text-slate-500"
          >
            排序 {{ node.sort }}
          </span>
          <span
            class="rounded-sm px-1.5 py-0.5 text-xs"
            :class="
              node.status === 2
                ? 'bg-green-50 text-green-600'
                : 'bg-slate-100 text-slate-500'
            "
          >
            {{ node.status === 2 ? '启用' : '禁用' }}
          </span>
          <div class="ml-auto flex items-center gap-1">
            <AdminActionButton
              type="link"
              size="small"
              codes="admin:category:add"
              @click.stop="onAdd(node.categoryId)"
            >
              新增子类
            </AdminActionButton>
            <AdminActionButton
              type="link"
              size="small"
              codes="admin:category:edit"
              @click.stop="openEditModal(node as SkuCategoryTreeNode)"
            >
              编辑
            </AdminActionButton>
            <AdminActionButton
              type="link"
              size="small"
              danger
              codes="admin:category:remove"
              @click.stop="onDelete(node as SkuCategoryTreeNode)"
            >
              删除
            </AdminActionButton>
          </div>
        </div>
      </template>
    </Tree>

    <Modal
      v-model:open="addVisible"
      title="新增类目"
      :confirm-loading="addSubmitting"
      :width="640"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="父级类目" class="mb-0 md:col-span-2">
          <TreeSelect
            v-model:value="addForm.parentId"
            :tree-data="parentTreeOptions"
            show-search
            allow-clear
            placeholder="选择父级类目"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="类目名称" required class="mb-0">
          <Input
            v-model:value="addForm.categoryName"
            placeholder="请输入类目名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
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
      title="编辑类目"
      :confirm-loading="editSubmitting"
      :width="640"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <Form layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="父级类目" class="mb-0 md:col-span-2">
          <TreeSelect
            v-model:value="editForm.parentId"
            :tree-data="parentTreeOptionsForEdit"
            show-search
            allow-clear
            placeholder="选择父级类目"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="类目名称" required class="mb-0">
          <Input
            v-model:value="editForm.categoryName"
            placeholder="请输入类目名称"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="editForm.sort" :min="0" class="w-full" />
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
  </AdminPageShell>
</template>
