<script lang="ts" setup>
import { Page } from '@vben/common-ui';
</script>

<template>
  <Page
    auto-content-height
    content-class="!bg-slate-50 !p-4 md:!p-6"
    header-class="border-b border-slate-200 bg-white/90 backdrop-blur"
  >
    <template #title>
      <div class="space-y-1">
        <div
          v-if="$slots.eyebrow"
          class="text-xs font-medium tracking-[0.22em] text-slate-400 uppercase"
        >
          <slot name="eyebrow"></slot>
        </div>
        <div class="text-2xl font-semibold tracking-tight text-slate-900">
          <slot name="title"></slot>
        </div>
      </div>
    </template>

    <template #description>
      <div class="max-w-3xl text-sm leading-7 text-slate-500">
        <slot name="description"></slot>
      </div>
    </template>

    <template v-if="$slots['header-extra']" #extra>
      <div class="flex flex-wrap items-center gap-2">
        <slot name="header-extra"></slot>
      </div>
    </template>

    <div class="space-y-4 md:space-y-5">
      <section
        v-if="$slots.filters || $slots['filter-actions']"
        class="app-radius-panel border border-slate-200 bg-white p-4 shadow-sm md:p-5"
      >
        <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
          <div class="min-w-0">
            <slot name="filters"></slot>
          </div>
          <div
            v-if="$slots['filter-actions']"
            class="flex flex-wrap items-center gap-2 xl:justify-end"
          >
            <slot name="filter-actions"></slot>
          </div>
        </div>
      </section>

      <section
        class="app-radius-panel border border-slate-200 bg-white p-4 shadow-sm md:p-5"
      >
        <div
          v-if="$slots.toolbar"
          class="mb-4 flex flex-col gap-3 border-b border-slate-100 pb-4 md:flex-row md:items-center md:justify-between"
        >
          <slot name="toolbar"></slot>
        </div>
        <slot></slot>
      </section>
    </div>
  </Page>
</template>
