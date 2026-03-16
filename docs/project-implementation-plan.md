# 项目实施计划文档

本文档基于 [docs/project-prd.md](docs/project-prd.md) 与当前仓库代码，用于指导按阶段、按模块、按任务持续补全脚手架。状态仅使用：**已完成**、**可直接实施**、**依赖前置确认**、**依赖联调**、**暂不建议启动**、**待补充**。可信度：**已确认**（PRD/代码可证）、**推断**（合理推出）、**待确认**（无法可靠结论）。

---

## 1. 文档说明

| 项 | 说明 |
|----|------|
| **文档目的** | 将 PRD 从需求层转译为可执行实施层；明确先做什么、后做什么、依赖关系、落点文件与完成定义。 |
| **与 project-prd.md 的关系** | PRD 为需求基线（范围、模块、状态、接口、约束）；本实施计划在 PRD 基础上拆阶段、拆模块、拆任务，并绑定文件/接口/验收口径。不重复 PRD 大段背景，侧重“怎么做”。 |
| **使用方式** | 按阶段推进；单模块开发时可抽取“模块级实施计划”+“任务清单”中对应任务做单轮 Plan；联调与验收按第 7、8 节执行。 |
| **实施原则** | 最小闭环优先；复用已有页面/API 模式；最小改动；不推翻已结案逻辑；能写清落点则写清。 |
| **适用范围** | 前端 `vue-vben-admin/apps/web-antd` 与后端 `go-admin` 对接相关的系统管理模块补全与联调验收。 |

---

## 2. 实施总策略

- **总体思路**：先补齐已有页面的缺口（部门编辑/删除），再按“分页表 + 增删改查”母版复用到用户、角色；最后视需要补字典、配置、日志等。认证与公共链路先做校验，避免后续大面积返工。
- **为什么按当前顺序**：部门差编辑/删除即可闭环，且 dept API 已全（updateDeptApi、deleteDeptApi 已在 [api/core/dept.ts](vue-vben-admin/apps/web-antd/src/api/core/dept.ts)）；岗位页可作为“分页表 + 弹窗 CRUD”母版；用户/角色与菜单权限强相关，适合紧随其后。
- **最小闭环优先**：Phase 1 以“部门全 CRUD + 认证路径校验”为首批目标；Phase 2 以“用户管理 + 角色管理”为第二闭环。
- **复用已有模块**：岗位管理 [sys-post/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-post/index.vue) 与 [api/core/post.ts](vue-vben-admin/apps/web-antd/src/api/core/post.ts) 作为分页表 + 搜索 + 新增/编辑/删除弹窗的**母版**；部门树表 + 新增弹窗作为树形维护母版；菜单管理不改为模板，仅作约束参考。
- **最小改动策略**：不改 [router/access.ts](vue-vben-admin/apps/web-antd/src/router/access.ts)、[api/request.ts](vue-vben-admin/apps/web-antd/src/api/request.ts) 的解包与 401 逻辑；不改 getinfo 字段映射；新增页面沿用现有 Table/Modal/Form 与 requestClient 用法。
- **联调前置策略**：认证相关（刷新/登出路径）在 Phase 0 确认或修正；每完成一模块即做该模块联调与冒烟，不集中堆到末期。
- **验证与验收穿插**：每阶段结束做阶段验收；模块交付时做模块级验收（见第 8 节）。

---

## 3. 当前开发完成情况总表（基于仓库全检）

| 模块 | 后端接口 | 前端 API | 前端页面 | 联调状态 | 结论 |
|------|----------|----------|----------|----------|------|
| 登录/认证 | 有 | auth.ts | login.vue | 已对齐 | 已完成 |
| 菜单管理 | 有 | menu + sys-api | sys-menu | 已打通 | 已完成 |
| 部门管理 | 有 | dept.ts | sys-dept | 已打通 | 已完成 |
| 岗位管理 | 有 | post.ts | sys-post | 已打通 | 已完成 |
| 参数配置 | 有 | config.ts | sys-config | 待验证 | 已完成（前端） |
| 用户管理 | 有 | user.ts | sys-user | 待验证 | 已完成（前端） |
| 角色管理 | 有 | role.ts | sys-role | 待验证 | 已完成（前端） |
| 字典类型 | 有 | dict.ts | sys-dict-type | 待验证 | 已完成（前端） |
| 字典数据 | 有 | dict.ts | sys-dict-data | 待验证 | 已完成（前端）；dict 已补充导出 |
| 登录日志 | 有 | login-log.ts | sys-login-log | 待验证 | 已完成（前端） |
| 操作日志 | 有 | opera-log.ts | sys-opera-log | 待验证 | 已完成（前端） |
| 接口管理 | GET/PUT | sys-api.ts | sys-api | 待验证 | 已完成（前端） |

