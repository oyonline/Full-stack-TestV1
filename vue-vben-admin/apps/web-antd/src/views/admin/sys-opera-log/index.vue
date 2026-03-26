<script lang="ts" setup>
/**
 * 系统管理 - 操作日志
 * 列表 + 搜索 + 分页 + 删除 + 详情
 */
import { h, onMounted, ref } from 'vue';

import {
  Button,
  Descriptions,
  Input,
  Modal,
  Select,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import { deleteOperaLog, getOperaLogDetail, getOperaLogPage } from '#/api/core';
import type { SysOperaLogItem, SysOperaLogPageResult } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminDetailCodeBlock from '#/components/admin/detail-code-block.vue';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

const businessTypeLabels: Record<string, string> = {
  create: '新增',
  delete: '删除',
  password: '密码',
  run: '执行',
  start: '启动',
  status: '状态',
  stop: '停止',
  update: '更新',
  DELETE: '删除',
  INSERT: '新增',
  UPDATE: '更新',
};

const businessTypeColors: Record<string, string> = {
  create: 'green',
  delete: 'red',
  password: 'gold',
  run: 'purple',
  start: 'cyan',
  status: 'orange',
  stop: 'volcano',
  update: 'blue',
  DELETE: 'red',
  INSERT: 'green',
  UPDATE: 'blue',
};

const businessTypeOptions = [
  { label: '新增', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '状态', value: 'status' },
  { label: '密码', value: 'password' },
  { label: '启动', value: 'start' },
  { label: '停止', value: 'stop' },
];

const businessTypesLabels: Record<string, string> = {
  api: '接口',
  dept: '部门',
  'dict-data': '字典数据',
  'dict-type': '字典类型',
  generator: '代码生成',
  job: '定时任务',
  menu: '菜单',
  post: '岗位',
  role: '角色',
  'system-settings': '系统设置',
  user: '用户',
  'generator-import': '代码生成',
  'generator-remove': '代码生成',
  'generator-save': '代码生成',
  'menu-create': '菜单',
  'menu-delete': '菜单',
  'menu-update': '菜单',
  'role-create': '角色',
  'role-data-scope': '角色',
  'role-delete': '角色',
  'role-status': '角色',
  'role-update': '角色',
};

const businessTypesOptions = [
  { label: '系统设置', value: 'system-settings' },
  { label: '代码生成', value: 'generator' },
  { label: '角色', value: 'role' },
  { label: '菜单', value: 'menu' },
  { label: '用户', value: 'user' },
  { label: '部门', value: 'dept' },
  { label: '岗位', value: 'post' },
  { label: '接口', value: 'api' },
  { label: '字典类型', value: 'dict-type' },
  { label: '字典数据', value: 'dict-data' },
  { label: '定时任务', value: 'job' },
];

const operatorTypeOptions = [{ label: '后台管理', value: 'MANAGE' }];

const requestMethodOptions = ['GET', 'POST', 'PUT', 'DELETE'].map((value) => ({
  label: value,
  value,
}));

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '异常', value: 2 },
];

function getBusinessTypeLabel(value?: string) {
  return businessTypeLabels[value || ''] || renderEmpty(value);
}

function getBusinessTypeColor(value?: string) {
  return businessTypeColors[value || ''] || 'default';
}

function getBusinessTypesLabel(value?: string) {
  return businessTypesLabels[value || ''] || renderEmpty(value);
}

function getOperatorTypeLabel(value?: string) {
  if (value === 'MANAGE') {
    return '后台管理';
  }
  return renderEmpty(value);
}

function getStatusLabel(value?: string) {
  if (value === '1') {
    return '正常';
  }
  if (value === '2') {
    return '异常';
  }
  return renderEmpty(value);
}

function getStatusColor(value?: string) {
  if (value === '1') {
    return 'success';
  }
  if (value === '2') {
    return 'error';
  }
  return 'default';
}

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
  SysOperaLogItem,
  {
    beginTime: string;
    businessType: string;
    businessTypes: string;
    endTime: string;
    method: string;
    operIp: string;
    operUrl: string;
    operatorType: string;
    requestMethod: string;
    status: '' | number;
    title: string;
  },
  {
    beginTime?: string;
    businessType?: string;
    businessTypes?: string;
    endTime?: string;
    method?: string;
    operIp?: string;
    operUrl?: string;
    operatorType?: string;
    requestMethod?: string;
    status?: number;
    title?: string;
  }
