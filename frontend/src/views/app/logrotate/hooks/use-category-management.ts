import { ref, nextTick, Ref } from 'vue';
import { useLogger } from '@/hooks/use-logger';

export function useCategoryManagement(
  categoryTreeRef: Ref<any>,
  gridRef: Ref<any>,
  params: Ref<any>
) {
  // 记录最后一次手动设置的分类
  const lastManualCategory = ref<string>('');

  // 添加状态管理，防止并发操作
  const isRefreshing = ref(false);
  const isInitializing = ref(false);

  // 日志记录
  const { logInfo, logError } = useLogger('CategoryManagement');

  // 刷新并选择分类
  const refreshAndSelectCategory = async (category: string) => {
    if (!category || isRefreshing.value) return;

    isRefreshing.value = true;
    try {
      // 记录这是手动设置的分类
      lastManualCategory.value = category;

      // 刷新分类树并选择类别
      if (categoryTreeRef.value) {
        try {
          await categoryTreeRef.value.reload();
          await nextTick();
          categoryTreeRef.value.selectCategory(category);

          // 等待分类选择完成后再设置参数
          await nextTick();
          params.value.category = category;

          // 立即触发数据加载，不依赖监听器
          await nextTick();
          gridRef.value?.reload();
        } catch (e) {
          logError('刷新分类树失败', e as Error);
          // 即使刷新失败，也尝试选择分类并刷新表格
          categoryTreeRef.value.selectCategory(category);
          params.value.category = category;
          await nextTick();
          gridRef.value?.reload();
        }
      } else {
        // 如果分类树不可用，直接设置分类并刷新
        params.value.category = category;
        await nextTick();
        gridRef.value?.reload();
      }
    } finally {
      isRefreshing.value = false;
    }
  };

  // 选择合适的分类：优先使用当前分类，否则使用第一个可用分类
  const selectAppropriateCategory = async (currentCategory: string) => {
    if (currentCategory) {
      // 恢复原始分类
      lastManualCategory.value = currentCategory;
      await refreshAndSelectCategory(currentCategory);
      return;
    }

    if (!categoryTreeRef.value) return;

    // 如果分类树有内容，使用第一个分类
    const categoryTree = categoryTreeRef.value as any;
    if (!categoryTree.items?.value?.length) return;

    const categories = categoryTree.items.value;
    if (categories.length > 0) {
      const firstCategory = categories[0];
      lastManualCategory.value = firstCategory;
      await refreshAndSelectCategory(firstCategory);
    }
  };

  // 重置组件状态
  const resetComponentsState = async () => {
    if (isInitializing.value) return; // 防止重复初始化

    logInfo('开始重置组件状态');
    isInitializing.value = true;
    try {
      // 清空当前状态
      params.value.category = '';
      lastManualCategory.value = '';
      logInfo('已清空分类状态');

      // 刷新分类树，让分类树组件自己处理分类选择
      if (categoryTreeRef.value) {
        logInfo('开始刷新分类树');
        await categoryTreeRef.value.reload();
        await nextTick();
        logInfo('分类树刷新完成，等待分类自动选择');

        // 分类树组件会在 loadCategories 中自动选择第一个分类
        // 这会触发 v-model 绑定，进而触发我们的监听器
      } else {
        logInfo('分类树引用不存在');
      }
    } catch (e) {
      logError('重置组件状态失败', e as Error);
    } finally {
      isInitializing.value = false;
      logInfo('重置组件状态完成');
    }
  };

  return {
    lastManualCategory,
    refreshAndSelectCategory,
    selectAppropriateCategory,
    resetComponentsState,
    isRefreshing,
    isInitializing,
  };
}
