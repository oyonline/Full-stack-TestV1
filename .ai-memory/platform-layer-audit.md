# 平台底座层 / 平台能力层审计

## 目的

这份文档用于回答两个问题：

1. 当前项目哪些内容已经可以视为“统一平台底座”
2. 当前项目哪些内容已经具备“跨业务复用的平台能力雏形”

本结论用于指导后续业务模块接入，尤其是：

1. 财务预算管控模块
2. 项目管理模块
3. 产品中心
4. 后续供应链、销售等模块

## 当前总体判断

当前项目已经具备“统一底座 + 多业务模块接入”的雏形，但两层成熟度不均衡：

- 平台底座层：相对扎实
- 平台能力层：已有雏形，但明显偏弱

换句话说：

- 现在已经不适合继续把项目当成单一后台来修
- 但也还不到可以直接无负担接入复杂业务模块的程度

## 平台底座层（已形成）

### 1. 登录认证与登录态

- 后端：
  - `common/middleware/handler/login.go`
  - `common/middleware/auth.go`
- 前端：
  - `src/api/core/auth.ts`
  - `src/views/_core/authentication/login.vue`
- 能力：
  - 验证码
  - 登录
  - JWT
  - `getinfo`
  - `menurole`

### 2. 用户 / 角色 / 部门 / 岗位

- 后端：
  - `app/admin/apis/sys_user.go`
  - `app/admin/apis/sys_role.go`
  - `app/admin/apis/sys_dept.go`
  - `app/admin/apis/sys_post.go`
- 前端：
  - `src/views/admin/sys-user`
  - `src/views/admin/sys-role`
  - `src/views/admin/sys-dept`
  - `src/views/admin/sys-post`

### 3. 权限体系

- 菜单按主角色生效
- 按钮权限和接口权限按多角色并集生效
- 数据权限当前关闭

关键位置：

- `app/admin/apis/sys_menu.go`
- `app/admin/apis/sys_user.go`
- `src/api/core/menu.ts`
- `src/composables/use-admin-permission.ts`

### 4. 菜单与导航

- `sys_menu` 已是正式真相源
- 前端动态路由已大幅减少历史兼容
- 父级分组统一使用 `RouteView`

关键位置：

- `src/router/access.ts`
- `app/admin/apis/sys_menu.go`

### 5. 系统设置与运行时配置

- `参数设置` 已成为唯一正式入口
- 系统名称、Logo、占位色、登录页标题/说明、全局界面设置都已接通

关键位置：

- `app/admin/service/sys_config.go`
- `app/admin/service/sys_config_settings.go`
- `src/views/admin/sys-config/set.vue`
- `src/utils/system-settings.ts`

### 6. 字典与参数

- 字典能力已形成统一管理入口
- 系统参数能力已形成统一设置入口

关键位置：

- `app/admin/service/dto/sys_dict_type.go`
- `app/admin/service/dto/sys_dict_data.go`
- `app/admin/service/dto/sys_config.go`

### 7. 审计与日志

- `sys_login_log`
- `sys_opera_log`
- 审计分类、动作类型、摘要规范已形成

关键位置：

- `common/audit/audit.go`
- `common/middleware/logger.go`
- `src/views/admin/sys-login-log`
- `src/views/admin/sys-opera-log`

### 8. 前后端统一协议与后台标准页母版

- 请求层、响应结构、后台列表页/树页/详情抽屉/权限按钮/轻量表单 schema 已有统一支撑层

关键位置：

- `src/api/request.ts`
- `src/components/admin/*`
- `src/composables/use-admin-table.ts`
- `src/composables/use-admin-tree-list.ts`

## 平台能力层（雏形）

### 1. 代码生成

- 当前已支持：
  - 导入数据库表
  - 维护生成配置
  - 独立编辑页
  - 模板预览
  - 从生成器移除

关键位置：

- `app/other/apis/tools/gen.go`
- `app/other/apis/tools/sys_tables.go`
- `src/views/dev-tools/gen/index.vue`
- `src/views/dev-tools/gen/edit.vue`

当前判断：

- 已具备能力层雏形
- 但还不是“业务模块接入框架”

### 2. 表单构建

- 当前通过后端静态页 + Vue bridge 承接
- 已支持导入/导出 schema

关键位置：

- `src/views/dev-tools/build/index.vue`

当前判断：

- 已具备能力层雏形
- 但还不是稳定的原生平台能力

### 3. 上传能力

- 当前已有统一上传接口和存储抽象

关键位置：

- `app/other/apis/file.go`
- `common/file_store/*`

当前判断：

- 具备演进成“附件中心”的基础
- 但目前仍只是上传能力，不是附件平台

### 4. 定时任务

- 当前已具备统一任务调度能力

关键位置：

- `app/jobs/apis/sys_job.go`
- `src/views/admin/sys-job`

当前判断：

- 可为后续通知、周期处理、提醒任务提供基础
- 但还不能直接视为消息/待办平台

## 当前未发现的正式平台能力

以下内容当前未发现正式模块：

- 审批流
- 待办中心
- 我发起的
- 站内消息中心
- 公告能力
- 通用单据能力
- 通用状态流转引擎
- 附件中心

## 当前边界问题

### 1. 底座强、能力弱

最明显的问题不是底座没有，而是平台能力层还没真正立起来。

### 2. 代码生成与标准页母版没有完全对齐

当前代码生成偏 CRUD，后台标准页母版偏平台化后台，两者还没合并成“业务模块生产力工具”。

### 3. 上传能力和附件中心之间有明显断层

文件能传，但业务附件还无统一抽象。

### 4. 审计偏系统管理动作

统一审计已经形成，但“业务操作日志”还没有独立的平台规范。

### 5. 审批流完全缺位

如果现在直接上财务预算模块，很容易把审批逻辑硬塞进业务模块，后续再平台化代价会很高。

## 面向财务预算管控模块的优先补齐建议

### P0

- 模块注册 / 模块入口配置
- 审批流 MVP
- 我的待办 / 我发起的
- 附件中心
- 统一业务操作日志

### P1

- 通用单据状态机
- 模块级权限约定
- 站内消息 / 公告

### P2

- 数据权限基础框架

## 当前最合理的下一步

先补 1 到 2 个平台能力，再开始接财务预算模块。

最推荐顺序：

1. 审批流 MVP + 模块注册 / 模块入口配置
2. 我的待办 / 我发起的 + 统一业务操作日志
3. 附件中心最小版

然后再让财务预算模块成为第一个规范接入的平台业务模块。

第一期最小闭环详见：[platform-capability-phase1.md](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/platform-capability-phase1.md)
