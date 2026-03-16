# 项目级 PRD 基线文档

本文档基于当前仓库代码、目录结构、前后端实现与接口整理，作为后续开发、改造、联调、验收及拆 Plan 的**唯一需求基线**。状态仅使用：已完成、部分完成、待开发、待联调、待验证、待确认。可信度标注：**已确认**（代码/配置可证）、**推断**（结构合理推出）、**待确认**（仓库与上下文无法支撑）。

---

## 1. 项目背景

| 项 | 说明 |
|----|------|
| **项目定位** | 企业级前后端分离管理后台：前端 Vue 3 + Vite + Ant Design Vue（web-antd），后端 Go + Gin + GORM + JWT + Casbin（go-admin）。 |
| **当前建设目标** | 将 vue-vben-admin 的 web-antd 应用与 go-admin 后端对接，形成菜单、部门、岗位等系统管理模块可用的管理端，并沉淀可执行的需求基线。 |
| **项目范围** | 前端：`vue-vben-admin/apps/web-antd`；后端：`go-admin`；API 前缀 `/api/v1`，开发环境前端代理至 `http://localhost:10086`。 |
| **当前阶段目标** | 明确现状、目标态、缺口点、联调点与后续待办，支撑直接拆开发任务。 |
| **文档用途** | 后续开发与拆任务的唯一需求基线；非纯产品想象型 PRD，也非空泛总结。 |

---

## 2. 用户角色与权限

| 项 | 说明 | 可信度 |
|----|------|--------|
| **角色来源** | 由 go-admin 后端与 Casbin 控制；getinfo 返回 `roles`、`permissions`、`buttons`。 | 已确认 |
| **前端权限码** | `getAccessCodesApi()` 固定返回 `['*:*:*']`，登录后可进入首页；不请求后端权限码接口。 | 已确认 |
| **菜单权限** | 菜单与路由来自 GET `/api/v1/menurole`，按当前用户角色过滤；后端控制可见菜单。 | 已确认 |
| **页面/按钮权限** | 依赖 getinfo 的 `permissions`/`buttons`；前端 v-access、AccessControl 可做按权限码/角色控制。 | 推断 |
| **数据权限** | 配置中有 `enabledp: false`（数据权限功能开关）；部门/可见范围/祖先匹配等边界未在前端体现。 | 待确认 |
| **部门可见范围** | 角色与部门关联（roleDeptTreeselect 等）存在后端，前端未实现角色-部门配置页。 | 推断 |

---

## 3. 当前项目现状总览

### 3.1 后端已实现模块（路径与能力）

| 模块 | Router 注册 | API/Service/Model 位置 | 主要接口能力 |
|------|-------------|------------------------|--------------|
| 登录/认证 | `sys_router.go` 内 v1 | 使用 JWT LoginHandler/RefreshHandler | POST `/v1/login`、GET `/v1/refresh_token` |
| 用户信息 | `sys_user.go` | `apis/sys_user.go`、`service/sys_user.go`、`models/sys_user.go` | GET `/v1/getinfo`、CRUD `/v1/sys-user`、profile/avatar/pwd/status |
| 角色 | `sys_role.go` | apis/service/dto/models 对应 | CRUD `/v1/role`、PUT role-status、roledatascope |
| 菜单 | `sys_menu.go` | apis/service/dto/models 对应 | CRUD `/v1/menu`、GET `/v1/menurole` |
| 部门 | `sys_dept.go` | apis/service/dto/models 对应 | CRUD `/v1/dept`、GET `/v1/deptTree` |
| 岗位 | `sys_post.go` | apis/service/dto/models 对应 | CRUD `/v1/post` |
| 字典类型/数据 | `sys_dict.go` | dict type/data apis/services | `/v1/dict/type`、`/v1/dict/data`、option-select 等 |
| 参数配置 | `sys_config.go` | 同上 | CRUD `/v1/config`、configKey、app-config、set-config |
| 接口管理 | `sys_api.go` | apis/service/dto/models 对应 | GET/GET/:id/PUT `/v1/sys-api`（无 POST/DELETE） |
| 登录日志 | `sys_login_log.go` | 同上 | GET/DELETE `/v1/sys-login-log` |
| 操作日志 | `sys_opera_log.go` | 同上 | GET/DELETE `/v1/sys-opera-log` |
| 定时任务 | jobs/router | jobs/apis/service | 对应 job 相关路由 |
| 代码生成/表结构 | other/router gen_router | other/apis/tools | 表/列/生成接口 |
| 文件 | other/router file | other/apis/file | 文件上传等 |
| 服务监控 | other/router sys_server_monitor | 对应 apis | 监控接口 |
| 基础路由 | `sys_router.go` registerBaseRouter | SysMenu/SysDept apis | GET roleMenuTreeselect、roleDeptTreeselect、POST logout |

