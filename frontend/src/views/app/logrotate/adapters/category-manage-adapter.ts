import { LOGROTATE_TYPE } from '@/config/enum';
import type { CategoryManageConfig } from '@/components/idb-tree/types/category';
import { LogrotateCategoryApiAdapter } from './category-adapter';

export function createLogrotateCategoryManageConfig(
  type: LOGROTATE_TYPE,
  hostId: number
): CategoryManageConfig {
  const apiAdapter = new LogrotateCategoryApiAdapter(type);

  return {
    api: {
      getList: async (params?: any) => {
        const result = await apiAdapter.getCategories({
          type: 'logrotate',
          host: hostId,
          page: params?.page || 1,
          pageSize: params?.page_size || 10,
        });

        return {
          page: result.page || 1,
          page_size: result.pageSize || 10,
          total: result.total || 0,
          items: (result.items || []).map((item) => ({ name: item.name })),
        };
      },
      create: async (params) => {
        await apiAdapter.createCategory({
          type: 'logrotate',
          name: params.category,
          host: hostId,
        });
      },
      update: async (params) => {
        await apiAdapter.updateCategory({
          type: 'logrotate',
          oldName: params.category,
          newName: params.new_name,
          host: hostId,
        });
      },
      delete: async (params) => {
        await apiAdapter.deleteCategory({
          type: 'logrotate',
          name: params.category,
          host: hostId,
        });
      },
    },
    nameField: 'name',
  };
}
