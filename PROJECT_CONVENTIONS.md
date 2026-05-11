# PROJECT_CONVENTIONS

本文件不是理想架构说明，而是基于当前仓库代码与仓库内文档做出的“现状固化”草案。

- 扫描范围：仓库代码、SQL、迁移、README、docs
- 本次没有查询运行中数据库，也没有修改业务代码
- 文档中的结论分为三类：
  - `已核实事实`：可以直接从仓库代码或仓库文件证明
  - `待确认项`：仓库内有线索，但仅靠代码扫描无法最终确认
  - `建议规范`：基于当前已存在模式给出的收口建议，不代表重构目标

## 1. 已核实事实

### 1.1 仓库与模块边界

- 工作区是前后端同仓：
  - 后端：`go-admin`
  - 前端：`vue-vben-admin`
- 前端主应用是 `vue-vben-admin/apps/web-antd`，运行在 Vben monorepo 中，当前项目显式把 `accessMode` 设为 `backend`，默认首页是 `/home`。
- 后端业务主模块集中在：
  - `go-admin/app/admin`
  - `go-admin/app/platform`
  - `go-admin/app/jobs`
  - `go-admin/app/other`

### 1.2 前端目录结构与页面组织方式

- `src/views/_core` 放应用运行必需页：
  - 登录、注册、找回密码
  - 个人中心
  - 404 / 403 / 500 / offline 等兜底页
- `src/views/admin` 是当前后台系统页主目录，已存在目录包括：
  - `sys-api`
  - `sys-config`
  - `sys-dept`
  - `sys-dict-data`
  - `sys-dict-type`
  - `sys-job`
  - `sys-login-log`
  - `sys-menu`
  - `sys-opera-log`
  - `sys-post`
  - `sys-role`
  - `sys-server-monitor`
  - `sys-user`
- `src/views/finance/budget/*` 是新接入的业务模块页面，当前每个资源目录都以 `index.vue` 作为主页面：
  - `cost-center`
  - `fee-categories`
  - `versions`
  - `allocation-rules`
- `src/views/platform/workflow/*` 当前已落地：
  - `todo/index.vue`
  - `started/index.vue`
  - 若干 workflow 详情组件
- 当前“路由承载页”的主流命名方式是：
  - 列表页、目录页：`index.vue`
  - 隐藏详情页：`detail.vue`
  - 隐藏工作区页：`workspace.vue`
  - 共享逻辑：同目录 `shared.ts` / `constants.ts`
- `src/views/admin/sys-dict-data` 目录仍在，但当前没有页面文件；仓库保留的是旧地址兼容 redirect，而不是正式页面。

### 1.3 前端页面母版与列表页组织

- 当前后台标准列表页的主流组织方式是：
  - 页面壳：`src/components/admin/page-shell.vue`
  - 分页列表查询：`src/composables/use-admin-table.ts`
  - 树表查询：`src/composables/use-admin-tree-list.ts`
  - 列设置：`src/composables/use-admin-table-columns.ts`
  - 操作按钮权限包装：`src/components/admin/action-button.vue`
- `AdminPageShell` 当前已经承担“紧凑后台页”母版职责：
  - 筛选区
  - 工具栏
  - 列表区
  - 可选紧凑/完整页头
- 绝大多数 `admin` / `finance` / `platform` 列表页都在用这套壳与 composable，但并非全部完全统一：
  - `sys-config/index.vue` 仍保留了更多页面内手写列表状态
  - `sys-menu/index.vue` 直接在页面内调用 `requestClient` 做 CRUD

### 1.4 路由、菜单、component 的真实映射规则

- 当前前端根路由固定包含：
  - `src/router/routes/core.ts` 中的根路由 `/`
  - 认证路由 `/auth/*`
  - 404 兜底
- 代码里还有两类静态路由来源：
  - `src/router/routes/modules/*.ts`
  - `src/router/routes/static/*.ts`
- `modules/*.ts` 当前主要承载 dashboard、demo、vben 外链等代码路由。
- `static/*.ts` 当前主要承载“隐藏但需要直接访问”的页面：
  - `/profile`
  - `/admin/sys-dict-type/detail`
  - `/admin/sys-role/create`
  - `/admin/sys-role/edit`
  - `/admin/sys-dict-data` 旧地址 redirect

- 当前真实业务菜单并不是前端长久硬编码，而是登录后通过 `/api/v1/menurole` 从后端拿 `sys_menu` 树，再在 `src/router/access.ts` 里映射成 Vben 路由。
- 当前映射链路是：
  1. 登录后拿 token
  2. `getinfo` 拿用户信息与权限码
  3. `menurole` 拿菜单树
  4. `generateAccess` 生成可访问菜单与可访问路由
  5. 动态挂到根路由 `/` 下面

- `sys_menu.component` 到前端组件的实际映射规则如下：
  - 空 component 且当前节点有 children：映射为 `BasicLayout`
  - `Layout` / `BasicLayout`：映射为 `BasicLayout`
  - `RouteView`：映射为 `RouteView`
  - `IFrameView`：映射为 `IFrameView`
  - 其他字符串：会做标准化，然后去匹配 `src/views/**/*.vue`
- 当前标准化逻辑允许以下几种写法共存：
  - `/admin/sys-user/index`
  - `admin/sys-user/index`
  - `/views/admin/sys-user/index.vue`
  - `/admin/sys-user`
- 若 component 不存在于真实 view 集合，前端会落到 `/_core/fallback/not-found`。