路由注册链：`cmd/api/server.go` → `app/admin/router/init_router.go` → `InitSysRouter` + `InitExamplesRouter`；CRUD 通过 `routerCheckRole` 在 `app/admin/router/router.go` 中统一注册。

### 3.2 前端已实现页面

| 页面 | 文件位置 | 能力摘要 |
|------|----------|----------|
| 登录 | `vue-vben-admin/apps/web-antd/src/views/_core/authentication/login.vue` | 表单提交调用 `loginApi`，对接 POST `/v1/login`；含滑块验证码占位（code/uuid 固定传 0）。 |
| 菜单管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-menu/index.vue` | 树表、搜索、新增/编辑/删除、关联接口多选（GET `/v1/sys-api`）、component/icon 与 access 一致校验。 |
| 部门管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-dept/index.vue` | 树表、搜索、新增/编辑/删除、编辑时父级排除自身及子节点。 |
| 参数配置 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-config/index.vue` | 分页表、搜索、新增/编辑/删除；对接 GET/POST/PUT/DELETE `/v1/config`。 |
| 岗位管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-post/index.vue` | 分页表、搜索、新增/编辑/删除、分页。 |
| 用户管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-user/index.vue` | 分页表、搜索、新增/编辑/删除、部门/角色/岗位下拉；对接 GET/POST/PUT/DELETE `/v1/sys-user`、PUT 无 :id body 含 userId。 |
| 角色管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-role/index.vue` | 分页表、搜索、新增/编辑/删除、菜单树/部门树（roleMenuTreeselect、roleDeptTreeselect）；`api/core/role.ts` 全 CRUD + 树接口。 |
| 字典类型 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-dict-type/index.vue` | 分页表、搜索、新增/编辑/删除；`api/core/dict.ts` 字典类型 CRUD。 |
| 字典数据 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-dict-data/index.vue` | 分页表、搜索、新增/编辑/删除、字典类型下拉；`dict.ts` 已补充 `getDictTypeOptionSelect`、`DictTypeOption` 导出，与页面引用一致。 |
| 登录日志 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-login-log/index.vue` | 列表、搜索、分页、删除、详情抽屉；`api/core/login-log.ts` GET/DELETE。 |
| 操作日志 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-opera-log/index.vue` | 列表、搜索、分页、删除、详情抽屉；`api/core/opera-log.ts` GET/DELETE。 |
| 接口管理 | `vue-vben-admin/apps/web-antd/src/views/admin/sys-api/index.vue` | 列表、搜索、分页、编辑（无新增/删除）；`api/core/sys-api.ts` getSysApiPage、getSysApiDetail、updateSysApi。 |
| Dashboard/Profile/Fallback | `views/dashboard/*`、`views/_core/profile/*`、`views/_core/fallback/*` | 框架自带或通用页。 |

### 3.3 前端已存在但待完善

- **字典数据**：已修正。`api/core/dict.ts` 已补充 `getDictTypeOptionSelect`、`DictTypeOption` 导出，与 sys-dict-data 页面引用一致。
- **角色/字典类型/登录日志/操作日志/接口管理**：前端页面与 API 已全部接入，待真实环境联调验证；菜单/路由需后端 menurole 配置对应 component 路径（须落在 access validViewPathSet）。

### 3.4 后端已有接口与前端对接现状

- **用户管理**：前后端均已就绪；前端 `sys-user/index.vue` + `api/core/user.ts` 全 CRUD；接口口径已核对一致，待真实环境联调验证。
- **角色管理**：前后端均已就绪；前端 `sys-role/index.vue` + `api/core/role.ts` 全 CRUD 及 roleMenuTreeselect/roleDeptTreeselect；待联调验证。
- **字典类型/数据**：后端 `/v1/dict/type`、`/v1/dict/data` 等齐全；前端 `sys-dict-type`、`sys-dict-data` 页面与 `api/core/dict.ts` 已接入；字典数据页存在 getDictTypeOptionSelect 引用残留（见 3.3）。
- **参数配置**：前后端均已就绪；前端已闭环，待联调验证。
- **登录日志/操作日志**：后端 GET/DELETE 已存在；前端 `sys-login-log`、`sys-opera-log` 页面与 `api/core/login-log.ts`、`api/core/opera-log.ts` 已接入，待联调验证。
- **接口管理**：后端仅 GET/PUT；前端 `sys-api/index.vue` 列表+编辑已接入，菜单页继续使用 getSysApiList 供关联接口多选。

### 3.5 联调状态总览

