import type { FormData } from '../components/form-drawer/types';

type ScriptType = 'prerotate' | 'postrotate';

const SCRIPT_END = 'endscript';

const trimEmptyLines = (lines: string[]): string[] => {
  let start = 0;
  let end = lines.length;

  for (; start < lines.length; start += 1) {
    if (lines[start].trim() !== '') {
      break;
    }
  }

  for (; end > start; end -= 1) {
    if (lines[end - 1].trim() !== '') {
      break;
    }
  }

  return lines.slice(start, end);
};

const dedentMultiline = (value: string): string => {
  const normalized = value.replace(/\r\n/g, '\n');
  const lines = trimEmptyLines(normalized.split('\n'));
  if (lines.length === 0) {
    return '';
  }

  const indentLengths = lines
    .filter((line) => line.trim() !== '')
    .map((line) => {
      const match = line.match(/^[ \t]*/);
      return match ? match[0].length : 0;
    });

  const minIndent = indentLengths.length > 0 ? Math.min(...indentLengths) : 0;
  return lines
    .map((line) => line.slice(minIndent).replace(/\s+$/, ''))
    .join('\n')
    .trim();
};

export const normalizeCreateDirective = (value: string): string => {
  const trimmed = value.trim();
  if (!trimmed) {
    return '';
  }
  return /^create\b/i.test(trimmed) ? trimmed : `create ${trimmed}`;
};

export const extractCreateDirective = (configContent: string): string => {
  const createMatch = configContent.match(/^\s*create\s+(.+?)\s*$/m);
  return createMatch ? `create ${createMatch[1].trim()}` : '';
};

export const extractScriptContent = (
  configContent: string,
  scriptType: ScriptType
): string => {
  const regex = new RegExp(
    `^\\s*${scriptType}\\s*\\n([\\s\\S]*?)^\\s*${SCRIPT_END}\\s*$`,
    'im'
  );
  const match = configContent.match(regex);
  if (!match || !match[1]) {
    return '';
  }
  return dedentMultiline(match[1]);
};

const appendScriptBlock = (
  lines: string[],
  scriptType: ScriptType,
  scriptContent: string,
  indent: string
) => {
  const normalized = dedentMultiline(scriptContent);
  if (!normalized) {
    return;
  }

  const scriptIndent = `${indent}${indent}`;
  lines.push(`${indent}${scriptType}`);
  normalized.split('\n').forEach((line) => {
    lines.push(`${scriptIndent}${line}`);
  });
  lines.push(`${indent}${SCRIPT_END}`);
};

export const generateLogrotateContentFromForm = (
  formData: FormData,
  options?: { includeHeader?: boolean; indent?: string }
): string => {
  const indent = options?.indent ?? '    ';
  const lines: string[] = [];

  if (options?.includeHeader) {
    lines.push(`# Logrotate configuration for ${formData.name}`);
  }

  lines.push(`${formData.path} {`);
  lines.push(`${indent}${String(formData.frequency).toLowerCase()}`);
  lines.push(`${indent}rotate ${formData.count}`);

  if (formData.compress) {
    lines.push(`${indent}compress`);
  }
  if (formData.delayCompress) {
    lines.push(`${indent}delaycompress`);
  }
  if (formData.missingOk) {
    lines.push(`${indent}missingok`);
  }
  if (formData.notIfEmpty) {
    lines.push(`${indent}notifempty`);
  }

  const createDirective = normalizeCreateDirective(formData.create);
  if (createDirective) {
    lines.push(`${indent}${createDirective}`);
  }

  appendScriptBlock(lines, 'prerotate', formData.preRotate, indent);
  appendScriptBlock(lines, 'postrotate', formData.postRotate, indent);

  lines.push('}');
  return lines.join('\n');
};