- 当前路由名生成规则不是强制手写统一名，而是：
  - 优先用 `sys_menu.menu_name`
  - 没有 `menu_name` 时，用 path 转换为 `a-b-c` 形式

- 当前菜单显示状态来自 `sys_menu.visible`：
  - `visible !== '0'` 会被映射为 `hideInMenu: true`
- 当前按钮类型 `F` 会在菜单路由生成时被过滤掉，不进入可访问路由树。

- 仓库里还保留了几处“数据库菜单不够干净时的前端兼容补丁”：
  - 首页不存在时，前端强插 `/home`
  - `/admin/sys-config/set` 存在时，隐藏 `/admin/sys-config`
  - `/admin/dict` 和 `/log` 会被前端强制改成 `RouteView + redirect`
  - 定时任务菜单如果数据库配置坏掉，前端会做一次补丁修复
- 结论：当前“菜单真相源”是数据库 `sys_menu`，但前端 `src/router/access.ts` 仍然承担少量兼容修复层。

### 1.5 API 封装模式

- 当前前端 API 统一入口是：
  - `src/api/request.ts`
  - `src/api/core/*.ts`
  - `src/api/index.ts`
- `requestClient` 默认假设后端返回结构为：
  - `{ code, data, msg, requestId }`
  - 仅在 `code === 200` 时解包 `data`
- 当前 API 层有两种使用方式并存：
  - 标准业务请求：走 `requestClient.get/post/put/delete`
  - 需要保留原始响应的请求：走 `getApiRaw/getHttpRaw/postApiRaw/postHttpRaw`
- 当前明确使用原始响应处理的场景包括：
  - 验证码
  - 登录
  - 刷新 token
  - `menurole`
  - 个别需要直接处理 `res.data` 的接口

- `src/api/core` 目前是“按业务域拆文件”，但不是完全一个资源一个文件：
  - 老系统模块多为单资源文件，如 `role.ts`、`user.ts`、`dept.ts`
  - 新业务模块已开始把同一业务域聚合到一个文件，如 `finance-budget.ts`

- 当前后端 API 风格是双轨并存，不是单一规范：
  - 老系统管理模块偏 REST：
    - `/v1/role`
    - `/v1/post`
    - `/v1/sys-user`
    - `/v1/dict/type`
  - 新财务预算模块偏业务动作式：
    - `/v1/finance/budget/.../list`
    - `/v1/finance/budget/.../get/:id`
    - `/v1/finance/budget/.../add`
    - `/v1/finance/budget/.../edit/:id`
    - `/v1/finance/budget/.../remove`

### 1.6 权限链路：`sys_menu / sys_api / sys_menu_api_rule / sys_role_menu / casbin_rule`

- 当前登录链路会从 `sys_user_role` 读取用户所有角色，并识别主角色 `is_primary`。
- JWT / Gin context 中会同时保存：
  - 主角色：`primaryRoleId` / `primaryRoleKey` / `primaryRoleName`
  - 全角色集合：`roleIds` / `roleKeys` / `roleNames`

- 当前左侧菜单链路是：
  1. `/api/v1/menurole`
  2. 后端 `GetMenuRole`
  3. `SetMenuRole(authctx.GetPrimaryRoleKey(c))`
  4. 按主角色 key 取菜单树
- 这意味着：菜单当前按主角色生效，不是按多角色并集生效。

- 当前按钮权限码来源不是 `sys_api`，而是 `sys_menu.permission`：
  - `/api/v1/getinfo` 会读取当前用户所有角色
  - `service.SysRole.GetByIds` 会把这些角色挂载菜单上的 `permission` 去重汇总
  - 返回给前端的 `permissions` / `buttons` 就是这组并集
- 前端按钮是否显示，当前主要由：
  - `accessStore.accessCodes`
  - `useAdminPermission`
  - `AdminActionButton`
  这条链路控制

- 当前接口权限链路是：
  1. 角色保存时拿到所选 `menuIds`
  2. 预加载这些菜单上的 `SysApi`
  3. 通过 `sys_menu_api_rule` 找到菜单关联的 API
  4. 把 `(roleKey, api.Path, api.Action)` 写入 `sys_casbin_rule`
- 也就是说：
  - `sys_menu` 负责菜单树与按钮 permission 码
  - `sys_api` 负责接口元数据
  - `sys_menu_api_rule` 是菜单到接口的桥
  - `sys_role_menu` 是角色到菜单的桥
  - `sys_casbin_rule` 是真正给 Casbin 做接口校验的数据

- 当前后端 `AuthCheckRole` 的接口校验只使用 JWT 中的 `rolekey`，即主角色 key。
- `admin` 角色当前直接在中间件里短路放行，不走 Casbin 校验。
- 同时，`CasbinExclude` 中还有一批登录、用户信息、菜单树、配置项、部分用户接口等排除路由。

- 代码层已经确认一个现状差异：
  - 前端按钮权限码是多角色并集
  - 后端 Casbin 接口校验当前是主角色单角色
- 这不是文档推测，而是当前代码行为。

- 当前仓库里没有发现 `db.sql` 对 `sys_role_menu` 和 `sys_casbin_rule` 的通用初始化种子数据。
- 这与代码行为是一致的：
  - `admin` 菜单可直接走特殊分支拿全部 `M/C` 菜单
  - `admin` 接口可直接绕过 Casbin
  - 非 admin 角色则依赖角色编辑/角色创建流程去写 `sys_role_menu` 和 `sys_casbin_rule`

### 1.6.1 数据权限（dataScope）规约

