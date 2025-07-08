/**
 * 分类管理组合函数
 * 提供统一的分类管理逻辑和状态管理
 */

import { ref, computed, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { useLogger } from '@/composables/use-logger';
import { useConfirm } from '@/composables/confirm';
import {
  CategoryApiAdapter,
  CategoryItem,
  CategoryManagerState,
  CategoryManagerActions,
  CategoryManagerConfig,
  CategoryListParams,
} from '../types/category';

export interface UseCategoryManagerOptions {
  /** 分类管理配置 */
  config: CategoryManagerConfig;
  /** 主机ID */
  hostId: number;
  /** 是否自动加载 */
  autoLoad?: boolean;
  /** 初始选中的分类 */
  initialSelectedCategory?: string;
}

export function useCategoryManager(options: UseCategoryManagerOptions) {
  const { t } = useI18n();
  const { confirm } = useConfirm();
  const { logInfo, logError } = useLogger('CategoryManager');

  // 响应式状态
  const state = ref<CategoryManagerState>({
    categories: [],
    selectedCategory: null,
    loading: false,
    error: null,
  });

  // 计算属性
  const categoryNames = computed(() =>
    state.value.categories.map((cat) => cat.name)
  );

  const selectedCategoryName = computed(
    () => state.value.selectedCategory?.name || ''
  );

  const hasCategories = computed(() => state.value.categories.length > 0);

  const isEmpty = computed(
    () => !state.value.loading && state.value.categories.length === 0
  );

  // 内部方法
  const createApiParams = (
    extraParams: Partial<CategoryListParams> = {}
  ): CategoryListParams => ({
    type: options.config.type,
    host: options.hostId,
    page: 1,
    pageSize: 1000,
    ...extraParams,
  });

  const handleError = (error: any, operation: string) => {
    const message = error?.message || `${operation} failed`;
    state.value.error = message;
    logError(`${operation} failed:`, error);
    Message.error(message);
  };

  const findCategoryByName = (name: string): CategoryItem | undefined => {
    return state.value.categories.find((cat) => cat.name === name);
  };

  const ensureCategoryExists = (name: string): void => {
    if (!findCategoryByName(name)) {
      state.value.categories.push({
        name,
        type: options.config.type,
        count: 0,
      });
    }
  };

  // 分类管理操作
  const actions: CategoryManagerActions = {
    // 加载分类列表
    async loadCategories() {
      if (state.value.loading) {
        logInfo('Categories are already loading, skipping duplicate request');
        return;
      }

      state.value.loading = true;
      state.value.error = null;

      try {
        const params = createApiParams();
        const result = await options.config.apiAdapter.getCategories(params);

        state.value.categories = result.items;
        logInfo(`Loaded ${result.items.length} categories`);

        // 如果有初始选中的分类，尝试选择它
        if (options.initialSelectedCategory) {
          const category = findCategoryByName(options.initialSelectedCategory);
          if (category) {
            state.value.selectedCategory = category;
          } else if (state.value.categories.length > 0) {
            // 如果初始分类不存在，选择第一个分类
            state.value.selectedCategory = state.value.categories[0];
          }
        } else if (
          state.value.categories.length > 0 &&
          !state.value.selectedCategory
        ) {
          // 如果没有选中任何分类，选择第一个
          state.value.selectedCategory = state.value.categories[0];
        }
      } catch (error) {
        handleError(error, 'Load categories');
      } finally {
        state.value.loading = false;
      }
    },

    // 选择分类
    selectCategory(category: CategoryItem | null) {
      const oldCategory = state.value.selectedCategory;
      state.value.selectedCategory = category;

      if (oldCategory?.name !== category?.name) {
        logInfo(`Category selected: ${category?.name || 'none'}`);
      }
    },

    // 创建分类
    async createCategory(name: string, data: any = {}) {
      if (!name || !name.trim()) {
        Message.error(t('common.validation.name_required'));
        return;
      }

      // 检查分类是否已存在
      if (findCategoryByName(name)) {
        Message.error(t('common.validation.name_exists'));
        return;
      }

      state.value.loading = true;
      try {
        const params = {
          type: options.config.type,
          name: name.trim(),
          host: options.hostId,
          ...data,
        };

        await options.config.apiAdapter.createCategory(params);

        Message.success(t('common.message.create_success'));
        logInfo(`Category created: ${name}`);

        // 重新加载分类列表
        await actions.loadCategories();

        // 选择新创建的分类
        const newCategory = findCategoryByName(name);
        if (newCategory) {
          actions.selectCategory(newCategory);
        }
      } catch (error) {
        handleError(error, 'Create category');
      } finally {
        state.value.loading = false;
      }
    },

    // 更新分类
    async updateCategory(oldName: string, newName: string, data: any = {}) {
      if (!oldName || !newName || !newName.trim()) {
        Message.error(t('common.validation.name_required'));
        return;
      }

      // 检查新名称是否已存在
      if (oldName !== newName && findCategoryByName(newName)) {
        Message.error(t('common.validation.name_exists'));
        return;
      }

      state.value.loading = true;
      try {
        const params = {
          type: options.config.type,
          oldName,
          newName: newName.trim(),
          host: options.hostId,
          ...data,
        };

        await options.config.apiAdapter.updateCategory(params);

        Message.success(t('common.message.update_success'));
        logInfo(`Category updated: ${oldName} -> ${newName}`);

        // 重新加载分类列表
        await actions.loadCategories();

        // 选择更新后的分类
        const updatedCategory = findCategoryByName(newName);
        if (updatedCategory) {
          actions.selectCategory(updatedCategory);
        }
      } catch (error) {
        handleError(error, 'Update category');
      } finally {
        state.value.loading = false;
      }
    },

    // 删除分类
    async deleteCategory(name: string) {
      if (!name) {
        Message.error(t('common.validation.name_required'));
        return;
      }

      const category = findCategoryByName(name);
      if (!category) {
        Message.error(t('common.validation.category_not_found'));
        return;
      }

      try {
        const confirmMessage = t('common.confirm.delete_category', { name });
        await confirm(confirmMessage);
      } catch {
        // 用户取消删除
        return;
      }

      state.value.loading = true;
      try {
        const params = {
          type: options.config.type,
          name,
          host: options.hostId,
        };

        await options.config.apiAdapter.deleteCategory(params);

        Message.success(t('common.message.delete_success'));
        logInfo(`Category deleted: ${name}`);

        // 如果删除的是当前选中的分类，清空选择
        if (state.value.selectedCategory?.name === name) {
          state.value.selectedCategory = null;
        }

        // 重新加载分类列表
        await actions.loadCategories();
      } catch (error) {
        handleError(error, 'Delete category');
      } finally {
        state.value.loading = false;
      }
    },

    // 刷新分类列表
    async refreshCategories() {
      logInfo('Refreshing categories...');
      await actions.loadCategories();
    },
  };

  // 监听主机ID变化
  watch(
    () => options.hostId,
    async (newHostId) => {
      if (newHostId) {
        logInfo(`Host ID changed to: ${newHostId}`);
        state.value.selectedCategory = null;
        await actions.loadCategories();
      }
    },
    { immediate: false }
  );

  // 自动加载
  if (options.autoLoad && options.hostId) {
    nextTick(() => {
      actions.loadCategories();
    });
  }

  return {
    // 状态
    state: computed(() => state.value),
    categories: computed(() => state.value.categories),
    selectedCategory: computed(() => state.value.selectedCategory),
    loading: computed(() => state.value.loading),
    error: computed(() => state.value.error),

    // 计算属性
    categoryNames,
    selectedCategoryName,
    hasCategories,
    isEmpty,

    // 操作方法
    ...actions,

    // 工具方法
    findCategoryByName,
    ensureCategoryExists,
  };
}

/**
 * 分类管理组合函数的简化版本
 * 适用于只需要基本功能的场景
 */
export function useSimpleCategoryManager(
  apiAdapter: CategoryApiAdapter,
  hostId: number,
  autoLoad = true
) {
  const config = {
    type: 'custom' as const,
    apiAdapter,
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
  };

  return useCategoryManager({
    config,
    hostId,
    autoLoad,
  });
}
