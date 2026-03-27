<script lang="ts" setup>
import { computed } from 'vue';

import { Empty, Modal, Tabs } from 'ant-design-vue';

interface PreviewItem {
  key: string;
  label: string;
  content: string;
}

interface Props {
  loading?: boolean;
  open: boolean;
  previewData: Record<string, string>;
  title: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:open': [value: boolean];
}>();

const previewItems = computed<PreviewItem[]>(() =>
  Object.entries(props.previewData ?? {}).map(([key, value]) => ({
    key,
    label: key.replace('template/', ''),
    content: value,
  })),
);

function onUpdateOpen(value: boolean) {
  emit('update:open', value);
}
</script>

<template>
  <Modal
    :open="open"
    :title="title"
    width="1120px"
    :footer="null"
    @update:open="onUpdateOpen"
  >
    <div v-if="loading" class="py-20 text-center text-slate-500">
      正在加载预览...
    </div>
    <template v-else>
      <Empty
        v-if="!previewItems.length"
        description="暂无预览内容"
        class="py-12"
      />
      <Tabs v-else type="card">
        <Tabs.TabPane
          v-for="item in previewItems"
          :key="item.key"
          :tab="item.label"
        >
          <pre class="app-radius-box max-h-[60vh] overflow-auto bg-slate-950 p-4 text-xs leading-6 text-slate-100">{{ item.content }}</pre>
        </Tabs.TabPane>
      </Tabs>
    </template>
  </Modal>
</template>
