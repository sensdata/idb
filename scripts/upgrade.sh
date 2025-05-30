#!/bin/bash

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/upgrade.log
}

# 备份数据
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
        # 保存当前环境变量值和容器配置
        docker inspect idb > "$BACKUP_DIR/container_info.json"
        
        # 获取当前的安装目录
        PANEL_DIR=$(docker inspect idb --format '{{ range .Mounts }}{{ if eq .Destination "/var/lib/idb" }}{{ .Source }}{{ end }}{{ end }}')
        if [[ -z "$PANEL_DIR" ]]; then
            PANEL_DIR="/var/lib/idb"  # 使用默认目录
        fi
        
        # 备份关键文件
        cp "${PANEL_DIR}/.env" "$BACKUP_DIR/.env"
        cp "${PANEL_DIR}/docker-compose.yaml" "$BACKUP_DIR/docker-compose.yaml"
        
        # 备份数据目录
        if [[ -d "${PANEL_DIR}/data" ]]; then
            log "备份数据目录..."
            cp -r "${PANEL_DIR}/data" "$BACKUP_DIR/"
        fi
        
        # 备份日志目录
        if [[ -d "${PANEL_DIR}/logs" ]]; then
            log "备份日志目录..."
            cp -r "${PANEL_DIR}/logs" "$BACKUP_DIR/"
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

# 恢复数据
function Restore_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log "未找到备份数据，升级失败"
        exit 1
    fi
    
    # 获取安装目录
    PANEL_DIR=$(jq -r '.[] | select(.Name=="/idb") | .Mounts[] | select(.Destination=="/var/lib/idb") | .Source' "$BACKUP_DIR/container_info.json" 2>/dev/null)
    if [[ -z "$PANEL_DIR" ]]; then
        PANEL_DIR="/var/lib/idb"
    fi
    
    log "开始恢复数据到 ${PANEL_DIR}..."
    
    # 恢复 .env 和 docker-compose.yaml
    cp "$BACKUP_DIR/.env" "${PANEL_DIR}/.env"
    cp "$BACKUP_DIR/docker-compose.yaml" "${PANEL_DIR}/docker-compose.yaml"
    
    # 恢复数据目录
    if [[ -d "$BACKUP_DIR/data" ]]; then
        log "恢复数据目录..."
        cp -r "$BACKUP_DIR/data/." "${PANEL_DIR}/data/"
    fi
    
    # 恢复日志目录
    if [[ -d "$BACKUP_DIR/logs" ]]; then
        log "恢复日志目录..."
        cp -r "$BACKUP_DIR/logs/." "${PANEL_DIR}/logs/"
    fi
    
    log "数据恢复完成"
    
    # 清理临时目录
    rm -rf "$BACKUP_DIR"
}

# 升级 IDB
function Upgrade_IDB() {
    # 优先使用传入的版本号，否则获取最新版本
    VERSION="${1}"
    if [[ -z "$VERSION" ]]; then
        VERSION=$(curl -s https://static.sensdata.com/idb/release/latest)
        if [[ -z "$VERSION" ]]; then
            log "获取最新版本失败"
            exit 1
        fi
    fi
    
    log "开始升级到版本 ${VERSION}..."
    
    # 备份当前数据
    Backup_Data
    
    # 下载新版本配置文件
    ENV_URL="https://static.sensdata.com/idb/release/${VERSION}/.env"
    DOCKER_COMPOSE_URL="https://static.sensdata.com/idb/release/${VERSION}/docker-compose.yaml"
    
    log "下载新版本配置文件..."
    curl -fsSL "$ENV_URL" -o "${PANEL_DIR}/.env.new"
    curl -fsSL "$DOCKER_COMPOSE_URL" -o "${PANEL_DIR}/docker-compose.yaml.new"
    
    # 合并配置文件（保留原有的自定义配置）
    if [[ -f "${PANEL_DIR}/.env.new" ]]; then
        # 保存用户自定义的配置
        local USER_HOST=$(grep "^iDB_service_host_ip=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_PORT=$(grep "^iDB_service_port=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        local USER_CONTAINER_PORT=$(grep "^iDB_service_container_port=" "${PANEL_DIR}/.env" | cut -d'=' -f2)
        
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
    fi

    # docker-compose.yaml
    if [[ -f "${PANEL_DIR}/docker-compose.yaml.new" ]]; then
        mv "${PANEL_DIR}/docker-compose.yaml.new" "${PANEL_DIR}/docker-compose.yaml"
    fi
    
    # 启动新版本容器
    cd "${PANEL_DIR}" || exit 1
    docker-compose up -d
    
    if [[ $? -ne 0 ]]; then
        log "启动新版本失败，开始回滚..."
        Restore_Data
        docker-compose up -d
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