技术栈：后端 Go + Gin + GORM + JWT + Casbin（go-admin）；前端 Vue 3 + Vite + Ant Design Vue（web-antd）；API 前缀 `/api/v1`，开发代理 `/api` → `http://localhost:10086`；request 解包 code/data、401 刷新/跳转登录。

---

## 4. 项目实施分阶段规划

### Phase 0：前置确认与公共链路校验

| 项 | 内容 |
|----|------|
| **阶段目标** | 确认或修正认证相关路径；校验登录 → getinfo → menurole 主链路无回归；明确后续可并行范围。 |
| **包含模块** | 登录/认证/会话（仅校验与路径修正）。 |
| **前置依赖** | 无。 |
| **输出成果** | 刷新 token、登出路径与后端一致的结论或改动的代码/配置；公共链路校验记录。 |
| **完成定义** | 刷新/登出若需改则已改并验证；登录后能稳定进入首页并拉取菜单。 |
| **风险点** | 已确认：代码中 [api/core/auth.ts](vue-vben-admin/apps/web-antd/src/api/core/auth.ts) 已使用 `/v1/refresh_token`、`/v1/logout`，与后端一致；联调行为待验证。 |
| **当前状态** | 已完成（路径已对齐）；联调验证包已从代码核对 login/getinfo/menurole/refresh/logout 与后端一致；建议真实环境执行一次公共链路校验。 |

### Phase 1：已有页面补齐与现有闭环加固

| 项 | 内容 |
|----|------|
| **阶段目标** | 部门管理编辑/删除上线；系统管理最小闭环（菜单+部门+岗位）全部可运维。 |
| **包含模块** | 部门管理（编辑、删除）。 |
| **前置依赖** | Phase 0 可不阻塞（部门不依赖刷新/登出）；建议 Phase 0 与 Phase 1 可并行或先做 Phase 1。 |
| **输出成果** | 部门管理支持编辑、删除；操作列接 updateDeptApi、deleteDeptApi。 |
| **完成定义** | 部门列表可编辑单条（弹窗预填、提交 PUT）、可删除（含二次确认、body ids）；列表刷新正确。 |
| **风险点** | 后端 DELETE `/v1/dept` 需 body `ids`，已确认 [dept.ts](vue-vben-admin/apps/web-antd/src/api/core/dept.ts) deleteDeptApi 已按此实现。 |
| **当前状态** | 已完成。 |

### Phase 2：系统管理核心模块补齐

| 项 | 内容 |
|----|------|
| **阶段目标** | 用户管理、角色管理前后端闭环；形成可复用的“分页表 + CRUD”实施模式。 |
| **包含模块** | 用户管理（页 + API）、角色管理（页 + API）。 |
| **前置依赖** | Phase 1 完成更佳（可选）；后端接口已存在。 |
| **输出成果** | 用户管理页（分页、搜索、新增/编辑/删除、部门/角色/岗位下拉）；角色管理页（分页、搜索、新增/编辑/删除、菜单树/部门树）；对应 api/core 封装。 |
| **完成定义** | 用户/角色列表与后端一致；增删改查与后端联调通过；角色菜单/部门树选择与 roleMenuTreeselect、roleDeptTreeselect 联调通过。 |
| **风险点** | 用户 Update 为 PUT 空路径（无 :id），见 [sys_user.go](go-admin/app/admin/router/sys_user.go)；分页参数与后端 dto 命名需联调确认。 |
| **当前状态** | **已完成（前端）**。用户管理、角色管理前端页面与 API 均已就绪（[sys-user/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-user/index.vue)、[sys-role/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-role/index.vue)、[api/core/role.ts](vue-vben-admin/apps/web-antd/src/api/core/role.ts)）；真实联调待验证。 |

