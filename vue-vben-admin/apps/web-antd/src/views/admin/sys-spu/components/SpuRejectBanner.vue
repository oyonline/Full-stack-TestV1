<script lang="ts" setup>
import { Alert } from 'ant-design-vue';

import type { WorkflowInstanceAction } from '#/api/core';

const props = defineProps<{
  rejectAction: WorkflowInstanceAction;
}>();
</script>

<template>
  <Alert
    type="error"
    show-icon
    class="mb-4"
    :message="`驳回理由：${props.rejectAction.comment || '无备注'}`"
    :description="`由 ${props.rejectAction.operatorName || '-'} 于 ${props.rejectAction.operatedAt || '-'} 驳回`"
  />
</template>
