#!/bin/bash

# 停止并禁用 nftables 服务
disable_nftables() {
    if systemctl is-active --quiet nftables; then
        echo "Stopping and disabling nftables..."
        systemctl stop nftables
        systemctl disable nftables
        if [[ $? -ne 0 ]]; then
            echo "Failed to stop or disable nftables."
            exit 1
        fi
    else
        echo "nftables is already stopped and disabled."
    fi
}

# 启用并启动 iptables 服务
enable_iptables() {
    echo "Enabling and starting iptables..."
    systemctl enable iptables
    systemctl start iptables
    if [[ $? -ne 0 ]]; then
        echo "Failed to enable or start iptables."
        exit 1
    fi
}

# 主逻辑
main() {
    disable_nftables
    enable_iptables
    echo "Success"
}

main
