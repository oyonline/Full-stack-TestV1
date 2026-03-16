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
 * 短 key 到 Iconify 的映射表
 * 覆盖 go-admin 历史数据库中高频使用的短 key，映射到语义接近的 ant-design 图标
 */
const ICON_SHORT_KEY_MAP: Record<string, string> = {
  // === 已验证的 PoC 映射 ===
  'api-server': 'ant-design:cloud-server-outlined', // 系统管理 -> 服务器/云图标
  'tree-table': 'ant-design:table-outlined', // 菜单管理 -> 表格图标
  user: 'ant-design:user-outlined', // 用户管理 -> 用户图标

  // === 新增高频映射 ===
  peoples: 'ant-design:team-outlined', // 角色管理 -> 团队/人群
  swagger: 'ant-design:api-outlined', // API/接口文档 -> API
  guide: 'ant-design:read-outlined', // 指南/文档 -> 阅读
  education: 'ant-design:read-outlined', // 教育/培训 -> 阅读
  logininfor: 'ant-design:history-outlined', // 登录日志 -> 历史记录
  skill: 'ant-design:tool-outlined', // 技能/技术 -> 工具
  bug: 'ant-design:bug-outlined', // Bug管理 -> 虫子
  build: 'ant-design:build-outlined', // 构建/编译 -> 构建
  code: 'ant-design:code-outlined', // 代码/生成 -> 代码
  log: 'ant-design:file-text-outlined', // 日志 -> 文档/文本
  pass: 'ant-design:key-outlined', // 密码/密钥 -> 钥匙
  job: 'ant-design:clock-circle-outlined', // 定时任务 -> 时钟
  'system-tools': 'ant-design:tool-outlined', // 系统工具 -> 工具
  'dev-tools': 'ant-design:experiment-outlined', // 开发工具 -> 实验
  'time-range': 'ant-design:clock-circle-outlined', // 时间范围 -> 时钟
  tree: 'ant-design:cluster-outlined', // 树形/部门树 -> 层级/树

  // === 服务监控 & 接口管理（后端实际返回的 key） ===
  druid: 'ant-design:monitor-outlined', // 服务监控
  'api-doc': 'ant-design:api-outlined', // 接口管理
};

/**
 * 将历史短 key icon 转换为 Iconify 格式
 * - 若值已含冒号，视为 Iconify 格式，原样返回
 * - 若在映射表中存在，返回对应 Iconify 值
 * - 其他未知短 key，原样返回（便于排查）
 */
function normalizeMenuIcon(icon?: string): string | undefined {
  if (!icon) return undefined;
  // 已含冒号，视为 Iconify 格式
  if (icon.includes(':')) return icon;
  // 映射表转换
  return ICON_SHORT_KEY_MAP[icon] || icon;
}

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
      icon: normalizeMenuIcon(node.icon), // PoC: 短 key 转 Iconify
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
