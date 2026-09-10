# 2GB 服务器方案（后端专用，前台与图片外置）

机型：AWS **t3.small**（2 vCPU / 2 GiB）或同规格 VPS。

---

## 一、架构总览

```
                    用户
                     │
              Cloudflare（CDN + WAF + SSL）
                     │
   ┌─────────────────┼──────────────────┬─────────────────┐
   │                 │                  │                 │
yourbrand.com   admin.yourbrand.com  api.yourbrand.com  img.yourbrand.com
   │                 │                  │                 │
Vercel          本機 Nginx         本機 Nginx → Go API   Cloudflare R2
（前台 Next.js）  （后台静态）        （容器）            （图片，出站免费）
                                        │
                                   ┌────┴────┐
                                PostgreSQL  Redis
                                （本机容器）（本机容器）
```

**服务器只跑：Go API + PostgreSQL + Redis + Nginx（托管后台静态）**
**外置：前台（Vercel）、图片（R2）**

---

## 二、为什么这样分（三个关键理由）

| 外置项 | 理由 |
|---|---|
| **图片 → R2** | 2GB 机器磁盘/内存都撑不住图片；R2 出站流量**免费**，而 AWS 出站 $0.09/GB |
| **前台 → Vercel** | Node SSR 常驻需 300-500MB，2GB 放不下；Vercel 带全球 CDN，海外访问更快 |
| **后台 → 本机 Nginx** | 后台是纯静态 SPA，**零内存开销**，放本机最省事 |

---

## 三、内存预算（服务器）

| 组件 | 上限 | 关键调优 |
|---|---|---|
| PostgreSQL | 640MB | `shared_buffers=128MB`、`max_connections=40`、`work_mem=4MB` |
| Go API | 256MB | `GOGC=80`、`GOMEMLIMIT=220MiB` |
| Redis | 192MB | `maxmemory=128mb`、`allkeys-lru` |
| Nginx | 96MB | 仅托管后台静态 + 反代 |
| 系统 + Docker | ~250MB | — |
| **合计** | **~1.42GB / 2GB** | **剩余约 600MB** |

---

## 四、图片上传流程（关键设计：不经过服务器）

图片**绝不上传到服务器再转存 R2**（会打满内存和带宽）。正确做法是**前端直传**：

```
1. 后台点「上传图片」
        │
        ▼
2. 请求 Go API：POST /api/v1/admin/assets/presign
        │  返回 R2 预签名 URL（有效期 10 分钟）
        ▼
3. 浏览器用预签名 URL 直接 PUT 到 R2
        │  图片不经过服务器，零内存/带宽消耗
        ▼
4. 上传完成后，前端提交 object_key 给 Go API
        │
        ▼
5. Go API 把 key 存入数据库，返回可访问 URL
   （https://img.yourbrand.com/products/xxx.jpg）
```

**接口约定：**

```
POST /api/v1/admin/assets/presign
Request:  { "filename": "cover-a5.jpg", "content_type": "image/jpeg", "size": 2048576 }
Response: { "upload_url": "https://r2.../presigned", "object_key": "products/2026/09/xxx.jpg",
            "public_url": "https://img.yourbrand.com/products/2026/09/xxx.jpg" }
```

**校验规则（服务端必须做）：**
- 文件类型白名单：`jpg / png / webp / avif`
- 单文件 ≤ 5MB
- 按日期分目录：`products/YYYY/MM/`
- object_key 用 UUID 命名，避免覆盖与遍历

---

## 五、前台部署（Vercel）

### ⚠️ 重要：商业用途需要 Pro 计划

Vercel **Hobby（免费版）条款禁止商业使用**。你的站要卖货，属于商业用途，必须用 **Pro（$20/月）**。

| 方案 | 成本 | 说明 |
|---|---|---|
| **Vercel Pro** | $20/月 | ⭐ 完整支持 Next.js SSR/ISR，商用合规 |
| Cloudflare Pages | **免费** | 允许商用，但需静态导出或 `next-on-pages`，ISR/SSR 受限 |

**建议**：前期用 Vercel Pro（$20），等月流水起来这笔钱不值一提；若坚持零成本，可用 Cloudflare Pages + 静态导出（商品变动需重新构建）。

### Vercel 环境变量

在 Vercel 项目设置中配置：

```
NEXT_PUBLIC_API_BASE_URL=https://api.yourbrand.com
NEXT_PUBLIC_SITE_URL=https://yourbrand.com
NEXT_PUBLIC_IMG_BASE_URL=https://img.yourbrand.com
```

区域选择：`iad1`（美东）/ `fra1`（欧洲）/ `hnd1`（东京），按目标市场定。

---

## 六、R2 配置

1. Cloudflare 控制台 → R2 → 创建 bucket `shop-assets`
2. 绑定自定义域名 `img.yourbrand.com`（需在 Cloudflare 托管的域名下）
3. 创建 API Token（Object Read & Write），填入 `.env`
4. 缓存规则：图片路径 `Cache-Control: public, max-age=31536000, immutable`
5. 再建一个 bucket `shop-backups` 存数据库备份

**成本**：10GB 存储免费，超出 $0.015/GB/月，**出站流量永远免费**。

---

## 七、成本合计

| 项 | 费用 |
|---|---|
| t3.small（2GB 服务器） | ~$15/月（新账号首年可能免费） |
| Vercel Pro | $20/月（或 Cloudflare Pages $0） |
| Cloudflare R2 | $0（10GB 内免费） |
| Cloudflare 免费版 CDN/WAF | $0 |
| Neon | **不需要**（PG 在本机） |
| **合计** | **$15–35/月** |

---

## 八、部署步骤

```bash
# 1. 服务器初始化（2GB swap + 内核调优 + 防火墙 + 定时备份）
sudo bash infra/scripts/setup-single-box.sh

# 2. 配置 .env（复制「档位 B：2GB」参数块）
cp .env.example .env
nano .env    # 填 PG_PASSWORD、REDIS_PASSWORD、JWT_SECRET、R2、Stripe

# 3. 启动服务（默认不含前台容器）
make deploy-all

# 4. 健康检查
make check
```

前台部署：Vercel 导入 `apps/web`，配置上述环境变量即可。

---

## 九、升级信号

`make check` 每日巡检，出现任一则升级到 4GB 或迁移数据库到 Neon：

- 可用内存持续 < 200MB
- Swap 占用 > 500MB
- 日志出现 OOM killed
- API 响应 P95 > 2s
