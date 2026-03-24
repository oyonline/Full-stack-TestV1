# 环境搭建指南

## 前提条件

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Git

## 后端启动

### 1. 数据库配置
```bash
# 创建数据库
create database go_admin default charset utf8mb4;

# 导入初始数据
# 在 go-admin/ 目录下找到 sql 文件
```

### 2. 配置文件
```bash
cd go-admin/

# 编辑配置文件
vim config/settings.yml

# 关键配置：
# - 数据库连接
# - JWT密钥
# - 端口（默认10086）
```

### 3. 启动服务
```bash
# 安装依赖
go mod tidy

# 启动
go run cmd/api/server.go

# 或编译后运行
go build -o server cmd/api/server.go
./server
```

**验证：** http://localhost:10086/swagger/index.html

---

## 前端启动

### 1. 安装依赖
```bash
cd vue-vben-admin/
pnpm install
# 或 npm install
```

### 2. 配置代理
```bash
# 编辑 apps/web-antd/.env.development
VITE_PROXY_API=http://localhost:10086
```

### 3. 启动开发服务器
```bash
# 在 vue-vben-admin/ 根目录
pnpm dev

# 或进入子项目
cd apps/web-antd
pnpm dev
```

**访问：** http://localhost:3000

---

## 常见问题

### 后端无法启动
- [ ] 检查数据库连接
- [ ] 检查端口占用
- [ ] 查看日志报错

### 前端无法连接后端
- [ ] 检查代理配置
- [ ] 确认后端端口
- [ ] 检查跨域配置

---

*记录时间：2026-03-22（待完善）*
