<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import {
  Button,
  Descriptions,
  Modal,
  Table,
  Tag,
  Timeline,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  approveWorkflowTask,
  type BizActionLogItem,
  getBizActionLogPage,
  getWorkflowInstanceDetail,
  rejectWorkflowTask,
  withdrawWorkflowInstance,
  type WorkflowActionLog,
  type WorkflowInstanceDetailResult,
  type WorkflowTaskDetail,
} from '#/api/core';
import AdminDetailDrawer from '#/components/admin/detail-drawer.vue';
import AdminDetailSection from '#/components/admin/detail-section.vue';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import {
  getWorkflowActionLabel,
  getWorkflowApproverTypeLabel,
  getWorkflowStatusLabel,
  getWorkflowTaskStatusLabel,
  workflowStatusColors,
} from '../constants';
import WorkflowAttachmentSection from './workflow-attachment-section.vue';

const props = withDefaults(
  defineProps<{
    canApprove?: boolean;
    canReject?: boolean;
    canWithdraw?: boolean;
    currentTaskId?: number | null;
    instanceId?: null | number;
    open: boolean;
  }>(),
  {
    canApprove: false,
    canReject: false,
    canWithdraw: false,
    currentTaskId: null,
    instanceId: null,
  },
);

const emit = defineEmits<{
  acted: [];
  'update:open': [value: boolean];
}>();

const detailLoading = ref(false);
const actionLoading = ref(false);
const detail = ref<null | WorkflowInstanceDetailResult>(null);
const bizLogs = ref<BizActionLogItem[]>([]);

