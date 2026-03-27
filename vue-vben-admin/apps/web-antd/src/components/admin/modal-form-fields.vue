<script lang="ts" setup>
import { computed } from 'vue';

import { Form, FormItem, Input, InputNumber, Select, TreeSelect } from 'ant-design-vue';

import type { AdminFormFieldSchema } from './modal-form';

const props = defineProps<{
  fields: AdminFormFieldSchema[];
  model: Record<string, any>;
}>();

const visibleFields = computed(() => props.fields.filter((field) => !field.hidden));

function getItemClass(field: AdminFormFieldSchema) {
  return [
    'mb-0',
    field.span === 2 ? 'md:col-span-2' : '',
    field.itemClass ?? '',
  ]
    .filter(Boolean)
    .join(' ');
}
</script>

<template>
  <Form layout="vertical" class="mt-4 grid gap-x-4 gap-y-4 md:grid-cols-2">
    <FormItem
      v-for="field in visibleFields"
      :key="field.field"
      :label="field.label"
      :required="field.required"
      :class="getItemClass(field)"
    >
      <Input
        v-if="field.component === 'input'"
        v-model:value="model[field.field]"
        :allow-clear="field.allowClear ?? true"
        :disabled="field.disabled"
        :placeholder="field.placeholder"
        v-bind="field.props"
      />
      <Input.TextArea
        v-else-if="field.component === 'textarea'"
        v-model:value="model[field.field]"
        :allow-clear="field.allowClear ?? true"
        :disabled="field.disabled"
        :placeholder="field.placeholder"
        :rows="field.rows ?? 2"
        v-bind="field.props"
      />
      <InputNumber
        v-else-if="field.component === 'input-number'"
        v-model:value="model[field.field]"
        :disabled="field.disabled"
        :min="field.min"
        class="w-full"
        v-bind="field.props"
      />
      <Select
        v-else-if="field.component === 'select'"
        v-model:value="model[field.field]"
        :allow-clear="field.allowClear"
        :disabled="field.disabled"
        :filter-option="field.filterOption"
        :loading="field.loading"
        :options="field.options"
        :placeholder="field.placeholder"
        :show-search="field.showSearch"
        class="w-full"
        v-bind="field.props"
      />
      <TreeSelect
        v-else-if="field.component === 'tree-select'"
        v-model:value="model[field.field]"
        :allow-clear="field.allowClear"
        :disabled="field.disabled"
        :loading="field.loading"
        :placeholder="field.placeholder"
        class="w-full"
        v-bind="field.props"
      />
    </FormItem>
  </Form>
</template>
