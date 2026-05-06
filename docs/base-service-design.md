# BaseService[T] 通用服务基类设计（spike：fs-2m0.1）

> 状态：spike（试点）。已落 `BaseService[T any]` + 1 个示范 service（`SysPost`）+ 单元测试。
> 批量迁移 44 个 service 的工作在 fs-2m0.2，代码生成模板更新在 fs-2m0.3。

## 1. 解决的问题

`go-admin/app/*/service/*.go` 当前 44 个 model 各自手写 GetPage/Get/Insert/Update/Remove
五件套。每个新业务模块都要复制一遍同样的查询模板，方法体除了 model 类型几乎完全一致。
脚手架"对接业务模块成本"的主要 boilerplate 都集中在这里。

目标：抽出泛型 `BaseService[T any]`，业务 service 只在差异点上覆盖；同时不破坏已有调用约定
（API 层 / DTO 接口 / 错误信息）。

## 2. 设计

### 2.1 包位置

新增包：`go-admin/common/baseservice`。

不放在已有的 `go-admin/common/service` 包，因为该包名会和 SDK 的
`github.com/go-admin-team/go-admin-core/sdk/service` 冲突，业务 service 文件普遍以
`service "github.com/go-admin-team/go-admin-core/sdk/service"` 显式导入，
新增 `baseservice` 名不冲突，import 语句也最直观。

### 2.2 核心类型

```go
type BaseService[T any] struct {
    service.Service  // 来自 SDK：保留 Orm / Log / Cache / AddError 等字段
}
```

业务 service 通过嵌入获得 5 件套：

```go
type SysPost struct {
    baseservice.BaseService[models.SysPost]
}
```

此时 `s.Orm`、`s.Log`、`s.Service`（用于 `MakeService(&s.Service)`）依然按 Go 的
"嵌入字段提升"规则可用，**API 层无需任何改动**。

### 2.3 DTO 接口约束

历史 DTO 已经实现了下面的方法集合，BaseService 只把它们抽成接口：

```go
type PageReq interface {
    GetNeedSearch() interface{}
    GetPageSize() int
    GetPageIndex() int
}
type IDReq interface {
    GetId() interface{}        // int / int64 / []int 都可（GORM 自适配）
}
type MutateReq[T any] interface {
    Generate(*T)
}
```

因此存量 DTO **不需要任何改动** 即可通过类型检查。

### 2.4 五件套行为

| 方法           | 行为                                                                                | SQL 形态                                                       |
| -------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `GetPage`      | `cDto.MakeCondition` + `cDto.Paginate` 组合 scopes                                  | `Find(list).Limit(-1).Offset(-1).Count(count)` —— 与历史一致   |
| `Get`          | 按主键 `First`，`gorm.ErrRecordNotFound` → 中文 "查看对象不存在或无权查看"          | `SELECT ... WHERE id=? LIMIT 1`                                |
| `Insert`       | `c.Generate(&data)` 后 `Create(&data)`                                              | `INSERT ...`                                                   |
| `InsertReturn` | 同 `Insert` 但把保存后的 model（含自增主键）回填给 `out`                            | 同上                                                           |
| `Update`       | `First(&model, id)` → `c.Generate(&model)` → `Save(&model)`，`RowsAffected==0` 报错 | `SELECT ... + UPDATE ...`                                      |
| `Remove`       | `Delete(&data, c.GetId())`，软删除 model 自动写 `deleted_at`                        | `DELETE` 或 `UPDATE deleted_at=...`                            |

错误信息全部沿用历史 service 的中文文案（"查看对象不存在或无权查看"、"无权更新该数据"、
"无权删除该数据"），保证前端不需要任何感知。

## 3. 覆盖范式（如何在差异点上扩展）

Go 没有虚拟方法分发，**外层 struct 的同名方法 shadow 嵌入字段方法**；只要业务 service
直接被 API 调用（`s := service.SysX{}; s.GetPage(...)`），覆盖就生效。

### 3.1 完全覆盖（推荐用于复杂场景）

直接重写整个方法。例如 `sys_role` 在 `Insert` 中要联动 `casbin` 写权限规则：

```go
type SysRole struct {
    baseservice.BaseService[models.SysRole]
}

// 完全覆盖：BaseService.Insert 不再被调用
func (e *SysRole) Insert(c *dto.SysRoleInsertReq) error {
    // 1) 业务前置（角色编码唯一性、casbin 规则等）
    // 2) 自己写 e.Orm.Create(...)
    // 3) 业务后置（写 sys_role_menu 关联表等）
    return nil
}
```

适用：多表 join、跨表事务、外部副作用（发消息、写日志、刷新缓存）。

### 3.2 局部增强（前置 / 后置 hook）

不需要复杂逻辑时，可以在覆盖里调用 `b.BaseService.Method(...)`：

```go
func (e *SysFoo) Insert(c *dto.SysFooInsertReq) error {
    // 前置校验
    if err := validate(c); err != nil { return err }
    // 复用默认实现
    if err := e.BaseService.Insert(c); err != nil { return err }
    // 后置副作用
    return notifyDownstream(c)
}
```

