# Full-stack-TestV1

全栈脚手架工作区，包含：

- `/Users/linshen/Cursor/Full-stack-TestV1/go-admin`：Go + Gin 后端
- `/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin`：Vue 3 + Vben 前端

## 当前项目状态与推进方式

这份 README 只负责给出当前工作区的统一入口，不单独作为最高真相源。当前建议按下面顺序判断事实：

1. 已提交代码
2. 已提交 migration / `config/db.sql` / 菜单种子
3. 当前运行链路能证明的行为
4. 本地 working tree 未提交改动
5. `.ai-memory/*`
6. README 与说明性文档

当前代码与 git 历史可以直接证明：

- 项目主方向仍是“基于已有后端能力，继续补真实页面和真实链路”，而不是新起一套 mock 页面。
- 平台底座、系统设置、字典管理、表单构建桥接、代码生成承接已经进入可继续推进状态。
- `go-admin/app/platform` 下的 `workflow / module_registry / attachment` 及对应前端最小验收页已经进入已提交历史。
- clone 基线本身已经包含同事融合进来的业务代码，当前仓库不能被描述成“回到纯空脚手架”。
- 本地叠加的后端业务实验层已基本收口，当前 `go-admin` 已回到 clone 基线附近。
- finance 本地扩展不再作为当前主线继续推进；feishu / kingdee / biz_action_log / `sys_user` 扩展类本地后端实验也按同一口径收口。
- 当前 working tree 剩余的主要是前端原型、样式、文档和脚手架增强相关改动，而不是继续叠加新的本地后端业务线。

当前推荐推进方式：

1. 先保护现有业务逻辑和真相源，不把菜单、权限、迁移真相搬回前端或临时 SQL。
2. 先收口文档和真相源口径，再做前端原型补全、页面承接和脚手架增强。
3. 当前不建议重新在本地随意扩写后端业务线；如果后续确实要重开，应先单独对齐真相源边界，而不是顺手叠加到现有 working tree。

## 本地启动

### 后端

当前本地开发以 [settings.dev.yml](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/config/settings.dev.yml) 为准。

如果后端源码或 migration 有变动，先重新编译本地二进制：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
go build -o ./go-admin .
```

数据库初始化：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
./go-admin migrate -c config/settings.dev.yml
```

启动：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
./go-admin server -c config/settings.dev.yml
```

健康检查：

```bash
curl -sS http://127.0.0.1:10082/info
```

### 前端

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin
pnpm install

cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

开发环境默认地址：

- 前端：`http://localhost:5666`
- API：`/api -> http://127.0.0.1:10082`
- 表单构建：`/form-generator -> http://127.0.0.1:10082/form-generator`

### 依赖服务

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`
- 当前本地数据库名：`full_stack_test_v1`

## 登录说明

- 登录链路：`/api/v1/captcha -> /api/v1/login -> /api/v1/getinfo -> /api/v1/menurole`
- 当前后端启用了真实验证码校验，不能再使用 `code=0/uuid=0` 的旧测试方式
- 当前默认管理员账号：`admin / 123456`
- 如果左侧导航缺失，先查数据库初始化是否完整，尤其是 `sys_menu` 是否为空

## 开发工具

### 表单构建

- 当前通过 `/form-generator/` 同域代理承接后端静态 form-generator
- Vue 承接页提供：
  - 下载 schema
  - 复制 schema
  - 导入 schema

### 代码生成

- 当前开放：
  - 数据库表导入
  - 已纳管表配置维护
  - 模板预览
  - 从生成器移除
- 当前隐藏：
  - 生成到项目
  - 生成菜单/API

## 参数设置

- `参数设置` 是唯一正式入口
- 系统名称、Logo、Logo 占位色、登录页标题/说明、全局界面设置都通过后端配置持久化
- 启动时前端会从 `/api/v1/app-config` 拉取运行时配置

## 字典管理

- 当前正式主链路为：
  - `字典类型目录页`
  - `字典类型详情页`
- 当前类型下的字典数据维护统一在类型详情页完成。
- 旧的 `/admin/sys-dict-data` 独立页面已废弃；历史地址当前只保留兼容跳转。

## 质量检查

提交前建议直接执行：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1
./scripts/check-local.sh
```

### 前端类型检查

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm typecheck
```

### 前端本地生产构建

生产构建要求显式设置 `VITE_GLOB_API_URL`。本地可直接使用：

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm build:local
```

### 后端测试

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/go-admin
go test ./...
```

### E2E 冒烟

E2E 默认依赖本地已启动的前后端服务：

- 前端：`http://127.0.0.1:5666`
- 后端：`http://127.0.0.1:10082`

Playwright 会通过开发环境专用的 `X-E2E-Test: true` 验证码调试头获取验证码答案，正常页面和正常接口行为不受影响。

```bash
cd /Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```

可选环境变量：

- `PLAYWRIGHT_BASE_URL`
- `PLAYWRIGHT_API_URL`
- `PLAYWRIGHT_ADMIN_USERNAME`
- `PLAYWRIGHT_ADMIN_PASSWORD`

## 文档

- 当前脚手架与推进规则：[docs/scaffold-guide.md](/Users/linshen/Cursor/Full-stack-TestV1/docs/scaffold-guide.md)
- 本地联调交接：[docs/local-dev-handoff.md](/Users/linshen/Cursor/Full-stack-TestV1/docs/local-dev-handoff.md)
- 头像系统设计：[docs/avatar-system.md](/Users/linshen/Cursor/Full-stack-TestV1/docs/avatar-system.md)
- 项目记忆入口：[.ai-memory/README.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/README.md)
- 项目关键决策：[.ai-memory/project-decisions.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/project-decisions.md)
- 前后端协议：[.ai-memory/backend-frontend-contracts.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/backend-frontend-contracts.md)
- 已知问题：[.ai-memory/known-issues.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/known-issues.md)
- 排查手册：[.ai-memory/runbooks.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/runbooks.md)
- 模块地图：[.ai-memory/module-map.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/module-map.md)
- 后端配置说明：[go-admin/config/README.md](/Users/linshen/Cursor/Full-stack-TestV1/go-admin/config/README.md)
