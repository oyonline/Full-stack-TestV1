#!/bin/bash
# AnnouncementAttachmentGC 监控告警脚本
# 用法: ./scripts/gc_monitor.sh [check|report]
# 建议 cron: 0 9 * * * cd /path/to/go-admin && ./scripts/gc_monitor.sh check

set -euo pipefail

DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-1Qazwsxedc01@}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_NAME="${DB_NAME:-full_stack_test_v1}"
LOG_FILE="${LOG_FILE:-./access.log}"
ALERT_LOG="${ALERT_LOG:-./temp/logs/gc_alert.log}"
ATT_DIR="${ATT_DIR:-./static/uploadfile/attachment}"

# 阈值
ATT_GROWTH_THRESHOLD="${ATT_GROWTH_THRESHOLD:-50}"
DISK_THRESHOLD="${DISK_THRESHOLD:-80}"
MATCHED_SPIKE_THRESHOLD="${MATCHED_SPIKE_THRESHOLD:-100}"

mysql_cmd() {
    mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" -e "$1" "$DB_NAME" 2>/dev/null
}

check_att_growth() {
    local today_count yesterday_count growth
    today_count=$(mysql_cmd "SELECT COUNT(*) FROM att_file WHERE module_key='admin' AND business_type LIKE 'announcement%';" | tail -1)
    # 简化: 如果当前 > 50 就告警(实际生产应对比昨日)
    if [ "$today_count" -gt "$ATT_GROWTH_THRESHOLD" ] 2>/dev/null; then
        echo "[ALERT] att_file announcement rows=$today_count > threshold=$ATT_GROWTH_THRESHOLD"
        return 1
    fi
    echo "[OK] att_file announcement rows=$today_count"
}

check_disk() {
    local usage
    if [ -d "$ATT_DIR" ]; then
        usage=$(df -h "$ATT_DIR" | awk 'NR==2 {gsub(/%/,""); print $5}')
    else
        usage=0
    fi
    if [ "${usage:-0}" -gt "$DISK_THRESHOLD" ] 2>/dev/null; then
        echo "[ALERT] disk usage=$usage% > threshold=$DISK_THRESHOLD%"
        return 1
    fi
    echo "[OK] disk usage=${usage:-0}%"
}

check_gc_matched() {
    local last_matched
    last_matched=$(grep '\[AnnouncementAttachmentGC\] summary total_matched=' "$LOG_FILE" 2>/dev/null | tail -1 | sed -n 's/.*total_matched=\([0-9]*\).*/\1/p')
    if [ -z "$last_matched" ]; then
        echo "[WARN] no GC log found"
        return 1
    fi
    if [ "$last_matched" -gt "$MATCHED_SPIKE_THRESHOLD" ] 2>/dev/null; then
        echo "[ALERT] GC total_matched=$last_matched > threshold=$MATCHED_SPIKE_THRESHOLD"
        return 1
    fi
    echo "[OK] GC last total_matched=$last_matched"
}

gc_report() {
    echo "=== AnnouncementAttachmentGC Daily Report $(date '+%Y-%m-%d %H:%M:%S') ==="
    check_att_growth
    check_disk
    check_gc_matched
    echo "=== END ==="
}

case "${1:-check}" in
    check)
        gc_report >> "$ALERT_LOG" 2>&1
        ;;
    report)
        gc_report
        ;;
    *)
        echo "Usage: $0 [check|report]"
        exit 1
        ;;
esac
