# Step 1: 使用 Golang 镜像作为构建阶段
FROM golang:1.22-alpine AS builder

# Step 2: 设置工作目录
WORKDIR /app

# Step 3: 复制 idb 目录下的所有文件到工作目录
COPY . /app

# Step 4: 切换到 center 目录并安装依赖，编译项目
RUN cd /app/center && \
    go mod tidy && \
    go build -o idb .

# Step 5: 使用更小的镜像作为最终运行环境
FROM alpine:3.18

# Step 6: 安装 bash, curl 和 sed
RUN apk add --no-cache bash curl sed

# Step 7: 创建必要的目录结构
RUN mkdir -p /etc/idb /var/lib/idb /var/log/idb /run/idb

# Step 8: 复制编译好的二进制文件和配置文件到对应的目录
COPY --from=builder /app/center/idb /var/lib/idb/idb
COPY center/idb.conf /etc/idb/idb.conf
COPY center/entrypoint.sh /var/lib/idb/entrypoint.sh

# Step 9: 设置执行权限
RUN chmod +x /var/lib/idb/entrypoint.sh /var/lib/idb/idb

# Step 10: 设置工作目录，并指定容器启动命令
WORKDIR /var/lib/idb

# 使用脚本作为 ENTRYPOINT，确保启动时端口写入配置文件
ENTRYPOINT ["/var/lib/idb/entrypoint.sh", "/var/lib/idb/idb"]
