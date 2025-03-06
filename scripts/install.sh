#!/bin/bash

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/install.log
}

echo
cat << EOF     
██╗██████╗ ██████╗ 
██║██╔══██╗██╔══██╗
██║██║  ██║██████╔╝
██║██║  ██║██╔══██╗
██║██████╔╝██████╔╝
╚═╝╚═════╝ ╚═════╝ 
EOF

log "======================= 开始安装 ======================="

function Check_Root() {
    if [[ $EUID -ne 0 ]]; then
        echo "请使用 root 或 sudo 权限运行此脚本"
        exit 1
    fi
}

function Check_Architecture() {
    osCheck=`uname -a`
    if [[ $osCheck =~ 'x86_64' ]];then
        architecture="amd64"
    # elif [[ $osCheck =~ 'arm64' ]] || [[ $osCheck =~ 'aarch64' ]];then
    #     architecture="arm64"
    # elif [[ $osCheck =~ 'armv7l' ]];then
    #     architecture="armv7"
    # elif [[ $osCheck =~ 'ppc64le' ]];then
    #     architecture="ppc64le"
    # elif [[ $osCheck =~ 's390x' ]];then
    #     architecture="s390x"
    else
        echo "暂不支持的系统架构，请参阅官方文档，选择受支持的系统。"
        exit 1
    fi
}

function Configure_Docker_Mirror() {
    if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
        log "配置 Docker 镜像加速器..."
        mkdir -p /etc/docker
        cat > /etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://docker.1panel.live",
        "https://hub.fast360.xyz",
        "https://docker-0.unsee.tech",
        "https://docker.tbedu.top",
        "https://dockerpull.cn"
    ]
}
EOF
        systemctl daemon-reload
        systemctl restart docker
        log "Docker 镜像加速器配置完成"
    fi
}

