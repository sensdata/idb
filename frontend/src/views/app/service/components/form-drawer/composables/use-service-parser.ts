/**
 * systemd服务配置解析工具
 * 从原始配置文件中解析出表单字段，实现单一数据源
 */
import { ref, computed } from 'vue';
import { camelCase } from 'lodash';
import { useLogger } from '@/composables/use-logger';

import type { ParsedServiceConfig, ServiceConfigStructured } from '../types';

// 导入常量
import {
  SECTION_UNIT,
  SECTION_SERVICE,
  SECTION_INSTALL,
  SERVICE_TYPES,
  RESTART_OPTIONS,
  DEFAULT_TIMEOUT,
  DEFAULT_SERVICE_TEMPLATE,
} from '../constants';

// 导入工具类
import {
  isArrayField,
  insertNewKeys,
  serializeSection,
  extractServiceContent,
  parseTimeValue,
} from '../utils/service-config-utils';
import {
  getFormValueForKey,
  getNewKeysToAdd,
} from '../utils/key-mapping-utils';

export function useServiceParser() {
  const rawContent = ref('');
  const { log } = useLogger('ServiceParser');

  /**
   * 创建服务配置对象，填充默认值
   */
  function createServiceConfig(
    partialConfig: Partial<ParsedServiceConfig>,
    content: string
  ): ParsedServiceConfig {
    return {
      description: partialConfig.description || '',
      serviceType: partialConfig.serviceType || SERVICE_TYPES[0],
      execStart: partialConfig.execStart || '',
      execStop: partialConfig.execStop || '',
      execReload: partialConfig.execReload || '',
      workingDirectory: partialConfig.workingDirectory || '',
      user: partialConfig.user || 'root',
      group: partialConfig.group || 'root',
      environment: partialConfig.environment || '',
      restart: partialConfig.restart || RESTART_OPTIONS[0],
      restartSec: partialConfig.restartSec || 0,
      timeoutStartSec: partialConfig.timeoutStartSec || DEFAULT_TIMEOUT,
      timeoutStopSec: partialConfig.timeoutStopSec || DEFAULT_TIMEOUT,
      rawContent: content,
    };
  }

  /**
   * 创建默认配置
   */
  function createDefaultConfig(): ParsedServiceConfig {
    return createServiceConfig({}, '');
  }

  /**
   * 默认模板，当没有原始内容时使用
   */
  function defaultTemplate(): string {
    return DEFAULT_SERVICE_TEMPLATE;
  }

  /**
   * 将键值对映射到表单数据
   */
  function mapKeyValueToFormData(
    key: string,
    value: string,
    formData: Partial<ParsedServiceConfig>
  ): void {
    log(`解析键值对: ${key}=${value}`);

    switch (key) {
      case 'Description':
        formData.description = value;
        break;
      case 'Type':
        formData.serviceType = value;
        break;
      case 'ExecStart':
        formData.execStart = value;
        break;
      case 'ExecStop':
        formData.execStop = value;
        break;
      case 'ExecReload':
        formData.execReload = value;
        break;
      case 'WorkingDirectory':
        formData.workingDirectory = value;
        break;
      case 'User':
        formData.user = value;
        break;
      case 'Group':
        formData.group = value;
        break;
      case 'Environment':
        log(`解析环境变量: ${value}`);
        // 直接使用原始字符串，不做额外处理
        formData.environment = value;
        break;
      case 'Restart':
        formData.restart = value;
        break;
      case 'RestartSec':
        formData.restartSec = parseTimeValue(value);
        break;
      case 'TimeoutStartSec':
        formData.timeoutStartSec = parseTimeValue(value);
        break;
      case 'TimeoutStopSec':
        formData.timeoutStopSec = parseTimeValue(value);
        break;
      default:
        log(`未处理的配置项: ${key}=${value}`);
        break; // 忽略未知的配置项
    }

    log(`解析后的表单数据:`, formData);
  }

  /**
   * 解析systemd service文件内容
   */
  function parseServiceConfig(content: string): ParsedServiceConfig {
    try {
      log('开始解析服务配置文件内容');
      const lines = content.split('\n');
      const serviceFormData: Partial<ParsedServiceConfig> = {};

      // 解析每一行
      for (const line of lines) {
        const trimmedLine = line.trim();

        // 跳过注释、空行和段标题
        if (
          !trimmedLine ||
          trimmedLine.startsWith('#') ||
          trimmedLine.startsWith(';') ||
          (trimmedLine.startsWith('[') && trimmedLine.endsWith(']'))
        ) {
          continue;
        }

        // 特殊处理Environment，因为它可能有多个引号
        if (trimmedLine.startsWith('Environment=')) {
          log('找到Environment行:', trimmedLine);
          const envValue = trimmedLine.substring('Environment='.length);
          serviceFormData.environment = envValue;
          continue;
        }

        // 解析其他键值对
        const equalIndex = trimmedLine.indexOf('=');
        if (equalIndex === -1) continue;

        const key = trimmedLine.substring(0, equalIndex).trim();
        const value = trimmedLine.substring(equalIndex + 1).trim();

        // 映射到表单字段
        mapKeyValueToFormData(key, value, serviceFormData);
      }

      // 返回包含默认值的完整配置
      return createServiceConfig(serviceFormData, content);
    } catch (error) {
      log('解析服务配置失败:', error);
      return createDefaultConfig();
    }
  }

  // 计算属性：解析后的配置
  const parsedConfig = computed(() => {
    if (!rawContent.value) return createDefaultConfig();
    return parseServiceConfig(rawContent.value);
  });

  /**
   * 将表单数据合并到原始配置内容中
   */
  function mergeFormToConfig(
    originalContent: string,
    formData: Partial<ParsedServiceConfig>
  ): string {
    try {
      log('合并表单数据到原始配置');
      const lines = originalContent.split('\n');
      const updatedLines: string[] = [];
      const processedKeys = new Set<string>();

      for (const line of lines) {
        const trimmedLine = line.trim();

        // 保留注释、空行和段标题
        if (
          !trimmedLine ||
          trimmedLine.startsWith('#') ||
          trimmedLine.startsWith(';') ||
          (trimmedLine.startsWith('[') && trimmedLine.endsWith(']'))
        ) {
          updatedLines.push(line);
          continue;
        }

        // 处理键值对
        const equalIndex = trimmedLine.indexOf('=');
        if (equalIndex !== -1) {
          const key = trimmedLine.substring(0, equalIndex).trim();
          const newValue = getFormValueForKey(key, formData);

          if (newValue !== undefined) {
            // 使用新值替换
            updatedLines.push(`${key}=${newValue}`);
            processedKeys.add(key);
          } else {
            // 保留原值
            updatedLines.push(line);
          }
        } else {
          updatedLines.push(line);
        }
      }

      // 添加新的键值对
      const newKeys = getNewKeysToAdd(formData, processedKeys);
      insertNewKeys(updatedLines, newKeys);

      return updatedLines.join('\n');
    } catch (error) {
      log('合并表单数据失败:', error);
      return originalContent;
    }
  }

  /**
   * 从行内容中提取段落名称
   */
  function extractSectionName(line: string): string | null {
    const trimmedLine = line.trim();
    if (!trimmedLine) return null;

    // 检查是否是节标题 [Section]
    const sectionMatch = trimmedLine.match(/^\[(.+)\]$/);
    if (sectionMatch) {
      return sectionMatch[1].toLowerCase();
    }

    return null;
  }

  /**
   * 确保配置对象中存在指定的段落
   */
  function ensureSectionExists(
    config: ServiceConfigStructured,
    sectionName: string
  ): void {
    if (!config[sectionName]) {
      config[sectionName] = {};
    }
  }

  /**
   * 从行内容中提取键值对
   */
  function extractKeyValue(line: string): {
    key: string | null;
    value: string | null;
  } {
    const trimmedLine = line.trim();

    // 跳过空行和注释
    if (
      !trimmedLine ||
      trimmedLine.startsWith('#') ||
      trimmedLine.startsWith(';')
    ) {
      return { key: null, value: null };
    }

    // 解析键值对
    const keyValueMatch = trimmedLine.match(/^([^=]+)=(.*)$/);
    if (keyValueMatch) {
      const key = keyValueMatch[1].trim();
      const value = keyValueMatch[2].trim();
      return { key, value };
    }

    return { key: null, value: null };
  }

  /**
   * 向段落中添加键值对
   */
  function addKeyValueToSection(
    config: ServiceConfigStructured,
    sectionName: string,
    key: string,
    value: string
  ): void {
    // 确保段落存在
    ensureSectionExists(config, sectionName);

    const section = config[sectionName];
    if (!section) return;

    // 转换为驼峰命名
    const camelKey = camelCase(key);

    // 处理可能有多个值的字段
    if (isArrayField(key)) {
      if (!section[camelKey]) {
        section[camelKey] = [];
      }

      // 拆分空格分隔的多个项
      const values = value.split(/\s+/).filter((v) => v.length > 0);

      if (Array.isArray(section[camelKey])) {
        (section[camelKey] as string[]).push(...values);
      } else {
        section[camelKey] = values;
      }
    } else {
      section[camelKey] = value;
    }
  }

  /**
   * 解析 systemd 服务配置文件内容为结构化对象
   */
  function parseServiceConfigStructured(
    content: string
  ): ServiceConfigStructured {
    try {
      log('解析服务配置为结构化对象');
      const structuredConfig: ServiceConfigStructured = {};

      if (!content || typeof content !== 'string') {
        return structuredConfig;
      }

      const lines = content.split('\n');
      let currentSection = '';

      for (const line of lines) {
        const sectionName = extractSectionName(line);
        if (sectionName) {
          currentSection = sectionName;
          ensureSectionExists(structuredConfig, currentSection);
          continue;
        }

        if (!currentSection) continue;

        const { key, value } = extractKeyValue(line);
        if (!key || !value) continue;

        addKeyValueToSection(structuredConfig, currentSection, key, value);
      }

      return structuredConfig;
    } catch (error) {
      log('解析结构化配置失败:', error);
      return {};
    }
  }

  /**
   * 将配置对象转换回 systemd 服务配置文件格式
   */
  function serializeServiceConfig(config: ServiceConfigStructured): string {
    try {
      log('将结构化对象序列化为配置内容');
      const lines: string[] = [];

      // 按照标准顺序处理节
      const sectionOrder = [SECTION_UNIT, SECTION_SERVICE, SECTION_INSTALL];

      // 处理标准节
      for (const sectionName of sectionOrder) {
        serializeSection(config, sectionName, lines);
      }

      // 处理其他自定义节
      for (const sectionName of Object.keys(config)) {
        if (!sectionOrder.includes(sectionName.toLowerCase())) {
          serializeSection(config, sectionName, lines);
        }
      }

      // 移除末尾的空行
      while (lines.length > 0 && lines[lines.length - 1] === '') {
        lines.pop();
      }

      return lines.join('\n');
    } catch (error) {
      log('序列化配置失败:', error);
      return '';
    }
  }

  /**
   * 解析原始内容为表单数据
   */
  function parseRawContentToForm(content: string): ParsedServiceConfig {
    log('解析原始内容为表单数据');
    rawContent.value = content;
    return parseServiceConfig(content);
  }

  /**
   * 从表单数据生成原始内容
   */
  function generateRawContent(
    formData: Partial<ParsedServiceConfig>,
    originalContent = ''
  ): string {
    log('从表单数据生成原始内容');
    const content = mergeFormToConfig(
      originalContent || defaultTemplate(),
      formData
    );
    rawContent.value = content;
    return content;
  }

  return {
    rawContent,
    parsedConfig,
    parseServiceConfig,
    mergeFormToConfig,
    parseServiceConfigStructured,
    serializeServiceConfig,
    extractServiceContent,
    parseRawContentToForm,
    generateRawContent,
    defaultTemplate,
  };
}
