# Full-stack-TestV1 脚手架使用说明

这份文档只保留“当前脚手架怎么用”的信息。更细的长期约定请同时查看：

- [项目记忆入口](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/README.md)
- [项目关键决策](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/project-decisions.md)
- [前后端协议](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/backend-frontend-contracts.md)
- [排查手册](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/runbooks.md)
- [本地联调交接](/Users/linshen/Cursor/Full-stack-TestV1/docs/local-dev-handoff.md)

## 0. 先看真相源

当前项目协作时，默认按下面顺序判断事实：

1. 已提交代码
2. 已提交 migration / `config/db.sql` / 菜单种子
3. 当前运行链路能证明的行为
4. 本地 working tree 未提交改动
5. `.ai-memory/*`
6. README 与一般说明文档

这份文档的职责是“说明当前脚手架怎么推进”，不是用来宣布某条业务线已经完成。

## 1. 系统设置

- `参数设置` 是唯一正式系统设置入口。
- 系统名称、Logo、Logo 占位色、登录页标题/说明、全局界面设置都会持久化到后端。
- 页面启动时前端会从 `/api/v1/app-config` 拉取运行时配置。

## 2. 权限模型

- 用户支持多角色。
- 菜单按主角色生效。
- 按钮权限按所有角色并集生效。
- 接口权限当前按主角色 key 走 Casbin 校验，不是按多角色并集校验。
- 本期不做角色切换，也不做数据权限。

## 3. 开发工具

### 系统接口

- 使用当前后台原生系统接口页承接。

### 代码生成

当前已支持：

- 数据库表导入
- 已纳管表配置维护
- 独立编辑页修改表级/字段级配置
- 模板预览
- 从生成器移除

当前仍然隐藏：

- 生成到项目
- 生成菜单/API

推荐工作流：

1. 在 `开发工具 -> 代码生成` 搜索数据库表
2. 执行“导入到生成器”
3. 切到“生成配置”
4. 进入“编辑配置”
5. 保存后再做“模板预览”

### 表单构建

- 前端通过同域 `/form-generator/` 承接后端静态 form-generator。
- 当前桥接支持：
  - 下载 schema
  - 复制 schema
  - 导入 schema
- 第一版唯一输出格式是 `FormSchemaJson`。

最少结构：

```json
{
  "drawingItems": [],
  "formConf": {},
  "idGlobal": "100",
  "treeNodeId": "100",
  "version": "1.1"
}
```

## 4. 菜单与页面映射

- 当前菜单最终以数据库 `sys_menu` 为真相源。
- 新增菜单时，`component` 必须对齐 `src/views` 下真实页面，或使用显式承接组件：
  - `RouteView`
  - `IFrameView`
- 父级分组菜单不要再映射到 `BasicLayout`。

## 5. 后台页面母版

- 后台 `CRUD / 树表 / 日志` 页面使用紧凑母版：
  - 筛选操作区
  - 轻量列表头
  - 数据区
- 页面内局部 `刷新` 按钮默认取消。
- `新增` 等主操作下沉到筛选操作区。
- `参数设置 / 表单构建 / 代码生成编辑页 / 服务监控 / 首页类` 保留压缩版轻标题条。

### 标准列表列配置

- 路由级标准列表页统一通过 `useAdminTableColumns` 接入列个性化配置。
- 列设置 UI 统一使用 `AdminTableColumnSettings`，入口通常挂在 `PageShell` 的 `toolbar-extra` 槽位。
- 页面原先直接传给表格的静态 `columns`，改为使用 composable 返回的 `tableColumns`。
- 页面原先手写的 `scroll.x`，改为使用 composable 返回的 `scrollX`。
- 表头列宽拖拽后会自动记住，不需要额外保存按钮。

接入要求：

- 每一列都要有稳定 `key`；没有显式 `key` 时，至少保证 `dataIndex` 稳定。
- 每个列表页都要传稳定 `tableId`。
- 同一路由下如果存在多张表，必须使用不同 `tableId`。
- `操作` 列按系统列处理，接入时应放进 `systemColumnKeys`。

范围说明：

- 当前只接入标准 `ant-design-vue Table` 路由页。
- 组件内嵌表格、编辑态子表格、`vxe-table` 和代码生成字段表不在这一层通用能力范围内。

## 6. 本地验收

### 本地启动顺序

如果后端源码或 migration 有变动，先重新编译：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
go build -o ./go-admin .
```

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
./go-admin migrate -c config/settings.dev.yml
./go-admin server -c config/settings.dev.yml
```

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin
pnpm install

cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

默认管理员账号：

- 用户名：`admin`
- 密码：`123456`

每次收尾至少跑：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1
./scripts/check-local.sh
```

如果前后端服务已启动，再追加：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```

## 7. 当前推荐推进方式

- 先收口文档、迁移、菜单真相源，再继续做页面承接。
- 默认优先“前端承接已有后端能力”，不扩散改业务后端。
- 当前默认不继续扩写本地后端业务线，优先推进前端原型补全、脚手架增强、页面承接和交互完善。
