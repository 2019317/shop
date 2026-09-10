# ===========================================
# 前台商城（Next.js App Router）多阶段构建
# 产物为 standalone 运行包，由 node 直接运行
# 在仓库根目录执行：
#   docker build -f apps/web/Dockerfile --build-arg NEXT_PUBLIC_API_BASE_URL=http://IP:8888/api/v1 -t shop-web:latest .
# ===========================================
FROM node:20-alpine AS deps

WORKDIR /app

# 依赖清单（apps/web 使用 npm，锁定文件在其目录内）
COPY apps/web/package.json apps/web/package-lock.json* ./
RUN npm install --no-audit --no-fund

# ---------------- 构建阶段 ----------------
FROM node:20-alpine AS builder

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY apps/web/ ./

# NEXT_PUBLIC_* 在构建时会被内联进浏览器产物，因此必须作为构建参数传入
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8888/api/v1
ARG NEXT_PUBLIC_IMG_BASE_URL=
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_IMG_BASE_URL=$NEXT_PUBLIC_IMG_BASE_URL
# 生产构建
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN npm run build

# ---------------- 运行阶段 ----------------
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# standalone 输出仅包含运行所需文件
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
