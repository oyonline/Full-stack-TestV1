<script lang="ts" setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import {
  onBeforeRouteLeave,
  onBeforeRouteUpdate,
  useRoute,
  useRouter,
} from 'vue-router';

import { Fallback } from '@vben/common-ui';
import {
  Alert,
  Button,
  Card,
  Input,
  InputNumber,
  Select,
  Tag,
  Tree,
  message,
} from 'ant-design-vue';
import type { TreeProps } from 'ant-design-vue';

import {
  createRole,
  getRoleDetail,
  getRoleDeptTreeselect,
  getRoleMenuTreeselect,
  getRolePage,
  updateRole,
  updateRoleDataScope,
} from '#/api/core/role';
import type {
  CreateRoleData,
  MenuLabel,
  SysRoleItem,
  UpdateRoleData,
} from '#/api/core/role';
import type { DeptLabel } from '#/api/core/dept';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminPermission } from '#/composables/use-admin-permission';
import { resolveAdminErrorMessage } from '#/utils/admin-crud';
import {
  CUSTOM_ROLE_DATA_SCOPE,
  DEFAULT_ROLE_DATA_SCOPE,
  ROLE_DATA_SCOPE_OPTIONS,
  normalizeRoleDataScope,
} from './shared';

type TreeKey = string | number;
type TreeNode = NonNullable<TreeProps['treeData']>[number];

const route = useRoute();
const router = useRouter();

const createPermission = useAdminPermission({ codes: 'admin:sysRole:add' });
const editPermission = useAdminPermission({ codes: 'admin:sysRole:update' });

const isCreateMode = computed(() => route.name === 'SysRoleCreate');
const pageTitle = computed(() => (isCreateMode.value ? '新增角色' : '编辑角色'));
const hasPagePermission = computed(() =>
  isCreateMode.value
    ? createPermission.hasPermission.value
    : editPermission.hasPermission.value,
);
const roleId = computed(() => {
  const raw = route.query.roleId;
  const value = Array.isArray(raw) ? raw[0] : raw;
  const normalized = Number(value ?? 0);
  return Number.isFinite(normalized) ? normalized : 0;
});

const statusOptions = [
  { label: '停用', value: '1' },
  { label: '启用', value: '2' },
];

const loading = ref(false);
const saving = ref(false);
const errorMsg = ref('');
const suppressLeaveGuard = ref(false);

const form = reactive({
  dataScope: DEFAULT_ROLE_DATA_SCOPE,
  remark: '',
  roleKey: '',
  roleName: '',
  roleSort: 0,
  status: '2',
});

const menuKeyword = ref('');
const rawMenuTree = ref<TreeNode[]>([]);
const rawDeptTree = ref<TreeNode[]>([]);
const menuCheckedKeys = ref<TreeKey[]>([]);
const menuHalfCheckedKeys = ref<TreeKey[]>([]);
const deptCheckedKeys = ref<TreeKey[]>([]);
const deptHalfCheckedKeys = ref<TreeKey[]>([]);
const menuExpandedKeys = ref<TreeKey[]>([]);
const deptExpandedKeys = ref<TreeKey[]>([]);
const menuAutoExpandParent = ref(false);
const deptAutoExpandParent = ref(false);
const initialSnapshot = ref('');

function labelOf(node: TreeNode) {
  return String(node.title ?? '').trim();
}

function menuLabelToTreeData(nodes: MenuLabel[]): TreeNode[] {
  return (nodes || []).map((node) => ({
    key: node.id,
    title: node.label || '',
    children: node.children?.length
      ? menuLabelToTreeData(node.children)
      : undefined,
  }));
}

function deptLabelToTreeData(nodes: DeptLabel[]): TreeNode[] {
  return (nodes || []).map((node) => ({
    key: node.id,
    title: node.label || '',
    children: node.children?.length
      ? deptLabelToTreeData(node.children)
      : undefined,
  }));
}

function collectAllKeys(nodes: TreeNode[]): TreeKey[] {
  const keys: TreeKey[] = [];
  const walk = (items: TreeNode[]) => {
    for (const item of items) {
      keys.push(item.key);
      if (item.children?.length) {
        walk(item.children as TreeNode[]);
      }
    }
  };
  walk(nodes);
  return keys;
}