### Phase 3：日志 / 辅助模块补齐

| 项 | 内容 |
|----|------|
| **阶段目标** | 登录日志、操作日志、字典、接口管理独立页可查可维。 |
| **包含模块** | 登录日志、操作日志、字典类型/数据、接口管理（独立页）。参数配置已在前端闭环（T-015 已完成），仅待联调验证。 |
| **前置依赖** | Phase 2 完成；业务优先级低于用户/角色。 |
| **输出成果** | 各模块至少列表 + 搜索/筛选 + 必要操作（如日志删除）；字典类型/数据可维护。 |
| **完成定义** | 列表与后端一致；关键操作联调通过。 |
| **风险点** | 接口管理后端无 POST/DELETE，仅查改；字典数据页存在引用残留（见 4.8）。 |
| **当前状态** | **已完成（前端）**。登录日志、操作日志、字典类型、字典数据、接口管理前端页面与 API 均已就绪（[sys-login-log](vue-vben-admin/apps/web-antd/src/views/admin/sys-login-log/index.vue)、[sys-opera-log](vue-vben-admin/apps/web-antd/src/views/admin/sys-opera-log/index.vue)、[sys-dict-type](vue-vben-admin/apps/web-antd/src/views/admin/sys-dict-type/index.vue)、[sys-dict-data](vue-vben-admin/apps/web-antd/src/views/admin/sys-dict-data/index.vue)、[sys-api](vue-vben-admin/apps/web-antd/src/views/admin/sys-api/index.vue)；[api/core/login-log.ts](vue-vben-admin/apps/web-antd/src/api/core/login-log.ts)、[opera-log.ts](vue-vben-admin/apps/web-antd/src/api/core/opera-log.ts)、[dict.ts](vue-vben-admin/apps/web-antd/src/api/core/dict.ts)、[sys-api.ts](vue-vben-admin/apps/web-antd/src/api/core/sys-api.ts)）；字典数据页引用 getDictTypeOptionSelect/DictTypeOption 与 dict.ts 导出不一致，需修正后联调。 |

### Phase 4：联调、验收、收尾

| 项 | 内容 |
|----|------|
| **阶段目标** | 全链路联调、权限与异常场景验收、文档与已知问题收敛。 |
| **包含模块** | 全项目。 |
| **前置依赖** | Phase 1～3 主体完成。 |
| **输出成果** | 联调报告、验收结论、待办/已知问题列表。 |
| **完成定义** | 见第 8 节验证与验收计划。 |
| **当前状态** | 依赖联调。 |

---

## 5. 模块级实施计划

### 4.1 登录 / 认证 / 会话

| 项 | 内容 |
|----|------|
| 当前状态 | 已完成（主流程）；代码核查：login/getinfo/menurole/refresh/logout 路径已与后端一致；联调行为待验证。 |
| 是否建议当前阶段启动 | 仅做公共链路校验，建议 Phase 0 执行一次。 |
| 前置依赖 | 无。 |
| 推荐实施顺序 | Phase 0 第一项。 |
| 前端任务 | 核对 [api/core/auth.ts](vue-vben-admin/apps/web-antd/src/api/core/auth.ts) 中 refreshTokenApi、logoutApi 的 URL 与后端或代理一致；若不一致则改为 `/v1/refresh_token`、`/v1/logout` 并验证。 |
| 后端任务 | 无（已确认后端路径）。 |
| 联调任务 | 登录 → getinfo → menurole 全流程；刷新 token（若启用）；登出后清除并跳转登录。 |
| 验证任务 | 401 后刷新或跳转登录；登出后无法访问需鉴权接口。 |
| 关键文件 | `vue-vben-admin/apps/web-antd/src/api/core/auth.ts`、`api/request.ts`。 |
| 完成定义 | 路径一致且主链路无回归。 |
| 备注 | 已确认：后端 GET `/v1/refresh_token`、POST `/v1/logout`；前端 auth.ts 已使用上述路径。 |

### 4.2 部门管理

