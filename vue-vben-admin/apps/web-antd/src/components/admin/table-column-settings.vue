<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';

import { useSortable } from '@vben/hooks';

import { Button, Drawer, Select, Switch, Tag } from 'ant-design-vue';
import type { Sortable } from '@vben/hooks';

import type {
  AdminTableColumnFixed,
  AdminTableColumnSettingsItem,
} from '#/utils/admin-table-columns';

interface Props {
  columns: AdminTableColumnSettingsItem[];
  open: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  changeFixed: [payload: { fixed: AdminTableColumnFixed; key: string }];
  reorder: [payload: { newIndex: number; oldIndex: number }];
  reset: [];
  toggleVisible: [payload: { key: string; visible: boolean }];
  'update:open': [value: boolean];
}>();

const sortableContainerRef = ref<HTMLElement | null>(null);
const sortableRef = ref<null | Sortable>(null);

const fixedOptions = computed(() => [
  { label: '不固定', value: 'none' },
  { label: '固定左侧', value: 'left' },
  { label: '固定右侧', value: 'right' },
]);

function closeDrawer() {
  emit('update:open', false);
}

function destroySortable() {
  sortableRef.value?.destroy();
  sortableRef.value = null;
}

async function initializeSortable() {
  const sortableContainer = sortableContainerRef.value;
  if (!sortableContainer) {
    return;
  }

  destroySortable();

  const { initializeSortable } = useSortable(sortableContainer, {
    animation: 180,
    handle: '[data-drag-handle="true"]',
    onEnd(event: { newIndex?: number; oldIndex?: number }) {
      if (
        event.oldIndex == null ||
        event.newIndex == null ||
        event.oldIndex === event.newIndex
      ) {
        return;
      }
      emit('reorder', {
        newIndex: event.newIndex,
        oldIndex: event.oldIndex,
      });
    },
  });

  sortableRef.value = await initializeSortable();
}

function onFixedChange(key: string, value: string) {
  emit('changeFixed', {
    fixed: value === 'none' ? false : (value as AdminTableColumnFixed),
    key,
  });
}

function onVisibilityChange(key: string, visible: boolean) {
  emit('toggleVisible', {
    key,
    visible,
  });
}

watch(
  () => [props.open, props.columns.map((column) => column.key).join('|')].join(':'),
  async () => {
    if (!props.open) {
      destroySortable();
      return;
    }
    await nextTick();
    await initializeSortable();
  },
  {
    flush: 'post',
  },
);

onBeforeUnmount(() => {
  destroySortable();
});
</script>

<template>
  <Button @click="emit('update:open', true)">列设置</Button>

  <Drawer
    :open="open"
    placement="right"
    title="列设置"
    :width="420"
    @close="closeDrawer"
  >
    <div class="mb-4 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-500">
      列宽请直接拖拽表头边缘调整。排序会按固定列分组生效。
    </div>

    <div ref="sortableContainerRef" class="space-y-3">
      <div
        v-for="column in columns"
        :key="column.key"
        class="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-3 py-3 shadow-sm"
      >
        <button
          type="button"
          class="w-5 shrink-0 text-xs font-semibold tracking-widest text-slate-400 transition hover:text-slate-600"
          :class="column.draggable ? 'cursor-grab active:cursor-grabbing' : 'cursor-not-allowed opacity-40'"
          :data-drag-handle="column.draggable ? 'true' : undefined"
          :disabled="!column.draggable"
          :tabindex="column.draggable ? 0 : -1"
        >
          ::
        </button>

        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <div class="truncate text-sm font-medium text-slate-900">
              {{ column.title }}
            </div>
            <Tag v-if="column.fixed === 'left'" color="blue">固定左侧</Tag>
            <Tag v-else-if="column.fixed === 'right'" color="cyan">固定右侧</Tag>
            <Tag v-if="column.system" color="gold">系统列</Tag>
          </div>
          <div class="mt-1 text-xs text-slate-400">{{ column.key }}</div>
        </div>

        <Select
          class="w-28 shrink-0"
          :disabled="column.disableFixed"
          :options="fixedOptions"
          :value="column.fixed || 'none'"
          @update:value="(value) => onFixedChange(column.key, String(value))"
        />

        <Switch
          :checked="column.visible"
          :disabled="column.disableVisibility"
          @update:checked="(value) => onVisibilityChange(column.key, Boolean(value))"
        />
      </div>
    </div>

    <template #footer>
      <div class="flex items-center justify-between">
        <Button @click="emit('reset')">恢复默认</Button>
        <Button @click="closeDrawer">关闭</Button>
      </div>
    </template>
  </Drawer>
</template>
