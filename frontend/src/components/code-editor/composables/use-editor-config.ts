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

interface FileItem {
  name: string;
  path: string;
  size?: number;
}

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

// nftables配置文件的语法高亮规则
const nftablesSyntax = {
  start: [
    // 注释
    { regex: /#.*/, token: 'comment' },

    // 表、链、规则等关键词
    {
      regex:
        /\b(?:table|chain|rule|set|map|element|flush|add|insert|replace|delete|list|reset|export|monitor)\b/,
      token: 'keyword',
    },

    // 表类型
    {
      regex: /\b(?:inet|ip|ip6|bridge|netdev|arp)\b/,
      token: 'atom',
    },

    // 链类型和优先级
    {
      regex: /\b(?:type|hook|priority|policy)\b/,
      token: 'property',
    },

    // 钩子类型
    {
      regex:
        /\b(?:prerouting|input|forward|output|postrouting|ingress|egress)\b/,
      token: 'def',
    },

    // 动作
    {
      regex:
        /\b(?:accept|drop|reject|queue|continue|return|jump|goto|masquerade|redirect|snat|dnat)\b/,
      token: 'builtin',
    },

    // 协议和服务
    {
      regex:
        /\b(?:tcp|udp|icmp|icmpv6|esp|ah|sctp|dccp|http|https|ssh|ftp|smtp|dns|dhcp|ntp|snmp)\b/,
      token: 'variable-2',
    },

    // 匹配条件
    {
      regex:
        /\b(?:iif|oif|iifname|oifname|ip|ip6|tcp|udp|icmp|icmpv6|ether|vlan|arp|rt|ct|meta|socket|osf|fib)\b/,
      token: 'variable',
    },

    // 端口和地址
    { regex: /\b\d{1,5}(?:-\d{1,5})?\b/, token: 'number' },
    {
      regex: /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(?:\/\d{1,2})?\b/,
      token: 'string',
    },
    { regex: /\b[0-9a-fA-F:]+\/\d{1,3}\b/, token: 'string' },

    // 运算符
    { regex: /[<>=!&|]/, token: 'operator' },

    // 大括号和小括号
    { regex: /[{}()]/, token: 'bracket' },

    // 字符串
    { regex: /"(?:[^"\\]|\\.)*"/, token: 'string' },
    { regex: /'(?:[^'\\]|\\.)*'/, token: 'string' },

    // 数字
    { regex: /\b\d+\b/, token: 'number' },
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
        /\b(?:Type|ExecStart|ExecReload|ExecStop|Restart|RestartSec|User|Group|WorkingDirectory|Environment|EnvironmentFile|TimeoutStartSec|TimeoutStopSec|RuntimeMaxSec|LimitNOFILE|LimitNPROC|PrivateTmp|NoNewPrivileges|ProtectSystem|ProtectHome|ReadWritePaths|ReadOnlyPaths|InaccessiblePaths|DynamicUser|SupplementaryGroups|Requires|Wants|After|Before|BindsTo|PartOf|Conflicts|Requisite|OnFailure|PropagatesReloadTo|ReloadPropagatedFrom|JoinsNamespaceOf|RequiresMountsFor|OnFailureJobMode|IgnoreOnIsolate|StopWhenUnneeded|RefuseManualStart|RefuseManualStop|AllowIsolate|DefaultDependencies|JobTimeoutSec|JobTimeoutAction|StartLimitIntervalSec|StartLimitBurst|StartLimitAction|RebootArgument|SourcePath|Description|Documentation|WantedBy|RequiredBy|Also|Alias)\b/,
      token: 'variable',
    },

    // systemd 特殊值
    {
      regex:
        /\b(?:simple|exec|forking|oneshot|dbus|notify|idle|always|on-success|on-failure|on-abnormal|on-watchdog|on-abort|true|false|yes|no|on|off|strict|full|read-only|tmpfs|multi-user\.target|graphical\.target|default\.target)\b/,
      token: 'atom',
    },

    // 赋值符号
    { regex: /=/, token: 'operator' },

    // 路径
    { regex: /\/[^\s]*/, token: 'string' },

    // 数字和时间单位
    { regex: /\b\d+[smhd]?\b/, token: 'number' },

    // 字符串
    { regex: /"(?:[^"\\]|\\.)*"/, token: 'string' },
    { regex: /'(?:[^'\\]|\\.)*'/, token: 'string' },
  ],
};

/**
 * 根据语言类型获取对应的扩展
 */
function getLanguageExtensions(languageType: LanguageType) {
  switch (languageType) {
    case 'javascript':
      return [javascript()];
    case 'json':
      return [json()];
    case 'html':
      return [html()];
    case 'css':
      return [css()];
    case 'markdown':
      return [markdown()];
    case 'shell':
      return [StreamLanguage.define(shell)];
    case 'yaml':
      return [StreamLanguage.define(yaml)];
    case 'properties':
      return [StreamLanguage.define(properties)];
    case 'nginx':
      return [StreamLanguage.define(nginx)];
    case 'toml':
      return [StreamLanguage.define(toml)];
    case 'log':
      return [StreamLanguage.define(simpleMode(logSyntax))];
    default:
      return [];
  }
}

/**
 * 判断是否为 bash 配置文件
 */
function isBashConfigFile(fileName: string): boolean {
  return BASH_FILES.includes(fileName as any);
}

/**
 * 获取文件扩展名
 */
function getFileExtension(fileName: string): SupportedExtension | undefined {
  const lastDotIndex = fileName.lastIndexOf('.');
  if (lastDotIndex === -1) return undefined;

  const extension = fileName.slice(lastDotIndex + 1).toLowerCase();
  return EXTENSION_MAP[extension as SupportedExtension]
    ? (extension as SupportedExtension)
    : undefined;
}

export default function useEditorConfig(file: Ref<FileItem | null>) {
  // Define helper functions before they are used
  const getExtensionsForLanguage = (languageType: LanguageType) => {
    return getLanguageExtensions(languageType);
  };

  const getLogrotateExtensions = () => {
    return [
      StreamLanguage.define(simpleMode(logrotateSyntax)),
      EditorState.readOnly.of(false),
    ];
  };

  const getNftablesExtensions = () => {
    return [
      StreamLanguage.define(simpleMode(nftablesSyntax)),
      EditorState.readOnly.of(false),
    ];
  };

  const getServiceExtensions = () => {
    return [
      StreamLanguage.define(simpleMode(serviceSyntax)),
      EditorState.readOnly.of(false),
    ];
  };

  const extensions = computed(() => {
    if (!file.value) {
      return [];
    }

    const fileName = file.value.name;
    const filePath = file.value.path;

    // 特殊文件名处理
    if (isBashConfigFile(fileName)) {
      return getExtensionsForLanguage('shell');
    }

    // 基于文件扩展名判断
    const extension = getFileExtension(fileName);
    if (extension) {
      const languageType = EXTENSION_MAP[extension];
      return getExtensionsForLanguage(languageType);
    }

    // 特殊文件类型处理（基于文件名模式）
    if (/logrotate/.test(filePath) || fileName === 'logrotate') {
      return getLogrotateExtensions();
    }

    if (/nftables|\.nft$/.test(filePath) || fileName.includes('nftables')) {
      return getNftablesExtensions();
    }

    if (/\.service$/.test(fileName) || /systemd/.test(filePath)) {
      return getServiceExtensions();
    }

    return [];
  });

  return {
    extensions,
    getExtensionsForLanguage,
    getLogrotateExtensions,
    getNftablesExtensions,
    getServiceExtensions,
  };
}
