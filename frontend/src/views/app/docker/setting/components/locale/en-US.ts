export default {
  'app.docker.setting.status.running': 'Running',
  'app.docker.setting.status.stopped': 'Stopped',
  'app.docker.setting.status.stop': 'Stop',
  'app.docker.setting.status.start': 'Start',
  'app.docker.setting.status.restart': 'Restart',
  'app.docker.setting.status.stop.help':
    'Stop Docker service. Running containers will be interrupted.',
  'app.docker.setting.status.start.help':
    'Start Docker service and restore container management capability.',
  'app.docker.setting.status.restart.help':
    'Restart Docker service. Containers may be interrupted briefly.',
  'app.docker.setting.version': 'Version',
  'app.docker.setting.mode.form': 'Edit Form',
  'app.docker.setting.mode.file': 'Edit File',
  'app.docker.setting.mode.help':
    'Use form mode for common settings, and file mode for advanced daemon.json edits.',
  'app.docker.setting.iptables.help':
    'Controls whether Docker manages iptables rules automatically.',
  'app.docker.setting.live_restore.help':
    'When enabled, containers try to keep running across Docker daemon restarts.',
  'app.docker.setting.cgroup_driver.help':
    'Keep this aligned with host init/system configuration to avoid runtime issues.',
  'app.docker.setting.mirror.empty': 'Not Set',
  'app.docker.setting.registry.empty': 'Not Set',
  'app.docker.setting.ipv6.advanced': 'Advanced',
  'common.message.formatError': 'Format Error',

  'app.docker.setting.log.title': 'Log Rotation',
  'app.docker.setting.log.max_size': 'File Size',
  'app.docker.setting.log.max_size.placeholder':
    'Enter max log file size, e.g. 10m/100m/1g',
  'app.docker.setting.log.max_size.help':
    'Supports k/m/g or KB/MB/GB, e.g. 100m.',
  'app.docker.setting.log.max_size.required': 'Please enter max log file size',
  'app.docker.setting.log.max_size.format':
    'Invalid format, only numbers + unit (k/m/g/B/KB/MB/GB) allowed',
  'app.docker.setting.log.max_file': 'Retain Count',
  'app.docker.setting.log.max_file.placeholder':
    'Enter number of log files to retain',
  'app.docker.setting.log.max_file.help': 'Positive integer only, e.g. 3.',
  'app.docker.setting.log.max_file.required': 'Please enter retain count',
  'app.docker.setting.log.max_file.format':
    'Invalid format, only positive integers allowed',
  'app.docker.setting.log.confirm.title': 'Configuration Change',
  'app.docker.setting.log.confirm.content':
    'Changes require a restart to take effect. Continue?',
  'app.docker.setting.log.guide.title': 'Log rotation tips',
  'app.docker.setting.log.guide.desc':
    'Limit container log size and retention count to prevent disk exhaustion. Docker restart is required after saving.',

  'app.docker.setting.mirror.title': 'Registry Mirror',
  'app.docker.setting.mirror.mirrors': 'Mirror Address',
  'app.docker.setting.mirror.mirrors.placeholder':
    'One per line, e.g. https://mirror.ccs.tencentyun.com',
  'app.docker.setting.mirror.mirrors.help':
    'Multiple mirror addresses are supported and tried in order.',
  'app.docker.setting.mirror.mirrors.required': 'Please enter mirror address',
  'app.docker.setting.mirror.mirrors.format':
    'Invalid format, must start with http(s):// or docker://',
  'app.docker.setting.mirror.confirm.title': 'Configuration Change',
  'app.docker.setting.mirror.confirm.content':
    'Changes require a restart to take effect. Continue?',
  'app.docker.setting.mirror.guide.title': 'Registry mirror tips',
  'app.docker.setting.mirror.guide.desc':
    'Useful in restricted networks to speed up image pulls. Prefer trusted mirror providers.',

  'app.docker.setting.registry.title': 'Private Registry',
  'app.docker.setting.registry.registries': 'Registry Address',
  'app.docker.setting.registry.registries.placeholder':
    'One per line, e.g. my-registry.com:5000',
  'app.docker.setting.registry.registries.help':
    'Use this only for HTTP or self-signed registries. HTTPS is recommended in production.',
  'app.docker.setting.registry.registries.required':
    'Please enter registry address',
  'app.docker.setting.registry.registries.format':
    'Invalid format, only http(s)://, IP:PORT, or domain:PORT allowed',
  'app.docker.setting.registry.confirm.title': 'Configuration Change',
  'app.docker.setting.registry.confirm.content':
    'Changes require a restart to take effect. Continue?',
  'app.docker.setting.registry.guide.title': 'Private registry tips',
  'app.docker.setting.registry.guide.desc':
    'This updates insecure-registries. Ensure your network and certificate policy allows it.',

  'app.docker.setting.ipv6.title': 'IPv6 Settings',
  'app.docker.setting.ipv6.fixed_cidr_v6': 'Subnet',
  'app.docker.setting.ipv6.fixed_cidr_v6.help':
    'Use IPv6 CIDR format, e.g. fd00::/64.',
  'app.docker.setting.ipv6.ip6_tables': 'ip6tables',
  'app.docker.setting.ipv6.experimental': 'experimental',
  'app.docker.setting.ipv6.confirm.title': 'Configuration Change',
  'app.docker.setting.ipv6.confirm.content':
    'Changes require a restart to take effect. Continue?',
  'app.docker.setting.ipv6.guide.title': 'IPv6 tips',
  'app.docker.setting.ipv6.guide.desc':
    'After enabling IPv6, make sure the host and upstream network support IPv6 routing.',

  'app.docker.setting.socketPath.title': 'Socket Path',
  'app.docker.setting.socketPath.socket_path': 'Socket Path',
  'app.docker.setting.socketPath.socket_path.placeholder':
    'Enter Docker Socket path, e.g. /var/run/docker.sock or tcp://host:port',
  'app.docker.setting.socketPath.socket_path.help':
    'Use /var/run/docker.sock for local access. Use TCP only in trusted networks.',
  'app.docker.setting.socketPath.socket_path.required':
    'Please enter socket path',
  'app.docker.setting.socketPath.socket_path.format':
    'Invalid format, only unix path or tcp://host:port allowed',
  'app.docker.setting.socketPath.confirm.title': 'Configuration Change',
  'app.docker.setting.socketPath.confirm.content':
    'Changes require a restart to take effect. Continue?',
  'app.docker.setting.socketPath.guide.title': 'Socket path tips',
  'app.docker.setting.socketPath.guide.desc':
    'This updates the hosts field in daemon.json. Invalid values may prevent Docker from starting.',
};
