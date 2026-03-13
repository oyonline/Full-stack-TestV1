import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
  RouteRecordStringComponent,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

/** 与 generate-routes-backend 一致的视图路径标准化（用于 component 与 pageMap key 匹配） */
function normalizeViewPath(path: string): string {
  const n = path.replace(/^(\.\/|\.\.\/)+/, '');
  const viewPath = n.startsWith('/') ? n : `/${n}`;
  return viewPath.replace(/^\/views/, '');
}

/** 从 glob 构建“有效视图路径”集合（不含 .vue），供 component 映射时查找 */
function buildValidViewPathSet(pageMap: Record<string, unknown>): Set<string> {
  const set = new Set<string>();
  for (const key of Object.keys(pageMap)) {
    const n = normalizeViewPath(key);
    set.add(n.endsWith('.vue') ? n.slice(0, -4) : n);
  }
  return set;
}

const NOT_FOUND_COMPONENT = '/_core/fallback/not-found';

/**
 * 将 go-admin 的 component 字符串映射为 Vben 可识别的 component（layoutMap 或 pageMap key 对应路径）
 * - Layout / BasicLayout / 目录无 component -> BasicLayout
 * - IFrameView -> IFrameView
 * - 其它：标准化后与当前项目 views 匹配；带/不带 /index、.vue 均尝试；无匹配则 not-found
 */
function mapComponent(
  backendComp: string | undefined,
  hasChildren: boolean,
  validViewPathSet: Set<string>,
): string {
  const comp = backendComp?.trim() ?? '';
  if (!comp && hasChildren) return 'BasicLayout';
  if (/^Layout$/i.test(comp) || /^BasicLayout$/i.test(comp)) return 'BasicLayout';
  if (/^IFrameView$/i.test(comp)) return 'IFrameView';

  let candidate = normalizeViewPath(comp);
  if (candidate.endsWith('.vue')) candidate = candidate.slice(0, -4);
  if (validViewPathSet.has(candidate)) return candidate;
  if (candidate.endsWith('/index')) {
    const withoutIndex = candidate.replace(/\/index$/, '');
    if (validViewPathSet.has(withoutIndex)) return withoutIndex;
  } else {
    if (validViewPathSet.has(candidate + '/index')) return candidate + '/index';
  }
  return NOT_FOUND_COMPONENT;
}

/** go-admin menurole 返回的菜单节点（仅映射用到的字段） */
interface GoAdminSysMenu {
  menuId?: number;
  menuName?: string;
  title?: string;
  icon?: string;
  path?: string;
  component?: string;
  sort?: number;
  menuType?: string;
  children?: GoAdminSysMenu[];
}

function pathToName(path: string): string {
  if (!path) return 'Route';
  return path.replace(/^\//, '').replace(/\//g, '-') || 'Route';
}

/** 将 go-admin SysMenu 树转为 RouteRecordStringComponent[]（最小映射） */
function mapSysMenuToRoute(
  node: GoAdminSysMenu,
  validViewPathSet: Set<string>,
): RouteRecordStringComponent {
  const name =
    (node.menuName && node.menuName.trim()) || pathToName(node.path ?? '');
  const path = node.path ?? '';
  const hasChildren = Array.isArray(node.children) && node.children.length > 0;
  const rawComp = node.component?.trim() || (hasChildren ? 'BasicLayout' : '');
  const component = mapComponent(rawComp, hasChildren, validViewPathSet);
  const children = hasChildren
    ? node.children!
        .filter((c) => c.menuType !== 'F')
        .map((c) => mapSysMenuToRoute(c, validViewPathSet))
    : undefined;

  return {
    name,
    path,
    component: component || 'BasicLayout',
    meta: {
      title: node.title ?? name,
      icon: node.icon,
      order: node.sort,
    },
    ...(children?.length ? { children } : {}),
  };
}

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
  const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => {
      message.loading({
        content: `${$t('common.loadingMenu')}...`,
        duration: 1.5,
      });
      const pageMap = import.meta.glob('../views/**/*.vue');
      const validViewPathSet = buildValidViewPathSet(pageMap);
      const raw = (await getAllMenusApi()) as unknown as GoAdminSysMenu[];
      if (!Array.isArray(raw)) return [];
      return raw
        .filter((n) => n.menuType !== 'F')
        .map((n) => mapSysMenuToRoute(n, validViewPathSet));
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap,
  });
}

export { generateAccess };
