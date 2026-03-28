<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button, Input, Modal, Select, Table, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  deleteRole,
  getRoleDetail,
  getRolePage,
  updateRoleStatus,
} from '#/api/core/role';
import type { SysRoleItem } from '#/api/core/role';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import AdminTableColumnSettings from '#/components/admin/table-column-settings.vue';
import { useAdminTableColumns } from '#/composables/use-admin-table-columns';
import { useAdminTable } from '#/composables/use-admin-table';
import {
  formatAdminDateTime,
  renderAdminEmpty,
  resolveAdminErrorMessage,
} from '#/utils/admin-crud';
import { getRoleDataScopeLabel } from './shared';

const router = useRouter();

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

function renderStatus(status?: string) {
  if (status === '2') return '启用';
  if (status === '1') return '停用';
  return status || '-';
}

function renderDataScope(dataScope?: string) {
  return renderAdminEmpty(getRoleDataScopeLabel(dataScope));
}

const baseColumns: TableColumnType[] = [
  { title: '角色ID', dataIndex: 'roleId', key: 'roleId', width: 90 },
  { title: '角色名称', dataIndex: 'roleName', key: 'roleName', width: 140 },
  { title: '权限字符', dataIndex: 'roleKey', key: 'roleKey', width: 140 },
  { title: '排序', dataIndex: 'roleSort', key: 'roleSort', width: 90 },
  {
    title: '数据范围',
    dataIndex: 'dataScope',
    key: 'dataScope',
    width: 140,
    customRender: ({ text }: { text: string }) => renderDataScope(text),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 90,
    customRender: ({ text }: { text: string }) => renderStatus(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
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
  tableId: 'sys-role-list',
});

function openCreateWorkspace() {
  void router.push('/admin/sys-role/create');
}

function openEditWorkspace(record: SysRoleItem) {
  void router.push({
    path: '/admin/sys-role/edit',
    query: { roleId: String(record.roleId) },
  });
}

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
        await fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

async function onToggleStatus(record: SysRoleItem) {
  const nextStatus = record.status === '2' ? '1' : '2';
  const label = nextStatus === '2' ? '启用' : '停用';
  try {
    await updateRoleStatus(record.roleId, nextStatus);
    message.success(`${label}成功`);
    await fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, `${label}失败`));
  }
}

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
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>角色管理</template>
    <template #description>
      管理角色名称、权限字符和授权范围。新增与编辑已切换为独立工作区，便于集中配置菜单权限与数据权限。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:sysRole:add"
        @click="openCreateWorkspace"
      >
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
      :row-key="(record: SysRoleItem) => record.roleId"
      :scroll="{ x: scrollX }"
      size="middle"
      @change="onTableChange"
      @resizeColumn="handleResizeColumn"
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
            @click="openEditWorkspace(record as SysRoleItem)"
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
              <dd class="mt-1 text-sm text-slate-900">
                {{ renderAdminEmpty(detailRecord.roleName) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">权限字符</dt>
              <dd class="mt-1 text-sm text-slate-900">
                {{ renderAdminEmpty(detailRecord.roleKey) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">排序</dt>
              <dd class="mt-1 text-sm text-slate-900">
                {{ renderAdminEmpty(detailRecord.roleSort) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">状态</dt>
              <dd class="mt-1 text-sm text-slate-900">
                {{ renderStatus(detailRecord.status) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">创建时间</dt>
              <dd class="mt-1 text-sm text-slate-900">
                {{ formatAdminDateTime(detailRecord.createdAt) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-slate-500">数据范围</dt>
              <dd class="mt-1 text-sm text-slate-900">
                {{ renderDataScope(detailRecord.dataScope) }}
              </dd>
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
