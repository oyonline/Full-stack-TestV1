# Full-stack-TestV1 - 企业级前后端分离管理后台

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

企业级前后端分离管理后台系统，基于 **Go + Gin + GORM** 后端与 **Vue 3 + Vite + Ant Design Vue** 前端构建，提供完整的 RBAC 权限管理、动态菜单、系统监控等功能。

---

## 📁 项目结构

```
Full-stack-TestV1/
├── go-admin/                 # 后端服务 (Go)
│   ├── app/
│   │   ├── admin/           # 管理后台模块
│   │   │   ├── apis/        # API 处理器 (Controller)
│   │   │   ├── models/      # 数据模型 (Model)
│   │   │   ├── router/      # 路由配置
│   │   │   └── service/     # 业务逻辑层
│   │   ├── jobs/            # 定时任务模块
│   │   └── other/           # 其他模块(文件、监控等)
│   ├── cmd/                 # 命令行入口
│   ├── common/              # 公共组件
│   │   ├── middleware/      # 中间件
│   │   ├── models/          # 通用模型
│   │   └── service/         # 通用服务
│   └── config/              # 配置文件
│
├── vue-vben-admin/          # 前端项目 (Vue 3)
│   ├── apps/
│   │   └── web-antd/        # Ant Design Vue 版本主应用
│   │       ├── src/
│   │       │   ├── api/     # API 接口封装
│   │       │   ├── views/   # 页面视图
│   │       │   ├── router/  # 路由配置
│   │       │   └── store/   # 状态管理
│   │       └── package.json
│   └── packages/            # 共享组件包
│       ├── @core/           # 核心框架
│       ├── effects/         # 副作用处理
│       ├── stores/          # Pinia 状态库
│       └── types/           # TypeScript 类型
│
└── docs/                     # 项目文档
    ├── project-prd.md        # 需求文档
    ├── project-implementation-plan.md  # 实施计划
    └── *.md                  # 其他技术文档
```

---

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        前端层 (Frontend)                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Vue 3 + Vite + Ant Design Vue + Pinia + Vue Router     │   │
│  │  ─────────────────────────────────────────────────────  │   │
│  │  • 动态路由与菜单生成                                     │   │
│  │  • 组件级权限控制 (v-access)                              │   │
│  │  • 请求拦截与响应处理                                     │   │
│  │  • 状态管理与本地缓存                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │ HTTP/RESTful API
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        后端层 (Backend)                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Go + Gin + GORM + JWT + Casbin                         │   │
│  │  ─────────────────────────────────────────────────────  │   │
│  │  • 认证授权 (JWT + Casbin RBAC)                          │   │
│  │  • 业务逻辑处理                                          │   │
│  │  • 数据访问层 (GORM)                                     │   │
│  │  • 接口文档 (Swagger)                                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        数据层 (Data)                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   MySQL     │  │  PostgreSQL │  │        SQLite           │ │
│  │  (主数据库)  │  │  (可选)     │  │    (开发/测试)           │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 技术栈

### 后端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.24+ | 编程语言 |
| Gin | v1.10.0 | Web 框架 |
| GORM | v1.25.12 | ORM 框架 |
| JWT | v5.2.2 | 身份认证 |
| Casbin | v2.104.0 | RBAC 权限控制 |
| Cobra | v1.9.1 | CLI 命令框架 |
| Swagger | v1.16.4 | API 文档 |
| Zap | v1.27.0 | 日志框架 |

### 前端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.x | 前端框架 |
| Vite | 5.x | 构建工具 |
| Ant Design Vue | 4.x | UI 组件库 |
| Pinia | 2.x | 状态管理 |
| Vue Router | 4.x | 路由管理 |
| TypeScript | 5.x | 类型系统 |
| Axios | - | HTTP 客户端 |

---

## 📦 核心模块

### 1. 系统管理模块

| 模块 | 功能描述 | 状态 |
|------|----------|------|
| **用户管理** | 用户增删改查、状态管理、密码修改、头像上传 | ✅ 已完成 |
| **角色管理** | 角色增删改查、菜单权限分配、数据权限配置 | ✅ 已完成 |
| **菜单管理** | 菜单树配置、路由生成、接口关联、图标配置 | ✅ 已完成 |
| **部门管理** | 部门树维护、层级关系、数据权限范围 | ✅ 已完成 |
| **岗位管理** | 岗位增删改查、状态管理 | ✅ 已完成 |

### 2. 系统监控模块

| 模块 | 功能描述 | 状态 |
|------|----------|------|
| **登录日志** | 登录记录查询、异常登录监控、日志清理 | ✅ 已完成 |
| **操作日志** | 操作记录审计、详情查看、日志清理 | ✅ 已完成 |
| **服务监控** | 服务器 CPU/内存/磁盘监控 | ✅ 已完成 |
| **定时任务** | 任务调度管理、执行日志 | ✅ 已完成 |

### 3. 系统工具模块

| 模块 | 功能描述 | 状态 |
|------|----------|------|
| **字典管理** | 字典类型/数据维护、下拉选项 | ✅ 已完成 |
| **参数管理** | 系统参数配置、动态配置 | ✅ 已完成 |
| **接口管理** | 接口文档管理、权限标识配置 | ✅ 已完成 |
| **代码生成** | 根据数据表生成前后端代码 | 🔄 部分完成 |

---

## 🔐 权限系统

