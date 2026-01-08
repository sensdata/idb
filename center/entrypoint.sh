#!/bin/bash
set -e

DEFAULT_HOST=127.0.0.1
DEFAULT_PORT=9918

HOST=${HOST:-$DEFAULT_HOST}
PORT=${PORT:-$DEFAULT_PORT}
ADMIN_PASS=${PASSWORD}

LATEST=https://static.sensdata.com/idb/release/latest
CONFIG_FILE=/etc/idb/idb.conf
LOG_FILE=/var/log/idb/idb.log
IDB_EXECUTABLE="$1"

# 创建日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 确保必要目录存在
ensure_directories() {
    REQUIRED_DIRS=(
        /etc/idb
        /var/log/idb
        /var/lib/idb
        /var/lib/idb/data
        /var/lib/idb/agent
        /run/idb
    )

    for dir in "${REQUIRED_DIRS[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log "创建目录: $dir"
        fi
    done

    # 设置 /run/idb 权限
    chmod 755 /run/idb
}

# 修改或添加相关配置
update_config() {
    local key=$1
    local value=$2
    
    # 转义 key 中的特殊字符
    local escaped_key=$(printf '%s\n' "$key" | sed 's/[][\.*^$/]/\\&/g')
    
    # 对于包含斜杠的值，使用 # 作为 sed 的分隔符
    if grep -q "^${escaped_key}=" "$CONFIG_FILE"; then
        if ! sed -i "s#^${escaped_key}=.*#${key}=${value}#" "$CONFIG_FILE"; then
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

# 启动逻辑
main() {
    if [ -z "$IDB_EXECUTABLE" ]; then
        echo "Error: IDB executable path not provided"
        exit 1
    fi

    log "ensure directories"
    ensure_directories

    log "configure $CONFIG_FILE"
    update_config "host" "$HOST"
    update_config "port" "$PORT"
    update_config "latest" "$LATEST"

    log "配置文件内容："
    cat "$CONFIG_FILE" || {
        log "读取配置文件失败"
        exit 1
    }

    # 设置资源限制（添加错误处理，允许非特权环境下继续运行）
    ulimit -n 1048576 || log "警告: 无法设置nofile限制, 将使用系统默认值"
    ulimit -u 1048576 || log "警告: 无法设置nproc限制, 将使用系统默认值"
    ulimit -c 1048576 || log "警告: 无法设置core文件大小限制, 将使用系统默认值"

    log "Starting IDB service..."
    exec "$IDB_EXECUTABLE" start
}

main