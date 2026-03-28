<script lang="ts" setup>
import { computed } from 'vue';

import { Page } from '@vben/common-ui';

interface Props {
  headerMode?: 'compact' | 'full' | 'none';
}

const props = withDefaults(defineProps<Props>(), {
  headerMode: 'none',
});

const showHeader = computed(() => props.headerMode !== 'none');
const showEyebrow = computed(() => props.headerMode === 'full');
const mergeHeaderExtraIntoFilters = computed(() => props.headerMode === 'none');
</script>

<template>
  <Page
    auto-content-height
    content-class="!bg-slate-50 !p-4 md:!p-6"
    :header-class="
      headerMode === 'compact'
        ? 'border-b border-slate-200 bg-white/95 backdrop-blur'
        : 'border-b border-slate-200 bg-white/90 backdrop-blur'
    "
  >
    <template v-if="showHeader" #title>
      <div :class="headerMode === 'compact' ? 'space-y-0.5' : 'space-y-1'">
        <div
          v-if="showEyebrow && $slots.eyebrow"
          class="text-xs font-medium tracking-[0.22em] text-slate-400 uppercase"
        >
          <slot name="eyebrow"></slot>
        </div>
        <div
          :class="
            headerMode === 'compact'
              ? 'text-lg font-semibold tracking-tight text-slate-900'
              : 'text-2xl font-semibold tracking-tight text-slate-900'
          "
        >
          <slot name="title"></slot>
        </div>
      </div>
    </template>

    <template v-if="showHeader" #description>
      <div
        :class="
          headerMode === 'compact'
            ? 'max-w-3xl text-sm leading-6 text-slate-500'
            : 'max-w-3xl text-sm leading-7 text-slate-500'
        "
      >
        <slot name="description"></slot>
      </div>
    </template>

    <template v-if="showHeader && $slots['header-extra']" #extra>
      <div class="flex flex-wrap items-center gap-2">
        <slot name="header-extra"></slot>
      </div>
    </template>

    <div class="space-y-4 md:space-y-5">
      <section
        v-if="$slots.filters || $slots['filter-actions'] || (mergeHeaderExtraIntoFilters && $slots['header-extra'])"
        class="app-radius-panel bg-white p-4 shadow-sm md:p-5"
      >
        <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
          <div class="min-w-0">
            <slot name="filters"></slot>
          </div>
          <div
            v-if="$slots['filter-actions'] || (mergeHeaderExtraIntoFilters && $slots['header-extra'])"
            class="flex flex-wrap items-center gap-2 xl:justify-end"
          >
            <slot name="filter-actions"></slot>
            <slot
              v-if="mergeHeaderExtraIntoFilters && $slots['header-extra']"
              name="header-extra"
            ></slot>
          </div>
        </div>
      </section>

      <section
        class="app-radius-panel bg-white p-4 shadow-sm md:p-5"
      >
        <div
          v-if="$slots.toolbar || $slots['toolbar-extra']"
          class="mb-3 flex flex-col gap-2 border-b border-slate-100 pb-3 md:flex-row md:items-center md:justify-between"
        >
          <div class="min-w-0">
            <slot name="toolbar"></slot>
          </div>
          <div
            v-if="$slots['toolbar-extra']"
            class="flex flex-wrap items-center gap-2 md:justify-end"
          >
            <slot name="toolbar-extra"></slot>
          </div>
        </div>
        <slot></slot>
      </section>
    </div>
  </Page>
</template>
