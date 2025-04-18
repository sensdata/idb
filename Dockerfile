# 构建阶段

# 编译 frontend 项目
FROM node:16 AS frontend-builder

# 设置工作目录
WORKDIR /app/frontend

# 复制 frontend 项目目录
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 编译 go 项目
FROM golang:1.22 AS builder

# 设置工作目录
WORKDIR /app

# 复制整个项目目录
COPY . .

# 生成证书
RUN openssl req -x509 -nodes -days 730 -newkey rsa:2048 -keyout key.pem -out cert.pem -config ssl.cnf

# 拷贝证书到 center 和 agent 目录
RUN mkdir -p center/global/certs agent/global/certs && \
    cp cert.pem center/global/certs/ && \
    cp cert.pem agent/global/certs/ && \
    cp key.pem center/global/certs/ && \
    cp key.pem agent/global/certs/

# 生成随机密钥
RUN KEY=$(openssl rand -base64 64 | tr -dc 'a-z0-9' | head -c 24 && echo) && \
    echo "KEY=${KEY}" >> /app/.env
    
# 设置构建参数
ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION

# 设置环境变量
ENV CGO_ENABLED=1

# 进入 center 目录
WORKDIR /app/center

# 下载依赖
RUN go mod download

# 编译 center
RUN KEY=$(cat /app/.env | grep KEY | cut -d'=' -f2) && \
    go mod tidy && \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -tags=xpack -trimpath \
    -ldflags="-s -w -X 'github.com/sensdata/idb/center/global.Version=${VERSION}' -X 'github.com/sensdata/idb/center/global.DefaultKey=${KEY}'" \
    -o idb .

# 进入 agent 目录
WORKDIR /app/agent

# 下载依赖
RUN go mod download

# 编译 agent
RUN go mod tidy && \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -tags=xpack -trimpath \
    -ldflags="-s -w -X 'github.com/sensdata/idb/agent/global.Version=${VERSION}'" \
    -o idb-agent .

# 创建 agent 包
RUN mkdir -p /app/agent-pkg && \
    cp idb-agent /app/agent-pkg/ && \
    cp idb-agent.service /app/agent-pkg/ && \
    cp idb-agent.conf /app/agent-pkg/ && \
    cp install-agent.sh /app/agent-pkg/ && \
    KEY=$(cat /app/.env | grep KEY | cut -d'=' -f2) && \
    sed -i "s/secret_key=.*/secret_key=${KEY}/" /app/agent-pkg/idb-agent.conf && \
    tar -czvf /app/idb-agent.tar.gz -C /app/agent-pkg .

# 创建 agent.version文件
RUN echo "${VERSION}" > /app/idb-agent.version

# 运行阶段
FROM debian:bookworm

# 安装运行时必要的工具
RUN apt-get update && apt-get install -y \
    bash \
    procps \
    curl \
    sed \
    && rm -rf /var/lib/apt/lists/*

# 创建 center 必要的目录结构
RUN mkdir -p /etc/idb /var/log/idb /run/idb /var/lib/idb /var/lib/idb/data /var/lib/idb/agent

# 从构建阶段复制编译好的 center 应用和必要文件
COPY --from=frontend-builder /app/frontend/dist/. /var/lib/idb/home
COPY --from=builder /app/center/idb /var/lib/idb/idb
COPY --from=builder /app/center/global/certs/cert.pem /var/lib/idb/cert.pem
COPY --from=builder /app/center/global/certs/key.pem /var/lib/idb/key.pem
COPY --from=builder /app/idb-agent.tar.gz /var/lib/idb/agent/idb-agent.tar.gz
COPY --from=builder /app/idb-agent.version /var/lib/idb/agent/idb-agent.version
COPY center/idb.conf /etc/idb/idb.conf
COPY center/entrypoint.sh /var/lib/idb/entrypoint.sh

# 创建软链接到 /usr/local/bin
RUN ln -sf /var/lib/idb/idb /usr/local/bin/idb

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