>({
  createParams: (currentQuery) => ({
    title: currentQuery.title.trim() || undefined,
    businessType: currentQuery.businessType || undefined,
    businessTypes: currentQuery.businessTypes || undefined,
    method: currentQuery.method.trim() || undefined,
    requestMethod: currentQuery.requestMethod || undefined,
    operatorType: currentQuery.operatorType || undefined,
    operUrl: currentQuery.operUrl.trim() || undefined,
    operIp: currentQuery.operIp.trim() || undefined,
    status:
      currentQuery.status === '' ? undefined : Number(currentQuery.status),
    beginTime: currentQuery.beginTime.trim() || undefined,
    endTime: currentQuery.endTime.trim() || undefined,
  }),
  createQuery: () => ({
    title: '',
    businessType: '',
    businessTypes: '',
    method: '',
    requestMethod: '',
    operatorType: '',
    operUrl: '',
    operIp: '',
    status: '' as '' | number,
    beginTime: '',
    endTime: '',
  }),
  fallbackErrorMessage: '加载列表失败',
  fetcher: async (params) => {
    const res: SysOperaLogPageResult = await getOperaLogPage(params);
    return res;
  },
});

const renderEmpty = renderAdminEmpty;

const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  {
    title: '操作模块',
    dataIndex: 'title',
    key: 'title',
    width: 110,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '操作类型',
    dataIndex: 'businessType',
    key: 'businessType',
    width: 96,
    customRender: ({ text }: { text: string }) =>
      text
        ? h(
            Tag,
            { color: getBusinessTypeColor(text) },
            () => getBusinessTypeLabel(text),
          )
        : renderEmpty(text),
  },
  {
    title: '审计分类',
    dataIndex: 'businessTypes',
    key: 'businessTypes',
    width: 140,
    ellipsis: true,
    customRender: ({ text }: { text: string }) =>
      renderEmpty(getBusinessTypesLabel(text)),
  },
  {
    title: '请求方式',
    dataIndex: 'requestMethod',
    key: 'requestMethod',
    width: 90,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '操作者',
    dataIndex: 'operName',
    key: 'operName',
    width: 90,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '访问地址',
    dataIndex: 'operUrl',
    key: 'operUrl',
    width: 180,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '客户端IP',
    dataIndex: 'operIp',
    key: 'operIp',
    width: 110,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '操作时间',
    dataIndex: 'operTime',
    key: 'operTime',
    width: 165,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 86,
    customRender: ({ text }: { text: string }) =>
      text
        ? h(Tag, { color: getStatusColor(text) }, () => getStatusLabel(text))
        : renderEmpty(text),
  },
  {
    title: '审计摘要',
    dataIndex: 'remark',
    key: 'remark',
    width: 240,
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
];

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailRecord = ref<SysOperaLogItem | null>(null);

