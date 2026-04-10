# ============================================
# Aurora Go - 多阶段构建 (极简镜像 ~15-25MB)
# ============================================

# Stage 1: 构建阶段
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先复制依赖文件，利用Docker缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并构建
COPY . .

# 编译: 静态链接, 去除调试信息, 优化体积
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /aurora-server \
    ./cmd/server

# Stage 2: 运行阶段 (极简alpine)
FROM alpine:3.20
LABEL maintainer="Aurora Team"
LABEL description="Aurora Blog Go Backend with AI Agent (tRPC-Agent-Go)"
LABEL version="1.0.0"

# 安装基础依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && rm -rf /var/cache/apk/*

# 创建非root用户
RUN adduser -D -u 1000 aurora

WORKDIR /app

# 从构建阶段复制二进制和配置
COPY --from=builder /aurora-server .
COPY --from=builder /app/configs ./configs

# 赋权
USER aurora

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./aurora-server"]
CMD ["--config", "configs/config.yaml"]
