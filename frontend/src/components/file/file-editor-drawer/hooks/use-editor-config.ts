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

// 定义日志文件的高亮规则
const logSyntax = {
  // 起始状态
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

// 创建日志语法高亮器
const logMode = simpleMode({ start: logSyntax.start });

// 添加统一的行分隔符配置，确保一致的换行符处理
const commonExtensions = [EditorState.lineSeparator.of('\n')];

export default function useEditorConfig(file: Ref<FileItem | null>) {
  const extensions = computed(() => {
    if (!file.value) {
      return commonExtensions;
    }

    const fileName = file.value.name.toLowerCase();

    // 特殊处理没有扩展名的文件，如 .bashrc, .bash_profile 等
    if (
      fileName === '.bashrc' ||
      fileName === '.bash_profile' ||
      fileName === '.bash_login' ||
      fileName === '.bash_logout' ||
      fileName === '.profile'
    ) {
      return [StreamLanguage.define(shell), ...commonExtensions];
    }

    const ext = fileName.split('.').pop();

    if (!ext) {
      return commonExtensions;
    }

    switch (ext) {
      case 'js':
      case 'jsx':
      case 'ts':
      case 'tsx':
        return [javascript(), ...commonExtensions];
      case 'json':
        return [json(), ...commonExtensions];
      case 'html':
      case 'htm':
      case 'vue':
        return [html(), ...commonExtensions];
      case 'css':
      case 'scss':
      case 'less':
        return [css(), ...commonExtensions];
      case 'md':
      case 'markdown':
        return [markdown(), ...commonExtensions];
      case 'sh':
      case 'bash':
        return [StreamLanguage.define(shell), ...commonExtensions];
      case 'yaml':
      case 'yml':
        return [StreamLanguage.define(yaml), ...commonExtensions];
      case 'properties':
      case 'env':
        return [StreamLanguage.define(properties), ...commonExtensions];
      case 'conf':
      case 'nginx':
        return [StreamLanguage.define(nginx), ...commonExtensions];
      case 'toml':
      case 'ini':
        return [StreamLanguage.define(toml), ...commonExtensions];
      case 'log':
        // 使用自定义的日志高亮模式，并确保一致的换行符处理
        return [StreamLanguage.define(logMode), ...commonExtensions];
      default:
        return commonExtensions;
    }
  });

  return {
    extensions,
  };
}
