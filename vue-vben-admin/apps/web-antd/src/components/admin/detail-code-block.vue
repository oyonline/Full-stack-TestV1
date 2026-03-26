<script lang="ts" setup>
defineProps<{
  emptyText?: string;
  value?: null | string;
}>();
</script>

<template>
  <pre
    class="app-radius-box max-h-40 overflow-auto bg-white p-3 text-xs leading-6 text-slate-600"
    >{{ value || emptyText || '-' }}</pre
  >
</template>
