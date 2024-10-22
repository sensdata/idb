#!/bin/bash
set -e

DEFAULT_PORT=9918
PORT=${PORT:-$DEFAULT_PORT}
CONFIG_FILE=/etc/idb/idb.conf
LOG_FILE=/var/log/idb/idb.log
IDB_EXECUTABLE="$1"

if [ -z "$IDB_EXECUTABLE" ]; then
    echo "Error: IDB executable path not provided"
    exit 1
fi

echo "Starting configuration with PORT=$PORT"

# 修改或添加端口配置
if grep -q "^port=" "$CONFIG_FILE"; then
    sed -i "s/^port=.*/port=$PORT/" "$CONFIG_FILE"
else
    echo "port=$PORT" >> "$CONFIG_FILE"
fi

echo "Configured idb.conf with port=$PORT"

# 设置文件描述符限制
ulimit -n 1048576
ulimit -u 1048576
ulimit -c 1048576

# 启动应用
echo "Starting IDB service..."
exec "$IDB_EXECUTABLE" start >> "$LOG_FILE" 2>&1