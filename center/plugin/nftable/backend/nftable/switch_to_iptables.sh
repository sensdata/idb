#!/bin/bash
set -euo pipefail

# 停止并禁用 nftables 服务
disable_nftables() {
    if systemctl is-active --quiet nftables; then
        echo "Stopping and disabling nftables..."
        systemctl stop nftables
        systemctl disable nftables
    else
        echo "nftables is already stopped and disabled."
    fi
}

# 启用并启动 iptables 服务（兼容 RHEL 和 Debian 系）
enable_iptables() {
    echo "Enabling and starting iptables..."

    if systemctl list-unit-files | grep -q "^iptables.service"; then
        systemctl enable iptables
        systemctl start iptables
    elif systemctl list-unit-files | grep -q "^netfilter-persistent.service"; then
        systemctl enable netfilter-persistent
        systemctl start netfilter-persistent
    else
        echo "No iptables systemd unit found."
        echo "👉 RHEL/CentOS: yum install -y iptables-services"
        echo "👉 Debian/Ubuntu: apt install -y iptables-persistent"
        exit 1
    fi
}

# 添加规则以放通必要端口
add_iptables_rules() {
    echo "Adding iptables rules for ports 22, 9918, and 9919..."
    for port in 22 9918 9919; do
        if ! iptables -C INPUT -p tcp --dport $port -j ACCEPT &>/dev/null; then
            iptables -A INPUT -p tcp --dport $port -j ACCEPT
        else
            echo "iptables rule for port $port already exists. Skipping."
        fi
    done
}

main() {
    disable_nftables
    enable_iptables
    add_iptables_rules
    echo "Success: switched to iptables"
}

main
