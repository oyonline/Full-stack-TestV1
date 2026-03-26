# Full-stack-TestV1 脚手架使用说明

## 1. 系统设置

- `参数设置` 是唯一正式系统设置入口
- 系统名称、Logo、Logo 占位色、登录页标题/说明、全局界面设置都会持久化到后端
- 页面启动时前端会从 `/api/v1/app-config` 拉取运行时配置

## 2. 代码生成

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
2. 先执行“导入到生成器”
3. 切到“生成配置”
4. 进入“编辑配置”
5. 保存后再做“模板预览”

## 3. 表单构建

- 前端通过同域 `/form-generator/` 承接后端静态 form-generator
- 当前补了一层 schema bridge，支持：
  - 下载 schema
  - 复制 schema
  - 导入 schema
- 第一版约定的唯一输出格式是 `FormSchemaJson`

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

## 4. 菜单与前端组件映射

- 当前菜单最终以数据库 `sys_menu` 为真相源
- 前端仍保留少量历史兼容映射，主要用于承接旧路径
- 新增菜单时，`component` 必须对齐 `src/views` 下真实页面或显式承接组件（如 `IFrameView`、`RouteView`）
- 父级分组菜单不要再映射到 `BasicLayout`，避免二次侧栏布局问题

## 5. 本地验收

建议每次收尾至少跑：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1
./scripts/check-local.sh
```

如果前后端服务已启动，再追加：

```bash
cd /Users/linshen/Desktop/Full-stack-TestV1/vue-vben-admin/apps/web-antd
pnpm test:e2e
```
