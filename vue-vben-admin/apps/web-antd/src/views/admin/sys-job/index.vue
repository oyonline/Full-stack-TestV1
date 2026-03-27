<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Input,
  Modal,
  Select,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnType, TablePaginationConfig } from 'ant-design-vue';

import {
  createSysJob,
  deleteSysJob,
  getSysJobDetail,
  getSysJobPage,
  removeSysJob,
  startSysJob,
  updateSysJob,
} from '#/api/core';
import type { SysJobItem } from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import type { AdminFormFieldSchema } from '#/components/admin/modal-form';
import AdminModalFormFields from '#/components/admin/modal-form-fields.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { renderAdminEmpty, resolveAdminErrorMessage } from '#/utils/admin-crud';

const submitting = ref(false);
const editLoading = ref(false);

const renderEmpty = renderAdminEmpty;

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
  SysJobItem,
  {
    invokeTarget: string;
    jobGroup: string;
    jobName: string;
    status: number | undefined;
  },
  {
    invokeTarget?: string;
    jobGroup?: string;
    jobName?: string;
    status?: number;
  }
>({
  createParams: (currentQuery) => ({
    jobName: currentQuery.jobName.trim() || undefined,
    jobGroup: currentQuery.jobGroup.trim() || undefined,
    invokeTarget: currentQuery.invokeTarget.trim() || undefined,
    status: currentQuery.status,
  }),
  createQuery: () => ({
    jobName: '',
    jobGroup: '',
    invokeTarget: '',
    status: undefined,
  }),
  fallbackErrorMessage: '加载任务列表失败',
  fetcher: async (params) => getSysJobPage(params),
});

const modalVisible = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const currentId = ref<number | null>(null);

const formState = reactive({
  jobName: '',
  jobGroup: '',
  jobType: 1,
  cronExpression: '',
  invokeTarget: '',
  args: '',
  misfirePolicy: 1,
  concurrent: 2,
  status: 2,
  entryId: 0,
});

const modalTitle = computed(() =>
  modalMode.value === 'create' ? '新增定时任务' : '编辑定时任务',
);

const statusOptions = [
  { label: '全部', value: undefined },
  { label: '停用', value: 1 },
  { label: '启用', value: 2 },
];

const formStatusOptions = [
  { label: '停用', value: 1 },
  { label: '启用', value: 2 },
];

const jobTypeOptions = [
  { label: 'HTTP', value: 1 },
  { label: 'EXEC', value: 2 },
];

const misfirePolicyOptions = [
  { label: '立即执行', value: 1 },
  { label: '执行一次', value: 2 },
  { label: '放弃执行', value: 3 },
];

const concurrentOptions = [
  { label: '允许', value: 1 },
  { label: '禁止', value: 2 },
];

const jobFormFields = computed<AdminFormFieldSchema[]>(() => [
  {
    component: 'input',
    field: 'jobName',
    label: '任务名称',
    required: true,
  },
  {
    component: 'input',
    field: 'jobGroup',
    label: '任务分组',
  },
  {
    component: 'select',
    field: 'jobType',
    label: '任务类型',
    options: jobTypeOptions,
  },
  {
    component: 'select',
    field: 'status',
    label: '状态',
    options: formStatusOptions,
  },
  {
    component: 'input',
    field: 'cronExpression',
    label: 'Cron',
    placeholder: '例如：0 */5 * * * *',
    required: true,
    span: 2,
  },
  {
    component: 'input',
    field: 'invokeTarget',
    label: '调用目标',
    required: true,
    span: 2,
  },
  {
    component: 'input',
    field: 'args',
    label: '参数',
    span: 2,
  },
  {
    component: 'select',
    field: 'misfirePolicy',
    label: '执行策略',
    options: misfirePolicyOptions,
  },
  {
    component: 'select',
    field: 'concurrent',
    label: '并发控制',
    options: concurrentOptions,
  },
]);

