<script lang="ts" setup>
import { computed, ref } from 'vue';

import { Button, message } from 'ant-design-vue';

import { uploadPlatformAttachment } from '#/api/core';

const props = defineProps<{
  businessId?: number | string;
  disabled?: boolean;
  modelValue: string;
}>();

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
      businessType: 'sku-brand-logo',
      moduleKey: 'admin',
    });
    const url = result.storagePath ? `/${result.storagePath}` : '';
    if (!url) {
      throw new Error('上传成功但未返回存储路径');
    }
    emit('update:modelValue', url);
    message.success('Logo 上传成功');
  } catch (error) {
    message.error((error as Error)?.message || 'Logo 上传失败');
  } finally {
    uploading.value = false;
  }
}

function clearLogo() {
  emit('update:modelValue', '');
}
</script>

<template>
  <div class="flex items-start gap-3">
    <input
      ref="fileInputRef"
      data-testid="sku-brand-logo-input"
      type="file"
      accept="image/*"
      class="hidden"
      @change="onFileChange"
    />
    <div
      class="flex-center size-24 overflow-hidden rounded-sm border border-dashed border-slate-300 bg-slate-50"
    >
      <img
        v-if="previewUrl"
        :src="previewUrl"
        class="size-full object-contain"
        alt="logo"
      />
      <span v-else class="text-xs text-slate-400">未上传 Logo</span>
    </div>
    <div class="flex flex-col gap-2">
      <Button
        data-testid="sku-brand-logo-upload"
        type="primary"
        size="small"
        :loading="uploading"
        :disabled="disabled"
        @click="triggerPick"
      >
        {{ previewUrl ? '更换 Logo' : '上传 Logo' }}
      </Button>
      <Button
        v-if="previewUrl"
        data-testid="sku-brand-logo-clear"
        size="small"
        :disabled="disabled || uploading"
        @click="clearLogo"
      >
        清除
      </Button>
      <div class="text-xs text-slate-400">建议正方形，jpg/png/webp</div>
    </div>
  </div>
</template>
