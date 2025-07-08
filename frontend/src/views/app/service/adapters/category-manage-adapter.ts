import {
  createServiceCategoryApi,
  updateServiceCategoryApi,
  deleteServiceCategoryApi,
  getServiceCategoryListApi,
} from '@/api/service';
import { SERVICE_TYPE } from '@/config/enum';
import type { CategoryManageConfig } from '@/components/idb-tree/types/category';

export function createServiceCategoryManageConfig(
  type: SERVICE_TYPE,
  hostId: number
): CategoryManageConfig {
  return {
    api: {
      getList: async (params?: any) => {
        const result = await getServiceCategoryListApi({
          type,
          page: params?.page || 1,
          page_size: params?.page_size || 10,
          host: hostId,
        });
        return {
          page: result.page || 1,
          page_size: result.page_size || 10,
          total: result.total || 0,
          items: result.items || [],
        };
      },
      create: async (params) => {
        await createServiceCategoryApi({
          type,
          category: params.category,
        });
      },
      update: async (params) => {
        await updateServiceCategoryApi({
          type,
          category: params.category,
          new_name: params.new_name,
        });
      },
      delete: async (params) => {
        await deleteServiceCategoryApi({
          type,
          category: params.category,
        });
      },
    },
    nameField: 'name',
  };
}
