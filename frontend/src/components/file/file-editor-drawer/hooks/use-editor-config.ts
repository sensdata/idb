import { computed, Ref } from 'vue';
import { javascript } from '@codemirror/lang-javascript';
import { json } from '@codemirror/lang-json';
import { html } from '@codemirror/lang-html';
import { css } from '@codemirror/lang-css';
import { markdown } from '@codemirror/lang-markdown';
import { StreamLanguage } from '@codemirror/language';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { properties } from '@codemirror/legacy-modes/mode/properties';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { toml } from '@codemirror/legacy-modes/mode/toml';
import { simpleMode } from '@codemirror/legacy-modes/mode/simple-mode';
import { EditorState } from '@codemirror/state';
import { FileItem } from '../types';

// 常量定义 - 提取到文件顶部以提高可维护性
const BASH_FILES = [
  '.bashrc',
  '.bash_profile',
  '.bash_login',
  '.bash_logout',
  '.profile',
] as const;

// 文件扩展名到语言的映射
const EXTENSION_MAP = {
  // JavaScript/TypeScript
  js: 'javascript',
  jsx: 'javascript',
  ts: 'javascript',
  tsx: 'javascript',

  // Web
  json: 'json',
  html: 'html',
  htm: 'html',
  vue: 'html',
  css: 'css',
  scss: 'css',
  less: 'css',

  // Documentation
  md: 'markdown',
  markdown: 'markdown',

  // Shell
  sh: 'shell',
  bash: 'shell',

  // Configuration
  yaml: 'yaml',
  yml: 'yaml',
  properties: 'properties',
  env: 'properties',
  conf: 'nginx',
  nginx: 'nginx',
  toml: 'toml',
  ini: 'toml',

  // Logs
  log: 'log',
} as const;

type SupportedExtension = keyof typeof EXTENSION_MAP;
type LanguageType = (typeof EXTENSION_MAP)[SupportedExtension];

// 日志文件的高亮规则
const logSyntax = {
  start: [
    // 匹配 ISO 格式日期 (2023-04-23T10:15:30)
    {
      regex:
        /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?/,
      token: 'number',
    },
    // 匹配常见日期格式 (yyyy-mm-dd, dd/mm/yyyy, mm/dd/yyyy)
    {
      regex: /\d{4}-\d{2}-\d{2}|\d{2}\/\d{2}\/\d{4}|\d{2}\.\d{2}\.\d{4}/,
      token: 'number',
    },
    // 匹配时间格式 (hh:mm:ss)
    { regex: /\d{2}:\d{2}(?::\d{2}(?:\.\d+)?)?/, token: 'number' },
    // 匹配括号中的内容 (通常是上下文信息)
    { regex: /\[[^\]]+\]|\([^)]+\)/, token: 'string' },
    // 匹配错误相关关键词
    { regex: /\b(?:error|fail|exception|fatal|critical)\b/i, token: 'error' },
    // 匹配警告相关关键词
    {
      regex: /\b(?:warning|warn|caution|alert|attention)\b/i,
      token: 'keyword',
    },
    // 匹配信息相关关键词
    { regex: /\b(?:info|information|notice|log)\b/i, token: 'comment' },
    // 匹配成功相关关键词
    {
      regex: /\b(?:success|successful|succeeded|completed|done|ok)\b/i,
      token: 'atom',
    },
    // 匹配 IP 地址
    { regex: /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/, token: 'def' },
    // 匹配引号中的内容
    { regex: /"(?:[^"\\]|\\.)*"/, token: 'string' },
    { regex: /'(?:[^'\\]|\\.)*'/, token: 'string' },
  ],
};

// logrotate配置文件的高亮规则
const logrotateSyntax = {
  start: [
    // 注释
    { regex: /#.*/, token: 'comment' },

    // 文件路径 (行首的路径)
    { regex: /^\/[^\s{]+/, token: 'string' },

    // 大括号
    { regex: /[{}]/, token: 'bracket' },

    // 频率关键词
    {
      regex: /\b(?:daily|weekly|monthly|yearly|hourly)\b/,
      token: 'keyword',
    },

    // rotate 指令
    { regex: /\brotate\s+\d+/, token: 'number' },

    // 布尔选项
    {
      regex:
        /\b(?:compress|delaycompress|missingok|notifempty|copytruncate|create|dateext|ifempty|nocompress|nocopytruncate|nocreate|nodateext|nodelaycompress|nomissingok|noolddir|nosharedscripts|olddir|sharedscripts|size|maxage|minsize|maxsize)\b/,
      token: 'atom',
    },

    // create 指令的权限和用户组
    { regex: /\bcreate\s+\d{3,4}\s+\w+\s+\w+/, token: 'def' },

    // size 指令
    { regex: /\bsize\s+\d+[kmgKMG]?/, token: 'number' },

    // maxage 指令
    { regex: /\bmaxage\s+\d+/, token: 'number' },

    // 脚本块关键词
    {
      regex: /\b(?:prerotate|postrotate|firstaction|lastaction|endscript)\b/,
      token: 'keyword',
    },

    // 数字
    { regex: /\b\d+\b/, token: 'number' },

    // 字符串 (引号包围)
    { regex: /"(?:[^"\\]|\\.)*"/, token: 'string' },
    { regex: /'(?:[^'\\]|\\.)*'/, token: 'string' },

    // 文件路径和通配符
    { regex: /\/[^\s]*/, token: 'string' },
    { regex: /\*+/, token: 'operator' },
  ],
};