`b.BaseService.Insert` 是 Go 显式访问嵌入字段方法的语法。

### 3.3 自定义 Where / 多表 join（仅 GetPage）

`MakeCondition` 已经能从 DTO struct tag 解析大多数条件（`type:contains`/`type:exact` /
`column:` / `table:` 等）。如果 DTO struct tag 已能表达（即使是跨表 join），**不需要覆盖**。
只有以下情况建议覆盖：

- 需要 `LEFT JOIN ... ON ...` 之外的复杂联表（窗口函数、子查询）；
- 需要按当前用户角色 / 数据权限动态拼 where；
- 需要 `Preload` 关联的子集合。

此时直接覆盖 `GetPage` 用 `b.Orm.Model(...).Joins(...).Where(...).Find(...).Count(...)` 重写。
不要在 `BaseService` 内增加配置 API（hook 列表、scopes 注入参数等），那样会把"如何扩展"
变成另一个需要文档化的维度，得不偿失。

### 3.4 软删除

GORM 通过 model 内嵌 `gorm.DeletedAt`（`go-admin/common/models.ModelTime` 已经包含）即可自动启用。
`Remove` 在软删除模型上自动改成 `UPDATE deleted_at=...`，`Get`/`GetPage` 自动忽略已删除记录。
**BaseService 不区分硬删除 / 软删除**，差异由 model 自身决定。

如果某个 model 需要"硬删除一行 + 同时跑额外 SQL"（如清理外键关联），覆盖 `Remove` 即可。

### 3.5 ID 回填

历史 API 在 Insert 后用 `req.GetId()` 返回新建 ID，但 DTO 在 Insert 前 ID=0，所以这条返回
**长期返回 0**。这是既有 bug，本次 spike **不修复**（不引入语义变化以便迁移）。
如果新代码需要真实 ID，应使用 `InsertReturn(req, &saved)` 然后从 `saved.PrimaryKey` 读。

## 4. 试点：`SysPost` 改造前后

### 改造前（104 行）

`go-admin/app/admin/service/sys_post.go` 含 5 个手写方法，方法体除了 model 类型外几乎模板化。

### 改造后（10 行）

```go
package service

import (
    "go-admin/app/admin/models"
    "go-admin/common/baseservice"
)

type SysPost struct {
    baseservice.BaseService[models.SysPost]
}
```

API 层（`apis/sys_post.go`）零改动。

### 测试覆盖（`sys_post_test.go`）

用 in-memory sqlite + 真实 GORM 跑了：

- `TestSysPost_BaseServiceCRUD`：Insert × 2 → GetPage 计数 → Get 命中 / 未命中 → Update → 软删除 → 删除后 Get/GetPage → 删除不存在 ID 的错误路径。
- `TestSysPost_InsertReturnPopulatesID`：验证 `InsertReturn` 能回填自增主键。
- `TestSysPost_GetPageRespectsPagination`：验证 `pageIndex/pageSize` 在 BaseService 默认实现下的分页边界。

## 5. 边界 / 已知限制

| 场景                                  | 处理方式                                                                     |
| ------------------------------------- | ---------------------------------------------------------------------------- |
| 多表 join / 自定义 where              | 覆盖 `GetPage`（见 3.3）                                                     |
| 跨表事务（写关联表）                  | 完全覆盖 Insert/Update/Remove；BaseService 不内置事务装饰器                  |
| 软删除 vs 硬删除                      | 由 model 自身决定（是否含 `gorm.DeletedAt`）；BaseService 不感知             |
| Insert 后回填新 ID 给前端             | 用 `InsertReturn`，不要依赖 `req.GetId()`（旧链路 bug）                      |
| DTO 约束（`Generate` / `GetId`）      | 已和现存 DTO 完全兼容；新模块按现有 DTO 模板编写即可                         |
| 一对多 Preload                        | 覆盖 `Get`/`GetPage`，自己写 `Preload(...)`                                  |
| 数据权限（按部门 / 角色过滤）         | 覆盖 `GetPage`，把当前用户上下文从外部传入；BaseService 不内置数据权限 hook |
| 主键非 int                            | `IDReq.GetId() interface{}` 已开放，GORM `First`/`Delete` 自适配             |

## 6. 不做什么（防止过度设计）

- **不引入泛型化的 hook 列表 / scopes 注入 API**：覆盖整个方法比叠加 hook 更清晰，迁移更可预期。
- **不内置事务装饰器**：跨表事务的边界由业务 service 决定，BaseService 只覆盖单表 CRUD。
- **不修改 DTO 接口**：所有现存 DTO 不改一行即可使用 BaseService。
- **不一次性迁移 44 个 service**：spike 阶段只做 1 个，等 review 通过再批量推进（fs-2m0.2）。

## 7. 后续

- **fs-2m0.2**：批量迁移 44 个 service。先把"无差异"的服务（与 SysPost 同形态）一刀切；
  含自定义逻辑的 service（如 `sys_role`、`sys_user`、`sys_dept`）单独评估。
- **fs-2m0.3**：更新 `cmd/api/template`（如有）/ 代码生成模板，让新模块默认产出嵌入
  `BaseService[T]` 的 service 文件。