function Install_Docker(){
    if which docker >/dev/null 2>&1; then
        log "检测到 Docker 已安装，跳过安装步骤"
        log "启动 Docker "
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log
        Configure_Docker_Mirror  # 配置镜像加速
    else
        log "... 在线安装 docker"

        if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
            sources=(
                "https://mirrors.aliyun.com/docker-ce"
                "https://mirrors.tencent.com/docker-ce"
                "https://mirrors.163.com/docker-ce"
                "https://mirrors.cernet.edu.cn/docker-ce"
            )

            get_average_delay() {
                local source=$1
                local total_delay=0
                local iterations=3

                for ((i = 0; i < iterations; i++)); do
                    delay=$(curl -o /dev/null -s -w "%{time_total}\n" "$source")
                    total_delay=$(awk "BEGIN {print $total_delay + $delay}")
                done

                average_delay=$(awk "BEGIN {print $total_delay / $iterations}")
                echo "$average_delay"
            }

            min_delay=${#sources[@]}
            selected_source=""

            for source in "${sources[@]}"; do
                average_delay=$(get_average_delay "$source")

                if (( $(awk 'BEGIN { print '"$average_delay"' < '"$min_delay"' }') )); then
                    min_delay=$average_delay
                    selected_source=$source
                fi
            done

            if [ -n "$selected_source" ]; then
                echo "选择延迟最低的源 $selected_source，延迟为 $min_delay 秒"
                export DOWNLOAD_URL="$selected_source"
                curl -fsSL "https://get.docker.com" -o get-docker.sh
                sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log

                log "... 启动 docker"
                systemctl enable docker; systemctl daemon-reload; systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log

                docker_config_folder="/etc/docker"
                if [[ ! -d "$docker_config_folder" ]];then
                    mkdir -p "$docker_config_folder"
                fi

                docker version >/dev/null 2>&1
                if [[ $? -ne 0 ]]; then
                    log "docker 安装失败"
                    exit 1
                else
                    log "docker 安装成功"
                    Configure_Docker_Mirror  # 配置镜像加速
                fi
            else
                log "无法选择源进行安装"
                exit 1
            fi
        else
            log "非中国大陆地区，无需更改源"
            export DOWNLOAD_URL="https://download.docker.com"
            curl -fsSL "https://get.docker.com" -o get-docker.sh
            sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log

            log "... 启动 docker"
            systemctl enable docker; systemctl daemon-reload; systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log

            docker_config_folder="/etc/docker"
            if [[ ! -d "$docker_config_folder" ]];then
                mkdir -p "$docker_config_folder"
            fi

            docker version >/dev/null 2>&1
            if [[ $? -ne 0 ]]; then
                log "docker 安装失败"
                exit 1
            else
                log "docker 安装成功"
            fi
        fi
    fi
}

function Install_Compose(){
    docker-compose version >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        log "... 在线安装 docker-compose"
        
        arch=$(uname -m)
        if [ "$arch" == 'armv7l' ]; then
            arch='armv7'
        fi
        curl -L https://github.com/docker/compose/releases/download/v2.26.1/docker-compose-$(uname -s | tr '[:upper:]' '[:lower:]')-$arch -o /usr/local/bin/docker-compose 2>&1 | tee -a ${CURRENT_DIR}/install.log
        if [[ ! -f /usr/local/bin/docker-compose ]];then
            log "docker-compose 下载失败，请稍候重试"
            exit 1
        fi
        chmod +x /usr/local/bin/docker-compose
        ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

        docker-compose version >/dev/null 2>&1
        if [[ $? -ne 0 ]]; then
            log "docker-compose 安装失败"
            exit 1
        else
            log "docker-compose 安装成功"
        fi
    else
        compose_v=`docker-compose -v`
        if [[ $compose_v =~ 'docker-compose' ]];then
            read -p "检测到已安装 Docker Compose 版本较低（需大于等于 v2.0.0 版本），是否升级 [y/n] : " UPGRADE_DOCKER_COMPOSE
            if [[ "$UPGRADE_DOCKER_COMPOSE" == "Y" ]] || [[ "$UPGRADE_DOCKER_COMPOSE" == "y" ]]; then
                rm -rf /usr/local/bin/docker-compose /usr/bin/docker-compose
                Install_Compose
            else
                log "Docker Compose 版本为 $compose_v，可能会影响应用商店的正常使用"
            fi
        else
            log "检测到 Docker Compose 已安装，跳过安装步骤"
        fi
    fi
}

function Check_Installation() {
    if docker ps -a -q -f name=idb >/dev/null 2>&1; then
        log "检测到已安装的 IDB 容器"
        read -p "是否要升级安装？这将备份现有数据 [y/n]: " UPGRADE_IDB
        if [[ "$UPGRADE_IDB" == "Y" ]] || [[ "$UPGRADE_IDB" == "y" ]]; then
            Backup_Data
            log "准备安装新版本..."
            return 0
        else
            log "取消安装"
            exit 1
        fi
    fi
}

function Backup_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    log "清理临时目录..."
    rm -rf "$BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    
    if docker ps -q -f name=idb >/dev/null 2>&1; then
        log "停止 IDB 容器..."
        docker stop idb
    fi
    
    if docker ps -a -q -f name=idb >/dev/null 2>&1; then
        # 保存当前环境变量值
        docker inspect idb > "$BACKUP_DIR/container_info.json"
        
        # 从容器配置中获取实际的挂载路径
        local DATA_PATH=$(docker inspect idb --format '{{ range .Mounts }}{{ if eq .Destination "/var/lib/idb/data" }}{{ .Source }}{{ end }}{{ end }}')
        local LOG_PATH=$(docker inspect idb --format '{{ range .Mounts }}{{ if eq .Destination "/var/log/idb" }}{{ .Source }}{{ end }}{{ end }}')
        
        # 备份实际的数据目录
        if [[ -n "$DATA_PATH" && -d "$DATA_PATH" ]]; then
            log "备份数据目录: $DATA_PATH"
            cp -r "$DATA_PATH" "$BACKUP_DIR/data"
        fi
        
        if [[ -n "$LOG_PATH" && -d "$LOG_PATH" ]]; then
            log "备份日志目录: $LOG_PATH"
            cp -r "$LOG_PATH" "$BACKUP_DIR/logs"
        fi
        
        log "删除旧容器..."
        docker rm idb
        
        return 0
    fi
    
    return 1
}

function Restore_Data() {
    local BACKUP_DIR="/tmp/idb-cache"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log "未找到备份数据，执行全新安装"
        return 0
    fi
    
    log "开始恢复数据..."
    
    # 确保目标目录存在
    mkdir -p "${PANEL_DIR}/data"
    mkdir -p "${PANEL_DIR}/logs"
    
    # 恢复数据到宿主机目录
    if [[ -d "$BACKUP_DIR/data" ]]; then
        log "恢复数据目录到: ${PANEL_DIR}/data"
        cp -r "$BACKUP_DIR/data/." "${PANEL_DIR}/data/"
    fi
    
    if [[ -d "$BACKUP_DIR/logs" ]]; then
        log "恢复日志目录到: ${PANEL_DIR}/logs"
        cp -r "$BACKUP_DIR/logs/." "${PANEL_DIR}/logs/"
    fi
    
    log "数据恢复完成"
    
    # 清理临时目录
    log "清理临时文件..."
    rm -rf "$BACKUP_DIR"
}

