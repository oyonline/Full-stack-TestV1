<script lang="ts" setup>
/**
 * 系统管理 - 菜单管理
 * 最小可用：树形列表 + 搜索 + 操作列（编辑弹窗已接通、删除已接通）。
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

import { requestClient } from '#/api/request';

/* -------- component 最小校验：与 access.ts 的路径集合对齐 -------- */
/** 与 access.ts normalizeViewPath 一致：去 ./ ../、补前导 /、去 /views 前缀 */
function normalizeViewPath(path: string): string {
  const n = path.replace(/^(\.\/|\.\.\/)+/, '');
  const viewPath = n.startsWith('/') ? n : `/${n}`;
  return viewPath.replace(/^\/views/, '');
}

/** 与 access.ts buildValidViewPathSet 一致：从 glob 得到“有效视图路径”集合（不含 .vue） */
function buildValidViewPathSet(pageMap: Record<string, unknown>): Set<string> {
  const set = new Set<string>();
  for (const key of Object.keys(pageMap)) {
    const n = normalizeViewPath(key);
    set.add(n.endsWith('.vue') ? n.slice(0, -4) : n);
  }
  return set;
}

const validViewPathSet = buildValidViewPathSet(
  import.meta.glob('../../**/*.vue') as Record<string, unknown>,
);

/**
 * 校验 component 是否允许提交：空允许；Layout/BasicLayout/IFrameView 允许；
 * 非空则必须在 validViewPathSet 内（与 access 的 mapComponent 一致，支持 /index 双向匹配）。
 */
function isValidComponent(comp: string): boolean {
  const c = comp?.trim() ?? '';
  if (c === '') return true;
  if (/^Layout$/i.test(c) || /^BasicLayout$/i.test(c) || /^IFrameView$/i.test(c))
    return true;
  let candidate = normalizeViewPath(c);
  if (candidate.endsWith('.vue')) candidate = candidate.slice(0, -4);
  if (validViewPathSet.has(candidate)) return true;
  if (candidate.endsWith('/index')) {
    if (validViewPathSet.has(candidate.replace(/\/index$/, ''))) return true;
  } else {
    if (validViewPathSet.has(candidate + '/index')) return true;
  }
  return false;
}

/** 候选项：Layout/BasicLayout/IFrameView + 当前项目有效视图路径（与 validViewPathSet 一致） */
const componentOptions = [
  { value: 'Layout', label: 'Layout' },
  { value: 'BasicLayout', label: 'BasicLayout' },
  { value: 'IFrameView', label: 'IFrameView' },
  ...Array.from(validViewPathSet)
    .sort()
    .map((v) => ({ value: v, label: v })),
];

/** 可搜索 Select 的过滤：按 label 模糊匹配 */
function filterComponentOption(input: string, option?: unknown): boolean {
  const o = option as { value?: string; label?: string } | undefined;
  const text = (o?.label ?? o?.value ?? '').toString().toLowerCase();
  return text.includes((input ?? '').trim().toLowerCase());
}

/**
 * 将后端返回的 component 转为候选项中的规范值（用于编辑回显，使 Select 能匹配到选项）
 */
function getComponentOptionValue(comp: string): string {
  const c = comp?.trim() ?? '';
  if (c === '') return '';
  if (/^Layout$/i.test(c) || /^BasicLayout$/i.test(c) || /^IFrameView$/i.test(c))
    return c;
  let candidate = normalizeViewPath(c);
  if (candidate.endsWith('.vue')) candidate = candidate.slice(0, -4);
  if (validViewPathSet.has(candidate)) return candidate;
  if (candidate.endsWith('/index')) {
    const without = candidate.replace(/\/index$/, '');
    if (validViewPathSet.has(without)) return without;
  } else {
    if (validViewPathSet.has(candidate + '/index')) return candidate + '/index';
  }
  return comp;
}

/** go-admin 菜单树节点（列表用，含 children） */
interface SysMenuRow {
  menuId: number;
  menuName?: string;
  title?: string;
  path?: string;
  menuType?: string;
  sort?: number;
  visible?: string;
  children?: SysMenuRow[];
}

/** TreeSelect 节点：value=menuId，title=显示名 */
interface ParentTreeNode {
  value: number;
  title: string;
  children?: ParentTreeNode[];
}

