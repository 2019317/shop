# Stationery Shop — 手账独立站

面向海外销售手账制品的 DTC 独立站。前后端分离 + 前后台分离，后端 Go，可平滑演进为微服务。

## 一、部署架构（2GB 服务器方案，详见 `docs/deployment-2gb.md`）

```
用户 → Cloudflare (CDN / WAF / SSL)
         ├── yourbrand.com        → Vercel（Next.js 前台）
         ├── admin.yourbrand.com  → 本机 Nginx（React 后台静态，零内存）
         ├── api.yourbrand.com    → 2GB 服务器（Nginx → Go API 容器）
         └── img.yourbrand.com    → Cloudflare R2（制品图片，出站免费）

Go API (Docker)
         ├── PostgreSQL 容器（本机，640MB 上限）
         └── Redis 容器（192MB 上限）
```

| 组成 | 服务 | 月成本 |
|---|---|---|
| 应用服务器 | t3.small 2GB（Go API + PG + Redis + Nginx） | ~$15（新账号首年免费） |
| 图片存储 | Cloudflare R2（10GB 内免费） | $0 |
| CDN / WAF | Cloudflare 免费版 | $0 |
| 前台托管 | Vercel Pro（商用必需） | $20 |
| 邮件 / 监控 | Resend / Sentry 免费层 | $0 |

## 二、目录结构

```
.
├── apps/
│   ├── api/          # Go 后端（go-zero）
│   ├── web/          # Next.js 前台商城（Vercel）
│   └── admin/        # React SPA 管理后台（Cloudflare Pages）
├── packages/         # 前端共享（types / api-client / ui）
├── infra/
│   ├── nginx/        # Nginx 反代配置
│   ├── scripts/      # VPS 初始化、数据库备份
│   ├── sql/          # 数据库初始化脚本
│   ├── docker-compose.dev.yml   # 本地依赖（PG/Redis/MinIO）
│   └── docker-compose.prod.yml  # 生产（API/Nginx/Redis）
├── .github/workflows/# CI/CD：构建镜像 → GHCR → Vultr
└── Makefile
```

## 三、技术选型

| 层 | 选型 |
|---|---|
| 后端 | Go 1.22 + go-zero |
| 数据库 | PostgreSQL 16（Neon），`bigint` 存分 |
| 缓存 | Redis 7 |
| 对象存储 | Cloudflare R2（S3 协议） |
| 前台 | Next.js 14（SSR/ISR，利于 Google SEO） |
| 后台 | React + Vite SPA |
| 支付 | Stripe（适配器封装，可换 PayPal） |
| 部署 | Docker + GitHub Actions |

**架构纪律（保证后期可拆分）：**
1. `handler` 只做参数绑定与响应，业务逻辑在 `logic`；
2. 跨模块禁止直接查表，只能调用模块导出的 Service 接口；
3. 跨模块副作用走领域事件（当前内存实现，后期无缝换 `asynq`/MQ）；
4. 金额一律 `bigint` 存分，时间一律 `timestamptz` UTC。

## 四、本地开发

```bash
# 1. 启动依赖（PostgreSQL / Redis / MinIO 模拟 R2）
make db-up

# 2. 准备配置
cp .env.2gb.example .env
cp apps/api/etc/api.yaml.example apps/api/etc/api.yaml

# 3. 执行数据库迁移与种子数据
psql "postgres://shop:shop@localhost:5432/shop" -f infra/sql/migrations/000001_init.up.sql
psql "postgres://shop:shop@localhost:5432/shop" -f infra/sql/migrations/000002_seed.up.sql

# 4. 创建后台管理员
go run ./apps/api/cmd/seed -f apps/api/etc/api.yaml \
  -email admin@yourbrand.com -password 'your-password'

# 5. 运行 API
make api-run

# 6. 启动前台与后台
pnpm install
pnpm dev:admin    # http://localhost:5173
pnpm dev:web      # http://localhost:3000

# 7. 验证
curl http://localhost:8888/api/v1/healthz
```

## 五、生产部署

### 1. Vultr VPS（Go API + Nginx）

```bash
# 首次初始化（Ubuntu 22.04+）
sudo bash infra/scripts/setup-vps.sh

# 克隆仓库到 /opt/shop，配置 .env 后启动
cd /opt/shop
docker compose -f infra/docker-compose.prod.yml up -d
```

证书：在 Cloudflare 申请**源证书（Origin Certificate）**，放到 `infra/certs/origin.crt|origin.key`，
SSL/TLS 加密模式设为 **Full (strict)**。

### 2. Neon（PostgreSQL）

- 创建项目，复制连接串填入 `.env` 的 `DATABASE_URL`（需带 `sslmode=require`）；
- 开启自动备份与 PITR；
- 每日额外异地备份：

```bash
# crontab：0 3 * * * /opt/shop/infra/scripts/backup-db.sh
DATABASE_URL=xxx R2_BACKUP_BUCKET=shop-backups bash infra/scripts/backup-db.sh
```

### 3. Cloudflare R2（图片）