function Set_Dir(){
    DEFAULT_DIR='/var/lib/idb'

    while true; do
        read -p "设置 idb 的目录 (默认为 ${DEFAULT_DIR}): " PANEL_DIR

        if [[ "$PANEL_DIR" == "" ]]; then
            PANEL_DIR=$DEFAULT_DIR
        fi

        # 判断目录是否合法
        if [[ ! "$PANEL_DIR" =~ ^/ ]]; then
            echo "错误：目录必须是绝对路径。"
            continue
        fi

        # 判断目录是否存在，如果不存在，则创建
        if [[ ! -d "$PANEL_DIR" ]]; then
            log "目录 ${PANEL_DIR} 不存在，正在创建..."
            mkdir -p "$PANEL_DIR"
            if [[ $? -ne 0 ]]; then
                log "创建目录 ${PANEL_DIR} 失败，请检查权限。"
                exit 1
            fi
        fi

        log "您设置的目录为：${PANEL_DIR}"
        break
    done
}

function Set_Port(){
    DEFAULT_PORT='9918'

    while true; do
        read -p "设置 idb 端口 (默认为 ${DEFAULT_PORT}): " PANEL_PORT

        if [[ "$PANEL_PORT" == "" ]];then
            PANEL_PORT=$DEFAULT_PORT
        fi

        if ! [[ "$PANEL_PORT" =~ ^[1-9][0-9]{0,4}$ && "$PANEL_PORT" -le 65535 ]]; then
            echo "错误：输入的端口号必须在 1 到 65535 之间"
            continue
        fi

        if command -v ss >/dev/null 2>&1; then
            if ss -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                echo "端口${PANEL_PORT}被占用，请重新输入..."
                continue
            fi
        elif command -v netstat >/dev/null 2>&1; then
            if netstat -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                echo "端口${PANEL_PORT}被占用，请重新输入..."
                continue
            fi
        fi

        log "您设置的端口为：${PANEL_PORT}"
        break
    done
}

function Set_Firewall(){
    if which firewall-cmd >/dev/null 2>&1; then
        if systemctl status firewalld | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 ${PANEL_PORT} 端口"
            firewall-cmd --zone=public --add-port=$PANEL_PORT/tcp --permanent
            firewall-cmd --reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi

    if which ufw >/dev/null 2>&1; then
        if systemctl status ufw | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 ${PANEL_PORT} 端口"
            ufw allow $PANEL_PORT/tcp
            ufw reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi
}

function Set_Container_Port(){
    DEFAULT_CONTAINER_PORT='9918'

    while true; do
        read -p "设置容器端口 (默认为${DEFAULT_CONTAINER_PORT}) :" CONTAINER_PORT

        if [[ "$CONTAINER_PORT" == "" ]];then
            CONTAINER_PORT=$DEFAULT_CONTAINER_PORT
        fi

        if ! [[ "$CONTAINER_PORT" =~ ^[1-9][0-9]{0,4}$ && "$CONTAINER_PORT" -le 65535 ]]; then
            echo "错误：输入的端口号必须在 1 到 65535 之间"
            continue
        fi

        log "您设置的容器端口为：${CONTAINER_PORT}"
        break
    done
}

