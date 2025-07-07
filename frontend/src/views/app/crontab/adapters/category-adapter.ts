/**
 * Crontab 分类管理适配器
 * 业务层的分类管理实现
 */

import { CRONTAB_TYPE } from '@/config/enum';
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
  getCrontabCategoryListApi,
  createCrontabCategoryApi,
  updateCrontabCategoryApi,
  deleteCrontabCategoryApi,
  CrontabCategoryListApiParams,
} from '@/api/crontab';

/**
 * Crontab 分类管理API适配器
 */
export class CrontabCategoryApiAdapter implements CategoryApiAdapter {
  private crontabType: CRONTAB_TYPE;

  constructor(crontabType: CRONTAB_TYPE) {
    this.crontabType = crontabType;
  }

  async getCategories(params: CategoryListParams): Promise<CategoryListResult> {
    if (!params.host) {
      throw new Error('Host parameter is required');
    }

    const apiParams: CrontabCategoryListApiParams = {
      type: this.crontabType,
      page: params.page || 1,
      page_size: params.pageSize || 100,
      host: params.host!,
    };

    const response = await getCrontabCategoryListApi(apiParams);

    const items: CategoryItem[] = response.items.map((item) => ({
      name: item.name,
      type: 'crontab',
      count: item.size || 0,
      modTime: item.mod_time,
    }));

    return {
      items,
      total: response.total || 0,
      page: response.page || 1,
      pageSize: response.page_size || 10,
    };
  }

  async createCategory(params: CategoryCreateParams): Promise<void> {
    await createCrontabCategoryApi({
      type: this.crontabType,
      category: params.name,
    });
  }

  async updateCategory(params: CategoryUpdateParams): Promise<void> {
    await updateCrontabCategoryApi({
      type: this.crontabType,
      category: params.oldName,
      new_name: params.newName,
    });
  }

  async deleteCategory(params: CategoryDeleteParams): Promise<void> {
    await deleteCrontabCategoryApi({
      type: this.crontabType,
      category: params.name,
    });
  }
}

/**
 * 创建 Crontab 分类管理配置
 */
export function createCrontabCategoryConfig(
  type: CRONTAB_TYPE,
  options: Partial<CategoryManagerConfig> = {}
): CategoryManagerConfig {
  return {
    type: 'crontab',
    apiAdapter: new CrontabCategoryApiAdapter(type),
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
    emptyText: '暂无分类',
    createText: '立即创建',
    i18nPrefix: 'app.crontab.category',
    customData: {
      crontabType: type,
    },
    ...options,
  };
}

/**
 * 预定义的 Crontab 分类管理配置
 */
export const CRONTAB_CATEGORY_CONFIGS = {
  LOCAL: (options?: Partial<CategoryManagerConfig>) =>
    createCrontabCategoryConfig(CRONTAB_TYPE.Local, options),
  GLOBAL: (options?: Partial<CategoryManagerConfig>) =>
    createCrontabCategoryConfig(CRONTAB_TYPE.Global, options),
} as const;