### RBAC 权限模型

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    用户      │────▶│    角色      │────▶│    权限      │
│  (SysUser)  │ N:M │  (SysRole)  │ N:M │  (SysMenu)  │
└─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   接口(API)  │
                                        │  (SysApi)   │
                                        └─────────────┘
```

### 权限控制层级

1. **路由级权限**: 通过 `menurole` 接口返回的菜单控制页面访问
2. **按钮级权限**: 通过 `v-access` 指令控制按钮显示
3. **接口级权限**: 通过 Casbin 控制 API 访问
4. **数据级权限**: 通过部门层级控制数据可见范围

---

## 🔌 API 接口规范

### 接口前缀

```
Base URL: /api/v1
```

### 核心接口列表

| 接口 | 方法 | 描述 |
|------|------|------|
| `/login` | POST | 用户登录 |
| `/logout` | POST | 用户登出 |
| `/refresh_token` | GET | 刷新 Token |
| `/getinfo` | GET | 获取当前用户信息 |
| `/menurole` | GET | 获取角色菜单 |
| `/sys-user` | CRUD | 用户管理 |
| `/role` | CRUD | 角色管理 |
| `/menu` | CRUD | 菜单管理 |
| `/dept` | CRUD | 部门管理 |
| `/post` | CRUD | 岗位管理 |
| `/dict/type` | CRUD | 字典类型 |
| `/dict/data` | CRUD | 字典数据 |
| `/config` | CRUD | 参数配置 |
| `/sys-api` | GET/PUT | 接口管理 |
| `/sys-login-log` | GET/DELETE | 登录日志 |
| `/sys-opera-log` | GET/DELETE | 操作日志 |

---

## 🚀 快速开始

### 环境要求

- **Go**: 1.24+
- **Node.js**: 18+
- **pnpm**: 8+
- **数据库**: MySQL 8.0+ / PostgreSQL / SQLite

### 后端启动

```bash
# 进入后端目录
cd go-admin

# 安装依赖
go mod tidy

# 初始化数据库（首次）
go run main.go migrate -c config/settings.dev.yml

# 启动服务
go run main.go server -c config/settings.dev.yml
# 或编译后运行
go build -o go-admin
./go-admin server -c config/settings.yml
```

后端服务默认运行在 `http://localhost:10086`

### 前端启动

```bash
# 进入前端目录
cd vue-vben-admin

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 或只启动 web-antd 应用
cd apps/web-antd
pnpm dev
```

前端服务默认运行在 `http://localhost:3000`

---

## ⚙️ 配置文件

### 后端配置 (go-admin/config/settings.yml)

```yaml
settings:
  application:
    mode: dev
    port: 10086
  database:
    driver: mysql
    source: root:password@tcp(127.0.0.1:3306)/go-admin?charset=utf8mb4
  jwt:
    secret: your-secret-key
    timeout: 3600
  casbin:
    model-path: ./config/rbac_model.conf
```

### 前端代理配置

开发环境代理已配置在 `vite.config.ts` 中，将 `/api` 代理到后端服务。

---

## 📚 文档索引

| 文档 | 说明 |
|------|------|
| [docs/project-prd.md](docs/project-prd.md) | 项目需求文档 (PRD) |
| [docs/project-implementation-plan.md](docs/project-implementation-plan.md) | 实施计划 |
| [docs/project-batch-delivery-plan.md](docs/project-batch-delivery-plan.md) | 分批交付计划 |
| [docs/dict-type-integration-guide.md](docs/dict-type-integration-guide.md) | 字典集成指南 |
| [go-admin/README.Zh-cn.md](go-admin/README.Zh-cn.md) | 后端详细文档 |
| [vue-vben-admin/README.md](vue-vben-admin/README.md) | 前端详细文档 |

---

## 🤝 前后端协作规范

### 接口对接流程

```
1. 后端定义 API 接口 (Swagger)
      │
      ▼
2. 前端封装 API 模块 (api/core/*.ts)
      │
      ▼
3. 前端开发页面组件 (views/admin/*.vue)
      │
      ▼
4. 联调测试验证
      │
      ▼
5. 完善字段映射与异常处理
```

### 数据流规范

```
前端页面 ──▶ API 封装 ──▶ 请求拦截 ──▶ 后端 API
    │                                    │
    │                                    ▼
    │                              认证/权限校验
    │                                    │
    │                                    ▼
    │                              业务逻辑处理
    │                                    │
    │                                    ▼
    └──────────────────────────◀ 响应数据
         响应拦截 ──▶ 数据处理 ──▶ 页面渲染
```

---

## 📝 开发规范

### 后端规范

- 遵循 RESTful API 设计规范
- 使用 DTO 进行数据传输
- 接口统一返回格式: `{ code: 200, data: {}, msg: "" }`
- 错误码规范: 200 成功，其他为业务错误码

### 前端规范

- 组件命名: PascalCase
- API 函数命名: camelCase，以 Api 结尾
- 类型定义: PascalCase，以 Type 结尾
- 页面文件: 放在 `views/admin/` 目录下

---

## 🔍 调试技巧

### 后端调试

```bash
# 带日志级别启动
./go-admin server -c config/settings.yml --log-level debug

# 自动同步 API 到数据库
./go-admin server -c config/settings.yml -a true
```

### 前端调试

```bash
# 开启详细日志
pnpm dev --debug

# 构建分析
pnpm build:analyze
```

---