| 项 | 内容 |
|----|------|
| 当前状态 | 已完成。 |
| 是否建议当前阶段启动 | 不新增开发；作为树表+弹窗 CRUD+父级排除的**母版**参考。 |
| 前置依赖 | 无。 |
| 推荐复用对象 | 本页已有新增弹窗与表单结构；编辑弹窗可参考岗位编辑（getDeptDetailApi 已存在）。 |
| 前端任务拆解 | ①～④ 已实现：操作列编辑/删除、编辑弹窗 getDeptDetailApi 回填与 updateDeptApi 提交、删除二次确认与 deleteDeptApi、父级树 filterDeptTreeExcludeNode 排除自身及子节点。 |
| 后端任务 | 无；接口已全。 |
| 联调任务 | PUT `/v1/dept/:id`、DELETE `/v1/dept` body `ids`。 |
| 验证任务 | 编辑后列表更新；删除后节点消失；树形展开收缩正常。 |
| 关键文件 | [views/admin/sys-dept/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-dept/index.vue)、[api/core/dept.ts](vue-vben-admin/apps/web-antd/src/api/core/dept.ts)。 |
| 完成定义 | 部门支持完整增删改查且与后端一致。 |
| 备注 | 已确认：编辑/删除/父级排除均已实现。 |

### 4.3 岗位管理

| 项 | 内容 |
|----|------|
| 当前状态 | 已完成。 |
| 是否建议当前阶段启动 | 不新增开发；作为 Phase 2 用户/角色页的**母版**参考。 |
| 推荐复用对象 | 分页表、搜索、新增/编辑弹窗、delete 二次确认、requestClient 与 api/core/post.ts 用法。 |
| 关键文件 | [views/admin/sys-post/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-post/index.vue)、[api/core/post.ts](vue-vben-admin/apps/web-antd/src/api/core/post.ts)。 |
| 备注 | 已确认：后端分页 pageIndex/pageSize 已与前端对齐。 |

### 4.4 用户管理

| 项 | 内容 |
|----|------|
| 当前状态 | 部分完成（前端已闭环）；与后端接口已核对；待真实联调验证。 |
| 是否建议当前阶段启动 | 不新增开发；作为复杂关联页**母版**参考；建议先执行用户管理联调验证。 |
| 前置依赖 | 岗位页、部门页、参数配置页已作为母版复用。 |
| 推荐复用对象 | 已实现：分页表 + 搜索 + 弹窗 CRUD + 部门/角色/岗位下拉（getDeptTreeApi、getRolePage、getPostPage）；[api/core/user.ts](vue-vben-admin/apps/web-antd/src/api/core/user.ts) 含 getSysUserPage、getSysUserDetail、createSysUser、updateSysUser、deleteSysUser（PUT 无 :id、body 含 userId）。 |
| 前端任务拆解 | ①～⑤ 已实现：api/core/user.ts sys-user CRUD、sys-user/index.vue 列表/搜索/分页、新增/编辑弹窗（部门/角色/岗位下拉）、删除二次确认、列表刷新。 |
| 后端任务 | 无；CRUD 已存在。 |
| 联调任务 | 分页参数、Update 请求体与路径（PUT 无 :id）；getinfo 与列表字段区分。 |
| 验证任务 | 列表/搜索/分页、新增/编辑/删除、状态切换（若做）。 |
| 关键文件 | [api/core/user.ts](vue-vben-admin/apps/web-antd/src/api/core/user.ts)、[views/admin/sys-user/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-user/index.vue)；后端 [router/sys_user.go](go-admin/app/admin/router/sys_user.go)。 |
| 完成定义 | 用户管理页与后端 /v1/sys-user 全流程打通（当前前端已闭环；接口口径已确认）。 |
| 备注 | 已确认：后端 PUT 空路径、body 含 userId；DELETE body ids（SysUserById）；前端已按此实现；真实联调待验证。 |

### 4.5 角色管理

