#!/bin/bash

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/install.log
}

echo
cat << EOF     
██╗██████╗ ██████╗ 
██║██╔══██╗██╔══██╗
██║██║  ██║██████╔╝
██║██║  ██║██╔══██╗
██║██████╔╝██████╔╝
╚═╝╚═════╝ ╚═════╝ 
EOF

log "======================= 开始安装 ======================="

# 加速代理支持
# 用法: IDB_GITHUB_PROXY=https://dl.idb.net bash install.sh
# 如果未指定代理，自动检测 GitHub 连通性，不通则使用 dl.idb.net
IDB_DEFAULT_PROXY="https://dl.idb.net"

function Probe_HTTP_200() {
    local url="$1"
    local attempts="${2:-3}"
    local i
    local code

    for ((i = 1; i <= attempts; i++)); do
        code=$(curl -s --connect-timeout 5 --max-time 10 -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
        if [[ "$code" == "200" ]]; then
            return 0
        fi
        sleep 1
    done

    return 1
}

function Is_China_Region() {
    # 1. 通过 ipinfo.io 检测国家代码（最准确）
    local country=$(curl -s --connect-timeout 3 --max-time 5 https://ipinfo.io/country 2>/dev/null | tr -d '[:space:]')
    if [[ "$country" == "CN" || "$country" == "HK" || "$country" == "MO" ]]; then
        return 0
    fi
    # 2. 通过时区判断（ipinfo.io 不可用时的后备）
    local tz=$(timedatectl 2>/dev/null | grep "Time zone" | awk '{print $3}')
    [[ -z "$tz" ]] && tz="$TZ"
    if [[ "$tz" == *"Shanghai"* || "$tz" == *"Chongqing"* || "$tz" == *"Hong_Kong"* ]]; then
        return 0
    fi
    return 1
}

function Auto_Detect_Proxy() {
    if [[ -n "$IDB_GITHUB_PROXY" ]]; then
        log "使用指定代理: ${IDB_GITHUB_PROXY}"
        return
    fi

    if Is_China_Region; then
        # 中国区域：优先检测代理
        log "检测到中国区域，优先检测加速代理..."
        if Probe_HTTP_200 "${IDB_DEFAULT_PROXY}/github-api/repos/sensdata/idb/releases/latest"; then
            log "加速代理可用: ${IDB_DEFAULT_PROXY}"
            export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
            return
        fi
        log "加速代理不可用，尝试 GitHub 直连..."
        if Probe_HTTP_200 "https://api.github.com/repos/sensdata/idb/releases/latest"; then
            log "GitHub 直连正常"
            return
        fi
        log "GitHub 直连也不可用，仍使用代理: ${IDB_DEFAULT_PROXY}"
        export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
    else
        # 非中国区域：优先检测 GitHub
        log "检测 GitHub 连通性..."
        if Probe_HTTP_200 "https://api.github.com/repos/sensdata/idb/releases/latest"; then
            log "GitHub 直连正常"
            return
        fi
        log "GitHub 连接不可用，尝试加速代理..."
        if Probe_HTTP_200 "${IDB_DEFAULT_PROXY}/github-api/repos/sensdata/idb/releases/latest"; then
            log "加速代理可用: ${IDB_DEFAULT_PROXY}"
            export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
            return
        fi
        log "所有源均不可用，默认使用 GitHub 直连"
    fi
}

Auto_Detect_Proxy

GITHUB_API_URL="${IDB_GITHUB_PROXY:+${IDB_GITHUB_PROXY}/github-api}"
GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"
GITHUB_RELEASES_URL="${IDB_GITHUB_PROXY:+${IDB_GITHUB_PROXY}/github-releases}"
GITHUB_RELEASES_URL="${GITHUB_RELEASES_URL:-https://github.com}"

if [[ -n "$IDB_GITHUB_PROXY" ]]; then
    log "使用加速代理: ${IDB_GITHUB_PROXY}"
fi

function Check_Root() {
    if [[ $EUID -ne 0 ]]; then
        log "请使用 root 或 sudo 权限运行此脚本"
        exit 1
    fi
}

IDB_DOCKER_USER=""
IDB_DOCKER_HOME=""

function Resolve_User_Home() {
    local username="$1"
    local user_home=""

    if command -v getent >/dev/null 2>&1; then
        user_home=$(getent passwd "$username" 2>/dev/null | cut -d: -f6)
    fi

    if [[ -z "$user_home" ]]; then
        user_home=$(eval echo "~${username}" 2>/dev/null)
        if [[ "$user_home" == "~${username}" ]]; then
            user_home=""
        fi
    fi

    if [[ -z "$user_home" ]]; then
        user_home=$(sudo -u "$username" -H sh -lc 'printf %s "$HOME"' 2>/dev/null)
    fi

    printf '%s\n' "$user_home"
}

function Configure_Docker_Access() {
    if ! command -v docker >/dev/null 2>&1; then
        return 0
    fi

    if command docker info >/dev/null 2>&1; then
        return 0
    fi

    if [[ -n "$SUDO_USER" && "$SUDO_USER" != "root" ]]; then
        local docker_home
        docker_home=$(Resolve_User_Home "$SUDO_USER")
        if [[ -n "$docker_home" ]]; then
            if sudo -u "$SUDO_USER" -H env HOME="$docker_home" DOCKER_CONFIG="$docker_home/.docker" docker info >/dev/null 2>&1; then
                IDB_DOCKER_USER="$SUDO_USER"
                IDB_DOCKER_HOME="$docker_home"
                log "检测到 Docker 需通过用户 ${IDB_DOCKER_USER} 的上下文访问"
                return 0
            fi
        fi
    fi

    return 1
}

function docker() {
    if [[ -n "$IDB_DOCKER_USER" ]]; then
        command sudo -u "$IDB_DOCKER_USER" -H env HOME="$IDB_DOCKER_HOME" DOCKER_CONFIG="$IDB_DOCKER_HOME/.docker" docker "$@"
    else
        command docker "$@"
    fi
}

function Check_Architecture() {
    osCheck=`uname -a`
    if [[ $osCheck =~ 'x86_64' ]];then
        architecture="amd64"
    # elif [[ $osCheck =~ 'arm64' ]] || [[ $osCheck =~ 'aarch64' ]];then
    #     architecture="arm64"
    # elif [[ $osCheck =~ 'armv7l' ]];then
    #     architecture="armv7"
    # elif [[ $osCheck =~ 'ppc64le' ]];then
    #     architecture="ppc64le"
    # elif [[ $osCheck =~ 's390x' ]];then
    #     architecture="s390x"
    else
        log "暂不支持的系统架构，请参阅官方文档，选择受支持的系统。"
        exit 1
    fi
}

function Install_Docker(){
    if which docker >/dev/null 2>&1; then
        log "检测到 Docker 已安装，跳过安装步骤"
        if Configure_Docker_Access; then
            log "Docker daemon 连接正常"
            return 0
        fi

        if command -v systemctl >/dev/null 2>&1 && systemctl cat docker.service >/dev/null 2>&1; then
            log "启动 Docker "
            systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log
            if Configure_Docker_Access; then
                log "Docker daemon 已启动"
                return 0
            fi
        fi

        log "检测到 Docker 客户端，但当前环境无法连接 Docker daemon。"
        log "如果您使用 Docker Desktop，请用普通用户执行脚本，或保留 sudo 环境变量后重试。"
        exit 1
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
        Configure_Docker_Access >/dev/null 2>&1
        log "docker 安装成功"
    fi
}

function Install_Compose(){
    # 检查 docker compose 是否可用
    docker compose version >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        log "... 在线安装 Docker Compose 插件"

        if command -v apt-get >/dev/null 2>&1; then
            log "尝试通过 apt 安装 docker-compose-plugin..."
            apt-get update 2>&1 | tee -a ${CURRENT_DIR}/install.log
            apt-get install -y docker-compose-plugin 2>&1 | tee -a ${CURRENT_DIR}/install.log
        elif command -v dnf >/dev/null 2>&1; then
            log "尝试通过 dnf 安装 docker-compose-plugin..."
            dnf install -y docker-compose-plugin 2>&1 | tee -a ${CURRENT_DIR}/install.log
        elif command -v yum >/dev/null 2>&1; then
            log "尝试通过 yum 安装 docker-compose-plugin..."
            yum install -y docker-compose-plugin 2>&1 | tee -a ${CURRENT_DIR}/install.log
        fi

        if ! docker compose version >/dev/null 2>&1; then
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

            log "包管理器安装失败，回退到二进制下载: ${COMPOSE_URL}"
            curl -f --progress-bar -L "$COMPOSE_URL" -o ~/.docker/cli-plugins/docker-compose || {
                log "docker compose 下载失败，请检查网络，或先手动安装 docker-compose-plugin 后重试"
                exit 1
            }

            chmod +x ~/.docker/cli-plugins/docker-compose
            ln -sf ~/.docker/cli-plugins/docker-compose /usr/local/bin/docker-compose
        fi

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

function Check_Installation() {
    local containers
    containers=$(docker ps -a -q -f name=idb)
    if [[ -n "$containers" ]]; then
        log "检测到已安装的 IDB 容器"
        read -p "是否要升级安装？这将备份现有数据 [y/n]: " UPGRADE_IDB
        if [[ "$UPGRADE_IDB" == "Y" ]] || [[ "$UPGRADE_IDB" == "y" ]]; then
            Backup_Data
            log "准备安装新版本..."
            return 0
        else
            log "取消安装"
            exit 1
        fi
    fi
}

function Backup_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    log "清理临时目录..."
    rm -rf "$BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    
    if docker ps -q -f name=idb >/dev/null 2>&1; then
        log "停止 IDB 容器..."
        docker stop idb
    fi
    
    if docker ps -a -q -f name=idb >/dev/null 2>&1; then
        # 保存当前环境变量值
        docker inspect idb > "$BACKUP_DIR/container_info.json"
        
        # 从容器配置中获取实际的挂载路径
        local DATA_PATH=$(docker inspect idb --format '{{ range .Mounts }}{{ if eq .Destination "/var/lib/idb/data" }}{{ .Source }}{{ end }}{{ end }}')
        local LOG_PATH=$(docker inspect idb --format '{{ range .Mounts }}{{ if eq .Destination "/var/log/idb" }}{{ .Source }}{{ end }}{{ end }}')
        
        # 备份实际的数据目录
        if [[ -n "$DATA_PATH" && -d "$DATA_PATH" ]]; then
            log "备份数据目录: $DATA_PATH"
            cp -r "$DATA_PATH" "$BACKUP_DIR/data"
        fi
        
        if [[ -n "$LOG_PATH" && -d "$LOG_PATH" ]]; then
            log "备份日志目录: $LOG_PATH"
            cp -r "$LOG_PATH" "$BACKUP_DIR/logs"
        fi
        
        log "删除旧容器..."
        docker rm idb
        
        return 0
    fi
    
    return 1
}

function Restore_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log "未找到备份数据，执行全新安装"
        return 0
    fi
    
    log "开始恢复数据..."
    
    # 确保目标目录存在
    mkdir -p "${PANEL_DIR}/data"
    mkdir -p "${PANEL_DIR}/logs"
    
    # 恢复数据到宿主机目录
    if [[ -d "$BACKUP_DIR/data" ]]; then
        log "恢复数据目录到: ${PANEL_DIR}/data"
        cp -r "$BACKUP_DIR/data/." "${PANEL_DIR}/data/"
    fi
    
    if [[ -d "$BACKUP_DIR/logs" ]]; then
        log "恢复日志目录到: ${PANEL_DIR}/logs"
        cp -r "$BACKUP_DIR/logs/." "${PANEL_DIR}/logs/"
    fi
    
    log "数据恢复完成"
    
    # 清理临时目录
    log "清理临时文件..."
    rm -rf "$BACKUP_DIR"
}

function Set_Dir(){
    DEFAULT_DIR='/var/lib/idb'

    while true; do
        read -p "设置 idb 的目录 (默认为 ${DEFAULT_DIR}): " PANEL_DIR

        if [[ "$PANEL_DIR" == "" ]]; then
            PANEL_DIR=$DEFAULT_DIR
        fi

        # 判断目录是否合法
        if [[ ! "$PANEL_DIR" =~ ^/ ]]; then
            echo "错误：目录必须是绝对路径。"
            continue
        fi

        # 判断目录是否存在，如果不存在，则创建
        if [[ ! -d "$PANEL_DIR" ]]; then
            log "目录 ${PANEL_DIR} 不存在，正在创建..."
            mkdir -p "$PANEL_DIR"
            if [[ $? -ne 0 ]]; then
                log "创建目录 ${PANEL_DIR} 失败，请检查权限。"
                exit 1
            fi
        fi

        log "您设置的目录为：${PANEL_DIR}"
        break
    done
}

function Set_Port(){
    DEFAULT_PORT='9918'
    DEFAULT_AGENT_PORT='9919'

    while true; do
        read -p "设置 idb 端口 (默认为 ${DEFAULT_PORT}): " PANEL_PORT

        if [[ "$PANEL_PORT" == "" ]];then
            PANEL_PORT=$DEFAULT_PORT
        fi

        if ! [[ "$PANEL_PORT" =~ ^[1-9][0-9]{0,4}$ && "$PANEL_PORT" -le 65535 ]]; then
            log "错误：输入的端口号必须在 1 到 65535 之间"
            continue
        fi

        if command -v ss >/dev/null 2>&1; then
            if ss -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                log "端口${PANEL_PORT}被占用，请重新输入..."
                continue
            fi
        elif command -v netstat >/dev/null 2>&1; then
            if netstat -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                log "端口${PANEL_PORT}被占用，请重新输入..."
                continue
            fi
        fi

        log "您设置的端口为：${PANEL_PORT}"
        break
    done
}

function Set_Firewall(){
    if which firewall-cmd >/dev/null 2>&1; then
        if systemctl status firewalld | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 ${PANEL_PORT} 端口"
            firewall-cmd --zone=public --add-port=$PANEL_PORT/tcp --permanent
            firewall-cmd --zone=public --add-port=$DEFAULT_AGENT_PORT/tcp --permanent
            firewall-cmd --reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi

    if which ufw >/dev/null 2>&1; then
        if systemctl status ufw | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 ${PANEL_PORT} 端口"
            ufw allow $PANEL_PORT/tcp
            ufw allow $DEFAULT_AGENT_PORT/tcp
            ufw reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi
}

function Set_Container_Port(){
    DEFAULT_CONTAINER_PORT='9918'

    while true; do
        read -p "设置容器端口 (默认为${DEFAULT_CONTAINER_PORT}) :" CONTAINER_PORT

        if [[ "$CONTAINER_PORT" == "" ]];then
            CONTAINER_PORT=$DEFAULT_CONTAINER_PORT
        fi

        if ! [[ "$CONTAINER_PORT" =~ ^[1-9][0-9]{0,4}$ && "$CONTAINER_PORT" -le 65535 ]]; then
            log "错误：输入的端口号必须在 1 到 65535 之间"
            continue
        fi

        log "您设置的容器端口为：${CONTAINER_PORT}"
        break
    done
}

# 从 GitHub Releases 下载镜像（Docker Hub 拉取失败时的 fallback）
function Pull_Image_From_GitHub() {
    local VERSION="$1"
    local IMAGE_TAR="idb_image_${VERSION}.tar"
    local DOWNLOAD_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${IMAGE_TAR}"

    if [[ -n "$IDB_GITHUB_PROXY" ]]; then
        log "通过加速代理下载镜像: ${DOWNLOAD_URL}"
    else
        log "从 GitHub Releases 下载镜像: ${DOWNLOAD_URL}"
    fi
    curl -fL --progress-bar --retry 3 --retry-delay 5 "$DOWNLOAD_URL" -o "/tmp/${IMAGE_TAR}"
    if [[ $? -ne 0 ]]; then
        log "GitHub 镜像下载失败"
        rm -f "/tmp/${IMAGE_TAR}"
        return 1
    fi

    # 检查下载文件是否有效（至少 1MB）
    local FILE_SIZE=$(stat -c%s "/tmp/${IMAGE_TAR}" 2>/dev/null || stat -f%z "/tmp/${IMAGE_TAR}" 2>/dev/null || echo "0")
    if [[ "$FILE_SIZE" -lt 1048576 ]]; then
        log "下载的镜像文件异常（大小: ${FILE_SIZE} bytes），可能下载不完整"
        rm -f "/tmp/${IMAGE_TAR}"
        return 1
    fi

    log "加载镜像..."
    docker load -i "/tmp/${IMAGE_TAR}" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
        log "镜像加载失败"
        rm -f "/tmp/${IMAGE_TAR}"
        return 1
    fi

    # tar 中保存的是 idb:VERSION，重新标记为 compose 期望的镜像名
    local IMAGE_REPO=$(grep "^iDB_image_repo=" "${PANEL_DIR}/.env" 2>/dev/null | cut -d'=' -f2)
    IMAGE_REPO="${IMAGE_REPO:-sensdb/idb}"
    docker tag "idb:${VERSION}" "${IMAGE_REPO}:${VERSION}"

    rm -f "/tmp/${IMAGE_TAR}"
    if [[ -n "$IDB_GITHUB_PROXY" ]]; then
        log "镜像加载完成（来源: 加速代理）"
    else
        log "镜像加载完成（来源: GitHub Releases）"
    fi
}

function Install_IDB() {
    # 获取版本号
    VERSION=$(curl -s ${GITHUB_API_URL}/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

    if [[ "x${VERSION}" == "x" ]];then
        log "获取最新版本失败，请稍候重试"
        exit 1
    fi

    # 下载 .env和docker-compose.yaml 到 PANEL_DIR 中
    # .env 地址: "https://github.com/sensdata/idb/releases/download/${VERSION}/idb.env"
    # docker-compose.yaml 地址: "https://github.com/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml"
    ENV_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/idb.env"
    DOCKER_COMPOSE_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml"

    log "正在下载 .env 文件..."
    curl -fsSL "$ENV_URL" -o "${PANEL_DIR}/.env" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log ".env 文件下载失败，请检查网络连接或 URL 是否正确。"
        exit 1
    fi

    log "正在下载 docker-compose.yaml 文件..."
    curl -fsSL "$DOCKER_COMPOSE_URL" -o "${PANEL_DIR}/docker-compose.yaml" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "docker-compose.yaml 文件下载失败，请检查网络连接或 URL 是否正确。"
        exit 1
    fi

    log ".env 和 docker-compose.yaml 文件下载成功。"
    
    # 生成随机8位密码
    ADMIN_PASS=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 8)
    
    # 修改 .env 文件中的配置
    log "正在修改 .env 文件中的配置..."
    sed -i "s/^iDB_service_host_ip=.*/iDB_service_host_ip=${PUBLIC_IP}/" "${PANEL_DIR}/.env"
    sed -i "s/^iDB_service_port=.*/iDB_service_port=${PANEL_PORT}/" "${PANEL_DIR}/.env"
    sed -i "s/^iDB_service_container_port=.*/iDB_service_container_port=${CONTAINER_PORT}/" "${PANEL_DIR}/.env"
    sed -i "s/^iDB_admin_pass=.*/iDB_admin_pass=${ADMIN_PASS}/" "${PANEL_DIR}/.env"
    
    log ".env 文件内容已更新为：\n$(cat ${PANEL_DIR}/.env)"

    # 在启动容器前恢复数据
    Restore_Data

    # 预拉取镜像
    log "拉取镜像..."
    cd "$PANEL_DIR" || { log "无法进入目录 $PANEL_DIR"; exit 1; }
    if ! docker compose pull; then
        if [[ -n "$IDB_GITHUB_PROXY" ]]; then
            log "Docker Hub 拉取失败，尝试通过加速代理下载镜像..."
        else
            log "Docker Hub 拉取失败，尝试从 GitHub Releases 下载镜像..."
        fi
        if ! Pull_Image_From_GitHub "${VERSION}"; then
            # 如果是直连 GitHub 失败，再尝试通过代理下载
            if [[ -z "$IDB_GITHUB_PROXY" ]]; then
                log "GitHub 直连下载失败，尝试通过加速代理下载..."
                export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
                GITHUB_RELEASES_URL="${IDB_GITHUB_PROXY}/github-releases"
                if ! Pull_Image_From_GitHub "${VERSION}"; then
                    log "所有镜像源均不可用，安装失败"
                    exit 1
                fi
            else
                log "所有镜像源均不可用，安装失败"
                exit 1
            fi
        fi
    fi
    log "镜像拉取完成"

    # 启动新容器
    log "正在启动 IDB..."
    cd "$PANEL_DIR" || { log "无法进入目录 $PANEL_DIR"; exit 1; }
    
    docker compose up -d 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "启动 IDB 失败，请检查 docker-compose 配置。"
        exit 1
    fi

    # 检查容器是否真的起来了
    if ! docker ps -a --format '{{.Names}}' | grep -q '^idb$'; then
        log "IDB 容器未成功创建，请检查日志：docker compose logs"
        exit 1
    fi

    # 进一步确认容器是否是 running 状态
    if ! docker inspect -f '{{.State.Running}}' idb 2>/dev/null | grep -q true; then
        log "IDB 容器未处于运行状态，请检查日志：docker compose logs idb"
        exit 1
    fi

    log "IDB 启动成功！"

    # 从 idb 容器的 /var/lib/idb/agent 目录下，拷贝 idb-agent_${VERSION}.tar.gz 至当前目录下的 agent目录
    log "正在拷贝 idb-agent 文件..."
    # 清理 agent 目录
    rm -rf "${CURRENT_DIR}/agent"
    # 创建 agent 目录
    mkdir -p "${CURRENT_DIR}/agent"  
    docker cp "idb:/var/lib/idb/agent/idb-agent.tar.gz" "${CURRENT_DIR}/agent/" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "拷贝 idb-agent 文件失败，请检查容器是否存在。"
        exit 1
    fi

    log "正在解压 idb-agent 文件..."
    tar -xzvf "${CURRENT_DIR}/agent/idb-agent.tar.gz" -C "${CURRENT_DIR}/agent/" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "解压 idb-agent 文件失败。"
        exit 1
    fi

    log "idb-agent 文件解压成功。"

    # 进入 agent 目录并执行 install-agent.sh
    log "正在执行 install-agent.sh..."
    cd "${CURRENT_DIR}/agent" || { log "无法进入目录 ${CURRENT_DIR}/agent"; exit 1; }
    # 如果旧版本配置存在，则覆盖当前版本idb-agent.conf
    if [[ -f "/etc/idb-agent/idb-agent.conf" ]]; then
        log "旧版本配置存在，则覆盖当前版本配置"
        cp "/etc/idb-agent/idb-agent.conf" "idb-agent.conf"
    fi
    bash ./install-agent.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "执行 install-agent.sh 失败，请检查脚本内容。"
        exit 1
    fi

    log "install-agent.sh 执行成功。"
}

function Get_Ip(){
    # 优先获取默认路由对应的源IP（仅IPv4）
    LOCAL_IP=$(ip -4 route get 8.8.8.8 2>/dev/null | grep -oP 'src \K[^ ]+')
    if [[ -z "$LOCAL_IP" ]]; then
        # 备选方案：获取默认网卡的IPv4地址
        default_interface=$(ip -4 route 2>/dev/null | grep '^default' | awk '{print $5}' | head -n1)
        if [[ -n "$default_interface" ]]; then
            LOCAL_IP=$(ip -4 addr show $default_interface 2>/dev/null | grep -oP 'inet \K[\d.]+' | head -n1)
        fi
    fi
    
    # 如果上述方法都失败，使用默认值
    if [[ -z "$LOCAL_IP" ]]; then
        LOCAL_IP="127.0.0.1"
    fi

    # 获取公网IP（优先IPv4）
    PUBLIC_IP=$(curl -s -4 https://api.ipify.org 2>/dev/null)
    if [[ -z "$PUBLIC_IP" ]]; then
        # 如果IPv4获取失败，尝试其他API
        PUBLIC_IP=$(curl -s -4 https://api64.ipify.org 2>/dev/null)
    fi
    if [[ -z "$PUBLIC_IP" ]]; then
        PUBLIC_IP="N/A"
    fi
    # 如果获取到的是IPv6地址（包含冒号），则设为N/A
    if echo "$PUBLIC_IP" | grep -q ":"; then
        PUBLIC_IP="N/A"
    fi
}

function Show_Result(){
    log ""
    log "=================感谢您的耐心等待，安装已经完成=================="
    log ""
    log "请用浏览器访问面板:"
    log "外网地址: http://${PUBLIC_IP}:${PANEL_PORT}/idb"
    log "内网地址: http://${LOCAL_IP}:${PANEL_PORT}/idb"
    log "初始用户: admin"
    log "初始密码: ${ADMIN_PASS}"
    log ""
    log "项目官网: https://idb.net"
    log "项目文档: https://idb.net/docs"
    log "代码仓库: https://github.com/sensdata/idb"
    log ""
    log "如果使用的是云服务器，请至安全组开放 $PANEL_PORT 端口"
    log ""
    log "为了您的服务器安全，在您离开此界面后您将无法再看到您的密码，请务必牢记您的密码。"
    log ""
    log "================================================================"
}

function main(){
    Check_Root
    Install_Docker
    Install_Compose
    Check_Installation
    Set_Dir
    Set_Port
    Set_Firewall
    Set_Container_Port
    Get_Ip
    Install_IDB
    Show_Result
}
main
