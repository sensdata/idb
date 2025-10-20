# ---------- 构建 frontend ---------- #
FROM node:22-bookworm AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---------- 生成证书和密钥 ---------- #
FROM golang:1.23.4 AS certs-builder
WORKDIR /app/certs
COPY ssl.cnf ./
# 生成 PKCS#8 格式的私钥
RUN openssl genpkey -algorithm RSA -out key.pem -pkeyopt rsa_keygen_bits:2048
# 生成自签名的 CA 证书，使用已生成的 PKCS#8 私钥
RUN openssl req -x509 -new -key key.pem -out cert.pem -days 3650 -config ssl.cnf -extensions v3_ca
# 密钥
RUN KEY=$(openssl rand -base64 64 | tr -dc 'a-z0-9' | head -c 24 && echo) && \
    echo "KEY=${KEY}" >> /app/.env

# ---------- 构建 agent (使用debian较低版本) ---------- #
FROM debian:10 AS agent-builder

# 设置为非交互模式，防止 tzdata 等包交互式安装
ENV DEBIAN_FRONTEND=noninteractive

# 替换已过期的 buster 源为 archive 源，并安装构建依赖
RUN sed -i 's|http://deb.debian.org/debian|http://archive.debian.org/debian|g' /etc/apt/sources.list && \
    sed -i 's|http://security.debian.org|http://archive.debian.org|g' /etc/apt/sources.list && \
    echo 'Acquire::Check-Valid-Until "false";' > /etc/apt/apt.conf.d/99no-check-valid-until && \
    apt-get update && \
    apt-get install -y \
        curl \
        git \
        gcc \
        libc6-dev \
        make \
        sed

# 安装 golang
RUN curl -L -o /tmp/go.tar.gz https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/local/bin/go
    
# 环境变量
ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH
ENV CGO_ENABLED=1
# 编译参数
ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION
# 拷贝代码
WORKDIR /app
COPY . .
# 拷贝证书和密钥
COPY --from=certs-builder /app/certs/cert.pem agent/global/certs/
COPY --from=certs-builder /app/certs/key.pem agent/global/certs/
COPY --from=certs-builder /app/.env ./

# 进入agent目录
WORKDIR /app/agent
# 下载依赖
RUN go mod download
# 编译 agent
RUN GOOS=${GOOS} \
    GOARCH=${GOARCH} \
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

# ---------- 构建 center ---------- #
FROM golang:1.23.4 AS center-builder
# 环境变量
ENV CGO_ENABLED=1
# 编译参数
ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION
# 拷贝代码
WORKDIR /app
COPY . .
# 拷贝证书和密钥
COPY --from=certs-builder /app/certs/cert.pem center/global/certs/
COPY --from=certs-builder /app/certs/key.pem center/global/certs/
COPY --from=certs-builder /app/.env ./

# 进入 center 目录
WORKDIR /app/center
# 下载依赖
RUN go mod download
# 编译 center
RUN KEY=$(cat /app/.env | grep KEY | cut -d'=' -f2) && \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -tags=xpack -trimpath \
    -ldflags="-s -w -X 'github.com/sensdata/idb/center/global.Version=${VERSION}' -X 'github.com/sensdata/idb/center/global.DefaultKey=${KEY}'" \
    -o idb .

# ---------- 运行阶段 ---------- #
FROM debian:bookworm
# 安装运行时必要的工具
RUN apt-get update && apt-get install -y \
    bash \
    procps \
    curl \
    sed \
    vim-tiny \
    lsof \
    net-tools \
    ca-certificates \
    rsync \
    sshpass \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# 创建 center 必要的目录结构
RUN mkdir -p /etc/idb /var/log/idb /run/idb /var/lib/idb /var/lib/idb/data /var/lib/idb/agent

# 从构建阶段复制编译好的 center 应用和必要文件
COPY --from=frontend-builder /app/frontend/dist/. /var/lib/idb/home
COPY --from=center-builder /app/center/idb /var/lib/idb/idb
COPY --from=center-builder /app/center/global/certs/cert.pem /var/lib/idb/cert.pem
COPY --from=center-builder /app/center/global/certs/key.pem /var/lib/idb/key.pem
COPY --from=agent-builder /app/idb-agent.tar.gz /var/lib/idb/agent/idb-agent.tar.gz
COPY --from=agent-builder /app/idb-agent.version /var/lib/idb/agent/idb-agent.version
COPY center/idb.conf /etc/idb/idb.conf
COPY center/entrypoint.sh /var/lib/idb/entrypoint.sh

# 创建软链接到 /usr/local/bin
RUN ln -sf /var/lib/idb/idb /usr/local/bin/idb
# 设置执行权限
RUN chmod +x /var/lib/idb/entrypoint.sh /var/lib/idb/idb
# 设置工作目录
WORKDIR /var/lib/idb
# # 设置健康检查
# HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
#   CMD /var/lib/idb/healthcheck.sh
# 设置入口点
ENTRYPOINT ["/var/lib/idb/entrypoint.sh", "/var/lib/idb/idb"]