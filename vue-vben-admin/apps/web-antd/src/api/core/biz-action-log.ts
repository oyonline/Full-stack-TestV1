import { requestClient } from '#/api/request';

export interface BizActionLogItem {
  logId: number;
  moduleKey: string;
  businessType: string;
  businessId: string;
  businessNo?: string;
  action: string;
  fromStatus?: string;
  toStatus?: string;
  remark?: string;
  operatorId: number;
  operatorName: string;
  createdAt: string;
}

export interface GetBizActionLogPageParams {
  pageIndex?: number;
  pageSize?: number;
  moduleKey: string;
  businessType: string;
  businessId: string;
}

interface BizActionLogPageResult {
  list: BizActionLogItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export async function getBizActionLogPage(
  params: GetBizActionLogPageParams,
): Promise<BizActionLogPageResult> {
  return requestClient.get<BizActionLogPageResult>('/v1/platform/biz-action-log', {
    params: {
      pageIndex: params.pageIndex ?? 1,
      pageSize: params.pageSize ?? 100,
      moduleKey: params.moduleKey,
      businessType: params.businessType,
      businessId: params.businessId,
    },
  });
}
