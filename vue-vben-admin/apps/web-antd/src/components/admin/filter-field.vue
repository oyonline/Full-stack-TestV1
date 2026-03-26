<script lang="ts" setup>
import { computed } from 'vue';

interface Props {
  label: string;
  span?: 1 | 2;
}

const props = withDefaults(defineProps<Props>(), {
  span: 1,
});

const wrapperClass = computed(() =>
  props.span === 2 ? 'md:col-span-2 xl:col-span-2' : '',
);
</script>

<template>
  <div :class="wrapperClass" class="space-y-2">
    <label class="block text-sm font-medium text-slate-700">
      {{ label }}
    </label>
    <slot></slot>
  </div>
</template>
