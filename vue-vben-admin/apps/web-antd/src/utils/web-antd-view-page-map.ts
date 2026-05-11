/**
 * 统一构建 web-antd 项目的视图 pageMap，供路由(access.ts)与菜单管理(sys-menu)共享，
 * 避免两边各自 import.meta.glob 导致路径集合不一致。
 */

/** 与 access.ts 的 normalizeViewPath 一致：去 ./ ../、补前导 /、去 /views 前缀 */
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

const pageMap = import.meta.glob('#/views/**/*.vue');
const validViewPathSet = buildValidViewPathSet(pageMap);

export { buildValidViewPathSet, normalizeViewPath, pageMap, validViewPathSet };
