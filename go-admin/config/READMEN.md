# ⚙ 配置详情

## 当前仓库约定

- `config/settings.yml` 只保留可提交的模板配置，不再放真实 secret 或数据库密码。
- 本地开发请使用 `config/settings.local.yml` 覆盖敏感项；可从 `config/settings.local.yml.example` 复制。
- 也可以使用环境变量覆盖关键配置：
  - `GO_ADMIN_JWT_SECRET`
  - `GO_ADMIN_JWT_TIMEOUT`
  - `GO_ADMIN_DB_DRIVER`
  - `GO_ADMIN_DB_SOURCE`
  - `GO_ADMIN_APP_MODE`
  - `GO_ADMIN_APP_HOST`
  - `GO_ADMIN_APP_PORT`
  - `GO_ADMIN_GEN_FRONTPATH`
- 代码生成前端目标目录为 `../vue-vben-admin/apps/web-antd/src`。
- 登录链路使用 `/api/v1/captcha -> /api/v1/login -> /api/v1/getinfo -> /api/v1/menurole`。

1. 配置文件说明

```yml
settings:
  application:
    # 项目启动环境
    mode: dev # dev开发环境 prod线上环境；
    host: 0.0.0.0 # 主机ip 或者域名，默认0.0.0.0
    # 服务名称
    name: go-admin
    # 服务端口
    port: 8000
    readtimeout: 1
    writertimeout: 2
  log:
    # 日志文件存放路径
    dir: temp/logs
  jwt:
    # JWT加密字符串
    secret: go-admin
    # 过期时间单位：秒
    timeout: 3600
  database:
    # 数据库名称
    name: dbname
    # 数据库类型
    dbtype: mysql
    # 数据库地址
    host: 127.0.0.1
    # 数据库密码
    password: password
    # 数据库端口
    port: 3306
    # 数据库用户名
    username: root
```