function collectRootKeys(nodes: TreeNode[]): TreeKey[] {
  return nodes.map((node) => node.key);
}

function normalizeKeys(keys: TreeKey[]) {
  return [...new Set(keys.map((key) => Number(key)).filter(Number.isFinite))].sort(
    (left, right) => left - right,
  );
}

function filterTree(nodes: TreeNode[], keyword: string) {
  const needle = keyword.trim().toLowerCase();
  if (!needle) {
    return {
      expandedKeys: [] as TreeKey[],
      nodes,
    };
  }

  const expanded = new Set<TreeKey>();

  const walk = (items: TreeNode[]): TreeNode[] => {
    return items
      .map((item) => {
        const children = item.children?.length
          ? walk(item.children as TreeNode[])
          : [];
        const matched = labelOf(item).toLowerCase().includes(needle);
        if (!matched && children.length === 0) {
          return null;
        }
        if (children.length > 0) {
          expanded.add(item.key);
        }
        return {
          ...item,
          children: children.length > 0 ? children : undefined,
        } as TreeNode;
      })
      .filter(Boolean) as TreeNode[];
  };

  const tree = walk(nodes);
  return {
    expandedKeys: [...expanded],
    nodes: tree,
  };
}

const filteredMenuTree = computed(() =>
  filterTree(rawMenuTree.value, menuKeyword.value),
);

const visibleMenuExpandedKeys = computed(() =>
  menuKeyword.value.trim()
    ? filteredMenuTree.value.expandedKeys
    : menuExpandedKeys.value,
);

const isCustomDataScope = computed(
  () => normalizeRoleDataScope(form.dataScope) === CUSTOM_ROLE_DATA_SCOPE,
);

const selectedMenuCount = computed(
  () => normalizeKeys([...menuCheckedKeys.value, ...menuHalfCheckedKeys.value]).length,
);
const selectedDeptCount = computed(
  () => normalizeKeys([...deptCheckedKeys.value, ...deptHalfCheckedKeys.value]).length,
);

const snapshot = computed(() =>
  JSON.stringify({
    dataScope: normalizeRoleDataScope(form.dataScope),
    deptIds: normalizeKeys([...deptCheckedKeys.value, ...deptHalfCheckedKeys.value]),
    menuIds: normalizeKeys([...menuCheckedKeys.value, ...menuHalfCheckedKeys.value]),
    remark: form.remark.trim(),
    roleKey: form.roleKey.trim(),
    roleName: form.roleName.trim(),
    roleSort: Number(form.roleSort ?? 0),
    status: form.status,
  }),
);

const isDirty = computed(() => initialSnapshot.value !== snapshot.value);

function markSnapshot() {
  initialSnapshot.value = snapshot.value;
}

function resetForm() {
  form.dataScope = DEFAULT_ROLE_DATA_SCOPE;
  form.remark = '';
  form.roleKey = '';
  form.roleName = '';
  form.roleSort = 0;
  form.status = '2';
  menuKeyword.value = '';
  menuCheckedKeys.value = [];
  menuHalfCheckedKeys.value = [];
  deptCheckedKeys.value = [];
  deptHalfCheckedKeys.value = [];
}

function applyRoleDetail(detail: SysRoleItem) {
  form.roleName = detail.roleName ?? '';
  form.roleKey = detail.roleKey ?? '';
  form.roleSort = detail.roleSort ?? 0;
  form.status = detail.status ?? '2';
  form.remark = detail.remark ?? '';
  form.dataScope = normalizeRoleDataScope(detail.dataScope);
}

async function resolveCreatedRoleId(
  roleKey: string,
  roleName: string,
): Promise<number> {
  const result = await getRolePage({
    pageIndex: 1,
    pageSize: 10,
    roleKey,
    roleName,
  });
  return (
    result.list.find((item) => item.roleKey === roleKey)?.roleId ??
    result.list[0]?.roleId ??
    0
  );
}

