#!/bin/bash

# detect-firewall.sh - 判断当前系统是使用 iptables (legacy)、iptables-nft，还是 nftables

# 检查 iptables 是否存在并判断其后端
if ! command -v iptables >/dev/null 2>&1; then
    IPTABLES_BACKEND="none"
else
    IPTABLES_BACKEND=$(iptables --version 2>/dev/null | grep -oE '\((nf_tables|legacy)\)' | tr -d '()')
fi

# 检查 nftables 是否活跃（是否存在原生表）
if command -v nft >/dev/null 2>&1; then
    # 使用 subshell 截断输出
    NFT_RULESET=$(nft list ruleset 2>/dev/null)
    if echo "$NFT_RULESET" | grep -q 'table'; then
        NFT_ACTIVE="yes"
    else
        NFT_ACTIVE="no"
    fi
else
    NFT_ACTIVE="no"
fi

# 输出最终判断结果
if [ "$IPTABLES_BACKEND" = "legacy" ] && [ "$NFT_ACTIVE" = "no" ]; then
    echo "iptables (legacy) is active"
elif [ "$IPTABLES_BACKEND" = "nf_tables" ] && [ "$NFT_ACTIVE" = "no" ]; then
    echo "iptables-nft (compatibility layer) is active"
elif [ "$NFT_ACTIVE" = "yes" ]; then
    echo "nftables is active"
elif [ "$IPTABLES_BACKEND" = "none" ] && [ "$NFT_ACTIVE" = "no" ]; then
    echo "no firewall system detected"
else
    echo "uncertain state: please check manually"
fi