function resetFormState() {
  formState.jobName = '';
  formState.jobGroup = '';
  formState.jobType = 1;
  formState.cronExpression = '';
  formState.invokeTarget = '';
  formState.args = '';
  formState.misfirePolicy = 1;
  formState.concurrent = 2;
  formState.status = 2;
  formState.entryId = 0;
}

function getStatusText(status: number) {
  return status === 2 ? '启用' : '停用';
}

function getStatusColor(status: number) {
  return status === 2 ? 'success' : 'default';
}

function getJobTypeText(jobType: number) {
  return jobType === 2 ? 'EXEC' : 'HTTP';
}

function getMisfirePolicyText(value: number) {
  switch (value) {
    case 1: {
      return '立即执行';
    }
    case 2: {
      return '执行一次';
    }
    case 3: {
      return '放弃执行';
    }
    default: {
      return String(value ?? '-');
    }
  }
}

function getConcurrentText(value: number) {
  return value === 1 ? '允许' : '禁止';
}

function handleTableChange(pager: TablePaginationConfig) {
  onTableChange(pager);
}

function openCreateModal() {
  modalMode.value = 'create';
  currentId.value = null;
  resetFormState();
  modalVisible.value = true;
}

async function openEditModal(record: SysJobItem) {
  modalMode.value = 'edit';
  currentId.value = record.jobId;
  modalVisible.value = true;
  editLoading.value = true;

  try {
    const detail = await getSysJobDetail(record.jobId);

    formState.jobName = detail.jobName || '';
    formState.jobGroup = detail.jobGroup || '';
    formState.jobType = detail.jobType ?? 1;
    formState.cronExpression = detail.cronExpression || '';
    formState.invokeTarget = detail.invokeTarget || '';
    formState.args = detail.args || '';
    formState.misfirePolicy = detail.misfirePolicy ?? 1;
    formState.concurrent = detail.concurrent ?? 2;
    formState.status = detail.status ?? 2;
    formState.entryId = detail.entryId ?? 0;
  } finally {
    editLoading.value = false;
  }
}

function validateForm() {
  if (!formState.jobName.trim()) {
    message.error('请输入任务名称');
    return false;
  }
  if (!formState.cronExpression.trim()) {
    message.error('请输入 Cron 表达式');
    return false;
  }
  if (!formState.invokeTarget.trim()) {
    message.error('请输入调用目标');
    return false;
  }
  return true;
}

async function handleSubmit() {
  if (!validateForm()) {
    return;
  }

  submitting.value = true;

  try {
    const payload = {
      ...(modalMode.value === 'edit' && currentId.value !== null
        ? { jobId: currentId.value }
        : {}),
      jobName: formState.jobName.trim(),
      jobGroup: formState.jobGroup.trim(),
      jobType: formState.jobType,
      cronExpression: formState.cronExpression.trim(),
      invokeTarget: formState.invokeTarget.trim(),
      args: formState.args.trim(),
      misfirePolicy: formState.misfirePolicy,
      concurrent: formState.concurrent,
      status: formState.status,
      entryId: formState.entryId,
    };

    if (modalMode.value === 'create') {
      await createSysJob(payload);
      message.success('新增成功');
    } else {
      await updateSysJob(payload);
      message.success('编辑成功');
    }

    modalVisible.value = false;
    fetchList();
  } finally {
    submitting.value = false;
  }
}

