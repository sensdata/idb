# Step 1: 构建阶段，使用 Golang 1.22
FROM golang:1.22-alpine AS builder

# Step 2: 设置工作目录
WORKDIR /app

# Step 3: 复制 idb 项目到工作目录
COPY . /app

# Step 4: 设置环境变量
ENV CGO_ENABLED=1
ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION

# Step 5: 编译center
RUN cd /app/center && \
    go mod tidy && \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -tags=xpack -trimpath \
    -ldflags="-s -w -X 'github.com/sensdata/idb/center/global.Version=${VERSION}'" \
    -o idb .

# Step 6: 选择运行时镜像(alpine:3.18支持arch amd64)
FROM alpine:3.18

# Step 7: 安装必要的工具 (alpine镜像可能没有)
RUN apk add --no-cache bash curl sed gcc musl-dev

# Step 8: 创建必要的目录结构（应用目录）
RUN mkdir -p /etc/idb /var/lib/idb /var/log/idb /run/idb

# Step 8: 复制可执行文件和配置文件到对应的目录
COPY --from=builder /app/center/idb /var/lib/idb/idb
COPY center/idb.conf /etc/idb/idb.conf
COPY center/entrypoint.sh /var/lib/idb/entrypoint.sh

# Step 9: 设置执行权限
RUN chmod +x /var/lib/idb/entrypoint.sh /var/lib/idb/idb

# Debugging: 查看 /var/lib/idb/ 中的文件
RUN ls -l /var/lib/idb/

# Step 10: 设置工作目录，并指定容器启动命令
WORKDIR /var/lib/idb

# 使用脚本作为 ENTRYPOINT，确保启动时端口写入配置文件
ENTRYPOINT ["/var/lib/idb/entrypoint.sh", "/var/lib/idb/idb"]
