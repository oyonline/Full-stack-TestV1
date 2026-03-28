export type RoleDataScopeValue = '1' | '2' | '3' | '4' | '5';

export const DEFAULT_ROLE_DATA_SCOPE: RoleDataScopeValue = '1';
export const CUSTOM_ROLE_DATA_SCOPE: RoleDataScopeValue = '2';

export const ROLE_DATA_SCOPE_OPTIONS: Array<{
  label: string;
  value: RoleDataScopeValue;
}> = [
  { label: '全部数据', value: '1' },
  { label: '自定义数据权限', value: '2' },
  { label: '本部门数据', value: '3' },
  { label: '本部门及以下数据', value: '4' },
  { label: '仅本人数据', value: '5' },
];

const ROLE_DATA_SCOPE_LABEL_MAP = new Map(
  ROLE_DATA_SCOPE_OPTIONS.map((item) => [item.value, item.label]),
);

export function normalizeRoleDataScope(
  value?: string | null,
): RoleDataScopeValue {
  const normalized = `${value ?? ''}`.trim() as RoleDataScopeValue;
  return ROLE_DATA_SCOPE_LABEL_MAP.has(normalized)
    ? normalized
    : DEFAULT_ROLE_DATA_SCOPE;
}

export function getRoleDataScopeLabel(value?: string | null) {
  return ROLE_DATA_SCOPE_LABEL_MAP.get(normalizeRoleDataScope(value));
}
