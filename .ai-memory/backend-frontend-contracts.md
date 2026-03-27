# 前后端关键协议

## 登录链路

固定链路：

1. `GET /api/v1/captcha`
2. `POST /api/v1/login`
3. `GET /api/v1/getinfo`
4. `GET /api/v1/menurole`

说明：

- 验证码是一次性消费。
- 登录失败后前端必须刷新验证码与 `uuid`。
- 前端登录页负责展示失败提示并重置验证码状态。

## 用户状态

- `2 = 启用 / 正常 / 可登录`
- `1 = 停用 / 关闭 / 不可登录`

该协议同时影响：

- 用户管理页状态展示与保存
- 登录时用户可用性判断
- 字典 `sys_common_status`

## 多角色协议

### 用户请求体

新增/编辑用户使用：

- `primaryRoleId`
- `roleIds`

### 用户返回体

- `primaryRoleId`
- `roleIds`
- `roles`

### 登录态与用户信息

`getinfo` 返回：

- 主角色：
  - `primaryRoleId`
  - `primaryRoleKey`
  - `primaryRoleName`
- 全角色：
  - `roleIds`
  - `roleKeys`
  - `roleNames`
- 权限结果：
  - `permissions`
  - `buttons`

### 权限规则

- 菜单：按主角色
- 按钮：按全部角色并集
- 接口权限：按全部角色并集

## 菜单组件映射

数据库 `sys_menu.component` 必须满足以下之一：

- 对应 `src/views` 下真实页面路径
- `RouteView`
- `IFrameView`

约束：

- 父级分组不要映射到 `BasicLayout`
- 旧历史路径不再作为正式值继续写入数据库

## 开发工具入口

- `系统接口`：
  - path: `/admin/sys-api`
  - component: `/admin/sys-api/index`
- `表单构建`：
  - path: `/dev-tools/build`
  - component: `/dev-tools/build/index`
- `代码生成`：
  - path: `/dev-tools/gen`
  - component: `/dev-tools/gen/index`
- `代码生成修改`：
  - path: `/dev-tools/editTable`
  - component: `/dev-tools/gen/edit`

## 审计分类

当前统一分类包括：

- `system-settings`
- `generator`
- `role`
- `menu`
- `user`
- `dept`
- `post`
- `api`
- `dict-type`
- `dict-data`
- `job`

动作类型统一为：

- `create`
- `update`
- `delete`
- `status`
- `password`
- `start`
- `stop`
- `run`