- 创建 bucket `shop-assets`，开启公共访问并绑定自定义域名 `img.yourbrand.com`；
- 在 Cloudflare 配置图片缓存规则（长期缓存 + 版本号或哈希命名）；
- Go 侧用 `aws-sdk-go-v2` 以 S3 协议访问，R2 **出站流量免费**。

### 4. Vercel（前台）

- 导入 `apps/web`，框架选 Next.js；
- 配置环境变量：`NEXT_PUBLIC_API_BASE_URL`、`NEXT_PUBLIC_SITE_URL`、`NEXT_PUBLIC_IMG_BASE_URL`；
- 建议区域：北美 `iad1`，欧洲 `fra1`（按目标市场调整）。

### 5. CI/CD

推送到 `main` 且 `apps/api/**` 有变更时自动构建镜像并部署到 Vultr。
需在 GitHub Secrets 配置：`VPS_HOST`、`VPS_USER`、`VPS_SSH_KEY`、`DEPLOY_PATH`。

## 六、方案 A'：单机全量部署（1GB 机型，如 AWS t3.micro）

适用于把 **Go API + PostgreSQL + Redis + Nginx + 静态后台** 全部塞进一台 1GB 内存机器。

⚠️ **前提约束（必须遵守）：**

| 约束 | 原因 |
|---|---|
| **前台 Next.js 必须放 Vercel** | Node SSR 约需 300MB，放本机必 OOM |
| **图片必须放 Cloudflare R2** | 本机磁盘有限，且 AWS 出站流量按 $0.09/GB 计费 |
| **禁止在本机 `docker build`** | t3 是突发性能实例，编译会耗尽 CPU 积分导致降频 |
| **必须配置 2GB swap** | 1GB 内存无缓冲空间，瞬时峰值直接 OOM |

**内存分配（总计约 635MB / 1024MB）：**

```
系统+Docker ~250MB │ Nginx 64MB │ Go API 192MB │ Redis 96MB │ PostgreSQL 320MB
```

### 部署步骤

```bash
# 1. 系统初始化（swap、内核调优、防火墙、定时备份）
sudo bash infra/scripts/setup-single-box.sh

# 2. 配置环境变量（数据库改为本机）
cp .env.example .env
# 修改 .env：DATABASE_URL=postgres://shop:密码@postgres:5432/shop?sslmode=disable

# 3. 构建后台静态产物（本地执行，不要在生产机构建）
cd apps/admin && pnpm build && cd ../..

# 4. 一键启动全部服务
make deploy-all

# 5. 每日巡检（内存/OOM/磁盘/服务状态）
make check
```

### 机型档位（在 `.env` 中选择一套）

| 档位 | 机型 | PG | Redis | API | Nginx | 合计 |
|---|---|---|---|---|---|---|
| A | t3.micro（1GB） | 320M | 96M | 192M | 64M | ~920M |
| B | t3.small（2GB） | 640M | 192M | 256M | 96M | ~1.4G |
| **C** | **t3.medium（4GB）** | **1G** | **320M** | **384M** | **128M** | **~2.2G** |

> ⚠️ **AWS 型号对照**：2C4G 是 **t3.medium**，不是 t3.small（t3.small 仅 2GB）。AWS 不支持自定义 CPU/内存配比。

4GB 机型剩余约 1.8GB，可启用可选的前台容器（不推荐，见下）：

```bash
docker compose -f infra/docker-compose.all-in-one.yml --profile frontend up -d
```

**建议始终将前台放 Vercel**：免费、全球 CDN、自动扩容，不消耗你的 CPU 积分与内存。

### 关于 T3 突发性能

- T3 **默认启用 Unlimited 模式**，积分耗尽**不会降频**（与 T2 不同）；
- 但超出基线部分按 **$0.05 / vCPU-hour** 计费（t3.medium 基线为 20% CPU）；
- 独立站日常负载远低于基线，**基本不会产生超额费**；
- 建议在 CloudWatch 对 `CPUCreditBalance` 设置告警，持续为 0 时排查是否有异常任务（如在机器上编译）。

```bash
make all-stats   # 实时看容器内存占用
make all-logs    # 查看日志
make check       # 健康巡检，剩余内存 <150MB 会告警
```

### 必须迁移的信号

出现以下任一情况，请立即把数据库迁到 Neon 或升级机型：

- `make check` 显示可用内存 **< 150MB**
- Swap 持续占用 **> 500MB**（说明内存长期不足）
- 日志中出现 **OOM killed**（PostgreSQL 被杀有数据损坏风险）
- 页面响应 **P95 > 2s**（t3 CPU 积分耗尽降频）

### AWS 免费层注意事项

- 免费期限 **12 个月**，到期后 t3.micro 约 $7.6/月
- 出站流量 100GB/月免费，超出按 **$0.09/GB** —— 图片务必走 Cloudflare CDN
- 在 Billing 面板设置 **$1 账单告警**
- EBS 默认 30GB，注意日志轮转（已在 compose 中配置 `max-size: 10m`）

## 七、注意事项

