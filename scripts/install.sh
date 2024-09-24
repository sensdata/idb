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

function Prepare_System(){
    if which idb >/dev/null 2>&1; then
        log "IDB Linux 服务器运维管理面板已安装，请勿重复安装"
        exit 1
    fi
}

function Install_Docker(){
    if which docker >/dev/null 2>&1; then
        log "检测到 Docker 已安装，跳过安装步骤"
        log "启动 Docker "
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log
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

function Set_Port(){
    DEFAULT_PORT=`25900`

    while true; do
        read -p "设置 idb 端口（默认为$DEFAULT_PORT）：" PANEL_PORT

        if [[ "$PANEL_PORT" == "" ]];then
            PANEL_PORT=$DEFAULT_PORT
        fi

        if ! [[ "$PANEL_PORT" =~ ^[1-9][0-9]{0,4}$ && "$PANEL_PORT" -le 65535 ]]; then
            echo "错误：输入的端口号必须在 1 到 65535 之间"
            continue
        fi

        if command -v ss >/dev/null 2>&1; then
            if ss -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                echo "端口$PANEL_PORT被占用，请重新输入..."
                continue
            fi
        elif command -v netstat >/dev/null 2>&1; then
            if netstat -tlun | grep -q ":$PANEL_PORT " >/dev/null 2>&1; then
                echo "端口$PANEL_PORT被占用，请重新输入..."
                continue
            fi
        fi

        log "您设置的端口为：$PANEL_PORT"
        break
    done
}

function Set_Firewall(){
    if which firewall-cmd >/dev/null 2>&1; then
        if systemctl status firewalld | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 $PANEL_PORT 端口"
            firewall-cmd --zone=public --add-port=$PANEL_PORT/tcp --permanent
            firewall-cmd --reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi

    if which ufw >/dev/null 2>&1; then
        if systemctl status ufw | grep -q "Active: active" >/dev/null 2>&1;then
            log "防火墙开放 $PANEL_PORT 端口"
            ufw allow $PANEL_PORT/tcp
            ufw reload
        else
            log "防火墙未开启，忽略端口开放"
        fi
    fi
}

function Init_Panel(){
    log "配置 IDB Service"

    # 目录
    BIN_DIR=/usr/local/bin
    SERVICE_DIR=/etc/systemd/system/
    CONFIG_DIR=/etc/idb
    DATA_DIR=/var/lib/idb
    LOG_DIR =/var/log/idb

    # 配置文件
    CONFIG_FILE=idb.conf

    # 注册服务文件
    SERVICE_FILE=idb.service

    # 日志文件路径
    LOG_FILE=/var/log/idb/idb-run.log

    # 创建
    mkdir -p $CONFIG_DIR
    mkdir -p $DATA_DIR
    mkdir -p $LOG_DIR

    # 清理
    rm -rf $CONFIG_DIR/*
    rm -rf $DATA_DIR/*
    rm -rf $LOG_DIR/*

    cd ${CURRENT_DIR}

    # 拷贝二进制文件
    cp ./idb $BIN_DIR && chmod +x $BIN_DIR/idb
    if [[ ! -f /usr/bin/idb ]]; then
        ln -s /usr/local/bin/idb /usr/bin/idb >/dev/null 2>&1
    fi
    
    # 拷贝配置文件
    cp ./$CONFIG_FILE $CONFIG_DIR
    
    # 修改端口
    sed -i 's/port=25900/port=${PANEL_PORT}/' config.conf

    # 拷贝服务定义文件
    cp ./$SERVICE_FILE $SERVICE_DIR

    # 运行日志
    touch $LOG_FILE

    systemctl enable $SERVICE_FILE; systemctl daemon-reload 2>&1 | tee -a ${CURRENT_DIR}/install.log

    log "启动 IDB 服务"
    systemctl start $SERVICE_FILE | tee -a ${CURRENT_DIR}/install.log

    for b in {1..30}
    do
        sleep 3
        service_status=`systemctl status $SERVICE_FILE 2>&1 | grep Active`
        if [[ $service_status == *running* ]];then
            log "IDB 服务启动成功!"
            break;
        else
            log "IDB 服务启动出错!"
            exit 1
        fi
    done
}

function Get_Ip(){
    active_interface=$(ip route get 8.8.8.8 | awk 'NR==1 {print $5}')
    if [[ -z $active_interface ]]; then
        LOCAL_IP="127.0.0.1"
    else
        LOCAL_IP=`ip -4 addr show dev "$active_interface" | grep -oP '(?<=inet\s)\d+(\.\d+){3}'`
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
    log "外网地址: http://$PUBLIC_IP:$PANEL_PORT/idb"
    log "内网地址: http://$LOCAL_IP:$PANEL_PORT/idb"
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
    Prepare_System
    Install_Docker
    Install_Compose
    Set_Port
    Set_Firewall
    Init_Panel
    Get_Ip
    Show_Result
}
main