# SKU 模块使用指南

本指南面向需要使用、运维或扩展 SKU 模块的工程师，配合 `.ai-memory/module-map.md`
中的"平台能力的业务使用者"小节看。重点回答三件事：

1. 模块在做什么、由什么组成（数据层 / 后端 / 前端 / 流程）
2. 上线前必须配置好哪些角色和流程节点（运营落地清单）
3. 端到端验收和审计排查应该看哪几张表 / 哪几条 audit method

> 本模块是项目"平台能力的第二个使用者"（第一个是 announcement）。
> 任何修改都需要保证不破坏 platform.workflow / dataScope / 审计日志 三条共享链路。

---

## 1. 模块组成

### 1.1 数据层（C4-A）

迁移：`go-admin/cmd/migrate/migration/version/1779000000000_sku_module.go`

落库内容：

| 表 | 角色 |
|----|------|
| `sku_category` | 类目树（自引用 ParentId） |
| `sku_brand` | 品牌 |
| `spu` | 标准产品单元（含 `status`, `workflow_instance_id`, `submitted_at`, `approved_at`, `creator_id`, `dept_id`） |
| `sku` | 销售单元，挂在 SPU 下 |

同时落库：

- `sys_menu`：'产品中心' 顶级菜单 + 4 个 C 子菜单（SPU/SKU/类目/品牌）+ 16 个按钮
- `sys_api`：19 行接口元数据
- `sys_menu_api_rule`：菜单与接口的桥接

> 迁移幂等。重跑不会重复插入。

工作流定义种子：`1779000000001_spu_workflow_seed.go`

- 创建 `spu_create_review` 流程定义
- 如 `product_admin` 角色已存在，自动给定义挂一个角色审批节点
- 如不存在，跳过节点种子（迁移不卡死，后台手工补）

### 1.2 后端（C4-B）

模块 | 关键文件
---|---
DTO | `app/admin/service/dto/spu.go`、`sku.go`、`sku_category.go`、`sku_brand.go`
Service | `app/admin/service/spu.go`、`sku.go`、`sku_category.go`、`sku_brand.go`
SPU 终态回写 | `app/admin/service/spu_workflow_handler.go`
API | `app/admin/apis/spu.go`、`sku.go`、`sku_category.go`、`sku_brand.go`
Router | `app/admin/router`（菜单 path 与 sys_api 一一对应）

关键服务约束：

- **Spu.Update** 拒绝审核中（`status=Reviewing`）的 SPU 直接编辑——必须先撤回审批。
- **Spu.Remove** 拒绝审核中 SPU 的删除；SKU 随 SPU 软删（GORM `Delete`）。
- **Spu.SubmitForReview** 只在 `Draft` 或 `Rejected` 时允许提交；提交成功后
  - 创建 `wf_instance` + `wf_business_binding` + 第一个 pending task
  - 写回 SPU：`status=Reviewing`、`submitted_at`、`workflow_instance_id`

终态回写规则（来自 `spu_workflow_handler.go`）：

| 终态 | SPU.status | 写回字段 |
|------|------------|----------|
| `approved` | `SpuStatusApproved` (3) | `approved_at=now` |
| `rejected` | `SpuStatusRejected` (4) | — |
| `cancelled` | `SpuStatusDraft` (1)（撤回视为回到草稿） | — |

### 1.3 前端（C4-C）

| 页面 | 路径 | Component |
|------|------|-----------|
| 类目管理 | `/product/category` | `admin/sku-category/index` |
| 品牌管理 | `/product/brand` | `admin/sku-brand/index` |
| SKU 主管理 | `/product/sku` | `admin/sku/index` |
| SPU 主管理 | `/product/spu` | `admin/sys-spu/index` |

API client：`vue-vben-admin/apps/web-antd/src/api/admin/{spu,sku,skuCategory,skuBrand}.ts`

> 前端走后端菜单（backend access mode）。component 路径必须与代码内真实文件对应，
> 否则前端落到 not-found。详见 `.ai-memory/known-issues.md` "菜单组件路径必须真实可映射"。

---

## 2. 上线前的运营落地清单

### 2.1 角色

至少需要预创建 / 验证两个角色：

