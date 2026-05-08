export type RoleDataScopeValue = '1' | '2' | '3' | '4' | '5';

export const DEFAULT_ROLE_DATA_SCOPE: RoleDataScopeValue = '1';
export const CUSTOM_ROLE_DATA_SCOPE: RoleDataScopeValue = '2';

export interface RoleDataScopeOption {
  description: string;
  label: string;
  value: RoleDataScopeValue;
}

export const ROLE_DATA_SCOPE_OPTIONS: RoleDataScopeOption[] = [
  {
    description: '用户可见所有业务模块的数据（适合 admin / 财务总监）',
    label: '全部数据权限',
    value: '1',
  },
  {
    description:
      '按角色关联的部门集合 → 该角色用户能看到关联部门内同事 create 的数据',
    label: '自定义数据权限',
    value: '2',
  },
  {
    description: '用户只能看到与自己同部门同事 create 的数据',
    label: '本部门数据权限',
    value: '3',
  },
  {
    description: '用户能看到自己部门 + 下级部门同事 create 的数据',
    label: '本部门及以下数据权限',
    value: '4',
  },
  {
    description: '用户只能看到自己 create 的数据',
    label: '仅本人数据权限',
    value: '5',
  },
];

const ROLE_DATA_SCOPE_LABEL_MAP = new Map(
  ROLE_DATA_SCOPE_OPTIONS.map((item) => [item.value, item.label]),
);

export function normalizeRoleDataScope(
  value?: null | string,
): RoleDataScopeValue {
  const normalized = `${value ?? ''}`.trim() as RoleDataScopeValue;
  return ROLE_DATA_SCOPE_LABEL_MAP.has(normalized)
    ? normalized
    : DEFAULT_ROLE_DATA_SCOPE;
}

export function getRoleDataScopeLabel(value?: null | string) {
  return ROLE_DATA_SCOPE_LABEL_MAP.get(normalizeRoleDataScope(value));
}