| 项 | 内容 |
|----|------|
| 当前状态 | **已完成（前端闭环）**；待联调验证。 |
| 是否建议当前阶段启动 | 不新增开发；建议优先做联调验证。 |
| 前置依赖 | 无；后端 CRUD + roleMenuTreeselect、roleDeptTreeselect 已存在。 |
| 已落点 | [api/core/role.ts](vue-vben-admin/apps/web-antd/src/api/core/role.ts) 全 CRUD + getRoleMenuTreeselect、getRoleDeptTreeselect；[views/admin/sys-role/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-role/index.vue) 分页表、搜索、新增/编辑/删除、菜单树/部门树勾选。 |
| 联调任务 | 分页、树选择接口返回结构与前端树组件兼容性。 |
| 验证任务 | 角色增删改查；菜单/部门权限分配后生效（如 menurole 变化）。 |
| 关键文件 | [api/core/role.ts](vue-vben-admin/apps/web-antd/src/api/core/role.ts)、[views/admin/sys-role/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-role/index.vue)；后端 [router/sys_role.go](go-admin/app/admin/router/sys_role.go)。 |
| 完成定义 | 角色管理页与后端 /v1/role 及树选择接口联调通过。 |
| 备注 | 前端已闭环；树接口返回结构联调时确认。 |

### 4.6 菜单管理

| 项 | 内容 |
|----|------|
| 当前状态 | 已完成。 |
| 是否建议当前阶段启动 | 不开发；仅遵守既有约束，不做重构。 |
| 既有约束 | component 必须在 access validViewPathSet 内；编辑全量回传；ICON_SHORT_KEY_MAP 与 access 一致。 |
| 备注 | 已确认。 |

### 4.7 参数配置

| 项 | 内容 |
|----|------|
| 当前状态 | 已完成（前端闭环）；与后端接口已核对；待联调验证。 |
| 是否建议当前阶段启动 | 不新增开发；可作为分页表+弹窗 CRUD 的**母版**参考。 |
| 前端文件 | [views/admin/sys-config/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-config/index.vue)、[api/core/config.ts](vue-vben-admin/apps/web-antd/src/api/core/config.ts)（getConfigPage、getConfigDetail、createConfig、updateConfig、deleteConfig）。 |
| 完成定义 | 分页、搜索、新增/编辑/删除与后端一致。 |
| 备注 | 已确认：前端已闭环；后端 GET/POST/PUT/DELETE `/v1/config` 已存在。 |

### 4.8 字典类型 / 字典数据、登录日志、操作日志、接口管理（独立页）

| 项 | 内容 |
|----|------|
| 当前状态 | **已完成（前端）**；待联调验证。 |
| 是否建议当前阶段启动 | 不新增开发；建议统一联调。 |
| 已落点 | 字典类型：dict.ts 类型 CRUD、sys-dict-type。字典数据：dict 数据 CRUD、sys-dict-data；dict 已补充 getDictTypeOptionSelect、DictTypeOption。登录/操作日志：login-log.ts、opera-log.ts 与对应 views。接口管理：sys-api 列表+编辑。 |
| 完成定义 | 列表与后端一致；关键操作联调通过。 |
| 备注 | 接口管理后端无 POST/DELETE，仅查改；已按此实现。 |

---

## 6. 任务清单（可执行视角）

