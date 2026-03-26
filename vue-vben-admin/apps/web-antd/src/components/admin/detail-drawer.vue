<script lang="ts" setup>
import { Drawer } from 'ant-design-vue';

withDefaults(
  defineProps<{
    loading?: boolean;
    open: boolean;
    title: string;
    width?: number | string;
  }>(),
  {
    loading: false,
    width: 560,
  },
);

defineEmits<{
  'update:open': [value: boolean];
}>();
</script>

<template>
  <Drawer
    :open="open"
    :title="title"
    :width="width"
    :footer-style="{ textAlign: 'right' }"
    @update:open="$emit('update:open', $event)"
  >
    <div v-if="loading" class="py-8 text-center text-sm text-slate-400">
      加载中…
    </div>
    <div v-else class="space-y-4">
      <slot />
    </div>
  </Drawer>
</template>
