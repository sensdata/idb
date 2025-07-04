/**
 * 非组件文件日志工具
 * 封装控制台方法，仅在开发环境下输出日志
 */

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

class Logger {
  private isDev: boolean;

  private prefix: string;

  constructor(module?: string) {
    this.isDev = import.meta.env.DEV;
    this.prefix = module ? `[${module}]` : '';
  }

  // 标准日志
  log(...args: LogParams): void {
    if (this.isDev) {
      // eslint-disable-next-line no-console
      console.log(this.prefix, ...args);
    }
  }

  // 调试日志
  logDebug(...args: LogParams): void {
    if (this.isDev) {
      // eslint-disable-next-line no-console
      console.log(this.prefix, ...args);
    }
  }

  // 信息日志
  logInfo(...args: LogParams): void {
    if (this.isDev) {
      // eslint-disable-next-line no-console
      console.info(this.prefix, ...args);
    }
  }

  // 警告日志
  logWarn(...args: LogParams): void {
    if (this.isDev) {
      console.warn(this.prefix, ...args);
    }
  }

  // 错误日志
  logError(...args: LogParams): void {
    console.error(this.prefix, ...args);
  }
}

/**
 * 创建日志实例
 * @param module - 模块名称
 * @returns Logger 实例
 */
export const createLogger = (module?: string): Logger => {
  return new Logger(module);
};

export default createLogger;
