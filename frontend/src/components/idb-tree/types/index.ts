/**
 * IdbTree 组件类型定义
 * 统一的树形组件数据结构和接口
 */

import { Component } from 'vue';
import { CategoryManagerConfig } from './category';

/**
 * 树项目数据结构
 */
export interface TreeItem {
  /** 唯一标识 */
  id: string | number;
  /** 显示文本 */
  label: string;
  /** 图标组件或图标名 */
  icon?: string | Component;
  /** 是否禁用 */
  disabled?: boolean;
  /** 是否加载中 */
  loading?: boolean;
  /** 是否展开（仅树模式） */
  expanded?: boolean;
  /** 子项列表（仅树模式） */
  children?: TreeItem[];
  /** 是否隐藏 */
  hidden?: boolean;
  /** 扩展数据 */
  [key: string]: any;
}

/**
 * 组件模式
 */
export type TreeMode = 'list' | 'tree';

/**
 * 选中值类型
 */
export type SelectedValue = string | number | TreeItem | null;

/**
 * 组件Props接口
 */
export interface TreeProps {
  /** 组件模式 */
  mode?: TreeMode;
  /** 数据源 */
  items: TreeItem[];
  /** 选中值 */
  selected?: SelectedValue;
  /** 是否多选 */
  multiple?: boolean;
  /** 是否显示图标 */
  showIcon?: boolean;
  /** 是否显示展开/折叠按钮 */
  showToggle?: boolean;
  /** 是否显示隐藏项 */
  showHidden?: boolean;
  /** 空状态文本 */
  emptyText?: string;
  /** 创建按钮文本 */
  createText?: string;
  /** 是否显示创建按钮 */
  showCreate?: boolean;
  /** 默认图标 */
  defaultIcon?: string | Component;
  /** 展开图标 */
  expandIcon?: string | Component;
  /** 收起图标 */
  collapseIcon?: string | Component;
  /** 分类管理配置 */
  categoryConfig?: CategoryManagerConfig;
  /** 是否启用分类管理模式 */
  enableCategoryManagement?: boolean;
  /** 主机ID（用于分类管理） */
  hostId?: number;
}

/**
 * 组件Events接口
 */
export interface TreeEvents {
  /** 选中项变化 */
  'update:selected': (value: SelectedValue) => void;
  /** 点击项 */
  'select': (item: TreeItem) => void;
  /** 双击项 */
  'doubleClick': (item: TreeItem) => void;
  /** 展开/折叠状态变化 */
  'toggle': (item: TreeItem, expanded: boolean) => void;
  /** 点击创建按钮 */
  'create': () => void;
  /** 分类创建成功 */
  'category-created': (categoryName: string) => void;
  /** 分类更新成功 */
  'category-updated': (oldName: string, newName: string) => void;
  /** 分类删除成功 */
  'category-deleted': (categoryName: string) => void;
  /** 分类管理操作 */
  'category-manage': () => void;
}

/**
 * 组件暴露的方法
 */
export interface TreeExposed {
  /** 刷新数据 */
  refresh: () => void;
  /** 选择指定项 */
  selectItem: (id: string | number) => void;
  /** 展开指定项 */
  expandItem: (id: string | number) => void;
  /** 折叠指定项 */
  collapseItem: (id: string | number) => void;
  /** 获取选中项 */
  getSelected: () => TreeItem | null;
  /** 刷新分类列表 */
  refreshCategories?: () => Promise<void>;
  /** 选择分类 */
  selectCategory?: (categoryName: string) => void;
  /** 创建分类 */
  createCategory?: (name: string, data?: any) => Promise<void>;
  /** 更新分类 */
  updateCategory?: (
    oldName: string,
    newName: string,
    data?: any
  ) => Promise<void>;
  /** 删除分类 */
  deleteCategory?: (name: string) => Promise<void>;
  /** 获取分类列表 */
  getCategories?: () => string[];
  /** 获取选中的分类 */
  getSelectedCategory?: () => string | null;
}
