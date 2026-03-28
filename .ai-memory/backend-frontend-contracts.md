# 前后端关键协议

说明：

- 这份文档记录的是“当前约定”，不是替代代码的最高真相源。
- 如果和已提交代码、已提交 migration 冲突，以已提交真相源为准。
- 本地 working tree 未提交改动只能标记为“本地状态”，不能直接写成正式协议。

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
- 接口权限：当前按主角色 key 走 Casbin 校验

当前代码已确认的差异：

- `/api/v1/getinfo` 返回给前端的 `permissions / buttons` 是多角色并集结果
- 后端 `AuthCheckRole` 实际调用 `Enforce(rolekey, path, method)`，这里的 `rolekey` 是主角色 key

## 菜单组件映射

数据库 `sys_menu.component` 必须满足以下之一：

- 对应 `src/views` 下真实页面路径
- `RouteView`
- `IFrameView`

约束：

- 父级分组不要映射到 `BasicLayout`
- 旧历史路径不再作为正式值继续写入数据库

## 字典管理路由与权限契约

- 正式主入口：
  - `/admin/sys-dict-type`
- 正式详情入口：
  - `/admin/sys-dict-type/detail?dictId=<id>`
  - 这是前端静态隐藏路由，不依赖后端菜单单独下发。
- 历史路径：
  - `/admin/sys-dict-data` 已废弃为独立页面。
  - 当前仅保留隐藏 redirect 到 `/admin/sys-dict-type`，用于兼容旧书签和旧标签页。

数据库菜单契约：

- `menu_id = 543` 是正式的“字典类型”页面节点。
- `menu_id = 59` 的“字典数据”页面节点已废弃，不应再作为正式种子或迁移目标保留。
- `menu_id = 240` 的“查询数据”按钮节点已废弃。
- `menu_id = 241/242/243` 当前应挂在 `543` 下面，而不是旧的 `59` 下面。

接口权限契约：

- `GET /api/v1/dict/data` 当前读取权限并入字典类型查看权限。
- 正式 API 绑定应至少包含：
  - `sys_menu_api_rule(543, 24)`
  - `sys_menu_api_rule(236, 24)`
- `admin:sysDictData:add/edit/remove` 继续保留，供类型详情页使用。

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

## 标准列表列配置

当前后台路由级标准列表页统一支持列个性化配置，不新增后端接口。

范围约定：

- 只接入标准 `ant-design-vue Table` 路由页。
- 不覆盖编辑态子表格、代码生成字段表、`vxe-table`、组件内嵌表格。

接入约定：

- 每一列必须提供稳定 `key`；没有显式 `key` 时，至少保证 `dataIndex` 稳定。
- 每个表格必须提供稳定 `tableId`。
- 同一路由下如果存在多张表，必须使用不同 `tableId` 做隔离。

持久化约定：

- 配置范围是“当前用户 + 当前路由 + 当前表格”。
- 当前实现存浏览器本地 `localStorage`，不做跨设备同步。
- 存储键口径为：`admin-table-columns:${userId}:${route.path}:${tableId}`。

列行为约定：

- 支持显示/隐藏、拖拽排序、列宽拖拽、固定左侧/右侧。
- 不允许把普通业务列全部隐藏到只剩系统列。
- 未声明宽度的列，默认宽度为 `160`。
- 列宽范围统一限制为 `80` 到 `600`。
- 表格 `scroll.x` 由可见列宽自动汇总计算，不再手写固定值。

系统列约定：

- 当前 `action` 作为系统列接入。
- 系统列始终显示、始终固定右侧、始终位于最右。
- 系统列允许调整列宽，但不允许隐藏、取消固定或参与拖拽排序。

## 头像契约（当前本地 working tree 方案）

这部分当前只能按“本地状态”理解，原因是相关 migration 和部分前端基础组件尚未进入已提交历史。

当前用户相关接口补充以下字段：

- `avatar`
- `avatarType`
- `avatarColor`

### GET `/api/v1/getinfo`

返回约定：

- `avatar`: 用户图片头像 URL，可为空
- `avatarType`: `image` / `letter` / 空字符串
- `avatarColor`: 用户自定义背景色，HEX 格式，可为空

前端消费规则：

1. `avatarType = image` 且 `avatar` 非空，展示图片头像
2. `avatarType = letter`，展示字母头像
3. `avatarType` 为空但 `avatar` 非空，按历史数据兼容展示图片头像
4. 其余情况展示字母头像，并按“姓名优先、登录账号兜底”的显示种子稳定映射颜色

### PUT `/api/v1/user/profile`

请求体约定：

- `introduction: string`
- `avatarType: "image" | "letter"`
- `avatarColor: string`

当前首版使用说明：

- 个人简介和头像模式/背景色走同一个保存接口。
- 图片路径本身不走这个接口更新。

### POST `/api/v1/user/avatar`

请求约定：

- `multipart/form-data`
- 字段名优先 `file`
- 兼容旧前端字段 `upload[]`

校验约定：

- 仅支持 `jpg` / `jpeg` / `png` / `webp`
- 文件大小不超过 `2MB`

成功响应：

- `avatar`
- `avatarType`

补充说明：

- 在正式 migration 进入已提交链前，这组字段不能被当作“所有环境默认可用”的初始化保证。
