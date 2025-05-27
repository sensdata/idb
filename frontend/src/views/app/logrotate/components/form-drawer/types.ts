import type { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';
import type { LogrotateEntity } from '@/entity/Logrotate';

// 权限配置接口
export interface PermissionConfig {
  mode: string;
  user: string;
  group: string;
  description: string;
}

// 权限访问类型
export interface PermissionAccess {
  owner: string[];
  group: string[];
  other: string[];
}

// 权限选项
export interface PermissionOption {
  label: string;
  value: string;
}

// 权限解析结果
export interface ParsedPermission {
  isValid: boolean;
  mode: string;
  user: string;
  group: string;
  access: PermissionAccess;
}

export interface FormData {
  name: string;
  category: string;
  path: string;
  frequency: LOGROTATE_FREQUENCY;
  count: number;
  compress: boolean;
  delayCompress: boolean;
  missingOk: boolean;
  notIfEmpty: boolean;
  create: string;
  preRotate: string;
  postRotate: string;
}

export interface ShowParams {
  name?: string;
  type?: LOGROTATE_TYPE;
  category?: string;
  isEdit?: boolean;
  record?: LogrotateEntity;
}

export interface ApiError {
  message: string;
  code?: string;
  details?: any;
}

export type SubmitData = {
  name: string;
  type: LOGROTATE_TYPE;
  category: string;
  form: Array<{ key: string; value: string }>;
};

export type ActiveMode = 'form' | 'raw';

export interface FormRules {
  name: Array<{ required?: boolean; message: string; pattern?: RegExp }>;
  category: Array<{ required?: boolean; message: string }>;
  path: Array<{ required?: boolean; message: string }>;
  count: Array<{
    required?: boolean;
    message: string;
    type?: string;
    min?: number;
  }>;
}

export type CategoryValue =
  | string
  | number
  | boolean
  | Record<string, any>
  | (string | number | boolean | Record<string, any>)[];

export type SelectOption = { label: string; value: string };
