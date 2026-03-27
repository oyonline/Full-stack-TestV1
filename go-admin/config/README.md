# 配置说明

## 当前仓库约定

- `settings.yml` 只保留可提交模板，不再放真实 secret 或数据库密码。
- 本地开发请使用 `settings.local.yml` 覆盖敏感项；可从 `settings.local.yml.example` 复制。
- 本地覆盖当前支持：
  - `application`
  - `jwt`
  - `database`
  - `gen`
  - `logger.enableddb`

## 环境变量覆盖

当前支持：

- `GO_ADMIN_JWT_SECRET`
- `GO_ADMIN_JWT_TIMEOUT`
- `GO_ADMIN_DB_DRIVER`
- `GO_ADMIN_DB_SOURCE`
- `GO_ADMIN_APP_MODE`
- `GO_ADMIN_APP_HOST`
- `GO_ADMIN_APP_PORT`
- `GO_ADMIN_GEN_FRONTPATH`

## 当前关键约定

- 代码生成前端目标目录：
  - `../vue-vben-admin/apps/web-antd/src`
- 登录链路：
  - `/api/v1/captcha`
  - `/api/v1/login`
  - `/api/v1/getinfo`
  - `/api/v1/menurole`
- 数据库日志是否落库由 `logger.enableddb` 控制。

## 本地使用建议

首次本地开发：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/go-admin
cp config/settings.local.yml.example config/settings.local.yml
```

然后按实际环境填写：

- `settings.jwt.secret`
- `settings.database.source`
- `settings.gen.frontpath`
- `settings.logger.enableddb`