| 链路 | 状态 | 说明 |
|------|------|------|
| 登录 → getinfo → menurole | 已完成 | 登录后拿 token、拉用户信息、拉菜单并生成路由。 |
| 菜单管理 列表/详情/增删改 | 已完成 | GET/POST/PUT/DELETE `/v1/menu`、GET `/v1/menu/:id`、GET `/v1/sys-api`。 |
| 部门 列表/新增/编辑/删除 | 已完成 | GET/POST/PUT/DELETE `/v1/dept`、树形、编辑时父级排除自身及子节点。 |
| 岗位 分页/增删改查 | 已完成 | GET/POST/PUT/DELETE `/v1/post`，分页参数 pageIndex/pageSize。 |
| 参数配置 分页/增删改查 | 已完成 | 前端与后端接口口径已核对；真实环境联调待验证。 |
| 用户管理 分页/增删改查/部门角色岗位下拉 | 部分完成 | 前端已闭环、接口口径已核对；真实环境联调待验证。 |
| 角色管理 分页/增删改/菜单树/部门树 | 已接入 | 前端页面与 API 已就绪（role.ts 全 CRUD + 树接口）；待联调验证。 |
| 字典类型 分页/增删改查 | 已接入 | 前端页面与 dict 类型接口已就绪；待联调验证。 |
| 字典数据 分页/增删改查 | 已接入 | dict.ts 已补充 getDictTypeOptionSelect、DictTypeOption 导出；待联调验证。 |
| 登录日志/操作日志 列表/筛选/删除 | 已接入 | 前端页面与 API 已就绪；待联调验证。 |
| 接口管理 列表/编辑 | 已接入 | 仅列表+编辑，无新增/删除；待联调验证。 |

### 3.6 当前已知风险点

- 刷新/登出路径：**已确认** 代码中 `api/core/auth.ts` 已使用 GET `/v1/refresh_token`、POST `/v1/logout`，与后端 `sys_router.go` 注册一致；真实环境行为待验证。
- 部门删除：后端 DELETE `/v1/dept` 需传 body `ids`，前端已实现 deleteDeptApi(ids)。

### 3.8 联调验证包结论（本轮）

- **参数配置**：列表/搜索/分页（pageIndex、pageSize、configName、configKey、isFrontend）、新增/编辑/删除、详情回填、删除 body `{ ids }` 与后端 dto/router 一致；**已从代码确认**；真实环境联调待验证。
- **用户管理**：列表/搜索/分页、PUT 无 `:id`（body 含 userId）、DELETE body `{ ids }`、部门/角色/岗位选项来源（getDeptTreeApi、getRolePage、getPostPage）、列表返回含 dept（后端 Preload("Dept")）与前端接法一致；**已从代码确认**；真实环境联调待验证。
- **公共链路**：login（POST /v1/login）、getinfo（GET /v1/getinfo）、menurole（GET /v1/menurole）、refresh（GET /v1/refresh_token）、logout（POST /v1/logout）前后端路径一致；baseURL /api、代理 /api → localhost:10086 与后端 /api/v1 前缀一致；**已从代码确认**；真实环境行为待验证。

### 3.7 历史结案内容汇总（当前实现基线）

- **菜单与 access 一致**：component 必须落在 `access.ts` 的 `validViewPathSet`（由 `import.meta.glob('../views/**/*.vue')` 推导）；菜单类型 C 必填 component 且须为有效视图路径；Layout/BasicLayout/IFrameView 允许。编辑时全量回传（fullDetail + editForm 合并）避免后端 Generate 零值覆盖。
- **Icon**：go-admin 历史库常用短 key（如 user、peoples、tree-table）通过 ICON_SHORT_KEY_MAP 映射为 Iconify；与 `access.ts` 的 `normalizeMenuIcon` 及 sys-menu 编辑回填一致。
- **权限码**：getAccessCodesApi 固定返回 `['*:*:*']`，不请求后端，保证登录后可进首页。
- **getinfo 字段映射**：在 `api/core/user.ts` 中已将 userName→username、name→realName、userId 转 string 等，与 @vben/types UserInfo 对齐。
- **最小改动原则**：后续改菜单/权限/access 时，优先沿用现有 component 校验、icon 映射、编辑全量提交方案；新增菜单 component 须在 validViewPathSet 内。

---

## 4. 模块级 PRD

### 4.1 登录 / 认证 / 会话

