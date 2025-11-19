#!/bin/bash
set -euo pipefail

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/install.log
}

function Check_Root() {
    if [[ $EUID -ne 0 ]]; then
        log "请使用 root 或 sudo 权限运行此脚本"
        exit 1
    fi
}

function Configure_Docker_Mirror() {
    if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
        log "配置 Docker 镜像加速器..."
        mkdir -p /etc/docker
        cat > /etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://docker.1ms.run",
        "https://docker.1panel.live",
        "https://hub.fast360.xyz",
        "https://docker-0.unsee.tech",
        "https://docker.tbedu.top",
        "https://dockerpull.cn"
    ]
}
EOF
        systemctl daemon-reload
        systemctl restart docker
        log "Docker 镜像加速器配置完成"
    fi
}

function Install_Docker(){
    if which docker >/dev/null 2>&1; then
        log "检测到 Docker 已安装，跳过安装步骤"
        log "启动 Docker "
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log
        Configure_Docker_Mirror  # 配置镜像加速
    else
        log "... 在线安装 docker"

        if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
            sources=(
                "https://mirrors.aliyun.com/docker-ce"
                "https://mirrors.tencent.com/docker-ce"
                "https://mirrors.163.com/docker-ce"
                "https://mirrors.cernet.edu.cn/docker-ce"
            )

            # 测试源的延迟和可用性
            test_source() {
                local url=$1
                local test_url="$url/linux/ubuntu/dists/"
                local delay=$(curl -o /dev/null -s -w "%{time_total}" --connect-timeout 3 --max-time 5 "$test_url")
                local code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 --max-time 5 "$test_url")

                # 如果返回码不是 200，则认为不可用
                if [[ "$code" != "200" ]]; then
                    echo "fail"
                else
                    echo "$delay"
                fi
            }

            min_delay=99999
            selected_source=""

            for source in "${sources[@]}"; do
                result=$(test_source "$source")
                if [[ "$result" == "fail" ]]; then
                    log "$source 不可用，跳过"
                    continue
                fi
                log "$source 延迟: ${result}s"
                if (( $(awk 'BEGIN {print '"$result"' < '"$min_delay"'}') )); then
                    min_delay=$result
                    selected_source=$source
                fi
            done

            if [ -n "$selected_source" ]; then
                log "选择延迟最低且可用的源: $selected_source (延迟 ${min_delay}s)"
                export DOWNLOAD_URL="$selected_source"
            else
                log "所有国内源不可用，fallback 到官方源 https://download.docker.com"
                export DOWNLOAD_URL="https://download.docker.com"
            fi
        else
            log "非中国大陆地区，无需更改源"
            export DOWNLOAD_URL="https://download.docker.com"
        fi

        # 保证 /etc/apt/sources.list.d 存在（某些精简系统会缺失）
        if [[ ! -d /etc/apt/sources.list.d ]]; then
            log "/etc/apt/sources.list.d 不存在，正在创建..."
            mkdir -p /etc/apt/sources.list.d
            chmod 755 /etc/apt/sources.list.d
        fi

        # 安装 docker
        curl -fsSL "https://get.docker.com" -o get-docker.sh
        sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log

        log "... 启动 docker"
        systemctl enable docker
        systemctl daemon-reload
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log

        docker_config_folder="/etc/docker"
        [[ ! -d "$docker_config_folder" ]] && mkdir -p "$docker_config_folder"

        if ! docker version >/dev/null 2>&1; then
            log "docker 安装失败"
            exit 1
        fi
        log "docker 安装成功"
        Configure_Docker_Mirror
    fi
}

function Install_Compose(){
    # 检查 docker compose 是否可用
    docker compose version >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        log "... 在线安装 Docker Compose 插件"

        mkdir -p ~/.docker/cli-plugins
        COMPOSE_VERSION="v2.26.1"
        OS=$(uname | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        case "$ARCH" in
            x86_64) ARCH="amd64";;
            aarch64|arm64) ARCH="arm64";;
            armv7l) ARCH="armv7";;
            *) log "不支持的架构 $ARCH"; exit 1;;
        esac

        COMPOSE_URL="https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-${OS}-${ARCH}"

        log "下载 docker compose 插件..."
        curl -fL "$COMPOSE_URL" -o ~/.docker/cli-plugins/docker-compose || {
            log "docker compose 下载失败"
            exit 1
        }

        chmod +x ~/.docker/cli-plugins/docker-compose
        ln -sf ~/.docker/cli-plugins/docker-compose /usr/local/bin/docker-compose

        docker compose version >/dev/null 2>&1 || {
            log "Docker Compose 插件安装失败"
            exit 1
        }
        log "Docker Compose 插件安装成功"
    else
        compose_v=$(docker compose version 2>/dev/null)
        log "检测到 Docker Compose 已安装: $compose_v"
    fi
}

function main() {
    Check_Root
    Install_Docker
    Install_Compose
    log "Docker + Compose 环境安装完成"
}

main
