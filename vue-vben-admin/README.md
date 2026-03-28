# Full-stack-TestV1 前端说明

这个目录是当前工作区的前端子项目，基于 Vue 3 + Vben Admin 定制。

## 以哪些文档为准

- 工作区总说明：[README.md](/Users/linshen/Desktop/Full-stack-TestV1/README.md)
- 脚手架使用说明：[docs/scaffold-guide.md](/Users/linshen/Desktop/Full-stack-TestV1/docs/scaffold-guide.md)
- 项目记忆入口：[.ai-memory/README.md](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/README.md)

## 常用命令

安装依赖：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin
pnpm install
```

启动前端：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

类型检查：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin
pnpm check:type
```

本地生产构建：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm build:local
```

Playwright 冒烟：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```

## 当前前端约定

- 后台 `CRUD / 树表 / 日志` 页面使用紧凑母版，不再保留大页头。
- `参数设置` 是唯一正式系统设置入口。
- 用户支持多角色：
  - 菜单按主角色
  - 按钮和接口权限按并集
- 菜单最终以数据库 `sys_menu` 为准，前端不再长期维护旧路径兼容。
- 用户字段前端展示统一口径为：
  - `username` 显示为 `登录账号`
  - `realName` / `nickName` 显示为 `姓名`
  - 这只影响前端文案，不改字段名、接口字段名和数据库列名
- 字母头像的文字大小占比统一由底层 `VbenAvatar` 控制，业务页面不要再单独覆写字号。

## 不再以这些内容为准

这个子项目原始上游 README 中关于：

- 官方 demo 账号
- 上游部署方式
- 原始前端结构

都不代表当前工作区的真实运行方式。当前请以本 README、根 README 和 `.ai-memory/` 为准。