// systemd service配置文件的高亮规则
const serviceSyntax = {
  start: [
    // 注释
    { regex: /#.*/, token: 'comment' },

    // 段标题 [Unit], [Service], [Install]
    { regex: /^\[[^\]]+\]/, token: 'header' },

    // systemd 配置键（主要的配置键名）
    {
      regex:
        /^(?:Description|Documentation|Requires|Wants|After|Before|Conflicts|ConditionPathExists|ConditionFileNotEmpty|AssertPathExists)(?=\s*=)/,
      token: 'keyword',
    },

    // Service 段的配置键
    {
      regex:
        /^(?:Type|PIDFile|BusName|ExecStart|ExecStartPre|ExecStartPost|ExecReload|ExecStop|ExecStopPost|RestartSec|TimeoutStartSec|TimeoutStopSec|TimeoutSec|Restart|SuccessExitStatus|RestartPreventExitStatus|RestartForceExitStatus|PermissionsStartOnly|RootDirectoryStartOnly|RemainAfterExit|GuessMainPID|KillMode|KillSignal|SendSIGKILL|SendSIGHUP|UMask|NotifyAccess|Sockets|StandardInput|StandardOutput|StandardError|TTYPath|TTYReset|TTYVHangup|TTYVTDisallocate|SyslogIdentifier|SyslogFacility|SyslogLevel|SyslogLevelPrefix|LogLevelMax|SecureBits|CapabilityBoundingSet|AmbientCapabilities|User|Group|DynamicUser|SupplementaryGroups|PAMName|WorkingDirectory|RootDirectory|Nice|OOMScoreAdjust|IOSchedulingClass|IOSchedulingPriority|CPUSchedulingPolicy|CPUSchedulingPriority|CPUSchedulingResetOnFork|CPUAffinity|LimitCPU|LimitFSIZE|LimitDATA|LimitSTACK|LimitCORE|LimitRSS|LimitNOFILE|LimitAS|LimitNPROC|LimitMEMLOCK|LimitLOCKS|LimitSIGPENDING|LimitMSGQUEUE|LimitNICE|LimitRTPRIO|LimitRTTIME|ReadWriteDirectories|ReadOnlyDirectories|InaccessibleDirectories|PrivateTmp|PrivateDevices|PrivateNetwork|ProtectSystem|ProtectHome|MountFlags|Environment|EnvironmentFile|PassEnvironment|UnsetEnvironment)(?=\s*=)/,
      token: 'keyword',
    },

    // Install 段的配置键
    {
      regex: /^(?:WantedBy|RequiredBy|Also|DefaultInstance)(?=\s*=)/,
      token: 'keyword',
    },

    // 其他通用配置键
    { regex: /^[A-Za-z][A-Za-z0-9]*(?=\s*=)/, token: 'property' },

    // 等号
    { regex: /=/, token: 'operator' },

    // 布尔值
    { regex: /\b(?:true|false|yes|no|on|off|1|0)\b/i, token: 'atom' },

    // 数字（包括带单位的时间值）
    { regex: /\b\d+[smhd]?\b/, token: 'number' },

    // 服务类型关键词
    {
      regex: /\b(?:simple|forking|oneshot|notify|dbus|idle)\b/,
      token: 'builtin',
    },

    // 重启策略关键词
    {
      regex:
        /\b(?:no|always|on-success|on-failure|on-abnormal|on-abort|on-watchdog)\b/,
      token: 'builtin',
    },

    // Kill 模式
    { regex: /\b(?:control-group|process|mixed|none)\b/, token: 'builtin' },

    // 标准流重定向
    {
      regex:
        /\b(?:inherit|null|tty|journal|syslog|kmsg|journal\+console|syslog\+console|kmsg\+console|socket|fd)\b/,
      token: 'builtin',
    },

    // 文件路径
    { regex: /\/[^\s]*/, token: 'string' },

    // 引号字符串
    { regex: /"(?:[^"\\]|\\.)*"/, token: 'string' },
    { regex: /'(?:[^'\\]|\\.)*'/, token: 'string' },

    // 变量引用 ${VAR} 或 $VAR
    { regex: /\$\{[^}]+\}|\$[A-Za-z_][A-Za-z0-9_]*/, token: 'variable' },

    // 常见的用户和组名
    {
      regex:
        /\b(?:root|www-data|nobody|daemon|apache|nginx|mysql|postgres|systemd-journal|systemd-network|systemd-resolve)\b/,
      token: 'def',
    },

    // systemd targets
    { regex: /\b[a-z-]+\.target\b/, token: 'tag' },

    // systemd 服务单位
    { regex: /\b[a-z-]+\.service\b/, token: 'tag' },

    // systemd 其他单位类型
    {
      regex:
        /\b[a-z-]+\.(?:socket|timer|mount|automount|swap|path|slice|scope)\b/,
      token: 'tag',
    },
  ],
};

