#!/usr/bin/env bash
# ===========================================
# 数据库备份：pg_dump（Neon）→ 压缩 → 上传到 Cloudflare R2
# 托管数据库也有自动备份，但异地备份必须自己留一份
# crontab：0 3 * * * /opt/shop/infra/scripts/backup-db.sh >> /var/log/shop-backup.log 2>&1
# ===========================================
set -euo pipefail

: "${DATABASE_URL:?请设置 DATABASE_URL}"
: "${R2_BACKUP_BUCKET:=shop-backups}"

TIMESTAMP=$(date -u +"%Y%m%d-%H%M%S")
BACKUP_FILE="/tmp/shop-${TIMESTAMP}.sql.gz"
RETENTION_DAYS=${RETENTION_DAYS:-14}

echo "[$(date -u)] 开始备份..."
pg_dump "$DATABASE_URL" --no-owner --no-acl | gzip > "$BACKUP_FILE"
echo "[$(date -u)] 备份完成：$(du -h "$BACKUP_FILE" | cut -f1)"

if command -v rclone &> /dev/null; then
  echo "[$(date -u)] 上传到 R2..."
  rclone copy "$BACKUP_FILE" "r2:${R2_BACKUP_BUCKET}/daily/"
  echo "[$(date -u)] 上传完成"
else
  echo "[WARN] 未安装 rclone，备份仅保存在本地：${BACKUP_FILE}"
fi

# 清理过期本地备份
find /tmp -name "shop-*.sql.gz" -mtime +${RETENTION_DAYS} -delete

echo "[$(date -u)] 备份流程结束"
