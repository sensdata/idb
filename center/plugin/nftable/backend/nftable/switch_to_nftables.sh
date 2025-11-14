#!/bin/bash
set -euo pipefail

# 停止并禁用 iptables 服务
disable_iptables() {
    if systemctl is-active --quiet iptables; then
        echo "Stopping and disabling iptables..."
        systemctl stop iptables
        systemctl disable iptables
    elif systemctl is-active --quiet netfilter-persistent; then
        echo "Stopping and disabling netfilter-persistent..."
        systemctl stop netfilter-persistent
        systemctl disable netfilter-persistent
    else
        echo "iptables service not active."
    fi
}

# 启用并启动 nftables 服务
enable_nftables() {
    echo "Enabling and starting nftables..."
    systemctl enable nftables
    systemctl start nftables
}

# 写入规则文件并加载
load_idb_nftables_rules() {
    RULES_FILE="/etc/nftables.conf"

    if [[ -s "$RULES_FILE" ]]; then
        echo "Skip: $RULES_FILE already exists and is not empty, not overwriting."
    else
        echo "Writing idb-filter rules to $RULES_FILE..."
        cat > "$RULES_FILE" <<EOF
table inet idb-filter {
    chain input {
        type filter hook input priority 0; policy drop;

        icmp type echo-request drop
        iifname "lo" accept
        iifname "docker0" accept
        iifname "br-+" accept
        iifname "veth+" accept
        iifname "docker_gwbridge" accept
        ct state established,related accept
        tcp dport 22 accept
        tcp dport 9918 accept
        tcp dport 9919 accept
    }

    chain output {
        type filter hook output priority 0; policy accept;
    }

    chain forward {
        type filter hook forward priority 0; policy accept;
    }
}
EOF
    fi

    echo "Loading rules from $RULES_FILE..."
    nft -f "$RULES_FILE"
}

main() {
    disable_iptables
    enable_nftables
    load_idb_nftables_rules
    echo "Success: switched to nftables with idb-filter table"
}

main
