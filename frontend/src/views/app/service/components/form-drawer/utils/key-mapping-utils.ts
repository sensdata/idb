/**
 * 表单数据与配置键的映射工具
 */
import { ParsedServiceConfig } from '../types';

/**
 * 创建表单数据到配置键的映射
 */
function createKeyMappings(
  formData: Partial<ParsedServiceConfig>
): Record<string, string | undefined> {
  return {
    Description: formData.description,
    Type: formData.serviceType,
    ExecStart: formData.execStart,
    ExecStop: formData.execStop,
    ExecReload: formData.execReload,
    WorkingDirectory: formData.workingDirectory,
    User: formData.user,
    Group: formData.group,
    Environment: formData.environment,
    Restart: formData.restart,
    RestartSec: formData.restartSec ? `${formData.restartSec}s` : undefined,
    TimeoutStartSec: formData.timeoutStartSec
      ? `${formData.timeoutStartSec}s`
      : undefined,
    TimeoutStopSec: formData.timeoutStopSec
      ? `${formData.timeoutStopSec}s`
      : undefined,
  };
}

/**
 * 获取表单字段对应的配置键值
 */
export function getFormValueForKey(
  key: string,
  formData: Partial<ParsedServiceConfig>
): string | undefined {
  const mapping = createKeyMappings(formData);
  return mapping[key];
}

/**
 * 获取需要添加的新键值对
 */
export function getNewKeysToAdd(
  formData: Partial<ParsedServiceConfig>,
  existingKeys: Set<string>
): [string, string][] {
  const newKeys: [string, string][] = [];
  const keyMappings = createKeyMappings(formData);

  for (const [key, value] of Object.entries(keyMappings)) {
    if (value !== undefined && !existingKeys.has(key)) {
      newKeys.push([key, value]);
    }
  }

  return newKeys;
}