- 云服务器、Vercel、Cloudflare 账号注册时**固定网络环境**，避免频繁切换 IP 触发风控；
- 为所有云服务设置**账单告警与额度上限**；
- 面向欧盟访客需提供隐私政策与 Cookie 同意（GDPR）；
- 境外收款（Stripe）需要海外主体，这是独立于技术架构的前置条件。

## 八、API 接口

统一前缀 `/api/v1`，响应格式 `{ code, msg, data }`，`code=0` 表示成功。

### 前台（公开）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 健康检查 |
| GET | `/products` | 商品列表，支持 `category` / `q` / `sort` / `page` / `page_size` |
| GET | `/products/:slug` | 商品详情（含变体与图片） |
| GET | `/categories` | 类目列表 |
| POST | `/orders` | 提交订单（服务端计价、库存预占） |
| GET | `/orders/:order_no` | 按订单号查询 |
| POST | `/payments/notify` | 支付回调（幂等，当前供 mock 渠道调用） |
| GET | `/readyz` | 就绪探针（检查数据库与关键表是否存在） |
| GET | `/coupons` | 前台可用优惠券列表 |
| POST | `/coupons/validate` | 优惠券试算（返回可抵扣金额，不占用名额） |

### 后台（需 `Authorization: Bearer <token>`）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/admin/auth/login` | 登录获取 JWT |
| GET | `/admin/products` | 商品列表（支持 `keyword` 搜索） |
| GET | `/admin/products/:id` | 商品详情 |
| POST | `/admin/products` | 创建商品（含变体与图片） |
| PUT | `/admin/products/:id` | 更新商品 |
| DELETE | `/admin/products/:id` | 归档商品 |
| GET | `/admin/categories` | 类目列表 |
| POST | `/admin/categories` | 创建类目 |
| PUT | `/admin/categories/:id` | 更新类目 |
| DELETE | `/admin/categories/:id` | 隐藏类目 |
| POST | `/admin/assets/presign` | 获取图片直传 R2 的预签名地址 |
| GET | `/admin/orders` | 订单列表（支持 `status` / `keyword`） |
| GET | `/admin/orders/:id` | 订单详情 |
| POST | `/admin/orders/:id/ship` | 发货（填物流单号） |
| POST | `/admin/orders/:id/delivered` | 将运单标记为已送达 |
| POST | `/admin/orders/:id/cancel` | 取消订单并回滚库存 |
| GET | `/admin/coupons` | 优惠券列表（支持 `keyword` / `status`） |
| GET | `/admin/coupons/:id` | 优惠券详情 |
| POST | `/admin/coupons` | 创建优惠券 |
| PUT | `/admin/coupons/:id` | 更新优惠券（券码与已用量不可改） |
| POST | `/admin/coupons/:id/status` | 启用/停用优惠券 |
| GET | `/admin/shipping/rules` | 运费规则列表 |
| POST | `/admin/shipping/rules` | 创建运费规则 |
| PUT | `/admin/shipping/rules/:id` | 更新运费规则 |
| DELETE | `/admin/shipping/rules/:id` | 停用运费规则 |
| GET | `/admin/logs` | 请求日志查询（支持 `level` / `keyword` / `path`） |
| DELETE | `/admin/logs` | 清空请求日志 |

## 九、订单与支付流程

```
下单 POST /orders
  1. 按 SKU 从数据库取价（不信任前端金额）
  2. 匹配运费规则（国家 / 金额 / 重量，满额包邮）
  3. 创建订单（pending）+ 订单项价格快照
  4. 预占库存（SELECT FOR UPDATE 行锁，防超卖）
  5. 调用支付渠道创建支付意图
  6. 发布 order.created 事件

支付回调 POST /payments/notify
  1. event_id 幂等去重（唯一约束）
  2. 校验金额与订单状态
  3. 提交库存预占（真正扣减）
  4. 订单流转 pending → paid，发布 order.paid

发货 POST /admin/orders/:id/ship
  paid → fulfilled，写入物流单号，发布 order.shipped
```

**支付渠道为适配器模式**（`pkg/payment/gateway.go`）：
当前仅 `mock` 实现，不引入任何 SDK。接入 Stripe 时新增 `stripe.go` 实现同一接口，
修改配置 `Payment.Driver: stripe` 即可，**业务代码零改动**。

**领域事件**（`pkg/event/bus.go`）：当前为内存同步实现，
后期换 asynq / RabbitMQ 只需替换 Bus 实现，订阅方代码不变。

## 十、当前进度

## 九、当前进度

- [x] 部署架构与基础设施配置（2GB 服务器 / R2 / Vercel）
- [x] 单机全量部署方案（含内存档位与巡检脚本）
- [x] Go API 骨架与 CI/CD
- [x] 数据库表结构（商品 / SKU / 库存 / 订单 / 支付 / 履约 / 营销）
- [x] 后台：登录、商品管理、类目管理、图片直传 R2、订单管理
- [x] 前台：首页、商品列表、商品详情、购物车、结算、订单查询
- [x] 订单闭环：下单、库存预占、支付回调幂等、发货、取消回滚
- [x] 支付适配器抽象（mock 实现，未引入 SDK）
- [ ] 接入 Stripe 真实支付
- [ ] 优惠券核销、邮件通知、多语言多币种
