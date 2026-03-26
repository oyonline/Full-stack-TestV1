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
  Tag,
  TreeSelect,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import { IconifyIcon } from '@vben/icons';
import { IconPicker } from '@vben/common-ui';

import { getSysApiList } from '#/api/core';
import type { SysApiItem } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTreeList } from '#/composables/use-admin-tree-list';
import { renderAdminEmpty, resolveAdminErrorMessage } from '#/utils/admin-crud';
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
	if (
		/^Layout$/i.test(c) ||
		/^BasicLayout$/i.test(c) ||
		/^RouteView$/i.test(c) ||
		/^IFrameView$/i.test(c)
	)
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
  { value: 'RouteView', label: 'RouteView' },
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
	if (
		/^Layout$/i.test(c) ||
		/^BasicLayout$/i.test(c) ||
		/^RouteView$/i.test(c) ||
		/^IFrameView$/i.test(c)
	)
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

/**
 * 短 key 转 Iconify（与 access.ts 的 normalizeMenuIcon 一致，供编辑回填 IconPicker 使用）
 */
const ICON_SHORT_KEY_MAP: Record<string, string> = {
  'api-server': 'ant-design:cloud-server-outlined',
  'tree-table': 'ant-design:table-outlined',
  user: 'ant-design:user-outlined',
  peoples: 'ant-design:team-outlined',
  swagger: 'ant-design:api-outlined',
  guide: 'ant-design:read-outlined',
  education: 'ant-design:read-outlined',
  logininfor: 'ant-design:history-outlined',
  skill: 'ant-design:tool-outlined',
  bug: 'ant-design:bug-outlined',
  build: 'ant-design:build-outlined',
  code: 'ant-design:code-outlined',
  log: 'ant-design:file-text-outlined',
  pass: 'ant-design:key-outlined',
  job: 'ant-design:clock-circle-outlined',
  'system-tools': 'ant-design:tool-outlined',
  'dev-tools': 'ant-design:experiment-outlined',
  'time-range': 'ant-design:clock-circle-outlined',
  tree: 'ant-design:cluster-outlined',
  druid: 'ant-design:monitor-outlined',
  'api-doc': 'ant-design:api-outlined',
};
function normalizeMenuIcon(icon?: string): string | undefined {
  if (!icon) return undefined;
  if (icon.includes(':')) return icon;
  return ICON_SHORT_KEY_MAP[icon] ?? 'ant-design:question-outlined';
}

/** 列表/详情中接口项（与后端 SysApi 对齐） */
interface SysApiInMenu {
  id: number;
  title?: string;
  path?: string;
  action?: string;
}

