<script lang="ts" setup>
import { computed, ref } from 'vue';

import { message } from 'ant-design-vue';

import { uploadPlatformAttachment } from '#/api/core';

const props = withDefaults(
  defineProps<{
    businessId?: number | string;
    disabled?: boolean;
    max?: number;
    modelValue: string[];
  }>(),
  {
    businessId: undefined,
    disabled: false,
    max: 3,
  },
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void;
}>();

const fileInputRef = ref<HTMLInputElement | null>(null);
const uploading = ref(false);

const list = computed<string[]>(() => props.modelValue ?? []);
const canAdd = computed(() => list.value.length < props.max && !props.disabled);

function triggerPick() {
  if (!canAdd.value || uploading.value) return;
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
      businessType: 'spu-detail-image',
      moduleKey: 'admin',
    });
    const url = result.storagePath ? `/${result.storagePath}` : '';
    if (!url) {
      throw new Error('上传成功但未返回存储路径');
    }
    emit('update:modelValue', [...list.value, url]);
    message.success('详情图上传成功');
  } catch (error) {
    message.error((error as Error)?.message || '详情图上传失败');
  } finally {
    uploading.value = false;
  }
}

function removeAt(index: number) {
  if (props.disabled) return;
  const next = [...list.value];
  next.splice(index, 1);
  emit('update:modelValue', next);
}
</script>

<template>
  <div>
    <input
      ref="fileInputRef"
      data-testid="spu-detail-image-input"
      type="file"
      accept="image/*"
      class="hidden"
      @change="onFileChange"
    />
    <div class="flex flex-wrap items-start gap-3">
      <div
        v-for="(url, idx) in list"
        :key="`${url}-${idx}`"
        class="relative h-24 w-24 overflow-hidden rounded-sm border border-slate-200"
      >
        <img :src="url" class="size-full object-cover" alt="detail" />
        <button
          v-if="!disabled"
          type="button"
          class="absolute right-1 top-1 flex h-5 w-5 items-center justify-center rounded-full bg-black/60 text-xs leading-none text-white"
          @click="removeAt(idx)"
        >
          ×
        </button>
      </div>
      <button
        v-if="canAdd"
        type="button"
        :disabled="uploading"
        class="flex h-24 w-24 cursor-pointer flex-col items-center justify-center rounded-sm border border-dashed border-slate-300 bg-slate-50 text-xs text-slate-500 hover:border-slate-400 disabled:cursor-not-allowed"
        @click="triggerPick"
      >
        <span class="text-xl leading-none">+</span>
        <span class="mt-1">{{ uploading ? '上传中…' : '添加图片' }}</span>
      </button>
    </div>
    <div class="mt-2 text-xs text-slate-400">
      最多 {{ max }} 张，jpg/png/webp（已上传 {{ list.length }} / {{ max }}）
    </div>
  </div>
</template>
