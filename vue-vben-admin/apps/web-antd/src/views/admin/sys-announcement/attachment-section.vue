<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';

import type { PlatformAttachmentItem } from '#/api/core';

import { computed, ref, watch } from 'vue';

import { Button, message, Popconfirm, Table, Tag } from 'ant-design-vue';

import {
  deletePlatformAttachment,
  downloadPlatformAttachment,
  getPlatformAttachmentPage,
  uploadPlatformAttachment,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

const props = defineProps<{
  announcementId?: null | number;
}>();

const listLoading = ref(false);
const uploadLoading = ref(false);
const attachmentList = ref<PlatformAttachmentItem[]>([]);
const fileInputRef = ref<HTMLInputElement | null>(null);

const canOperate = computed(
  () => props.announcementId !== undefined && props.announcementId !== null,
);

const businessId = computed(() =>
  props.announcementId === undefined || props.announcementId === null
    ? ''
    : String(props.announcementId),
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
  if (!bytes) return '0 B';
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
      businessId: businessId.value,
      businessType: 'announcement',
      moduleKey: 'admin',
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
  if (!file || !canOperate.value) return;

  uploadLoading.value = true;
  try {
    await uploadPlatformAttachment({
      file,
      businessId: businessId.value,
      businessType: 'announcement',
      moduleKey: 'admin',
    });
    message.success('附件上传成功');
    await loadAttachments();
  } catch (error) {
    message.error((error as Error)?.message || '附件上传失败');
  } finally {
    uploadLoading.value = false;
  }
}

async function handleDownload(
  record: PlatformAttachmentItem | Record<string, any>,
) {
  try {
    await downloadPlatformAttachment(record as PlatformAttachmentItem);
    message.success('附件下载已开始');
  } catch (error) {
    message.error((error as Error)?.message || '下载附件失败');
  }
}

async function handleDelete(
  record: PlatformAttachmentItem | Record<string, any>,
) {
  try {
    await deletePlatformAttachment(
      (record as PlatformAttachmentItem).attachmentId,
    );
    message.success('附件删除成功');
    await loadAttachments();
  } catch (error) {
    message.error((error as Error)?.message || '附件删除失败');
  }
}

watch(
  () => props.announcementId,
  () => {
    void loadAttachments();
  },
  { immediate: true },
);
</script>

<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="text-sm font-medium text-slate-700">附件</div>
      <div class="flex items-center gap-2">
        <input
          ref="fileInputRef"
          data-testid="announcement-attachment-input"
          type="file"
          class="hidden"
          @change="handleFileChange"
        />
        <Button
          data-testid="announcement-attachment-upload-trigger"
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
              data-testid="announcement-attachment-download"
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
              <AdminActionButton
                data-testid="announcement-attachment-delete"
                size="small"
                type="link"
                danger
                roles="admin"
                :disabled="!canOperate"
              >
                删除
              </AdminActionButton>
            </Popconfirm>
          </div>
        </template>
      </template>

      <template #emptyText>
        <div class="flex-col-center gap-2 py-4 text-slate-400">
          <Tag color="default">无附件</Tag>
          <span v-if="canOperate">当前公告还没有上传附件</span>
          <span v-else>请先创建并保存公告，再上传附件</span>
        </div>
      </template>
    </Table>
  </div>
</template>
