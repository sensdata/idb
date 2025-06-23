import { ref } from 'vue';
import { LOGROTATE_FREQUENCY } from '@/config/enum';
import { useLogger } from '@/composables/use-logger';
import type { FormData } from '../types';

export function useRawContentParser() {
  const rawContent = ref('');
  const { log } = useLogger('RawContentParser');

  const generateRawContent = (formData: FormData) => {
    let content = `# Logrotate configuration for ${formData.name}\n`;
    content += `${formData.path} {\n`;

    // 将枚举值转换为小写字符串以匹配配置文件格式
    const frequencyStr = formData.frequency.toLowerCase();
    content += `  ${frequencyStr}\n`;

    content += `  rotate ${formData.count}\n`;

    if (formData.compress) {
      content += '  compress\n';
    }

    if (formData.delayCompress) {
      content += '  delaycompress\n';
    }

    if (formData.missingOk) {
      content += '  missingok\n';
    }

    if (formData.notIfEmpty) {
      content += '  notifempty\n';
    }

    if (formData.create) {
      content += `  create ${formData.create}\n`;
    }

    if (formData.preRotate) {
      content += '  prerotate\n';
      content += `    ${formData.preRotate}\n`;
      content += '  endscript\n';
    }

    if (formData.postRotate) {
      content += '  postrotate\n';
      content += `    ${formData.postRotate}\n`;
      content += '  endscript\n';
    }

    content += '}\n';
    rawContent.value = content;
  };

  // 解析文件路径
  const parsePath = (content: string): string => {
    const pathMatch = content.match(/^([^#\n]+?)\s*\{/m);
    return pathMatch ? pathMatch[1].trim() : '';
  };

  // 解析频率选项
  const parseFrequency = (configContent: string): LOGROTATE_FREQUENCY => {
    const frequencyMatch = configContent.match(
      /^\s*(daily|weekly|monthly|yearly)\s*$/m
    );

    if (frequencyMatch) {
      const value = frequencyMatch[1].toLowerCase();
      switch (value) {
        case 'daily':
          return LOGROTATE_FREQUENCY.Daily;
        case 'weekly':
          return LOGROTATE_FREQUENCY.Weekly;
        case 'monthly':
          return LOGROTATE_FREQUENCY.Monthly;
        case 'yearly':
          return LOGROTATE_FREQUENCY.Yearly;
        default:
          return LOGROTATE_FREQUENCY.Daily;
      }
    }

    return LOGROTATE_FREQUENCY.Daily;
  };

  // 解析轮转次数
  const parseRotateCount = (configContent: string): number => {
    const rotateMatch = configContent.match(/^\s*rotate\s+(\d+)\s*$/m);
    return rotateMatch ? parseInt(rotateMatch[1], 10) : 7;
  };

  // 解析布尔选项
  const parseBooleanOptions = (configContent: string) => {
    return {
      compress: /^\s*compress\s*$/m.test(configContent),
      delayCompress: /^\s*delaycompress\s*$/m.test(configContent),
      missingOk: /^\s*missingok\s*$/m.test(configContent),
      notIfEmpty: /^\s*notifempty\s*$/m.test(configContent),
    };
  };

  // 解析create选项
  const parseCreateOption = (configContent: string): string => {
    const createMatch = configContent.match(/^\s*create\s+(.+?)\s*$/m);
    return createMatch ? createMatch[1].trim() : '';
  };

  // 解析脚本内容
  const parseScriptContent = (
    configContent: string,
    scriptType: 'prerotate' | 'postrotate'
  ): string => {
    const regex = new RegExp(
      `^\\s*${scriptType}\\s*\\n([\\s\\S]*?)^\\s*endscript\\s*$`,
      'm'
    );
    const match = configContent.match(regex);
    if (match) {
      // 清理缩进，保留脚本内容
      return match[1].replace(/^\s{4}/gm, '').trim();
    }
    return '';
  };

  const parseRawContentToForm = (baseFormData: FormData): FormData | null => {
    const content = rawContent.value;
    if (!content) return null;

    log('🔍 开始解析原始内容:', content);

    // 基于现有的表单数据创建新的数据对象，确保所有字段都存在
    const parsedData: FormData = {
      name: baseFormData.name, // 保持现有的 name
      category: baseFormData.category, // 保持现有的 category
      path: '',
      frequency: LOGROTATE_FREQUENCY.Daily,
      count: 7,
      compress: false,
      delayCompress: false,
      missingOk: false,
      notIfEmpty: false,
      create: '',
      preRotate: '',
      postRotate: '',
    };

    // 解析路径
    parsedData.path = parsePath(content);

    // 提取 {} 内的配置内容
    const configMatch = content.match(/\{([\s\S]*)\}/);
    const configContent = configMatch ? configMatch[1] : content;

    // 解析频率
    parsedData.frequency = parseFrequency(configContent);

    // 解析轮转次数
    parsedData.count = parseRotateCount(configContent);

    // 解析布尔选项
    const boolOptions = parseBooleanOptions(configContent);
    log('🗜️ compress 选项存在:', boolOptions.compress);
    parsedData.compress = boolOptions.compress;
    parsedData.delayCompress = boolOptions.delayCompress;
    parsedData.missingOk = boolOptions.missingOk;
    parsedData.notIfEmpty = boolOptions.notIfEmpty;

    // 解析create选项
    parsedData.create = parseCreateOption(configContent);

    // 解析prerotate脚本
    parsedData.preRotate = parseScriptContent(configContent, 'prerotate');

    // 解析postrotate脚本
    parsedData.postRotate = parseScriptContent(configContent, 'postrotate');

    log('✅ 解析结果:', parsedData);
    return parsedData;
  };

  return {
    rawContent,
    generateRawContent,
    parseRawContentToForm,
  };
}
