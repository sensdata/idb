#!/bin/bash

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/upgrade.log
}

# 加速代理支持
# 用法: IDB_GITHUB_PROXY=https://dl.idb.net bash upgrade.sh
# 或在 .env 中配置: IDB_GITHUB_PROXY=https://dl.idb.net
# 如果未指定代理，自动检测 GitHub 连通性，不通则使用 dl.idb.net
IDB_DEFAULT_PROXY="https://dl.idb.net"

# 从 .env 读取代理配置（如果环境变量未设置）
if [[ -z "$IDB_GITHUB_PROXY" ]]; then
    _PANEL_DIR=$(docker inspect --format '{{ index .Config.Labels "com.docker.compose.project.working_dir" }}' idb 2>/dev/null)
    _PANEL_DIR="${_PANEL_DIR:-/var/lib/idb}"
    if [[ -f "${_PANEL_DIR}/.env" ]]; then
        IDB_GITHUB_PROXY=$(grep "^IDB_GITHUB_PROXY=" "${_PANEL_DIR}/.env" 2>/dev/null | cut -d'=' -f2)
    fi
    unset _PANEL_DIR
fi

function Auto_Detect_Proxy() {
    if [[ -n "$IDB_GITHUB_PROXY" ]]; then
        log "使用指定代理: ${IDB_GITHUB_PROXY}"
        return
    fi

    log "检测 GitHub 连通性..."
    local github_ok=false
    if curl -s --connect-timeout 5 --max-time 10 -o /dev/null -w "%{http_code}" https://api.github.com/repos/sensdata/idb/releases/latest 2>/dev/null | grep -q "200"; then
        github_ok=true
    fi

    if [[ "$github_ok" == "true" ]]; then
        log "GitHub 直连正常"
    else
        log "GitHub 连接超时或不可用，自动切换到加速代理: ${IDB_DEFAULT_PROXY}"
        export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
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

# 备份数据（PANEL_DIR 由调用方提前设置）
function Backup_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    log "清理临时目录..."
    rm -rf "$BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    
    # 检查容器是否存在并运行
    if docker ps -q -f name="^idb$" > /dev/null 2>&1; then
        log "停止 IDB 容器..."
        docker stop -t 30 idb || docker kill idb
    fi
    
    if docker ps -a -q -f name=idb >/dev/null 2>&1; then
        # 分别获取三个挂载的宿主机真实路径
        local DATA_DIR=$(docker inspect --format '{{ range .Mounts }}{{ if eq .Destination "/var/lib/idb/data" }}{{ .Source }}{{ end }}{{ end }}' idb 2>/dev/null)
        local LOG_DIR=$(docker inspect --format '{{ range .Mounts }}{{ if eq .Destination "/var/log/idb" }}{{ .Source }}{{ end }}{{ end }}' idb 2>/dev/null)
        local CONF_DIR=$(docker inspect --format '{{ range .Mounts }}{{ if eq .Destination "/etc/idb" }}{{ .Source }}{{ end }}{{ end }}' idb 2>/dev/null)
        
        # 写入元数据，供回滚时使用（不依赖容器状态）
        cat > "$BACKUP_DIR/paths.meta" <<EOF
PANEL_DIR=${PANEL_DIR}
DATA_DIR=${DATA_DIR}
LOG_DIR=${LOG_DIR}
CONF_DIR=${CONF_DIR}
EOF
        
        # 备份 .env 和 docker-compose.yaml
        cp "${PANEL_DIR}/.env" "$BACKUP_DIR/.env"
        cp "${PANEL_DIR}/docker-compose.yaml" "$BACKUP_DIR/docker-compose.yaml"
        
        # 备份数据目录
        if [[ -n "$DATA_DIR" && -d "$DATA_DIR" ]]; then
            log "备份数据目录: ${DATA_DIR}"
            cp -r "$DATA_DIR" "$BACKUP_DIR/data"
        fi
        
        # 备份日志目录
        if [[ -n "$LOG_DIR" && -d "$LOG_DIR" ]]; then
            log "备份日志目录: ${LOG_DIR}"
            cp -r "$LOG_DIR" "$BACKUP_DIR/logs"
        fi
        
        # 备份配置目录
        if [[ -n "$CONF_DIR" && -d "$CONF_DIR" ]]; then
            log "备份配置目录: ${CONF_DIR}"
            cp -r "$CONF_DIR" "$BACKUP_DIR/conf"
        fi
        
        log "删除旧容器..."
        if ! docker rm idb; then
            log "删除容器失败，尝试强制删除..."
            docker rm -f idb
        fi
        
        return 0
    fi
    
    return 1
}

