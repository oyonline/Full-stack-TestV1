# Project Memory - Full-stack-TestV1

**项目路径：** ~/Desktop/Full-stack-TestV1  
**技术栈：** Vue3 + Go + Gin + GORM  
**创建时间：** 2026-03-22  
**状态：** 开发中

---

## 🏗️ 项目记忆架构

基于 WAL 协议的三层记忆 + 专项问题追踪

```
.ai-memory/
├── episodic/          # 时间线记录（做了什么）
│   └── 2026-03-22.md
├── semantic/          # 知识沉淀（是什么）
│   ├── tech-stack.md      # 技术栈要点
│   ├── api-contracts.md   # 接口约定
│   └── dependencies.md    # 依赖版本
├── procedural/        # 操作手册（怎么做）
│   ├── setup-guide.md     # 环境搭建
│   ├── debug-guide.md     # 调试步骤
│   └── deploy-guide.md    # 部署流程
└── issues/            # 问题与解决
    ├── resolved/          # 已解决
    └── pending/           # 待解决
```

---

## 🔄 记录规范（WAL 协议应用）

### 1. 即时记录触发条件

| 场景 | 记录位置 | 示例 |
|------|---------|------|
| 遇到报错/坑 | `issues/pending/` | 启动时报错、接口不通 |
| 解决问题 | `issues/resolved/` + `episodic/` | 如何解决、原因分析 |
| 发现正确路径 | `procedural/` | 正确的配置步骤 |
| 理解技术点 | `semantic/` | 接口约定、框架特性 |
| 调试过程 | `episodic/` | 试了哪些方法 |

### 2. 问题记录模板

```markdown
# Issue: [简要描述]

**时间：** 2026-03-22 20:00  
**现象：** [报错信息/异常表现]  
**环境：** [前端/后端/数据库/网络]  
**尝试过的方法：**
- [ ] 方法1 → 结果
- [ ] 方法2 → 结果  
**最终解决：** [正确的解决方法]  
**原因分析：** [为什么会出现这个问题]  
**预防措施：** [下次如何避免]
```

### 3. 知识沉淀模板

```markdown
# [知识点名称]

**出处：** [文档/代码/调试发现]  
**核心内容：** [简明扼要]  
**相关代码：** [文件路径]  
**注意事项：** [容易忽略的点]  
**更新记录：**
- 2026-03-22: 初始记录
```

---

## 📋 当前已知信息（初始化）

### 技术栈要点
- 前端：Vue 3 + Vite + Ant Design Vue (web-antd)
- 后端：Go + Gin + GORM + JWT + Casbin
- API：/api/v1 前缀
- 开发代理：localhost:10086

### 模块清单
- [x] 登录/认证
- [x] 用户管理
- [x] 角色管理
- [x] 菜单管理
- [x] 部门管理
- [x] 岗位管理
- [x] 字典管理
- [x] 参数配置
- [x] 日志管理
- [x] 接口管理

### 接口约定
- 后端路由：`go-admin/app/admin/router/`
- 前端页面：`vue-vben-admin/apps/web-antd/src/views/admin/`
- 前端 API：`vue-vben-admin/apps/web-antd/src/api/core/`

---

## 🎯 记忆使用流程

1. **遇到问题** → 先查 `issues/resolved/` 看是否已有
2. **解决问题** → 写入 `issues/resolved/` + 原因分析
3. **重复操作** → 整理成 `procedural/` 手册
4. **理解原理** → 沉淀到 `semantic/` 知识库
5. **每日结束** → 更新 `episodic/` 时间线

---

## 🔗 与主记忆系统关联

- **SESSION-STATE.md**：记录当前调试的任务状态
- **MEMORY.md**：记录项目级的重要决策和里程碑
- **memory/YYYY-MM-DD.md**：记录与项目相关的对话
- **本项目 .ai-memory/**：项目专属的详细技术记忆

---

*项目记忆初始化完成*
