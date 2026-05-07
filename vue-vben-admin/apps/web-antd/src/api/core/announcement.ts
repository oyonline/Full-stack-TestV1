import { requestClient } from '#/api/request';

/** 公告项（与后端 dto.AnnouncementListItem 对齐） */
export interface AnnouncementItem {
  announcementId: number;
  title: string;
  content?: string;
  coverImageUrl?: string;
  status: number; // 1=草稿 2=已发布 3=已下线
  isTop: number; // 0=否 1=是
  topSort: number;
  publishAt?: null | string;
  expireAt?: null | string;
  creatorId?: number;
  remark?: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
  // 派生字段
  deptIds?: number[];
  isRead?: boolean;
  readCount?: number;
}

export interface AnnouncementPageResult {
  list: AnnouncementItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface GetAnnouncementPageParams {
  pageIndex?: number;
  pageSize?: number;
  title?: string;
  status?: number;
  isTop?: number;
  /** 1 = 仅返回当前生效中的（按 publish_at/expire_at 过滤） */
  onlyValid?: number;
  /** 1 = 仅返回当前用户可见（按部门 scope 过滤） */
  onlyVisible?: number;
}

export interface CreateAnnouncementData {
  title: string;
  content?: string;
  coverImageUrl?: string;
  status?: number;
  isTop?: number;
  topSort?: number;
  publishAt?: null | string;
  expireAt?: null | string;
  deptIds?: number[];
  remark?: string;
}

export type UpdateAnnouncementData = CreateAnnouncementData;

/** GET /api/v1/announcement */
export async function getAnnouncementPage(
  params: GetAnnouncementPageParams = {},
): Promise<AnnouncementPageResult> {
  return requestClient.get<AnnouncementPageResult>('/v1/announcement', {
    params,
  });
}

/** GET /api/v1/announcement/:id */
export async function getAnnouncementDetail(
  id: number,
): Promise<AnnouncementItem> {
  return requestClient.get<AnnouncementItem>(`/v1/announcement/${id}`);
}

/** POST /api/v1/announcement */
export async function createAnnouncement(
  data: CreateAnnouncementData,
): Promise<number> {
  return requestClient.post('/v1/announcement', data);
}

/** PUT /api/v1/announcement/:id */
export async function updateAnnouncement(
  id: number,
  data: UpdateAnnouncementData,
): Promise<number> {
  return requestClient.put(`/v1/announcement/${id}`, data);
}

/** DELETE /api/v1/announcement */
export async function deleteAnnouncement(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/announcement', { data: { ids } });
}

/** POST /api/v1/announcement/:id/read（幂等） */
export async function markAnnouncementRead(id: number): Promise<number> {
  return requestClient.post(`/v1/announcement/${id}/read`);
}
