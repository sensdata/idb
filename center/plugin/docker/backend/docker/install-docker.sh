#!/bin/bash
set -euo pipefail

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function Check_Root() {
    if [[ $EUID -ne 0 ]]; then
        echo "请使用 root 或 sudo 权限运行此脚本"
        exit 1
    fi
}

function Configure_Docker_Mirror() {
    if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
        echo "配置 Docker 镜像加速器..."
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
        echo "Docker 镜像加速器配置完成"
    fi
}

function Install_Docker() {
    if which docker >/dev/null 2>&1; then
        echo "检测到 Docker 已安装，跳过安装步骤"
        echo "启动 Docker "
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/docker-install.log
        Configure_Docker_Mirror
    else
        echo "... 在线安装 docker"
        echo "当前地区: $(curl -s ipinfo.io/country)"

        if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
            sources=(
                "https://mirrors.aliyun.com/docker-ce"
                "https://mirrors.tencent.com/docker-ce"
                "https://mirrors.163.com/docker-ce"
                "https://mirrors.cernet.edu.cn/docker-ce"
            )

            # 测试源延迟
            echo "测试源延迟..."
            test_source() {
                local url=$1
                local test_url="$url/linux/ubuntu/dists/"
                local delay=$(curl -o /dev/null -s -w "%{time_total}" --connect-timeout 3 --max-time 5 "$test_url")
                local code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 --max-time 5 "$test_url")

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
                    echo "$source 不可用，跳过"
                    continue
                fi
                echo "$source 延迟: ${result}s"
                if (( $(awk 'BEGIN {print '"$result"' < '"$min_delay"'}') )); then
                    min_delay=$result
                    selected_source=$source
                fi
            done

            if [ -n "$selected_source" ]; then
                echo "选择延迟最低的源: $selected_source (延迟 ${min_delay}s)"
                export DOWNLOAD_URL="$selected_source"
            else
                echo "所有国内源不可用，fallback 到官方源 https://download.docker.com"
                export DOWNLOAD_URL="https://download.docker.com"
            fi
        else
            echo "非中国大陆地区，无需更改源"
            export DOWNLOAD_URL="https://download.docker.com"
        fi

        curl -fsSL "https://get.docker.com" -o get-docker.sh
        sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/docker-install.log

        echo "... 启动 docker"
        systemctl enable docker
        systemctl daemon-reload
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/docker-install.log

        if ! docker version >/dev/null 2>&1; then
            echo "docker 安装失败"
            exit 1
        fi
        echo "docker 安装成功"
        Configure_Docker_Mirror
    fi
}

function Install_Compose() {
    docker compose version >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "... 在线安装 Docker Compose 插件"

        mkdir -p ~/.docker/cli-plugins
        COMPOSE_VERSION="v2.26.1"
        OS=$(uname | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        case "$ARCH" in
            x86_64) ARCH="amd64";;
            aarch64|arm64) ARCH="arm64";;
            armv7l) ARCH="armv7";;
            *) echo "不支持的架构 $ARCH"; exit 1;;
        esac

        COMPOSE_URL="https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-${OS}-${ARCH}"

        echo "下载 docker compose 插件..."
        curl -fL "$COMPOSE_URL" -o ~/.docker/cli-plugins/docker-compose || {
            echo "docker compose 下载失败"
            exit 1
        }

        chmod +x ~/.docker/cli-plugins/docker-compose
        ln -sf ~/.docker/cli-plugins/docker-compose /usr/local/bin/docker-compose

        docker compose version >/dev/null 2>&1 || {
            echo "Docker Compose 插件安装失败"
            exit 1
        }
        echo "Docker Compose 插件安装成功"
    else
        compose_v=$(docker compose version 2>/dev/null)
        echo "检测到 Docker Compose 已安装: $compose_v"
    fi
}

function main() {
    Check_Root
    Install_Docker
    Install_Compose
    echo "Docker + Compose 环境安装完成"
}

main
