/**
 * Logrotate 分类管理适配器
 * 业务层的分类管理实现
 */

import { LOGROTATE_TYPE } from '@/config/enum';
import {
  CategoryApiAdapter,
  CategoryListParams,
  CategoryListResult,
  CategoryCreateParams,
  CategoryUpdateParams,
  CategoryDeleteParams,
  CategoryItem,
  CategoryManagerConfig,
} from '@/components/idb-tree/types/category';

import {
  getLogrotateCategoriesApi,
  createLogrotateCategoryApi,
  updateLogrotateCategoryApi,
  deleteLogrotateCategoryApi,
} from '@/api/logrotate';

/**
 * Logrotate 分类管理API适配器
 */
export class LogrotateCategoryApiAdapter implements CategoryApiAdapter {
  private logrotateType: LOGROTATE_TYPE;

  constructor(logrotateType: LOGROTATE_TYPE) {
    this.logrotateType = logrotateType;
  }

  async getCategories(params: CategoryListParams): Promise<CategoryListResult> {
    if (!params.host) {
      throw new Error('Host parameter is required');
    }

    const response = await getLogrotateCategoriesApi(
      this.logrotateType,
      params.page || 1,
      params.pageSize || 100,
      params.host!
    );

    const items: CategoryItem[] = response.items.map((item) => ({
      name: item.name,
      type: 'logrotate',
      count: item.count || 0,
      modTime: undefined, // logrotate API 没有返回修改时间
    }));

    return {
      items,
      total: response.total || 0,
      page: response.page || 1,
      pageSize: response.page_size || 10,
    };
  }

  async createCategory(params: CategoryCreateParams): Promise<void> {
    await createLogrotateCategoryApi(
      this.logrotateType,
      params.name,
      params.host
    );
  }

  async updateCategory(params: CategoryUpdateParams): Promise<void> {
    await updateLogrotateCategoryApi(
      this.logrotateType,
      params.oldName,
      params.newName,
      params.host
    );
  }

  async deleteCategory(params: CategoryDeleteParams): Promise<void> {
    await deleteLogrotateCategoryApi(
      this.logrotateType,
      params.name,
      params.host
    );
  }
}

/**
 * 创建 Logrotate 分类管理配置
 */
export function createLogrotateCategoryConfig(
  type: LOGROTATE_TYPE,
  options: Partial<CategoryManagerConfig> = {}
): CategoryManagerConfig {
  return {
    type: 'logrotate',
    apiAdapter: new LogrotateCategoryApiAdapter(type),
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
    emptyText: '暂无分类',
    createText: '立即创建',
    i18nPrefix: 'app.logrotate.category',
    customData: {
      logrotateType: type,
    },
    ...options,
  };
}

/**
 * 预定义的 Logrotate 分类管理配置
 */
export const LOGROTATE_CATEGORY_CONFIGS = {
  LOCAL: (options?: Partial<CategoryManagerConfig>) =>
    createLogrotateCategoryConfig(LOGROTATE_TYPE.Local, options),
  GLOBAL: (options?: Partial<CategoryManagerConfig>) =>
    createLogrotateCategoryConfig(LOGROTATE_TYPE.Global, options),
} as const;