/** 从菜单树中移除指定 menuId 节点（及其整棵子树），避免“选自己或后代为父级” */
function filterTreeExcludeNode(
  nodes: SysMenuRow[],
  excludeMenuId: number,
): SysMenuRow[] {
  return nodes
    .filter((n) => n.menuId !== excludeMenuId)
    .map((n) => ({
      ...n,
      children: n.children?.length
        ? filterTreeExcludeNode(n.children, excludeMenuId)
        : undefined,
    }));
}

/** SysMenuRow 转为 TreeSelect 节点 */
function menuRowToParentNode(row: SysMenuRow): ParentTreeNode {
  return {
    value: row.menuId,
    title: row.title ?? row.menuName ?? `菜单 ${row.menuId}`,
    children: row.children?.length
      ? row.children.map(menuRowToParentNode)
      : undefined,
  };
}

/** GET /api/v1/menu/:id 返回的完整详情（与后端 models.SysMenu 对齐） */
interface SysMenuDetail {
  menuId: number;
  menuName: string;
  title: string;
  icon: string;
  path: string;
  paths: string;
  menuType: string;
  action: string;
  permission: string;
  parentId: number;
  noCache: boolean;
  breadcrumb: string;
  component: string;
  sort: number;
  visible: string;
  isFrame: string;
  sysApi: any[];
  apis: number[];
}

const loading = ref(false);
const treeData = ref<SysMenuRow[]>([]);
const errorMsg = ref('');

/** 搜索：标题（模糊） */
const searchTitle = ref('');
/** 搜索：显示状态（后端 visible 为 int：1 显示 0 隐藏） */
const searchVisible = ref<'' | '0' | '1'>('');

const visibleOptions = [
  { value: '' as const, label: '全部' },
  { value: '1' as const, label: '显示' },
  { value: '0' as const, label: '隐藏' },
];

/** 编辑/新增弹窗共用的可见下拉（不含"全部"）*/
const yesNoOptions = [
  { value: '1', label: '显示' },
  { value: '0', label: '隐藏' },
];

/** 菜单类型：与 go-admin 后端 sys_menu.menu_type 一致（M=目录 C=菜单 F=按钮） */
const menuTypeOptions = [
  { value: 'M', label: '目录' },
  { value: 'C', label: '菜单' },
  { value: 'F', label: '按钮' },
];

async function fetchMenuList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: Record<string, string | number> = {};
    if (searchTitle.value.trim()) params.title = searchTitle.value.trim();
    if (searchVisible.value !== '') params.visible = Number(searchVisible.value);
    const data = await requestClient.get<SysMenuRow[]>('/v1/menu', { params });
    treeData.value = Array.isArray(data) ? data : [];
  } catch (e: any) {
    errorMsg.value = e?.message || e?.response?.data?.msg || '加载菜单列表失败';
    treeData.value = [];
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  fetchMenuList();
}

function onReset() {
  searchTitle.value = '';
  searchVisible.value = '';
  fetchMenuList();
}

/* -------- 编辑弹窗 -------- */

const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
/** 从 GET /api/v1/menu/:id 取到的完整详情，提交时作为基底与表单合并 */
const fullDetail = ref<SysMenuDetail | null>(null);
/** 编辑表单数据：含 component、parentId、menuType 等可编辑字段，其余由 fullDetail 保持原值 */
const editForm = reactive({
  menuName: '',
  title: '',
  path: '',
  component: '',
  parentId: 0 as number,
  menuType: 'C' as string,
  permission: '',
  sort: 0,
  visible: '1',
});

/** 父级候选项（新增用）：根节点 + 完整菜单树，基于当前页 treeData */
const parentTreeOptionsAdd = computed<ParentTreeNode[]>(() => [
  { value: 0, title: '根节点' },
  ...treeData.value.map(menuRowToParentNode),
]);

/** 父级候选项（编辑用）：根节点 + 排除当前节点及其子树，避免选自己或后代为父级 */
const parentTreeOptionsForEdit = computed<ParentTreeNode[]>(() => {
  const excludeId = fullDetail.value?.menuId;
  if (excludeId == null) return parentTreeOptionsAdd.value;
  const filtered = filterTreeExcludeNode(treeData.value, excludeId);
  return [{ value: 0, title: '根节点' }, ...filtered.map(menuRowToParentNode)];
});

/**
 * 打开编辑弹窗：先请求 GET /api/v1/menu/:id 取完整详情。
 * 原因：PUT /api/v1/menu/:id 后端 Generate() 全量覆盖所有字段，
 * 必须把未编辑字段（icon、component、permission 等）一并传回，才能避免被零值覆盖。
 */