| RoleKey | 名称 | 关键权限 |
|---------|------|----------|
| `product_admin` | 产品管理员 | `admin:spu:approve`、`admin:spu:reject`、`admin:spu:list`、`admin:spu:edit`（看待办、审批） |
| `product_operator` | 产品操作员 | `admin:spu:add`、`admin:spu:edit`、`admin:spu:submit`、`admin:spu:list` |

> dataScope：`product_operator` 通常配 `data_scope=3`（本部门）或 `data_scope=5`（仅本人），
> `product_admin` 配 `data_scope=1`（全部，方便审批）。

### 2.2 流程定义

迁移已落 `spu_create_review` 定义：

- `definition_key=spu_create_review`
- `module_key=admin`
- `business_type=spu`
- `status=2`（启用）

**必要手工动作**（管理员后台进 `流程中心` → `定义管理` → 'SPU 创建审核'）：

1. 确认存在至少一个 `approve` 类型节点
2. 节点的 `approver_type=role`、`approver_value=<product_admin role_id>`
3. 如果迁移时 `product_admin` 不存在，节点不会被自动种子——必须手工补

> 没有审批节点会让 `Spu.SubmitForReview` 直接返回 "流程定义未配置审批节点"。

### 2.3 验收测试场景

按 `bd show my-pwj` 中 C4-D 验收清单，至少跑下面 5 条：

1. `product_operator` 创建 SPU + 多 SKU + 上传主图 / 详情图 / 富文本描述 → 提交审核
2. `product_admin` 查看待办 → 通过 → SPU.status 推进到 3（已通过）+ `approved_at` 写回
3. `product_admin` 驳回 → SPU.status=4（已驳回）
4. 创建人改一改 → 再提交 → 新建 wf_instance + 旧 binding 被替换 + status 推回 Reviewing
5. dataScope：`data_scope=3` 的 `product_operator` 列表 SPU 时仅看到本部门同事的记录

---

## 3. 验证与排查

### 3.1 本地端到端跑通

```bash
./scripts/check-local.sh        # 全栈：go test + pnpm typecheck + pnpm build:local
go test ./app/admin/service/... # 单跑后端 SPU/SKU 相关测试
go test -run TestE2E_Spu ./app/admin/service/  # 仅跑 C4-D 端到端测试
```

后端 e2e 入口文件：`go-admin/app/admin/service/spu_e2e_test.go`，覆盖：

- `TestE2E_Spu_Submit_Approve_Writeback`：完整提交→审批→回写
- `TestE2E_Spu_Submit_Reject_Writeback`：驳回路径
- `TestE2E_Spu_Resubmit_After_Reject`：驳回后重提，新 wf_instance + 旧 binding 替换
- `TestE2E_Spu_DataScope_DeptOnly`：dataScope=3 仅看本部门
- `TestE2E_Spu_AuditMethod_Contract` / `TestE2E_Spu_AuditEmit_OnSubmit`：审计 method 名稳定契约

### 3.2 审计日志（sys_opera_log）

SPU 关键操作产出的 `Method` 值（即排查时的 `where method=...`）：

| 场景 | Method | 来源 |
|------|--------|------|
| 创建 SPU | `admin.spu.insert` | `app/admin/apis/spu.go` |
| 修改 SPU | `admin.spu.update` | 同上 |
| 删除 SPU | `admin.spu.delete` | 同上 |
| 提交审核 | `admin.spu.submit` | 同上 |
| 审批通过 | `platform.workflow.task.approve` | `app/platform/apis/workflow.go` |
| 审批驳回 | `platform.workflow.task.reject` | 同上 |
| 流程撤回 | `platform.workflow.instance.withdraw` | 同上 |

> 这些字符串被 `TestE2E_Spu_AuditMethod_Contract` 钉死。改名需要同步改测试和这份文档。

排查路径：

```sql
-- 谁在什么时候改过 SPU id=123
SELECT oper_time, oper_name, method, oper_param
FROM sys_opera_log
WHERE business_types = 'spu'
  AND oper_param LIKE '%"id":123%'
ORDER BY oper_time DESC;

-- SPU id=123 的全部审批动作
SELECT oper_time, method, oper_name, oper_param
FROM sys_opera_log
WHERE method IN (
  'admin.spu.submit',
  'platform.workflow.task.approve',
  'platform.workflow.task.reject',
  'platform.workflow.instance.withdraw'
)
ORDER BY oper_time DESC;
```

