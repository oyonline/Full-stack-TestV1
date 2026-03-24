# ISSUE-004: 左侧导航为空

## 问题描述
后端返回4个菜单，接口请求成功，但左侧导航不显示，页面进入404。

## 现象
- 后端 `/api/v1/menurole` 返回 4 个菜单 ✅
- `getAllMenusApi()` 获取数据正常 ✅
- `generateAccess` 构建列表正常 ✅
- `@vben/access` 的 `generateMenus()` 过滤后返回空数组 ❌
- localStorage.accessMenus 为空 ❌

## 根因分析

后端 `visible` 字段语义与前端判断逻辑相反：

| 后端 `visible` | 实际含义 | 前端原判断 (`=== '0'`) | 结果 |
|---------------|---------|---------------------|------|
| `'0'` | 显示菜单 | `hideInMenu: true` | 被隐藏 ❌ |
| `'1'` | 隐藏菜单 | `hideInMenu: false` | 显示 |

**关键代码位置：** `src/router/access.ts` 第240行

```typescript
// 原代码（错误）
hideInMenu: node.visible === '0'
```

后端返回的所有菜单 `visible: '0'`，表示"显示"，但前端判断为 `hideInMenu: true`，全部被过滤。

## 解决方案

**修改文件：** `apps/web-antd/src/router/access.ts`

```typescript
// 修复前（错误）
hideInMenu: node.visible === '0'

// 修复后（正确）
hideInMenu: node.visible !== '0'
```

## 经验总结

1. **后端字段语义必须明确文档化**：`visible` 字段的 `'0'` 和 `'1'` 分别代表什么
2. **布尔值判断要双向验证**：`===` 和 `!==` 都要考虑
3. **添加日志追踪中间状态**：`visible` → `hideInMenu` → `show`
4. **数据库初始数据与代码逻辑要一致**

## 相关文件

- 修复位置：`src/router/access.ts` 第240行
- 日志位置：同文件 `mapSysMenuToRoute` 函数
- 过滤逻辑：`packages/utils/src/helpers/generate-menus.ts` 第64行

---

*解决时间：2026-03-23 18:03*
