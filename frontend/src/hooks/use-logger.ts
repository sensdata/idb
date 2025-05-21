/**
 * 开发环境日志工具
 * 封装控制台方法，仅在开发环境下输出日志
 */

import { ref, onMounted, onUnmounted } from 'vue';

// 日志参数类型
type LogParams = (
  | string
  | number
  | boolean
  | object
  | null
  | undefined
  | unknown
)[];

export const useLogger = (component?: string) => {
  const isDev = import.meta.env.DEV;
  // 控制日志是否启用
  const isEnabled = ref(isDev);
  // 日志前缀
  const prefix = component ? `[${component}]` : '';

  // 记录组件挂载和卸载
  onMounted(() => {
    if (isEnabled.value && component) {
      console.log(`${prefix} 组件已挂载`);
    }
  });

  onUnmounted(() => {
    if (isEnabled.value && component) {
      console.log(`${prefix} 组件已卸载`);
    }
  });

  // 标准日志（与logDebug相同，但命名更简单）
  const log = (...args: LogParams): void => {
    if (isEnabled.value) {
      console.log(prefix, ...args);
    }
  };

  // 调试日志
  const logDebug = (...args: LogParams): void => {
    if (isEnabled.value) {
      console.log(prefix, ...args);
    }
  };

  // 信息日志
  const logInfo = (...args: LogParams): void => {
    if (isEnabled.value) {
      console.info(prefix, ...args);
    }
  };

  // 警告日志
  const logWarn = (...args: LogParams): void => {
    if (isEnabled.value) {
      console.warn(prefix, ...args);
    }
  };

  // 错误日志
  const logError = (...args: LogParams): void => {
    if (isEnabled.value) {
      console.error(prefix, ...args);
    }
  };

  // 设置日志级别
  const setLogEnabled = (enabled: boolean): void => {
    isEnabled.value = enabled;
  };

  return {
    log,
    logDebug,
    logInfo,
    logWarn,
    logError,
    setLogEnabled,
    isEnabled,
  };
};

export default useLogger;
