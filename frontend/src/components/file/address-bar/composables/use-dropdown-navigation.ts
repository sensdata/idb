import { ref, Ref, computed, watch } from 'vue';

export interface DropdownOption {
  value: string;
  label: string;
  isDir?: boolean;
  displayValue?: string;
}

/**
 * 下拉导航组合函数
 * 使用Vue响应式系统来管理选中状态和滚动行为
 */
export default function useDropdownNavigation(
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>
) {
  const currentSelectedIndex = ref(-1);
  const hoverItem = ref<DropdownOption | null>(null);

  // 响应式计算当前选中的选项
  const currentSelectedOption = computed(() => {
    const index = currentSelectedIndex.value;
    return index >= 0 && index < allOptions.value.length
      ? allOptions.value[index]
      : null;
  });

  // 响应式计算是否有选中项
  const hasSelection = computed(() => currentSelectedIndex.value >= 0);

  /**
   * 处理向上键导航
   */
  function handleKeyUp() {
    if (!popupVisible.value) {
      popupVisible.value = true;
      return;
    }

    const optionsLength = allOptions.value.length;
    if (optionsLength === 0) return;

    if (currentSelectedIndex.value > 0) {
      currentSelectedIndex.value--;
    } else {
      // 循环到最后一项
      currentSelectedIndex.value = optionsLength - 1;
    }
  }

  /**
   * 处理向下键导航
   */
  function handleKeyDown() {
    if (!popupVisible.value) {
      popupVisible.value = true;
      return;
    }

    const optionsLength = allOptions.value.length;
    if (optionsLength === 0) return;

    if (currentSelectedIndex.value < optionsLength - 1) {
      currentSelectedIndex.value++;
    } else {
      // 循环到第一项
      currentSelectedIndex.value = 0;
    }
  }

  /**
   * 设置选中项（通过索引）
   */
  function setSelectedIndex(index: number) {
    const optionsLength = allOptions.value.length;
    if (index >= 0 && index < optionsLength) {
      currentSelectedIndex.value = index;
    }
  }

  /**
   * 设置悬停项
   */
  function setHoverItem(option: DropdownOption | null) {
    hoverItem.value = option;
  }

  /**
   * 通过值选择选项
   */
  function selectByValue(value: string) {
    const index = allOptions.value.findIndex(
      (option) => option.value === value
    );
    if (index >= 0) {
      currentSelectedIndex.value = index;
    }
  }

  /**
   * 重置选中状态
   */
  function resetSelection() {
    currentSelectedIndex.value = -1;
    hoverItem.value = null;
  }

  /**
   * 获取当前选中项的值
   */
  function getCurrentSelectedValue(): string | null {
    return currentSelectedOption.value?.value || null;
  }

  // 监听选项变化，重置无效的选中状态
  watch(allOptions, (newOptions) => {
    if (currentSelectedIndex.value >= newOptions.length) {
      currentSelectedIndex.value = -1;
    }
  });

  // 监听弹窗关闭，重置选中状态
  watch(popupVisible, (visible) => {
    if (!visible) {
      resetSelection();
    }
  });

  return {
    // 状态
    currentSelectedIndex,
    currentSelectedOption,
    hoverItem,
    hasSelection,

    // 方法
    handleKeyUp,
    handleKeyDown,
    setSelectedIndex,
    setHoverItem,
    selectByValue,
    resetSelection,
    getCurrentSelectedValue,
  };
}
