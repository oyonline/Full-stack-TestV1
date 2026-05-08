# C7-2 数据权限：路由策略审计（哪些路由应用 dataScope，哪些豁免）

> 关联 bead: my-t68（C7-2）
> 关联 phase: phase2 数据权限拆分（C7-1 ~ C7-7）
> 立项日期：2026-05-08
> 范围：本文档**只产策略**，不改代码。代码落地由 C7-3 起逐模块实施。

---

## 1. 背景与机制要点

### 1.1 三个独立概念

| 概念 | 位置 | 作用 |
|------|------|------|
| `config.ApplicationConfig.EnableDP` | `settings*.yml` 全局开关 | 关掉则 `actions.Permission` 短路返回原 db；中间件仍执行但注入的范围实际不生效 |
| `actions.PermissionAction()` | gin middleware（路由层） | 查询当前用户的 `data_scope / dept_id / role_id`，写入 `c.Set(PermissionKey, *DataPermission)`。**只注入上下文，不过滤数据。** |
| `actions.Permission(tableName, p)` | GORM scope（service 层） | 真正按 `data_scope` 拼接 `WHERE create_by IN (...)`。**没在 service 调用就等于没接入。** |

> **关键结论**：路由挂 `PermissionAction()` 只是"准备好了上下文"。如果对应 service 没有 `db.Scopes(actions.Permission(...))`，**该路由实际不受 dataScope 影响**。这一点决定了"豁免"和"接入"在代码层的真正含义。

### 1.2 data_scope 取值（来自 `common/actions/permission.go:62`）

| 值 | 含义 | SQL |
|----|------|-----|
| `"1"` | 全部数据 | 不加过滤（default 分支） |
| `"2"` | 自定义（按角色绑定的部门集合） | `create_by IN (sys_role_dept ⋈ sys_user)` |
| `"3"` | 本部门 | `create_by IN (本部门所有 user)` |
| `"4"` | 本部门及以下 | `create_by IN (dept_path LIKE '%/$DeptId/%' 的所有 user)` |
| `"5"` | 仅本人 | `create_by = $UserId` |

---

## 2. 当前挂载情况盘点（grep `PermissionAction()`）

代码现状（2026-05-08，分支 `polecat/fury/my-t68`）：

| 文件 | 路由组 | 中间件 | service 是否 wire `actions.Permission` |
|------|-------|--------|----------------------------------------|
| `app/admin/router/sys_user.go:18` | `/sys-user/**` | ✅ `PermissionAction()` | ✅ 已接入（`service/sys_user.go` 多处） |
| `app/admin/router/sys_user.go:27` | `/user/**` | ✅ `PermissionAction()` | ✅ 已接入 |
| `app/admin/apis/sys_api.go` (route 在 router/sys_api.go 未挂中间件) | `/sys-api/**` | ❌ 未挂 | ✅ 已接入（`service/sys_api.go` 三处） |
| `app/platform/router/attachment.go:20` | `/platform/attachments/**` | ✅ | ❌ service 未 wire |
| `app/platform/router/module_registry.go:20` | `/platform/modules/**` | ✅ | ❌ service 未 wire |
| `app/platform/router/workflow.go:17` | `/platform/workflow/**` | ✅ | ❌ service 未 wire |
| `app/jobs/router/sys_job.go:23-32` | `/sysjob/**` | ✅（逐 handler） | ❌ service 未 wire |

> **盘点观察**：
> 1. `sys_user` 是唯一一个"中间件 + service scope"都齐的模块——但它是**平台底座**，按 C7-2 策略反而**应该解除**接入。
> 2. `sys_api` service 调了 `actions.Permission` 但路由没挂中间件——`getPermissionFromContext` 读不到值会返回零值 `*DataPermission{}`，`Permission()` 走 default 分支不过滤，**目前是空操作**。
> 3. `platform/{attachment,module_registry,workflow}` 中间件挂了但 service 没 wire——**目前不生效**，C7-3 才会真正接入。
> 4. `jobs/sysjob` 用的是老 `actions.IndexAction/ViewAction` 框架内部的 scope，不是手写 service。需要单独评估。

