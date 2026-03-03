import { ref } from 'vue';
import { LOGROTATE_FREQUENCY } from '@/config/enum';
import { useLogger } from '@/composables/use-logger';
import type { FormData } from '../types';
import {
  extractCreateDirective,
  extractScriptContent,
  generateLogrotateContentFromForm,
} from '../../../utils/content';

export function useRawContentParser() {
  const rawContent = ref('');
  const { log } = useLogger('RawContentParser');

  const generateRawContent = (formData: FormData) => {
    rawContent.value = generateLogrotateContentFromForm(formData, {
      includeHeader: true,
      indent: '  ',
    });
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
    parsedData.create = extractCreateDirective(configContent);

    // 解析prerotate脚本
    parsedData.preRotate = extractScriptContent(configContent, 'prerotate');

    // 解析postrotate脚本
    parsedData.postRotate = extractScriptContent(configContent, 'postrotate');

    log('✅ 解析结果:', parsedData);
    return parsedData;
  };

  return {
    rawContent,
    generateRawContent,
    parseRawContentToForm,
  };
}