function Install_IDB() {
    # 获取版本号
    VERSION=$(curl -s https://static.sensdata.com/idb/release/latest)

    if [[ "x${VERSION}" == "x" ]];then
        echo "获取最新版本失败，请稍候重试"
        exit 1
    fi

    # 下载 .env和docker-compose.yml 到 PANEL_DIR 中
    # .env 地址: "https://static.sensdata.com/idb/release/${VERSION}/.env"
    # docker-compose.yml 地址: "https://static.sensdata.com/idb/release/${VERSION}/docker-compose.yml"
    ENV_URL="https://static.sensdata.com/idb/release/${VERSION}/.env"
    DOCKER_COMPOSE_URL="https://static.sensdata.com/idb/release/${VERSION}/docker-compose.yml"

    log "正在下载 .env 文件..."
    curl -fsSL "$ENV_URL" -o "${PANEL_DIR}/.env" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log ".env 文件下载失败，请检查网络连接或 URL 是否正确。"
        exit 1
    fi

    log "正在下载 docker-compose.yml 文件..."
    curl -fsSL "$DOCKER_COMPOSE_URL" -o "${PANEL_DIR}/docker-compose.yml" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "docker-compose.yml 文件下载失败，请检查网络连接或 URL 是否正确。"
        exit 1
    fi

    log ".env 和 docker-compose.yml 文件下载成功。"
    
    # 修改 .env 文件中的端口配置
    log "正在修改 .env 文件中的配置..."
    sed -i "s/^iDB_service_host_ip=.*/iDB_service_host_ip=${LOCAL_IP}/" "${PANEL_DIR}/.env"
    sed -i "s/^iDB_service_port=.*/iDB_service_port=${PANEL_PORT}/" "${PANEL_DIR}/.env"
    sed -i "s/^iDB_service_container_port=.*/iDB_service_container_port=${CONTAINER_PORT}/" "${PANEL_DIR}/.env"
    
    log ".env 文件内容已更新为：\n$(cat ${PANEL_DIR}/.env)"

    # 在启动容器前恢复数据
    Restore_Data

    # 启动新容器
    log "正在启动 IDB..."
    cd "$PANEL_DIR" || { log "无法进入目录 $PANEL_DIR"; exit 1; }
    
    docker-compose up -d 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "启动 IDB 失败，请检查 docker-compose 配置。"
        exit 1
    fi

    log "IDB 启动成功！"

    # 从 idb 容器的 /var/lib/idb/data/agent 目录下，拷贝 idb-agent_${VERSION}.tar.gz 至当前目录下的 agent目录
    log "正在拷贝 idb-agent 文件..."
    mkdir -p "${CURRENT_DIR}/agent"  # 创建 agent 目录
    docker cp "idb:/var/lib/idb/data/agent/idb-agent.tar.gz" "${CURRENT_DIR}/agent/" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "拷贝 idb-agent 文件失败，请检查容器是否存在。"
        exit 1
    fi

    log "正在解压 idb-agent 文件..."
    tar -xzvf "${CURRENT_DIR}/agent/idb-agent.tar.gz" -C "${CURRENT_DIR}/agent/" 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "解压 idb-agent 文件失败。"
        exit 1
    fi

    log "idb-agent 文件解压成功。"

    # 进入 agent 目录并执行 install-agent.sh
    log "正在执行 install-agent.sh..."
    cd "${CURRENT_DIR}/agent" || { log "无法进入目录 ${CURRENT_DIR}/agent"; exit 1; }
    
    sudo bash ./install-agent.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log
    if [[ $? -ne 0 ]]; then
        log "执行 install-agent.sh 失败，请检查脚本内容。"
        exit 1
    fi

    log "install-agent.sh 执行成功。"
}

function Get_Ip(){
    # 优先获取默认路由对应的源IP
    LOCAL_IP=$(ip route get 8.8.8.8 | grep -oP 'src \K[^ ]+')
    if [[ -z "$LOCAL_IP" ]]; then
        # 备选方案：获取默认网卡的IP
        default_interface=$(ip route | grep '^default' | awk '{print $5}')
        if [[ -n "$default_interface" ]]; then
            LOCAL_IP=$(ip addr show $default_interface | grep -oP 'inet \K[\d.]+')
        fi
    fi
    
    # 如果上述方法都失败，使用默认值
    if [[ -z "$LOCAL_IP" ]]; then
        LOCAL_IP="127.0.0.1"
    fi

    PUBLIC_IP=`curl -s https://api64.ipify.org`
    if [[ -z "$PUBLIC_IP" ]]; then
        PUBLIC_IP="N/A"
    fi
    if echo "$PUBLIC_IP" | grep -q ":"; then
        PUBLIC_IP=[${PUBLIC_IP}]
    fi
}

function Show_Result(){
    log ""
    log "=================感谢您的耐心等待，安装已经完成=================="
    log ""
    log "请用浏览器访问面板:"
    log "外网地址: http://${PUBLIC_IP}:${PANEL_PORT}/idb"
    log "内网地址: http://${LOCAL_IP}:${PANEL_PORT}/idb"
    log "初始用户: admin"
    log "初始密码: admin123"
    log ""
    log "项目官网: https://idb.sensdata.com"
    log "项目文档: https://idb.sensdata.com/docs"
    log "代码仓库: https://github.com/sensdata/idb"
    log ""
    log "如果使用的是云服务器，请至安全组开放 $PANEL_PORT 端口"
    log ""
    log "为了您的服务器安全，在您离开此界面后您将无法再看到您的密码，请务必牢记您的密码。"
    log ""
    log "================================================================"
}

function main(){
    Check_Root
    Install_Docker
    Install_Compose
    Check_Installation
    Set_Dir
    Set_Port
    Set_Firewall
    Set_Container_Port
    Get_Ip
    Install_IDB
    Show_Result
}
main