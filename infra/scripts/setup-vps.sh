#!/usr/bin/env bash
# ===========================================
# Vultr VPS 首次初始化（Ubuntu 22.04+）
# 用法：sudo bash infra/scripts/setup-vps.sh
# ===========================================
set -euo pipefail

echo "==> 更新系统"
apt-get update -y && apt-get upgrade -y

echo "==> 安装 Docker"
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi

echo "==> 安装 Docker Compose 插件"
apt-get install -y docker-compose-plugin git ufw wget

echo "==> 配置防火墙（仅放行 22/80/443）"
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "==> 创建部署目录"
mkdir -p /opt/shop/{infra/certs,infra/logs,data}
mkdir -p /opt/shop/infra/certbot/www

echo "==> 开启 swap（2C4G 机器防 OOM）"
if [ ! -f /swapfile ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
fi

echo "==> 完成。接下来："
echo "  1) 将仓库克隆到 /opt/shop"
echo "  2) 配置 .env（指向 Neon 数据库、R2、Stripe）"
echo "  3) 放置 Cloudflare 源证书到 /opt/shop/infra/certs/origin.crt|origin.key"
echo "  4) docker compose -f infra/docker-compose.prod.yml up -d"
