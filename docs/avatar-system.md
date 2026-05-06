# 用户头像系统说明

更新时间：2026-05-06

## 文档定位

- 这份文档描述当前已提交的头像方案。schema 与基础后端字段已收口为正式真相源。
- 部分前端组件（`user-avatar.ts` / `user-avatar.vue`）和完整的 profile API 字段回路仍然在落地中，
  这份文档只对"已提交"部分给正式描述，对"仍在落地"部分明确标注本地状态。

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

## 仍未收口的部分

下列项仍属于本地 working tree，不能写成正式状态：

- `getinfo` / `user/profile` / `user/avatar` API 与 DTO 层 **尚未把 `avatar_type` / `avatar_color`
  写入 / 读出**。schema 已就绪但 API 数据回路还需要补一层（DTO 增字段、Generate 写回 model、
  getinfo 返回 map 加两个 key）。这是 fs-2ie 之外的工作，不在本次范围。
- [user-avatar.ts](../vue-vben-admin/apps/web-antd/src/utils/user-avatar.ts) 与
  [user-avatar.vue](../vue-vben-admin/apps/web-antd/src/components/user-avatar.vue) 仍属本地内容。
- [docs/sql/20260328_sys_user_avatar_profile.sql](sql/20260328_sys_user_avatar_profile.sql)
  仅是 migration 的早期参考脚本，**不再是真相源**；migration 进入 commit 后以 migration 为准。

## 如果继续推进

- 后端：把 `UpdateSysUserProfileReq` / `UpdateSysUserAvatarReq` 加上 `AvatarType` / `AvatarColor`，
  在 `Generate(model *SysUser)` 中写回；`getinfo` 返回 map 加 `avatarType` / `avatarColor` key。
- 前端 working tree 工件按既有节奏入提交链。
- 文档同步把"仍未收口的部分"段移除或缩小。
