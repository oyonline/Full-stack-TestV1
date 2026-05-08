<script lang="ts" setup>
import type {
  IDomEditor,
  IEditorConfig,
  IToolbarConfig,
} from '@wangeditor/editor';

import { onBeforeUnmount, shallowRef } from 'vue';

import { Editor, Toolbar } from '@wangeditor/editor-for-vue';
import { message } from 'ant-design-vue';

import { uploadPlatformAttachment } from '#/api/core';

import '@wangeditor/editor/dist/css/style.css';

const props = withDefaults(
  defineProps<{
    height?: number | string;
    modelValue: string;
    placeholder?: string;
    uploadBusinessType?: string;
    uploadModuleKey?: string;
  }>(),
  {
    height: 320,
    placeholder: '请输入正文…',
    uploadBusinessType: 'announcement-inline',
    uploadModuleKey: 'admin',
  },
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const editorRef = shallowRef<IDomEditor>();

type InsertFn = (url: string, alt?: string, href?: string) => void;

const toolbarConfig: Partial<IToolbarConfig> = {
  excludeKeys: ['fullScreen'],
};

const editorConfig: Partial<IEditorConfig> = {
  placeholder: props.placeholder,
  MENU_CONF: {
    uploadImage: {
      maxFileSize: 10 * 1024 * 1024,
      allowedFileTypes: ['image/*'],
      async customUpload(file: File, insertFn: InsertFn) {
        try {
          const result = await uploadPlatformAttachment({
            file,
            businessId: '0',
            businessType: props.uploadBusinessType,
            moduleKey: props.uploadModuleKey,
          });
          const url = result.storagePath ? `/${result.storagePath}` : '';
          if (!url) {
            throw new Error('上传成功但未返回存储路径');
          }
          insertFn(url, file.name, url);
        } catch (error) {
          message.error((error as Error)?.message || '图片上传失败');
        }
      },
    },
  },
};

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor;
}

function handleUpdate(value: string) {
  emit('update:modelValue', value);
}

onBeforeUnmount(() => {
  const editor = editorRef.value;
  if (!editor) return;
  editor.destroy();
});

const toolbarStyle = { borderBottom: '1px solid #e5e7eb' };
const editorBoxStyle = {
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
  overflowY: 'auto' as const,
};
</script>

<template>
  <div class="overflow-hidden rounded-sm border border-slate-300 bg-white">
    <Toolbar
      :default-config="toolbarConfig"
      :editor="editorRef"
      :style="toolbarStyle"
      mode="default"
    />
    <Editor
      :default-config="editorConfig"
      :model-value="modelValue"
      :style="editorBoxStyle"
      mode="default"
      @on-created="handleCreated"
      @update:model-value="handleUpdate"
    />
  </div>
</template>
