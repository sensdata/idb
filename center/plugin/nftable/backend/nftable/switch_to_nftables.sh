#!/bin/bash

# 停止并禁用 iptables 服务
disable_iptables() {
    if systemctl is-active --quiet iptables; then
        echo "Stopping and disabling iptables..."
        systemctl stop iptables
        systemctl disable iptables
        if [[ $? -ne 0 ]]; then
            echo "Failed to stop or disable iptables."
            exit 1
        fi
    else
        echo "iptables is already stopped and disabled."
    fi
}

# 启用并启动 nftables 服务
enable_nftables() {
    echo "Enabling and starting nftables..."
    systemctl enable nftables
    systemctl start nftables
    if [[ $? -ne 0 ]]; then
        echo "Failed to enable or start nftables."
        exit 1
    fi
}

# 添加规则，确保无重复
add_nft_rule() {
    echo "Adding nftables rule for ports 9918 and 9919..."
    existing_rule=$(nft list ruleset | grep 'tcp dport { 9918, 9919 } accept')
    if [[ -z "$existing_rule" ]]; then
        nft add rule ip filter input tcp dport { 9918, 9919 } accept
        if [[ $? -ne 0 ]]; then
            echo "Failed to add nftables rule."
            exit 1
        fi
    else
        echo "Rule already exists. Skipping."
    fi
}

# 主逻辑
main() {
    disable_iptables
    enable_nftables
    add_nft_rule
    echo "Success"
}

main