---

## 3. 路由 → 策略判定总表

判定规则：
- **平台底座**：跨租户/跨业务的全局配置/治理资源，由超管/系统管理员维护全量 → **豁免（不加 PermissionAction，service 不 wire scope）**
- **业务数据**：业务用户产生的"我的/我部门/全部"语义可区分的数据 → **接入（路由挂中间件 + service 调用 actions.Permission(table, p)）**
- **公开/认证特殊路由**：登录、刷新、健康检查、ws、上传等 → **N/A（与 dataScope 无关）**

### 3.1 `app/admin/router/`（核心 sys_*，本 bead 主战场）

| 路由 | 文件:行 | 当前状态 | 应用 dataScope？ | 理由 |
|------|---------|----------|------------------|------|
| `GET /sys-user` | sys_user.go:20 | mw✅ + scope✅ | ❌ **豁免** | 用户管理是平台底座，超管要看全量；按部门切片会导致"创建用户的人不在我部门→我看不到该用户" |
| `GET /sys-user/:id` | sys_user.go:21 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `POST /sys-user` | sys_user.go:22 | mw✅ | ❌ **豁免** | 写操作，按角色权限控制（菜单/按钮权限），不按 dataScope |
| `PUT /sys-user` | sys_user.go:23 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `DELETE /sys-user` | sys_user.go:24 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `GET /user/profile` | sys_user.go:29 | mw✅ + scope✅ | ❌ **豁免** | 当前用户读自己 profile，scope 是干扰；按 `user_id from JWT` 过滤即可 |
| `PUT /user/profile` | sys_user.go:30 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `POST /user/avatar` | sys_user.go:31 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `PUT /user/pwd/set` | sys_user.go:32 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `PUT /user/pwd/reset` | sys_user.go:33 | mw✅ + scope✅ | ❌ **豁免** | 重置他人密码——管理员动作，按按钮权限控；scope 不应介入 |
| `PUT /user/status` | sys_user.go:34 | mw✅ + scope✅ | ❌ **豁免** | 同上 |
| `GET /getinfo` | sys_user.go:38 | 仅 auth | ❌ **豁免** | 当前用户的菜单/路由/profile，scope 不适用 |
| `GET /role` | sys_role.go:20 | 仅 auth | ❌ **豁免** | 角色定义是底座 |
| `GET /role/:id` | sys_role.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /role` | sys_role.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /role/:id` | sys_role.go:23 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /role` | sys_role.go:24 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /role-status` | sys_role.go:28 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /roledatascope` | sys_role.go:29 | 仅 auth | ❌ **豁免** | 设置角色 dataScope 自身——元配置不能再受 scope |
| `GET /dept` | sys_dept.go:20 | 仅 auth | ❌ **豁免** | 部门树是底座 |
| `GET /dept/:id` | sys_dept.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /dept` | sys_dept.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /dept/:id` | sys_dept.go:23 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /dept` | sys_dept.go:24 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /deptTree` | sys_dept.go:29 | 仅 auth | ❌ **豁免** | dataScope 计算自身就依赖 deptTree，反向依赖会成环 |
| `GET /menu` | sys_menu.go:20 | 仅 auth | ❌ **豁免** | 菜单/前端路由是底座 |
| `GET /menu/:id` | sys_menu.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /menu` | sys_menu.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /menu/:id` | sys_menu.go:23 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /menu` | sys_menu.go:24 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /menurole` | sys_menu.go:29 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /sys-api` | sys_api.go:20 | 仅 auth（service 残留 scope） | ❌ **豁免** | API 注册表是底座；现状 service 调 `Permission` 但路由没挂中间件，是空操作，C7-3 应清理 |
| `GET /sys-api/:id` | sys_api.go:21 | 仅 auth（service 残留 scope） | ❌ **豁免** | 同上 |
| `PUT /sys-api/:id` | sys_api.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /config` | sys_config.go:20 | 仅 auth | ❌ **豁免** | 系统参数是底座 |
| `GET /config/:id` | sys_config.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /config` | sys_config.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /config/:id` | sys_config.go:23 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /config` | sys_config.go:24 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /configKey/:configKey` | sys_config.go:29 | 仅 auth | ❌ **豁免** | 按 key 查值，全局配置 |
| `GET /app-config` | sys_config.go:34 | 公开 | ❌ **豁免** | 前端启动时拉取 app 名/logo，无身份概念 |
| `PUT /set-config` | sys_config.go:39 | 仅 auth | ❌ **豁免** | 全局参数 |
| `GET /set-config` | sys_config.go:40 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /dict/data` | sys_dict.go:20 | 仅 auth | ❌ **豁免** | 字典是底座 |
| `GET /dict/data/:dictCode` | sys_dict.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /dict/data` | sys_dict.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /dict/data/:dictCode` | sys_dict.go:23 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /dict/data` | sys_dict.go:24 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /dict/type-option-select` | sys_dict.go:26 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /dict/type` | sys_dict.go:27 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /dict/type/:id` | sys_dict.go:28 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /dict/type` | sys_dict.go:29 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /dict/type/:id` | sys_dict.go:30 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /dict/type` | sys_dict.go:31 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /dict-data/option-select` | sys_dict.go:35 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /sys-login-log` | sys_login_log.go:20 | 仅 auth | ❌ **豁免** | 安全审计日志，超管全量；按部门切片会让管理员漏掉跨部门入侵线索 |
| `GET /sys-login-log/:id` | sys_login_log.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /sys-login-log` | sys_login_log.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /sys-opera-log` | sys_opera_log.go:19 | 仅 auth | ❌ **豁免** | 操作日志，同登录日志理由 |
| `GET /sys-opera-log/:id` | sys_opera_log.go:20 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /sys-opera-log` | sys_opera_log.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `GET /post` | sys_post.go:19 | 仅 auth | ❌ **豁免** | 岗位定义是底座 |
| `GET /post/:id` | sys_post.go:20 | 仅 auth | ❌ **豁免** | 同上 |
| `POST /post` | sys_post.go:21 | 仅 auth | ❌ **豁免** | 同上 |
| `PUT /post/:id` | sys_post.go:22 | 仅 auth | ❌ **豁免** | 同上 |
| `DELETE /post` | sys_post.go:23 | 仅 auth | ❌ **豁免** | 同上 |

