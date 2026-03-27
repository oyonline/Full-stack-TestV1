<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import { Button, Popconfirm, Table, Tag, message } from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  deletePlatformAttachment,
  downloadPlatformAttachment,
  getPlatformAttachmentPage,
  uploadPlatformAttachment,
  type PlatformAttachmentItem,
} from '#/api/core';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

const props = defineProps<{
  businessId?: string;
  businessNo?: string;
  businessType?: string;
  moduleKey?: string;
}>();

const listLoading = ref(false);
const uploadLoading = ref(false);
const attachmentList = ref<PlatformAttachmentItem[]>([]);
const fileInputRef = ref<HTMLInputElement | null>(null);

const canOperate = computed(
  () =>
    !!props.moduleKey &&
    !!props.businessType &&
    !!props.businessId,
);

const columns: TableColumnType[] = [
  {
    title: '文件名',
    dataIndex: 'fileName',
    key: 'fileName',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '大小',
    dataIndex: 'fileSize',
    key: 'fileSize',
    width: 100,
    customRender: ({ text }: { text: number }) => formatFileSize(text),
  },
  {
    title: '上传人',
    dataIndex: 'uploaderName',
    key: 'uploaderName',
    width: 120,
    customRender: ({ text }: { text: string }) => renderAdminEmpty(text),
  },
  {
    title: '上传时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 170,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 160,
  },
];

function formatFileSize(size?: number) {
  const bytes = Number(size || 0);
  if (!bytes) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

async function loadAttachments() {
  if (!canOperate.value) {
    attachmentList.value = [];
    return;
  }
  listLoading.value = true;
  try {
    const result = await getPlatformAttachmentPage({
      businessId: props.businessId!,
      businessType: props.businessType!,
      moduleKey: props.moduleKey!,
      pageIndex: 1,
      pageSize: 100,
    });
    attachmentList.value = result.list || [];
  } catch (error) {
    message.error((error as Error)?.message || '加载附件失败');
    attachmentList.value = [];
  } finally {
    listLoading.value = false;
  }
}

function triggerUpload() {
  fileInputRef.value?.click();
}

async function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  target.value = '';
  if (!file || !canOperate.value) {
    return;
  }

  uploadLoading.value = true;
  try {
    await uploadPlatformAttachment({
      file,
      businessId: props.businessId!,
      businessNo: props.businessNo,
      businessType: props.businessType!,
      moduleKey: props.moduleKey!,
    });
    message.success('附件上传成功');
    await loadAttachments();
  } catch (error) {
    message.error((error as Error)?.message || '附件上传失败');
  } finally {
    uploadLoading.value = false;
  }
}

async function handleDownload(record: PlatformAttachmentItem | Record<string, any>) {
  try {
    await downloadPlatformAttachment(record as PlatformAttachmentItem);
    message.success('附件下载已开始');
  } catch (error) {
    message.error((error as Error)?.message || '下载附件失败');
  }
}

async function handleDelete(record: PlatformAttachmentItem | Record<string, any>) {
  try {
    await deletePlatformAttachment((record as PlatformAttachmentItem).attachmentId);
    message.success('附件删除成功');
    await loadAttachments();
  } catch (error) {
    message.error((error as Error)?.message || '附件删除失败');
  }
}

watch(
  () => [props.moduleKey, props.businessType, props.businessId] as const,
  () => {
    void loadAttachments();
  },
  { immediate: true },
);
</script>

<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="text-xs text-slate-500">
        当前流程实例关联的业务附件。第一版支持上传、下载、删除。
      </div>
      <div class="flex items-center gap-2">
        <input
          ref="fileInputRef"
          data-testid="workflow-attachment-input"
          type="file"
          class="hidden"
          @change="handleFileChange"
        />
        <Button
          data-testid="workflow-attachment-upload-trigger"
          type="primary"
          size="small"
          :disabled="!canOperate"
          :loading="uploadLoading"
          @click="triggerUpload"
        >
          上传附件
        </Button>
      </div>
    </div>

    <Table
      :columns="columns"
      :data-source="attachmentList"
      :loading="listLoading"
      :pagination="false"
      row-key="attachmentId"
      size="small"
      :scroll="{ x: 760 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <div class="flex items-center gap-2">
            <Button
              data-testid="workflow-attachment-download"
              size="small"
              type="link"
              :disabled="!canOperate"
              @click="handleDownload(record)"
            >
              下载
            </Button>
            <Popconfirm
              title="确认删除该附件吗？"
              ok-text="确认"
              cancel-text="取消"
              @confirm="handleDelete(record)"
            >
              <Button
                data-testid="workflow-attachment-delete"
                size="small"
                type="link"
                danger
                :disabled="!canOperate"
              >
                删除
              </Button>
            </Popconfirm>
          </div>
        </template>
      </template>

      <template #emptyText>
        <div class="flex flex-col items-center justify-center gap-2 py-4 text-slate-400">
          <Tag color="default">无附件</Tag>
          <span>当前业务单据还没有上传附件</span>
        </div>
      </template>
    </Table>
  </div>
</template>
