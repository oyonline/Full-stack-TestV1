<script lang="ts" setup>
import { computed, ref } from 'vue';

import { Button, message } from 'ant-design-vue';

import { uploadPlatformAttachment } from '#/api/core';

const props = withDefaults(
  defineProps<{
    businessId?: number | string;
    businessType?: string;
    disabled?: boolean;
    hint?: string;
    modelValue: string;
    moduleKey?: string;
  }>(),
  {
    businessId: undefined,
    businessType: 'announcement-cover',
    disabled: false,
    hint: '建议 1200×400，jpg/png/webp',
    moduleKey: 'admin',
  },
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const fileInputRef = ref<HTMLInputElement | null>(null);
const uploading = ref(false);

const previewUrl = computed(() => props.modelValue || '');

function triggerPick() {
  if (props.disabled || uploading.value) return;
  fileInputRef.value?.click();
}

async function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  target.value = '';
  if (!file) return;
  if (!file.type.startsWith('image/')) {
    message.error('仅支持图片格式');
    return;
  }
  uploading.value = true;
  try {
    const businessId =
      props.businessId === undefined || props.businessId === null
        ? '0'
        : String(props.businessId);
    const result = await uploadPlatformAttachment({
      file,
      businessId,
      businessType: props.businessType,
      moduleKey: props.moduleKey,
    });
    const url = result.storagePath ? `/${result.storagePath}` : '';
    if (!url) {
      throw new Error('上传成功但未返回存储路径');
    }
    emit('update:modelValue', url);
    message.success('封面上传成功');
  } catch (error) {
    message.error((error as Error)?.message || '封面上传失败');
  } finally {
    uploading.value = false;
  }
}

function clearCover() {
  emit('update:modelValue', '');
}
</script>

<template>
  <div class="flex items-start gap-3">
    <input
      ref="fileInputRef"
      data-testid="announcement-cover-input"
      type="file"
      accept="image/*"
      class="hidden"
      @change="onFileChange"
    />
    <div
      class="flex-center h-24 w-40 overflow-hidden rounded-sm border border-dashed border-slate-300 bg-slate-50"
    >
      <img
        v-if="previewUrl"
        :src="previewUrl"
        class="size-full object-cover"
        alt="cover"
      />
      <span v-else class="text-xs text-slate-400">未上传封面</span>
    </div>
    <div class="flex flex-col gap-2">
      <Button
        data-testid="announcement-cover-upload"
        type="primary"
        size="small"
        :loading="uploading"
        :disabled="disabled"
        @click="triggerPick"
      >
        {{ previewUrl ? '更换封面' : '上传封面' }}
      </Button>
      <Button
        v-if="previewUrl"
        data-testid="announcement-cover-clear"
        size="small"
        :disabled="disabled || uploading"
        @click="clearCover"
      >
        清除
      </Button>
      <div class="text-xs text-slate-400">{{ hint }}</div>
    </div>
  </div>
</template>