# 恢复数据（仅从备份元数据读取路径，不依赖容器状态）
function Restore_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log "未找到备份数据，升级失败"
        exit 1
    fi
    
    # 从备份元数据读取路径
    if [[ -f "$BACKUP_DIR/paths.meta" ]]; then
        source "$BACKUP_DIR/paths.meta"
    else
        log "未找到路径元数据，使用默认路径"
        PANEL_DIR="/var/lib/idb"
        DATA_DIR="/var/lib/idb/data"
        LOG_DIR="/var/log/idb"
        CONF_DIR="/etc/idb"
    fi
    
    log "开始恢复数据..."
    log "  项目目录: ${PANEL_DIR}"
    log "  数据目录: ${DATA_DIR}"
    log "  日志目录: ${LOG_DIR}"
    log "  配置目录: ${CONF_DIR}"
    
    # 恢复 .env 和 docker-compose.yaml
    cp "$BACKUP_DIR/.env" "${PANEL_DIR}/.env"
    cp "$BACKUP_DIR/docker-compose.yaml" "${PANEL_DIR}/docker-compose.yaml"
    
    # 恢复数据目录
    if [[ -d "$BACKUP_DIR/data" && -n "$DATA_DIR" ]]; then
        log "恢复数据目录到: ${DATA_DIR}"
        mkdir -p "$DATA_DIR"
        cp -r "$BACKUP_DIR/data/." "${DATA_DIR}/"
    fi
    
    # 恢复日志目录
    if [[ -d "$BACKUP_DIR/logs" && -n "$LOG_DIR" ]]; then
        log "恢复日志目录到: ${LOG_DIR}"
        mkdir -p "$LOG_DIR"
        cp -r "$BACKUP_DIR/logs/." "${LOG_DIR}/"
    fi
    
    # 恢复配置目录
    if [[ -d "$BACKUP_DIR/conf" && -n "$CONF_DIR" ]]; then
        log "恢复配置目录到: ${CONF_DIR}"
        mkdir -p "$CONF_DIR"
        cp -r "$BACKUP_DIR/conf/." "${CONF_DIR}/"
    fi
    
    log "数据恢复完成"
    
    # 清理临时目录
    rm -rf "$BACKUP_DIR"
}

# 从 GitHub Releases 下载镜像（Docker Hub 拉取失败时的 fallback）
function Pull_Image_From_GitHub() {
    local VERSION="$1"
    local IMAGE_TAR="idb_image_${VERSION}.tar"
    local DOWNLOAD_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${IMAGE_TAR}"

    log "从 GitHub Releases 下载镜像: ${DOWNLOAD_URL}"
    curl -fSL "$DOWNLOAD_URL" -o "/tmp/${IMAGE_TAR}"
    if [[ $? -ne 0 ]]; then
        log "GitHub 镜像下载失败"
        rm -f "/tmp/${IMAGE_TAR}"
        return 1
    fi

    log "加载镜像..."
    docker load -i "/tmp/${IMAGE_TAR}"
    if [[ $? -ne 0 ]]; then
        log "镜像加载失败"
        rm -f "/tmp/${IMAGE_TAR}"
        return 1
    fi

    # tar 中保存的是 idb:VERSION，重新标记为 compose 期望的镜像名
    local IMAGE_REPO=$(grep "^iDB_image_repo=" "${PANEL_DIR}/.env.new" 2>/dev/null | cut -d'=' -f2)
    [[ -z "$IMAGE_REPO" ]] && IMAGE_REPO=$(grep "^iDB_image_repo=" "${PANEL_DIR}/.env" 2>/dev/null | cut -d'=' -f2)
    IMAGE_REPO="${IMAGE_REPO:-sensdb/idb}"
    docker tag "idb:${VERSION}" "${IMAGE_REPO}:${VERSION}"

    rm -f "/tmp/${IMAGE_TAR}"
    log "GitHub 镜像加载完成"
}