async function openDetail(record: SysOperaLogItem) {
  detailRecord.value = null;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const d = await getOperaLogDetail(record.id);
    detailRecord.value = d;
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

function onDelete(record: SysOperaLogItem) {
  const name = record.title || record.operName || `ID:${record.id}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除操作日志「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteOperaLog([record.id]);
        message.success('删除成功');
        fetchList();
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

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Audit</template>
    <template #title>操作日志</template>
    <template #description>
      重点展示系统设置、代码生成、角色与菜单变更等关键审计动作，并支持按审计分类快速筛选。
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="操作模块">
          <Input
            v-model:value="query.title"
            placeholder="请输入操作模块"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="操作类型">
          <Select
            v-model:value="query.businessType"
            :options="businessTypeOptions"
            class="w-full"
            placeholder="请选择操作类型"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="审计分类">
          <Select
            v-model:value="query.businessTypes"
            :options="businessTypesOptions"
            class="w-full"
            placeholder="请选择审计分类"
            allow-clear
            show-search
            option-filter-prop="label"
          />
        </AdminFilterField>
        <AdminFilterField label="请求方式">
          <Select
            v-model:value="query.requestMethod"
            :options="requestMethodOptions"
            class="w-full"
            placeholder="请选择请求方式"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="函数">
          <Input
            v-model:value="query.method"
            placeholder="请输入函数"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="操作主体">
          <Select
            v-model:value="query.operatorType"
            :options="operatorTypeOptions"
            class="w-full"
            placeholder="请选择操作主体"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="访问地址" :span="2">
          <Input
            v-model:value="query.operUrl"
            placeholder="请输入访问地址"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="IP">
          <Input
            v-model:value="query.operIp"
            placeholder="请输入 IP"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            class="w-full"
            placeholder="请选择状态"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="开始时间">
          <Input
            v-model:value="query.beginTime"
            placeholder="如 2024-01-01"
            allow-clear
          />
        </AdminFilterField>
        <AdminFilterField label="结束时间">
          <Input
            v-model:value="query.endTime"
            placeholder="如 2024-12-31"
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
        <div class="text-base font-semibold text-slate-900">操作日志列表</div>
        <p class="mt-1 text-sm text-slate-500">
          可按审计分类、操作类型和主体快速筛选，并在详情抽屉中查看摘要、请求参数与返回数据。
        </p>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysOperaLogItem) => record.id"
      :scroll="{ x: 1580 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:sysOperLog:query"
            @click="openDetail(record as SysOperaLogItem)"
          >
            详情
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:sysOperLog:remove"
            @click="onDelete(record as SysOperaLogItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <AdminDetailDrawer
      v-model:open="detailVisible"
      title="操作日志详情"
      width="680"
      :loading="detailLoading"
    >
      <template v-if="detailRecord">
        <AdminDetailSection title="操作概览" description="模块、请求方式和操作主体。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="ID">{{ detailRecord.id }}</Descriptions.Item>
            <Descriptions.Item label="操作模块">{{
              renderEmpty(detailRecord.title)
            }}</Descriptions.Item>
            <Descriptions.Item label="操作类型">
              <Tag :color="getBusinessTypeColor(detailRecord.businessType)">
                {{ getBusinessTypeLabel(detailRecord.businessType) }}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="审计分类">{{
              getBusinessTypesLabel(detailRecord.businessTypes)
            }}</Descriptions.Item>
            <Descriptions.Item label="请求方式">{{
              renderEmpty(detailRecord.requestMethod)
            }}</Descriptions.Item>
            <Descriptions.Item label="函数">{{
              renderEmpty(detailRecord.method)
            }}</Descriptions.Item>
            <Descriptions.Item label="操作主体">{{
              getOperatorTypeLabel(detailRecord.operatorType)
            }}</Descriptions.Item>
            <Descriptions.Item label="操作者">{{
              renderEmpty(detailRecord.operName)
            }}</Descriptions.Item>
            <Descriptions.Item label="部门">{{
              renderEmpty(detailRecord.deptName)
            }}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag :color="getStatusColor(detailRecord.status)">
                {{ getStatusLabel(detailRecord.status) }}
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>

        <AdminDetailSection title="审计摘要" description="关键系统操作会补充更可读的审计说明。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="摘要">{{
              renderEmpty(detailRecord.remark)
            }}</Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>

        <AdminDetailSection title="访问信息" description="快速查看访问入口、来源和耗时。">
          <Descriptions
            :column="1"
            bordered
            size="middle"
            :label-style="{ width: '120px' }"
          >
            <Descriptions.Item label="访问地址">{{
              renderEmpty(detailRecord.operUrl)
            }}</Descriptions.Item>
            <Descriptions.Item label="客户端IP">{{
              renderEmpty(detailRecord.operIp)
            }}</Descriptions.Item>
            <Descriptions.Item label="访问位置">{{
              renderEmpty(detailRecord.operLocation)
            }}</Descriptions.Item>
            <Descriptions.Item label="操作时间">{{
              formatAdminDateTime(detailRecord.operTime)
            }}</Descriptions.Item>
            <Descriptions.Item label="耗时">{{
              renderEmpty(detailRecord.latencyTime)
            }}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{{
              formatAdminDateTime(detailRecord.createdAt)
            }}</Descriptions.Item>
            <Descriptions.Item label="User Agent">{{
              renderEmpty(detailRecord.userAgent)
            }}</Descriptions.Item>
          </Descriptions>
        </AdminDetailSection>

        <AdminDetailSection title="请求参数" description="保留原始请求体，便于排查问题。">
          <AdminDetailCodeBlock :value="detailRecord.operParam" />
        </AdminDetailSection>

        <AdminDetailSection title="返回数据" description="查看接口返回内容。">
          <AdminDetailCodeBlock :value="detailRecord.jsonResult" />
        </AdminDetailSection>
      </template>
    </AdminDetailDrawer>
  </AdminPageShell>
</template>
