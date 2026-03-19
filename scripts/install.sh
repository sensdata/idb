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

function Probe_HTTP_200() {
    local url="$1"
    local attempts="${2:-3}"
    local i
    local code

    for ((i = 1; i <= attempts; i++)); do
        code=$(curl -s --connect-timeout 5 --max-time 10 -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
        if [[ "$code" == "200" ]]; then
            return 0
        fi
        sleep 1
    done

    return 1
}

function Is_China_Region() {
    local country
    local tz

    country=$(curl -s --connect-timeout 3 --max-time 5 https://ipinfo.io/country 2>/dev/null | tr -d '[:space:]')
    if [[ "$country" == "CN" || "$country" == "HK" || "$country" == "MO" ]]; then
        return 0
    fi

    tz=$(timedatectl 2>/dev/null | grep "Time zone" | awk '{print $3}')
    [[ -z "$tz" ]] && tz="${TZ:-}"
    if [[ "$tz" == *"Shanghai"* || "$tz" == *"Chongqing"* || "$tz" == *"Hong_Kong"* ]]; then
        return 0
    fi
    return 1
}

function Auto_Detect_Proxy() {
    if [[ -n "${IDB_GITHUB_PROXY:-}" ]]; then
        log "使用指定代理: ${IDB_GITHUB_PROXY}"
        return
    fi

    if Is_China_Region; then
        log "检测到中国区域，优先检测加速代理..."
        if Probe_HTTP_200 "${IDB_DEFAULT_PROXY}/github-api/repos/sensdata/idb/releases/latest"; then
            export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
            log "加速代理可用: ${IDB_GITHUB_PROXY}"
            return
        fi
    fi

    if Probe_HTTP_200 "https://api.github.com/repos/sensdata/idb/releases/latest"; then
        log "GitHub 直连正常"
        return
    fi

    if Probe_HTTP_200 "${IDB_DEFAULT_PROXY}/github-api/repos/sensdata/idb/releases/latest"; then
        export IDB_GITHUB_PROXY="${IDB_DEFAULT_PROXY}"
        log "GitHub 直连不可用，切换到加速代理: ${IDB_GITHUB_PROXY}"
        return
    fi

    log "所有源均不可用，后续下载可能失败"
}

function Check_Root() {
    if [[ $EUID -ne 0 ]]; then
        log "请使用 root 或 sudo 权限运行此脚本"
        exit 1
    fi
}

function Check_Architecture() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64)
            IDB_ARCH="amd64"
            ;;
        *)
            log "暂不支持的系统架构: ${arch}"
            exit 1
            ;;
    esac
}

function Prepare_Download_Env() {
    GITHUB_API_URL="${IDB_GITHUB_PROXY:+${IDB_GITHUB_PROXY}/github-api}"
    GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"
    GITHUB_RELEASES_URL="${IDB_GITHUB_PROXY:+${IDB_GITHUB_PROXY}/github-releases}"
    GITHUB_RELEASES_URL="${GITHUB_RELEASES_URL:-https://github.com}"

    if [[ -n "${IDB_GITHUB_PROXY:-}" ]]; then
        log "使用加速代理: ${IDB_GITHUB_PROXY}"
    fi
}

function Detect_Version() {
    VERSION=$(curl -s "${GITHUB_API_URL}/repos/sensdata/idb/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [[ -z "${VERSION}" ]]; then
        log "获取最新版本失败"
        exit 1
    fi

    CENTER_PKG="idb_${VERSION}_linux_${IDB_ARCH}.tar.gz"
    AGENT_PKG="idb-agent_${VERSION}_linux_${IDB_ARCH}.tar.gz"
    CENTER_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${CENTER_PKG}"
    AGENT_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${AGENT_PKG}"
}

function Download_Assets() {
    rm -rf "${TMP_DIR}"
    mkdir -p "${TMP_DIR}"

    log "下载 center 安装包: ${CENTER_PKG}"
    curl -fsSL "${CENTER_URL}" -o "${TMP_DIR}/${CENTER_PKG}"

    log "下载 agent 安装包: ${AGENT_PKG}"
    curl -fsSL "${AGENT_URL}" -o "${TMP_DIR}/${AGENT_PKG}"

    log "解压 center 安装包"
    mkdir -p "${TMP_DIR}/center"
    tar -xzf "${TMP_DIR}/${CENTER_PKG}" -C "${TMP_DIR}/center"

    log "解压 agent 安装包"
    mkdir -p "${TMP_DIR}/agent"
    tar -xzf "${TMP_DIR}/${AGENT_PKG}" -C "${TMP_DIR}/agent"
}

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
        admin_pass=$(head -c 100 /dev/urandom | tr -dc 'a-z0-9' | head -c 8)
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

    local_ip=$(ip -4 route get 8.8.8.8 2>/dev/null | grep -oE 'src [0-9.]*' | awk '{print $2}' | head -n1 || true)

    TARGET_HOST="${public_ip:-${local_ip:-127.0.0.1}}"
    LOCAL_IP="${local_ip:-127.0.0.1}"
    PUBLIC_IP="${public_ip:-N/A}"
}

function Ensure_Config_Value() {
    local file="$1"
    local key="$2"
    local value="$3"

    if grep -q "^${key}=" "${file}"; then
        sed -i "s#^${key}=.*#${key}=${value}#" "${file}"
    else
        echo "${key}=${value}" >> "${file}"
    fi
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

function Generate_Certs() {
    local cert_file="${IDB_INSTALL_ROOT}/cert.pem"
    local key_file="${IDB_INSTALL_ROOT}/key.pem"

    if [[ -f "${cert_file}" && -f "${key_file}" ]]; then
        return
    fi

    log "生成本地自签名证书"
    openssl req -x509 -nodes -days 3650 \
        -newkey rsa:2048 \
        -keyout "${key_file}" \
        -out "${cert_file}" \
        -subj "/CN=localhost" >/dev/null 2>&1
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
    Detect_Version
    Download_Assets
    Ensure_Directories
    Ensure_Admin_Env
    Detect_Target_Host
    Install_Center
    Generate_Certs
    Start_Services
    Install_Agent_With_Warning
    Show_Result
}

main "$@"
