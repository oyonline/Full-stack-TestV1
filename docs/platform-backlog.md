# 平台能力待办（摘自模块地图）

更新时间：2026-05-13

本文从 [`.ai-memory/module-map.md`](../.ai-memory/module-map.md) 抽取「仍待收口 / 未接真实数据」的平台向条目，便于排期；完成某项后应同步更新 `module-map` 与本文件。

## 通知中心

- 前端仍有右上角入口占位，**未接真实通知数据**。
- 后续需定义接口契约、未读策略与后端真相源。

## 统一业务操作日志

- `common/audit` 已提供契约与 helper；部分业务路径仍需统一改用 middleware helper，避免手拼 `AuditMeta`。
- 详见 `.ai-memory/backend-frontend-contracts.md` 中的业务操作日志段落。

## 代码生成（二阶段）

- 当前刻意关闭「生成到项目」「生成菜单/API」等危险动作（产品决策）。
- 若开放二阶段能力，需单独立项：权限、回滚与 CI 门禁。

## 其他模块地图中的边界表述

- `module-map` 内关于 finance-budget、通用单据状态流转等条目以该文件为准；此处不重复展开。