# 升级 IDB
function Upgrade_IDB() {
    # 优先使用传入的版本号，否则获取最新版本
    VERSION="${1}"
    if [[ -z "$VERSION" ]]; then
        VERSION=$(curl -s ${GITHUB_API_URL}/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
        if [[ -z "$VERSION" ]]; then
            log "获取最新版本失败"
            exit 1
        fi
    fi
    
    log "开始升级到版本 ${VERSION}..."
    
    # 通过 compose label 获取项目目录（在停容器之前获取）
    PANEL_DIR=$(docker inspect --format '{{ index .Config.Labels "com.docker.compose.project.working_dir" }}' idb 2>/dev/null)
    if [[ -z "$PANEL_DIR" ]]; then
        PANEL_DIR="/var/lib/idb"
    fi
    
    # 下载新版本配置文件（在停容器之前下载到临时位置）
    ENV_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/idb.env"
    DOCKER_COMPOSE_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml"
    
    log "下载新版本配置文件..."
    curl -fsSL "$ENV_URL" -o "${PANEL_DIR}/.env.new"
    if [[ $? -ne 0 ]]; then
        log "下载 .env 失败，中止升级"
        rm -f "${PANEL_DIR}/.env.new"
        exit 1
    fi
    curl -fsSL "$DOCKER_COMPOSE_URL" -o "${PANEL_DIR}/docker-compose.yaml.new"
    if [[ $? -ne 0 ]]; then
        log "下载 docker-compose.yaml 失败，中止升级"
        rm -f "${PANEL_DIR}/.env.new" "${PANEL_DIR}/docker-compose.yaml.new"
        exit 1
    fi
    
    # 预拉取新镜像（旧容器仍在运行，服务不中断）
    # 使用 .env.new 中的镜像名来拉取，不修改现有 .env
    log "预拉取新版本镜像..."
    local NEW_IMAGE_REPO=$(grep "^iDB_image_repo=" "${PANEL_DIR}/.env.new" 2>/dev/null | cut -d'=' -f2)
    NEW_IMAGE_REPO="${NEW_IMAGE_REPO:-sensdb/idb}"
    if ! docker pull "${NEW_IMAGE_REPO}:${VERSION}" 2>/dev/null; then
        log "Docker Hub 拉取失败，尝试从 GitHub Releases 下载镜像..."
        if ! Pull_Image_From_GitHub "${VERSION}"; then
            log "所有镜像源均不可用，中止升级（旧版本未受影响）"
            rm -f "${PANEL_DIR}/.env.new" "${PANEL_DIR}/docker-compose.yaml.new"
            exit 1
        fi
    fi
    log "镜像拉取完成"

    # 镜像已准备好，现在才更新配置文件
    if [[ -f "${PANEL_DIR}/.env.new" ]]; then
        # 保存用户自定义的配置（包括路径配置）
        local USER_HOST=$(grep "^iDB_service_host_ip=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_PORT=$(grep "^iDB_service_port=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_CONTAINER_PORT=$(grep "^iDB_service_container_port=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_DATA_PATH=$(grep "^iDB_service_data_path=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_LOG_PATH=$(grep "^iDB_service_log_path=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_CONF_PATH=$(grep "^iDB_service_conf_path=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_NETWORK_MODE=$(grep "^iDB_service_network_mode=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_IMAGE_REPO=$(grep "^iDB_image_repo=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        
        # 使用新的配置文件
        mv "${PANEL_DIR}/.env.new" "${PANEL_DIR}/.env"
        
        # 恢复用户的自定义配置
        if [[ -n "$USER_HOST" ]]; then
            sed -i "s/^iDB_service_host_ip=.*/iDB_service_host_ip=${USER_HOST}/" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_PORT" ]]; then
            sed -i "s/^iDB_service_port=.*/iDB_service_port=${USER_PORT}/" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_CONTAINER_PORT" ]]; then
            sed -i "s/^iDB_service_container_port=.*/iDB_service_container_port=${USER_CONTAINER_PORT}/" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_DATA_PATH" ]]; then
            sed -i "s|^iDB_service_data_path=.*|iDB_service_data_path=${USER_DATA_PATH}|" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_LOG_PATH" ]]; then
            sed -i "s|^iDB_service_log_path=.*|iDB_service_log_path=${USER_LOG_PATH}|" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_CONF_PATH" ]]; then
            sed -i "s|^iDB_service_conf_path=.*|iDB_service_conf_path=${USER_CONF_PATH}|" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_NETWORK_MODE" ]]; then
            sed -i "s|^iDB_service_network_mode=.*|iDB_service_network_mode=${USER_NETWORK_MODE}|" "${PANEL_DIR}/.env"
        fi
        if [[ -n "$USER_IMAGE_REPO" ]]; then
            sed -i "s|^iDB_image_repo=.*|iDB_image_repo=${USER_IMAGE_REPO}|" "${PANEL_DIR}/.env"
        fi
    fi

    # docker-compose.yaml
    if [[ -f "${PANEL_DIR}/docker-compose.yaml.new" ]]; then
        mv "${PANEL_DIR}/docker-compose.yaml.new" "${PANEL_DIR}/docker-compose.yaml"
    fi
    
    # 所有文件和镜像都已就绪，现在才停止旧容器并备份
    Backup_Data
    
    # 启动新版本容器（镜像已预拉取，几乎瞬时启动）
    cd "${PANEL_DIR}" || exit 1
    docker compose up -d
    
    if [[ $? -ne 0 ]]; then
        log "启动新版本失败，开始回滚..."
        Restore_Data
        docker compose up -d
        exit 1
    fi
    
    log "升级完成，版本：${VERSION}"
}

# 主函数
function main() {
    log "======================= 开始升级 ======================="
    Upgrade_IDB "$1"
    log "======================= 升级完成 ======================="
}

main "$1"