**`app/admin/router/` 内业务模块（应接入）**

| 路由 | 文件:行 | 当前状态 | 应用 dataScope？ | 理由 |
|------|---------|----------|------------------|------|
| `GET /announcement` | announcement.go:19 | 仅 auth | ✅ **接入** | 公告是业务数据，分发"我发的/我部门发的/全部"语义存在；C7-3 首个落地样板 |
| `GET /announcement/:id` | announcement.go:20 | 仅 auth | ✅ **接入** | 详情查询在 service 层加 `actions.Permission` 防越权 |
| `POST /announcement` | announcement.go:21 | 仅 auth | ❌ **豁免（写）** | 创建按按钮权限控；scope 是读侧概念 |
| `PUT /announcement/:id` | announcement.go:22 | 仅 auth | ✅ **接入（更新前查）** | 更新前的"找到这条记录"要走 scope，防止越权改他人公告 |
| `DELETE /announcement` | announcement.go:23 | 仅 auth | ✅ **接入（删除前查）** | 同上 |
| `POST /announcement/:id/read` | announcement.go:24 | 仅 auth | ✅ **接入** | 标记已读前要确认这条公告确实在当前用户可见范围 |
| `GET /kingdee-customer` | kingdee_customer.go:20 | 仅 auth | ✅ **接入** | 客户档案是业务数据；"我的客户/我部门客户/全部"是产品默认形态 |
| `GET /kingdee-customer/:id` | kingdee_customer.go:21 | 仅 auth | ✅ **接入** | 同上 |
| `POST /kingdee-customer` | kingdee_customer.go:22 | 仅 auth | ❌ **豁免（写）** | 写按按钮权限；scope 不参与创建 |
| `PUT /kingdee-customer/:id` | kingdee_customer.go:23 | 仅 auth | ✅ **接入（更新前查）** | 防越权改他人客户 |
| `DELETE /kingdee-customer` | kingdee_customer.go:24 | 仅 auth | ✅ **接入（删除前查）** | 同上 |
| `GET /kingdee-customer/template` | kingdee_customer.go:25 | 仅 auth | ❌ **豁免** | 静态模板下载，无数据 |
| `POST /kingdee-customer/import` | kingdee_customer.go:26 | 仅 auth | ❌ **豁免（写）** | 导入是写操作 |
| `GET /kingdee-customer/export` | kingdee_customer.go:27 | 仅 auth | ✅ **接入** | 导出 = 批量读，必须走 scope，否则导出泄露 |
| `GET /kingdee-customer/pull` | kingdee_customer.go:29 | 仅 auth | ❌ **豁免** | 从金蝶外部系统拉取，是写流程的入口 |
| `POST /kingdee-customer/group` | kingdee_customer.go:30 | 仅 auth | ❌ **豁免** | 同上，写流程 |
| `GET /kingdee-customer/group` | kingdee_customer.go:31 | 仅 auth | ⚠️ **待评估** | 客户分组是分类元数据；如果分组也按部门隔离则接入，否则豁免——**留给 C7-3 实施时按业务复核** |
| `GET /kingdee-customer/organize` | kingdee_customer.go:32 | 仅 auth | ❌ **豁免** | 拉取金蝶组织信息，元数据 |