async function onEdit(record: SysMenuRow) {
  fullDetail.value = null;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await requestClient.get<SysMenuDetail>(
      `/v1/menu/${record.menuId}`,
    );
    fullDetail.value = detail;
    editForm.menuName = detail.menuName ?? '';
    editForm.title = detail.title ?? '';
    editForm.path = detail.path ?? '';
    editForm.component = getComponentOptionValue(detail.component ?? '');
    editForm.parentId = detail.parentId ?? 0;
    editForm.menuType = detail.menuType ?? 'C';
    editForm.permission = detail.permission ?? '';
    editForm.sort = detail.sort ?? 0;
    editForm.visible = detail.visible ?? '1';
  } catch (e: any) {
    message.error(e?.message || '获取菜单详情失败，请重试');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

/** menuType 对 component 的保存前校验：C 必填且有效，M/F 非空则须有效 */
function validateComponentByMenuType(
  menuType: string,
  component: string,
): { ok: boolean; message?: string } {
  const comp = (component ?? '').trim();
  if (menuType === 'C') {
    if (!comp) return { ok: false, message: '菜单类型为「菜单」时，组件路径不能为空' };
  }
  if (comp && !isValidComponent(component)) {
    return {
      ok: false,
      message: '组件路径不在当前项目有效视图中，请填写如 views/admin/sys-menu/index 或 Layout',
    };
  }
  return { ok: true };
}

/**
 * 提交编辑：将 fullDetail（完整原始数据）与表单修改字段合并后，
 * 整体送入 PUT /api/v1/menu/:id，确保未编辑字段不被覆盖。
 */
async function onEditOk() {
  if (!fullDetail.value) return;
  const v = validateComponentByMenuType(editForm.menuType, editForm.component);
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await requestClient.put(`/v1/menu/${fullDetail.value.menuId}`, {
      ...fullDetail.value,
      menuName: editForm.menuName,
      title: editForm.title,
      path: editForm.path,
      component: editForm.component,
      parentId: editForm.parentId ?? 0,
      menuType: editForm.menuType,
      permission: editForm.permission ?? '',
      sort: editForm.sort,
      visible: editForm.visible,
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchMenuList();
  } catch (e: any) {
    message.error(e?.message || e?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

/** 编辑弹窗内切换 menuType 为 F 时清空 component，避免误带页面组件 */
function onEditMenuTypeChange(value: unknown) {
  if (value === 'F') editForm.component = '';
}

/** 新增弹窗内切换 menuType 为 F 时清空 component */
function onAddMenuTypeChange(value: unknown) {
  if (value === 'F') addForm.component = '';
}

/* -------- 新增弹窗 -------- */

const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  menuName: '',
  title: '',
  path: '',
  component: '',
  parentId: 0 as number,
  menuType: 'C' as string,
  permission: '',
  sort: 0,
  visible: '1',
});

function onAdd() {
  addForm.menuName = '';
  addForm.title = '';
  addForm.path = '';
  addForm.component = '';
  addForm.parentId = 0;
  addForm.menuType = 'C';
  addForm.permission = '';
  addForm.sort = 0;
  addForm.visible = '1';
  addVisible.value = true;
}

/**
 * 提交新增：POST /api/v1/menu。menuType 由表单选择（M/C/F）；isFrame 固定 "0"。
 */
async function onAddOk() {
  const v = validateComponentByMenuType(addForm.menuType, addForm.component);
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await requestClient.post('/v1/menu', {
      menuName: addForm.menuName,
      title: addForm.title,
      path: addForm.path,
      component: addForm.component,
      parentId: addForm.parentId ?? 0,
      menuType: addForm.menuType,
      permission: addForm.permission ?? '',
      sort: addForm.sort,
      visible: addForm.visible,
      isFrame: '0',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchMenuList();
  } catch (e: any) {
    message.error(e?.message || e?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

function onAddCancel() {
  addVisible.value = false;
}

/** 删除：二次确认后调用后端删除并刷新列表 */
function onDelete(record: SysMenuRow) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除菜单「${record.title ?? record.menuName ?? record.menuId}」吗？`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await requestClient.delete('/v1/menu', { data: { ids: [record.menuId] } });
        message.success('删除成功');
        fetchMenuList();
      } catch (e: any) {
        message.error(e?.message || e?.response?.data?.msg || '删除失败');
      }
    },
  });
}

onMounted(() => {
  fetchMenuList();
});

const columns: TableColumnType[] = [
  { title: '菜单ID', dataIndex: 'menuId', key: 'menuId', width: 90 },
  { title: '菜单名称', dataIndex: 'menuName', key: 'menuName', width: 140 },
  { title: '显示标题', dataIndex: 'title', key: 'title', width: 120 },
  { title: '路径', dataIndex: 'path', key: 'path', ellipsis: true },
  { title: '类型', dataIndex: 'menuType', key: 'menuType', width: 80 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 70 },
  { title: '可见', dataIndex: 'visible', key: 'visible', width: 70 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">菜单管理</h2>
      <Button type="primary" @click="onAdd">新增菜单</Button>
    </div>
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">标题：</span>
      <Input
        v-model:value="searchTitle"
        placeholder="模糊匹配"
        allow-clear
        class="w-40"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">可见：</span>
      <Select
        v-model:value="searchVisible"
        :options="visibleOptions"
        class="w-28"
      />
      <Button type="primary" size="small" @click="onSearch">查询</Button>
      <Button size="small" @click="onReset">重置</Button>
    </div>
    <div v-if="errorMsg" class="mb-4 text-red-600">
      {{ errorMsg }}
    </div>
    <Table
      :columns="columns"
      :data-source="treeData"
      :loading="loading"
      :pagination="false"
      row-key="menuId"
      size="small"
      bordered
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="onEdit(record)">编辑</Button>
          <Button type="link" size="small" danger @click="onDelete(record)">
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑菜单"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载中…
      </div>
      <Form
        v-else
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        class="mt-4"
      >
        <FormItem label="菜单名称">
          <Input v-model:value="editForm.menuName" placeholder="menuName" />
        </FormItem>
        <FormItem label="显示标题">
          <Input v-model:value="editForm.title" placeholder="title" />
        </FormItem>
        <FormItem label="类型(menuType)">
          <Select
            v-model:value="editForm.menuType"
            :options="menuTypeOptions"
            class="w-full"
            @change="onEditMenuTypeChange"
          />
        </FormItem>
        <FormItem label="路径">
          <Input v-model:value="editForm.path" placeholder="path" />
        </FormItem>
        <FormItem label="上级菜单">
          <TreeSelect
            v-model:value="editForm.parentId"
            :tree-data="parentTreeOptionsForEdit"
            show-search
            allow-clear
            placeholder="根节点或选择父级菜单"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="组件(component)">
          <Select
            v-model:value="editForm.component"
            show-search
            allow-clear
            placeholder="选择或输入 Layout / 视图路径"
            :options="componentOptions"
            :filter-option="filterComponentOption"
            option-filter-prop="label"
            class="w-full"
          />
        </FormItem>
        <FormItem label="权限标识(permission)">
          <Input
            v-model:value="editForm.permission"
            placeholder="如 system:user:list，按钮(F)类型常用"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="editForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="可见">
          <Select v-model:value="editForm.visible" :options="yesNoOptions" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增菜单"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="菜单名称">
          <Input v-model:value="addForm.menuName" placeholder="menuName" />
        </FormItem>
        <FormItem label="显示标题">
          <Input v-model:value="addForm.title" placeholder="title" />
        </FormItem>
        <FormItem label="类型(menuType)">
          <Select
            v-model:value="addForm.menuType"
            :options="menuTypeOptions"
            class="w-full"
            @change="onAddMenuTypeChange"
          />
        </FormItem>
        <FormItem label="路径">
          <Input v-model:value="addForm.path" placeholder="path" />
        </FormItem>
        <FormItem label="上级菜单">
          <TreeSelect
            v-model:value="addForm.parentId"
            :tree-data="parentTreeOptionsAdd"
            show-search
            allow-clear
            placeholder="根节点或选择父级菜单"
            tree-node-filter-prop="title"
            class="w-full"
          />
        </FormItem>
        <FormItem label="组件(component)">
          <Select
            v-model:value="addForm.component"
            show-search
            allow-clear
            placeholder="选择或输入 Layout / 视图路径"
            :options="componentOptions"
            :filter-option="filterComponentOption"
            option-filter-prop="label"
            class="w-full"
          />
        </FormItem>
        <FormItem label="权限标识(permission)">
          <Input
            v-model:value="addForm.permission"
            placeholder="如 system:user:list，按钮(F)类型常用"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="可见">
          <Select v-model:value="addForm.visible" :options="yesNoOptions" />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
