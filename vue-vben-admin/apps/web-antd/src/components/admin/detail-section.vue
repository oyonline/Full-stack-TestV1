<script lang="ts" setup>
defineProps<{
  description?: string;
  title: string;
}>();
</script>

<template>
  <section class="space-y-3">
    <div class="space-y-1">
      <div class="text-sm font-semibold text-slate-900">{{ title }}</div>
      <p v-if="description" class="text-xs leading-5 text-slate-500">
        {{ description }}
      </p>
    </div>
    <div class="app-radius-box border border-slate-200 bg-slate-50/70 p-3">
      <slot />
    </div>
  </section>
</template>