function handleDelete(record: SysJobItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定删除任务「${record.jobName}」吗？`,
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSysJob([record.jobId]);
        message.success('删除成功');
        fetchList();
      } catch (error) {
        message.error(resolveAdminErrorMessage(error, '删除失败'));
      }
    },
  });
}

async function handleStart(record: SysJobItem) {
  try {
    await startSysJob(record.jobId);
    message.success('启动成功');
    fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '启动失败'));
  }
}

async function handleRemove(record: SysJobItem) {
  try {
    await removeSysJob(record.jobId);
    message.success('停止成功');
    fetchList();
  } catch (error) {
    message.error(resolveAdminErrorMessage(error, '停止失败'));
  }
}

const columns: TableColumnType<SysJobItem>[] = [
  {
    title: 'ID',
    dataIndex: 'jobId',
    width: 80,
  },
  {
    title: '任务名称',
    dataIndex: 'jobName',
    width: 160,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '任务分组',
    dataIndex: 'jobGroup',
    width: 120,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '任务类型',
    dataIndex: 'jobType',
    width: 100,
  },
  {
    title: 'Cron表达式',
    dataIndex: 'cronExpression',
    width: 180,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '调用目标',
    dataIndex: 'invokeTarget',
    ellipsis: true,
    width: 220,
  },
  {
    title: '参数',
    dataIndex: 'args',
    ellipsis: true,
    width: 160,
  },
  {
    title: '执行策略',
    dataIndex: 'misfirePolicy',
    width: 110,
  },
  {
    title: '并发',
    dataIndex: 'concurrent',
    width: 90,
  },
  {
    title: '状态',
    dataIndex: 'status',
    width: 100,
  },
  {
    title: 'EntryId',
    dataIndex: 'entryId',
    width: 100,
  },
  {
    title: '操作',
    key: 'action',
    fixed: 'right',
    width: 260,
  },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>定时任务</template>
    <template #description>
      统一维护任务名称、分组、调用目标和状态。列表页收口为标准搜索网格，操作区和表格层级更清晰。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="job:sysJob:add"
        @click="openCreateModal"
      >
        新增
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="任务名称">
          <Input
            v-model:value="query.jobName"
            allow-clear
            placeholder="请输入任务名称"
          />
        </AdminFilterField>
        <AdminFilterField label="任务分组">
          <Input
            v-model:value="query.jobGroup"
            allow-clear
            placeholder="请输入任务分组"
          />
        </AdminFilterField>
        <AdminFilterField label="调用目标" :span="2">
          <Input
            v-model:value="query.invokeTarget"
            allow-clear
            placeholder="请输入调用目标"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            allow-clear
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
        <div class="text-base font-semibold text-slate-900">任务列表</div>
      </div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysJobItem) => record.jobId"
      :scroll="{ x: 1300 }"
      size="middle"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record, text }">
        <template v-if="column.dataIndex === 'jobType'">
          {{ getJobTypeText(text) }}
        </template>

        <template v-else-if="column.dataIndex === 'misfirePolicy'">
          {{ getMisfirePolicyText(text) }}
        </template>

        <template v-else-if="column.dataIndex === 'concurrent'">
          {{ getConcurrentText(text) }}
        </template>

        <template v-else-if="column.dataIndex === 'status'">
          <Tag :color="getStatusColor(text)">
            {{ getStatusText(text) }}
          </Tag>
        </template>

        <template v-else-if="column.key === 'action'">
          <div class="flex flex-wrap justify-end gap-x-2 gap-y-1">
            <AdminActionButton
              type="link"
              size="small"
              codes="job:sysJob:edit"
              @click="openEditModal(record as SysJobItem)"
            >
              编辑
            </AdminActionButton>
            <AdminActionButton
              type="link"
              size="small"
              codes="job:sysJob:edit"
              @click="handleStart(record as SysJobItem)"
            >
              启动
            </AdminActionButton>
            <AdminActionButton
              type="link"
              size="small"
              codes="job:sysJob:edit"
              @click="handleRemove(record as SysJobItem)"
            >
              停止
            </AdminActionButton>
            <AdminActionButton
              danger
              type="link"
              size="small"
              codes="job:sysJob:remove"
              @click="handleDelete(record as SysJobItem)"
            >
              删除
            </AdminActionButton>
          </div>
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="modalVisible"
      :confirm-loading="submitting"
      :title="modalTitle"
      :width="820"
      destroy-on-close
      ok-text="保存"
      cancel-text="取消"
      @ok="handleSubmit"
    >
      <div v-if="editLoading" class="py-8 text-center text-[#999]">
        加载中...
      </div>

      <AdminModalFormFields
        v-else
        :model="formState"
        :fields="jobFormFields"
      />
    </Modal>
  </AdminPageShell>
</template>
