import { SSHStatus, SSHFormConfig } from '@/views/app/ssh/types';

// API状态与前端状态的映射
export const STATUS_MAP: Record<string, SSHStatus> = {
  enable: 'running',
  running: 'running',
  disable: 'stopped',
  disabled: 'stopped',
  stopped: 'stopped',
  failed: 'error',
};

// 状态对应的徽章样式
export const BADGE_STATUS_MAP: Record<SSHStatus, string> = {
  running: 'success',
  stopped: 'danger',
  starting: 'warning',
  stopping: 'warning',
  error: 'danger',
  unhealthy: 'warning',
  loading: 'normal',
};

// 状态对应的颜色
export const COLOR_MAP: Record<SSHStatus, string> = {
  running: 'green',
  stopped: 'red',
  starting: 'orange',
  stopping: 'orange',
  error: 'red',
  unhealthy: 'orange',
  loading: 'gray',
};

// 缓存时间（毫秒）
export const CACHE_TTL = 30000; // 30秒

// 默认表单配置
export const DEFAULT_FORM_CONFIG: SSHFormConfig = {
  port: '22',
  listenAddress: '0.0.0.0',
  permitRootLogin: 'yes',
  passwordAuth: true,
  keyAuth: true,
  reverseLookup: false,
};
