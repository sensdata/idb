/**
 * 开发环境日志工具
 * 封装控制台方法，仅在开发环境下输出日志
 */

/* eslint-disable no-console */
import { ref, onMounted, onUnmounted } from 'vue';
import { createLogger } from '@/utils/logger';

export const useLogger = (component?: string) => {
  // 创建底层日志实例
  const logger = createLogger(component);

  // 控制日志是否启用（保持响应式，用于组件内部状态管理）
  const isEnabled = ref(import.meta.env.DEV);

  // 记录组件挂载和卸载
  onMounted(() => {
    if (isEnabled.value && component) {
      logger.log('组件已挂载');
    }
  });

  onUnmounted(() => {
    if (isEnabled.value && component) {
      logger.log('组件已卸载');
    }
  });

  // 设置日志级别
  const setLogEnabled = (enabled: boolean): void => {
    isEnabled.value = enabled;
  };

  return {
    log: logger.log.bind(logger),
    logDebug: logger.logDebug.bind(logger),
    logInfo: logger.logInfo.bind(logger),
    logWarn: logger.logWarn.bind(logger),
    logError: logger.logError.bind(logger),
    setLogEnabled,
    isEnabled,
  };
};

export default useLogger;