### 3.2 `app/admin/router/sys_router.go`（系统骨架路由）

| 路由 | 行 | 应用 dataScope？ | 理由 |
|------|----|------------------|------|
| `GET /` | 42 | ❌ N/A | go-admin 着陆页，非生产 |
| `GET /info` | 44 | ❌ N/A | 健康/版本 |
| `/static/*` | 52 | ❌ N/A | 静态资源 |
| `/form-generator/*` | 54 | ❌ N/A | dev 静态资源 |
| `/swagger/admin/*any` | 59 | ❌ N/A | 文档 |
| `/ws/:id/:channel` | 65 | ❌ N/A | WebSocket，与 dataScope 无关 |
| `/wslogout/:id/:channel` | 66 | ❌ N/A | 同上 |
| `POST /api/v1/login` | 71 | ❌ N/A | 登录 |
| `GET /api/v1/refresh_token` | 73 | ❌ N/A | 刷新 token |
| `GET /roleMenuTreeselect/:roleId` | 83 | ❌ **豁免** | 平台底座（角色↔菜单） |
| `GET /roleDeptTreeselect/:roleId` | 85 | ❌ **豁免** | 平台底座（角色↔部门） |
| `POST /logout` | 86 | ❌ N/A | 登出 |

### 3.3 `app/platform/router/`（平台能力层，业务复用基座）

> 这一层是"业务模块共享的能力组件"。按 platform-layer-audit.md，三者都是已落地的平台能力。

