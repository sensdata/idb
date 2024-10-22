#!/bin/bash

# 确保脚本以 root 权限运行
if [ "$(id -u)" != "0" ]; then
   echo "此脚本必须以 root 权限运行" 1>&2
   exit 1
fi

# 创建必要的目录
mkdir -p /etc/idb-agent /var/log/idb-agent /run/idb-agent /var/lib/idb-agent /var/lib/idb-agent/data 

# 复制文件到正确的位置
cp ./idb-agent /var/lib/idb-agent/idb-agent
cp ./idb-agent.conf /etc/idb-agent/idb-agent.conf
cp ./idb-agent.service /etc/systemd/system/idb-agent.service

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