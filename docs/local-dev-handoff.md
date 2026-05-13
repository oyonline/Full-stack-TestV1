# Full-stack-TestV1 本地联调交接

更新时间：2026-05-13

## 必读（先看这一段）

1. **迁移必须与二进制同步**：新增或修改 `go-admin/cmd/migrate/migration/version/*.go` 后，**禁止**只运行旧的 `./go-admin migrate`（二进制里的迁移列表仍是旧的）。请使用下面 **`make migrate-dev`**（本地联调）或 **`make migrate`**（使用 `config/settings.yml` 时），二者都会 **先 `go build` 再 migrate**。详见下文「本地启动迁移规范」。
2. **管理员密码以数据库为准**：文档里的 `admin` / `123456` 仅对齐 `config/db.sql` 种子；若曾改过密码或导入过别的库，请以 `sys_user` 实际哈希为准。
3. **仓库路径**：下文命令以本 workspace 根目录为准（`/Users/linshen/Documents/Full-stack-TestV1` 仅作示例时可替换为你的克隆路径）。

## 文档定位

- 这份文档只记录“当前这台机器上的本地联调基线”和“当前 working tree 里能观察到的本地状态”。
- 这不是仓库主线完成度说明；如果和已提交代码、已提交 migration 或 `config/db.sql` 冲突，以已提交真相源为准。
- 任何本地未提交改动，都只能写成“本地状态”，不能写成“项目已完成”。

## 当前本地联调基线

- 后端本地配置： [settings.dev.yml](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/config/settings.dev.yml)
- 本地 MySQL：`127.0.0.1:3306`
- 当前目标库：`full_stack_test_v1`

