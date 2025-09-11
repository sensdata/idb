# 找到docker-compose.yaml文件所在的目录
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# 确认卸载
echo "警告：此操作将完全删除 IDB 所有组件和数据！"
read -p "确认要继续卸载吗？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "卸载已取消"
    exit 1
fi

echo "开始卸载..."

# 停止并删除所有的容器
echo "停止并删除容器..."
docker compose -f $DIR/docker-compose.yaml down --remove-orphans

# 删除相关的 Docker 镜像
echo "清理 Docker 镜像..."
docker images | grep "idb" | awk '{print $3}' | xargs -r docker rmi -f

# 停止本机agent服务
echo "停止 agent 服务..."
if systemctl is-active --quiet idb-agent; then
    sudo systemctl stop idb-agent
fi

# 卸载本机agent服务
echo "卸载 agent 服务..."
if systemctl is-enabled --quiet idb-agent; then
    sudo systemctl disable idb-agent
fi

# 清理center挂载目录
echo "清理 center 目录..."
for dir in "/var/lib/idb" "/var/log/idb" "/etc/idb"; do
    if [ -d "$dir" ]; then
        sudo rm -rf "$dir"
    fi
done

# 清理agent目录
echo "清理 agent 目录..."
for dir in "/var/lib/idb-agent" "/var/log/idb-agent" "/etc/idb-agent" "/run/idb-agent"; do
    if [ -d "$dir" ]; then
        rm -rf "$dir"
    fi
done

echo "卸载完成！"