import { ref } from 'vue';
import { LOGROTATE_TYPE } from '@/config/enum';
import useCurrentHost from '@/hooks/current-host';
import { getLogrotateCategoriesApi } from '@/api/logrotate';
import type { SelectOption } from '../types';

export function useCategories() {
  const categoryLoading = ref(false);
  const categoryOptions = ref<SelectOption[]>([]);

  const { currentHostId } = useCurrentHost();

  const fetchCategories = async (type: LOGROTATE_TYPE) => {
    if (!currentHostId.value) return;

    categoryLoading.value = true;
    try {
      const response = await getLogrotateCategoriesApi(
        type,
        1,
        100,
        currentHostId.value
      );
      categoryOptions.value = response.items.map((category: any) => ({
        label: category.name,
        value: category.name,
      }));
    } catch (error) {
      categoryOptions.value = [];
    } finally {
      categoryLoading.value = false;
    }
  };

  const handleCategoryChange = () => {
    // 处理分类变化逻辑
  };

  const handleCategoryVisibleChange = (
    visible: boolean,
    type: LOGROTATE_TYPE
  ) => {
    if (visible) {
      fetchCategories(type);
    }
  };

  const ensureCategoryInOptions = (category: string) => {
    if (
      category &&
      !categoryOptions.value.find((option) => option.value === category)
    ) {
      categoryOptions.value.unshift({
        label: category,
        value: category,
      });
    }
  };

  const refreshCategories = (type: LOGROTATE_TYPE) => {
    fetchCategories(type);
  };

  return {
    categoryLoading,
    categoryOptions,
    fetchCategories,
    handleCategoryChange,
    handleCategoryVisibleChange,
    ensureCategoryInOptions,
    refreshCategories,
  };
}
