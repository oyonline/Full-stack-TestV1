<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue';

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

const loading = ref(false);
const submitting = ref(false);
const editLoading = ref(false);

const list = ref<SysJobItem[]>([]);
const total = ref(0);

const searchForm = reactive({
  jobName: '',
  jobGroup: '',
  invokeTarget: '',
  status: undefined as number | undefined,
});

const pagination = reactive({
  current: 1,
  pageSize: 10,
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

async function fetchList() {
  loading.value = true;
  try {
    const res = await getSysJobPage({
      pageIndex: pagination.current,
      pageSize: pagination.pageSize,
      jobName: searchForm.jobName || undefined,
      jobGroup: searchForm.jobGroup || undefined,
      invokeTarget: searchForm.invokeTarget || undefined,
      status: searchForm.status,
    });

    list.value = res?.list || [];
    total.value = res?.count || 0;
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchList();
}

function handleReset() {
  searchForm.jobName = '';
  searchForm.jobGroup = '';
  searchForm.invokeTarget = '';
  searchForm.status = undefined;
  pagination.current = 1;
  fetchList();
}

function handleTableChange(pager: TablePaginationConfig) {
  pagination.current = pager.current || 1;
  pagination.pageSize = pager.pageSize || 10;
  fetchList();
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
      await deleteSysJob([record.jobId]);
      message.success('删除成功');
      fetchList();
    },
  });
}

async function handleStart(record: SysJobItem) {
  await startSysJob(record.jobId);
  message.success('启动成功');
  fetchList();
}

async function handleRemove(record: SysJobItem) {
  await removeSysJob(record.jobId);
  message.success('停止成功');
  fetchList();
}

const columns = [
  {
    title: 'ID',
    dataIndex: 'jobId',
    width: 80,
  },
  {
    title: '任务名称',
    dataIndex: 'jobName',
    width: 160,
  },
  {
    title: '任务分组',
    dataIndex: 'jobGroup',
    width: 120,
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
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <div class="text-[18px] font-semibold">定时任务 / SCHEDULE</div>
      <Space>
        <Button @click="fetchList">刷新</Button>
        <Button type="primary" @click="openCreateModal">新增</Button>
      </Space>
    </div>

    <div class="mb-4 rounded bg-white p-4">
      <Form layout="inline">
        <FormItem label="任务名称">
          <Input
            v-model:value="searchForm.jobName"
            allow-clear
            placeholder="请输入任务名称"
          />
        </FormItem>
        <FormItem label="任务分组">
          <Input
            v-model:value="searchForm.jobGroup"
            allow-clear
            placeholder="请输入任务分组"
          />
        </FormItem>
        <FormItem label="调用目标">
          <Input
            v-model:value="searchForm.invokeTarget"
            allow-clear
            placeholder="请输入调用目标"
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="searchForm.status"
            :options="statusOptions"
            allow-clear
            placeholder="请选择状态"
            style="width: 140px"
          />
        </FormItem>
        <FormItem>
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </FormItem>
      </Form>
    </div>

    <div class="rounded bg-white p-4">
      <Table
        :columns="columns"
        :data-source="list"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total,
          showSizeChanger: true,
          showTotal: (v: number) => `共 ${v} 条`,
        }"
        :row-key="(record: SysJobItem) => record.jobId"
        bordered
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
            <Space>
              <Button type="link" size="small" @click="openEditModal(record)">
                编辑
              </Button>
              <Button type="link" size="small" @click="handleStart(record)">
                启动
              </Button>
              <Button type="link" size="small" @click="handleRemove(record)">
                停止
              </Button>
              <Button
                danger
                type="link"
                size="small"
                @click="handleDelete(record)"
              >
                删除
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </div>

    <Modal
      v-model:open="modalVisible"
      :confirm-loading="submitting"
      :title="modalTitle"
      destroy-on-close
      ok-text="保存"
      cancel-text="取消"
      @ok="handleSubmit"
    >
      <div v-if="editLoading" class="py-8 text-center text-[#999]">
        加载中...
      </div>

      <Form v-else :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <FormItem label="任务名称" required>
          <Input v-model:value="formState.jobName" allow-clear />
        </FormItem>
        <FormItem label="任务分组">
          <Input v-model:value="formState.jobGroup" allow-clear />
        </FormItem>
        <FormItem label="任务类型">
          <Select
            v-model:value="formState.jobType"
            :options="jobTypeOptions"
          />
        </FormItem>
        <FormItem label="Cron" required>
          <Input
            v-model:value="formState.cronExpression"
            allow-clear
            placeholder="例如：0 */5 * * * *"
          />
        </FormItem>
        <FormItem label="调用目标" required>
          <Input v-model:value="formState.invokeTarget" allow-clear />
        </FormItem>
        <FormItem label="参数">
          <Input v-model:value="formState.args" allow-clear />
        </FormItem>
        <FormItem label="执行策略">
          <Select
            v-model:value="formState.misfirePolicy"
            :options="misfirePolicyOptions"
          />
        </FormItem>
        <FormItem label="并发控制">
          <Select
            v-model:value="formState.concurrent"
            :options="concurrentOptions"
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="formState.status"
            :options="formStatusOptions"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
