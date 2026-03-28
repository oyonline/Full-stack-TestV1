import { requestClient } from '#/api/request';

// ========== 字典类型 ==========

/** 字典类型列表项（与后端 models.SysDictType 对齐） */
export interface SysDictTypeItem {
  id: number;
  dictName: string;
  dictType: string;
  status: number;
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 字典类型分页响应 */
export interface SysDictTypePageResult {
  list: SysDictTypeItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 字典类型列表查询参数（与 dto.SysDictTypeGetPageReq form 对齐） */
interface GetDictTypePageParams {
  pageIndex?: number;
  pageSize?: number;
  dictName?: string;
  dictType?: string;
  status?: number;
}

/**
 * 获取字典类型分页列表
 * 对接 go-admin GET /api/v1/dict/type
 */
export async function getDictTypePage(
  params: GetDictTypePageParams = {},
): Promise<SysDictTypePageResult> {
  return requestClient.get<SysDictTypePageResult>('/v1/dict/type', { params });
}

/**
 * 获取字典类型详情
 * 对接 go-admin GET /api/v1/dict/type/:id
 */
export async function getDictTypeDetail(id: number): Promise<SysDictTypeItem> {
  return requestClient.get<SysDictTypeItem>(`/v1/dict/type/${id}`);
}

/**
 * 获取字典类型全部（供下拉等使用）
 * 对接 go-admin GET /api/v1/dict/type-option-select
 */
export async function getDictTypeAll(
  params?: { dictName?: string; dictType?: string },
): Promise<SysDictTypeItem[]> {
  const res = await requestClient.get<SysDictTypeItem[]>(
    '/v1/dict/type-option-select',
    { params: params ?? {} },
  );
  return Array.isArray(res) ? res : [];
}

/** 字典类型下拉选项（与 SysDictTypeItem 一致，供字典类型详情页等使用） */
export type DictTypeOption = SysDictTypeItem;

/**
 * 获取字典类型选项（供下拉等使用），与 getDictTypeAll 等价
 * 对接 go-admin GET /api/v1/dict/type-option-select
 */
export async function getDictTypeOptionSelect(
  params?: { dictName?: string; dictType?: string },
): Promise<DictTypeOption[]> {
  return getDictTypeAll(params);
}

/** 新增字典类型请求体（与 dto.SysDictTypeInsertReq 对齐） */
interface CreateDictTypeData {
  dictName: string;
  dictType: string;
  status?: number;
  remark?: string;
}

/**
 * 新增字典类型
 * 对接 go-admin POST /api/v1/dict/type
 */
export async function createDictType(data: CreateDictTypeData): Promise<number> {
  return requestClient.post('/v1/dict/type', data);
}

/** 更新字典类型请求体 */
interface UpdateDictTypeData {
  dictName?: string;
  dictType?: string;
  status?: number;
  remark?: string;
}

/**
 * 更新字典类型
 * 对接 go-admin PUT /api/v1/dict/type/:id
 */
export async function updateDictType(
  id: number,
  data: UpdateDictTypeData,
): Promise<number> {
  return requestClient.put(`/v1/dict/type/${id}`, data);
}

/**
 * 删除字典类型（支持批量）
 * 对接 go-admin DELETE /api/v1/dict/type body { ids }
 */
export async function deleteDictType(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/dict/type', { data: { ids } });
}

// ========== 字典数据（主要由字典类型详情页消费） ==========

/** 字典数据列表项（与后端 models.SysDictData 对齐） */
export interface SysDictDataItem {
  dictCode: number;
  dictSort: number;
  dictLabel: string;
  dictValue: string;
  dictType: string;
  cssClass: string;
  listClass: string;
  isDefault: string;
  status: number;
  default: string;
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 字典数据分页响应 */
export interface SysDictDataPageResult {
  list: SysDictDataItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 字典数据列表查询参数（与 dto.SysDictDataGetPageReq form 对齐） */
interface GetDictDataPageParams {
  pageIndex?: number;
  pageSize?: number;
  dictLabel?: string;
  dictValue?: string;
  dictType?: string;
  status?: string;
}

/**
 * 获取字典数据分页列表
 * 对接 go-admin GET /api/v1/dict/data
 */
export async function getDictDataPage(
  params: GetDictDataPageParams = {},
): Promise<SysDictDataPageResult> {
  return requestClient.get<SysDictDataPageResult>('/v1/dict/data', {
    params,
  });
}

/**
 * 获取字典数据详情
 * 对接 go-admin GET /api/v1/dict/data/:dictCode
 */
export async function getDictDataDetail(
  dictCode: number,
): Promise<SysDictDataItem> {
  return requestClient.get<SysDictDataItem>(`/v1/dict/data/${dictCode}`);
}

/** 新增字典数据请求体（与 dto.SysDictDataInsertReq 对齐） */
interface CreateDictDataData {
  dictSort?: number;
  dictLabel: string;
  dictValue: string;
  dictType: string;
  cssClass?: string;
  listClass?: string;
  isDefault?: string;
  status?: number;
  default?: string;
  remark?: string;
}

/**
 * 新增字典数据
 * 对接 go-admin POST /api/v1/dict/data
 */
export async function createDictData(
  data: CreateDictDataData,
): Promise<number> {
  return requestClient.post('/v1/dict/data', data);
}

/** 更新字典数据请求体 */
interface UpdateDictDataData {
  dictSort?: number;
  dictLabel?: string;
  dictValue?: string;
  dictType?: string;
  cssClass?: string;
  listClass?: string;
  isDefault?: string;
  status?: number;
  default?: string;
  remark?: string;
}

/**
 * 更新字典数据
 * 对接 go-admin PUT /api/v1/dict/data/:dictCode
 */
export async function updateDictData(
  dictCode: number,
  data: UpdateDictDataData,
): Promise<number> {
  return requestClient.put(`/v1/dict/data/${dictCode}`, data);
}

/**
 * 删除字典数据（支持批量）
 * 对接 go-admin DELETE /api/v1/dict/data body { ids }
 */
export async function deleteDictData(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/dict/data', { data: { ids } });
}
