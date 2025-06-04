/**
 * systemd服务配置处理工具类
 */
import { includes, isEmpty, findLastIndex, snakeCase } from 'lodash';
import { ARRAY_FIELDS, UNIT_FIELDS, INSTALL_FIELDS } from '../constants';
import { ServiceConfigStructured } from '../types';

/**
 * 判断字段是否应该作为数组处理
 */
export function isArrayField(key: string): boolean {
  return includes(ARRAY_FIELDS, key.toLowerCase());
}

/**
 * 判断是否是环境变量字段（需要每行一个值）
 */
export function isEnvironmentField(key: string): boolean {
  return key.toLowerCase() === 'environment';
}

/**
 * 判断字段应该属于哪个段落
 */
export function getFieldSection(key: string): string {
  const lowerKey = key.toLowerCase();

  // Unit段落的字段
  if (UNIT_FIELDS.includes(lowerKey)) {
    return 'Unit';
  }

  // Install段落的字段
  if (INSTALL_FIELDS.includes(lowerKey)) {
    return 'Install';
  }

  // 默认为Service段落
  return 'Service';
}

/**
 * 在适当的段落插入新键
 */
export function insertNewKeys(
  lines: string[],
  newKeys: [string, string][]
): void {
  if (isEmpty(newKeys)) return;

  // 按段落分组新键
  const keysBySection: Record<string, [string, string][]> = {
    Unit: [],
    Service: [],
    Install: [],
  };

  for (const [key, value] of newKeys) {
    const section = getFieldSection(key);
    keysBySection[section].push([key, value]);
  }

  // 为每个段落插入相应的键
  for (const [sectionName, sectionKeys] of Object.entries(keysBySection)) {
    if (sectionKeys.length === 0) continue;

    const sectionHeader = `[${sectionName}]`;
    const sectionIndex = findLastIndex(
      lines,
      (line) => line.trim() === sectionHeader
    );

    if (sectionIndex >= 0) {
      // 在现有段落之后添加新键
      let insertIndex = sectionIndex + 1;
      // 找到段落内容的最后一行（下一个段落之前或文件末尾）
      while (
        insertIndex < lines.length &&
        !lines[insertIndex].trim().startsWith('[') &&
        lines[insertIndex].trim() !== ''
      ) {
        insertIndex++;
      }

      // 在段落内容的最后插入新键
      for (const [key, value] of sectionKeys) {
        lines.splice(insertIndex, 0, `${key}=${value}`);
        insertIndex++;
      }
    } else {
      // 如果没找到对应段落，创建一个新的段落
      // 找到合适的位置插入段落（按Unit -> Service -> Install的顺序）
      let insertPosition = lines.length;

      if (sectionName === 'Unit') {
        // Unit段落应该在最前面
        insertPosition = 0;
      } else if (sectionName === 'Service') {
        // Service段落在Unit之后，Install之前
        const installIndex = findLastIndex(
          lines,
          (line) => line.trim() === '[Install]'
        );
        if (installIndex >= 0) {
          insertPosition = installIndex;
        }
      }
      // Install段落默认在最后

      // 插入段落标题和键值对
      lines.splice(insertPosition, 0, sectionHeader);
      insertPosition++;
      for (const [key, value] of sectionKeys) {
        lines.splice(insertPosition, 0, `${key}=${value}`);
        insertPosition++;
      }
      // 添加空行分隔
      lines.splice(insertPosition, 0, '');
    }
  }
}

/**
 * 序列化单个配置段
 */
export function serializeSection(
  config: ServiceConfigStructured,
  sectionName: string,
  lines: string[]
): void {
  const section = config[sectionName];
  if (!section || typeof section !== 'object') return;

  // 添加段标题
  const capitalizedName =
    sectionName.charAt(0).toUpperCase() + sectionName.slice(1);
  lines.push(`[${capitalizedName}]`);

  // 添加段内容
  for (const [key, value] of Object.entries(section)) {
    if (value === undefined || value === null || value === '') continue;

    const snakeKey = snakeCase(key);

    if (Array.isArray(value)) {
      if (isEnvironmentField(snakeKey)) {
        // Environment 字段每个值一行
        for (const item of value) {
          lines.push(`${snakeKey}=${item}`);
        }
      } else {
        // 其他数组字段用空格分隔
        lines.push(`${snakeKey}=${value.join(' ')}`);
      }
    } else {
      lines.push(`${snakeKey}=${value}`);
    }
  }

  // 段之间添加空行
  lines.push('');
}

/**
 * 从 API 响应中提取内容
 */
export function extractServiceContent(response: any): string {
  try {
    if (typeof response === 'string') {
      return response;
    }

    if (response && typeof response === 'object') {
      // 检查各种可能的数据结构
      if (response.content && typeof response.content === 'string') {
        return response.content;
      }

      if (
        response.items &&
        Array.isArray(response.items) &&
        response.items.length > 0 &&
        response.items[0].content
      ) {
        return response.items[0].content;
      }

      if (response.data) {
        return extractServiceContent(response.data);
      }
    }

    return '';
  } catch (error) {
    console.error('提取服务内容失败:', error);
    return '';
  }
}

/**
 * 解析时间值（支持s, m, h等单位）
 */
export function parseTimeValue(value: string): number {
  try {
    const match = value.match(/^(\d+)([smh]?)$/);
    if (!match) return parseInt(value, 10) || 0;

    const num = parseInt(match[1], 10);
    const unit = match[2] || 's';

    switch (unit) {
      case 'm':
        return num * 60;
      case 'h':
        return num * 3600;
      case 's':
      default:
        return num;
    }
  } catch (error) {
    console.error('解析时间值失败:', error);
    return 0;
  }
}
