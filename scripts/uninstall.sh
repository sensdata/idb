#!/bin/bash

set -euo pipefail

echo "警告：此操作将完全删除 IDB 所有组件和数据。"
read -r -p "确认要继续卸载吗？(y/N) " reply
if [[ ! "$reply" =~ ^[Yy]$ ]]; then
    echo "卸载已取消"
    exit 1
fi

echo "开始卸载..."

if command -v systemctl >/dev/null 2>&1; then
    echo "停止并禁用 center 服务..."
    sudo systemctl stop idb.service 2>/dev/null || true
    sudo systemctl disable idb.service 2>/dev/null || true

    echo "停止并禁用 agent 服务..."
    sudo systemctl stop idb-agent.service 2>/dev/null || true
    sudo systemctl disable idb-agent.service 2>/dev/null || true
    sudo systemctl daemon-reload 2>/dev/null || true
fi

if command -v docker >/dev/null 2>&1; then
    echo "清理遗留 Docker 容器..."
    docker rm -f idb 2>/dev/null || true
fi

echo "清理 center 目录..."
for dir in "/var/lib/idb" "/var/log/idb" "/etc/idb" "/run/idb"; do
    sudo rm -rf "$dir" 2>/dev/null || true
done

echo "清理 agent 目录..."
for dir in "/var/lib/idb-agent" "/var/log/idb-agent" "/etc/idb-agent" "/run/idb-agent"; do
    sudo rm -rf "$dir" 2>/dev/null || true
done

echo "清理 systemd 服务文件..."
sudo rm -f /etc/systemd/system/idb.service /etc/systemd/system/idb-agent.service 2>/dev/null || true
sudo rm -f /usr/local/bin/idb /usr/local/bin/idb-agent 2>/dev/null || true
sudo systemctl daemon-reload 2>/dev/null || true

echo "卸载完成。"
