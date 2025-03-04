#!/bin/bash

# 确保脚本以 root 权限运行
if [ "$(id -u)" != "0" ]; then
   echo "此脚本必须以 root 权限运行" 1>&2
   exit 1
fi

# 检查并停止已存在的服务
if systemctl is-active --quiet idb-agent.service; then
    echo "停止现有 idb-agent 服务..."
    systemctl stop idb-agent.service
fi

# 如果服务已启用，先禁用它
if systemctl is-enabled --quiet idb-agent.service; then
    echo "禁用现有 idb-agent 服务..."
    systemctl disable idb-agent.service
fi

# 创建必要的目录
mkdir -p /etc/idb-agent /var/log/idb-agent /run/idb-agent /var/lib/idb-agent /var/lib/idb-agent/data 

# 备份现有配置（如果存在）
if [ -f "/etc/idb-agent/idb-agent.conf" ]; then
    echo "备份现有配置文件..."
    cp -f /etc/idb-agent/idb-agent.conf /etc/idb-agent/idb-agent.conf.bak
fi

# 复制文件到正确的位置，使用 -f 强制覆盖
cp -f ./idb-agent /var/lib/idb-agent/idb-agent
cp -f ./idb-agent.conf /etc/idb-agent/idb-agent.conf
cp -f ./idb-agent.service /etc/systemd/system/idb-agent.service

# 设置正确的权限
chmod +x /var/lib/idb-agent/idb-agent
touch /var/log/idb-agent/idb-agent.log

# 重新加载 systemd 配置
systemctl daemon-reload

# 设置 idb-agent 服务开机自启
systemctl enable idb-agent.service

# 启动 idb-agent 服务
systemctl start idb-agent.service

echo "idb-agent 安装完成并已启动"