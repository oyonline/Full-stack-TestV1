# 已知问题与限制

## 表单构建依赖外部资源

- 表单构建当前承接的是后端静态页（`go-admin/static/form-generator/`）。
- 资源已 vendor 化到 `go-admin/static/form-generator/vendor/`：vue、vue-router、element-ui（含 `theme-chalk/fonts/`）、monaco-editor（`min/vs/` 完整目录，AMD basePath）、tinymce 5.3.2 完整包、jquery、js-beautify。
- `index.html`、`preview.html`、`js/index.*.js`、`js/preview.*.js` 中所有原 `https://lib.baomitu.com/...` 已改写为 `/form-generator/vendor/...`，与 gin `r.Static("/form-generator", "./static/form-generator")` 路由匹配。
- 验收：`grep -rn 'baomitu' go-admin/static/` 必须为 0；表单构建页在屏蔽 baomitu 域名 / 断网情况下仍能渲染。
- 重新升级 vue/element-ui/monaco/tinymce 版本时，需要同步刷新 vendor/ 内的对应资源，并确认 HTML 与 JS bundle 的引用路径仍然正确。

## 登录失败文案与排查

- JWT 中间件可能将「用户不存在 / 密码错误 / `status≠2`」等多种失败统一返回 **`incorrect Username or Password`**，易误判为仅密码问题。
- **验证码**错误通常另有提示（如「验证码错误」）；验证码与 `uuid` **一次性**，失败后须重新获取。
- **账号真相源**在数据库 `sys_user`，勿假设 README、文档示例密码与本地库一致。
- 本地联调排查简表见 [`docs/local-dev-handoff.md`](../docs/local-dev-handoff.md)。

## 数据库连接串日志脱敏

- 后端 `database.Setup` 打印 DSN 时已对密码段脱敏（`user:***@tcp(...)`）。
- 自定义日志中请勿打印完整 `database.source`。

## 数据库日志开关默认不在主配置启用

- 数据库操作日志和登录日志是否落库，依赖 `logger.enableddb`。
- 当前建议在 `settings.local.yml` 中显式配置。
- 如果改了本地日志开关但后端未重启，页面会继续看起来“没有日志”。

## 代码生成危险动作仍关闭

- 当前代码生成支持：
  - 导入
  - 配置维护
  - 保存
  - 预览
  - 移除
- 当前不支持：
  - 生成到项目
  - 生成菜单/API

## 菜单组件路径必须真实可映射

- 数据库中一旦写了不存在的 `component`，前端就会落到 not-found。
- 父级分组菜单如果误写成会渲染布局的组件，可能导致二次侧栏或内容区错位。

## 字典管理的旧菜单方案已废弃

- `字典数据` 作为独立页面的历史方案已经废弃。
- 如果有人手工执行旧的 `menu-batch3-web-antd.sql` / `menu-batch4-dict-log-fix.sql` 历史片段，或按旧认知恢复 `menu_id=59`、`menu_id=240`，左侧导航和权限模型会重新回到错误状态。
- 当前正式状态应为：
  - 左侧导航只保留 `字典类型`
  - `241/242/243` 挂在 `543` 下
  - `1775300000000_remove_dict_data_page.go` 已执行

## 空库或半初始化库会直接导致导航缺失

- 本地库如果 `sys_menu` 为空，`/api/v1/menurole` 就没有菜单可返回。
- 这种情况下首页可能还能进，但左侧导航会完全缺失。
- 先查 `sys_menu`、`sys_user`、`sys_role`、`sys_user_role`、`sys_migration`，不要先在前端写临时菜单。

## 修改迁移源码后，旧二进制不会自动带上新迁移

- 当前项目执行 `./go-admin migrate ...` 时，实际跑的是本地 `./go-admin` 二进制里编进去的迁移版本。
- 如果只改了 `go-admin/cmd/migrate/migration/version/*.go` 源码，但没有重新 `go build -o ./go-admin .`，新迁移不会执行。

## `sys_user` 字段容易与业务模型脱节

- 当前业务模型已经依赖 `open_id`、`job_title`、`open_department_id`、`open_department_ids`、`cn_name`。
- 如果本地库缺这些列，新增用户会报 `Error 1054 (42S22): Unknown column 'open_id' in 'field list'`。
- 这类问题应通过正式 migration 修复，不能靠手工改一台机器的表结构长期维持。

## 头像配置依赖新字段

- 当前头像方案依赖 `sys_user.avatar_type` 和 `sys_user.avatar_color` 两列。
- migration `1775000000000_user_avatar_profile.go` 已进入已提交历史（fs-2ie），含老数据回填
  `avatar_type='image'`；Go model（`adminModels.SysUser` 与 migration 私有副本）同步含两字段。
- 后端 API / DTO 数据回路已完整收口（profile + admin CRUD + 上传保色，my-3fo），
  `getinfo` / `user/profile` / `user/avatar` / admin `SysUserInsert` / `SysUserUpdate`
  全链路读写 `avatar_type` 与 `avatar_color`，且图片上传分支保留原 `avatar_color`。

## 数据权限（dataScope）已启用，未接入的业务模块视为缺陷

