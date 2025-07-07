/**
 * CategoryManageButton 组件类型定义
 * 用于封装分类管理按钮的通用逻辑
 */

import type { CategoryManageConfig } from '@/components/idb-tree/types/category';

/**
 * CategoryManageButton 组件属性
 */
export interface CategoryManageButtonProps {
  /** 分类管理配置 */
  config: CategoryManageConfig;
  /** 按钮文本（可选，默认使用通用的分类管理文本） */
  buttonText?: string;
  /** 按钮大小 */
  size?: 'mini' | 'small' | 'medium' | 'large';
  /** 按钮类型 */
  type?: 'primary' | 'secondary' | 'outline' | 'dashed' | 'text';
  /** 是否禁用 */
  disabled?: boolean;
}

/**
 * CategoryManageButton 组件事件
 */
export interface CategoryManageButtonEmits {
  /** 分类管理完成事件 */
  (e: 'ok'): void;
  /** 分类管理开始事件 */
  (e: 'manage'): void;
}

/**
 * CategoryManageButton 组件暴露的方法
 */
export interface CategoryManageButtonExposed {
  /** 显示分类管理抽屉 */
  show: () => void;
  /** 隐藏分类管理抽屉 */
  hide: () => void;
}