| 项 | 内容 |
|----|------|
| 模块目标 | 用户登录、Token 管理、会话与登出。 |
| 页面/入口 | 登录页：`views/_core/authentication/login.vue`；路由由框架 core 配置。 |
| 前端文件 | `api/core/auth.ts`（loginApi、logoutApi、refreshTokenApi、getAccessCodesApi）、`store/auth.ts`、`api/request.ts`（Bearer、401 处理、refresh）。 |
| 后端接口 | POST `/api/v1/login`、GET `/api/v1/refresh_token`、POST `/api/v1/logout`（registerBaseRouter）；getinfo 见用户信息。 |
| 核心功能 | 用户名密码登录；Token 写入 AccessStore；请求头带 Bearer；401 可刷新或跳转登录；登出清 Store 并跳转登录页。 |
| 关键字段 | 请求：username、password、code、uuid（当前固定 0）；响应：code、token、expire。 |
| 权限控制 | 无；登录接口免鉴权。 |
| 历史结案/约束 | 登录成功条件：HTTP 成功且 body.code===200 且 body.token 存在；getAccessCodesApi 不请求后端。 |
| 当前差距 | 验证码未接真实接口；刷新/登出真实环境行为待验证。 |
| 状态 | 已完成（登录+getinfo+menurole 主流程）；路径已与后端代码对齐，联调行为待验证。 |

### 4.2 用户管理

| 项 | 内容 |
|----|------|
| 模块目标 | 系统用户列表、增删改查、状态/密码等。 |
| 页面/入口 | `views/admin/sys-user/index.vue`；路由由 menurole 决定。 |
| 前端文件 | `api/core/user.ts`（getinfo + getSysUserPage、getSysUserDetail、createSysUser、updateSysUser、deleteSysUser、updateSysUserStatus）；`views/admin/sys-user/index.vue`。 |
| 后端接口 | GET/POST/PUT/DELETE `/api/v1/sys-user`、GET `/v1/getinfo`、profile/avatar/pwd/status 等（见 sys_user.go router）。 |
| 核心功能 | 前端：分页列表、搜索、新增/编辑/删除弹窗、部门/角色/岗位下拉（getDeptTreeApi、getRolePage、getPostPage）；编辑为 PUT 无 :id、body 含 userId。后端：分页列表、详情、新增、修改、删除、头像、改密、状态。 |
| 关键字段 | 见 `go-admin/app/admin/service/dto/sys_user.go`、models sys_user；前端 SysUserItem、CreateSysUserData、UpdateSysUserData 已对齐。 |
| 当前差距 | 无；前后端接口口径已从代码核对一致。 |
| 状态 | 部分完成（前端已闭环；接口口径已确认）；真实环境联调待验证。 |

### 4.3 角色管理

| 项 | 内容 |
|----|------|
| 模块目标 | 角色列表、增删改、菜单/部门数据权限分配。 |
| 页面/入口 | `views/admin/sys-role/index.vue`；路由由 menurole 决定。 |
| 前端文件 | `api/core/role.ts`（getRolePage、getRoleDetail、createRole、updateRole、deleteRole、updateRoleStatus、getRoleMenuTreeselect、getRoleDeptTreeselect）；`views/admin/sys-role/index.vue`。 |
| 后端接口 | CRUD `/api/v1/role`、PUT `/v1/role-status`；GET `/v1/roleMenuTreeselect/:roleId`、`/v1/roleDeptTreeselect/:roleId`。 |
| 核心功能 | 分页表、搜索、新增/编辑/删除、编辑时菜单树/部门树勾选（与后端树接口对接）。 |
| 当前差距 | 无；前端已闭环，待真实环境联调验证。 |
| 状态 | 已完成（前端闭环）；待联调验证。 |

### 4.4 菜单管理

| 项 | 内容 |
|----|------|
| 模块目标 | 系统菜单树配置、增删改、关联接口、与动态路由/左侧菜单一致。 |
| 页面/入口 | 路由由 menurole 决定；入口 `views/admin/sys-menu/index.vue`。 |
| 前端文件 | `views/admin/sys-menu/index.vue`、`api/core/menu.ts`（getAllMenusApi）、`api/core/sys-api.ts`（getSysApiList）、`router/access.ts`（菜单转路由、component/icon 映射）。 |
| 后端接口 | GET/POST/PUT/DELETE `/api/v1/menu`、GET `/api/v1/menu/:id`、GET `/api/v1/menurole`、GET `/api/v1/sys-api`（分页取 list）。 |
| 核心功能 | 树表展示、标题/可见搜索、新增/编辑/删除、父级树选、菜单类型 M/C/F、component 下拉（Layout/BasicLayout/IFrameView + validViewPathSet）、icon Iconify、关联接口多选。 |
| 关键字段 | menuId、menuName、title、icon、path、component、parentId、menuType、sort、visible、apis/sysApi 等（见 dto sys_menu.go、models sys_menu.go）。 |
| 数据来源 | 列表 GET `/v1/menu`；详情 GET `/v1/menu/:id`；提交 POST/PUT；关联接口 GET `/v1/sys-api`。 |
| 交互规则 | 类型 C 必填 component 且有效；类型 F 不填 component；编辑用 fullDetail+表单合并 PUT，避免零值覆盖；icon 回显用 normalizeMenuIcon。 |
| 权限控制 | 依赖后端 AuthCheckRole；前端无额外控制。 |
| 异常/边界 | component 不在 validViewPathSet 时前端校验不通过；父级不能选自身或后代（编辑时 filterTreeExcludeNode）。 |
| 历史结案/约束 | component 与 access.ts validViewPathSet 一致；ICON_SHORT_KEY_MAP 与 access 共用；编辑全量回传。 |
| 当前差距 | 无；主流程已闭环。 |
| 状态 | 已完成。 |