| 路由 | 文件:行 | 当前状态 | 应用 dataScope？ | 理由 |
|------|---------|----------|------------------|------|
| `GET /platform/attachments` | attachment.go:22 | mw✅ + scope❌ | ✅ **接入** | 附件挂载到具体业务记录，应受业务侧 scope；附件本身有 `create_by`，可按 scope 过滤"列我能看到的附件" |
| `POST /platform/attachments/upload` | attachment.go:23 | mw✅ | ❌ **豁免（写）** | 上传是写 |
| `GET /platform/attachments/:id/download` | attachment.go:24 | mw✅ + scope❌ | ✅ **接入** | 下载前要确认在 scope 内，防越权下载他人附件 |
| `DELETE /platform/attachments/:id` | attachment.go:25 | mw✅ + scope❌ | ✅ **接入（删除前查）** | 防越权删除 |
| `GET /platform/modules` | module_registry.go:22 | mw✅ + scope❌ | ❌ **豁免** | 模块注册表是平台底座（描述系统拥有哪些业务模块），所有人看一致 |
| `GET /platform/modules/:id` | module_registry.go:23 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `POST /platform/modules` | module_registry.go:24 | mw✅ + scope❌ | ❌ **豁免** | 同上，超管动作 |
| `PUT /platform/modules` | module_registry.go:25 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `DELETE /platform/modules/:id` | module_registry.go:26 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `GET /platform/workflow/definitions` | workflow.go:19 | mw✅ + scope❌ | ❌ **豁免** | 流程定义是平台元数据，超管/流程管理员维护，不按部门切 |
| `GET /platform/workflow/definitions/:id` | workflow.go:20 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `POST /platform/workflow/definitions` | workflow.go:21 | mw✅ | ❌ **豁免** | 同上 |
| `PUT /platform/workflow/definitions` | workflow.go:22 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `DELETE /platform/workflow/definitions/:id` | workflow.go:23 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `GET /platform/workflow/definitions/:id/nodes` | workflow.go:24 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `PUT /platform/workflow/definitions/:id/nodes` | workflow.go:25 | mw✅ + scope❌ | ❌ **豁免** | 同上 |
| `POST /platform/workflow/instances/start` | workflow.go:27 | mw✅ | ❌ **豁免（写）** | 写 |
| `GET /platform/workflow/instances/started` | workflow.go:28 | mw✅ + scope❌ | ⚠️ **特殊：按 starter 过滤** | "我发起的"——业务语义就是 `starter_user_id = 当前用户`；不走通用 scope，而是 service 显式按 user 过滤 |
| `GET /platform/workflow/instances/:id` | workflow.go:29 | mw✅ + scope❌ | ⚠️ **特殊：业务规则** | 实例可见性 = "我发起的 ∪ 我审批过的 ∪ 我可见的"；通用 scope 不适用，C7-3 实施时单独建模 |
| `POST /platform/workflow/instances/:id/withdraw` | workflow.go:30 | mw✅ | ⚠️ **特殊** | 撤回前判定 starter，不走 scope |
| `GET /platform/workflow/tasks/todo` | workflow.go:32 | mw✅ + scope❌ | ⚠️ **特殊：按 assignee** | "我的待办"——按 `assignee_user_id`，不走 scope |
| `POST /platform/workflow/tasks/:id/approve` | workflow.go:33 | mw✅ | ⚠️ **特殊** | 审批前判定 assignee，不走 scope |
| `POST /platform/workflow/tasks/:id/reject` | workflow.go:34 | mw✅ | ⚠️ **特殊** | 同上 |

> **workflow 模块的特殊性**：审批流的"可见性"由"发起人/审批人/抄送"建模，不是部门数据范围。**C7-3 实施时不要把 workflow 路由加进通用 dataScope**，应保留特殊建模。建议：从路由组中**移除 `actions.PermissionAction()` 中间件**，避免误导后人去 wire scope。

### 3.4 `app/jobs/router/`（调度任务）

