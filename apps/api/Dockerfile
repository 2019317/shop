# ===========================================
# Go API 多阶段构建（产物约 15-25MB，非 root 运行）
# 在仓库根目录执行：
#   docker build -f apps/api/Dockerfile -t shop-api:latest .
# ===========================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 先拷贝依赖清单，最大化利用构建缓存
COPY go.mod go.sum* ./
RUN go mod download

# 单 module 多应用，仅拷贝后端目录
COPY apps/api/ ./apps/api/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /build/bin/api ./apps/api/cmd/api

# ---------------- 运行阶段 ----------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app -g 10001 \
    && adduser -S -G app -u 10001 app

WORKDIR /app

COPY --from=builder /build/bin/api /app/api
COPY --from=builder /build/apps/api/etc /app/etc

RUN chown -R app:app /app

USER 10001:10001
EXPOSE 8888

ENV TZ=UTC
ENV APP_ENV=production

CMD ["/app/api", "-f", "/app/etc/api.yaml"]