### 4.5 部门管理

| 项 | 内容 |
|----|------|
| 模块目标 | 部门树维护、增删改查。 |
| 页面/入口 | `views/admin/sys-dept/index.vue`。 |
| 前端文件 | `views/admin/sys-dept/index.vue`、`api/core/dept.ts`（getDeptListApi、getDeptTreeApi、createDeptApi、updateDeptApi、deleteDeptApi、getDeptDetailApi 等）。 |
| 后端接口 | GET/POST/PUT/DELETE `/api/v1/dept`、GET `/api/v1/dept/:id`、GET `/api/v1/deptTree`。 |
| 核心功能 | 树表、按部门名称/状态搜索、新增/编辑/删除；编辑时父级部门树排除自身及子节点（filterDeptTreeExcludeNode）；后端 GET dept 返回树形（SetDeptPage 递归 Children）。 |
| 关键字段 | deptId、parentId、deptPath、deptName、sort、leader、phone、email、status（见 models sys_dept、api core dept.ts）。 |
| 当前差距 | 无；主流程已闭环。 |
| 状态 | 已完成。 |

### 4.6 岗位管理

| 项 | 内容 |
|----|------|
| 模块目标 | 岗位分页列表、增删改查。 |
| 页面/入口 | `views/admin/sys-post/index.vue`。 |
| 前端文件 | 同上、`api/core/post.ts`（getPostPage、getPostDetail、createPost、updatePost、deletePost）。 |
| 后端接口 | GET/POST/PUT/DELETE `/api/v1/post`、GET `/api/v1/post/:id`；分页参数 pageIndex、pageSize。 |
| 核心功能 | 分页表、岗位名称/状态搜索、新增/编辑/删除、状态 1 停用 2 启用。 |
| 关键字段 | postId、postName、postCode、sort、status、remark（见 api core post.ts、后端 dto/models）。 |
| 状态 | 已完成。 |

### 4.7 字典类型 / 字典数据

| 项 | 内容 |
|----|------|
| 模块目标 | 字典类型与字典数据维护、选项下拉等。 |
| 页面/入口 | 字典类型 `views/admin/sys-dict-type/index.vue`；字典数据 `views/admin/sys-dict-data/index.vue`。 |
| 前端文件 | `api/core/dict.ts`（类型/数据分页、详情、增删改；getDictTypeAll 供下拉）；两页独立列表+弹窗 CRUD。 |
| 后端接口 | `/api/v1/dict/type`、`/v1/dict/type/:id`、`/v1/dict/type-option-select`；`/v1/dict/data`、`/v1/dict/data/:dictCode`；POST/PUT/DELETE。 |
| 当前差距 | 无；dict.ts 已补充 getDictTypeOptionSelect、DictTypeOption 导出。 |
| 状态 | 字典类型、字典数据：已完成（前端闭环）；待联调验证。 |

### 4.8 参数配置

| 项 | 内容 |
|----|------|
| 模块目标 | 系统参数键值配置。 |
| 页面/入口 | `views/admin/sys-config/index.vue`。 |
| 前端文件 | `views/admin/sys-config/index.vue`、`api/core/config.ts`（getConfigPage、getConfigDetail、createConfig、updateConfig、deleteConfig）。 |
| 后端接口 | CRUD `/api/v1/config`、configKey、app-config、set-config。 |
| 核心功能 | 分页表、参数名称/键名/是否前台搜索、新增/编辑/删除、详情回填。 |
| 当前差距 | 无；前后端接口口径已从代码核对一致（列表/搜索/分页/增删改/删除 body ids）。 |
| 状态 | 已完成（前端闭环；接口口径已确认）；真实环境联调待验证。 |

### 4.9 接口管理（SysApi）