| 任务编号 | 所属阶段 | 所属模块 | 任务名称 | 任务目标 | 落点文件/目录 | 前置依赖 | 输出结果 | 完成定义 | 当前状态 | 优先级 | 备注 |
|----------|----------|----------|----------|----------|----------------|----------|----------|----------|----------|--------|------|
| T-001 | Phase 0 | 认证 | 认证路径与后端对齐 | 刷新/登出请求路径与后端或代理一致 | `api/core/auth.ts` | 无 | 代码修改 + 结论说明 | 调用与后端一致且无回归 | 已完成 | P0 | 已确认 auth.ts 使用 /v1/refresh_token、/v1/logout |
| T-002 | Phase 0 | 认证 | 公共链路校验 | 登录→getinfo→menurole 全流程无回归 | 无代码必改 | 无 | 校验记录 | 登录后能进首页且菜单正确 | 可直接实施 | P0 | 联调验证包已从代码核对路径一致；建议真实环境执行 |
| T-003 | Phase 1 | 部门管理 | 部门编辑能力 | 支持单条编辑并提交 | `views/admin/sys-dept/index.vue`、`api/core/dept.ts`（已存在） | 无 | 编辑弹窗 + getDeptDetailApi 回填 + updateDeptApi 提交 | 编辑保存后列表刷新正确 | 已完成 | P0 | |
| T-004 | Phase 1 | 部门管理 | 部门删除能力 | 支持单条/批量删除 | `views/admin/sys-dept/index.vue` | 无 | 删除按钮 + 二次确认 + deleteDeptApi(ids) | 删除后树节点消失 | 已完成 | P0 | |
| T-005 | Phase 1 | 部门管理 | 部门编辑父级排除自身及子节点 | 编辑时父级不能选自己或后代 | `views/admin/sys-dept/index.vue` | T-003 | 父级树选项过滤逻辑 | 选择父级时无非法选项 | 已完成 | P1 | 已实现 filterDeptTreeExcludeNode |
| T-006 | Phase 2 | 用户管理 | 用户 API 封装 | sys-user 分页/详情/增删改 | `api/core/user.ts` | 无 | getPage、get、create、update、delete 等 | 类型与请求与后端一致 | 已完成 | P0 | 已实现；Update 为 PUT 无 :id |
| T-007 | Phase 2 | 用户管理 | 用户管理页骨架 | 列表页 + 路由占位 | `views/admin/sys-user/index.vue` | 无 | 页面文件 + 表格列 + 搜索区 | 页面可打开且表格可请求 | 已完成 | P0 | 已实现 |
| T-008 | Phase 2 | 用户管理 | 用户列表与搜索 | 分页、username 等搜索 | `views/admin/sys-user/index.vue`、`api/core/user.ts` | T-006 | 列表展示与搜索生效 | 与后端 GET /v1/sys-user 一致 | 已完成 | P0 | 与后端 PUT 无 :id、DELETE body ids、分页已核对；待联调验证 |
| T-009 | Phase 2 | 用户管理 | 用户新增/编辑弹窗 | 表单字段与提交 | `views/admin/sys-user/index.vue` | T-006、T-007 | 弹窗 + 表单 + create/update 调用 | 新增/编辑成功并刷新列表 | 已完成 | P0 | 部门/角色/岗位下拉已接入 |
| T-010 | Phase 2 | 用户管理 | 用户删除 | 删除与二次确认 | `views/admin/sys-user/index.vue` | T-006 | 操作列删除 + delete 调用 | 删除后列表刷新 | 已完成 | P0 | 与后端 PUT 无 :id、DELETE body ids 已核对；待联调验证 |
| T-011 | Phase 2 | 角色管理 | 角色 API 封装 | role 分页/详情/增删改及树接口 | `api/core/role.ts` | 无 | getPage、get、create、update、delete、roleMenuTreeselect、roleDeptTreeselect | 与后端一致 | 已完成 | P1 | 已实现 |
| T-012 | Phase 2 | 角色管理 | 角色管理页 | 分页表 + 搜索 + 增删改 + 菜单/部门树选择 | `views/admin/sys-role/index.vue` | T-011 | 完整角色管理页 | 与后端 /v1/role 打通 | 已完成（前端）；依赖联调 | P1 | 页面已就绪 |
| T-013 | Phase 3 | 登录日志 | 登录日志 API + 列表页 | GET/DELETE sys-login-log | `api/core/login-log.ts`、`views/admin/sys-login-log/index.vue` | 无 | 列表 + 筛选 + 删除 + 详情 | 可查可删 | 已完成（前端）；待联调 | P2 | 已实现 |
| T-014 | Phase 3 | 操作日志 | 操作日志 API + 列表页 | GET/DELETE sys-opera-log | `api/core/opera-log.ts`、`views/admin/sys-opera-log/index.vue` | 无 | 列表 + 筛选 + 删除 + 详情 | 可查可删 | 已完成（前端）；待联调 | P2 | 已实现 |
| T-015 | Phase 3 | 参数配置 | 参数配置 API + 页 | CRUD config | `api/core/config.ts`、`views/admin/sys-config/index.vue` | 无 | 列表 + 增删改 | 与后端一致 | 已完成 | P2 | 前端已闭环；待联调验证 |
| T-016 | Phase 3 | 字典 | 字典类型/数据 API + 页 | dict/type、dict/data | `api/core/dict.ts`、`views/admin/sys-dict-type/index.vue`、`sys-dict-data/index.vue` | 无 | 类型列表 + 数据列表/编辑 | 可维护字典 | 已完成 | P2 | dict 已补充 getDictTypeOptionSelect、DictTypeOption |
| T-017 | Phase 3 | 接口管理 | 接口管理只读/编辑页 | GET/PUT sys-api | `api/core/sys-api.ts`、`views/admin/sys-api/index.vue` | 无 | 列表 + 编辑（无新增/删除） | 可查可改 | 已完成（前端）；待联调 | P2 | 已实现 |
| T-018 | Phase 4 | 全项目 | 联调与验收 | 按第 8、9 节执行 | - | Phase 1～3 主体完成 | 联调报告、验收结论 | 见第 9 节 | 依赖联调 | P0 | |