> phase2 已接入并通过端到端验收（C7-1 ~ C7-5），本节是业务模块开发者接入数据权限的**真相源**。
> 配套文档：
> - 接入手册（含代码片段与落地清单）：`.ai-memory/procedural/data-permission-wiring.md`
> - 路由策略审计（每条路由"接入/豁免"判定）：`.ai-memory/audits/data-permission-routing.md`
> - 落地样板：`go-admin/app/admin/{router,service}/announcement.go` + 测试 `announcement_data_scope_test.go` / `announcement_permission_test.go`
> 代码核心：`go-admin/common/actions/permission.go`

#### A. EnableDP 全局开关状态

- `config.ApplicationConfig.EnableDP`（`settings*.yml` 配置项）控制整套 dataScope 是否生效。
- **phase2 起默认开启**（C7-7 切换完成）。关掉后 `actions.Permission` 走短路分支返回原 `*gorm.DB`，**不加任何 SQL 过滤**——仅作过渡期保护，**不应作为长期关闭手段**。
- 关掉 EnableDP **不**会卸掉 `PermissionAction()` 中间件，仅让 service 端的 scope 成为空操作。

#### B. 业务模块接入步骤（开发者必读）

落地一个新业务模块时按下列顺序完成 4 步，**任何一步缺失即视为未接入**：

1. **router**：在业务路由组上 `.Use(actions.PermissionAction())`（**整组挂载**，不要 per-handler 挂）：

   ```go
   r := v1.Group("/<resource>").
       Use(authMiddleware.MiddlewareFunc()).
       Use(middleware.AuthCheckRole()).
       Use(actions.PermissionAction())
   ```

2. **apis**：handler 内通过 `actions.GetPermissionFromContext(c)` 取出 `*DataPermission`，传给 service。**严禁把 `*gin.Context` 传到 service 层**（保持 service 不依赖 gin，便于单测）：

   ```go
   p := actions.GetPermissionFromContext(c)
   if err := s.GetPage(&req, p, &list, &count, user.GetUserId(c)); err != nil { ... }
   ```

3. **service**：`GetPage / GetList / Get / Update 写前查 / MarkRead` 等所有"读侧"路径调 `.Scopes(actions.Permission(tableName, p))`：

   ```go
   q := e.Orm.Model(&data).Scopes(
       cDto.MakeCondition(c.GetNeedSearch()),
       actions.Permission(data.TableName(), p),  // ← scope 必加
   )
   ```

   **`Remove` 必须先按 scope 过滤 `Ids` 拿到 allowed 子集再级联删**，否则 dataScope=5 用户传一组 ID 会把别人的也带删（参考 announcement 的 `Remove` 实现）。

4. **tableName 与 model TableName 一致**：`actions.Permission(table, p)` 拼出来的是 `<table>.create_by`。table 必须是该业务主模型 `TableName()` 返回的物理表名（用 `(&MyModel{}).TableName()` 取，不要写字符串字面量），否则在 join / alias 场景下 SQL 会拼错。

> 写操作（`POST /xxx`）按按钮权限控制，**不调 scope**——scope 是读侧概念。

#### C. 5 个 dataScope 语义对照表

下表 SQL 片段与 `common/actions/permission.go` 实现一致；UI tooltip（角色管理页 `vue-vben-admin/apps/web-antd/src/views/admin/sys-role/data/data-scope-options.ts`）必须保持同步：

| 值 | 含义 | SQL 片段 | UI 文案 |
|----|------|---------|--------|
| `"1"` 或空 | 全部数据权限 | 不加过滤 | "全部数据权限" |
| `"2"` | 自定义数据权限 | `create_by IN (sys_role_dept ⋈ sys_user where role_id=?)` | "自定义数据权限"（按当前角色绑定的部门集合） |
| `"3"` | 本部门数据权限 | `create_by IN (SELECT user_id FROM sys_user WHERE dept_id=?)` | "本部门数据权限" |
| `"4"` | 本部门及以下数据权限 | `create_by IN (SELECT user_id FROM sys_user WHERE dept_id IN (SELECT dept_id FROM sys_dept WHERE dept_path LIKE '%/${dept}/%'))` | "本部门及以下数据权限" |
| `"5"` | 仅本人数据权限 | `create_by = ?` | "仅本人数据权限" |

#### D. 平台底座豁免清单（不挂中间件、service 不调 scope）

下面这些模块**整体豁免**——它们是跨业务的全局配置/治理资源，由超管/系统管理员维护全量。**修改时请同步检查**，不要因为"业务模块都接入了"就反过来给底座加 scope：

1. 用户管理 `/sys-user/*`、`/user/*`、`/getinfo`
2. 角色管理 `/role/*`、`/role-status`、`/roledatascope`
3. 部门 `/dept/*`、`/deptTree`
4. 菜单 `/menu/*`、`/menurole`、`/roleMenuTreeselect/*`、`/roleDeptTreeselect/*`
5. API 注册表 `/sys-api/*`
6. 系统参数 `/config/*`、`/configKey/*`、`/app-config`、`/set-config`
7. 字典 `/dict/*`、`/dict-data/*`
8. 登录日志 `/sys-login-log/*`
9. 操作日志 `/sys-opera-log/*`
10. 岗位 `/post/*`
11. 平台模块注册 `/platform/modules/*`
12. 平台附件元数据上传写路径 `POST /platform/attachments/upload`（读路径需接入，见下文"接入清单"）
13. 流程定义与实例 `/platform/workflow/*`（**整个 workflow 不走通用 scope**，由业务侧按 starter / assignee 自行建模）
14. 调度任务 `/sysjob/*`、`/job/*`
15. 代码生成器 `/gen/*`、`/db/*`、`/sys/tables/*`
16. 服务器监控 `/server-monitor`、`/metrics`、`/health`
17. 登录、刷新 token、登出、验证码、健康检查、ws、静态资源、飞书回调等公开/认证特殊路由（与 dataScope 无关）

