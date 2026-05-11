<script lang="ts" setup>
import { computed } from 'vue';

import { Timeline } from 'ant-design-vue';

import type { WorkflowActionLog } from '#/api/core';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import { getWorkflowActionLabel } from '#/views/platform/workflow/constants';

const props = defineProps<{
  actions: WorkflowActionLog[];
  compact?: boolean;
}>();

const timelineItems = computed(() =>
  props.actions.map((item) => ({
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
</script>

<template>
  <div>
    <Timeline v-if="timelineItems.length" :items="timelineItems" />
    <div v-else class="text-sm text-slate-400">暂无审批记录</div>
  </div>
</template>