---

## 7. 推荐开发顺序

- **已完成（前端）**：Phase 0 认证路径（T-001）；Phase 1 部门编辑/删除（T-003～T-005）；Phase 2 用户管理（T-006～T-010）、角色管理（T-011～T-012）；Phase 3 参数配置（T-015）、登录/操作日志（T-013～T-014）、字典类型/数据（T-016）、接口管理（T-017）。上述模块前端页面与 API 均已就绪。
- **推荐先做**：**联调验证**：公共链路（T-002）→ 用户/角色 → 参数配置/字典/日志/接口管理。字典数据页引用已通过 dict.ts 补充 getDictTypeOptionSelect、DictTypeOption 导出修正。
- **作为模板复用**：岗位、参数配置（分页表 + 弹窗 CRUD）；用户、角色（分页表 + 关联下拉/树）；部门（树表 + 父级排除）；日志（列表 + 筛选 + 删除 + 详情）。
- **必须后置**：Phase 4 全链路联调与验收在联调验证通过后进行。
- **已修残**：字典数据页 API 引用已与 dict.ts 一致（dict 已补充 getDictTypeOptionSelect、DictTypeOption）。

---

## 8. 联调计划

| 联调项 | 模块 | 接口/路径风险 | 顺序建议 | 完成定义 |
|--------|------|----------------|----------|----------|
| 认证 | 登录/认证 | refreshTokenApi、logoutApi 路径 | 最先 | 路径一致且 401/登出行为正确 |
| 部门 | 部门管理 | PUT /v1/dept/:id、DELETE body ids | Phase 1 已完成 | 编辑/删除与后端一致 |
| 参数配置 | 参数配置 | GET/POST/PUT/DELETE /v1/config | 前端已闭环 | 与后端联调待验证 |
| 用户 | 用户管理 | PUT 无 :id、分页参数、getinfo 与列表字段 | 前端已闭环，建议优先联调 | 列表与增删改与后端一致 |
| 角色 | 角色管理 | roleMenuTreeselect、roleDeptTreeselect 返回结构 | Phase 2 角色页后 | 角色保存后菜单/数据权限生效 |
| 公共链路 | 全局 | request 解包 code/data、成功码 0/200 | 与认证一起 | 各模块响应格式统一处理 |

联调顺序建议：认证与公共链路 → 部门 → 用户 → 角色 → Phase 3 各模块。

---

## 9. 验证与验收计划

- **阶段级验收**：Phase 0 路径已对齐（建议执行一次公共链路校验）；Phase 1 已结束：部门全 CRUD 已实现；Phase 2 用户管理前端已闭环、待联调验证，角色管理待开发；Phase 3 结束：按本期完成的模块做列表与关键操作验收；Phase 4：全链路 + 权限 + 异常场景。
- **模块级验收**：按 PRD 第 7 节模块级验收标准（菜单、部门、岗位、用户、角色等）。
- **页面级验收**：列表加载/空态/错误态、提交成功/失败提示、必填与校验。
- **权限验收**：当前 getAccessCodesApi 固定 `*:*:*` 时登录即可访问已配置菜单；若后续接真实权限码，则验证无权限时不可见或 403。
- **异常场景验收**：401 跳转或刷新、接口报错提示、菜单 component 非法时前端拦截。
- **回归范围建议**：动 access/request/auth 后回归登录与菜单；动部门/用户/角色页后回归对应模块 + 菜单（因菜单可能依赖角色）。

