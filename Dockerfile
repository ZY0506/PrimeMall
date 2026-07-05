# ============================================================
# Dockerfile - PrimeMall Server
# 构建阶段一次性编译所有 9 个服务二进制
# 运行阶段通过 target 选择具体服务（避免并行编译拥塞）
# ============================================================
# docker compose build 会自动并行，但 BuildKit 会缓存
# 公共层（builder），只有首次需要编译，后续秒完成
# ============================================================

# ========================
# Stage 1: 构建阶段 — 编译全部 9 个服务
# ========================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# 国内 Go 代理（解决 go mod download 被墙问题）
ENV GOPROXY=https://goproxy.cn,direct

# 缓存依赖层
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制全部源码
COPY . .

# 一次性编译所有服务（顺序执行，避免并行竞争）
RUN set -e && \
    echo "=== Building user-rpc ===" && \
    go build -ldflags="-s -w" -o /app/user-rpc     ./apps/service/user/rpc && \
    echo "=== Building product-rpc ===" && \
    go build -ldflags="-s -w" -o /app/product-rpc  ./apps/service/product/rpc && \
    echo "=== Building order-rpc ===" && \
    go build -ldflags="-s -w" -o /app/order-rpc    ./apps/service/order/rpc && \
    echo "=== Building payment-rpc ===" && \
    go build -ldflags="-s -w" -o /app/payment-rpc  ./apps/service/payment/rpc && \
    echo "=== Building marketing-rpc ===" && \
    go build -ldflags="-s -w" -o /app/marketing-rpc ./apps/service/marketing/rpc && \
    echo "=== Building admin-rpc ===" && \
    go build -ldflags="-s -w" -o /app/admin-rpc    ./apps/service/admin/rpc && \
    echo "=== Building search-rpc ===" && \
    go build -ldflags="-s -w" -o /app/search-rpc   ./apps/service/search/rpc && \
    echo "=== Building shop-api ===" && \
    go build -ldflags="-s -w" -o /app/shop-api     ./apps/gateway/shop && \
    echo "=== Building admin-api ===" && \
    go build -ldflags="-s -w" -o /app/admin-api    ./apps/gateway/admin && \
    echo "=== All builds complete ==="

# ========================
# 运行阶段基镜像（通用配置）
# ========================
FROM alpine:3.21 AS base

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai
ENV LANG=C.UTF-8

WORKDIR /app

# 各服务 main() 会加载 .env 文件，提供占位
COPY .env.example ./.env

ENTRYPOINT ["./server"]

# ============================================================
# 以下为 9 个服务的运行阶段
# 每个仅 COPY 自己的二进制和配置，镜像极小
# ============================================================

# --- user.rpc (实际端口 8081) ---
FROM base AS user-rpc
EXPOSE 8081
COPY --from=builder /app/user-rpc ./server
COPY --from=builder /build/apps/service/user/rpc/etc/ ./etc/

# --- product.rpc (实际端口 8082) ---
FROM base AS product-rpc
EXPOSE 8082
COPY --from=builder /app/product-rpc ./server
COPY --from=builder /build/apps/service/product/rpc/etc/ ./etc/

# --- order.rpc (实际端口 8083) ---
FROM base AS order-rpc
EXPOSE 8083
COPY --from=builder /app/order-rpc ./server
COPY --from=builder /build/apps/service/order/rpc/etc/ ./etc/

# --- payment.rpc (实际端口 8084) ---
FROM base AS payment-rpc
EXPOSE 8084
COPY --from=builder /app/payment-rpc ./server
COPY --from=builder /build/apps/service/payment/rpc/etc/ ./etc/

# --- marketing.rpc (实际端口 8085) ---
FROM base AS marketing-rpc
EXPOSE 8085
COPY --from=builder /app/marketing-rpc ./server
COPY --from=builder /build/apps/service/marketing/rpc/etc/ ./etc/

# --- admin.rpc (实际端口 8086) ---
FROM base AS admin-rpc
EXPOSE 8086
COPY --from=builder /app/admin-rpc ./server
COPY --from=builder /build/apps/service/admin/rpc/etc/ ./etc/

# --- search.rpc (实际端口 8087) ---
FROM base AS search-rpc
EXPOSE 8087
COPY --from=builder /app/search-rpc ./server
COPY --from=builder /build/apps/service/search/rpc/etc/ ./etc/
# 替换硬编码的 ES 地址为环境变量
RUN sed -i 's|http://localhost:19200|http://${ES_HOST}:${ES_PORT}|g' ./etc/search.yaml

# --- shop-api (实际端口 9091，3 个实例各自容器隔离) ---
FROM base AS shop-api
EXPOSE 9091
COPY --from=builder /app/shop-api ./server
COPY --from=builder /build/apps/gateway/shop/etc/ ./etc/
# Casbin 路径兼容（config 中写死了 apps/gateway/shop/etc/ 的相对路径）
RUN mkdir -p apps/gateway/shop && cp -r etc apps/gateway/shop/

# --- admin-api (实际端口 8088) ---
FROM base AS admin-api
EXPOSE 8088
COPY --from=builder /app/admin-api ./server
COPY --from=builder /build/apps/gateway/admin/etc/ ./etc/
