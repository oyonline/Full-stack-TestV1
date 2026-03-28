# 用户头像系统本地方案说明

更新时间：2026-03-28

## 文档定位

- 这份文档描述的是“当前本地 working tree 中的头像方案”，不是已提交主线的完成说明。
- 相关 migration、通用头像组件和部分前端实现目前仍包含本地未提交状态，因此不能把这里的内容直接写成“项目已完成”。
- 如果和已提交代码、已提交 migration 冲突，以已提交真相源为准。

## 当前本地方案包含的内容

当前 working tree 中，头像方案按下面思路组织：

- 用户头像配置继续挂在 `sys_user`
- `GET /api/v1/getinfo`、`PUT /api/v1/user/profile`、`POST /api/v1/user/avatar` 承接头像相关字段与保存
- 前端优先走“姓名优先、登录账号兜底”的字母头像规则
- 个人中心承担头像模式切换、背景色设置和图片上传入口

## 当前不能写成正式已完成的部分

下面这些内容目前只能按“本地状态”描述：

- [1775000000000_user_avatar_profile.go](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/cmd/migrate/migration/version/1775000000000_user_avatar_profile.go) 仍未进入已提交 git 历史
- [user-avatar.ts](/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd/src/utils/user-avatar.ts) 与 [user-avatar.vue](/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd/src/components/user-avatar.vue) 仍属于本地 working tree 内容
- [docs/sql/20260328_sys_user_avatar_profile.sql](/Users/linshen/Cursor/Full-stack-TestV1/docs/sql/20260328_sys_user_avatar_profile.sql) 仅是本地参考 SQL，不是正式 schema 真相源

## 如果继续推进这条线

继续推进头像相关工作前，建议先收口下面三件事：

1. 把正式 migration 放进已提交链路
2. 再把接口契约和 README / `.ai-memory` 口径同步成正式状态
3. 最后再把“已落地”描述写回文档
