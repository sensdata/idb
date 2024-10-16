#!/bin/bash
set -e

DEFAULT_PORT=25900
PORT=${PORT:-$DEFAULT_PORT}
CONFIG_FILE=/etc/idb/idb.conf

echo "Starting configuration with PORT=$PORT"

# 修改或添加端口配置
if grep -q "^port=" "$CONFIG_FILE"; then
    sed -i "s/^port=.*/port=$PORT/" "$CONFIG_FILE"
else
    echo "port=$PORT" >> "$CONFIG_FILE"
fi

echo "Configured idb.conf with port=$PORT"

# 启动应用
exec "$@"