import {
  createCrontabCategoryApi,
  updateCrontabCategoryApi,
  deleteCrontabCategoryApi,
  getCrontabCategoryListApi,
} from '@/api/crontab';
import { CRONTAB_TYPE } from '@/config/enum';
import type { CategoryManageConfig } from '@/components/idb-tree/types/category';

export function createCrontabCategoryManageConfig(
  type: CRONTAB_TYPE
): CategoryManageConfig {
  return {
    api: {
      getList: async (params?: any) => {
        const result = await getCrontabCategoryListApi({ type, ...params });
        return {
          page: result.page || 1,
          page_size: result.page_size || 10,
          total: result.total || 0,
          items: result.items || [],
        };
      },
      create: async (params) => {
        await createCrontabCategoryApi({
          type,
          category: params.category,
        });
      },
      update: async (params) => {
        await updateCrontabCategoryApi({
          type,
          category: params.category,
          new_name: params.new_name,
        });
      },
      delete: async (params) => {
        await deleteCrontabCategoryApi({
          type,
          category: params.category,
        });
      },
    },
    params: { type },
    nameField: 'name',
  };
}
