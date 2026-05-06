# 代码生成模板 v4

本目录是 go-admin 代码生成器（`POST /api/v1/sys/tables/info/gen/:tableId` 等）使用的模板集合。生成后的产物分布在：

- `app/{Package}/models/{Table}.go`            ← `model.go.template`
- `app/{Package}/service/dto/{Table}.go`       ← `dto.go.template`
- `app/{Package}/service/{Table}.go`           ← `no_actions/service.go.template`
- `app/{Package}/apis/{Table}.go`              ← `no_actions/apis.go.template`
- `app/{Package}/router/{Table}.go`            ← `no_actions/router_*.go.template`
- 前端 `api/{Package}/{table}.js` / `views/{Package}/{table}/index.vue`

## BaseService 范式（自 fs-2m0.3 起）

新模块的 service 默认通过嵌入 `baseservice.BaseService[models.{ClassName}]` 获得 CRUD 五件套，避免每个 service 重复样板：

```go
// app/{Package}/service/{Table}.go（生成产物）
type Widget struct {
    baseservice.BaseService[models.Widget]
}
```

`baseservice.BaseService[T]` 提供：

| 方法           | 接受参数（接口）                          | 行为                                        |
|----------------|-------------------------------------------|---------------------------------------------|
| `GetPage`      | `PageReq`, `*[]T`, `*int64`               | `MakeCondition + Paginate` 标准分页查询     |
| `Get`          | `IDReq`, `*T`                             | `First` 主键查询；`ErrRecordNotFound` 包装  |
| `Insert`       | `MutateReq[T]`                            | `Generate(&model)` 后 `Create`              |
| `InsertReturn` | `MutateReq[T]`, `*T`                      | 与 `Insert` 相同，额外回填新建后的 model    |
| `Update`       | `MutateReq[T] & IDReq`                    | `First` 加载 → `Generate` 覆盖 → `Save`     |
| `Remove`       | `IDReq`                                   | `Delete`，`GetId()` 可返回单值或 `[]int`    |

DTO 模板 (`dto.go.template`) 已天然产出满足这些接口的请求结构：

- `{Class}GetPageReq.GetNeedSearch()` + 嵌入 `dto.Pagination` → `PageReq`
- `{Class}{Insert,Update}Req.Generate(*T)` + `GetId()`        → `MutateReq[T]` (+ `IDReq`)
- `{Class}{Get,Delete}Req.GetId()`                             → `IDReq`

## 与旧模板的差异

| 维度                    | 旧模板                                               | 新模板（BaseService）                          |
|-------------------------|------------------------------------------------------|------------------------------------------------|
| service 文件大小        | ~110 行 5 个方法体                                   | ~17 行（仅类型嵌入 + 注释）                    |
| `*actions.DataPermission` 过滤 | service 内置                                  | 不内置；按需在业务 service 覆盖单个方法实现     |
| api 调用 service        | `s.GetPage(&req, p, ...)` 等需传 `p`                 | `s.GetPage(&req, ...)` 直接调用                |
| `actions` 包导入        | api 必须导入                                          | api 不再默认导入                                |

## 差异行为如何处理

外层 service 类型上声明同名方法即可覆盖 `BaseService` 的默认实现（Go 嵌入字段方法被外层同名方法 shadow）：

```go
type Widget struct {
    baseservice.BaseService[models.Widget]
}

// 覆盖 GetPage：增加 actions.Permission 数据权限 scope
func (s *Widget) GetPage(req *dto.WidgetGetPageReq, p *actions.DataPermission, list *[]models.Widget, count *int64) error {
    var data models.Widget
    return s.Orm.Model(&data).
        Scopes(
            cDto.MakeCondition(req.GetNeedSearch()),
            cDto.Paginate(req.GetPageSize(), req.GetPageIndex()),
            actions.Permission(data.TableName(), p),
        ).
        Find(list).Limit(-1).Offset(-1).
        Count(count).Error
}
```

> 注意：`BaseService` 不接受 `*actions.DataPermission` 参数，覆盖方法可以使用任意签名。但 api 层调用时需相应调整。

其它常见覆盖场景：

- 多表 join：覆盖 `GetPage` / `Get`，`b.Orm.Joins(...).Where(...).Find(...)`
- 自定义 where：同上
- 软删除回滚：覆盖 `Remove`，用 `b.Orm.Unscoped().Delete(...)` 或 `Update("deleted_at", nil)`
- 新建后回填字段：调用方使用 `s.InsertReturn(&req, &out)`

## 验证

`fs-2m0.3` 在 `app/_genverify/`（已清理）渲染出 `model + dto + service + apis`，对 SQLite 内存库执行了五件套 CRUD 全部通过。`SysPost` 试点 (`fs-2m0.1`) 的 `app/admin/service/sys_post_test.go` 持续守护这一范式。

## 模板文件清单

```
template/v4/
├── README.md                                ← 本文件
├── model.go.template
├── dto.go.template
├── js.go.template
├── vue.go.template
├── actions/
│   ├── router_check_role.go.template
│   └── router_no_check_role.go.template
└── no_actions/
    ├── apis.go.template                     ← 已使用 BaseService 调用约定
    ├── service.go.template                  ← 已使用 BaseService 嵌入
    ├── router_check_role.go.template
    └── router_no_check_role.go.template
```