async function loadWorkspace() {
  if (!hasPagePermission.value) {
    return;
  }

  loading.value = true;
  errorMsg.value = '';

  try {
    if (isCreateMode.value) {
      resetForm();
      const [menuRes, deptRes] = await Promise.all([
        getRoleMenuTreeselect(0),
        getRoleDeptTreeselect(0),
      ]);
      rawMenuTree.value = menuLabelToTreeData(menuRes?.menus || []);
      rawDeptTree.value = deptLabelToTreeData(deptRes?.depts || []);
      menuExpandedKeys.value = collectRootKeys(rawMenuTree.value);
      deptExpandedKeys.value = collectAllKeys(rawDeptTree.value);
      markSnapshot();
      return;
    }

    if (!roleId.value) {
      rawMenuTree.value = [];
      rawDeptTree.value = [];
      errorMsg.value = '缺少 roleId 参数';
      return;
    }

    const [detail, menuRes, deptRes] = await Promise.all([
      getRoleDetail(roleId.value),
      getRoleMenuTreeselect(roleId.value),
      getRoleDeptTreeselect(roleId.value),
    ]);

    applyRoleDetail(detail);
    rawMenuTree.value = menuLabelToTreeData(menuRes?.menus || []);
    rawDeptTree.value = deptLabelToTreeData(deptRes?.depts || []);
    menuCheckedKeys.value = menuRes?.checkedKeys || [];
    menuHalfCheckedKeys.value = [];
    deptCheckedKeys.value = deptRes?.checkedKeys || [];
    deptHalfCheckedKeys.value = [];
    menuExpandedKeys.value = collectRootKeys(rawMenuTree.value);
    deptExpandedKeys.value = collectAllKeys(rawDeptTree.value);
    markSnapshot();
  } catch (error) {
    errorMsg.value = resolveAdminErrorMessage(error, '加载角色配置失败');
  } finally {
    loading.value = false;
  }
}

function validateForm() {
  const roleName = form.roleName.trim();
  const roleKey = form.roleKey.trim();
  if (!roleName) return '请输入角色名称';
  if (!roleKey) return '请输入权限字符';
  return '';
}

async function persistDataScope(targetRoleId: number, dataScope: string) {
  const deptIds = isCustomDataScope.value
    ? normalizeKeys([...deptCheckedKeys.value, ...deptHalfCheckedKeys.value])
    : [];

  await updateRoleDataScope({
    dataScope,
    deptIds,
    roleId: targetRoleId,
  });
}

async function saveWorkspace() {
  if (!hasPagePermission.value) {
    return;
  }

  const validationMessage = validateForm();
  if (validationMessage) {
    message.error(validationMessage);
    return;
  }

  saving.value = true;
  const dataScope = normalizeRoleDataScope(form.dataScope);
  const payload: CreateRoleData & UpdateRoleData = {
    dataScope,
    menuIds: normalizeKeys([...menuCheckedKeys.value, ...menuHalfCheckedKeys.value]),
    remark: form.remark.trim() || undefined,
    roleKey: form.roleKey.trim(),
    roleName: form.roleName.trim(),
    roleSort: Number(form.roleSort ?? 0),
    status: form.status,
  };

  try {
    if (isCreateMode.value) {
      const created = await createRole(payload);
      let targetRoleId = Number(created ?? 0);
      if (!targetRoleId) {
        targetRoleId = await resolveCreatedRoleId(payload.roleKey, payload.roleName);
      }
      if (!targetRoleId) {
        throw new Error('角色已创建，但未能定位新角色 ID');
      }

      let scopeError: unknown = null;
      try {
        await persistDataScope(targetRoleId, dataScope);
      } catch (error) {
        scopeError = error;
      }

      suppressLeaveGuard.value = true;
      await router.replace({
        path: '/admin/sys-role/edit',
        query: { roleId: String(targetRoleId) },
      });
      suppressLeaveGuard.value = false;

      if (scopeError) {
        message.error(
          resolveAdminErrorMessage(
            scopeError,
            '角色已创建，但数据权限保存失败，请继续在编辑页检查',
          ),
        );
      } else {
        message.success('新增成功');
      }
      return;
    }

    if (!roleId.value) {
      throw new Error('缺少 roleId 参数');
    }

    await updateRole(roleId.value, payload);

    try {
      await persistDataScope(roleId.value, dataScope);
    } catch (error) {
      await loadWorkspace();
      message.error(
        resolveAdminErrorMessage(
          error,
          '角色基础信息已保存，但数据权限同步失败',
        ),
      );
      return;
    }

    await loadWorkspace();
    message.success('保存成功');
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '保存失败'));
  } finally {
    saving.value = false;
  }
}