---

## 10. 风险、阻塞项与待确认项

| 类型 | 内容 |
|------|------|
| **当前高风险点** | 用户 Update 路径与 body 使用错误导致编辑失败（前端已按 PUT 无 :id、body 含 userId 实现；联调时验证）。 |
| **可能阻塞实施的问题** | 若后端分页参数与前端不一致（如 page/pageSize vs pageIndex/pageSize），需统一或适配；角色树选择接口返回结构若与前端组件不兼容需适配。 |
| **必须先确认的问题** | 数据权限 enabledp 是否启用及对前端的含义（若启用则用户/角色页可能受影响）。刷新/登出路径代码已与后端一致。 |
| **可边做边确认的问题** | 字典/配置/日志的字段与筛选条件；接口管理是否仅只读+编辑。 |
| **暂不建议过早启动的模块** | 验证码接入（登录当前固定 code/uuid）；大规模重构 access 或 request。 |
| **容易导致返工的点** | 改 access.ts 的菜单→路由映射；改 request 成功码或解包逻辑；改 getinfo 字段映射而不改 store 与各引用处。 |

---

## 11. 附录：实施约束与既有原则

- **最小改动原则**：菜单管理 component/icon、access 映射、编辑全量回传不改；新增菜单 component 须在 validViewPathSet 内；getAccessCodesApi 不请求后端保持现状除非产品明确要接权限码。
- **已结案模块延续约束**：菜单管理：不重写 mapComponent、normalizeMenuIcon；编辑提交必须 fullDetail + editForm 合并 PUT；部门/岗位沿用现有 API 与页面模式。
- **不建议重构的链路**： [router/access.ts](vue-vben-admin/apps/web-antd/src/router/access.ts)、[api/request.ts](vue-vben-admin/apps/web-antd/src/api/request.ts)、[api/core/user.ts](vue-vben-admin/apps/web-antd/src/api/core/user.ts) 的 getinfo 映射。
- **推荐优先复用的页面/模式**： [views/admin/sys-post/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-post/index.vue) + [api/core/post.ts](vue-vben-admin/apps/web-antd/src/api/core/post.ts)（分页表 + 弹窗 CRUD）； [views/admin/sys-config/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-config/index.vue) + [api/core/config.ts](vue-vben-admin/apps/web-antd/src/api/core/config.ts)（分页表 + 弹窗 CRUD）； [views/admin/sys-user/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-user/index.vue) + [api/core/user.ts](vue-vben-admin/apps/web-antd/src/api/core/user.ts)（分页表 + 弹窗 + 部门/角色/岗位下拉，复杂关联页母版）； [views/admin/sys-dept/index.vue](vue-vben-admin/apps/web-antd/src/views/admin/sys-dept/index.vue)（树表 + 弹窗 + 父级排除）；新增 API 时沿用 requestClient、code/data 解包、类型与后端 dto 对齐。
- **后续继续拆任务时的口径建议**：单模块开发时从第 5 节取对应模块、从第 6 节取任务编号与落点、从第 9 节取验收口径；联调时按第 8 节顺序与完成定义执行。

---

## 第一批建议启动任务

以下任务为当前推荐执行顺序（开发已基本就绪，以联调与修残为主）。

| 任务编号 | 任务名称 | 理由 |
|----------|----------|------|
| **已修正** | 字典数据页 API 引用 | dict.ts 已补充 getDictTypeOptionSelect、DictTypeOption 导出，字典数据页可正常构建与运行。 |
| **T-002** | 公共链路校验 | 验证登录→getinfo→menurole 稳定，为后续所有页提供基础。 |
| **联调** | 用户/角色/参数配置/字典/日志/接口管理 | 前端已闭环，真实环境联调验证列表/增删改/树选择等与后端一致。 |

**已完成（前端）**：T-001～T-017 除 T-018 外均已落地；Phase 0～3 前端页面与 API 就绪。Phase 4 全链路联调依赖上述联调与字典数据引用修正。

**说明**：下一步以“真实环境联调”为主线，无新页面开发必要。

---

*文档版本：基于仓库全检更新；已反映角色/字典/日志/接口管理前端已接入；dict 已补充导出。*
