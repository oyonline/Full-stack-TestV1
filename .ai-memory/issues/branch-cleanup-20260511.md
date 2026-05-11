# Branch Cleanup Report — 2026-05-11

## 已归档并删除的分支

| 原分支 | archive tag | 说明 |
|--------|-------------|------|
| origin/agent/agent/1145ee57 | archive/agent-agent-1145ee57 | ahead=2, behind=27, 无独有业务价值 |
| origin/agent/agent/294190a3 | archive/agent-agent-294190a3 | ahead=2, behind=34, 无独有业务价值 |
| origin/agent/agent/3eb3abca | archive/agent-agent-3eb3abca | ahead=2, behind=34, 无独有业务价值 |
| origin/agent/02/5bcc69ee | archive/agent-02-5bcc69ee | ahead=0, 安全删除 |
| origin/agent/02/08cb6ea6 | archive/agent-02-08cb6ea6 | ahead=1 (78fe1de), main 已有等价功能 (609bab0) |
| origin/agent/02/f7541f8c | archive/agent-02-f7541f8c | ahead=1 (6248929), main 已有等价功能 (a22c38f) |

## 待处理分支清单（建议 PR / 人工 review）

以下分支含有未合入 main 的 commit，**不要直接删除**，建议由对应开发 agent 提 PR 或确认是否仍需保留。

### 1. origin/agent/02/04ce92ff
- **ahead of main**: 3 commits
- **commits**:
  - `233f6d8` feat(spu): Drawer before-close 拦截未保存变动确认
  - `e94290d` feat(spu): 补齐 4 项缺失 spec
  - `d0348c3` feat(spu): 补齐 SPU 详情场景 Drawer + 独立页 + 审批历史 Tab
- **建议**: 功能完整，建议由后端开发02 或前端开发提 PR 合入 main。

### 2. origin/agent/02/f9f9ee40
- **ahead of main**: 1 commit
- **commits**:
  - `ac09b43` feat(module_gate): remove EnsureModuleEnabled fallback passthrough
- **建议**: 单 commit，改动范围小，建议由后端开发01/02 提 PR 合入 main。

### 3. origin/feat/epo-88-spu-online-offline
- **ahead of main**: 3 commits
- **commits**:
  - `b42a935` feat(spu): 实施上下架方案 B + is_online 字段 (EPO-88)
  - `773a3d1` docs(decisions): add SPU offline strategy decision record
  - `d5d1db9` feat(sku): 接入 dataScope，SKU 主管理页改为只读
- **建议**: 属于 EPO-88 功能分支，建议由产品/架构师确认方案 B 是否仍有效，再决定提 PR 或废弃。

## 本地分支状态

- `local-backup-20260509-1814`：保留，等待下次清理周期。
- 其他已合入 main 的本地 feat 分支：已清理。
