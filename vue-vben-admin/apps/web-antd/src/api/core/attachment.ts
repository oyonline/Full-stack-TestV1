import { useAccessStore } from '@vben/stores';

import { requestClient } from '#/api/request';

export interface PlatformAttachmentItem {
  attachmentId: number;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo?: string;
  fileName: string;
  fileExt?: string;
  fileSize: number;
  contentType?: string;
  storageType: string;
  storagePath: string;
  uploaderId: number;
  uploaderName: string;
  createdAt?: string;
  updatedAt?: string;
}

interface PlatformAttachmentPageResult {
  list: PlatformAttachmentItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface GetPlatformAttachmentPageParams {
  pageIndex?: number;
  pageSize?: number;
  moduleKey: string;
  businessType: string;
  businessId: string;
}

export interface UploadPlatformAttachmentParams {
  file: File;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo?: string;
}

export async function getPlatformAttachmentPage(
  params: GetPlatformAttachmentPageParams,
): Promise<PlatformAttachmentPageResult> {
  return requestClient.get<PlatformAttachmentPageResult>('/v1/platform/attachments', {
    params: {
      pageIndex: params.pageIndex ?? 1,
      pageSize: params.pageSize ?? 100,
      moduleKey: params.moduleKey,
      businessType: params.businessType,
      businessId: params.businessId,
    },
  });
}

export async function uploadPlatformAttachment(
  params: UploadPlatformAttachmentParams,
): Promise<PlatformAttachmentItem> {
  const formData = new FormData();
  formData.append('file', params.file);
  formData.append('moduleKey', params.moduleKey);
  formData.append('businessType', params.businessType);
  formData.append('businessId', params.businessId);
  if (params.businessNo) {
    formData.append('businessNo', params.businessNo);
  }
  const accessStore = useAccessStore();
  const headers: HeadersInit = {};
  if (accessStore.accessToken) {
    headers.Authorization = `Bearer ${accessStore.accessToken}`;
  }

  const response = await fetch('/api/v1/platform/attachments/upload', {
    body: formData,
    headers,
    method: 'POST',
  });

  const body = (await response.json()) as {
    code?: number;
    data?: PlatformAttachmentItem;
    msg?: string;
  };
  if (body.code !== 200 || !body.data) {
    throw new Error(body.msg || '上传附件失败');
  }
  return body.data;
}

export async function deletePlatformAttachment(id: number): Promise<void> {
  return requestClient.delete<void>(`/v1/platform/attachments/${id}`);
}

function getDownloadFileName(headers: Headers, fallback: string) {
  const disposition = headers.get('content-disposition') || '';
  const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1]);
  }
  const plainMatch = disposition.match(/filename="?([^"]+)"?/i);
  if (plainMatch?.[1]) {
    return decodeURIComponent(plainMatch[1]);
  }
  return fallback;
}

export async function downloadPlatformAttachment(
  item: PlatformAttachmentItem,
): Promise<void> {
  const accessStore = useAccessStore();
  const headers: HeadersInit = {};
  if (accessStore.accessToken) {
    headers.Authorization = `Bearer ${accessStore.accessToken}`;
  }

  const response = await fetch(
    `/api/v1/platform/attachments/${item.attachmentId}/download`,
    {
      headers,
    },
  );
  if (!response.ok) {
    throw new Error('下载附件失败');
  }

  const blob = await response.blob();
  const fileName = getDownloadFileName(response.headers, item.fileName || 'attachment');
  const objectUrl = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = objectUrl;
  link.download = fileName;
  document.body.append(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(objectUrl);
}