function confirmDiscardChanges() {
  return window.confirm('当前角色配置尚未保存，确认离开吗？');
}

function goBack() {
  void router.push('/admin/sys-role');
}

function clearMenuSelection() {
  menuCheckedKeys.value = [];
  menuHalfCheckedKeys.value = [];
}

function clearDeptSelection() {
  deptCheckedKeys.value = [];
  deptHalfCheckedKeys.value = [];
}

function expandAllMenus() {
  menuExpandedKeys.value = collectAllKeys(rawMenuTree.value);
  menuAutoExpandParent.value = false;
}

function collapseAllMenus() {
  menuExpandedKeys.value = [];
  menuAutoExpandParent.value = false;
}

function expandAllDepts() {
  deptExpandedKeys.value = collectAllKeys(rawDeptTree.value);
  deptAutoExpandParent.value = false;
}

function collapseAllDepts() {
  deptExpandedKeys.value = [];
  deptAutoExpandParent.value = false;
}

const onMenuCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  menuCheckedKeys.value = checkedKeys as TreeKey[];
  menuHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

const onDeptCheck: TreeProps['onCheck'] = (checkedKeys, info) => {
  deptCheckedKeys.value = checkedKeys as TreeKey[];
  deptHalfCheckedKeys.value = (info?.halfCheckedKeys as TreeKey[]) || [];
};

const onMenuExpand: TreeProps['onExpand'] = (expandedKeys) => {
  menuExpandedKeys.value = expandedKeys as TreeKey[];
  menuAutoExpandParent.value = false;
};

const onDeptExpand: TreeProps['onExpand'] = (expandedKeys) => {
  deptExpandedKeys.value = expandedKeys as TreeKey[];
  deptAutoExpandParent.value = false;
};

function handleBeforeUnload(event: BeforeUnloadEvent) {
  if (!isDirty.value || suppressLeaveGuard.value) {
    return;
  }
  event.preventDefault();
  event.returnValue = '';
}

watch(
  () => [route.name, route.query.roleId, hasPagePermission.value],
  async () => {
    await loadWorkspace();
  },
  { immediate: true },
);

watch(
  () => menuKeyword.value,
  () => {
    menuAutoExpandParent.value = true;
  },
);

watch(
  () => isCustomDataScope.value,
  (enabled) => {
    if (enabled && deptExpandedKeys.value.length === 0) {
      deptExpandedKeys.value = collectAllKeys(rawDeptTree.value);
      deptAutoExpandParent.value = false;
    }
  },
);

window.addEventListener('beforeunload', handleBeforeUnload);

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload);
});

onBeforeRouteLeave(() => {
  if (suppressLeaveGuard.value || !isDirty.value) {
    return true;
  }
  return confirmDiscardChanges();
});

onBeforeRouteUpdate(() => {
  if (suppressLeaveGuard.value || !isDirty.value) {
    return true;
  }
  return confirmDiscardChanges();
});
</script>

