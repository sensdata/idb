/**
 * 分类管理统一接口和类型定义
 * 用于支持不同应用的分类管理功能
 */

import { Component } from 'vue';

/**
 * 分类管理支持的应用类型
 * 通用组件不限制具体的业务类型
 */
export type CategoryAppType = string;

/**
 * 分类数据结构
 */
export interface CategoryItem {
  /** 分类名称 */
  name: string;
  /** 分类下的项目数量 */
  count?: number;
  /** 分类类型 */
  type: CategoryAppType;
  /** 最后修改时间 */
  modTime?: string;
  /** 扩展数据 */
  [key: string]: any;
}

/**
 * 分类列表查询参数
 */
export interface CategoryListParams {
  /** 应用类型 */
  type: CategoryAppType;
  /** 主机ID */
  host: number;
  /** 分页页码 */
  page?: number;
  /** 分页大小 */
  pageSize?: number;
  /** 额外参数 */
  [key: string]: any;
}

/**
 * 分类列表查询结果
 */
export interface CategoryListResult {
  /** 分类列表 */
  items: CategoryItem[];
  /** 总数 */
  total: number;
  /** 当前页码 */
  page: number;
  /** 分页大小 */
  pageSize: number;
}

/**
 * 分类创建参数
 */
export interface CategoryCreateParams {
  /** 应用类型 */
  type: CategoryAppType;
  /** 分类名称 */
  name: string;
  /** 主机ID */
  host: number;
  /** 额外参数 */
  [key: string]: any;
}

/**
 * 分类更新参数
 */
export interface CategoryUpdateParams {
  /** 应用类型 */
  type: CategoryAppType;
  /** 原分类名称 */
  oldName: string;
  /** 新分类名称 */
  newName: string;
  /** 主机ID */
  host: number;
  /** 额外参数 */
  [key: string]: any;
}

/**
 * 分类删除参数
 */
export interface CategoryDeleteParams {
  /** 应用类型 */
  type: CategoryAppType;
  /** 分类名称 */
  name: string;
  /** 主机ID */
  host: number;
  /** 额外参数 */
  [key: string]: any;
}

/**
 * 分类管理 API 适配器接口
 */
export interface CategoryApiAdapter {
  /** 获取分类列表 */
  getCategories: (params: CategoryListParams) => Promise<CategoryListResult>;
  /** 创建分类 */
  createCategory: (params: CategoryCreateParams) => Promise<void>;
  /** 更新分类 */
  updateCategory: (params: CategoryUpdateParams) => Promise<void>;
  /** 删除分类 */
  deleteCategory: (params: CategoryDeleteParams) => Promise<void>;
}

/**
 * 分类表单字段定义
 */
export interface CategoryFormField {
  /** 字段名 */
  name: string;
  /** 字段标签 */
  label: string;
  /** 字段类型 */
  type: 'input' | 'select' | 'textarea' | 'checkbox' | 'radio';
  /** 默认值 */
  defaultValue?: any;
  /** 是否必填 */
  required?: boolean;
  /** 验证规则 */
  validation?: {
    pattern?: string;
    maxLength?: number;
    minLength?: number;
    message?: string;
  };
  /** 选项（用于 select、radio） */
  options?: Array<{ label: string; value: any }>;
  /** 提示信息 */
  placeholder?: string;
  /** 帮助文本 */
  hint?: string;
}

/**
 * 分类管理配置
 */
export interface CategoryManagerConfig {
  /** 应用类型标识 */
  type: CategoryAppType;
  /** API 适配器 */
  apiAdapter: CategoryApiAdapter;
  /** 表单字段定义 */
  formFields?: CategoryFormField[];
  /** 表单组件 */
  formComponent?: Component;
  /** 自定义图标 */
  icon?: string | Component;
  /** 是否允许创建 */
  allowCreate?: boolean;
  /** 是否允许编辑 */
  allowEdit?: boolean;
  /** 是否允许删除 */
  allowDelete?: boolean;
  /** 空状态文本 */
  emptyText?: string;
  /** 创建按钮文本 */
  createText?: string;
  /** 国际化键前缀 */
  i18nPrefix?: string;
  /** 自定义数据 */
  customData?: Record<string, any>;
}

/**
 * 分类管理事件
 */
export interface CategoryManagerEvents {
  /** 分类选择变化 */
  'category-select': (category: CategoryItem | null) => void;
  /** 分类创建成功 */
  'category-created': (category: CategoryItem) => void;
  /** 分类更新成功 */
  'category-updated': (category: CategoryItem) => void;
  /** 分类删除成功 */
  'category-deleted': (categoryName: string) => void;
  /** 分类列表刷新 */
  'categories-refreshed': (categories: CategoryItem[]) => void;
}

/**
 * 分类管理状态
 */
export interface CategoryManagerState {
  /** 分类列表 */
  categories: CategoryItem[];
  /** 当前选中的分类 */
  selectedCategory: CategoryItem | null;
  /** 是否正在加载 */
  loading: boolean;
  /** 错误信息 */
  error: string | null;
}

/**
 * 分类管理操作
 */
export interface CategoryManagerActions {
  /** 加载分类列表 */
  loadCategories: () => Promise<void>;
  /** 选择分类 */
  selectCategory: (category: CategoryItem | null) => void;
  /** 创建分类 */
  createCategory: (name: string, data?: any) => Promise<void>;
  /** 更新分类 */
  updateCategory: (
    oldName: string,
    newName: string,
    data?: any
  ) => Promise<void>;
  /** 删除分类 */
  deleteCategory: (name: string) => Promise<void>;
  /** 刷新分类列表 */
  refreshCategories: () => Promise<void>;
}

/**
 * 简化的分类数据结构（兼容性）
 */
export interface Category {
  name: string;
  [key: string]: any;
}

/**
 * 简化的分类管理 API 接口（兼容性）
 */
export interface CategoryManageAPI {
  /** 获取分类列表 */
  getList: (params?: any) => Promise<{
    page: number;
    page_size: number;
    total?: number;
    items: Category[];
  }>;
  /** 创建分类 */
  create: (params: { category: string; [key: string]: any }) => Promise<void>;
  /** 更新分类 */
  update: (params: {
    category: string;
    new_name: string;
    [key: string]: any;
  }) => Promise<void>;
  /** 删除分类 */
  delete: (params: { category: string; [key: string]: any }) => Promise<void>;
}

/**
 * 简化的分类管理配置（兼容性）
 */
export interface CategoryManageConfig {
  /** 分类管理API接口 */
  api: CategoryManageAPI;
  /** 额外的请求参数 */
  params?: Record<string, any>;
  /** 分类名称字段 */
  nameField?: string;
  /** 表格列配置 */
  columns?: Array<{
    dataIndex: string;
    title: string;
    width?: number;
    ellipsis?: boolean;
    align?: 'left' | 'center' | 'right';
    slotName?: string;
  }>;
}
