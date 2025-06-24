/**
 * 地址栏组件相关类型定义
 */

/**
 * 事件发射函数类型
 */
export interface DropdownOption {
  value: string;
  label: string;
  isDir?: boolean;
  displayValue?: string;
}

export interface EmitFn {
  (e: 'goto', path: string): void;
  (e: 'search', params: { path: string; word: string }): void;
  (e: 'clear'): void;
}

export interface SearchParams {
  path: string;
  word: string;
}

export interface AddressBarProps {
  path: string;
  items?: Array<{
    name: string;
    is_dir: boolean;
  }>;
}
