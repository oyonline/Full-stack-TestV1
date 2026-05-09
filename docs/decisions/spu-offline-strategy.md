# SPU 上下架长期方案决策记录

**决策来源**: [EPO-70](mention://issue/f0b58d96-502e-44de-97c3-8fe7a6925d34) — 产品提案(评论 §1~§5)+ 架构师 ack(评论 §8.1~§8.4)  
**决策日期**: 2026-05-09  
**状态**: 已确认(产品 + 架构师双方评论 ack)

---

## 1. 背景与三选一对比

详细对比见 [EPO-47](mention://issue/8d660f06-a90d-44f0-8a25-0f3c36c974ea) §3.B2。当前 SPU.status 状态机中 status=5(Offline)无任何后端接口入口,[EPO-53](mention://issue/67b07ca5-aae4-49dc-884a-034c4315dd78) 已临时去掉前端选项。长期方案三选一:

| 方案 | 描述 | 特点 |
|---|---|---|
| **A** | 永远不下架,SPU 一旦审核通过即"事实上线",通过 SKU 维度的 status 控制可售性 | 最简,但有语义问题 |
| **B** | 加一对动作 `Spu.GoOffline / Spu.GoOnline` + 审计 method,不走 workflow,直接 is_online 切换 | 适合无审批的运营动作,已选 |
| **C** | 给"下架"也走 workflow,复用 platform.workflow,新建 `spu_offline_review` definition_key | 最重,合规友好,不作为默认 |

### 否决方案 A 的理由

- **批量负担**:一个 SPU 平均带多个 SKU,商品停售要求一秒生效时,运营无法逐个翻 SKU。
- **语义错位**:SKU.status 控的是"该规格"是否可售;SPU 整体下架是"这件商品不卖了",两件事不同。
- **召回/合规场景**:质量、版权、合规问题时,客服/法务需要清晰的"商品已下架"系统状态作为举证依据。
- **报表分析**:GMV/活跃 SPU 指标需按 SPU 维度区分在架/下架。

### 否决方案 C 作为默认的理由

- 下架是防御动作,速度第一;审批等待期内继续可售 = 风险敞口。
- 加审计 method + `sys_opera_log` 已能满足合规追溯要求,不需要事前审批。
- 若未来出现强监管类目,可针对该类目单独开 `spu_offline_review` definition_key,本轮不强制全局走 C。

---

## 2. 决策结果:方案 B + 新增 `is_online` 字段

**决定采用方案 B,但存储字段调整为新增 `is_online`,废弃复用 `status=5`。**

### 原因:废弃 status=5

复用 `spu.status=5` 会让 `status` 同时表达两件事:审核态 + 在售态。
- `status` 当前语义:审核状态机(1 草稿 → 2 审核中 → 3 通过 / 4 拒绝)。
- GoOnline 时目标 status 不确定:若拒绝态(status=4)的商品下架,回 3 会丢"曾被拒绝"的事实。
- 解法:状态正交化(比补 `previous_status` 字段更干净)。

### 新字段定义

```go
type Spu struct {
    // ... existing fields
    Status   int  `gorm:"size:4;not null;default:1"` // 审核状态机:1 草稿 / 2 审核中 / 3 通过 / 4 拒绝。废弃 5。
    IsOnline bool `gorm:"not null;default:false;column:is_online;index"` // 是否在售。仅 status=3 时允许 true。
}
```

### 迁移步骤(后端开发执行)

1. 加列 `is_online tinyint(1) default 0`,加索引。
2. 数据回填:`UPDATE spu SET is_online=1 WHERE status=3;`(当前所有审核通过的 SPU 默认在售)
3. 把 `status=5` 的存量(若有)迁回 `status=3, is_online=0`。
4. 应用 cutover 后,删除 `SpuStatusOffline=5` 常量(或保留注释 deprecated)。

---

## 3. status 与 is_online 二维状态语义

| 显示文案 | status | is_online | 说明 |
|---|---|---|---|
| 草稿 | 1 | false | 创建/撤回回退 |
| 审核中 | 2 | false | SubmitForReview |
| 在售 | 3 | true | 审核通过且已上架 |
| 已下架 | 3 | false | 审核通过但已下架(运营可一键 GoOnline) |
| 已驳回 | 4 | false | workflow Rejected 终态 |

约束:其它 status(`is_online` 必须为 false),由 service 层强约束。

GoOnline 需要先进行内容修改时的场景:
- 因临时缺货/促销结束下架 → 重新上架:不审批,直接 GoOnline 回 `status=3, is_online=true`。
- 因内容/价格/合规问题下架,且重新上架前**修改了 SPU 内容**:走原 SubmitForReview 重新审批。
- 因合规问题下架,**未修改内容**强行重新上架:系统不允许(强约束)。

---

## 4. 权限点

新增两个权限点:

- `admin:spu:offline` — 下架 SPU
- `admin:spu:online` — 上架 SPU

绑定策略:
- `product_admin`:全量(所有 SPU)
- `product_operator`:配 dataScope(仅本部门 / 本人创建的 SPU)

实现:复用现有 casbin + dataScope 中间件,无额外框架成本。

---

## 5. 单向级联:GoOffline 级联 SKU 下架,GoOnline 不反向恢复

### GoOffline 行为

同事务内:

```sql
UPDATE sku SET status=1 WHERE spu_id=? AND status=2
```

SPU + SKU 同事务,失败回滚。

### GoOnline 行为

**不**自动恢复 SKU。理由:下架期间运营可能调整过 SKU 库存/价格,自动批量打开 = 误开风险。GoOnline 后,运营根据需要手动启用 SKU。

前端在 GoOnline 成功后的 toast 里提示:"请检查 SKU 启用状态"。

---

## 6. 审计落点:sys_opera_log

`wf_action_log` 强绑 `workflow_instance_id`,设计上服务于 platform.workflow,**不适合**记非 workflow 动作。

**审计走 `sys_opera_log`(go-admin 内置通用操作日志)**:

- 操作人 / 时间 / IP / params 全有,无新表无迁移。
- params 格式:`{"spu_id":..., "action":"go_offline", "reason":"...", "category_id":...}`。
- 若产品/法务后续需要"商品下架历史"作为业务功能,再单建 `spu_offline_history` 业务表,本轮先不做。

---

## 7. hook 位但不抽象的实现取向

不新建 interface、不预留 strategy 模式、不留空函数。**只是把 service 写成"可以扩展的形状",不当下抽象**:

- `Spu.GoOffline` service 入口走"前置校验链"(目前为空 slice),将来插 `categoryComplianceCheck` 一行注册即可。
- audit 记录时把 `category_id` 一并写入 `sys_opera_log.params`,将来按类目筛审计/统计零成本。

---

## 8. 接口契约

```go
// service 层
func (s *Spu) GoOffline(c *dto.SpuGoOfflineReq) error  // 必须 status=3 AND is_online=true
func (s *Spu) GoOnline(c *dto.SpuGoOnlineReq) error    // 必须 status=3 AND is_online=false

// dto
type SpuGoOfflineReq struct {
    SpuId  int64  `json:"spuId" binding:"required"`
    Reason string `json:"reason" binding:"required,max=255"` // 必填,审计用
}
type SpuGoOnlineReq struct {
    SpuId int64 `json:"spuId" binding:"required"`
}

// 路由
POST /api/v1/spu/:id/offline   // 权限:admin:spu:offline
POST /api/v1/spu/:id/online    // 权限:admin:spu:online
```

---

## 9. 与临时方案的口径

[EPO-53](mention://issue/67b07ca5-aae4-49dc-884a-034c4315dd78) 已落地"前端去掉 5 选项"。本决策记录确定方向后:

- **实现落地前**:状态 5 仍按 EPO-53 处理,前端不出现该选项,后端无入口。代码注释 + 文档统一口径:**"状态 5 = 暂未启用,等上下架业务流定义后再开放"**。
- **实现落地后**:前端把"已下架"选项加回过滤下拉;SPU 列表/详情的状态展示恢复完整;新增"下架/上架"按钮。

代码注释样板:

```go
// SpuStatusOffline = 5,采用方案 B(GoOffline/GoOnline 无审批),决策见 docs/decisions/spu-offline-strategy.md
```

---

## 参考

- [EPO-47](mention://issue/8d660f06-a90d-44f0-8a25-0f3c36c974ea) §3.B2 — 架构师技术优化方案(三选一对比)
- [EPO-70](mention://issue/f0b58d96-502e-44de-97c3-8fe7a6925d34) — 产品提案 + 架构师 ack 原始评论
- [EPO-53](mention://issue/67b07ca5-aae4-49dc-884a-034c4315dd78) — 前端临时去掉 status=5 选项
