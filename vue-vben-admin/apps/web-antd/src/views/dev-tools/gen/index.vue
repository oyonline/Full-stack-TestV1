<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  Alert,
  Button,
  Input,
  Popconfirm,
  Table,
  Tabs,
  Tag,
  message,
} from 'ant-design-vue';
import type { Key } from 'ant-design-vue/es/_util/type';
import type { TableColumnType } from 'ant-design-vue';

import {
  createSysTables,
  deleteSysTable,
  getDbTablePage,
  getSysTablePage,
  previewGeneratedCode,
} from '#/api/core';
import type {
  DbTableItem,
  PageResult,
  PreviewResult,
  SysTableItem,
} from '#/api/core';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import GenPreviewModal from '#/views/dev-tools/gen/components/preview-modal.vue';

type GeneratorTabKey = 'db' | 'sys';

const router = useRouter();

const activeTab = ref<GeneratorTabKey>('db');

const dbLoading = ref(false);
const dbErrorMsg = ref('');
const dbTableData = ref<DbTableItem[]>([]);
const dbTableName = ref('');
const dbPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

const sysLoading = ref(false);
const sysErrorMsg = ref('');
const sysTableData = ref<SysTableItem[]>([]);
const sysTableName = ref('');
const sysTableComment = ref('');
const sysPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

const previewVisible = ref(false);
const previewLoading = ref(false);
const previewTitle = ref('');
const previewData = ref<PreviewResult>({});

async function fetchDbTables() {
  dbLoading.value = true;
  dbErrorMsg.value = '';
  try {
    const result: PageResult<DbTableItem> = await getDbTablePage({
      pageIndex: dbPagination.value.current,
      pageSize: dbPagination.value.pageSize,
      tableName: dbTableName.value.trim() || undefined,
    });
    dbTableData.value = result.list ?? [];
    dbPagination.value.total = result.count ?? 0;
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    dbErrorMsg.value = err?.message || err?.response?.data?.msg || '加载数据库表失败';
    dbTableData.value = [];
    dbPagination.value.total = 0;
  } finally {
    dbLoading.value = false;
  }
}

async function fetchSysTables() {
  sysLoading.value = true;
  sysErrorMsg.value = '';
  try {
    const result: PageResult<SysTableItem> = await getSysTablePage({
      pageIndex: sysPagination.value.current,
      pageSize: sysPagination.value.pageSize,
      tableName: sysTableName.value.trim() || undefined,
      tableComment: sysTableComment.value.trim() || undefined,
    });
    sysTableData.value = result.list ?? [];
    sysPagination.value.total = result.count ?? 0;
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    sysErrorMsg.value = err?.message || err?.response?.data?.msg || '加载生成配置失败';
    sysTableData.value = [];
    sysPagination.value.total = 0;
  } finally {
    sysLoading.value = false;
  }
}

function onSearch() {
  if (activeTab.value === 'db') {
    dbPagination.value.current = 1;
    fetchDbTables();
    return;
  }
  sysPagination.value.current = 1;
  fetchSysTables();
}

function onReset() {
  if (activeTab.value === 'db') {
    dbTableName.value = '';
    dbPagination.value.current = 1;
    fetchDbTables();
    return;
  }
  sysTableName.value = '';
  sysTableComment.value = '';
  sysPagination.value.current = 1;
  fetchSysTables();
}

function onDbTableChange(pag: { current?: number; pageSize?: number }) {
  if (pag.current) dbPagination.value.current = pag.current;
  if (pag.pageSize) dbPagination.value.pageSize = pag.pageSize;
  fetchDbTables();
}

function onSysTableChange(pag: { current?: number; pageSize?: number }) {
  if (pag.current) sysPagination.value.current = pag.current;
  if (pag.pageSize) sysPagination.value.pageSize = pag.pageSize;
  fetchSysTables();
}

function onTabChange(key: Key) {
  activeTab.value = String(key) as GeneratorTabKey;
  if (activeTab.value === 'db' && !dbTableData.value.length && !dbLoading.value) {
    fetchDbTables();
  }
  if (activeTab.value === 'sys' && !sysTableData.value.length && !sysLoading.value) {
    fetchSysTables();
  }
}

async function adoptDbTable(record: DbTableItem) {
  try {
    await createSysTables([record.tableName]);
    message.success('已加入代码生成配置');
    await Promise.all([fetchDbTables(), fetchSysTables()]);
    activeTab.value = 'sys';
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '纳入失败');
  }
}

function openEdit(record: SysTableItem) {
  router.push({
    path: '/dev-tools/editTable',
    query: { tableId: String(record.tableId) },
  });
}

async function openPreview(record: SysTableItem) {
  previewVisible.value = true;
  previewLoading.value = true;
  previewTitle.value = `${record.tableComment || record.tableName || '代码生成'} - 模板预览`;
  previewData.value = {};
  try {
    previewData.value = await previewGeneratedCode(record.tableId);
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '获取预览失败');
    previewVisible.value = false;
  } finally {
    previewLoading.value = false;
  }
}

async function removeSysTable(record: SysTableItem) {
  try {
    await deleteSysTable(record.tableId);
    message.success('已从代码生成器移除');

    if (sysTableData.value.length <= 1 && sysPagination.value.current > 1) {
      sysPagination.value.current -= 1;
    }

    await Promise.all([fetchDbTables(), fetchSysTables()]);
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '移除失败');
  }
}

