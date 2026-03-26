# Full-stack-TestV1

全栈脚手架工作区，包含：

- `/Users/linshen/Desktop/Full-stack-TestV1/go-admin`：Go + Gin 后端
- `/Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin`：Vue 3 + Vben 前端

## 本地启动

### 后端

先准备本地覆盖配置：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/go-admin
cp config/settings.local.yml.example config/settings.local.yml
```

按本机实际值填写：

- `settings.jwt.secret`
- `settings.database.source`
- `settings.gen.frontpath`

启动：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/go-admin
go run . server -c config/settings.yml
```

健康检查：

```bash
curl -sS http://localhost:10082/info
```

### 前端

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin
pnpm install

cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm dev
```

开发环境默认地址：

- 前端：`http://localhost:5666`
- API：`/api -> http://127.0.0.1:10082`
- 表单构建：`/form-generator -> http://127.0.0.1:10082/form-generator`

### 依赖服务

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`

## 登录说明

- 登录链路：`/api/v1/captcha -> /api/v1/login -> /api/v1/getinfo -> /api/v1/menurole`
- 当前后端启用了真实验证码校验，不能再使用 `code=0/uuid=0` 的旧测试方式

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

## 质量检查

提交前建议直接执行：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1
./scripts/check-local.sh
```

### 前端类型检查

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm typecheck
```

### 前端本地生产构建

生产构建要求显式设置 `VITE_GLOB_API_URL`。本地可直接使用：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm build:local
```

### 后端测试

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/go-admin
go test ./...
```

### E2E 冒烟

E2E 默认依赖本地已启动的前后端服务：

- 前端：`http://127.0.0.1:5666`
- 后端：`http://127.0.0.1:10082`

Playwright 会通过开发环境专用的 `X-E2E-Test: true` 验证码调试头获取验证码答案，正常页面和正常接口行为不受影响。

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```

可选环境变量：

- `PLAYWRIGHT_BASE_URL`
- `PLAYWRIGHT_API_URL`
- `PLAYWRIGHT_ADMIN_USERNAME`
- `PLAYWRIGHT_ADMIN_PASSWORD`

## 文档

- 脚手架使用说明：[docs/scaffold-guide.md](/Users/linshen/Desktop/Full-stack-TestV1/docs/scaffold-guide.md)
