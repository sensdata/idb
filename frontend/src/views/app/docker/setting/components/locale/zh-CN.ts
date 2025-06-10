export default {
  'app.docker.setting.status.running': '已启动',
  'app.docker.setting.status.stopped': '未启动',
  'app.docker.setting.status.stop': '停止',
  'app.docker.setting.status.start': '启动',
  'app.docker.setting.status.restart': '重启',
  'app.docker.setting.version': '版本: ',
  'app.docker.setting.mode.form': '编辑表单',
  'app.docker.setting.mode.file': '编辑配置文件',
  'app.docker.setting.mirror.empty': '未设置',
  'app.docker.setting.registry.empty': '未设置',
  'app.docker.setting.ipv6.advanced': '高级设置',
  'common.message.formatError': '格式错误',

  'app.docker.setting.log.title': '日志切割',
  'app.docker.setting.log.max_size': '文件大小',
  'app.docker.setting.log.max_size.placeholder':
    '请输入日志文件最大大小，如 10m/100m/1g',
  'app.docker.setting.log.max_size.required': '请输入日志文件最大大小',
  'app.docker.setting.log.max_size.format':
    '格式错误，仅支持数字+单位(k/m/g/B/KB/MB/GB)',
  'app.docker.setting.log.max_file': '保留份数',
  'app.docker.setting.log.max_file.placeholder': '请输入日志文件保留份数',
  'app.docker.setting.log.max_file.required': '请输入日志文件保留份数',
  'app.docker.setting.log.max_file.format': '格式错误，仅支持正整数',
  'app.docker.setting.log.confirm.title': '配置修改',
  'app.docker.setting.log.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',

  'app.docker.setting.mirror.title': '镜像加速器',
  'app.docker.setting.mirror.mirrors': '加速器地址',
  'app.docker.setting.mirror.mirrors.placeholder':
    '每行一个加速器地址，如 https://mirror.ccs.tencentyun.com',
  'app.docker.setting.mirror.mirrors.required': '请输入加速器地址',
  'app.docker.setting.mirror.mirrors.format':
    '格式错误，仅支持 http(s):// 或 docker:// 开头',
  'app.docker.setting.mirror.confirm.title': '配置修改',
  'app.docker.setting.mirror.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',

  'app.docker.setting.registry.title': '私有仓库',
  'app.docker.setting.registry.registries': '仓库地址',
  'app.docker.setting.registry.registries.placeholder':
    '每行一个仓库地址，如 my-registry.com:5000',
  'app.docker.setting.registry.registries.required': '请输入仓库地址',
  'app.docker.setting.registry.registries.format':
    '格式错误，仅支持 http(s)://、IP:PORT、域名:PORT',
  'app.docker.setting.registry.confirm.title': '配置修改',
  'app.docker.setting.registry.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',

  'app.docker.setting.ipv6.title': 'IPv6 配置',
  'app.docker.setting.ipv6.fixed_cidr_v6': '子网',
  'app.docker.setting.ipv6.ip6_tables': 'ip6tables',
  'app.docker.setting.ipv6.experimental': 'experimental',
  'app.docker.setting.ipv6.confirm.title': '配置修改',
  'app.docker.setting.ipv6.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',

  'app.docker.setting.socketPath.title': 'Socket 路径',
  'app.docker.setting.socketPath.socket_path': 'Socket 路径',
  'app.docker.setting.socketPath.socket_path.placeholder':
    '请输入 Docker Socket 路径，如 /var/run/docker.sock 或 tcp://host:port',
  'app.docker.setting.socketPath.socket_path.required': '请输入 Socket 路径',
  'app.docker.setting.socketPath.socket_path.format':
    '格式错误，仅支持 unix 路径或 tcp://host:port',
  'app.docker.setting.socketPath.confirm.title': '配置修改',
  'app.docker.setting.socketPath.confirm.content':
    '修改配置后需要重启生效，确认继续操作吗？',
};
