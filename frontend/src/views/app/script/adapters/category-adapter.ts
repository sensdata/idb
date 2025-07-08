/**
 * Script Category Adapter
 * Adapts script category API to work with idb-tree component
 */
import { SCRIPT_TYPE } from '@/config/enum';
import {
  CategoryApiAdapter,
  CategoryListParams,
  CategoryListResult,
  CategoryCreateParams,
  CategoryUpdateParams,
  CategoryDeleteParams,
  CategoryManagerConfig,
  CategoryAppType,
} from '@/components/idb-tree/types/category';
import {
  getScriptCategoryListApi,
  createScriptCategoryApi,
  updateScriptCategoryApi,
  deleteScriptCategoryApi,
} from '@/api/script';

export class ScriptCategoryApiAdapter implements CategoryApiAdapter {
  private scriptType: SCRIPT_TYPE;

  constructor(scriptType: SCRIPT_TYPE) {
    this.scriptType = scriptType;
  }

  async getCategories(params: CategoryListParams): Promise<CategoryListResult> {
    const response = await getScriptCategoryListApi({
      page: params.page || 1,
      page_size: params.pageSize || 1000,
      type: this.scriptType,
    });

    return {
      items: response.items.map((item) => ({
        name: item.name,
        type: 'script',
        modTime: item.mod_time,
        size: item.size,
        source: item.source,
      })),
      total: response.total || 0,
      page: response.page,
      pageSize: response.page_size,
    };
  }

  async createCategory(params: CategoryCreateParams): Promise<void> {
    await createScriptCategoryApi({
      type: this.scriptType,
      category: params.name,
    });
  }

  async updateCategory(params: CategoryUpdateParams): Promise<void> {
    await updateScriptCategoryApi({
      type: this.scriptType,
      category: params.oldName,
      new_name: params.newName,
    });
  }

  async deleteCategory(params: CategoryDeleteParams): Promise<void> {
    await deleteScriptCategoryApi({
      type: this.scriptType,
      category: params.name,
    });
  }
}

export function createScriptCategoryConfig(
  scriptType: SCRIPT_TYPE,
  options: Partial<CategoryManagerConfig> = {}
): CategoryManagerConfig {
  return {
    type: 'script' as CategoryAppType,
    apiAdapter: new ScriptCategoryApiAdapter(scriptType),
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
    emptyText: '暂无分组',
    createText: '立即创建',
    ...options,
  };
}
