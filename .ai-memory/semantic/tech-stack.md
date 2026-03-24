# 技术栈要点

## 前端

### Vue 3 + Vite + Ant Design Vue
- **框架：** Vue 3 Composition API
- **构建工具：** Vite
- **UI库：** Ant Design Vue
- **位置：** `vue-vben-admin/apps/web-antd/`

### 关键配置
- **API代理：** 开发环境代理到 `http://172.16.97.127:10082/api`
- **权限：** 前端权限码固定返回 `['*:*:*']`，实际权限由后端菜单控制
- **路由：** 从后端 `/api/v1/menurole` 动态获取

### 重要字段语义

#### `sys_menu.visible` 字段
| 值 | 含义 | 前端映射 |
|----|------|---------|
| `'0'` | 显示菜单 | `hideInMenu: false` |
| `'1'` | 隐藏菜单 | `hideInMenu: true` |

**注意**：前端代码中使用 `hideInMenu: node.visible !== '0'` 进行映射。

## 后端

### Go + Gin + GORM
- **框架：** Gin
- **ORM：** GORM
- **认证：** JWT
- **权限：** Casbin
- **位置：** `go-admin/`

### 项目结构
```
go-admin/
├── app/admin/router/      # 路由注册
├── app/admin/apis/        # API处理
├── app/admin/service/     # 业务逻辑
├── app/admin/models/      # 数据模型
└── cmd/api/server.go      # 入口
```

### API规范
- **前缀：** `/api/v1`
- **CRUD：** 标准 RESTful
- **认证：** JWT Token

### 响应格式
```json
{
  "code": 200,
  "data": [...],
  "requestId": "xxx",
  "msg": "success"
}
```

## 接口对接

### 关键接口
| 功能 | 后端路径 | 前端API |
|------|---------|---------|
| 登录 | POST /v1/login | loginApi |
| 用户信息 | GET /v1/getinfo | getUserInfoApi |
| 菜单 | GET /v1/menurole | getAllMenusApi |
| 用户CRUD | /v1/sys-user | userApi |

## 注意事项
- 前端权限码固定，真实权限由后端菜单接口控制
- 菜单路由需要后端配置 component 路径
- API代理配置在 vite.config.ts
- **重要**：后端 `visible` 字段 `'0'` 表示显示，`'1'` 表示隐藏

---

*更新时间：2026-03-23*