| 项 | 内容 |
|----|------|
| 模块目标 | 系统接口列表维护，供菜单关联。 |
| 页面/入口 | 独立页 `views/admin/sys-api/index.vue`；菜单管理页“关联接口”多选仍用 getSysApiList。 |
| 前端文件 | `api/core/sys-api.ts`（getSysApiPage、getSysApiDetail、updateSysApi、getSysApiList）。 |
| 后端接口 | GET `/api/v1/sys-api`（分页）、GET `/v1/sys-api/:id`、PUT `/v1/sys-api/:id`（无 POST/DELETE）。 |
| 核心功能 | 列表、搜索、分页、编辑弹窗（无新增/删除）。 |
| 当前差距 | 无；与后端能力一致，待联调验证。 |
| 状态 | 已完成（前端闭环，仅查改）；待联调验证。 |

### 4.10 登录日志 / 操作日志

| 项 | 内容 |
|----|------|
| 模块目标 | 登录日志、操作日志查询与清理。 |
| 页面/入口 | 登录日志 `views/admin/sys-login-log/index.vue`；操作日志 `views/admin/sys-opera-log/index.vue`。 |
| 前端文件 | `api/core/login-log.ts`（getLoginLogPage、getLoginLogDetail、deleteLoginLog）；`api/core/opera-log.ts`（getOperaLogPage、getOperaLogDetail、deleteOperaLog）。 |
| 后端接口 | GET/DELETE `/api/v1/sys-login-log`、GET `/v1/sys-login-log/:id`；GET/DELETE `/api/v1/sys-opera-log`、GET `/v1/sys-opera-log/:id`。 |
| 核心功能 | 列表、筛选、分页、删除（批量）、详情抽屉（只读）。 |
| 当前差距 | 无；前端已闭环，待联调验证。 |
| 状态 | 已完成（前端闭环）；待联调验证。 |

### 4.11 验证码 / 系统信息

| 项 | 内容 |
|----|------|
| 说明 | 登录请求当前固定 code/uuid 为 0；后端是否有验证码接口未在本次侦查确认。系统信息：GET `/info`（Ping）等。 |
| 状态 | 待确认。 |

---

## 5. 页面清单

| 页面名称 | 路由 | 文件位置 | 所属模块 | 当前状态 | 依赖接口 | 优先级 | 备注 |
|----------|------|----------|----------|----------|----------|--------|------|
| 登录 | /login（core） | `views/_core/authentication/login.vue` | 认证 | 已完成 | POST /v1/login | P0 | code/uuid 固定 |
| 菜单管理 | 由 menurole 决定 | `views/admin/sys-menu/index.vue` | 系统管理 | 已完成 | GET/POST/PUT/DELETE /v1/menu、GET /v1/sys-api | P0 | |
| 部门管理 | 由 menurole 决定 | `views/admin/sys-dept/index.vue` | 系统管理 | 已完成 | GET/POST/PUT/DELETE /v1/dept、GET /v1/deptTree | P0 | 含编辑/删除、父级排除 |
| 参数配置 | 由 menurole 决定 | `views/admin/sys-config/index.vue` | 系统管理 | 已完成 | GET/POST/PUT/DELETE /v1/config | P2 | 真实联调待验证 |
| 岗位管理 | 由 menurole 决定 | `views/admin/sys-post/index.vue` | 系统管理 | 已完成 | GET/POST/PUT/DELETE /v1/post | P0 | |
| 用户管理 | 由 menurole 决定 | `views/admin/sys-user/index.vue` | 系统管理 | 部分完成 | /v1/sys-user、getinfo、dept/post/role 下拉 | P1 | 前端已闭环；真实联调待验证 |
| 角色管理 | 由 menurole 决定 | `views/admin/sys-role/index.vue` | 系统管理 | 已完成（待联调） | /v1/role、roleMenuTreeselect、roleDeptTreeselect | P1 | 页面+API 已就绪 |
| 字典类型 | 由 menurole 决定 | `views/admin/sys-dict-type/index.vue` | 系统管理 | 已完成（待联调） | /v1/dict/type 等 | P2 | |
| 字典数据 | 由 menurole 决定 | `views/admin/sys-dict-data/index.vue` | 系统管理 | 已完成（待联调） | /v1/dict/data；dict 已补充导出 | P2 | |
| 接口管理 | 由 menurole 决定 | `views/admin/sys-api/index.vue` | 系统管理 | 已完成（待联调） | GET/PUT /v1/sys-api | P2 | 仅列表+编辑 |
| 登录日志 | 由 menurole 决定 | `views/admin/sys-login-log/index.vue` | 系统管理 | 已完成（待联调） | GET/DELETE /v1/sys-login-log | P2 | |
| 操作日志 | 由 menurole 决定 | `views/admin/sys-opera-log/index.vue` | 系统管理 | 已完成（待联调） | GET/DELETE /v1/sys-opera-log | P2 | |
| Dashboard | 由 menurole 决定 | `views/dashboard/*` | 工作台 | 已完成 | 无 | P1 | |
| 个人中心/Fallback | core | `views/_core/profile/*`、`views/_core/fallback/*` | 通用 | 已完成 | getinfo 等 | P1 | |