| 路由 | 文件:行 | 当前状态 | 应用 dataScope？ | 理由 |
|------|---------|----------|------------------|------|
| `GET /sysjob` | sys_job.go:23 | mw✅（per-handler） | ❌ **豁免** | 调度任务全局，超管维护 |
| `GET /sysjob/:id` | sys_job.go:27 | mw✅ | ❌ **豁免** | 同上 |
| `POST /sysjob` | sys_job.go:30 | 无 | ❌ **豁免** | 同上 |
| `PUT /sysjob` | sys_job.go:31 | mw✅ | ❌ **豁免** | 同上 |
| `DELETE /sysjob` | sys_job.go:32 | mw✅ | ❌ **豁免** | 同上 |
| `GET /job/remove/:id` | sys_job.go:36 | 无 auth | ❌ **豁免** | 内部维护接口 |
| `GET /job/start/:id` | sys_job.go:37 | 无 auth | ❌ **豁免** | 同上 |

> sysjob 现在是用 go-admin 老的 `actions.IndexAction/ViewAction` 通用框架，框架内会读 PermissionKey 自动 wire scope。**C7-3 应明确从路由层去掉 PermissionAction()**，并审查通用 action 是否会"暗中"应用 scope。

### 3.5 `app/other/router/`（杂项）

| 路由 | 文件 | 应用 dataScope？ | 理由 |
|------|------|------------------|------|
| `GET /metrics`, `GET /health` | monitor.go | ❌ N/A | 监控 |
| `GET /captcha` | gen_router.go | ❌ N/A | 验证码（无认证） |
| `GET/POST /gen/*`, `/db/*`, `/sys/tables/*` | gen_router.go | ❌ **豁免** | 代码生成器，开发期工具，超管使用 |
| `POST /public/uploadFile` | file.go | ❌ **豁免** | 通用上传，无业务语义 |
| `POST /feishu/*`, `GET /feishu/subscript` | feishu.go | ❌ N/A | 飞书回调（无认证） |
| `GET /server-monitor` | sys_server_monitor.go | ❌ **豁免** | 服务器监控，超管 |

---

## 4. 路由总数与豁免/接入比例

- **admin/router 业务路由总数**：约 67 条（不含 sys_router 骨架）
- **明确豁免（平台底座）**：49 条
- **明确接入（业务模块）**：announcement 5 条 + kingdee-customer 6 条 = 11 条
- **写操作豁免**：5 条
- **待评估（业务复核）**：1 条（kingdee-customer/group 列表）
- **跨目录额外（platform / jobs / other）**：~30 条；workflow 走特殊建模，attachment 接入下载/列表，其余豁免

---

## 5. 平台底座清单（"豁免"白名单）

下面这些**模块**整体豁免 dataScope（C7-3 不挂中间件、service 不调 `actions.Permission`）：

1. **用户管理** `/sys-user/*`、`/user/*`、`/getinfo`
2. **角色管理** `/role/*`、`/role-status`、`/roledatascope`
3. **部门** `/dept/*`、`/deptTree`
4. **菜单** `/menu/*`、`/menurole`、`/roleMenuTreeselect/*`、`/roleDeptTreeselect/*`
5. **API 注册表** `/sys-api/*`
6. **系统参数** `/config/*`、`/configKey/*`、`/app-config`、`/set-config`
7. **字典** `/dict/*`、`/dict-data/*`
8. **登录日志** `/sys-login-log/*`
9. **操作日志** `/sys-opera-log/*`
10. **岗位** `/post/*`
11. **平台模块注册** `/platform/modules/*`
12. **流程定义** `/platform/workflow/definitions/*`（**整个 workflow 模块不走通用 scope**，由业务侧建模）
13. **调度任务** `/sysjob/*`、`/job/*`
14. **代码生成器** `/gen/*`、`/db/*`、`/sys/tables/*`
15. **服务器监控** `/server-monitor`、`/metrics`、`/health`

---

## 6. 业务模块清单（"接入"路由）

C7-3 起按以下路由实施"中间件 + service scope"双侧 wire：