> 完整的逐路由判定与理由见 `.ai-memory/audits/data-permission-routing.md` §3。C7-3.5 已清理上述底座的"半接入"残留（中间件挂了但 service 没 wire，或反之）。

**当前已接入的业务路由：**

| 模块 | 接入路由 |
|------|---------|
| 公告 announcement | `GET /announcement`、`GET /announcement/:id`、`PUT /announcement/:id`、`DELETE /announcement`、`POST /announcement/:id/read` |
| 平台附件读路径 platform/attachments | `GET /platform/attachments`、`GET /platform/attachments/:id/download`、`DELETE /platform/attachments/:id` |
| 金蝶客户 kingdee_customer | `GET /kingdee-customer`、`GET /kingdee-customer/:id`、`PUT /kingdee-customer/:id`、`DELETE /kingdee-customer`、`GET /kingdee-customer/export`（C4 实施时按本规约接入） |

#### E. 测试要求（每个新业务模块上线前必须有）

参考 `go-admin/app/admin/service/announcement_data_scope_test.go` + `announcement_permission_test.go`：

1. **5 路 dataScope 单元/集成测试**：scope=`"1"`/`"2"`/`"3"`/`"4"`/`"5"` 各一个 case，构造跨部门用户 + 公告 fixture，断言 `GetPage` 返回的行集合与期望一致。
2. **`EnableDP=false` 短路用例**：测试开头/结尾用 `defer` 还原 `config.ApplicationConfig.EnableDP`，避免污染其他测试。
3. **跨用户越权用例**：覆盖 `Get / Update / Remove / MarkRead` 等"基于 ID"的写前查路径，dataScope=5 用户访问他人记录应返回"不存在"。
4. **`Remove` 越权批量用例**：dataScope=5 用户传 `[mine.Id, others.Id]`，断言只删 `mine`，`others` 仍存在。
5. **业务自带可见性 × dataScope 正交叠加用例**（如业务有自带可见性，见 G 节）。
6. 测试栈：in-memory sqlite + `db.AutoMigrate` 装业务模型 + `SysUser / SysDept`，many2many 关联表（如 `sys_role_dept`）手建。

#### F. 业务字段约定

- **默认所有者字段是 `create_by`**：`actions.Permission` 拼的过滤条件就是 `<table>.create_by`。业务主表必须嵌入 `common/models.ControlBy`（提供 `CreateBy/UpdateBy`），否则 scope SQL 会引用不存在的列。
- **自定义所有者字段（如 `owner_id` / `assignee_user_id`）**：当前 `actions.Permission` 不支持参数化所有者字段。如业务的"所有者"语义不是"创建者"（典型例子：CRM 客户负责人 owner_id、工单受理人 assignee_user_id），有两种扩展模式：
  1. **业务侧手动拼条件**：保持 `PermissionAction()` 中间件挂载，service 内不用 `actions.Permission(table, p)` 而是手动按 `p.DataScope` 拼 `WHERE owner_id IN (...)`。当前 workflow 模块（按 starter / assignee）走的就是这条路。
  2. **冗余 `create_by` 与 owner 同步**：每次创建/转交都同步 `create_by = 新 owner`。**不推荐**，与"创建者"语义混淆。
- 跨表/跨 join 场景下，调 `actions.Permission(<物理主表名>, p)`，不要传 alias；如必须用 alias，需在调用前手动重写 SQL 片段。

#### G. 正交语义：dataScope 与业务自带可见性

dataScope 与业务侧自带的"展示规则"是**两套独立机制，并存且 AND 叠加**，不是子集替代：

- **业务自带可见性**：业务记录创建者主动指定"我希望让谁看到"（典型：announcement 的 `announcement_scope` 表 + `OnlyVisible` 参数按部门可见）。
- **dataScope（数据权限）**：当前用户角色被允许读"哪些 `create_by` 创建的记录"——读侧权限规则，与创建者意图无关。

**先后顺序**：service 层先走业务可见性（如 `OnlyVisible` 拼 `JOIN announcement_scope`），再叠加 `Scopes(actions.Permission(table, p))`。一条记录要出现在结果中，必须**同时**满足两边。新模块**不要把 dataScope 等价替换业务自带可见性**，反之亦然。

#### H. 变更 dataScope 的实时性

- **管理员改了角色 `data_scope` 后，该角色用户必须重新登录才生效。**
- 原因：`PermissionAction()` 中间件按 JWT 中的 `roleId` 在每次请求时查 `sys_role.data_scope`，但用户的"主角色 roleId"是登录时写入 JWT 的，登录态期间不会自动刷新。
- **不要**在登录态轮询/推送变更——这是产品默认行为，文档化即可，不需要改代码。
- 角色"绑定部门集合"（dataScope=`"2"` 用）改了同样需要重新登录生效，理由同上。

### 1.6.2 业务模块接入 workflow 平台规约

> phase2 提供的轻量回调机制（C4-pre / my-c9x）。业务模块（C4 SKU、公告等）通过该机制在 workflow 终态时自动回写自身 status。

**核心机制：terminal callback registry**