// 创建语法高亮器
const logMode = simpleMode({ start: logSyntax.start });
const logrotateMode = simpleMode({ start: logrotateSyntax.start });
const serviceMode = simpleMode({ start: serviceSyntax.start });

// 统一的行分隔符配置，确保一致的换行符处理
const commonExtensions = [EditorState.lineSeparator.of('\n')] as const;

/**
 * 获取指定语言类型的扩展配置
 * @param languageType 语言类型
 * @returns CodeMirror扩展数组
 */
function getLanguageExtensions(languageType: LanguageType) {
  switch (languageType) {
    case 'javascript':
      return [javascript(), ...commonExtensions];
    case 'json':
      return [json(), ...commonExtensions];
    case 'html':
      return [html(), ...commonExtensions];
    case 'css':
      return [css(), ...commonExtensions];
    case 'markdown':
      return [markdown(), ...commonExtensions];
    case 'shell':
      return [StreamLanguage.define(shell), ...commonExtensions];
    case 'yaml':
      return [StreamLanguage.define(yaml), ...commonExtensions];
    case 'properties':
      return [StreamLanguage.define(properties), ...commonExtensions];
    case 'nginx':
      return [StreamLanguage.define(nginx), ...commonExtensions];
    case 'toml':
      return [StreamLanguage.define(toml), ...commonExtensions];
    case 'log':
      return [StreamLanguage.define(logMode), ...commonExtensions];
    default:
      return [...commonExtensions];
  }
}

/**
 * 检查文件名是否为bash配置文件
 * @param fileName 文件名
 * @returns 是否为bash配置文件
 */
function isBashConfigFile(fileName: string): boolean {
  return BASH_FILES.includes(fileName as (typeof BASH_FILES)[number]);
}

/**
 * 获取文件扩展名
 * @param fileName 文件名
 * @returns 文件扩展名或undefined
 */
function getFileExtension(fileName: string): SupportedExtension | undefined {
  const ext = fileName.split('.').pop()?.toLowerCase();
  return ext && ext in EXTENSION_MAP ? (ext as SupportedExtension) : undefined;
}

/**
 * 编辑器配置Hook
 * 根据文件类型自动配置CodeMirror编辑器的语法高亮
 *
 * @param file 文件对象的响应式引用
 * @returns 包含extensions和工具方法的配置对象
 */
export default function useEditorConfig(file: Ref<FileItem | null>) {
  const extensions = computed(() => {
    if (!file.value) {
      return [...commonExtensions];
    }

    const fileName = file.value.name.toLowerCase();

    // 特殊处理bash配置文件
    if (isBashConfigFile(fileName)) {
      return getLanguageExtensions('shell');
    }

    // 根据文件扩展名获取语言类型
    const ext = getFileExtension(fileName);
    if (!ext) {
      return [...commonExtensions];
    }

    const languageType = EXTENSION_MAP[ext];
    return getLanguageExtensions(languageType);
  });

  /**
   * 获取指定语言类型的扩展配置
   * @param languageType 语言类型
   * @returns CodeMirror扩展数组
   */
  const getExtensionsForLanguage = (languageType: LanguageType) => {
    return getLanguageExtensions(languageType);
  };

  /**
   * 获取logrotate配置文件的扩展配置
   * @returns logrotate语言扩展配置
   */
  const getLogrotateExtensions = () => {
    return [StreamLanguage.define(logrotateMode), ...commonExtensions];
  };

  /**
   * 获取systemd service配置文件的扩展配置
   * @returns service语言扩展配置
   */
  const getServiceExtensions = () => {
    return [StreamLanguage.define(serviceMode), ...commonExtensions];
  };

  return {
    extensions,
    getExtensionsForLanguage,
    getLogrotateExtensions,
    getServiceExtensions,
  } as const;
}
