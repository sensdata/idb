#!/bin/bash
set -e

DEFAULT_HOST=127.0.0.1
DEFAULT_PORT=9918
HOST=${HOST:-$DEFAULT_HOST}
PORT=${PORT:-$DEFAULT_PORT}
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

log "Starting configure idb.conf"

# 修改或添加相关配置
if grep -q "^host=" "$CONFIG_FILE"; then
    sed -i "s/^host=.*/host=$HOST/" "$CONFIG_FILE"
    log "更新配置: host=$HOST"
else
    echo "host=$HOST" >> "$CONFIG_FILE"
    log "新增配置: host=$HOST"
fi

if grep -q "^port=" "$CONFIG_FILE"; then
    sed -i "s/^port=.*/port=$PORT/" "$CONFIG_FILE"
    log "更新配置: port=$PORT"
else
    echo "port=$PORT" >> "$CONFIG_FILE"
    log "新增配置: port=$PORT"
fi

log "配置文件更新完成，当前配置内容：\n$(cat "$CONFIG_FILE")"

# 设置文件描述符限制
ulimit -n 1048576
ulimit -u 1048576
ulimit -c 1048576

# 启动应用
log "Starting IDB service..."
exec "$IDB_EXECUTABLE" start