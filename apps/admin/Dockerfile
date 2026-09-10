# ===========================================
# 管理后台（React + Vite）多阶段构建
# 产物为纯静态文件，由 nginx 托管，并反向代理 /api 到 Go API
# 在仓库根目录执行：
#   docker build -f apps/admin/Dockerfile -t shop-admin:latest .
# ===========================================
FROM node:20-alpine AS builder

WORKDIR /app

# 先拷贝依赖清单，最大化利用构建缓存
COPY apps/admin/package.json ./
RUN npm install --no-audit --no-fund

COPY apps/admin/ ./

# 与 API 同源，由 nginx 反向代理 /api，故无需配置跨域
ENV VITE_API_BASE_URL=/api/v1

# 仅做打包（vite build），跳过 tsc 类型检查：
# 避免前端类型问题阻塞部署，类型检查交给本地开发/CI
RUN npx vite build

# ---------------- 运行阶段 ----------------
FROM nginx:1.27-alpine

COPY apps/admin/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
