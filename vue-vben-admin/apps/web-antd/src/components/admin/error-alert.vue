<script lang="ts" setup>
import { Alert } from 'ant-design-vue';

defineProps<{
  message?: string;
}>();
</script>

<template>
  <Alert
    v-if="message"
    show-icon
    type="error"
    :message="message"
    class="mb-4"
  />
</template>
