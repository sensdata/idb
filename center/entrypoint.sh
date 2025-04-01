#!/bin/bash
set -e

DEFAULT_HOST=127.0.0.1
DEFAULT_PORT=9918
HOST=${HOST:-$DEFAULT_HOST}
PORT=${PORT:-$DEFAULT_PORT}
LATEST=https://static.sensdata.com/idb/release/latest
CONFIG_FILE=/etc/idb/idb.conf
LOG_FILE=/var/log/idb/idb.log
IDB_EXECUTABLE="$1"

if [ -z "$IDB_EXECUTABLE" ]; then
    echo "Error: IDB executable path not provided"
    exit 1
fi

# 创建日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 修改或添加相关配置
update_config() {
    local key=$1
    local value=$2
    
    if grep -q "^${key}=" "$CONFIG_FILE"; then
        if ! sed -i "s/^${key}=.*/${key}=${value}/" "$CONFIG_FILE"; then
            log "错误：更新配置 ${key}=${value} 失败"
            return 1
        fi
        log "更新配置: ${key}=${value}"
    else
        if ! echo "${key}=${value}" >> "$CONFIG_FILE"; then
            log "错误：新增配置 ${key}=${value} 失败"
            return 1
        fi
        log "新增配置: ${key}=${value}"
    fi
    return 0
}

log "Starting configure idb.conf"

# 修改或添加相关配置
if ! update_config "host" "$HOST" || \
   ! update_config "port" "$PORT" || \
   ! update_config "latest" "$LATEST"; then
    log "配置文件更新失败"
    exit 1
fi

# 显示更新后的配置
if ! cat "$CONFIG_FILE"; then
    log "读取配置文件失败"
    exit 1
fi

log "配置文件更新完成，当前配置内容：\n$(cat "$CONFIG_FILE")"

# 设置文件描述符限制
ulimit -n 1048576
ulimit -u 1048576
ulimit -c 1048576

# 启动应用
log "Starting IDB service..."
exec "$IDB_EXECUTABLE" start