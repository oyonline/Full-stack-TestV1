# 排查手册

## 登录失败排查

先查顺序：

1. 用户状态是否为 `2`
2. 密码是否为当前实际设置值
3. 验证码是否为当前图片对应的最新 `uuid`
4. 登录失败后是否已经刷新验证码

关键事实：

- 后端只允许 `status = 2` 的用户登录。
- 验证码是一次性消费。
- 前端登录页失败后会刷新验证码与 `uuid`。

## 菜单 404 / 空页面排查

先查：

1. `sys_menu.path`
2. `sys_menu.component`
3. 前端 `src/views` 是否存在对应页面
4. 父级菜单是否误用了布局组件

经验规则：

- 分组菜单用 `RouteView`
- 承接静态页用 `IFrameView`
- 真实页面必须能映射到 `src/views`

## 点击无响应排查

先查：

1. 目标路由是否已注册到 router
2. 当前 accessMode 下，该路由是否放在正确层级
3. 路由守卫是否允许访问
4. 最后才查组件点击事件和交互

经验规则：

- `backend accessMode` 下，不依赖后端菜单的隐藏前端页必须走静态私有路由
- 不要先在组件事件层连续叠加补丁，再回头查路由注册

## 个人中心问题排查

### 修改密码看起来成功但实际没生效

先查：

1. 前端页面是否真的调用了 `PUT /api/v1/user/pwd/set`
2. 新密码在整条链路里是否被重复 bcrypt 哈希
3. service 层是否显式写入单次哈希结果
4. 成功提示是否放在真实接口成功之后
5. 修改成功后，新密码是否真的能重新登录

经验规则：

- 密码修改不能保留“前端假成功”逻辑
- 密码哈希只能由一层负责，不能在 API 层和 model hook 中重复处理
- 当前项目以 service 层显式生成单次 bcrypt 哈希为准，不再依赖隐式多层加密

### 个人简介不持久化

先查：

1. `sys_user` 是否已有 `introduction` 字段
2. `GET /api/v1/getinfo` 返回的是数据库值还是硬编码文本
3. `PUT /api/v1/user/profile` 是否只更新允许自维护的字段
4. 前端保存后是否重新拉取用户信息

经验规则：

- 个人简介第一版直接挂 `sys_user.introduction`
- 角色、部门、岗位、权限不进入个人中心可编辑范围

## 操作日志为空排查

先查：

1. `settings.local.yml` 中 `logger.enableddb`
2. 后端是否已重启
3. 做一次明确会记审计的操作
4. 直接查库确认 `sys_opera_log`

示例动作：

- 保存参数设置
- 修改角色
- 修改菜单
- 代码生成导入/保存/移除

## 用户管理列表异常排查

如果用户列表报反射错误，先查用户查询 DTO 是否新增了未标注 `search:"-"` 的普通字符串字段。

## 本地验收

提交前最少执行：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1
./scripts/check-local.sh
```

如需冒烟：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```
