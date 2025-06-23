import { ref, Ref } from 'vue';
import {
  addRootSlash,
  findCommonPrefix,
  getSearchPath,
  getSearchTerm,
} from '../utils';
import { DropdownOption } from './use-dropdown-navigation';
import { EmitFn } from '../types';

export default function usePathNavigation(
  value: Ref<string>,
  inputRef: Ref<any>,
  emit: EmitFn,
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>,
  isSearching: Ref<boolean>,
  triggerByTab: Ref<boolean>
) {
  const lastTabTime = ref(0);
  const userTyping = ref(true);

  /**
   * 导航到输入的路径
   */
  function handleGo() {
    const v = value.value.trim();

    if (!v) {
      emit('goto', '/');
      return;
    }

    const targetPath = addRootSlash(v);

    emit('goto', targetPath);
  }

  function handleHome() {
    value.value = '';
    emit('goto', '/');
  }

  function handleClear() {
    value.value = '';
    emit('clear');
    emit('search', {
      path: '/',
      word: '',
    });
    popupVisible.value = false;
    isSearching.value = false;
  }

  function handleOptionClick(item: DropdownOption) {
    popupVisible.value = false;

    const currentPath = value.value;
    const lastSlashIndex = currentPath.lastIndexOf('/');
    const basePath =
      lastSlashIndex >= 0 ? currentPath.substring(0, lastSlashIndex + 1) : '';
    const selectedValue = item.isDir ? `${item.value}/` : item.value;

    value.value = basePath + selectedValue;

    const targetPath = addRootSlash(value.value);

    /**
     * 导航前清理状态
     */
    if (item.isDir) {
      allOptions.value = [];

      isSearching.value = false;
      triggerByTab.value = false;

      emit('goto', targetPath);

      emit('search', {
        path: targetPath,
        word: '',
      });
    } else {
      emit('goto', targetPath);
    }

    inputRef.value?.blur();
  }

  /**
   * 处理Tab键自动补全
   */
  function handleTab() {
    const currentTime = Date.now();
    const isDoubleTab = currentTime - lastTabTime.value < 300;
    lastTabTime.value = currentTime;

    /**
     * 路径以斜杠结尾表示在目录内，应列出内容
     */
    const endsWithSlash = value.value.endsWith('/');

    if (allOptions.value.length === 0) {
      triggerByTab.value = true;
      isSearching.value = true;

      const searchPath = getSearchPath(value.value);
      const searchTerm = getSearchTerm(value.value);

      emit('search', {
        path: searchPath,
        word: searchTerm,
      });
      return;
    }

    if (allOptions.value.length === 1) {
      /**
       * 选择唯一选项（可能是目录）
       */
      const selectedOption = allOptions.value[0];

      if (selectedOption.isDir) {
        allOptions.value = [];
      }

      handleOptionClick(selectedOption);

      isSearching.value = false;
      triggerByTab.value = false;
      return;
    }

    /**
     * 双击Tab显示所有选项
     */
    if (isDoubleTab) {
      popupVisible.value = true;
      return;
    }

    /**
     * 在目录中时显示下拉菜单
     */
    if (endsWithSlash) {
      allOptions.value = [];
      popupVisible.value = true;

      isSearching.value = false;
      triggerByTab.value = false;

      const searchPath = getSearchPath(value.value);
      emit('search', {
        path: searchPath,
        word: '',
      });
      return;
    }

    /**
     * 查找最长公共前缀进行自动补全
     */
    const prefix = findCommonPrefix(allOptions.value);
    if (!prefix) {
      popupVisible.value = true;
      return;
    }

    const currentPath = value.value;
    const lastSlashIndex = currentPath.lastIndexOf('/');

    if (lastSlashIndex >= 0) {
      const basePath = currentPath.substring(0, lastSlashIndex + 1);
      const currentTerm = currentPath.substring(lastSlashIndex + 1);

      /**
       * 仅在找到比已输入更长的匹配项时更新
       */
      if (
        prefix &&
        prefix.length > currentTerm.length &&
        prefix.toLowerCase().startsWith(currentTerm.toLowerCase())
      ) {
        value.value = basePath + prefix;

        emit('search', {
          path: addRootSlash(basePath),
          word: prefix,
        });
      } else if (allOptions.value.length > 1) {
        popupVisible.value = true;
      }
    } else if (
      prefix &&
      prefix.length > value.value.length &&
      prefix.toLowerCase().startsWith(value.value.toLowerCase())
    ) {
      value.value = prefix;

      emit('search', {
        path: '/',
        word: prefix,
      });
    } else if (allOptions.value.length > 1) {
      popupVisible.value = true;
    }

    /**
     * 当有多个匹配项时始终显示选项
     */
    if (allOptions.value.length > 1) {
      popupVisible.value = true;
    }
  }

  function handleKeyEnter(
    e: KeyboardEvent | null,
    currentSelectedIndex: Ref<number>
  ) {
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      if (e) {
        e.preventDefault();
      }
      handleOptionClick(allOptions.value[currentSelectedIndex.value]);
    }
  }

  /**
   * 处理输入框失焦时导航到输入的路径
   */
  function handleBlur(path: string) {
    if (popupVisible.value && allOptions.value.length > 0) {
      return;
    }

    setTimeout(() => {
      if (path === value.value) {
        return;
      }

      const v = value.value.trim();

      if (!v) {
        emit('goto', '/');
        return;
      }

      const targetPath = addRootSlash(v);

      emit('goto', targetPath);
    }, 200);
  }

  return {
    lastTabTime,
    userTyping,
    handleGo,
    handleHome,
    handleClear,
    handleOptionClick,
    handleTab,
    handleKeyEnter,
    handleBlur,
  };
}
