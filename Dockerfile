# 构建阶段
FROM golang:1.22-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制整个项目目录
COPY . .

# 进入 center 目录
WORKDIR /app/center

# 下载依赖
RUN go mod download

# 设置构建参数
ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION

# 设置环境变量
ENV CGO_ENABLED=1

# 编译 center
RUN cd /app/center && \
    go mod tidy && \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -tags=xpack -trimpath \
    -ldflags="-s -w -X 'github.com/sensdata/idb/center/global.Version=${VERSION}'" \
    -o idb .

# 运行阶段
FROM alpine:3.18

# 安装运行时必要的工具
RUN apk add --no-cache bash curl sed

# 创建必要的目录结构
RUN mkdir -p /etc/idb /var/lib/idb /var/log/idb /run/idb

# 从构建阶段复制编译好的应用和必要文件
COPY --from=builder /app/center/idb /var/lib/idb/idb
COPY center/idb.conf /etc/idb/idb.conf
COPY center/entrypoint.sh /var/lib/idb/entrypoint.sh

# 设置执行权限
RUN chmod +x /var/lib/idb/entrypoint.sh /var/lib/idb/idb

# # 创建健康检查脚本
# COPY center/healthcheck.sh /var/lib/idb/healthcheck.sh
# RUN chmod +x /var/lib/idb/healthcheck.sh

# 设置工作目录
WORKDIR /var/lib/idb

# # 设置健康检查
# HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
#   CMD /var/lib/idb/healthcheck.sh

# 设置入口点
ENTRYPOINT ["/var/lib/idb/entrypoint.sh", "/var/lib/idb/idb"]