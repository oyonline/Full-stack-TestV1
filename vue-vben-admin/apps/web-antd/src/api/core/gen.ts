import { requestClient } from '#/api/request';

export interface DbTableItem {
  tableName: string;
  engine?: string;
  tableRows?: string;
  tableCollation?: string;
  createTime?: string;
  updateTime?: string;
  tableComment?: string;
}

export interface SysTableColumnItem {
  columnId: number;
  tableId: number;
  columnName?: string;
  columnComment?: string;
  columnType?: string;
  goType?: string;
  goField?: string;
  jsonField?: string;
  isPk?: string;
  isRequired?: string;
  isInsert?: string;
  isEdit?: string;
  isList?: string;
  isQuery?: string;
  queryType?: string;
  htmlType?: string;
  dictType?: string;
  sort?: number;
  remark?: string;
  fkTableName?: string;
  fkLabelId?: string;
  fkLabelName?: string;
  isIncrement?: string;
  increment?: boolean;
  insert?: boolean;
  edit?: boolean;
  query?: boolean;
  required?: boolean;
  pk?: boolean;
}

export interface SysTableItem {
  tableId: number;
  tableName?: string;
  tableComment?: string;
  className?: string;
  packageName?: string;
  moduleName?: string;
  moduleFrontName?: string;
  businessName?: string;
  functionName?: string;
  functionAuthor?: string;
  tplCategory?: string;
  remark?: string;
  options?: string;
  crud?: boolean;
  tree?: boolean;
  isDataScope?: number;
  isActions?: number;
  isAuth?: number;
  isLogicalDelete?: string;
  logicalDelete?: boolean;
  logicalDeleteColumn?: string;
  treeCode?: string;
  treeParentCode?: string;
  treeName?: string;
  pkColumn?: string;
  pkGoField?: string;
  pkJsonField?: string;
  columns?: SysTableColumnItem[];
}

export interface PageResult<T> {
  list: T[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

interface PageData<T> {
  list?: T[];
  count?: number;
  pageIndex?: number;
  pageSize?: number;
}

export interface GetDbTablePageParams {
  pageIndex: number;
  pageSize: number;
  tableName?: string;
}

export interface GetSysTablePageParams {
  pageIndex: number;
  pageSize: number;
  tableName?: string;
  tableComment?: string;
}

export interface SysTableDetailResult {
  info: SysTableItem;
  list: SysTableColumnItem[];
}

export type PreviewResult = Record<string, string>;

export interface UpdateSysTablePayload extends SysTableItem {
  tableId: number;
  columns: SysTableColumnItem[];
}

export interface FormSchemaJson {
  drawingItems: unknown[];
  formConf: Record<string, unknown>;
  version?: string;
  idGlobal?: string;
  treeNodeId?: string;
}

export async function getDbTablePage(
  params: GetDbTablePageParams,
): Promise<PageResult<DbTableItem>> {
  const data = await requestClient.get<PageData<DbTableItem>>(
    '/v1/db/tables/page',
    { params },
  );
  return {
    list: data?.list ?? [],
    count: data?.count ?? 0,
    pageIndex: data?.pageIndex ?? params.pageIndex,
    pageSize: data?.pageSize ?? params.pageSize,
  };
}

export async function getSysTablePage(
  params: GetSysTablePageParams,
): Promise<PageResult<SysTableItem>> {
  const data = await requestClient.get<PageData<SysTableItem>>(
    '/v1/sys/tables/page',
    { params },
  );
  return {
    list: data?.list ?? [],
    count: data?.count ?? 0,
    pageIndex: data?.pageIndex ?? params.pageIndex,
    pageSize: data?.pageSize ?? params.pageSize,
  };
}

export async function createSysTables(tables: string[]): Promise<void> {
  await requestClient.post('/v1/sys/tables/info', undefined, {
    params: {
      tables: tables.join(','),
    },
  });
}

export async function getSysTableDetail(
  tableId: number,
): Promise<SysTableDetailResult> {
  const data = await requestClient.get<Partial<SysTableDetailResult>>(
    `/v1/sys/tables/info/${tableId}`,
  );
  return {
    info: (data?.info ?? {}) as SysTableItem,
    list: Array.isArray(data?.list) ? data.list : [],
  };
}

export async function previewGeneratedCode(
  tableId: number,
): Promise<PreviewResult> {
  const data = await requestClient.get<PreviewResult>(`/v1/gen/preview/${tableId}`);
  return data ?? {};
}

export async function updateSysTable(
  payload: UpdateSysTablePayload,
): Promise<void> {
  await requestClient.put('/v1/sys/tables/info', payload);
}

export async function deleteSysTable(tableId: number): Promise<void> {
  await requestClient.delete(`/v1/sys/tables/info/${tableId}`);
}

export async function generateCodeToProject(tableId: number): Promise<void> {
  await requestClient.get(`/v1/gen/toproject/${tableId}`);
}

export async function generateMenuAndApi(tableId: number): Promise<void> {
  await requestClient.get(`/v1/gen/todb/${tableId}`);
}
