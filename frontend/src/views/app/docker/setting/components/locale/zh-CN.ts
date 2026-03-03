export default {
  'app.docker.setting.status.running': '已启动',
  'app.docker.setting.status.stopped': '未启动',
  'app.docker.setting.status.stop': '停止',
  'app.docker.setting.status.start': '启动',
  'app.docker.setting.status.restart': '重启',
  'app.docker.setting.status.stop.help': '停止 Docker 服务，容器会中断运行。',
  'app.docker.setting.status.start.help':
    '启动 Docker 服务并恢复容器管理能力。',
  'app.docker.setting.status.restart.help':
    '重启 Docker 服务，短暂中断后恢复。',
  'app.docker.setting.version': '版本',
  'app.docker.setting.mode.form': '编辑表单',
  'app.docker.setting.mode.file': '编辑配置文件',
  'app.docker.setting.mode.help':
    '常规配置建议使用表单；复杂或批量调整建议直接编辑 daemon.json。',
  'app.docker.setting.iptables.help':
    '控制 Docker 是否自动维护 iptables 规则，关闭后需自行维护转发/NAT。',
  'app.docker.setting.live_restore.help':
    '开启后 Docker 重启时尽量保持容器继续运行（部分场景可能不生效）。',
  'app.docker.setting.cgroup_driver.help':
    '建议与宿主机 init/system 配置保持一致，配置不一致可能导致容器异常。',
  'app.docker.setting.mirror.empty': '未设置',
  'app.docker.setting.registry.empty': '未设置',
  'app.docker.setting.ipv6.advanced': '高级设置',
  'common.message.formatError': '格式错误',

  'app.docker.setting.log.title': '日志切割',
  'app.docker.setting.log.max_size': '文件大小',
  'app.docker.setting.log.max_size.placeholder':
    '请输入日志文件最大大小，如 10m/100m/1g',
  'app.docker.setting.log.max_size.help': '支持 k/m/g 或 KB/MB/GB，示例: 100m',
  'app.docker.setting.log.max_size.required': '请输入日志文件最大大小',
  'app.docker.setting.log.max_size.format':
    '格式错误，仅支持数字+单位(k/m/g/B/KB/MB/GB)',
  'app.docker.setting.log.max_file': '保留份数',
  'app.docker.setting.log.max_file.placeholder': '请输入日志文件保留份数',
  'app.docker.setting.log.max_file.help': '仅支持正整数，示例: 3',
  'app.docker.setting.log.max_file.required': '请输入日志文件保留份数',
  'app.docker.setting.log.max_file.format': '格式错误，仅支持正整数',
  'app.docker.setting.log.confirm.title': '配置修改',
  'app.docker.setting.log.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
  'app.docker.setting.log.guide.title': '日志切割说明',
  'app.docker.setting.log.guide.desc':
    '限制单个容器日志文件大小和保留份数，避免磁盘被日志占满。保存后会重启 Docker 生效。',

  'app.docker.setting.mirror.title': '镜像加速器',
  'app.docker.setting.mirror.mirrors': '加速器地址',
  'app.docker.setting.mirror.mirrors.placeholder':
    '每行一个加速器地址，如 https://mirror.ccs.tencentyun.com',
  'app.docker.setting.mirror.mirrors.help':
    '可配置多个地址，按顺序尝试，通常用于提升拉取速度。',
  'app.docker.setting.mirror.mirrors.required': '请输入加速器地址',
  'app.docker.setting.mirror.mirrors.format':
    '格式错误，仅支持 http(s):// 或 docker:// 开头',
  'app.docker.setting.mirror.confirm.title': '配置修改',
  'app.docker.setting.mirror.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
  'app.docker.setting.mirror.guide.title': '镜像加速说明',
  'app.docker.setting.mirror.guide.desc':
    '适合网络受限场景，可显著提升拉取镜像速度。建议使用可信镜像加速源。',

  'app.docker.setting.registry.title': '私有仓库',
  'app.docker.setting.registry.registries': '仓库地址',
  'app.docker.setting.registry.registries.placeholder':
    '每行一个仓库地址，如 my-registry.com:5000',
  'app.docker.setting.registry.registries.help':
    '仅在使用 HTTP 或自签名仓库时配置，生产环境建议优先使用 HTTPS。',
  'app.docker.setting.registry.registries.required': '请输入仓库地址',
  'app.docker.setting.registry.registries.format':
    '格式错误，仅支持 http(s)://、IP:PORT、域名:PORT',
  'app.docker.setting.registry.confirm.title': '配置修改',
  'app.docker.setting.registry.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
  'app.docker.setting.registry.guide.title': '私有仓库说明',
  'app.docker.setting.registry.guide.desc':
    '此项会将仓库加入 insecure-registries，请确认网络与证书策略符合你的安全要求。',

  'app.docker.setting.ipv6.title': 'IPv6 配置',
  'app.docker.setting.ipv6.fixed_cidr_v6': '子网',
  'app.docker.setting.ipv6.fixed_cidr_v6.help':
    '格式为 IPv6 CIDR，例如 fd00::/64。',
  'app.docker.setting.ipv6.ip6_tables': 'ip6tables',
  'app.docker.setting.ipv6.experimental': 'experimental',
  'app.docker.setting.ipv6.confirm.title': '配置修改',
  'app.docker.setting.ipv6.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
  'app.docker.setting.ipv6.guide.title': 'IPv6 说明',
  'app.docker.setting.ipv6.guide.desc':
    '启用后 Docker 会分配 IPv6 地址，请确保主机与上游网络具备 IPv6 路由能力。',

  'app.docker.setting.socketPath.title': 'Socket 路径',
  'app.docker.setting.socketPath.socket_path': 'Socket 路径',
  'app.docker.setting.socketPath.socket_path.placeholder':
    '请输入 Docker Socket 路径，如 /var/run/docker.sock 或 tcp://host:port',
  'app.docker.setting.socketPath.socket_path.help':
    '本地建议使用 /var/run/docker.sock；远程 TCP 请仅在受控网络中使用。',
  'app.docker.setting.socketPath.socket_path.required': '请输入 Socket 路径',
  'app.docker.setting.socketPath.socket_path.format':
    '格式错误，仅支持 unix 路径或 tcp://host:port',
  'app.docker.setting.socketPath.confirm.title': '配置修改',
  'app.docker.setting.socketPath.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
  'app.docker.setting.socketPath.guide.title': 'Socket 路径说明',
  'app.docker.setting.socketPath.guide.desc':
    '该配置会更新 daemon.json 的 hosts 字段。错误配置可能导致 Docker 无法启动，请谨慎修改。',
};
