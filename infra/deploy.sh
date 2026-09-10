#!/usr/bin/env bash
# ===========================================
# 本地/单机部署快捷脚本（无域名场景）
# 用法（在仓库根目录执行）：
#   ./infra/deploy.sh up        # 构建并启动
#   ./infra/deploy.sh down      # 停止并删除数据卷
#   ./infra/deploy.sh ps        # 查看状态
#   ./infra/deploy.sh logs      # 查看 API 日志
#   ./infra/deploy.sh restart   # 重启 API
#   ./infra/deploy.sh psql      # 进入数据库
#
# 说明：
#   必须显式把 .env 导出到 shell 环境变量。
#   docker compose 的变量替换只会读「项目目录」下的 .env，
#   而本项目 compose 文件在 infra/，项目目录就是 infra/，
#   导致仓库根目录的 .env 读不到，密码会静默回落到默认值 change_me，
#   最终出现 "password authentication failed"。
#   这里用 set -a + source 绕开该问题，确保 ${PG_PASSWORD} 等能正确解析。
# ===========================================
set -euo pipefail

cd "$(dirname "$0")/.."

ENV_FILE=".env"
COMPOSE_FILE="infra/docker-compose.local.yml"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "错误：未找到 $ENV_FILE，请先执行：cp .env.2gb.example $ENV_FILE 并填写密码"
  exit 1
fi

# 将 .env 导出为当前 shell 的环境变量，供 compose 做变量替换
set -a
# shellcheck disable=SC1090
. "./$ENV_FILE"
set +a

DC="docker-compose -f $COMPOSE_FILE"

case "${1:-up}" in
  up)
    $DC up -d --build
    ;;
  down)
    $DC down -v
    ;;
  ps)
    $DC ps
    ;;
  logs)
    $DC logs -f api
    ;;
  restart)
    $DC restart api
    ;;
  psql)
    docker exec -it shop-postgres psql -U "${PG_USER:-shop}" -d "${PG_DB:-shop}"
    ;;
  build)
    $DC build
    ;;
  *)
    echo "用法: $0 {up|down|ps|logs|restart|psql|build}"
    exit 1
    ;;
esac