---

## 6. 数据与接口依赖

### 6.1 已有接口能力与前端对应

| 后端路径 | 方法 | 前端封装位置 | 使用页面/场景 |
|----------|------|--------------|----------------|
| /v1/login | POST | api/core/auth.ts loginApi | 登录页 |
| /v1/refresh_token | GET | api/core/auth.ts refreshTokenApi | request 拦截器 |
| /v1/logout | POST | api/core/auth.ts logoutApi | 登出 |
| /v1/getinfo | GET | api/core/user.ts getUserInfoApi | 登录后、store |
| /v1/menurole | GET | api/core/menu.ts getAllMenusApi | access 动态路由/菜单 |
| /v1/menu | GET/POST/PUT/DELETE | sys-menu 内 requestClient | 菜单管理 |
| /v1/menu/:id | GET | 同上 | 菜单编辑详情 |
| /v1/dept | GET/POST/PUT/DELETE | api/core/dept.ts | 部门管理全 CRUD |
| /v1/dept/:id | GET | api/core/dept.ts getDeptDetailApi | 部门详情/编辑回填 |
| /v1/deptTree | GET | api/core/dept.ts getDeptTreeApi | 部门树选择器 |
| /v1/config | GET/POST/PUT/DELETE | api/core/config.ts | 参数配置全 CRUD |
| /v1/config/:id | GET | api/core/config.ts getConfigDetail | 参数配置编辑回填 |
| /v1/post | GET/POST/PUT/DELETE | api/core/post.ts | 岗位管理 |
| /v1/post/:id | GET | api/core/post.ts getPostDetail | 岗位编辑 |
| /v1/sys-user | GET/POST/PUT/DELETE | api/core/user.ts getSysUserPage、createSysUser、updateSysUser、deleteSysUser | 用户管理 |
| /v1/sys-user/:id | GET | api/core/user.ts getSysUserDetail | 用户编辑回填 |
| /v1/role | GET | api/core/role.ts getRolePage | 用户页角色下拉 |
| /v1/role | GET/POST/PUT/DELETE | api/core/role.ts 全 CRUD | 角色管理页 |
| /v1/roleMenuTreeselect/:id | GET | api/core/role.ts getRoleMenuTreeselect | 角色编辑菜单树 |
| /v1/roleDeptTreeselect/:id | GET | api/core/role.ts getRoleDeptTreeselect | 角色编辑部门树 |
| /v1/dict/type、/v1/dict/data | GET/POST/PUT/DELETE | api/core/dict.ts | 字典类型/数据页 |
| /v1/sys-login-log | GET/DELETE | api/core/login-log.ts | 登录日志页 |
| /v1/sys-opera-log | GET/DELETE | api/core/opera-log.ts | 操作日志页 |
| /v1/sys-api | GET | api/core/sys-api.ts getSysApiList | 菜单页关联接口 |
| /v1/sys-api | GET/PUT | api/core/sys-api.ts getSysApiPage、updateSysApi | 接口管理页 |

### 6.2 已打通链路

- 登录 → token → getinfo → getAccessCodesApi（固定）→ menurole → 动态路由与菜单渲染。
- 菜单管理：列表、详情、新增、编辑、删除、关联接口多选。
- 部门管理：树表、搜索、新增、编辑、删除、编辑时父级排除自身及子节点（GET dept 返回树形）。
- 岗位管理：分页、搜索、新增、编辑、删除。
- 参数配置：分页、搜索、新增、编辑、删除（前端已闭环且与后端接口口径已核对；待联调验证）。
- 用户管理：分页、搜索、新增、编辑、删除、部门/角色/岗位下拉（前端已闭环且与后端接口口径已核对；待联调验证）。

### 6.3 已接入待联调能力（前端）

- 用户 CRUD：已完成；与后端联调待验证。
- 角色 CRUD 及 roleMenuTreeselect、roleDeptTreeselect：api/core/role.ts 与 sys-role 页已就绪；待联调验证。
- 字典 type/data：api/core/dict.ts 与 sys-dict-type、sys-dict-data 页已接入；dict.ts 已补充 getDictTypeOptionSelect、DictTypeOption 导出。
- 登录/操作日志：api/core/login-log.ts、opera-log.ts 与对应页面已就绪；待联调验证。
- 接口管理独立页：api/core/sys-api.ts 与 sys-api 页已就绪（列表+编辑）；待联调验证。

