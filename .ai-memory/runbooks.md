# 排查手册

说明：

- 这份手册以“当前排查顺序”为主，不单独替代代码真相源。
- 如果某条手册提到的 migration 或文件尚未进入已提交历史，那它只能视为本地联调提示。

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

## 字典管理链路排查

先查：

1. 左侧导航下是否只剩 `字典类型`
2. `sys_menu` 中是否已不存在 `menu_id = 59`、`menu_id = 240`
3. `sys_menu` 中 `241/242/243` 的 `parent_id` 是否已是 `543`
4. `sys_menu_api_rule` 是否已补齐：
   - `543 -> 24`
   - `236 -> 24`
5. `sys_migration` 是否已有 `1775300000000`
6. 旧地址 `/admin/sys-dict-data` 是否已回到 `/admin/sys-dict-type`

经验规则：

- 当前正式产品形态是“字典类型目录页 -> 字典类型详情页”。
- 如果再次看到“全部字典数据”出现在左侧导航，优先怀疑旧菜单种子、旧批处理脚本或本地库状态回滚，而不是先改前端路由。
- 当前类型下的数据读取权限已并入字典类型查看权限，不要再试图恢复 `admin:sysDictData:list/query` 作为正式产品权限。

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

## 左侧导航缺失排查

先查：

1. `sys_menu` 是否为空
2. `sys_user` / `sys_role` / `sys_user_role` 是否至少有 `admin` 这套基础数据
3. `sys_migration` 是否有完整版本记录
4. 重新登录后 `/api/v1/menurole` 是否仍返回空

经验规则：

- 本项目里“首页能进但没有左侧导航”优先怀疑数据库未初始化完整。
- `admin` 角色菜单是后端按管理员特例直接从 `sys_menu` 取，不依赖先补前端假数据。

## 数据库初始化修复

推荐顺序：

1. 备份当前本地库
2. 检查 [settings.dev.yml](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/config/settings.dev.yml) 是否仍指向本地 MySQL
3. 重新编译后端二进制
4. 执行数据库迁移
5. 核对基础表和迁移版本
6. 再启动后端与前端联调

建议命令：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
go build -o ./go-admin .
./go-admin migrate -c config/settings.dev.yml
./go-admin server -c config/settings.dev.yml
```

关键核对项：

- `sys_menu`
- `sys_api`
- `sys_user`
- `sys_role`
- `sys_user_role`
- `sys_migration`

## 新增用户报 `open_id` 缺列排查

先查：

1. `sys_user` 是否存在 `open_id`
2. `sys_user` 是否存在 `job_title`
3. `sys_user` 是否存在 `open_department_id`
4. `sys_user` 是否存在 `open_department_ids`
5. `sys_user` 是否存在 `cn_name`
6. `sys_migration` 是否已有 `1774900000000`

经验规则：

- 业务模型字段多于库表字段时，新增用户和部分飞书同步逻辑会一起出错。
- 当前正式补丁是 migration `1774900000000_sys_user_feishu_fields.go`，不要只在一台机器上手工 `ALTER TABLE` 后就结束。

## 头像设置/上传排查

这一节当前只适用于“本地 working tree 已包含头像相关改动”的场景。

先查：

1. `sys_user` 是否存在 `avatar_type`
2. `sys_user` 是否存在 `avatar_color`
3. 当前本地是否确实带着头像相关 migration 或手工补库状态
4. 当前后端进程是否已重启到最新编译的 `./go-admin`
5. 浏览器 Network 中 `PUT /api/v1/user/profile` 是否返回 200
6. 浏览器 Network 中 `POST /api/v1/user/avatar` 是否返回 200

上传失败时重点看：

- 文件格式是否为 `jpg/png/webp`
- 裁剪后的 blob 是否仍超过 `2MB`
- 返回体里是否包含新的 `avatar` 与 `avatarType`

字母头像显示不对时重点看：

- `GET /api/v1/getinfo` 是否返回 `avatarType` / `avatarColor`
- 当前用户是否仍被旧的图片 URL 覆盖
- 前端是否已经重新拉取 `fetchUserInfo`

## 本地验收

提交前最少执行：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1
./scripts/check-local.sh
```

如需冒烟：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```
