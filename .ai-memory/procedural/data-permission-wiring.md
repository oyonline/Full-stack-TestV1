# 数据权限接入手册（C4 业务模块参考）

> 关联：phase2 数据权限（C7-1 ~ C7-7）
> 落地样板：announcement（C7-3）
> 路由策略真相源：`.ai-memory/audits/data-permission-routing.md`（C7-2 产出）
> 代码核心：`go-admin/common/actions/permission.go`

数据权限（dataScope）让"业务数据"按当前用户的角色配置切片为
"全部 / 自定义 / 本部门 / 本部门及以下 / 仅本人"。本文档把 announcement
的接入步骤抽成模板，给后续 C4 业务模块（如 kingdee_customer 等）参考。

---

## 1. 三个独立部件，必须同时齐备

| 部件 | 位置 | 作用 |
|------|------|------|
| `config.ApplicationConfig.EnableDP` | `settings*.yml` 全局开关 | 关掉则 `actions.Permission` 短路放行；过渡期保护 |
| `actions.PermissionAction()` | gin middleware（router） | 读 JWT user，查 `data_scope/dept_id/role_id`，写到 `c.Set(PermissionKey, *DataPermission)`。**只注入上下文，不过滤**。 |
| `actions.Permission(tableName, p)` | GORM scope（service） | 真正按 dataScope 拼 `WHERE create_by IN (...)`。**没在 service 调用 == 没接入**。 |

> 三件事缺一不可：路由没挂中间件 → service 里 p 是零值 → Permission 走 default
> 不过滤；service 没 Scopes → 中间件白挂；EnableDP=false → 全部短路。

---

## 2. dataScope 取值与 SQL（来自 `permission.go`）

| 值 | 含义 | SQL 片段 |
|----|------|---------|
| `"1"` 或空 | 全部 | 不加过滤 |
| `"2"` | 自定义（按角色绑定的部门集合） | `create_by IN (sys_role_dept ⋈ sys_user where role_id=?)` |
| `"3"` | 本部门 | `create_by IN (SELECT user_id FROM sys_user WHERE dept_id=?)` |
| `"4"` | 本部门及以下 | `create_by IN (SELECT user_id FROM sys_user WHERE dept_id IN (SELECT dept_id FROM sys_dept WHERE dept_path LIKE '%/${dept}/%'))` |
| `"5"` | 仅本人 | `create_by = ?` |

> 关键约束：业务模型的主表必须有 `create_by` 字段（嵌入 `common/models.ControlBy`），
> 否则上述 SQL 没意义。announcement 已经嵌入 ControlBy。

---

## 3. 接入步骤（以 announcement 为样板）

### 3.1 router：挂 `PermissionAction()` 中间件

`go-admin/app/admin/router/announcement.go`：

```go
import (
    "go-admin/common/actions"
    "go-admin/common/middleware"
)

r := v1.Group("/announcement").
    Use(authMiddleware.MiddlewareFunc()).
    Use(middleware.AuthCheckRole()).
    Use(actions.PermissionAction())   // ← 新增
```

> **业务路由组级挂载**就够了。不要 per-handler 挂（jobs 模块那种是历史遗留）。

### 3.2 apis：从 ctx 取 `*DataPermission` 传给 service

```go
import "go-admin/common/actions"

func (e Announcement) GetPage(c *gin.Context) {
    s := service.Announcement{}
    req := dto.AnnouncementPageReq{}
    if err := e.MakeContext(c).MakeOrm()....Errors; err != nil {
        ...
    }
    p := actions.GetPermissionFromContext(c)   // ← 取注入的 DataPermission
    if err := s.GetPage(&req, p, &list, &count, user.GetUserId(c)); err != nil {
        ...
    }
}
```

> **不要把 `*gin.Context` 传到 service**。service 只接收 `*actions.DataPermission`
> 这个简单结构体；这样 service 不依赖 gin，单测好做。

### 3.3 service：每个读路径加 `actions.Permission` scope

```go
import "go-admin/common/actions"

// 列表查询：
q := e.Orm.Model(&data).
    Scopes(
        cDto.MakeCondition(c.GetNeedSearch()),
        actions.Permission(data.TableName(), p),   // ← 新增
    )

// 详情读、写前查（Update/Delete/MarkRead）：
if err := tx.Model(&MyModel{}).
    Scopes(actions.Permission((&MyModel{}).TableName(), p)).
    First(&existing, id).Error; err != nil { ... }
```

