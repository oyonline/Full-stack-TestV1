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

