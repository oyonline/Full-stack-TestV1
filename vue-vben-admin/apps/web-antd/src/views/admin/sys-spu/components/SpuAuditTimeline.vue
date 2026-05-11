<script lang="ts" setup>
import { computed } from 'vue';

import { Alert, Button, Spin, Tag, Timeline } from 'ant-design-vue';

import type { WorkflowInstanceAction } from '#/api/core';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import { getWorkflowActionLabel } from '#/views/platform/workflow/constants';

const props = withDefaults(
  defineProps<{
    actions: WorkflowInstanceAction[];
    compact?: boolean;
    error?: string;
    loading?: boolean;
  }>(),
  {
    compact: false,
    error: '',
    loading: false,
  },
);

const emit = defineEmits<{
  retry: [];
}>();

const displayedActions = computed(() => {
  if (props.compact) {
    return props.actions.slice(0, 5);
  }
  return props.actions;
});

const hasMore = computed(() => props.compact && props.actions.length > 5);

function actionColor(action: string): string {
  switch (action) {
    case 'approve':
      return 'green';
    case 'reject':
      return 'red';
    case 'withdraw':
      return 'gray';
    default:
      return 'blue';
  }
}

function actionTagColor(action: string): string {
  switch (action) {
    case 'approve':
      return 'success';
    case 'reject':
      return 'error';
    case 'withdraw':
      return 'default';
    default:
      return 'processing';
  }
}

const timelineItems = computed(() =>
  displayedActions.value.map((item) => ({
    color: actionColor(item.action),
    children: `${formatAdminDateTime(item.operatedAt)}  ${renderAdminEmpty(item.operatorName)} · ${getWorkflowActionLabel(item.action)}${item.comment ? ` · ${item.comment}` : ''}`,
  })),
);
</script>

<template>
  <div>
    <!-- 加载态 -->
    <div v-if="loading" class="py-8 text-center">
      <Spin />
    </div>

    <!-- 错误态 -->
    <Alert
      v-else-if="error"
      type="error"
      show-icon
      class="mt-2 mb-4"
      :message="error"
    >
      <template #action>
        <Button size="small" @click="emit('retry')">重试</Button>
      </template>
    </Alert>

    <!-- 空态 -->
    <div v-else-if="!actions.length" class="py-8 text-center text-sm text-slate-400">
      暂无审批记录
    </div>

    <!-- 时间轴 -->
    <div v-else>
      <Timeline v-if="timelineItems.length" :items="timelineItems" />

      <!-- 紧凑模式：逐条渲染 + 角色徽标 -->
      <div v-if="compact" class="relative pl-6 pt-2">
        <div
          v-for="(item, idx) in displayedActions"
          :key="item.actionId"
          class="relative mb-6 last:mb-0"
        >
          <!-- 连接线 -->
          <div
            v-if="idx < displayedActions.length - 1"
            class="absolute left-[-17px] top-3 h-full w-px bg-slate-200"
          />
          <!-- 圆点 -->
          <div
            class="absolute left-[-21px] top-1.5 h-2.5 w-2.5 rounded-full border-2 border-white"
            :style="{ backgroundColor: actionColor(item.action) }"
          />
          <div class="flex flex-wrap items-center gap-2">
            <Tag :color="actionTagColor(item.action)">
              {{ getWorkflowActionLabel(item.action) }}
            </Tag>
            <span class="text-sm font-medium text-slate-700">
              {{ renderAdminEmpty(item.operatorName) }}
            </span>
            <Tag v-if="item.operatorRole" color="blue">
              {{ item.operatorRole }}
            </Tag>
            <span class="text-xs text-slate-400">{{ item.nodeKey }}</span>
          </div>
          <div v-if="item.comment" class="mt-1 text-sm text-slate-500">
            {{ item.comment }}
          </div>
          <div class="mt-0.5 text-xs text-slate-400">
            {{ formatAdminDateTime(item.operatedAt) }}
          </div>
        </div>
      </div>

      <div v-if="hasMore" class="mt-2 text-center text-sm text-slate-400">
        共 {{ actions.length }} 条记录，完整记录请在详情页查看
      </div>
    </div>
  </div>
</template>
