#!/bin/bash

# 检测nftables是否已安装
check_nftables_installed() {
    if command -v nft >/dev/null 2>&1; then
        return 0  # 已安装
    else
        return 1  # 未安装
    fi
}

# 根据操作系统安装nftables
install_nftables() {
    if command -v lsb_release >/dev/null 2>&1; then
        OS_TYPE=$(lsb_release -i | awk -F: '{print $2}' | tr -d '[:space:]')
    else
        return 1  # 无法检测系统类型
    fi
    
    if [ "$OS_TYPE" == "Debian" ] || [ "$OS_TYPE" == "Ubuntu" ]; then
        sudo apt-get update && sudo apt-get install -y nftables
    elif [ "$OS_TYPE" == "CentOS" ] || [ "$OS_TYPE" == "RedHat" ] || [ "$OS_TYPE" == "Fedora" ]; then
        sudo yum install -y nftables
    elif [ "$OS_TYPE" == "Arch" ]; then
        sudo pacman -Syu --noconfirm nftables
    else
        return 1  # 不支持的操作系统
    fi

    # 检查安装是否成功
    if ! command -v nft >/dev/null 2>&1; then
        return 1  # 安装失败
    fi
    return 0  # 安装成功
}

# 启用并启动nftables服务
enable_and_start_nftables() {
    sudo systemctl enable nftables
    sudo systemctl start nftables
}

# 禁用iptables并切换到nftables
handle_iptables_conflict() {
    if systemctl is-active --quiet iptables; then
        sudo systemctl stop iptables
        sudo systemctl disable iptables
    fi
}

# 主逻辑
main() {
    installed=-1  # 初始值设为 -1，表示默认未安装

    if check_nftables_installed; then
        installed=0  # 已安装，无需安装
    else
        if install_nftables; then
            installed=1  # 未安装，安装成功
        fi
    fi

    enable_and_start_nftables
    handle_iptables_conflict

    # 根据 installed 的值输出结果
    case $installed in
        -1) echo "Failed" ;;  # 安装失败
        0) echo "Installed" ;;  # 已安装，无需安装
        1) echo "Installed" ;;  # 安装成功
        *) echo "Failed" ;;  # 其他
    esac
}

main
