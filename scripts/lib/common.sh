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

function Fetch_Latest_Release_Tag() {
    local api_url="$1"
    local latest_url="${api_url}/repos/sensdata/idb/releases/latest?ts=$(date +%s)"

    curl -fsSL \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Cache-Control: no-cache" \
        -H "Pragma: no-cache" \
        "${latest_url}" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
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
    VERSION="${1:-}"
    if [[ -z "${VERSION}" ]]; then
        VERSION=$(Fetch_Latest_Release_Tag "${GITHUB_API_URL}")
    fi
    if [[ -z "${VERSION}" ]]; then
        log "获取目标版本失败"
        exit 1
    fi

    CENTER_PKG="idb_${VERSION}_linux_${IDB_ARCH}.tar.gz"
    AGENT_PKG="idb-agent_${VERSION}_linux_${IDB_ARCH}.tar.gz"
    CENTER_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${CENTER_PKG}"
    AGENT_URL="${GITHUB_RELEASES_URL}/sensdata/idb/releases/download/${VERSION}/${AGENT_PKG}"
}

function Download_Assets() {
    rm -rf "${TMP_DIR}"
    mkdir -p "${TMP_DIR}/center" "${TMP_DIR}/agent"

    log "下载 center 安装包: ${CENTER_PKG}"
    curl -fsSL "${CENTER_URL}" -o "${TMP_DIR}/${CENTER_PKG}"

    log "下载 agent 安装包: ${AGENT_PKG}"
    curl -fsSL "${AGENT_URL}" -o "${TMP_DIR}/${AGENT_PKG}"

    tar -xzf "${TMP_DIR}/${CENTER_PKG}" -C "${TMP_DIR}/center"
    tar -xzf "${TMP_DIR}/${AGENT_PKG}" -C "${TMP_DIR}/agent"
}

function Ensure_Config_Value() {
    local file="$1"
    local key="$2"
    local value="$3"

    if grep -q "^${key}=" "${file}" 2>/dev/null; then
        local safe_value
        safe_value=$(printf '%s' "$value" | sed 's/[#&\\/]/\\&/g')
        sed -i "s#^${key}=.*#${key}=${safe_value}#" "${file}"
    else
        printf '%s=%s\n' "${key}" "${value}" >> "${file}"
    fi
}

function Generate_Certs_If_Missing() {
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
