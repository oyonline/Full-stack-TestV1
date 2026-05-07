# 用户头像系统说明

更新时间：2026-05-07

## 文档定位

- 这份文档描述当前已提交的头像方案。schema 与后端 API 字段回路已完整收口为正式真相源。
- 部分前端组件（`user-avatar.ts` / `user-avatar.vue`）仍属本地 working tree，
  这份文档只对"已提交"部分给正式描述。

## 字段真相源（已提交）

`sys_user` 表通过 migration `1775000000000_user_avatar_profile.go` 引入两列：

| 列            | 类型           | 含义                                                                  |
| ------------- | -------------- | --------------------------------------------------------------------- |
| `avatar_type` | `varchar(16)`  | 头像类型，取值 `image` / `letter` / 空。决定前端渲染图片还是字母色块。 |
| `avatar_color`| `varchar(16)`  | hex 颜色（例如 `#1D4ED8`）。letter 模式下的字母色块背景色。            |

migration 同时做了一次回填：

```sql
UPDATE sys_user
SET avatar_type = 'image'
WHERE avatar <> '' AND (avatar_type IS NULL OR avatar_type = '');
```

避免老用户首屏头像突然回退成字母色块。

模型映射：
- `go-admin/app/admin/models/sys_user.go` 的 `SysUser` 结构含 `AvatarType` / `AvatarColor` 字段。
- `go-admin/cmd/migrate/migration/models/sys_user.go` 的 migration 私有副本同步含两字段，
  保证 AutoMigrate 与 AddColumn 推断列名一致。
- 单测：`go-admin/cmd/migrate/migration/version/1775000000000_user_avatar_profile_test.go`
  用 in-memory sqlite 验证加列 + 回填，并保证幂等。

## 当前真实链路

- 用户头像配置挂在 `sys_user`（`avatar` / `avatar_type` / `avatar_color` 三列）。
- `GET /api/v1/getinfo`、`PUT /api/v1/user/profile`、`POST /api/v1/user/avatar` 是头像相关入口。
- 前端字母头像规则："姓名优先、登录账号兜底"。
- 个人中心承担头像模式切换、背景色设置和图片上传入口。

## API 数据回路（已收口）

后端 API 已完整收口 `avatar_type` / `avatar_color` 的写入与读出（fs-ua0 + my-3fo）：

- `GET /api/v1/getinfo`：响应 map 包含 `avatarType` / `avatarColor`。
- `PUT /api/v1/user/profile`：`UpdateSysUserProfileReq` 含两字段，`Generate` 写回 model，
  service 显式 `Updates({introduction, avatar_type, avatar_color})`。
- `POST /api/v1/user/avatar`：上传成功后只覆盖 `avatar` + `avatar_type='image'`，
  **保留原 `avatar_color`**；响应回传完整三元组 `{avatar, avatarType, avatarColor}`，
  前端 store 切回 letter 模式时仍能复用原背景色。
- admin CRUD：`SysUserInsertReq` / `SysUserUpdateReq` 含 `avatarType` / `avatarColor`，
  `Generate` 写回 model；列表 / 详情通过 `models.SysUser` 序列化天然带出两字段。

## 残留待办

- [user-avatar.ts](../vue-vben-admin/apps/web-antd/src/utils/user-avatar.ts) 与
  [user-avatar.vue](../vue-vben-admin/apps/web-antd/src/components/user-avatar.vue) 仍属本地内容，
  按既有节奏入提交链。
- [docs/sql/20260328_sys_user_avatar_profile.sql](sql/20260328_sys_user_avatar_profile.sql)
  仅是 migration 的早期参考脚本，**不再是真相源**；migration 进入 commit 后以 migration 为准。
