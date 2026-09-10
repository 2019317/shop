# ===========================================
# 常用命令
# 方案 B：Vultr + Neon + R2 + Vercel（make deploy）
# 方案 A'：单机全量，1GB 机型如 t3.micro（make deploy-all）
# ===========================================
.PHONY: help dev db-up db-down api-run api-build deploy deploy-all all-stop all-logs all-stats check backup

help:
	@echo "dev        启动本地依赖（PG/Redis/MinIO）"
	@echo "db-up      启动本地数据库与缓存"
	@echo "db-down    停止本地依赖"
	@echo "api-run    本地运行 Go API"
	@echo "api-build  构建 Go 二进制"
	@echo "deploy     方案B：Vultr 拉取最新镜像并重启"
	@echo "deploy-all 单机全量：1GB 机型一键部署全部服务"
	@echo "all-logs   查看单机全量服务日志"
	@echo "all-stats  实时查看容器资源占用"
	@echo "check      健康巡检（内存/OOM/磁盘/服务状态）"
	@echo "backup     手动执行一次数据库备份"

# ---------- 本地开发 ----------
dev: db-up
	@echo "本地依赖已启动：PG:5432 Redis:6379 MinIO:9000/9001"

db-up:
	docker compose -f infra/docker-compose.dev.yml up -d

db-down:
	docker compose -f infra/docker-compose.dev.yml down

api-run:
	cd apps/api && go run ./cmd/api -f etc/api.yaml

api-build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./apps/api/cmd/api

# ---------- 生产（Vultr VPS 上执行）----------
deploy:
	docker compose -f infra/docker-compose.prod.yml pull api
	docker compose -f infra/docker-compose.prod.yml up -d --no-deps api
	docker image prune -f

backup:
	bash infra/scripts/backup-db.sh

# ---------- 生产：单机全量（1GB 机型，如 AWS t3.micro）----------
deploy-all:
	docker compose -f infra/docker-compose.all-in-one.yml pull
	docker compose -f infra/docker-compose.all-in-one.yml up -d
	docker image prune -f

all-stop:
	docker compose -f infra/docker-compose.all-in-one.yml down

all-logs:
	docker compose -f infra/docker-compose.all-in-one.yml logs -f --tail=100

all-stats:
	docker stats

check:
	bash infra/scripts/healthcheck.sh