> **签名约定**：在 list/detail/get 之外，所有"基于 ID 找记录的写前查"也都
> 要加这个 scope，否则 dataScope=5 的用户能 PUT/DELETE/标记已读他人记录。

### 3.4 批量删除：先按 scope 过滤 ID 集，再级联删

announcement 的 `Remove` 模式：

```go
func (e *Announcement) Remove(c *dto.AnnouncementDeleteReq, p *actions.DataPermission) error {
    var allowed []int64
    if err := e.Orm.Model(&Model{}).
        Scopes(actions.Permission(table, p)).
        Where(table+".id IN ?", c.Ids).
        Pluck(table+".id", &allowed).Error; err != nil {
        return err
    }
    if len(allowed) == 0 {
        return errors.New("不存在或无权删除")
    }
    return e.Orm.Transaction(func(tx *gorm.DB) error {
        // 用 allowed 而不是 c.Ids 做级联删
    })
}
```

> 不这样做的话，dataScope=5 的用户传 `[ann1.Id, ann3.Id]` 会把 ann3 一起删掉。

---

## 4. 业务自带的可见性（announcement_scope）与 dataScope 正交

公告自带"按部门可见"机制（`announcement_scope` 表 + `OnlyVisible` 参数）：
- 这是**展示规则**：公告作者指定"我希望让哪些部门看到"。
- dataScope 是**读权限规则**：当前用户角色被允许读"哪些 create_by 创建的记录"。

两套机制**叠加 AND**：一条记录要出现在结果中，必须**同时**满足两边。
不要在新模块里把 dataScope 等价替换业务自带的可见性，反之亦然。

---

## 5. 测试规范

`announcement_permission_test.go` 是参考。要点：

1. 测试开头 `config.ApplicationConfig.EnableDP = true`，结尾恢复（避免污染其他测试）。
2. 用 in-memory sqlite + `db.AutoMigrate` 装 `Announcement / SysUser / SysDept`，
   手建 `sys_role_dept`（many2many 关联表无独立 model）。
3. 5 个 dataScope 都要测：scope=1/2/3/4/5。
4. 至少覆盖：
   - 列表（GetPage）每个 scope 的可见集合
   - 越权读（Get scope=5 用户读他人记录 → 返回不存在）
   - 越权写（Update / Remove / MarkRead 同上）
   - EnableDP=false 时短路放行（兼容过渡期）
   - 业务自带可见性 × dataScope 正交（如果业务有自带可见性）

---

## 6. 不要做的事

- **不要把 `*gin.Context` 传到 service**：增加耦合，单测难做。仅传 `*DataPermission`。
- **不要在写接口上做 scope 过滤**（POST/Insert）：写按按钮权限控；scope 是读侧概念。
- **不要在平台底座路由（用户/角色/部门/菜单/字典/日志/配置等）挂 `PermissionAction()`**：
  这些是超管视角，按 dataScope 切会导致"创建用户的人不在我部门→看不见该用户"等
  反直觉行为。详见 C7-2 路由策略表。
- **不要忘记给非 search 字段加 `search:"-"`**：`MakeCondition` 走的 `search.ResolveSearchQuery`
  会对没有 `search` 标签的字段递归 reflect，遇到 int/string 会 panic。
  AnnouncementPageReq 的 `OnlyValid` / `OnlyVisible` 加 `search:"-"` 收口（C7-3 顺手修）。

---

## 7. 落地清单（C4 业务模块进入时）

- [ ] router 加 `actions.PermissionAction()`
- [ ] apis 各 handler 调 `actions.GetPermissionFromContext(c)` 并传给 service
- [ ] service 各读方法接收 `*actions.DataPermission` 入参
  - [ ] List/GetPage 加 `Scopes(actions.Permission(table, p))`
  - [ ] Get 详情加 `Scopes(actions.Permission(table, p))`
  - [ ] Update 写前查加 scope
  - [ ] Delete/Remove 按 scope 过滤 IDs 后再删
  - [ ] 其他基于 ID 的写操作（如 MarkRead 类）都加 scope 校验
- [ ] 单元测试覆盖 5 个 scope + EnableDP=false 短路 + 业务自带可见性正交（如有）
- [ ] PageReq DTO 中所有非搜索字段加 `search:"-"`
