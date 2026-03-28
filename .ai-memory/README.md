# Full-stack-TestV1 项目记忆

这套目录是当前仓库的长期上下文入口，目标是减少重复排查、重复踩坑和重复做决策。

但要注意：

- `.ai-memory/*` 是工作记忆和阶段记录，不高于已提交代码、已提交 migration 和正式菜单真相源。
- 如果 `.ai-memory/*` 和当前代码冲突，以当前代码为准。
- `platform-layer-audit.md`、`platform-capability-phase1.md` 这类文件更偏阶段判断与过程记录，不能直接当作当前完成度公告。

## 先看哪些文件

- [project-decisions.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/project-decisions.md)
  记录已经确定、不应反复摇摆的产品和技术决策。
- [backend-frontend-contracts.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/backend-frontend-contracts.md)
  记录前后端协议、状态值、登录链路、多角色权限等关键契约。
- [known-issues.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/known-issues.md)
  记录已知坑、限制和容易误判的问题。
- [runbooks.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/runbooks.md)
  记录排查和验收手册。
- [module-map.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/module-map.md)
  记录模块边界和职责归属。
- [project-roadmap.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/project-roadmap.md)
  记录当前阶段完成情况与后续方向。
- [platform-layer-audit.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/platform-layer-audit.md)
  记录“平台底座层 / 平台能力层”的现状、边界问题和业务模块接入前建议。
- [platform-capability-phase1.md](/Users/linshen/Cursor/Full-stack-TestV1/.ai-memory/platform-capability-phase1.md)
  记录平台能力层第一期的最小闭环、P0 范围和正式实施顺序。

## 当前交接

- [本地联调交接](/Users/linshen/Cursor/Full-stack-TestV1/docs/local-dev-handoff.md)
  记录当前本机和当前 working tree 的本地联调基线，不等于仓库主线完成度。
- [头像系统设计](/Users/linshen/Cursor/Full-stack-TestV1/docs/avatar-system.md)
  记录当前本地头像方案；其中涉及未提交改动的部分只能按“本地状态”理解。

## 使用规则

新增一个会长期影响项目协作的结论时，优先写入下面的文件：

- 新的系统级决策：写入 `project-decisions.md`
- 新的前后端协议：写入 `backend-frontend-contracts.md`
- 新发现的坑和限制：写入 `known-issues.md`
- 新形成的稳定排查流程：写入 `runbooks.md`
- 新模块或模块边界变化：写入 `module-map.md`
- 新的平台分层判断和模块接入前审计：写入 `platform-layer-audit.md`
- 新的平台能力层第一期闭环、实施顺序和接入前约束：写入 `platform-capability-phase1.md`

默认约定：

- 后续出现重要改动时，项目记忆会同步更新，不需要每次单独提醒。
- “重要改动”包括：
  - 权限模型变化
  - 前后端协议变化
  - 系统设置能力变化
  - 菜单与路由正式值变化
  - 审计分类与日志链变化
  - 后台标准列表的通用交互与持久化约定变化
  - 新平台模块落地
  - 新的高频问题与稳定排查手册
- 普通样式微调、文案调整、局部 UI 小修不要求每次都写入记忆文档，除非它改变了项目约定或容易重复踩坑。

历史调试记录和旧 issue 资料保留在现有子目录中，只作为追溯材料，不再作为当前项目事实来源。

## 当前记忆范围

当前这套记忆已覆盖：

- 参数设置作为唯一系统设置入口
- 多角色权限模型
- 菜单与按钮权限规则
- 审计与操作日志链路
- 动态路由和菜单真相源
- 字典管理“目录页 -> 类型详情页”的正式主链路
- 代码生成、表单构建和后台标准页母版
- 路由级标准列表页的列个性化配置与本地持久化约定
- 常见登录、菜单、日志、构建问题的排查路径
- 平台底座层 / 平台能力层的分层认知与业务模块接入前建议
- 平台能力层第一期的最小闭环与正式实施顺序
