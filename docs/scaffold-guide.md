# Full-stack-TestV1 脚手架使用说明

这份文档只保留“当前脚手架怎么用”的信息。更细的长期约定请同时查看：

- [项目记忆入口](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/README.md)
- [项目关键决策](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/project-decisions.md)
- [前后端协议](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/backend-frontend-contracts.md)
- [排查手册](/Users/linshen/Desktop/Full-stack-TestV1/.ai-memory/runbooks.md)

## 1. 系统设置

- `参数设置` 是唯一正式系统设置入口。
- 系统名称、Logo、Logo 占位色、登录页标题/说明、全局界面设置都会持久化到后端。
- 页面启动时前端会从 `/api/v1/app-config` 拉取运行时配置。

## 2. 权限模型

- 用户支持多角色。
- 菜单按主角色生效。
- 按钮权限和接口权限按所有角色并集生效。
- 本期不做角色切换，也不做数据权限。

## 3. 开发工具

### 系统接口

- 使用当前后台原生系统接口页承接。

### 代码生成

当前已支持：

- 数据库表导入
- 已纳管表配置维护
- 独立编辑页修改表级/字段级配置
- 模板预览
- 从生成器移除

当前仍然隐藏：

- 生成到项目
- 生成菜单/API

推荐工作流：

1. 在 `开发工具 -> 代码生成` 搜索数据库表
2. 执行“导入到生成器”
3. 切到“生成配置”
4. 进入“编辑配置”
5. 保存后再做“模板预览”

### 表单构建

- 前端通过同域 `/form-generator/` 承接后端静态 form-generator。
- 当前桥接支持：
  - 下载 schema
  - 复制 schema
  - 导入 schema
- 第一版唯一输出格式是 `FormSchemaJson`。

最少结构：

```json
{
  "drawingItems": [],
  "formConf": {},
  "idGlobal": "100",
  "treeNodeId": "100",
  "version": "1.1"
}
```

## 4. 菜单与页面映射

- 当前菜单最终以数据库 `sys_menu` 为真相源。
- 新增菜单时，`component` 必须对齐 `src/views` 下真实页面，或使用显式承接组件：
  - `RouteView`
  - `IFrameView`
- 父级分组菜单不要再映射到 `BasicLayout`。

## 5. 后台页面母版

- 后台 `CRUD / 树表 / 日志` 页面使用紧凑母版：
  - 筛选操作区
  - 轻量列表头
  - 数据区
- 页面内局部 `刷新` 按钮默认取消。
- `新增` 等主操作下沉到筛选操作区。
- `参数设置 / 表单构建 / 代码生成编辑页 / 服务监控 / 首页类` 保留压缩版轻标题条。

## 6. 本地验收

每次收尾至少跑：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1
./scripts/check-local.sh
```

如果前后端服务已启动，再追加：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```
