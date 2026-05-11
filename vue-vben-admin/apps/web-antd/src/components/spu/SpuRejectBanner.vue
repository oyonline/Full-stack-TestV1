<script lang="ts" setup>
import type { WorkflowActionLog } from '#/api/core';

import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

defineProps<{
  rejectAction: WorkflowActionLog;
}>();
</script>

<template>
  <div
    class="rounded-sm border border-red-200 bg-red-50 p-3 text-sm text-red-700"
  >
    <div class="font-semibold">驳回理由</div>
    <div>{{ rejectAction.comment || '无备注' }}</div>
    <div class="mt-1 text-xs text-red-500">
      {{ formatAdminDateTime(rejectAction.createdAt) }} · {{ renderAdminEmpty(rejectAction.operatorName) }}
    </div>
  </div>
</template>