### 3.3 工作流状态排查

SPU 当前流程实例：

```sql
SELECT spu_id, spu_code, status, workflow_instance_id, submitted_at, approved_at
FROM spu WHERE spu_id = ?;

SELECT * FROM wf_business_binding
WHERE module_key='admin' AND business_type='spu' AND business_id=?;

SELECT * FROM wf_action_log
WHERE instance_id=? ORDER BY log_id ASC;
```

> 如果 SPU.status 卡在 2（Reviewing）但 `wf_instance.status='approved'`，说明
> 终态回写 handler 没跑——常见原因是 init() 没注册（包未被 import）或
> `business_id` 不可解析为 int。`onSpuWorkflowTerminal` 对非法 id 静默返回，
> 这种情况手工 update SPU.status 后再排查。

---

## 4. 扩展指南

### 4.1 新业务接入审批流（参考 SKU 模块）

最小契约 3 步：

1. **建表**：业务表至少包含 `status`（业务态）、`workflow_instance_id`、提交人/部门字段。
2. **写流程定义种子**：迁移文件参考 `1779000000001_spu_workflow_seed.go`，落 `wf_definition`
   + `wf_definition_node`，`module_key` / `business_type` 自定（与 `Workflow.Start` 入参一致）。
3. **注册终态回写**：在业务模块 init() 调用：

   ```go
   platformService.RegisterTerminalHandler("<your_business_type>", func(tx *gorm.DB,
       binding *platformModels.WorkflowBusinessBinding, terminalStatus string) error {
       // 按 terminalStatus 把业务态写回业务表，错误返回会触发 platform 事务回滚
   })
   ```

   handler 必须在事务内运行（platform 已包好），不要在 handler 里再起新事务。

### 4.2 新业务接入 dataScope

参考 `app/admin/service/spu.go::GetPage` 与 `.ai-memory/procedural/data-permission-wiring.md`：

- 业务表必须有 `create_by` 列（来自 `models.ControlBy`）
- service.GetPage 走 `actions.Permission(tableName, p)` scope
- API 层用 `actions.GetPermissionFromContext(c)` 取 DataPermission

> 不要绕开 `actions.Permission`，否则跨部门数据会泄漏。

### 4.3 新业务接入审计

API handler 在写动作成功后调用：

```go
middleware.AuditLogCreate(c, "<标题>", middleware.AuditTarget{...}, after, "<method>")
middleware.AuditLogUpdate(c, ...)
middleware.AuditLogDelete(c, ...)
middleware.AuditLog(c, middleware.AuditEntry{...})  // 自定义 Action（如 start/approve）
```

`<method>` 用稳定的点分路径（`<层>.<域>.<动作>`），上线后不再改名——日志 / 告警 / 大盘
都会按这个 key 查询。

---

## 5. 已知约束

- **手工补流程节点的窗口期**：迁移在 `product_admin` 不存在时跳过节点种子。
  如果生产环境 fresh install，迁移先于角色种子执行，需要后续手工进流程中心补节点。
- **platform.workflow 撤回视为回到 Draft**：业务方如果想区分"撤回 by 提交人"与
  "驳回 by 审批人"，需要靠 `wf_action_log.action` 字段拆分，不能只看 SPU.status。
- **dataScope=2 (custom) 与 SPU 还没做端到端冒烟**：当前 e2e 只覆盖 dataScope=1/3。
  自定义角色绑定部门集合的口径应与 announcement C7-5 用例对齐，参考
  `app/admin/service/announcement_data_scope_test.go`。
- **审计 method 一旦改名，会让历史日志查询失效**。改名前先在 `TestE2E_Spu_AuditMethod_Contract`
  改契约，再改 apis/ 文件，再同步本指南第 3.2 节。

---

## 参考链接

- `.ai-memory/module-map.md` —— 模块全景与平台层使用者清单
- `.ai-memory/procedural/data-permission-wiring.md` —— dataScope 接入步骤
- `.ai-memory/audits/data-permission-routing.md` —— 5 路 dataScope 路由审计
- `docs/local-dev-handoff.md` —— 本地联调与依赖
- `go-admin/app/platform/service/workflow.go` —— 平台审批流入口
- `go-admin/app/platform/service/workflow_callbacks.go` —— 终态回写注册中心
