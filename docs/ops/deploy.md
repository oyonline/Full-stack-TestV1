# 部署 SOP：product 角色种子 (1779000000003)

适用版本：包含 `1779000000003_product_role_seed` migration 的构建版本起。

---

## 1. 顺序约束（硬规则）

```
go-admin migrate   →   （验证通过）   →   restart api
```

**不能颠倒顺序。**

casbin enforcer 在应用启动时执行 `LoadPolicy`（见 `common/global/casbin.go`），从 `sys_casbin_rule` 加载全部策略到内存。若在 migrate 完成前重启 api，`product_admin` / `product_operator` 的策略尚未落库，任何非 `admin` rolekey 的用户登录后调接口都会被 `permission.go` 中间件拦截返回 403。

---

## 2. 执行步骤

### 2A. Fresh Install（全新环境，无历史数据）

1. 确保 `config/settings.yml` 已按目标环境配置好数据库连接。
2. 运行全部 migration：

   ```bash
   cd go-admin
   make migrate
   ```

   `make migrate` 会强制重新编译二进制再执行迁移，避免使用旧二进制漏跑 migration（见 Makefile 注释）。
   
3. 执行[验收校验 SQL](#3-migrate-后校验-sql)，确认两个角色策略数量符合预期。
4. 确认无报错后，启动 api 服务：

   ```bash
   nohup ./go-admin server -c config/settings.yml >> access.log 2>&1 &
   ```

5. 执行[服务存活检查](#5-服务存活-smoke-test)。

### 2B. 升级（已部署环境，`1779000000000`～`1779000000002` 已跑过）

1. 拉取包含 `1779000000003` 的新代码/二进制。
2. **先跑 migrate，不要先重启服务**：

   ```bash
   cd go-admin
   make migrate
   ```

   框架在 `Migrate.Migrate()` 入口先 `SELECT count` 检查 `sys_migration.version`；`1779000000000`～`1779000000002` 已存在会跳过，只执行 `1779000000003`。

3. 执行[验收校验 SQL](#3-migrate-后校验-sql)，确认策略数量。
4. 确认后才重启 api：

   ```bash
   killall go-admin || true
   nohup ./go-admin server -c config/settings.yml >> access.log 2>&1 &
   ```

5. 执行[服务存活检查](#5-服务存活-smoke-test)。

---

## 3. Migrate 后校验 SQL

在 migrate 完成后、重启 api 前，连接数据库执行：

```sql
SELECT v0, COUNT(*) AS policy_count
FROM sys_casbin_rule
WHERE v0 IN ('product_admin', 'product_operator')
GROUP BY v0;
```

### 期望结果

| v0 | policy_count |
|----|-------------|
| `product_admin` | 22 |
| `product_operator` | 15 |

> **注**：上述数字基于 EPO-49 §4.3+§4.4+§4.5 架构设计推算（C 菜单 5 条 + 按钮 API 展开 + 平台审批 API）。
> **EPO-54 merge 后**请从 PR 描述取实际运行值并更新此表。若实测数字与此不符，停止部署，升级架构师排查。

### 校验说明

策略数量拆解（供参考）：

**product_admin (22 条)**：
- SPU/SKU/类目/品牌 C 菜单读 API：5 条
- SPU 4 个操作按钮（含 submit 复用 edit 按钮）：6 条
- SKU 3 个操作按钮：3 条
- 类目 3 个操作按钮：3 条
- 品牌 3 个操作按钮：3 条
- 平台审批 API（todo/approve/reject）：3 条

**product_operator (15 条)**：
- SPU/SKU/类目/品牌 C 菜单读 API：5 条
- SPU 4 个操作按钮（含 submit）：6 条
- SKU 3 个操作按钮：3 条
- 平台工作流 API（started/withdraw）：2 条

---

## 4. 回滚预案

Migration `1779000000003` 全程在单事务内执行；数据库层失败时 GORM tx 自动回滚，`sys_migration` 中不会留下此版本记录，可直接修复后重跑。

**上线后发现策略绑定有误**的正确处理方式：

1. **不要回滚 `sys_migration` 记录**（否则会导致已运行的幂等 migration 被误判为未执行，重跑出重复数据）。
2. 新建 migration `1779000000004`，用 `DELETE FROM sys_casbin_rule WHERE v0=? AND v1=? AND v2=?` 或 `INSERT IGNORE` 做修正。
3. 按本 SOP 顺序（migrate → 验证 → restart）走一遍升级流程。
4. 回滚后必须在对应 issue 记录原因（不吞掉）。

---

## 5. 服务存活 Smoke Test

重启后执行：

```bash
# 检查进程
ps aux | grep go-admin | grep -v grep

# 检查端口（默认 8000，按实际 settings.yml 调整）
lsof -i :8000

# 健康检查（若有）
curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/api/v1/ping
```

期望：进程存在、端口监听、ping 返回 200（或系统无此端点则 404 也可，重点看无 500 启动崩溃）。

---

## 6. 常见问题

| 现象 | 原因 | 处理 |
|------|------|------|
| product_admin/product_operator 用户登录返回 403 | 策略未落库就重启了 api | 确认 migrate 已完成，执行校验 SQL，重启 api |
| `make migrate` 报 `config/settings.yml 不存在` | 缺配置文件 | 基于 `config/settings.local.yml.example` 创建配置后重试 |
| 第二次 `make migrate` 后策略数量变多 | migration 未走幂等路径 | 升级架构师排查 `1779000000003` 实现 |
| 校验 SQL 返回 0 行 | migration 未执行（`sys_migration` 中无 `1779000000003`） | 检查 migrate 日志，确认 make migrate 成功 |

---

*本文档由运维 agent 维护，依赖变更请更新校验期望数字并注明 PR 来源。*