const dbColumns: TableColumnType[] = [
  { title: '表名', dataIndex: 'tableName', key: 'tableName', width: 180 },
  {
    title: '备注',
    dataIndex: 'tableComment',
    key: 'tableComment',
    width: 220,
    ellipsis: true,
  },
  { title: '引擎', dataIndex: 'engine', key: 'engine', width: 100 },
  { title: '行数', dataIndex: 'tableRows', key: 'tableRows', width: 90 },
  { title: '创建时间', dataIndex: 'createTime', key: 'createTime', width: 170 },
  { title: '更新时间', dataIndex: 'updateTime', key: 'updateTime', width: 170 },
  { title: '操作', key: 'actions', width: 120, fixed: 'right' },
];

const sysColumns: TableColumnType[] = [
  { title: '表名', dataIndex: 'tableName', key: 'tableName', width: 180 },
  {
    title: '功能名',
    dataIndex: 'functionName',
    key: 'functionName',
    width: 180,
    ellipsis: true,
  },
  {
    title: '类名',
    dataIndex: 'className',
    key: 'className',
    width: 160,
    ellipsis: true,
  },
  {
    title: '业务名',
    dataIndex: 'businessName',
    key: 'businessName',
    width: 140,
    ellipsis: true,
  },
  {
    title: '模板',
    dataIndex: 'tplCategory',
    key: 'tplCategory',
    width: 100,
  },
  {
    title: '备注',
    dataIndex: 'tableComment',
    key: 'tableComment',
    width: 220,
    ellipsis: true,
  },
  { title: '操作', key: 'actions', width: 180, fixed: 'right' },
];

onMounted(() => {
  fetchDbTables();
  fetchSysTables();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>DEV TOOLS</template>
    <template #title>代码生成</template>
    <template #description>
      当前页承接代码生成主流程：导入数据库表、维护已纳管配置，并通过预览验证模板输出。本轮先隐藏真正的写入动作。
    </template>

    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <AdminFilterField v-if="activeTab === 'db'" label="数据库表名">
          <Input
            v-model:value="dbTableName"
            allow-clear
            placeholder="请输入数据库表名"
            @pressEnter="onSearch"
          />
        </AdminFilterField>
        <template v-else>
          <AdminFilterField label="生成表名">
            <Input
              v-model:value="sysTableName"
              allow-clear
              placeholder="请输入生成表名"
              @pressEnter="onSearch"
            />
          </AdminFilterField>
          <AdminFilterField label="功能备注">
            <Input
              v-model:value="sysTableComment"
              allow-clear
              placeholder="请输入表备注"
              @pressEnter="onSearch"
            />
          </AdminFilterField>
        </template>
      </div>
    </template>

    <template #filter-actions>
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </template>

    <template #toolbar>
      <div class="flex flex-wrap items-center gap-2">
        <Tag color="blue">{{ activeTab === 'db' ? '数据库表' : '生成配置' }}</Tag>
        <span class="text-sm text-slate-500">
          {{ activeTab === 'db' ? '先把数据库表纳入生成器，之后进入独立编辑页维护详细配置。' : '当前仅开放配置维护与模板预览，危险写入动作已隐藏。' }}
        </span>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <Button @click="fetchDbTables">刷新数据库表</Button>
        <Button @click="fetchSysTables">刷新生成配置</Button>
      </div>
    </template>

    <Tabs v-model:activeKey="activeTab" @change="onTabChange">
      <Tabs.TabPane key="db" tab="数据库表">
        <Alert
          show-icon
          type="info"
          message="数据库表列表会自动排除已经纳入 sys_tables 的记录。"
          class="mb-4"
        />
        <Alert
          v-if="dbErrorMsg"
          show-icon
          type="error"
          :message="dbErrorMsg"
          class="mb-4"
        />
        <Table
          row-key="tableName"
          :columns="dbColumns"
          :data-source="dbTableData"
          :loading="dbLoading"
          :pagination="dbPagination"
          :scroll="{ x: 980 }"
          size="middle"
          @change="onDbTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'actions'">
              <Button type="link" class="px-0" @click="adoptDbTable(record as DbTableItem)">
                导入到生成器
              </Button>
            </template>
          </template>
        </Table>
      </Tabs.TabPane>

      <Tabs.TabPane key="sys" tab="生成配置">
        <Alert
          show-icon
          type="info"
          message="已纳管表支持进入独立编辑页维护表配置与字段配置，并通过模板预览验证生成结果。"
          class="mb-4"
        />
        <Alert
          v-if="sysErrorMsg"
          show-icon
          type="error"
          :message="sysErrorMsg"
          class="mb-4"
        />
        <Table
          row-key="tableId"
          :columns="sysColumns"
          :data-source="sysTableData"
          :loading="sysLoading"
          :pagination="sysPagination"
          :scroll="{ x: 1180 }"
          size="middle"
          @change="onSysTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'tplCategory'">
              <Tag color="purple">{{ (record as SysTableItem).tplCategory || '-' }}</Tag>
            </template>
            <template v-if="column.key === 'actions'">
              <Button type="link" class="px-0" @click="openEdit(record as SysTableItem)">
                编辑配置
              </Button>
              <Button type="link" class="px-2" @click="openPreview(record as SysTableItem)">
                模板预览
              </Button>
              <Popconfirm
                title="确认将这张表从代码生成器中移除吗？"
                ok-text="移除"
                cancel-text="取消"
                @confirm="removeSysTable(record as SysTableItem)"
              >
                <Button danger type="link" class="px-0">移除</Button>
              </Popconfirm>
            </template>
          </template>
        </Table>
      </Tabs.TabPane>
    </Tabs>

    <GenPreviewModal
      v-model:open="previewVisible"
      :loading="previewLoading"
      :preview-data="previewData"
      :title="previewTitle"
    />
  </AdminPageShell>
</template>
