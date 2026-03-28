# 已知问题与限制

## 表单构建依赖外部资源

- 表单构建当前承接的是后端静态页。
- 静态页依赖外部 CDN 资源。
- 若浏览器网络环境或拦截策略导致外链资源失败，页面可能白屏或不完整。

## 数据库日志开关默认不在主配置启用

- 数据库操作日志和登录日志是否落库，依赖 `logger.enableddb`。
- 当前建议在 `settings.local.yml` 中显式配置。
- 如果改了本地日志开关但后端未重启，页面会继续看起来“没有日志”。

## 代码生成危险动作仍关闭

- 当前代码生成支持：
  - 导入
  - 配置维护
  - 保存
  - 预览
  - 移除
- 当前不支持：
  - 生成到项目
  - 生成菜单/API

## 菜单组件路径必须真实可映射

- 数据库中一旦写了不存在的 `component`，前端就会落到 not-found。
- 父级分组菜单如果误写成会渲染布局的组件，可能导致二次侧栏或内容区错位。

## 字典管理的旧菜单方案已废弃

- `字典数据` 作为独立页面的历史方案已经废弃。
- 如果有人手工执行旧的 `menu-batch3-web-antd.sql` / `menu-batch4-dict-log-fix.sql` 历史片段，或按旧认知恢复 `menu_id=59`、`menu_id=240`，左侧导航和权限模型会重新回到错误状态。
- 当前正式状态应为：
  - 左侧导航只保留 `字典类型`
  - `241/242/243` 挂在 `543` 下
  - `1775300000000_remove_dict_data_page.go` 已执行

## 空库或半初始化库会直接导致导航缺失

- 本地库如果 `sys_menu` 为空，`/api/v1/menurole` 就没有菜单可返回。
- 这种情况下首页可能还能进，但左侧导航会完全缺失。
- 先查 `sys_menu`、`sys_user`、`sys_role`、`sys_user_role`、`sys_migration`，不要先在前端写临时菜单。

## 修改迁移源码后，旧二进制不会自动带上新迁移

- 当前项目执行 `./go-admin migrate ...` 时，实际跑的是本地 `./go-admin` 二进制里编进去的迁移版本。
- 如果只改了 `go-admin/cmd/migrate/migration/version/*.go` 源码，但没有重新 `go build -o ./go-admin .`，新迁移不会执行。

## `sys_user` 字段容易与业务模型脱节

- 当前业务模型已经依赖 `open_id`、`job_title`、`open_department_id`、`open_department_ids`、`cn_name`。
- 如果本地库缺这些列，新增用户会报 `Error 1054 (42S22): Unknown column 'open_id' in 'field list'`。
- 这类问题应通过正式 migration 修复，不能靠手工改一台机器的表结构长期维持。

## 头像配置依赖新字段

- 当前 working tree 中的头像方案依赖 `sys_user.avatar_type` 和 `sys_user.avatar_color`。
- 但相关 migration `1775000000000_user_avatar_profile.go` 仍未进入已提交历史。
- 这意味着：头像这条线目前仍是本地待收口状态，不能当作所有环境默认可用的正式主线能力。

## 平台能力层已部分落地，但仍有待收口项

- `workflow / module_registry / attachment` 及对应最小前端验收层已经进入已提交历史。
- 但 `统一业务操作日志最小规范` 仍未形成完整已提交真相源。
- 继续推进时应优先补齐真相源和承接层，不再把平台能力重新塞回单个业务模块内部。

## 阶段口径可能滞后

- 旧文档可能仍把 finance 或本地后端实验层写成当前主线，这是过时口径。
- 当前应以后端业务实验层已基本收口、回到 clone 基线附近为准。
- clone 基线本身已经包含同事融合进来的业务代码，不能把当前仓库描述成“已删掉业务代码”。
- 当前不建议重新在本地随意扩写 finance / feishu / kingdee / biz_action_log / `sys_user` 扩展类后端业务线。

## 阶段文档仍可能保留旧判断

- `platform-layer-audit.md`、`platform-capability-phase1.md`、`PROJECT_CONVENTIONS.md` 中保留了阶段性判断或工作草案。
- 这些文件可以作为追溯线索，但不能替代当前代码和正式迁移链。

## 旧历史文档仍可能存在

- `go-admin/README.md` 和 `go-admin/README.Zh-cn.md` 主要是上游项目说明，不代表当前工作区的真实运行方式。
- 当前仓库以根 README、`docs/` 和 `.ai-memory/` 为准。