<template>
  <Fallback v-if="!hasPagePermission" status="403" />

  <AdminPageShell v-else header-mode="compact">
    <template #eyebrow>System Admin</template>
    <template #title>{{ pageTitle }}</template>
    <template #description>
      {{
        isCreateMode
          ? '创建角色并集中配置菜单权限与数据权限。当前工作区会保留更大的树视图和清晰的保存节奏。'
          : '维护角色基础信息、菜单权限和数据权限。保存后会停留当前标签页，便于连续调整。'
      }}
    </template>

    <template #header-extra>
      <Button @click="goBack">返回列表</Button>
      <Button type="primary" :loading="saving" @click="saveWorkspace">
        保存
      </Button>
    </template>

    <Alert
      v-if="errorMsg"
      show-icon
      type="error"
      :message="errorMsg"
      class="mb-4"
    />

    <div v-if="!errorMsg" class="space-y-4">
      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>
          <div class="flex flex-wrap items-center gap-2">
            <span>基础信息</span>
            <Tag v-if="!isCreateMode && roleId" color="blue">
              角色ID {{ roleId }}
            </Tag>
          </div>
        </template>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <AdminFilterField label="角色名称">
            <Input v-model:value="form.roleName" placeholder="请输入角色名称" />
          </AdminFilterField>
          <AdminFilterField label="权限字符">
            <Input v-model:value="form.roleKey" placeholder="请输入权限字符" />
          </AdminFilterField>
          <AdminFilterField label="排序">
            <InputNumber
              v-model:value="form.roleSort"
              :min="0"
              class="w-full"
              placeholder="请输入排序"
            />
          </AdminFilterField>
          <AdminFilterField label="状态">
            <Select
              v-model:value="form.status"
              :options="statusOptions"
              placeholder="请选择状态"
            />
          </AdminFilterField>
          <AdminFilterField label="数据范围">
            <Select
              v-model:value="form.dataScope"
              :options="ROLE_DATA_SCOPE_OPTIONS"
              placeholder="请选择数据范围"
            />
          </AdminFilterField>
          <div class="space-y-2 md:col-span-2 xl:col-span-3">
            <label class="block text-sm font-medium text-slate-700">备注</label>
            <Input.TextArea
              v-model:value="form.remark"
              :auto-size="{ minRows: 2, maxRows: 4 }"
              placeholder="请输入备注"
            />
          </div>
        </div>
      </Card>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1.65fr)_minmax(320px,0.95fr)]">
        <Card :bordered="false" class="app-radius-panel shadow-sm">
          <template #title>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <span>菜单权限</span>
                <Tag color="processing">已选 {{ selectedMenuCount }}</Tag>
                <Tag color="default">总计 {{ collectAllKeys(rawMenuTree).length }}</Tag>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <Button size="small" @click="expandAllMenus">展开全部</Button>
                <Button size="small" @click="collapseAllMenus">收起全部</Button>
                <Button size="small" @click="clearMenuSelection">清空选择</Button>
              </div>
            </div>
          </template>

          <div class="space-y-4">
            <Input
              v-model:value="menuKeyword"
              allow-clear
              placeholder="搜索菜单权限"
            />

            <div
              class="app-radius-box h-[620px] overflow-auto border border-slate-200 bg-slate-50 p-3"
            >
              <Tree
                v-if="filteredMenuTree.nodes.length"
                block-node
                checkable
                :auto-expand-parent="menuKeyword.trim() ? true : menuAutoExpandParent"
                :checked-keys="menuCheckedKeys"
                :expanded-keys="visibleMenuExpandedKeys"
                :tree-data="filteredMenuTree.nodes"
                @check="onMenuCheck"
                @expand="onMenuExpand"
              />
              <div v-else class="py-8 text-center text-sm text-slate-400">
                {{
                  menuKeyword.trim()
                    ? '没有匹配的菜单权限'
                    : loading
                      ? '加载菜单权限中…'
                      : '暂无菜单权限数据'
                }}
              </div>
            </div>
          </div>
        </Card>

        <Card :bordered="false" class="app-radius-panel shadow-sm">
          <template #title>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <span>数据权限（部门）</span>
                <Tag :color="isCustomDataScope ? 'processing' : 'default'">
                  {{ isCustomDataScope ? `已选 ${selectedDeptCount}` : '按数据范围控制' }}
                </Tag>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <Button size="small" @click="expandAllDepts">展开全部</Button>
                <Button size="small" @click="collapseAllDepts">收起全部</Button>
                <Button size="small" @click="clearDeptSelection">清空选择</Button>
              </div>
            </div>
          </template>

          <div class="space-y-3">
            <Alert
              v-if="!isCustomDataScope"
              show-icon
              type="info"
              :message="'当前数据范围不是“自定义数据权限”，部门树仅展示结构，不参与保存。'"
            />

            <div
              class="app-radius-box h-[620px] overflow-auto border border-slate-200 bg-slate-50 p-3"
              :class="{ 'pointer-events-none opacity-60': !isCustomDataScope }"
            >
              <Tree
                v-if="rawDeptTree.length"
                block-node
                checkable
                :auto-expand-parent="deptAutoExpandParent"
                :checked-keys="deptCheckedKeys"
                :disabled="!isCustomDataScope"
                :expanded-keys="deptExpandedKeys"
                :tree-data="rawDeptTree"
                @check="onDeptCheck"
                @expand="onDeptExpand"
              />
              <div v-else class="py-8 text-center text-sm text-slate-400">
                {{ loading ? '加载部门权限中…' : '暂无部门树数据' }}
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  </AdminPageShell>
</template>
