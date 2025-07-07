/**
 * Service 分类管理适配器
 * 业务层的分类管理实现
 */

import { SERVICE_TYPE } from '@/config/enum';
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
  getServiceCategoryListApi,
  createServiceCategoryApi,
  updateServiceCategoryApi,
  deleteServiceCategoryApi,
} from '@/api/service';

/**
 * Service 分类管理API适配器
 */
export class ServiceCategoryApiAdapter implements CategoryApiAdapter {
  private serviceType: SERVICE_TYPE;

  constructor(serviceType: SERVICE_TYPE) {
    this.serviceType = serviceType;
  }

  async getCategories(params: CategoryListParams): Promise<CategoryListResult> {
    if (!params.host) {
      throw new Error('Host parameter is required');
    }

    const response = await getServiceCategoryListApi({
      type: this.serviceType,
      page: params.page || 1,
      page_size: params.pageSize || 100,
      host: params.host!,
    });

    const items: CategoryItem[] = response.items.map((item) => ({
      name: item.name,
      type: 'service',
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
    await createServiceCategoryApi({
      type: this.serviceType,
      category: params.name,
    });
  }

  async updateCategory(params: CategoryUpdateParams): Promise<void> {
    await updateServiceCategoryApi({
      type: this.serviceType,
      category: params.oldName,
      new_name: params.newName,
    });
  }

  async deleteCategory(params: CategoryDeleteParams): Promise<void> {
    await deleteServiceCategoryApi({
      type: this.serviceType,
      category: params.name,
    });
  }
}

/**
 * 创建 Service 分类管理配置
 */
export function createServiceCategoryConfig(
  type: SERVICE_TYPE,
  options: Partial<CategoryManagerConfig> = {}
): CategoryManagerConfig {
  return {
    type: 'service',
    apiAdapter: new ServiceCategoryApiAdapter(type),
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
    emptyText: '暂无分类',
    createText: '立即创建',
    i18nPrefix: 'app.service.category',
    customData: {
      serviceType: type,
    },
    ...options,
  };
}

/**
 * 预定义的 Service 分类管理配置
 */
export const SERVICE_CATEGORY_CONFIGS = {
  LOCAL: (options?: Partial<CategoryManagerConfig>) =>
    createServiceCategoryConfig(SERVICE_TYPE.Local, options),
  GLOBAL: (options?: Partial<CategoryManagerConfig>) =>
    createServiceCategoryConfig(SERVICE_TYPE.Global, options),
} as const;