const taskColumns: TableColumnType[] = [
  {
    title: '节点',
    dataIndex: 'nodeName',
    key: 'nodeName',
    width: 120,
  },
  {
    title: '审批对象',
    key: 'assignee',
    width: 140,
    customRender: ({ record }: { record: WorkflowTaskDetail }) =>
      `${getWorkflowApproverTypeLabel(record.assigneeType)}：${renderAdminEmpty(record.assigneeName)}`,
  },
  {
    title: '任务状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: string }) => getWorkflowTaskStatusLabel(text),
  },
  {
    title: '处理人',
    dataIndex: 'actionByName',
    key: 'actionByName',
    width: 110,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 160,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '处理时间',
    dataIndex: 'processedAt',
    key: 'processedAt',
    width: 160,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
];

const timelineItems = computed(() =>
  (detail.value?.actions || []).map((item: WorkflowActionLog) => ({
    children: `${formatAdminDateTime(item.createdAt)}  ${renderAdminEmpty(item.operatorName)} · ${getWorkflowActionLabel(item.action)}${item.comment ? ` · ${item.comment}` : ''}`,
    color:
      item.action === 'approve'
        ? 'green'
        : item.action === 'reject'
          ? 'red'
          : item.action === 'withdraw'
            ? 'gray'
            : 'blue',
  })),
);

const currentTask = computed(() =>
  detail.value?.tasks?.find((item) => item.taskId === props.currentTaskId),
);

const showApprove = computed(
  () =>
    props.canApprove &&
    detail.value?.instance.status === 'in_review' &&
    currentTask.value?.status === 'pending' &&
    !!props.currentTaskId,
);

const showReject = computed(
  () =>
    props.canReject &&
    detail.value?.instance.status === 'in_review' &&
    currentTask.value?.status === 'pending' &&
    !!props.currentTaskId,
);

const showWithdraw = computed(
  () =>
    props.canWithdraw &&
    detail.value?.instance.status === 'in_review' &&
    !!props.instanceId,
);

const title = computed(() => detail.value?.instance.title || '流程实例详情');

async function loadDetail() {
  if (!props.instanceId) {
    detail.value = null;
    bizLogs.value = [];
    return;
  }
  detailLoading.value = true;
  try {
    detail.value = await getWorkflowInstanceDetail(props.instanceId);
    if (detail.value?.instance) {
      const { moduleKey, businessType, businessId } = detail.value.instance;
      if (moduleKey && businessType && businessId) {
        const result = await getBizActionLogPage({ moduleKey, businessType, businessId });
        bizLogs.value = result.list ?? [];
      }
    }
  } catch (error) {
    message.error((error as Error)?.message || '加载流程详情失败');
    detail.value = null;
    bizLogs.value = [];
  } finally {
    detailLoading.value = false;
  }
}

async function handleAction(
  action: 'approve' | 'reject' | 'withdraw',
  targetId: number,
) {
  if (!targetId) {
    return;
  }

  const actionLabel = getWorkflowActionLabel(action);
  Modal.confirm({
    title: `确认${actionLabel}`,
    content: `确认执行“${actionLabel}”吗？`,
    async onOk() {
      actionLoading.value = true;
      try {
        if (action === 'approve') {
          detail.value = await approveWorkflowTask(targetId);
        } else if (action === 'reject') {
          detail.value = await rejectWorkflowTask(targetId);
        } else {
          detail.value = await withdrawWorkflowInstance(targetId);
        }
        message.success(`${actionLabel}成功`);
        emit('acted');
      } catch (error) {
        message.error((error as Error)?.message || `${actionLabel}失败`);
      } finally {
        actionLoading.value = false;
      }
    },
  });
}

watch(
  () => [props.open, props.instanceId] as const,
  ([open, instanceId]) => {
    if (open && instanceId) {
      void loadDetail();
    }
  },
  { immediate: true },
);
</script>

<template>
  <AdminDetailDrawer
    :loading="detailLoading"
    :open="open"
    :title="title"
    :width="860"
    @update:open="$emit('update:open', $event)"
  >
    <template v-if="detail">
      <AdminDetailSection title="基础信息">
        <Descriptions :column="2" bordered size="small">
          <Descriptions.Item label="标题">
            {{ renderAdminEmpty(detail.instance.title) }}
          </Descriptions.Item>
          <Descriptions.Item label="业务类型">
            {{ renderAdminEmpty(detail.instance.businessType) }}
          </Descriptions.Item>
          <Descriptions.Item label="业务编号">
            {{ renderAdminEmpty(detail.instance.businessNo) }}
          </Descriptions.Item>
          <Descriptions.Item label="发起人">
            {{ renderAdminEmpty(detail.instance.starterName) }}
          </Descriptions.Item>
          <Descriptions.Item label="当前状态">
            <Tag :color="workflowStatusColors[detail.instance.status] || 'default'">
              {{ getWorkflowStatusLabel(detail.instance.status) }}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="当前节点">
            {{ renderAdminEmpty(detail.instance.currentNodeName) }}
          </Descriptions.Item>
          <Descriptions.Item label="发起时间">
            {{ formatAdminDateTime(detail.instance.startedAt) }}
          </Descriptions.Item>
          <Descriptions.Item label="完成时间">
            {{ formatAdminDateTime(detail.instance.finishedAt) }}
          </Descriptions.Item>
        </Descriptions>
      </AdminDetailSection>

      <AdminDetailSection title="关联任务">
        <Table
          :columns="taskColumns"
          :data-source="detail.tasks"
          :pagination="false"
          row-key="taskId"
          size="small"
          :scroll="{ x: 860 }"
        />
      </AdminDetailSection>

      <AdminDetailSection title="审批记录时间线">
        <Timeline v-if="timelineItems.length" :items="timelineItems" />
        <div v-else class="text-sm text-slate-400">暂无审批记录</div>
      </AdminDetailSection>

      <AdminDetailSection title="业务操作记录">
        <Timeline
          v-if="bizLogs.length"
          :items="bizLogs.map((item: BizActionLogItem) => ({
            children: `${formatAdminDateTime(item.createdAt)}  ${item.operatorName} · ${item.action}${item.fromStatus || item.toStatus ? ` · ${item.fromStatus || '—'} → ${item.toStatus || '—'}` : ''}${item.remark ? ` · ${item.remark}` : ''}`,
            color: 'blue',
          }))"
        />
        <div v-else class="text-sm text-slate-400">暂无业务操作记录</div>
      </AdminDetailSection>

      <AdminDetailSection title="关联附件">
        <WorkflowAttachmentSection
          :business-id="detail.instance.businessId"
          :business-no="detail.instance.businessNo"
          :business-type="detail.instance.businessType"
          :module-key="detail.instance.moduleKey"
        />
      </AdminDetailSection>

      <div
        v-if="showApprove || showReject || showWithdraw"
        class="flex flex-wrap justify-end gap-2 border-t border-slate-100 pt-4"
      >
        <Button
          v-if="showReject && currentTaskId"
          data-testid="workflow-reject-action"
          danger
          :loading="actionLoading"
          @click="handleAction('reject', currentTaskId)"
        >
          驳回
        </Button>
        <Button
          v-if="showApprove && currentTaskId"
          data-testid="workflow-approve-action"
          type="primary"
          :loading="actionLoading"
          @click="handleAction('approve', currentTaskId)"
        >
          通过
        </Button>
        <Button
          v-if="showWithdraw && instanceId"
          data-testid="workflow-withdraw-action"
          :loading="actionLoading"
          @click="handleAction('withdraw', instanceId)"
        >
          撤回
        </Button>
      </div>
    </template>
  </AdminDetailDrawer>
</template>
