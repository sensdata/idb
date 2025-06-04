/**
 * systemd服务配置相关常量
 */

// 配置段落名称
export const SECTION_UNIT = 'unit';
export const SECTION_SERVICE = 'service';
export const SECTION_INSTALL = 'install';

// Unit段落的字段（基于 systemd 官方文档）
export const UNIT_FIELDS = [
  'description',
  'documentation',
  'after',
  'before',
  'requires',
  'wants',
  'conflicts',
  'bindsto',
  'partof',
  'requisite',
  'upholds',
  'onfailure',
  'onsuccess',
  'propagatesreloadto',
  'reloadpropagatedfrom',
  'propagatesstopsto',
  'stoppropagatedfrom',
  'joinsnamespaceOf',
  'requiresmountsfor',
  'wantsmountsfor',
];

// Install段落的字段（基于 systemd 官方文档）
export const INSTALL_FIELDS = [
  'wantedby',
  'requiredby',
  'alias',
  'also',
  'defaultinstance',
  'upheldby',
];

// 服务重启选项（官方完整列表）
export const RESTART_OPTIONS = [
  'no',
  'always',
  'on-success',
  'on-failure',
  'on-abnormal',
  'on-abort',
  'on-watchdog',
];

// 服务类型选项（包含最新的 notify-reload 和 exec）
export const SERVICE_TYPES = [
  'simple',
  'exec',
  'forking',
  'oneshot',
  'dbus',
  'notify',
  'notify-reload',
  'idle',
];

// 退出类型选项
export const EXIT_TYPES = ['main', 'cgroup'];

// 重启模式选项
export const RESTART_MODES = ['normal', 'direct', 'debug'];

// 超时失败模式
export const TIMEOUT_FAILURE_MODES = ['terminate', 'abort', 'kill'];

// 通知访问权限
export const NOTIFY_ACCESS_OPTIONS = ['none', 'main', 'exec', 'all'];

// OOM 策略选项
export const OOM_POLICY_OPTIONS = ['continue', 'stop', 'kill'];

// 默认超时时间(秒)
export const DEFAULT_TIMEOUT = 90;

// 数组类型字段（需要特殊处理的字段）
export const ARRAY_FIELDS = [
  'after',
  'before',
  'requires',
  'wants',
  'conflicts',
  'bindsto',
  'partof',
  'requisite',
  'upholds',
  'onfailure',
  'onsuccess',
  'wantedby',
  'requiredby',
  'upheldby',
  'alias',
  'also',
  'environment',
  'requiresmountsfor',
  'wantsmountsfor',
];

// 默认服务模板
export const DEFAULT_SERVICE_TEMPLATE = `[Unit]
Description=

[Service]
Type=simple
ExecStart=
WorkingDirectory=
User=root
Group=root

[Install]
WantedBy=multi-user.target
`;

// 常用的 systemd 目标单元
export const COMMON_TARGETS = [
  'multi-user.target',
  'graphical.target',
  'default.target',
  'basic.target',
  'sysinit.target',
  'local-fs.target',
  'remote-fs.target',
  'network.target',
  'network-online.target',
  'time-sync.target',
];

// 常用的系统服务依赖
export const COMMON_SYSTEM_SERVICES = [
  'dbus.socket',
  'dbus.service',
  'systemd-journald.service',
  'systemd-logind.service',
  'network.service',
  'NetworkManager.service',
  'sshd.service',
];
