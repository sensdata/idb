#!/bin/bash

# 检测screen是否已安装
check_screen_installed() {
    if command -v screen >/dev/null 2>&1; then
        SCREEN_VERSION=$(screen --version 2>&1 | head -n 1)
        return 0  # 已安装
    else
        return 1  # 未安装
    fi
}

# 根据操作系统安装screen
install_screen() {
    # 检测系统类型
    OS_TYPE="Unknown"
    OS_VERSION="Unknown"
    
    if command -v lsb_release >/dev/null 2>&1; then
        # 方法1: 使用lsb_release (首选)
        OS_TYPE=$(lsb_release -i | awk -F: '{print $2}' | tr -d '[:space:]')
        OS_VERSION=$(lsb_release -r | awk -F: '{print $2}' | tr -d '[:space:]')
    elif [ -f /etc/os-release ]; then
        # 方法2: 检查/etc/os-release文件
        source /etc/os-release
        if [ ! -z "$ID" ]; then
            OS_TYPE=$(echo "$ID" | tr '[:lower:]' '[:upper:]')
            if [ ! -z "$VERSION_ID" ]; then
                OS_VERSION="$VERSION_ID"
            fi
        fi
    elif [ -f /etc/redhat-release ]; then
        # 方法3: 检查/etc/redhat-release (RedHat系列)
        REDHAT_INFO=$(cat /etc/redhat-release)
        if echo "$REDHAT_INFO" | grep -i "CentOS" >/dev/null; then
            OS_TYPE="CentOS"
        elif echo "$REDHAT_INFO" | grep -i "Red Hat" >/dev/null; then
            OS_TYPE="RedHat"
        elif echo "$REDHAT_INFO" | grep -i "Fedora" >/dev/null; then
            OS_TYPE="Fedora"
        fi
        # 尝试从redhat-release提取版本号
        if [[ "$REDHAT_INFO" =~ ([0-9]+\.[0-9]+) ]]; then
            OS_VERSION="${BASH_REMATCH[1]}"
        fi
    elif [ -f /etc/debian_version ]; then
        # 方法4: 检查/etc/debian_version (Debian/Ubuntu系列)
        DEBIAN_VERSION=$(cat /etc/debian_version)
        if [ -f /etc/lsb-release ]; then
            source /etc/lsb-release
            if [ ! -z "$DISTRIB_ID" ]; then
                OS_TYPE="$DISTRIB_ID"
                if [ ! -z "$DISTRIB_RELEASE" ]; then
                    OS_VERSION="$DISTRIB_RELEASE"
                fi
            else
                OS_TYPE="Debian"
                OS_VERSION="$DEBIAN_VERSION"
            fi
        else
            OS_TYPE="Debian"
            OS_VERSION="$DEBIAN_VERSION"
        fi
    else
        return 1  # 无法检测系统类型
    fi
    
    # 标准化OS_TYPE以确保匹配逻辑正确
    if [[ "$OS_TYPE" == *"UBUNTU"* ]]; then
        OS_TYPE="Ubuntu"
    elif [[ "$OS_TYPE" == *"DEBIAN"* ]]; then
        OS_TYPE="Debian"
    elif [[ "$OS_TYPE" == *"CENTOS"* ]]; then
        OS_TYPE="CentOS"
    elif [[ "$OS_TYPE" == *"REDHAT"* || "$OS_TYPE" == *"RED HAT"* ]]; then
        OS_TYPE="RedHat"
    elif [[ "$OS_TYPE" == *"FEDORA"* ]]; then
        OS_TYPE="Fedora"
    elif [[ "$OS_TYPE" == *"ARCH"* ]]; then
        OS_TYPE="Arch"
    fi
    
    # 根据系统类型执行安装
    if [ "$OS_TYPE" == "Debian" ] || [ "$OS_TYPE" == "Ubuntu" ]; then
        sudo apt-get update
        sudo apt-get install -y screen
        INSTALL_RESULT=$?
    elif [ "$OS_TYPE" == "CentOS" ] || [ "$OS_TYPE" == "RedHat" ] || [ "$OS_TYPE" == "Fedora" ]; then
        sudo yum install -y screen
        INSTALL_RESULT=$?
    elif [ "$OS_TYPE" == "Arch" ]; then
        sudo pacman -Syu --noconfirm screen
        INSTALL_RESULT=$?
    else
        return 1  # 不支持的操作系统
    fi

    # 检查安装是否成功
    if command -v screen >/dev/null 2>&1; then
        SCREEN_VERSION=$(screen --version 2>&1 | head -n 1)
        return 0  # 安装成功
    else
        return 1  # 安装失败
    fi
}

# 主逻辑
main() {
    installed=-1  # 初始值设为 -1，表示默认未安装

    if check_screen_installed; then
        installed=0  # 已安装，无需安装
    else
        if install_screen; then
            installed=1  # 未安装，安装成功
        fi
    fi

    # 根据 installed 的值输出结果
    case $installed in
        -1) echo "Failed" ;;  # 安装失败
        0) echo "Installed" ;;  # 已安装，无需安装
        1) echo "Success" ;;  # 安装成功
        *) echo "Failed" ;;  # 其他
    esac
}

main