/** go-admin 菜单树节点（列表用，含 children；列表接口已 Preload SysApi） */
interface SysMenuRow {
  menuId: number;
  menuName?: string;
  title?: string;
  icon?: string;
  path?: string;
  menuType?: string;
  sort?: number;
  visible?: string;
  sysApi?: SysApiInMenu[];
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

/** 关联接口多选：选项来自 GET /api/v1/sys-api，value=id，label=title + path + action */
const apiOptions = ref<{ value: number; label: string }[]>([]);

const visibleOptions = [
  { value: '' as const, label: '全部' },
  { value: '1' as const, label: '显示' },
  { value: '0' as const, label: '隐藏' },
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
  SysMenuRow,
  {
    title: string;
    visible: '' | '0' | '1';
  },
  {
    title?: string;
    visible?: number;
  }
>({
  createParams: (currentQuery) => ({
    title: currentQuery.title.trim() || undefined,
    visible:
      currentQuery.visible !== '' ? Number(currentQuery.visible) : undefined,
  }),
  createQuery: () => ({
    title: '',
    visible: '' as const,
  }),
  fallbackErrorMessage: '加载菜单列表失败',
  fetcher: async (params) =>
    requestClient.get<SysMenuRow[]>('/v1/menu', { params }),
});

const fetchMenuList = fetchList;

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

/* -------- 编辑弹窗 -------- */

const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysMenuDetail | null>(null);
/** 从 GET /api/v1/menu/:id 取到的完整详情，提交时作为基底与表单合并 */
const fullDetail = ref<SysMenuDetail | null>(null);
/** 编辑表单数据：含 component、parentId、menuType、apis 等可编辑字段，其余由 fullDetail 保持原值 */
const editForm = reactive({
  menuName: '',
  title: '',
  icon: '',
  path: '',
  component: '',
  parentId: 0 as number,
  menuType: 'C' as string,
  permission: '',
  apis: [] as number[],
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
    editForm.icon = detail.icon ? (normalizeMenuIcon(detail.icon) ?? '') : '';
    editForm.path = detail.path ?? '';
    editForm.component = getComponentOptionValue(detail.component ?? '');
    editForm.parentId = detail.parentId ?? 0;
    editForm.menuType = detail.menuType ?? 'C';
    editForm.permission = detail.permission ?? '';
    editForm.apis = Array.isArray(detail.apis) ? [...detail.apis] : [];
    editForm.sort = detail.sort ?? 0;
    editForm.visible = detail.visible ?? '1';
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取菜单详情失败，请重试'));
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

async function openDetail(record: SysMenuRow) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detailRecord.value = await requestClient.get<SysMenuDetail>(
      `/v1/menu/${record.menuId}`,
    );
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '获取菜单详情失败，请重试'));
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

/** menuType 对 component 的保存前校验：C 必填且有效，M/F 非空则须有效 */
function validateComponentByMenuType(
  menuType: string,
  component: string,
): { ok: boolean; message?: string } {
  const comp = (component ?? '').trim();
  if (menuType === 'C') {
    if (!comp)
      return { ok: false, message: '菜单类型为「菜单」时，组件路径不能为空' };
  }
  if (comp && !isValidComponent(component)) {
    return {
      ok: false,
      message:
        '组件路径不在当前项目有效视图中，请填写如 views/admin/sys-menu/index 或 Layout',
    };
  }
  return { ok: true };
}

/**
 * 提交编辑：将 fullDetail（完整原始数据）与表单修改字段合并后，
 * 整体送入 PUT /api/v1/menu/:id，确保未编辑字段不被覆盖。
 * icon：若当前展示值与“原始 icon 的展示值”一致，视为未改图标，提交原始值避免污染短 key；否则提交表单值。
 */
async function onEditOk() {
  if (!fullDetail.value) return;
  const v = validateComponentByMenuType(editForm.menuType, editForm.component);
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  const iconUnchanged =
    editForm.icon === (normalizeMenuIcon(fullDetail.value.icon) ?? '');
  const iconToSubmit = iconUnchanged
    ? (fullDetail.value.icon ?? '')
    : (editForm.icon ?? '');
  editSubmitting.value = true;
  try {
    await requestClient.put(`/v1/menu/${fullDetail.value.menuId}`, {
      ...fullDetail.value,
      menuName: editForm.menuName,
      title: editForm.title,
      icon: iconToSubmit,
      path: editForm.path,
      component: editForm.component,
      parentId: editForm.parentId ?? 0,
      menuType: editForm.menuType,
      permission: editForm.permission ?? '',
      apis: editForm.apis ?? [],
      sort: editForm.sort,
      visible: editForm.visible,
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchMenuList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '编辑失败'));
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
  icon: '',
  path: '',
  component: '',
  parentId: 0 as number,
  menuType: 'C' as string,
  permission: '',
  apis: [] as number[],
  sort: 0,
  visible: '1',
});

function onAdd() {
  addForm.menuName = '';
  addForm.title = '';
  addForm.icon = '';
  addForm.path = '';
  addForm.component = '';
  addForm.parentId = 0;
  addForm.menuType = 'C';
  addForm.permission = '';
  addForm.apis = [];
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
      icon: addForm.icon ?? '',
      path: addForm.path,
      component: addForm.component,
      parentId: addForm.parentId ?? 0,
      menuType: addForm.menuType,
      permission: addForm.permission ?? '',
      apis: addForm.apis ?? [],
      sort: addForm.sort,
      visible: addForm.visible,
      isFrame: '0',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchMenuList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '新增失败'));
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
        await requestClient.delete('/v1/menu', {
          data: { ids: [record.menuId] },
        });
        message.success('删除成功');
        fetchMenuList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

/** 将 SysApiItem 转为多选选项：label = title + path + action */
function toApiOption(item: SysApiItem): { value: number; label: string } {
  const parts = [item.title, item.path, item.action].filter(Boolean);
  return {
    value: item.id,
    label: parts.length ? parts.join(' ') : String(item.id),
  };
}

/** 单条接口的展示文案（title / path / action 组合） */
function apiItemLabel(a: SysApiInMenu): string {
  const parts = [a.title, a.path, a.action].filter(Boolean);
  return parts.length ? parts.join(' ') : String(a.id);
}

/**
 * 列表「关联接口」列展示：优先用行数据 sysApi，否则用 apiOptions 按 apis id 映射
 * 无关联时返回 '-'
 */
/** 列表「关联接口」列 Tag 展示：返回标签文案数组，无关联时返回空数组（入参兼容表格 bodyCell 的 record 类型） */
function getApisDisplayLabels(record: unknown): string[] {
  const row = record as SysMenuRow;
  if (row?.sysApi?.length) {
    return row.sysApi.map(apiItemLabel);
  }
  const ids = (record as { apis?: number[] })?.apis;
  if (Array.isArray(ids) && ids.length && apiOptions.value.length) {
    return ids
      .map(
        (id) =>
          apiOptions.value.find((o) => o.value === id)?.label ?? String(id),
      )
      .filter(Boolean);
  }
  return [];
}

onMounted(() => {
  void fetchMenuList();
  getSysApiList()
    .then((list) => {
      apiOptions.value = list.map(toApiOption);
    })
    .catch(() => {
      apiOptions.value = [];
    });
});

const columns: TableColumnType[] = [
  { title: '菜单ID', dataIndex: 'menuId', key: 'menuId', width: 90 },
  {
    title: '菜单名称',
    dataIndex: 'menuName',
    key: 'menuName',
    width: 140,
    customRender: ({ text }) => renderAdminEmpty(text as string),
  },
  {
    title: '显示标题',
    dataIndex: 'title',
    key: 'title',
    width: 120,
    customRender: ({ text }) => renderAdminEmpty(text as string),
  },
  { title: '图标', key: 'icon', width: 70, align: 'center' },
  {
    title: '路径',
    dataIndex: 'path',
    key: 'path',
    width: 180,
    ellipsis: true,
    customRender: ({ text }) => renderAdminEmpty(text as string),
  },
  { title: '类型', dataIndex: 'menuType', key: 'menuType', width: 80 },
  { title: '关联接口', key: 'apisDisplay', width: 200, ellipsis: true },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 70 },
  { title: '可见', dataIndex: 'visible', key: 'visible', width: 70 },
  { title: '操作', key: 'action', width: 190, fixed: 'right' },
];
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>菜单管理</template>
    <template #description>
      管理菜单层级、组件映射和权限标识。树形页沿用统一的页头、筛选区和表格卡片结构。
    </template>
    <template #header-extra>
      <Button @click="fetchMenuList">刷新</Button>
      <AdminActionButton type="primary" codes="admin:sysMenu:add" @click="onAdd">
        新增菜单
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="标题">
          <Input
            v-model:value="query.title"
            placeholder="模糊匹配菜单标题"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="可见">
          <Select
            v-model:value="query.visible"
            :options="visibleOptions"
            placeholder="请选择可见状态"
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
        <div class="text-base font-semibold text-slate-900">菜单树列表</div>
        <p class="mt-1 text-sm text-slate-500">
          重点查看菜单层级、组件路径和接口绑定关系，树表和编辑弹窗的样式统一收口。
        </p>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="treeData"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: 1180 }"
      row-key="menuId"
      size="middle"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'icon'">
          <IconifyIcon
            v-if="normalizeMenuIcon(record.icon)"
            :icon="normalizeMenuIcon(record.icon)!"
            class="text-primary text-lg"
          />
          <span v-else class="text-gray-400">-</span>
        </template>
        <template v-else-if="column.key === 'apisDisplay'">
          <span
            v-if="getApisDisplayLabels(record).length"
            class="flex flex-wrap gap-1"
          >
            <Tag
              v-for="label in getApisDisplayLabels(record)"
              :key="label"
              class="m-0"
            >
              {{ label }}
            </Tag>
          </span>
          <span v-else class="text-gray-400">-</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysMenu:query"
            @click="openDetail(record as SysMenuRow)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysMenu:edit"
            @click="onEdit(record as SysMenuRow)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysMenu:remove"
            @click="onDelete(record as SysMenuRow)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑菜单"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      :width="860"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载中…
      </div>
      <Form v-else layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="菜单名称" class="mb-0">
          <Input v-model:value="editForm.menuName" placeholder="menuName" />
        </FormItem>
        <FormItem label="显示标题" class="mb-0">
          <Input v-model:value="editForm.title" placeholder="title" />
        </FormItem>
        <FormItem label="图标(icon)" class="mb-0 md:col-span-2">
          <IconPicker v-model="editForm.icon" class="w-full" />
        </FormItem>
        <FormItem label="类型(menuType)" class="mb-0">
          <Select
            v-model:value="editForm.menuType"
            :options="menuTypeOptions"
            class="w-full"
            @change="onEditMenuTypeChange"
          />
        </FormItem>
        <FormItem label="路径" class="mb-0">
          <Input v-model:value="editForm.path" placeholder="path" />
        </FormItem>
        <FormItem label="上级菜单" class="mb-0 md:col-span-2">
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
        <FormItem label="组件(component)" class="mb-0 md:col-span-2">
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
        <FormItem label="权限标识(permission)" class="mb-0 md:col-span-2">
          <Input
            v-model:value="editForm.permission"
            placeholder="如 system:user:list，按钮(F)类型常用"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="关联接口(apis)" class="mb-0 md:col-span-2">
          <Select
            v-model:value="editForm.apis"
            mode="multiple"
            :options="apiOptions"
            placeholder="选择关联的接口"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="editForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="可见" class="mb-0">
          <Select
            v-model:value="editForm.visible"
            :options="yesNoOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增菜单"
      :confirm-loading="addSubmitting"
      :width="860"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form layout="vertical" class="mt-4 grid gap-x-4 md:grid-cols-2">
        <FormItem label="菜单名称" class="mb-0">
          <Input v-model:value="addForm.menuName" placeholder="menuName" />
        </FormItem>
        <FormItem label="显示标题" class="mb-0">
          <Input v-model:value="addForm.title" placeholder="title" />
        </FormItem>
        <FormItem label="图标(icon)" class="mb-0 md:col-span-2">
          <IconPicker v-model="addForm.icon" class="w-full" />
        </FormItem>
        <FormItem label="类型(menuType)" class="mb-0">
          <Select
            v-model:value="addForm.menuType"
            :options="menuTypeOptions"
            class="w-full"
            @change="onAddMenuTypeChange"
          />
        </FormItem>
        <FormItem label="路径" class="mb-0">
          <Input v-model:value="addForm.path" placeholder="path" />
        </FormItem>
        <FormItem label="上级菜单" class="mb-0 md:col-span-2">
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
        <FormItem label="组件(component)" class="mb-0 md:col-span-2">
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
        <FormItem label="权限标识(permission)" class="mb-0 md:col-span-2">
          <Input
            v-model:value="addForm.permission"
            placeholder="如 system:user:list，按钮(F)类型常用"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="关联接口(apis)" class="mb-0 md:col-span-2">
          <Select
            v-model:value="addForm.apis"
            mode="multiple"
            :options="apiOptions"
            placeholder="选择关联的接口"
            allow-clear
            class="w-full"
          />
        </FormItem>
        <FormItem label="排序" class="mb-0">
          <InputNumber v-model:value="addForm.sort" :min="0" class="w-full" />
        </FormItem>
        <FormItem label="可见" class="mb-0">
          <Select
            v-model:value="addForm.visible"
            :options="yesNoOptions"
            class="w-full"
          />
        </FormItem>
      </Form>
    </Modal>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="菜单详情"
      :loading="detailLoading"
      width="760"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="基础信息" description="确认菜单名称、类型和显示规则。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs text-slate-500">菜单 ID</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ detailRecord.menuId }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">菜单名称</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.menuName) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">显示标题</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.title) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">菜单类型</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.menuType) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">可见</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ detailRecord.visible === '1' ? '显示' : '隐藏' }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">排序</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.sort) }}</dd>
            </div>
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">图标</dt>
              <dd class="mt-1 flex items-center gap-2 text-sm text-slate-900">
                <IconifyIcon
                  v-if="normalizeMenuIcon(detailRecord.icon)"
                  :icon="normalizeMenuIcon(detailRecord.icon)!"
                  class="text-primary text-lg"
                />
                <span>{{ renderAdminEmpty(detailRecord.icon) }}</span>
              </dd>
            </div>
          </dl>
        </AdminDetailSection>

        <AdminDetailSection title="路由与权限" description="定位路径、组件映射和权限标识。">
          <dl class="grid gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">路径</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.path) }}</dd>
            </div>
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">组件</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.component) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">父级菜单 ID</dt>
              <dd class="mt-1 text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.parentId) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">完整路径</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.paths) }}</dd>
            </div>
            <div class="md:col-span-2">
              <dt class="text-xs text-slate-500">权限标识</dt>
              <dd class="mt-1 break-all text-sm text-slate-900">{{ renderAdminEmpty(detailRecord.permission) }}</dd>
            </div>
          </dl>
        </AdminDetailSection>

        <AdminDetailSection title="关联接口" description="查看菜单与后端接口的绑定关系。">
          <div v-if="detailRecord.sysApi?.length" class="flex flex-wrap gap-2">
            <Tag
              v-for="api in detailRecord.sysApi"
              :key="api.id"
              class="m-0"
            >
              {{ apiItemLabel(api) }}
            </Tag>
          </div>
          <p v-else class="text-sm text-slate-500">暂无关联接口</p>
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