### 6.4 字段映射与联调风险

| 项 | 说明 | 风险 |
|----|------|------|
| getinfo | userName→username、name→realName、userId→string；已在 user.ts 映射。 | 低；已确认。 |
| 菜单 visible | 后端为 int（0/1），前端搜索/表单用 string（'0'/'1'）。 | 已兼容。 |
| 部门 status | 0 禁用 1 启用；前后端一致。 | 无。 |
| 岗位 status | 1 停用 2 启用；前后端一致。 | 无。 |
| 分页参数 | 前端 pageIndex、pageSize；后端 common/dto.Pagination 为 pageIndex、pageSize（form 绑定）。 | 已确认一致。 |
| 刷新/登出路径 | 代码已对齐：auth.ts 使用 /v1/refresh_token、/v1/logout，与后端 sys_router 一致。 | 已确认；真实环境联调行为待验证。 |

---

## 7. 验收标准

### 7.1 模块级

- **菜单管理**：树表展示与后端一致；新增/编辑/删除后列表刷新；编辑不丢字段；component 仅能选有效视图或 Layout/BasicLayout/IFrameView；关联接口展示与提交正确。
- **部门管理**：树表展示与 GET /v1/dept 一致；新增/编辑/删除后数据一致；编辑时父级不能选自身及子节点。
- **岗位管理**：分页、搜索、增删改查与后端一致；状态与备注展示正确。
- **参数配置**：分页、搜索、增删改查与后端一致（待联调验证）。
- **用户管理**：分页、搜索、增删改查、部门/角色/岗位下拉与后端一致（待联调验证）。

### 7.2 页面级

- 列表加载、空态、错误态有反馈；提交成功/失败有提示；必填与前端校验与后端一致。

### 7.3 联调验收

- 登录后能进入首页并看到与角色一致的菜单；菜单管理修改后重新登录或刷新菜单后左侧菜单与配置一致；部门/岗位增删改后列表与后端一致。

### 7.4 权限验收

- 当前 getAccessCodesApi 固定 `*:*:*`，登录即可访问已配置菜单；若后续接真实权限码，需验证无权限时菜单与按钮不可见或 403。

### 7.5 异常场景

- 401 时跳转登录或刷新 token；接口报错有 message 提示；菜单 component 无效时保存被前端拦截并提示。

---

## 8. 实施优先级建议

| 优先级 | 内容 | 说明 |
|--------|------|------|
| P0 | 菜单、部门、岗位、登录 | 已形成最小闭环。 |
| P1 | 用户管理、角色管理 | 前端均已闭环；与菜单/权限强相关，建议优先联调验证。 |
| P2 | 参数配置、字典类型/数据、接口管理、登录/操作日志 | 前端均已接入；dict 已补充导出，字典数据页可正常使用；待联调验证。 |

**推荐下一步**：真实环境联调验证：公共链路 → 用户/角色 → 参数配置/字典/日志/接口管理。

**当前闭环**：登录 + 动态菜单 + 菜单管理 + 部门/岗位/参数配置/用户/角色/字典类型/字典数据（含残留）/登录日志/操作日志/接口管理 前端均已接入；部门、岗位、菜单已联调完成；其余待真实环境联调验证。

**可复用母版**：岗位、参数配置（分页表+弹窗 CRUD）；部门（树表+父级排除）；用户、角色（分页表+关联下拉/树）；日志页（列表+筛选+删除+详情）。

---

## 9. 风险与注意事项

| 类型 | 说明 |
|------|------|
| 最小改动区 | 菜单管理：component 校验、ICON_SHORT_KEY_MAP、access.ts 的 mapComponent/normalizeMenuIcon；编辑全量回传。 |
| 易牵一发动全身 | `router/access.ts`（菜单转路由、component 解析）；`api/request.ts`（解包 code/data、401、refresh）；getinfo 字段映射（user.ts）。 |
| 避免重构式改动 | 不要重写 access 的菜单→路由映射逻辑；不要改 request 成功码 0/200 兼容逻辑；不要改 getinfo 的 UserInfo 字段约定。 |
| 既有约束延续 | 新增菜单 component 必须在 validViewPathSet 内；菜单类型 C 必填 component；icon 沿用短 key 映射表或完整 Iconify。 |
| 待确认项 | 验证码是否接入、后端路径；数据权限 enabledp 是否启用及对前端的含义。刷新/登出路径代码已与后端一致，联调行为待验证。 |

---

*文档版本：基于仓库全检更新；已反映角色/字典/日志/接口管理前端已接入；dict 已补充 getDictTypeOptionSelect/DictTypeOption 导出。*
