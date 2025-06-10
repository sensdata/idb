export default {
  'app.docker.setting.status.running': 'Running',
  'app.docker.setting.status.stopped': 'Stopped',
  'app.docker.setting.status.stop': 'Stop',
  'app.docker.setting.status.start': 'Start',
  'app.docker.setting.status.restart': 'Restart',
  'app.docker.setting.version': 'Version: ',
  'app.docker.setting.mode.form': 'Edit Form',
  'app.docker.setting.mode.file': 'Edit File',
  'app.docker.setting.mirror.empty': 'Not Set',
  'app.docker.setting.registry.empty': 'Not Set',
  'app.docker.setting.ipv6.advanced': 'Advanced',
  'common.message.formatError': 'Format Error',

  'app.docker.setting.log.title': 'Log Rotation',
  'app.docker.setting.log.max_size': 'File Size',
  'app.docker.setting.log.max_size.placeholder':
    'Enter max log file size, e.g. 10m/100m/1g',
  'app.docker.setting.log.max_size.required': 'Please enter max log file size',
  'app.docker.setting.log.max_size.format':
    'Invalid format, only numbers + unit (k/m/g/B/KB/MB/GB) allowed',
  'app.docker.setting.log.max_file': 'Retain Count',
  'app.docker.setting.log.max_file.placeholder':
    'Enter number of log files to retain',
  'app.docker.setting.log.max_file.required': 'Please enter retain count',
  'app.docker.setting.log.max_file.format':
    'Invalid format, only positive integers allowed',
  'app.docker.setting.log.confirm.title': 'Configuration Change',
  'app.docker.setting.log.confirm.content':
    'Changes require a restart to take effect. Continue?',

  'app.docker.setting.mirror.title': 'Registry Mirror',
  'app.docker.setting.mirror.mirrors': 'Mirror Address',
  'app.docker.setting.mirror.mirrors.placeholder':
    'One per line, e.g. https://mirror.ccs.tencentyun.com',
  'app.docker.setting.mirror.mirrors.required': 'Please enter mirror address',
  'app.docker.setting.mirror.mirrors.format':
    'Invalid format, must start with http(s):// or docker://',
  'app.docker.setting.mirror.confirm.title': 'Configuration Change',
  'app.docker.setting.mirror.confirm.content':
    'Changes require a restart to take effect. Continue?',

  'app.docker.setting.registry.title': 'Private Registry',
  'app.docker.setting.registry.registries': 'Registry Address',
  'app.docker.setting.registry.registries.placeholder':
    'One per line, e.g. my-registry.com:5000',
  'app.docker.setting.registry.registries.required':
    'Please enter registry address',
  'app.docker.setting.registry.registries.format':
    'Invalid format, only http(s)://, IP:PORT, or domain:PORT allowed',
  'app.docker.setting.registry.confirm.title': 'Configuration Change',
  'app.docker.setting.registry.confirm.content':
    'Changes require a restart to take effect. Continue?',

  'app.docker.setting.ipv6.title': 'IPv6 Settings',
  'app.docker.setting.ipv6.fixed_cidr_v6': 'Subnet',
  'app.docker.setting.ipv6.ip6_tables': 'ip6tables',
  'app.docker.setting.ipv6.experimental': 'experimental',
  'app.docker.setting.ipv6.confirm.title': 'Configuration Change',
  'app.docker.setting.ipv6.confirm.content':
    'Changes require a restart to take effect. Continue?',

  'app.docker.setting.socketPath.title': 'Socket Path',
  'app.docker.setting.socketPath.socket_path': 'Socket Path',
  'app.docker.setting.socketPath.socket_path.placeholder':
    'Enter Docker Socket path, e.g. /var/run/docker.sock or tcp://host:port',
  'app.docker.setting.socketPath.socket_path.required':
    'Please enter socket path',
  'app.docker.setting.socketPath.socket_path.format':
    'Invalid format, only unix path or tcp://host:port allowed',
  'app.docker.setting.socketPath.confirm.title': 'Configuration Change',
  'app.docker.setting.socketPath.confirm.content':
    'Changes require a restart to take effect. Continue?',
};
