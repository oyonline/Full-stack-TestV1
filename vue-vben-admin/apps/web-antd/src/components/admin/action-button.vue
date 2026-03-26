<script lang="ts" setup>
import { computed, useAttrs } from 'vue';

import { Button } from 'ant-design-vue';

import { useAdminPermission } from '#/composables/use-admin-permission';

defineOptions({
  inheritAttrs: false,
});

const props = withDefaults(
  defineProps<{
    codes?: string | string[];
    disabledMode?: 'disable' | 'hide';
    roles?: string | string[];
  }>(),
  {
    disabledMode: 'hide',
  },
);

const attrs = useAttrs();
const { hasPermission } = useAdminPermission({
  codes: props.codes,
  roles: props.roles,
});

const shouldRender = computed(() => {
  return props.disabledMode === 'disable' || hasPermission.value;
});

const mergedDisabled = computed(() => {
  return Boolean(attrs.disabled) || !hasPermission.value;
});
</script>

<template>
  <Button v-if="shouldRender" v-bind="attrs" :disabled="mergedDisabled">
    <slot />
  </Button>
</template>
