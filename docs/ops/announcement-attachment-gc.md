# 公告附件 GC 运维 SOP (EPO-96)

适用：`AnnouncementAttachmentGC` cron job(见 `go-admin/app/jobs/announcement_attachment_gc.go`)。

---

## 1. 部署

### 1.1 sys_job 配置

```sql
INSERT INTO sys_job (
  job_name, job_group, job_type,
  cron_expression, invoke_target, args,
  misfire_policy, concurrent, status,
  entry_id, created_at, updated_at
) VALUES (
  'AnnouncementAttachmentGC', 'DEFAULT', 2,
  '0 0 * * * *', 'AnnouncementAttachmentGC', '',
  1, 1, 2,
  0, NOW(3), NOW(3)
);
```

字段说明:

- `job_type=2` → 函数任务(对应 `app/jobs/jobbase.go` 的 `ExecJob` 分支)
- `invoke_target='AnnouncementAttachmentGC'` → 必须与 `app/jobs/examples.go` `InitJob()` 字典 key 完全一致
- `cron_expression='0 0 * * * *'` → robfig/cron v3 `NewWithSeconds()` 6 字段:`秒 分 时 日 月 周`,每小时整点触发
- `status=2` → 启用(`status=1` 是关闭,`setup()` 只加载 `WHERE status=2`)

### 1.2 启动验证

重启 `go-admin` 后,`access.log` 应有:

```
[INFO] JobCore Starting...
trace ... SELECT * FROM sys_job WHERE status = 2 ... [rows:1]
trace ... UPDATE sys_job SET entry_id=1, updated_at=... WHERE job_id=3 ...
[INFO] JobCore start success.
```

`sys_job.entry_id` 由 0 变非零 = cron 已注册到调度器。

---

## 2. 7 天 dry-run 观察期

默认 `GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN` 未设 → `dry_run=true`,只 SELECT 不 DELETE。

### 2.1 每日检查日志

```bash
cd /Users/linshen/Documents/Full-stack-TestV1/go-admin
grep "\[AnnouncementAttachmentGC\] summary" access.log | tail -24
```

每条日志格式:

```
2026-05-11 10:07:00.012+0800 info [AnnouncementAttachmentGC] summary total_matched=0 dry_run=true spend=10.970125ms
```

判断标准(MVP 数据量):

- `total_matched` 稳定在两位数以内 → 正常
- `total_matched` 突然 ≥ 100 → 异常,**停 cron 升级架构师**,见下方 §4

### 2.2 每日 att_file 行数与磁盘水位

```sql
-- 公告附件行数 + 增长
SELECT
  COUNT(*) AS total_rows,
  SUM(CASE WHEN created_at >= NOW() - INTERVAL 1 DAY THEN 1 ELSE 0 END) AS added_24h,
  SUM(CASE WHEN business_id = '0' AND created_at < NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END) AS temp_orphans_overdue
FROM att_file
WHERE module_key = 'admin'
  AND business_type IN ('announcement-inline', 'announcement-cover');
```

```bash
# 磁盘水位(本地 MVP 是单机)
df -h /Users/linshen/Documents/Full-stack-TestV1/go-admin/static/uploadfile/
```

阈值:

- `added_24h > 200` → 异常增长,人工查上传源
- 磁盘使用率 > 80% → 扩容 / 清理

---

## 3. 第 8 天切真删

```bash
# 1. 改 env(写到启动脚本 / systemd / docker-compose 三选一,本地用 .env)
export GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN=false

# 2. 重启 go-admin
cd /Users/linshen/Documents/Full-stack-TestV1/go-admin
./stop.sh
nohup ./go-admin server -c=config/settings.dev.yml >> access.log 2>&1 &

# 3. 盯下一个整点日志
tail -f access.log | grep --line-buffered "\[AnnouncementAttachmentGC\]"
```

预期看到:

```
[AnnouncementAttachmentGC] stage=temp_orphans removed=N
[AnnouncementAttachmentGC] summary total_removed=N dry_run=false spend=...
```

**`removed=N` 应与 dry-run 期最近一次 `total_matched` 接近**。差异巨大 → 见 §4。

---

## 4. 回滚 / 应急

### 4.1 临时停 cron(不删行,只关 schedule)

```sql
UPDATE sys_job SET status=1 WHERE job_id=3;
```

重启 go-admin,`setup()` 不再加载该行(`WHERE status=2`)。

### 4.2 切回 dry-run

```bash
unset GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN
# 或显式
export GO_ADMIN_ANNOUNCEMENT_GC_DRY_RUN=true
./stop.sh && nohup ./go-admin server -c=config/settings.dev.yml >> access.log 2>&1 &
```

### 4.3 完全卸载(EPO-95 回滚场景)

```sql
DELETE FROM sys_job WHERE job_id=3 AND invoke_target='AnnouncementAttachmentGC';
```

---

## 5. 升级路径

生产环境上 Prometheus/Grafana 后,把 §2.2 的 SQL 改写成 mysqld_exporter 自定义 metric +
PromQL alert:

```promql
# att_file 公告附件行数 24h 增长 > 200 告警
increase(announcement_attachment_rows_total[24h]) > 200

# 磁盘水位
node_filesystem_avail_bytes{mountpoint="/static/uploadfile"} / node_filesystem_size_bytes < 0.2
```

本 SOP 适用 MVP / 本地单机 / 无监控栈环境。