- 文件：`go-admin/app/platform/service/workflow_callbacks.go`
- 平台暴露两个 API：
  - `service.RegisterTerminalHandler(businessType string, h WorkflowTerminalHandler)`：业务模块在 `init()` 注册。
  - `WorkflowTerminalHandler = func(tx *gorm.DB, binding *models.WorkflowBusinessBinding, terminalStatus string) error`：handler 签名。
- 内部 `dispatchTerminalHandler` 在 `workflow.Approve / Reject / Withdraw` 三个终态分支的 `updateBusinessBinding` 之后、`actionLog` 写入之前调用，**与 workflow 状态变更同事务**。

**业务模块接入步骤：**

1. 在业务 service 包内新增 `init()`，注册 handler：
   ```go
   func init() {
       platformSvc.RegisterTerminalHandler("c4-spu", func(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
           // 用 tx（当前 workflow 事务）更新业务 status
           return tx.Model(&SPU{}).Where("spu_id = ?", binding.BusinessId).
               Update("status", mapTerminalToBusinessStatus(terminalStatus)).Error
       })
   }
   ```
2. handler 内部**只用传入的 `tx`**，不要新开 session；返回 `error` 即触发整个 workflow 事务回滚。
3. `binding.BusinessType` 必须与 `wf_business_binding.business_type` 一致，是注册键也是分发键。

**业务字段命名约定：**

- 业务表用 `workflow_status`（同步 wf 状态）+ `business_status`（业务自身状态机），与 `wf_business_binding` 的双字段语义保持一致。
- 单字段方案（仅 `status`）也可，但 handler 内必须自行映射 wf terminal → 业务状态值。

**终态判定：**

- 终态由 `isTerminalWorkflowStatus` 统一判定：`approved` / `rejected` / `cancelled`。
- 非终态（`in_review` 中间节点流转）**不会**触发 handler。

**旁路约定：**

- 没注册 handler 的 `business_type`：dispatch 静默跳过，不报错。
- 没有 `wf_business_binding` 记录的 workflow（旁路使用）：dispatch 静默跳过。
- 业务模块完全自由选择是否接入；接入后才在终态自动回写。

**测试规范：**

- 业务模块接入回调时，必须补 handler 单测，覆盖 `approved / rejected / cancelled` 三路至少各一例。
- 平台 callback registry 自身测试见 `workflow_callbacks_test.go`，**业务模块不要重复测注册分发本身**，只测自己 handler 的业务语义。

**禁止：**

- handler 不要做长事务、外部 IO（HTTP / 文件 / 锁）：会拖住 workflow 主事务。
- handler 不要静默吞 error；返回 nil 会让 workflow 状态写入但业务状态不一致。
- 不要绕过 registry 直接在 `workflow.go` 里塞业务分支判断；那是 platform / 业务边界回归。

### 1.6.3 业务模块接入 platform.workflow 的最小契约（平台能力使用者样板）

> 「产品中心」（SPU 经由 admin 模块挂载）是当前仓库内**第一个完整接入 platform.workflow 的业务模块**，已通过 fresh-install e2e。后续新业务模块按本节 6 条契约对齐即可，**任一缺失视为未接入，不得通过 review**。
>
> 样板源码（真相源）：
> - 业务表 / 字段：`go-admin/app/admin/models/spu.go`
> - service 与提交流：`go-admin/app/admin/service/spu.go`
> - 终态回调：`go-admin/app/admin/service/spu_workflow_handler.go`
> - 审计 method 契约测试：`go-admin/app/admin/service/spu_e2e_test.go::TestE2E_Spu_AuditMethod_Contract`
> - 角色种子迁移：`go-admin/cmd/migrate/migration/version/1779000000003_product_role_seed.go`
>
> 配套规约：dataScope 细节见 §1.6.1；终态回调机制见 §1.6.2。本节是把两者编排成「样板」的清单，**避免重复说明，只补样板特有的约束**。

#### 1. 业务表必须有的列

业务主表（aggregate root）必须包含下列列，字段名 / 类型与样板对齐——缺一列会让 SubmitForReview / 终态回调 / dataScope 其中一条链路断掉：

| 列名 | GORM 类型 | 用途 |
|------|-----------|------|
| `status` | `int` / `tinyint` | 业务自身状态机（典型：`1 Draft / 2 Reviewing / 3 Approved / 4 Rejected`） |
| `workflow_instance_id` | `bigint, index, default 0` | 当前关联的 platform workflow 实例 ID；终态后**保留**作为审计指针，不清零 |
| `submitted_at` | `*time.Time` NULL | `SubmitForReview` 成功后写入 |
| `approved_at` | `*time.Time` NULL | `Approved` 终态回调写入（`Rejected / Canceled` 不动） |
| `creator_id` | `int, index` | dataScope=5「仅本人」过滤；入库由 service 从 `user.GetUserId(c)` 填 |
| `dept_id` | `int, index` | dataScope=3/4「本部门 / 本部门及以下」过滤；入库由 service 从登录态部门填 |

`ControlBy.create_by` 是字符串账号名，**不能**当作 `creator_id` 使用；必须显式建数值列。模型同时嵌入 `common/models.ControlBy` 与 `common/models.ModelTime`（沿用现有惯例）。

子表（如 SPU → SKU）是否需要 `creator_id / dept_id` 取决于「子表是否独立可写」——独立可写就接 dataScope，否则继承父表（见 §1.6.1 §F 「自定义所有者字段」与 §3.6「最小改动原则」）。

#### 2. service 方法签名带 `p *actions.DataPermission`