后端启动：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
./go-admin server -c config/settings.dev.yml
```

数据库迁移（**推荐**，强制重新编译后再 migrate，避免漏跑 migration）：

```bash
cd /path/to/Full-stack-TestV1/go-admin
make migrate-dev
# 或：编译 + 迁移
make build-and-migrate-dev
```

不推荐单独执行 `./go-admin migrate`（除非刚执行过 `go build -o ./go-admin .`）。等价手动方式：

```bash
go build -o ./go-admin . && ./go-admin migrate -c config/settings.dev.yml
```

前端开发：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin
pnpm install

cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

当前本地默认管理员账号（与 **种子/SQL 基线** 一致时成立）：

- 用户名：`admin`
- 密码：`123456`

### 登录失败（incorrect Username or Password）排查简表

后端 JWT 层会把多种失败合并成同一句英文提示，请逐项排除：

| 检查项 | 说明 |
|--------|------|
| `sys_user.status` | 必须为 **`2`**（启用）；`1` 不可登录。 |
| 验证码 | `uuid` + `code` 与页面一致；验证码 **一次性**，失败后重新获取。 |
| 密码 | 与库里 `password` 哈希匹配；勿假设 README 与本地库一致。 |
| 用户名 | 区分大小写；与库里 `username` 完全一致。 |

更多协议见 [`.ai-memory/backend-frontend-contracts.md`](../.ai-memory/backend-frontend-contracts.md)。

### DSN 日志脱敏

后端初始化数据库时，日志中的连接串已 **隐去密码段**（仅打印 `user:***@tcp(...)`）。切勿在其他自定义日志里打印完整 `database.source`。

## 已提交主线可直接依赖的事项

- 用户扩展字段缺失的正式补丁是 [1774900000000_sys_user_feishu_fields.go](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/cmd/migrate/migration/version/1774900000000_sys_user_feishu_fields.go)。
- 字典管理正式收口版本是 [1775300000000_remove_dict_data_page.go](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/cmd/migrate/migration/version/1775300000000_remove_dict_data_page.go)。
- 平台模块的 `workflow / module_registry / attachment` 及对应最小前端验收页已经在已提交历史中。

## 当前本地未提交状态

当前以仓库实查结果为准：

- `go-admin` 当前没有本地业务代码差异；本地叠加的后端业务实验层已基本收口，回到 clone 基线附近。
- clone 基线本身已经包含同事融合进来的业务代码，因此当前仓库不能写成“业务代码已删掉”或“回到纯空脚手架”。
- 当前 working tree 剩余内容主要集中在：
  - 前端原型与页面交互补全
  - 样式与体验收口
  - 文档同步
  - 脚手架增强
- `docs/sql/*` 仅作为手工排查或本地修库参考，不是正式 schema 真相源。

## finance 协同边界

- finance 本地扩展已清理，不再作为当前主线继续推进。
- feishu / kingdee / biz_action_log / `sys_user` 扩展类本地后端实验也按同一口径收口，不再视为当前本地后端主线。
- 当前后端业务真相源默认以 clone 基线附近为准；后续联调优先围绕已提交后端基线做前端补全。
- 当前协同默认规则是：
  - 不把本地后端业务实验层重新带回主线
  - 不把历史本地实验状态写成当前正式完成度
  - 不在 finance 或其他业务后端线上顺手继续扩写

## 当前更适合的前端承接方向

- 当前更适合继续推进的方向是：
  - 前端原型补全
  - 页面与交互完善
  - 样式与体验收口
  - 脚手架增强
  - 文档同步
- 这条主线默认围绕已提交后端基线做真实页面承接，不再把 finance / feishu / kingdee / biz_action_log / `sys_user` 扩展当作当前本地后端推进目标。

## 本地核验建议

先区分“问题出在已提交主线”还是“问题出在本地未提交状态”。

推荐顺序：

1. 重新编译 `./go-admin`
2. 执行数据库迁移
3. 启动后端和前端
4. 执行本地检查
5. 再做浏览器最小冒烟

建议命令：

```bash
cd /path/to/Full-stack-TestV1/go-admin
make build-and-migrate-dev
./go-admin server -c config/settings.dev.yml
```

```bash
cd /path/to/Full-stack-TestV1
./scripts/check-local.sh
```

## 本地启动迁移规范

**本地联调（`settings.dev.yml`）**：优先 **`make migrate-dev`** / **`make build-and-migrate-dev`**（强制 `go build` 后再 migrate）。

**使用 `config/settings.yml` 的环境**：走 **`make migrate`** / **`make build-and-migrate`**。

**禁止**：在未重新编译的前提下单独运行 `./go-admin migrate ...`，否则会使用磁盘上的旧二进制，**漏注册新 migration**（详见 [`.ai-memory/known-issues.md`](../.ai-memory/known-issues.md)）。

```bash
cd /path/to/Full-stack-TestV1/go-admin

make migrate-dev
make build-and-migrate-dev

make migrate
make build-and-migrate
```

## 数据权限（phase2 已启用）

- `settings.application.enabledp` 自 C7-7 起在 `settings.yml` / `settings.full.yml` / `settings.sqlite.yml` / `settings.local.yml.example` 全部置为 `true`，`settings.demo.yml` 一直为 `true`。
- 本地启动后，默认 `admin` 用户挂的角色 `dataScope=1`（全部数据），观察到的公告 / 业务列表行为与 phase1 完全一致。
- 想验证数据范围真实生效，新建一个 `dataScope=5`（仅本人）的测试角色，分配给一个非 admin 测试用户登录，应仅能看到该用户 `create_by` 的行（C7-5 已端到端覆盖，本地 smoke 一次即可）。
- 启动前必须先跑迁移：`make build-and-migrate-dev`（或 `make build-and-migrate`），`1778200000000_data_permission_default` 会兜底已有 `sys_role.data_scope` 空值。
- 接入业务模块时遵循 `PROJECT_CONVENTIONS.md` 的数据权限规约，不要在新模块里再写 raw SQL 绕过 `dataScope`。

## 后续变更约束

- 新增 migration 后，必须先重新编译 `./go-admin`，再执行迁移。
- 遇到“菜单空白 / 用户新增报缺列”这类问题，先查数据库初始化与 schema，不要先在前端加兜底逻辑。
- 遇到“字典管理链路回退了”这类问题，先查 `1775300000000` 是否执行、旧 SQL 是否被重新导入，不要先改前端目录页。
- 任何会长期影响本地联调的结论，都需要同步写回 `docs/` 或 `.ai-memory/`。
