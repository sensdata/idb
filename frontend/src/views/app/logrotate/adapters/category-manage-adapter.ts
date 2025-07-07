import {
  createLogrotateCategoryApi,
  updateLogrotateCategoryApi,
  deleteLogrotateCategoryApi,
  getLogrotateCategoriesApi,
} from '@/api/logrotate';
import { LOGROTATE_TYPE } from '@/config/enum';
import type { CategoryManageConfig } from '@/components/idb-tree/types/category';

export function createLogrotateCategoryManageConfig(
  type: LOGROTATE_TYPE,
  hostId: number
): CategoryManageConfig {
  return {
    api: {
      getList: async (params?: any) => {
        const result = await getLogrotateCategoriesApi(
          type,
          params?.page || 1,
          params?.page_size || 10,
          hostId
        );
        return {
          page: result.page || 1,
          page_size: result.page_size || 10,
          total: result.total || 0,
          items: result.items || [],
        };
      },
      create: async (params) => {
        await createLogrotateCategoryApi(type, params.category, hostId);
      },
      update: async (params) => {
        await updateLogrotateCategoryApi(
          type,
          params.category,
          params.new_name,
          hostId
        );
      },
      delete: async (params) => {
        await deleteLogrotateCategoryApi(type, params.category, hostId);
      },
    },
    nameField: 'name',
  };
}