| 模块 | 路由 | 备注 |
|------|------|------|
| 公告 (announcement) | `GET /announcement`、`GET /announcement/:id`、`PUT /announcement/:id`、`DELETE /announcement`、`POST /announcement/:id/read` | C7-3 首个落地样板 |
| 金蝶客户 (kingdee_customer) | `GET /kingdee-customer`、`GET /kingdee-customer/:id`、`PUT /kingdee-customer/:id`、`DELETE /kingdee-customer`、`GET /kingdee-customer/export` | 接入后 export 必须先 scope |
| 平台附件 (platform/attachments) | `GET /platform/attachments`、`GET /platform/attachments/:id/download`、`DELETE /platform/attachments/:id` | scope 字段：`create_by` |

**workflow 不在此清单**——它使用业务建模的可见性（starter / assignee），不是通用 dataScope。

---

## 7. 风险与变更建议（输出给 C7-3）

### 7.1 应清理的现存"半接入"

C7-3 在落地业务模块前，应顺便修正下面这些"中间件挂了但 service 没 wire"或"service wire 了但路由没挂"的不一致：

| 位置 | 现状 | 建议（C7-3） |
|------|------|--------------|
| `app/admin/router/sys_user.go:18, 27` | 挂了 PermissionAction，service wire 了 scope | **解除**：删 `Use(actions.PermissionAction())`，service 内 `actions.Permission(...)` 也清理（或保留但走 default 分支即可——但不一致会让后人困惑） |
| `app/admin/service/sys_api.go` | service 调了 `Permission`，但路由没挂中间件 | **解除**：service 内 `actions.Permission(...)` 移除（当前是空操作） |
| `app/jobs/router/sys_job.go` | 每个 handler 单独挂 PermissionAction | **解除**：jobs 是平台底座 |
| `app/platform/router/module_registry.go` | 挂了中间件，service 未 wire | **解除中间件**：模块注册是底座 |
| `app/platform/router/workflow.go` | 挂了中间件，service 未 wire | **解除中间件**：workflow 不走通用 scope |
| `app/platform/router/attachment.go` | 挂了中间件，service 未 wire | **保留中间件 + C7-3 在 service wire scope** |

### 7.2 EnableDP=true 落地前必修

**C7-7 真把 `EnableDP` 切到 true 之前**，必须先完成：

1. C7-3：把 announcement/kingdee_customer/attachment 在 service 加 `actions.Permission(table, p)`。
2. 7.1 清单的"解除"项全部完成——否则 EnableDP=true 后，**sys_user/sys_api/sysjob 等平台底座会按 scope 切片**，普通管理员（data_scope='3' 本部门）会突然看不到跨部门用户/API/任务，造成回归。
3. C7-1 已完成的 `data_scope='1'` 兜底 migration 必须在生产 DB 执行。

### 7.3 不在本 bead 范围（仅记录）

- 字段层 dataScope（按字段隐藏列）
- 行级 scope 之外的 ABAC（按业务规则的可见性）
- 多租户隔离（与 dataScope 不是一回事）

---

## 8. 验收对照（来自 bead my-t68）

| 验收项 | 满足情况 |
|--------|----------|
| 路由策略文档完整列出 ~30+ 路由的策略判定 | ✅ 第 3 节列出 100+ 条路由（admin 67 + sys_router 12 + platform 24 + jobs 7 + other ~6） |
| 公告管理等业务模块标记为接入 dataScope | ✅ 第 3.1 节末段 + 第 6 节 |
| 平台底座清单明确不接入 | ✅ 第 5 节 |

---

## 9. 给 C7-3 实施者的清单

参考实施顺序（建议每模块独立 bead 或拆 commit）：

1. announcement service 接入（首个样板，已有 bead my-zh0 = C7-3）
2. kingdee_customer service 接入
3. platform/attachments service 接入
4. 清理 sys_user / sys_api / sysjob / module_registry / workflow 的"半接入"中间件（参考 7.1）
5. 全量回归（参考 C7-5 my-52n）
6. 切 `EnableDP=true`（参考 C7-7 my-8e2）
