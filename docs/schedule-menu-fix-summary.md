# 定时任务 / SCHEDULE 菜单修复 - 收尾说明

## 1. 本轮改动文件清单

| 路径 | 操作 |
|------|------|
| `vue-vben-admin/apps/web-antd/src/router/access.ts` | 修改 |

（定时任务页面、job API 等为前期已有或独立提交，本轮仅涉及菜单/路由 patch 逻辑。）

---

## 2. 本轮功能验收清单

- [ ] 登录后左侧是否只剩一套「定时任务」
- [ ] 点击「定时任务 → Schedule」是否进入 `/admin/sys-job`
- [ ] 地址栏直达 `/admin/sys-job` 与左侧点击进入是否为同一页面
- [ ] Network 是否出现 `/v1/sysjob`（或项目内定时任务列表接口）
- [ ] 页面搜索、列表加载是否正常
- [ ] 刷新页面后仍能正常进入 `/admin/sys-job` 且菜单正常

---

## 3. Git 提交说明

**Commit message：**

```
fix(web-antd): 定时任务菜单去重并修复 Schedule 子菜单 404
```

**备注（本次改动摘要）：**

- 后端原菜单已有「定时任务」父菜单（path /schedule）及子菜单「Schedule」（path /schedule/manage），映射后 component 为 not-found 导致 404。
- 在 access.ts 的 patchScheduleInMappedList 中增加：若未命中「坏掉的顶层定时任务」节点，则查找后端映射后的 Schedule 子节点（title=Schedule 或 path=/schedule/manage），将其 path 改为 /admin/sys-job、component 改为 /admin/sys-job/index，并 return true。
- 从而不再 push 第二条顶层「定时任务」，左侧只保留一套菜单，点击 Schedule 进入 /admin/sys-job 正常。
- 未改登录逻辑；未处理「日志」子菜单。

---

## 4. 风险提醒

- 「日志」子菜单（/schedule/log）仍可能为 404，本轮未接通。
- 后端菜单 path（/schedule、/schedule/manage）与前端实际页面（/admin/sys-job）不一致，目前仅通过前端 patch 兼容，未改后端菜单配置。
