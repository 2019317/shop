#!/usr/bin/env bash
# ===========================================
# 单机健康巡检（1GB 机型建议每日执行）
# 用法：bash infra/scripts/healthcheck.sh
# 建议加入 crontab：0 9 * * * /opt/shop/infra/scripts/healthcheck.sh
# ===========================================
set -uo pipefail

echo "===== 巡检时间：$(date '+%Y-%m-%d %H:%M:%S') ====="

echo ""
echo "--- 1. 内存（剩余 <150MB 需警惕）---"
free -h
AVAILABLE=$(free -m | awk '/^Mem:/{print $7}')
if [ "${AVAILABLE:-9999}" -lt 150 ]; then
  echo "⚠️  警告：可用内存仅 ${AVAILABLE}MB，建议升级机型或迁移数据库到 Neon"
fi

echo ""
echo "--- 2. Swap 使用（持续 >500MB 说明内存不足）---"
swapon --show

echo ""
echo "--- 3. 容器内存占用 Top ---"
docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPercy}}\t{{.CPUPerc}}" 2>/dev/null \
  | sed 's/MemPercy/MemPerc/'

echo ""
echo "--- 4. 磁盘（>80% 需清理）---"
df -h / | tail -1

echo ""
echo "--- 5. 服务健康状态 ---"
docker compose -f infra/docker-compose.all-in-one.yml ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null

echo ""
echo "--- 6. 最近是否有 OOM 杀进程 ---"
if command -v dmesg &> /dev/null; then
  dmesg -T 2>/dev/null | grep -i "out of memory" | tail -5 || echo "未发现 OOM 记录"
else
  grep -i "out of memory" /var/log/syslog 2>/dev/null | tail -5 || echo "未发现 OOM 记录"
fi

echo ""
echo "--- 7. API 存活 ---"
curl -s -o /dev/null -w "HTTP %{http_code} 耗时 %{time_total}s\n" --max-time 5 http://127.0.0.1:8888/healthz || echo "❌ API 无响应"

echo ""
echo "--- 8. AWS 突发实例 CPU 积分（非 AWS 可忽略）---"
TOKEN=$(curl -s -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 60" --max-time 2 2>/dev/null)
if [ -n "$TOKEN" ]; then
  # T3 默认 unlimited 模式：积分耗尽不会降频，但会产生超额费用($0.05/vCPU-hour)
  echo "实例类型: $(curl -s -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/instance-type 2>/dev/null)"
  echo "提示：请在 CloudWatch 监控 CPUCreditBalance，持续为 0 说明 CPU 长期超基线(20%)，会产生超额费用"
else
  echo "非 AWS 环境，跳过"
fi
