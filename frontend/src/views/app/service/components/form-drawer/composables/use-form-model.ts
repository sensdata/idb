import { ref } from 'vue';
import { useLogger } from '@/composables/use-logger';
import type { SERVICE_TYPE } from '@/config/enum';
import type { ParsedServiceConfig } from '../types';

export const useFormModel = (type: SERVICE_TYPE, category: string) => {
  const { logError } = useLogger('ServiceFormModel');

  // 初始化表单数据模型
  const formModel = ref({
    name: '',
    category: category || '',
    description: '',
    serviceType: 'simple',
    execStart: '',
    workingDirectory: '',
    user: 'root',
    group: 'root',
    environment: '',
    execStop: '',
    execReload: '',
    restart: 'no',
    restartSec: 0,
    timeoutStartSec: 90,
    timeoutStopSec: 90,
    rawContent: '',
  });

  /**
   * 从解析后的配置数据设置表单值
   * @param data 包含服务配置信息的对象
   */
  const setFormData = (data: {
    name?: string;
    category?: string;
    parsedConfig?: ParsedServiceConfig;
    fields?: Array<{ key: string; value: string }>;
  }) => {
    if (data.parsedConfig) {
      // 直接使用解析后的配置对象
      const config = data.parsedConfig;

      formModel.value = {
        name: data.name || '',
        category: data.category || category || '', // 优先使用传入的分类
        description: config.description || '',
        serviceType: config.serviceType || 'simple',
        execStart: config.execStart || '',
        workingDirectory: config.workingDirectory || '',
        user: config.user || 'root',
        group: config.group || 'root',
        environment: config.environment || '',
        execStop: config.execStop || '',
        execReload: config.execReload || '',
        restart: config.restart || 'no',
        restartSec: Number(config.restartSec) || 0,
        timeoutStartSec: Number(config.timeoutStartSec) || 90,
        timeoutStopSec: Number(config.timeoutStopSec) || 90,
        rawContent: '',
      };
    } else if (data.fields) {
      // 兼容旧的API格式（如果还需要的话）
      const fieldMap = data.fields.reduce(
        (
          map: Record<string, string>,
          field: { key: string; value: string }
        ) => {
          map[field.key] = field.value;
          return map;
        },
        {}
      );

      formModel.value = {
        name: data.name || '',
        category: data.category || category || '', // 优先使用传入的分类
        description: fieldMap.Description || '',
        serviceType: fieldMap.Type || 'simple',
        execStart: fieldMap.ExecStart || '',
        workingDirectory: fieldMap.WorkingDirectory || '',
        user: fieldMap.User || 'root',
        group: fieldMap.Group || 'root',
        environment: fieldMap.Environment || '',
        execStop: fieldMap.ExecStop || '',
        execReload: fieldMap.ExecReload || '',
        restart: fieldMap.Restart || 'no',
        restartSec: parseInt(fieldMap.RestartSec || '0', 10),
        timeoutStartSec: parseInt(fieldMap.TimeoutStartSec || '90', 10),
        timeoutStopSec: parseInt(fieldMap.TimeoutStopSec || '90', 10),
        rawContent: '',
      };
    }
  };

  /**
   * 获取表单数据，返回合并后的原始配置内容
   * @param rawContent 原始配置内容
   * @returns 表单数据对象
   */
  const getFormData = async (rawContent: string) => {
    try {
      // 导入解析器
      const { useServiceParser } = await import('./use-service-parser');
      const { mergeFormToConfig } = useServiceParser();

      // 获取原始配置内容
      const originalContent = rawContent || '';

      // 将表单数据合并到原始配置中
      const mergedContent = mergeFormToConfig(originalContent, {
        description: formModel.value.description,
        serviceType: formModel.value.serviceType,
        execStart: formModel.value.execStart,
        workingDirectory: formModel.value.workingDirectory,
        user: formModel.value.user,
        group: formModel.value.group,
        environment: formModel.value.environment,
        execStop: formModel.value.execStop,
        execReload: formModel.value.execReload,
        restart: formModel.value.restart,
        restartSec: formModel.value.restartSec,
        timeoutStartSec: formModel.value.timeoutStartSec,
        timeoutStopSec: formModel.value.timeoutStopSec,
      });

      // 返回原始内容格式
      return {
        type,
        category: formModel.value.category,
        name: formModel.value.name,
        content: mergedContent, // 返回合并后的完整配置内容
      };
    } catch (error) {
      logError('无法合并表单数据到配置文件:', error);
      throw new Error('无法合并表单数据到配置文件');
    }
  };

  return {
    formModel,
    setFormData,
    getFormData,
  };
};