读侧 4 类入口签名必须显式带 `p`，且 service 内部走 `Scopes(actions.Permission(table, p))`。**不允许只挂中间件不传 p**——否则 dataScope 在 service 层是空操作：

```go
func (e *MyModel) GetPage(c *dto.MyPageReq, p *actions.DataPermission, list *[]models.MyModel, count *int64) error
func (e *MyModel) Get(c *dto.MyGetReq, p *actions.DataPermission, model *models.MyModel) error
func (e *MyModel) Update(c *dto.MyUpdateReq, p *actions.DataPermission) error // 写前查必须走 scope
func (e *MyModel) Remove(c *dto.MyRemoveReq, p *actions.DataPermission) error // 先按 scope 过滤 Ids 再级联删
```

完整接入步骤（路由整组 `.Use(actions.PermissionAction())` / handler 内 `GetPermissionFromContext` 取 p / `Remove` 越权防御 / `tableName` 与 `model.TableName()` 一致）沿用 §1.6.1 §B，本节不重复。

#### 3. SubmitForReview 调 `platformService.Workflow.Start`，三元组定位 binding

业务侧的「提交审核」方法必须组装 `WorkflowInstanceStartReq` 调 platform（样板 `spu.go:228 SubmitForReview`）：

```go
const (
    MyModuleKey    = "admin"   // 与本模块所属 app 目录一致（admin / platform / jobs / other）
    MyBusinessType = "my-biz"  // 与 RegisterTerminalHandler / wf_business_binding.business_type 一致
)

func (e *MyModel) SubmitForReview(c *gin.Context, p *actions.DataPermission, req *dto.MySubmitReq) (int64, error) {
    // 1. 写前查走 scope，确认 caller 有权操作该记录
    // 2. 校验当前 status ∈ {Draft, Rejected}（允许的「可提交」前置）
    // 3. resolveDefinition(req.DefinitionId) 拿到 platform 上的 definition_id
    wf := &platformService.Workflow{Service: e.Service}
    detail, err := wf.Start(c, &platformDto.WorkflowInstanceStartReq{
        DefinitionId: defID,
        ModuleKey:    MyModuleKey,
        BusinessType: MyBusinessType,
        BusinessId:   int64ToString(model.Id), // wf_business_binding.business_id 是 string
        BusinessNo:   model.Code,
        Title:        model.Name,
        Remark:       req.Remark,
    })
    if err != nil { return 0, err }
    // 4. 同事务更新 status = Reviewing / submitted_at / workflow_instance_id
}
```

`(module_key, business_type, business_id)` 三元组**全仓唯一定位一条 binding**，本节列出的查询都依赖这个约定：

- 列表回灌 workflow_status：`WHERE module_key = ? AND business_type = ? AND business_id IN ?`（样板 `spu.go:94`）
- 「是否已有未结 binding」判定：同一三元组 `First(&binding)`（样板 `spu.go:133`）

新模块不要换其他写法（把 `module_key` 写死成空、把 `business_id` 用 int 列存等），否则上述跨模块 SQL 全部漏命中。

#### 4. `init()` 注册 `RegisterTerminalHandler`，终态在 platform 事务内回写

样板 `spu_workflow_handler.go`：

```go
func init() {
    platformService.RegisterTerminalHandler(MyBusinessType, onMyWorkflowTerminal)
}

func onMyWorkflowTerminal(tx *gorm.DB, binding *platformModels.WorkflowBusinessBinding, terminal string) error {
    // 只用传入的 tx；返回 error 触发整个 workflow 事务回滚
    updates := map[string]interface{}{}
    switch terminal {
    case platformDto.WorkflowStatusApproved:
        updates["status"] = StatusApproved
        now := time.Now(); updates["approved_at"] = &now
    case platformDto.WorkflowStatusRejected:
        updates["status"] = StatusRejected
    case platformDto.WorkflowStatusCanceled:
        // 「撤回 = 回草稿」是 SPU 的业务决策，新模块按自己业务定（也可保留 Canceled 终态）
        updates["status"] = StatusDraft
    default:
        return nil
    }
    return tx.Model(&MyModel{}).Where("id = ?", id).Updates(updates).Error
}
```

handler 禁项与边界完全沿用 §1.6.2「禁止」清单（不开新 session / 不做外部 IO / 不静默吞 error / 不绕过 registry）。

#### 5. 审计 method 命名：`<层>.<域>.<动作>`，改名走三处同步

业务侧落 audit log 的 `Method` 字段必须遵循 `<层>.<域>.<动作>` 三段式。当前仓库已落地的稳定 method 字符串：

```
admin.spu.insert
admin.spu.update
admin.spu.delete
admin.spu.submit
admin.announcement.insert
admin.announcement.update
admin.announcement.delete
admin.announcement.markRead
platform.workflow.task.approve
platform.workflow.task.reject
```

- 层 = app 目录名（`admin` / `platform` / `jobs` / `other`）
- 域 = 业务模型（spu / announcement / workflow.task / ...）
- 动作 = 全小写动词（insert / update / delete / submit / approve / reject / markRead / ...）

**改名约束（硬规则）**：这些字符串是后续日志查询、告警、排查链路的稳定 key，改名必须**同时**改三处，否则视为破坏契约：

1. `app/<层>/apis/<域>.go` 中 `audit.Entry.Method` 字段
2. `app/<层>/service/<域>_e2e_test.go` 的 `TestE2E_*_AuditMethod_Contract` 契约用例（SPU 样板 `spu_e2e_test.go:492`）
3. `docs/<模块>-guide.md` 与本规约引用同一字符串的段落

