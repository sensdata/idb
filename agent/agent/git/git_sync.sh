#!/bin/bash

# 检查参数
if [ $# -lt 2 ]; then
    echo "Usage: $0 <git_url> <local_path>"
    echo "Example: $0 https://8.138.47.21:9918/api/v1/git/scripts/global /var/lib/idb/data/scripts/global"
    exit 1
fi

GIT_URL=$1
LOCAL_PATH=$2

# 配置 Git 跳过 SSL 验证
export GIT_SSL_NO_VERIFY=1

# 确保本地路径存在
mkdir -p $(dirname "$LOCAL_PATH")

# 检查是否已经存在 Git 仓库
if [ -d "$LOCAL_PATH/.git" ]; then
    echo "Repository exists, performing force update..."
    cd "$LOCAL_PATH"
    
    # 保存当前分支名
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    
    # 获取远程最新代码并强制覆盖本地
    git -c http.sslVerify=false fetch origin
    git reset --hard origin/$CURRENT_BRANCH
    git clean -fd
    
    if [ $? -eq 0 ]; then
        echo "Repository updated successfully"
    else
        echo "Failed to update repository"
        exit 1
    fi
else
    echo "Repository does not exist, cloning..."
    git -c http.sslVerify=false clone "$GIT_URL" "$LOCAL_PATH"
    
    if [ $? -eq 0 ]; then
        echo "Repository cloned successfully"
    else
        echo "Failed to clone repository"
        exit 1
    fi
fi

exit 0