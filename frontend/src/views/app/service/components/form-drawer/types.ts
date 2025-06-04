/**
 * systemd服务配置相关类型定义
 */

/**
 * 解析后的服务配置表单数据
 */
export interface ParsedServiceConfig {
  // 基本信息
  description: string;
  serviceType: string;

  // 执行配置
  execStart: string;
  execStop: string;
  execReload: string;
  workingDirectory: string;

  // 用户权限
  user: string;
  group: string;

  // 环境配置
  environment: string;

  // 重启配置
  restart: string;
  restartSec: number;

  // 超时配置
  timeoutStartSec: number;
  timeoutStopSec: number;

  // 其他未解析的配置
  rawContent: string;
}

/**
 * systemd配置文件中的通用值类型
 */
export type SystemdConfigValue =
  | string
  | string[]
  | number
  | boolean
  | undefined;

/**
 * 通用的键值对映射类型
 */
export type StringKeyValueMapping = Record<string, SystemdConfigValue>;

/**
 * Unit段配置项
 */
export interface UnitSection {
  description?: string;
  after?: string[];
  before?: string[];
  requires?: string[];
  wants?: string[];
  [key: string]: SystemdConfigValue;
}

/**
 * Service段配置项
 */
export interface ServiceSection {
  type?: string;
  execStart?: string;
  execStop?: string;
  execReload?: string;
  restart?: string;
  restartSec?: string;
  workingDirectory?: string;
  user?: string;
  group?: string;
  environment?: string[];
  environmentFile?: string;
  [key: string]: SystemdConfigValue;
}

/**
 * Install段配置项
 */
export interface InstallSection {
  wantedBy?: string[];
  requiredBy?: string[];
  alias?: string[];
  [key: string]: SystemdConfigValue;
}

/**
 * 结构化的服务配置对象
 */
export interface ServiceConfigStructured {
  unit?: UnitSection;
  service?: ServiceSection;
  install?: InstallSection;
  [sectionName: string]: StringKeyValueMapping | undefined;
}
