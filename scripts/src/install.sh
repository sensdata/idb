#!/bin/bash

set -euo pipefail

CURRENT_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

IDB_DEFAULT_PROXY="https://dl.idb.net"
IDB_INSTALL_ROOT="/var/lib/idb"
IDB_HOME_DIR="${IDB_INSTALL_ROOT}/home"
IDB_DATA_DIR="${IDB_INSTALL_ROOT}/data"
IDB_AGENT_DIR="${IDB_INSTALL_ROOT}/agent"
IDB_PLUGIN_DIR="${IDB_DATA_DIR}/plugins"
IDB_LOG_DIR="/var/log/idb"
IDB_RUN_DIR="/run/idb"
IDB_CONF_DIR="/etc/idb"
IDB_AGENT_CONF_DIR="/etc/idb-agent"
IDB_AGENT_RUN_DIR="/run/idb-agent"
IDB_AGENT_LOG_DIR="/var/log/idb-agent"
IDB_AGENT_DATA_DIR="/var/lib/idb-agent/data"
TMP_DIR="/tmp/idb-install"
AGENT_INSTALL_FAILED="false"
AGENT_INSTALL_ERROR=""

function log() {
    message="[idb Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a "${CURRENT_DIR}/install.log"
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

# __IDB_COMMON__

function Ensure_Directories() {
    mkdir -p "${IDB_INSTALL_ROOT}" "${IDB_HOME_DIR}" "${IDB_DATA_DIR}" "${IDB_AGENT_DIR}" "${IDB_PLUGIN_DIR}"
    mkdir -p "${IDB_LOG_DIR}" "${IDB_RUN_DIR}" "${IDB_CONF_DIR}"
    mkdir -p "${IDB_AGENT_CONF_DIR}" "${IDB_AGENT_RUN_DIR}" "${IDB_AGENT_LOG_DIR}" "${IDB_AGENT_DATA_DIR}"
}

function Ensure_Admin_Env() {
    local admin_env="${IDB_CONF_DIR}/idb.env"
    local admin_pass

    touch "${admin_env}"
    chmod 600 "${admin_env}"

    if [[ ! -f "${IDB_DATA_DIR}/idb.db" ]] && ! grep -q '^PASSWORD=' "${admin_env}"; then
        admin_pass=$(tr -dc 'a-z0-9' </dev/urandom | head -c 8)
        echo "PASSWORD=${admin_pass}" > "${admin_env}"
        chmod 600 "${admin_env}"
        INITIAL_ADMIN_PASS="${admin_pass}"
    fi
}

function Detect_Target_Host() {
    local public_ip
    local local_ip

    public_ip=$(curl -s -4 --connect-timeout 3 --max-time 5 https://api.ipify.org 2>/dev/null || true)
    if [[ -z "${public_ip}" ]]; then
        public_ip=$(curl -s -4 --connect-timeout 3 --max-time 5 https://api64.ipify.org 2>/dev/null || true)
    fi
    if [[ -z "${public_ip}" || "${public_ip}" == *:* ]]; then
        public_ip=""
    fi

    local_ip=$(ip -4 route get 8.8.8.8 2>/dev/null | grep -oE 'src [0-9.]*' | awk '{print $2}' | head -n1)

    TARGET_HOST="${public_ip:-${local_ip:-127.0.0.1}}"
    LOCAL_IP="${local_ip:-127.0.0.1}"
    PUBLIC_IP="${public_ip:-N/A}"
}

function Install_Center() {
    local center_dir="${TMP_DIR}/center"
    local center_binary="${center_dir}/idb"
    local center_service="${center_dir}/idb.service"
    local center_conf="${center_dir}/idb.conf"
    local center_home="${center_dir}/home"
    local center_plugins="${center_dir}/plugins"
    local config_target="${IDB_CONF_DIR}/idb.conf"

    if [[ ! -f "${center_binary}" || ! -f "${center_service}" || ! -f "${center_conf}" ]]; then
        log "center 安装包内容不完整"
        exit 1
    fi

    if [[ ! -f "${config_target}" ]]; then
        cp "${center_conf}" "${config_target}"
    fi
    Ensure_Config_Value "${config_target}" "host" "${TARGET_HOST}"
    Ensure_Config_Value "${config_target}" "port" "9918"
    Ensure_Config_Value "${config_target}" "github_repo" "sensdata/idb"

    cp "${center_binary}" "${IDB_INSTALL_ROOT}/idb"
    chmod +x "${IDB_INSTALL_ROOT}/idb"
    cp "${center_service}" /etc/systemd/system/idb.service
    ln -sf "${IDB_INSTALL_ROOT}/idb" /usr/local/bin/idb

    rm -rf "${IDB_HOME_DIR:?}"/*
    if [[ -d "${center_home}" ]]; then
        cp -R "${center_home}/." "${IDB_HOME_DIR}/"
    fi

    if [[ -d "${center_plugins}" ]]; then
        cp -R "${center_plugins}/." "${IDB_PLUGIN_DIR}/"
        find "${IDB_PLUGIN_DIR}" -type f -exec chmod +x {} \;
    fi
}

function Install_Agent() {
    local agent_dir="${TMP_DIR}/agent"
    local packaged_conf="${agent_dir}/idb-agent.conf"
    local installed_conf="${IDB_AGENT_CONF_DIR}/idb-agent.conf"

    if [[ ! -f "${agent_dir}/idb-agent" || ! -f "${agent_dir}/install-agent.sh" || ! -f "${packaged_conf}" ]]; then
        log "agent 安装包内容不完整"
        return 1
    fi

    cp "${TMP_DIR}/${AGENT_PKG}" "${IDB_AGENT_DIR}/idb-agent.tar.gz"
    echo "${VERSION}" > "${IDB_AGENT_DIR}/idb-agent.version"

    if [[ -f "${installed_conf}" ]]; then
        cp "${installed_conf}" "${packaged_conf}"
    fi

    chmod +x "${agent_dir}/install-agent.sh"
    (
        cd "${agent_dir}"
        bash ./install-agent.sh
    )
}

function Install_Agent_With_Warning() {
    if ! Install_Agent; then
        AGENT_INSTALL_FAILED="true"
        AGENT_INSTALL_ERROR="本机 agent 安装失败，center 已保留运行，请稍后在面板内重试或检查本机安装环境"
        log "${AGENT_INSTALL_ERROR}"
    fi
}

function Start_Services() {
    systemctl daemon-reload
    systemctl enable idb.service
    systemctl restart idb.service
}

function Show_Result() {
    local password_display

    password_display="${INITIAL_ADMIN_PASS:-}"
    if [[ -z "${password_display}" && -f "${IDB_CONF_DIR}/idb.env" ]]; then
        password_display=$(grep '^PASSWORD=' "${IDB_CONF_DIR}/idb.env" 2>/dev/null | cut -d'=' -f2)
    fi

    log ""
    log "=================感谢您的耐心等待，安装已经完成=================="
    log ""
    log "请用浏览器访问面板:"
    if [[ "${PUBLIC_IP}" != "N/A" ]]; then
        log "外网地址: http://${PUBLIC_IP}:9918"
    fi
    log "内网地址: http://${LOCAL_IP}:9918"
    log "初始用户: admin"
    log "初始密码: ${password_display}"
    log ""
    log "项目官网: https://idb.net"
    log "项目文档: https://idb.net/docs"
    log "代码仓库: https://github.com/sensdata/idb"
    log ""
    if [[ "${AGENT_INSTALL_FAILED}" == "true" ]]; then
        log "注意事项: ${AGENT_INSTALL_ERROR}"
        log ""
    fi
    log "如果使用的是云服务器，请至安全组开放 9918 端口"
    log ""
    log "为了您的服务器安全，在您离开此界面后您将无法再看到您的密码，请务必牢记您的密码。"
    log ""
    log "================================================================"
}

function Cleanup() {
    rm -rf "${TMP_DIR}"
}

function main() {
    trap Cleanup EXIT
    Check_Root
    Auto_Detect_Proxy
    Prepare_Download_Env
    Check_Architecture
    Detect_Version "${1:-}"
    Download_Assets
    Ensure_Directories
    Ensure_Admin_Env
    Detect_Target_Host
    Install_Center
    Generate_Certs_If_Missing
    Start_Services
    Install_Agent_With_Warning
    Show_Result
}

main "${1:-}"