任一处漏改 e2e 即红。新增 Method 时也要补一行契约用例。

#### 6. 角色种子写 migration，不靠手工补

业务模块上线必须带「角色种子 migration」，落下列三类数据，**全程单事务、幂等**（`FirstOrCreate` / count-check）：

- `sys_role` — 业务管理员 / 业务操作员 两个最小角色，`data_scope` 显式给值（管理员通常 `"1"`，操作员通常 `"5"`）
- `sys_role_menu` — 把业务页 / 按钮菜单挂到上面两个角色；按钮**全集**（管理员）与**操作子集**（操作员）在 migration 内分别按 `permission IN ?` + `menu_type = 'F'` 查出，不要写魔法菜单 ID
- `sys_casbin_rule` — 由 menu → api 桥（`sys_menu_api_rule`）推出来的 `(role_key, api_path, api_action)` 三元组

样板：`1779000000003_product_role_seed.go`（落 `product_admin` + `product_operator`，覆盖 SPU / SKU / Category / Brand 四组按钮）。

**禁止**：

- 把角色种子写到 `config/db.sql`——那是初始化基线，演进性补丁不入 `db.sql`（见 §1.7 / §3.5）
- 把角色种子写到 `config/menu-batch*.sql` 或 `docs/sql/*.sql`——这两类是「应急 / 人工修库」目录，**不在自动迁移链上**，装机即丢
- 依赖运营在角色管理页手工勾菜单——fresh-install e2e 会缺数据导致 Casbin 拒绝

#### 验收：fresh-install e2e

满足上述 6 条契约的业务模块，在干净库上跑：

```bash
go-admin migrate -c config/settings.yml
go test ./go-admin/app/<层>/service/... -run TestE2E_<域>_ -count=1
```

应当**全绿**，**不需要任何手工 SQL / 手工角色赋权**步骤。这是本节的最终验收口径，也是后续业务模块照抄本样板的判定标准。

### 1.7 migration、seed、手工数据修复的边界

- 当前正式迁移入口是 `go-admin migrate -c ...`。
- 迁移执行顺序是：
  1. `AutoMigrate(sys_migration)`
  2. 加载 `cmd/migrate/migration/version/*.go`
  3. 按文件名前 13 位版本号排序执行
  4. 写入 `sys_migration`

- 当前 `1599190683659_tables.go` 是初始化基线：
  - 先 AutoMigrate 一批基础模型
  - 再执行 `config/db.sql`
  - MySQL 下还会先后执行：
    - `config/db-begin-mysql.sql`
    - `config/db.sql`
    - `config/db-end-mysql.sql`

- `version-local` 目录当前存在，但仓库内没有实际本地迁移脚本，只有占位 `doc.go`。

- 当前仓库里与“数据修正”有关的内容分成三类：
  - 正式迁移：`go-admin/cmd/migrate/migration/version/*.go`
  - 初始化 seed：`go-admin/config/db.sql`
  - 手工/一次性 SQL：
    - `go-admin/config/menu-batch*.sql`
    - `go-admin/config/user-role-multi-role.sql`
    - `docs/sql/*.sql`

- 当前已经有部分历史手工修库内容被正式纳入了迁移链：
  - `1774502400000_sys_user_role.go`
  - `1775100000000_sys_user_reset_pwd_permission.go`
  - `1775200000000_dict_data_menu_route.go`
  - `1775300000000_remove_dict_data_page.go`
  - `1775600000000_finance_budget_menus.go`

- 因此当前真实边界不是“只靠 db.sql”，而是：
  - 初始库靠 `db.sql`
  - 演进性结构和长期保留数据修正靠 `version/*.go`
  - 仍有部分人工修库脚本存在，但不属于自动迁移链

### 1.8 命名规范、最小改动原则、禁止事项

- 当前前端命名现状：
  - 目录名多用 kebab-case，如 `sys-login-log`、`fee-categories`
  - 主页面默认 `index.vue`
  - 辅助页常见 `detail.vue`、`workspace.vue`
  - 多词 API 文件常见 kebab-case，如 `sys-api.ts`、`finance-budget.ts`
- 当前后端命名现状：
  - Go 文件多用 snake_case
  - 目录结构稳定分层：
    - `apis`
    - `models`
    - `router`
    - `service`
    - `service/dto`
- 当前权限码命名并不完全统一，仓库内同时存在：
  - `admin:sysRole:update`
  - `finance-budget:costCenter:add`
  - `job:sysJob:list`
  - `system:sysdicttype:remove`
- 结论：仓库现状是“冒号分段”这一点稳定，但前缀、大小写、模块名风格并不完全统一。

- 当前最小改动原则已经体现在代码里：
  - 前端通过 `normalizeViewPath`、兼容 redirect、隐藏静态路由等方式吸收历史路径差异
  - 没有一次性改写所有旧菜单、旧路径、旧权限码
  - 新增修复优先通过 migration / 小范围 SQL / 单页收口完成

- 当前仓库里不应该做的事情，从现状能推导出以下禁止项：
  - 不要把前端路由真相从数据库重新搬回前端硬编码
  - 不要在没有迁移或 SQL 配套的前提下批量改 `sys_menu.path` / `component` / `permission`
  - 不要为了兼容脏菜单数据，直接在业务页里写更多临时路径判断；当前兼容层集中在 `src/router/access.ts`
  - 不要批量统一历史权限码命名；现有权限码已进入菜单、按钮和 Casbin 链路
  - 不要把需要长期保留的 schema/data 修复只写在 `docs/sql` 或 `menu-batch*.sql` 里
  - 不要把新页面直接绕过 `src/api/request.ts`，除非这个接口确实需要原始响应

