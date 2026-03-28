# Full-stack-TestV1 本地联调交接

更新时间：2026-03-28

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

数据库迁移：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
./go-admin migrate -c config/settings.dev.yml
```

前端开发：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin
pnpm install

cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

当前本地默认管理员账号：

- 用户名：`admin`
- 密码：`123456`

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
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
go build -o ./go-admin .
./go-admin migrate -c config/settings.dev.yml
./go-admin server -c config/settings.dev.yml
```

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1
./scripts/check-local.sh
```

## 后续变更约束

- 新增 migration 后，必须先重新编译 `./go-admin`，再执行迁移。
- 遇到“菜单空白 / 用户新增报缺列”这类问题，先查数据库初始化与 schema，不要先在前端加兜底逻辑。
- 遇到“字典管理链路回退了”这类问题，先查 `1775300000000` 是否执行、旧 SQL 是否被重新导入，不要先改前端目录页。
- 任何会长期影响本地联调的结论，都需要同步写回 `docs/` 或 `.ai-memory/`。
