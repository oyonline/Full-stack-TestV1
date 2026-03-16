import { requestClient } from '#/api/request';

/** 参数配置列表项（与后端 models.SysConfig 对齐） */
export interface SysConfigItem {
  id: number;
  configName: string;
  configKey: string;
  configValue: string;
  configType: string;
  isFrontend: string;
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 参数配置分页响应（requestClient 解包后为 data 内容） */
export interface SysConfigPageResult {
  list: SysConfigItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 参数配置列表查询参数 */
interface GetConfigPageParams {
  pageIndex?: number;
  pageSize?: number;
  configName?: string;
  configKey?: string;
  configType?: string;
  isFrontend?: string;
}

/**
 * 获取参数配置分页列表
 * 对接 go-admin GET /api/v1/config
 */
export async function getConfigPage(
  params: GetConfigPageParams = {},
): Promise<SysConfigPageResult> {
  return requestClient.get<SysConfigPageResult>('/v1/config', { params });
}

/**
 * 获取参数配置详情
 * 对接 go-admin GET /api/v1/config/:id
 */
export async function getConfigDetail(id: number): Promise<SysConfigItem> {
  return requestClient.get<SysConfigItem>(`/v1/config/${id}`);
}

/** 新增参数配置请求体 */
interface CreateConfigData {
  configName: string;
  configKey: string;
  configValue: string;
  configType?: string;
  isFrontend?: string;
  remark?: string;
}

/**
 * 新增参数配置
 * 对接 go-admin POST /api/v1/config
 */
export async function createConfig(data: CreateConfigData): Promise<number> {
  return requestClient.post('/v1/config', data);
}

/** 更新参数配置请求体 */
interface UpdateConfigData {
  configName?: string;
  configKey?: string;
  configValue?: string;
  configType?: string;
  isFrontend?: string;
  remark?: string;
}

/**
 * 更新参数配置
 * 对接 go-admin PUT /api/v1/config/:id
 */
export async function updateConfig(
  id: number,
  data: UpdateConfigData,
): Promise<number> {
  return requestClient.put(`/v1/config/${id}`, data);
}

/**
 * 删除参数配置（支持批量）
 * 对接 go-admin DELETE /api/v1/config
 */
export async function deleteConfig(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/config', { data: { ids } });
}