- phase2 起 `config.ApplicationConfig.EnableDP=true` 已默认开启（C7-1 ~ C7-7 收口）。
- **新增业务模块若未按 `PROJECT_CONVENTIONS.md` §1.6.1 接入 dataScope（路由挂 `actions.PermissionAction()` + service 调 `actions.Permission(table, p)`），视为缺陷**。
- 检查清单：
  - 路由组上有 `.Use(actions.PermissionAction())`（**整组挂**，不是 per-handler）
  - apis 通过 `actions.GetPermissionFromContext(c)` 取 `*DataPermission`，传给 service（**严禁传 `*gin.Context`**）
  - service `GetPage / GetList / Get / Update 写前查 / MarkRead` 全部加 `Scopes(actions.Permission(table, p))`
  - `Remove` 先按 scope 过滤 `Ids` 再级联删
  - 主模型嵌入 `common/models.ControlBy`（提供 `create_by`），或文档化自定义 owner 扩展
  - 5 路 dataScope（`"1"` ~ `"5"`）+ 跨用户越权 + `EnableDP=false` 短路 单/集成测试齐全
- 平台底座路由（用户/角色/部门/菜单/字典/日志/配置/sys-api/sysjob/module_registry/workflow 等）**豁免**，**不要**反向给底座挂中间件——会导致超管按部门切片，跨部门治理失效。完整豁免清单见 `PROJECT_CONVENTIONS.md` §1.6.1 D 节。
- 落地样板：`go-admin/app/admin/{router,service}/announcement.go`。C4 业务模块（`kingdee_customer` 等）按此样板照搬。

## 平台能力层已部分落地，但仍有待收口项

- `workflow / module_registry / attachment` 及对应最小前端验收层已经进入已提交历史。
- `统一业务操作日志最小规范` 已收口为已提交真相源：`go-admin/common/audit`（`Entry`/`Target` 契约 + `Log/LogCreate/LogUpdate/LogDelete` helper），文档见 `.ai-memory/backend-frontend-contracts.md` 的"业务操作日志最小契约"段。新业务模块应直接使用 `middleware.AuditLog*` helper，不再手拼 `AuditMeta`。
- 继续推进时应优先补齐真相源和承接层，不再把平台能力重新塞回单个业务模块内部。

## 阶段口径可能滞后

- 旧文档可能仍把 finance 或本地后端实验层写成当前主线，这是过时口径。
- 当前应以后端业务实验层已基本收口、回到 clone 基线附近为准。
- clone 基线本身已经包含同事融合进来的业务代码，不能把当前仓库描述成“已删掉业务代码”。
- 当前不建议重新在本地随意扩写 finance / feishu / kingdee / biz_action_log / `sys_user` 扩展类后端业务线。

## 阶段文档仍可能保留旧判断

- `platform-layer-audit.md`、`platform-capability-phase1.md`、`PROJECT_CONVENTIONS.md` 中保留了阶段性判断或工作草案。
- 这些文件可以作为追溯线索，但不能替代当前代码和正式迁移链。

## 旧历史文档仍可能存在

- `go-admin/README.md` 和 `go-admin/README.Zh-cn.md` 主要是上游项目说明，不代表当前工作区的真实运行方式。
- 当前仓库以根 README、`docs/` 和 `.ai-memory/` 为准。

## SKU 模块（C4）的运营落地需要手工补节点

- 迁移 `1779000000001_spu_workflow_seed.go` 在 `product_admin` 角色不存在时跳过节点种子。
- fresh install 场景下，迁移先于角色种子执行，需要后续手工进 `流程中心` → `定义管理` →
  'SPU 创建审核' 给定义补一个 `approver_type=role`、`approver_value=<product_admin role_id>` 的审批节点。
- 没有审批节点时 `Spu.SubmitForReview` 直接返回 "流程定义未配置审批节点"。
- 完整模块使用指南：[docs/sku-module-guide.md](/Users/linshen/Cursor/Full-stack-TestV1/docs/sku-module-guide.md)。

## SKU 模块审计 method 名稳定契约

- 审计日志的 `Method` 字段是历史日志、告警、大盘的稳定 key——一旦改名查询失效。
- SKU 模块当前的 method 契约（钉死在 `TestE2E_Spu_AuditMethod_Contract`）：
  - `admin.spu.insert` / `admin.spu.update` / `admin.spu.delete` / `admin.spu.submit`（apis/spu.go）
  - `platform.workflow.task.approve` / `platform.workflow.task.reject` / `platform.workflow.instance.withdraw`（platform/apis/workflow.go）
- 改 method 名前必须同步改测试和 `docs/sku-module-guide.md` 第 3.2 节，否则历史日志排查会断链。

## SKU dataScope 覆盖说明

- **服务层 SQLite 契约**：`go-admin/app/admin/service/spu_data_scope_test.go` 已覆盖 dataScope **1～5**（与 `announcement_data_scope_test.go` 同一 persona 拓扑）。
- **集成 E2E（shared sqlite + 完整迁移表）**：`go-admin/app/admin/service/spu_e2e_test.go` 中 `TestE2E_Spu_DataScope_*` 覆盖 scope **1、2、3、4、5**（含 `TestE2E_Spu_DataScope_DeptOnly` 对 1 与 3 的验证）。
- Playwright 浏览器 E2E 仍以列表页冒烟为主；细粒度 dataScope 以 Go 服务层测试为准。
