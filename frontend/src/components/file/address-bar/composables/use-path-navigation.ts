import { ref, Ref } from 'vue';
import type { ComponentPublicInstance } from 'vue';
import { findCommonPrefix } from '../utils';
import { DropdownOption } from './use-dropdown-navigation';
import { EmitFn } from '../types';

// 辅助函数：解析当前路径
function parseCurrentPath(currentPath: string) {
  const lastSlashIndex = currentPath.lastIndexOf('/');
  const basePath =
    lastSlashIndex >= 0 ? currentPath.substring(0, lastSlashIndex + 1) : '';
  return { basePath, lastSlashIndex };
}

// 辅助函数：构建选项值
function buildSelectedValue(item: DropdownOption): string {
  return item.isDir ? `${item.value}/` : item.value;
}

export default function usePathNavigation(
  value: Ref<string>,
  inputRef: Ref<ComponentPublicInstance | HTMLElement | null>,
  emit: EmitFn,
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>,
  isSearching: Ref<boolean>,
  triggerByTab: Ref<boolean>,
  currentPath: Ref<string> // 添加当前路径参数
) {
  const lastTabTime = ref(0);
  const userTyping = ref(true);

  const handleHome = () => {
    value.value = '';
    emit('goto', '/');
  };

  const handleClear = () => {
    value.value = '';
    emit('clear');
    emit('search', {
      path: currentPath.value, // 使用当前路径
      word: '',
    });
    popupVisible.value = false;
    isSearching.value = false;
  };

  const handleOptionClick = (item: DropdownOption) => {
    popupVisible.value = false;

    const { basePath } = parseCurrentPath(value.value);
    const selectedValue = buildSelectedValue(item);

    value.value = basePath + selectedValue;

    // 清理下拉状态，但不触发导航
    allOptions.value = [];
    isSearching.value = false;
    triggerByTab.value = false;
  };

  /**
   * 处理Tab键自动补全
   */
  const handleTab = () => {
    const currentTime = Date.now();
    const isDoubleTab = currentTime - lastTabTime.value < 300;
    lastTabTime.value = currentTime;

    const endsWithSlash = value.value.endsWith('/');

    // 没有选项时触发搜索
    if (allOptions.value.length === 0) {
      triggerByTab.value = true;
      isSearching.value = true;

      // 使用当前路径和输入值作为搜索词
      const searchPath = currentPath.value;
      const searchTerm = value.value.trim();

      emit('search', {
        path: searchPath,
        word: searchTerm,
      });
      return;
    }

    // 单个选项时自动选择并导航
    if (allOptions.value.length === 1) {
      const option = allOptions.value[0];
      const selectedValue = buildSelectedValue(option);

      // 构建完整路径：当前路径 + 选中的选项
      const targetPath = `${currentPath.value}/${selectedValue}`.replace(
        /\/+/g,
        '/'
      );

      // 清理下拉状态
      allOptions.value = [];
      isSearching.value = false;
      triggerByTab.value = false;
      popupVisible.value = false;
      value.value = '';

      // 触发导航
      emit('goto', targetPath);
      return;
    }

    // 多个选项时尝试公共前缀补全
    const commonPrefix = findCommonPrefix(allOptions.value);

    if (commonPrefix) {
      const { basePath } = parseCurrentPath(value.value);
      value.value = basePath + commonPrefix;
    }

    if (isDoubleTab || endsWithSlash || commonPrefix) {
      popupVisible.value = true;
    }

    userTyping.value = false;
  };

  const handleBlur = (path: string) => {
    if (value.value.trim() === '') {
      value.value = path;
    }
  };

  return {
    // 状态
    lastTabTime,
    userTyping,

    // 方法
    handleHome,
    handleClear,
    handleOptionClick,
    handleTab,
    handleBlur,
  };
}
