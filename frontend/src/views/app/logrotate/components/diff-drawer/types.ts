import type { LOGROTATE_TYPE } from '@/config/enum';

export interface DiffParams extends Record<string, unknown> {
  type: LOGROTATE_TYPE;
  category: string;
  name: string;
  commit: string;
}

export interface DiffDrawerExpose {
  show: (params: DiffParams, onRestoreSuccess?: () => void) => void;
}

export interface ParsedDiff {
  historical: string;
  current: string;
}

// 恢复成功回调函数类型
export type RestoreSuccessCallback = () => void;