## 2. 待确认项

- 仅凭仓库扫描，无法确认当前各环境数据库是否都已经执行了最新 migration 与历史手工 SQL。
- 仓库文档声称本地库存在若干当前计数，但本次没有直接查询数据库，因此这些数字不在“已核实事实”范围内。
- `sys_casbin_rule` 的建表是否完全由 `go-admin-core` 的 Casbin adapter 兜底完成，当前仓库代码只看得到使用点，看不到依赖内部实现。
- 非 admin 角色在全量初始化后的首个可用状态，是否总是依赖手工赋权或角色编辑流程，需要结合实际初始化库验证。

## 3. 建议规范

### 3.1 前端目录与页面

- 新增后台标准列表页时，优先延续现有主流：
  - `src/views/<domain>/<resource>/index.vue`
  - `AdminPageShell`
  - `useAdminTable` / `useAdminTreeList`
  - `useAdminTableColumns`
- 新增隐藏详情/工作区页时，优先延续：
  - `detail.vue`
  - `workspace.vue`
  - 对应 `src/router/routes/static/*.ts`

### 3.2 路由、菜单、component

- 继续把数据库 `sys_menu` 视为业务菜单真相源。
- 新增菜单时，`component` 只使用当前代码已经接受的形式：
  - `RouteView`
  - `IFrameView`
  - 指向真实 `src/views` 的页面路径
- 新增父级分组菜单时，优先使用当前已经大量出现的 `RouteView`，不要再引入新的布局表达方式。
- 对历史兼容路径，只在 `src/router/access.ts` 或 `src/router/routes/static/*.ts` 集中处理，不把兼容逻辑散落到业务页面。

### 3.3 API 封装

- 新 API 默认继续落在 `src/api/core/*`，不要在页面里大面积直接拼接口。
- 只有登录、验证码、下载、需要保留完整响应头/响应体等场景，才使用 raw client。
- 不要求把老模块 REST 风格统一改造成新模块 `/list|get|add|edit|remove`，也不反过来；新增接口优先跟随所在模块现有风格。

### 3.4 权限链路

- 新增按钮权限时，先决定它属于：
  - 仅前端按钮显隐：挂到 `sys_menu.permission`
  - 需要落到接口校验：同时补 `sys_api` 与 `sys_menu_api_rule`
- 若变更多角色语义，必须把以下三层一起审视：
  - `/menurole` 的主角色菜单逻辑
  - `/getinfo` 的并集按钮逻辑
  - `AuthCheckRole` 的主角色 Casbin 逻辑
- 在没有统一方案前，不要只改其中一层。

### 3.5 migration / seed / 手工修库

- 长期保留的结构变更、初始化后仍需自动补齐的数据修复，优先写入 `version/*.go`。
- `db.sql` 继续只承担“初始化基线”职责，不再持续堆叠每次演进补丁。
- `menu-batch*.sql`、`docs/sql/*.sql` 只作为应急或人工修复脚本保存；如果脚本结论要长期生效，应回收进正式 migration。
- 新增 migration 后，默认视为需要重新编译后端二进制再执行迁移。

### 3.6 最小改动原则

- 先收口真实约定，再做局部治理，不做一轮式大重命名、大搬家、大统一。
- 对现有 permission、path、menu_id、api_id 的改动，默认视为高风险操作。
- 治理优先级应先选“跨模块收益高、改动面可控”的地方，而不是先碰全仓公共底座。

## 4. 最适合先治理的 3 个模块

以下排序按“收益 / 风险比”综合判断，不是按技术难度排序。

### 1) 初始化与迁移边界

- 范围：
  - `go-admin/cmd/migrate/migration/*`
  - `go-admin/config/db.sql`
  - `go-admin/config/menu-batch*.sql`
  - `go-admin/config/user-role-multi-role.sql`
  - `docs/sql/*`
- 原因：
  - 当前正式 migration、基线 seed、手工修库脚本三套机制并存
  - 这是最容易继续漂移、但也最适合先通过文档和流程治理收口的区域
  - 先治理这块，能减少后续“菜单没了 / 字段缺了 / 数据不齐”的低级回归

### 2) 菜单 / 路由 / component 真相源

- 范围：
  - `vue-vben-admin/apps/web-antd/src/router/*`
  - `vue-vben-admin/apps/web-antd/src/views/admin/sys-menu/*`
  - `go-admin/app/admin/service/sys_menu.go`
  - 相关菜单 migration / SQL
- 原因：
  - 当前新增页面成败，高度依赖这条链路是否被正确理解
  - 仓库已经形成“数据库为主、前端兼容层兜底”的稳定模式
  - 先治理这块，可以直接提升后续新增页面和修菜单的成功率

### 3) 多角色权限链路

- 范围：
  - `go-admin/common/middleware/handler/*`
  - `go-admin/common/authctx/*`
  - `go-admin/common/middleware/permission.go`
  - `go-admin/app/admin/service/sys_role.go`
  - `go-admin/app/admin/service/sys_user.go`
  - 前端 `useAdminPermission` / `AdminActionButton`
- 原因：
  - 当前代码已经形成“菜单按主角色、按钮按并集、接口按主角色”的三段式语义
  - 收益非常高，但牵涉菜单、按钮、Casbin、登录态、角色编辑流程，风险也最高
  - 适合在前两项先收口之后，再做专项治理

