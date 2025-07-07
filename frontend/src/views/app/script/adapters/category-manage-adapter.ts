import {
  createScriptCategoryApi,
  updateScriptCategoryApi,
  deleteScriptCategoryApi,
  getScriptCategoryListApi,
} from '@/api/script';
import { SCRIPT_TYPE } from '@/config/enum';
import type { CategoryManageConfig } from '@/components/idb-tree/types/category';

export function createScriptCategoryManageConfig(
  type: SCRIPT_TYPE
): CategoryManageConfig {
  return {
    api: {
      getList: async (params?: any) => {
        const result = await getScriptCategoryListApi({ type, ...params });
        return {
          page: result.page || 1,
          page_size: result.page_size || 10,
          total: result.total || 0,
          items: result.items || [],
        };
      },
      create: async (params) => {
        await createScriptCategoryApi({
          type,
          category: params.category,
        });
      },
      update: async (params) => {
        await updateScriptCategoryApi({
          type,
          category: params.category,
          new_name: params.new_name,
        });
      },
      delete: async (params) => {
        await deleteScriptCategoryApi({
          type,
          category: params.category,
        });
      },
    },
    params: { type },
    nameField: 'name',
  };
}
