#!/usr/bin/env bash
# ===========================================
# 1GB 小内存单机初始化（AWS t3.micro / 1C1G VPS）
# 用法：sudo bash infra/scripts/setup-single-box.sh
# ===========================================
set -euo pipefail

echo "==> 1/6 更新系统并安装 Docker"
apt-get update -y && apt-get upgrade -y
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi
apt-get install -y docker-compose-plugin git ufw wget postgresql-client

echo "==> 2/6 创建 2GB Swap（1GB 内存机型必做，防 OOM）"
if [ ! -f /swapfile ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo "swap 已启用：$(swapon --show | tail -1)"
else
  echo "swap 已存在，跳过"
fi

echo "==> 3/6 内核参数调优"
cat > /etc/sysctl.d/99-shop.conf <<'EOF'
# 尽量不用 swap，但保留应急能力
vm.swappiness=10
# 允许内存超分配（Go/Java 类应用需要）
vm.overcommit_memory=1
# 减少脏页回写压力
vm.dirty_ratio=15
vm.dirty_background_ratio=5
net.core.somaxconn=1024
EOF
sysctl --system

echo "==> 4/6 防火墙（仅放行 22/80/443）"
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "==> 5/6 创建目录"
mkdir -p /opt/shop/{infra/certs,infra/logs,data}
mkdir -p /opt/shop/infra/certbot/www

echo "==> 6/6 配置每日数据库备份（凌晨 3 点，避开流量高峰）"
cat > /etc/cron.d/shop-backup <<'EOF'
0 3 * * * root cd /opt/shop && DATABASE_URL="postgres://shop:PASSWORD@localhost:5432/shop" /opt/shop/infra/scripts/backup-db.sh >> /var/log/shop-backup.log 2>&1
EOF
chmod 644 /etc/cron.d/shop-backup

echo ""
echo "==> 完成。重要提醒："
echo "  1) t3.micro 是突发性能实例，禁止在本机执行 docker build（CI 构建后推送镜像，本机只 pull）"
echo "  2) 前台 Next.js 放 Vercel，不要部署到本机（Node SSR 约需 300MB）"
echo "  3) 图片必须放 Cloudflare R2，经由 Cloudflare CDN 回源，避免 AWS 出站流量费"
echo "  4) 执行：docker stats 实时观察内存，剩余低于 150MB 时请立即升级机型"
