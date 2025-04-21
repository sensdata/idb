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
      // emit('loading', true); // 已移除，父组件未监听此事件
      emit('goto', '/');
      return;
    }

    const targetPath = addRootSlash(v);

    // emit('loading', true); // 已移除，父组件未监听此事件
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

    // Clear navigation state before navigating
    if (item.isDir) {
      // Clear options to prevent stale entries
      allOptions.value = [];

      // Reset search state when entering a directory
      isSearching.value = false;
      triggerByTab.value = false;
    }

    // Navigate to the path
    emit('goto', targetPath);

    if (item.isDir) {
      // Request fresh directory contents
      emit('search', {
        path: targetPath,
        word: '',
      });
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

    // Check if path ends with a slash - this means we're inside a directory and should list contents
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
      // Immediately select the only option - this could be a directory
      const selectedOption = allOptions.value[0];

      // Clear options before navigating to prevent stale entries
      if (selectedOption.isDir) {
        allOptions.value = [];
      }

      handleOptionClick(selectedOption);

      // Reset search state when selecting with tab
      isSearching.value = false;
      triggerByTab.value = false;
      return;
    }

    // On double tab, just show all options
    if (isDoubleTab) {
      popupVisible.value = true;
      return;
    }

    // If we're in a directory (path ends with slash), just show dropdown
    if (endsWithSlash) {
      // When in a directory, request fresh directory contents
      allOptions.value = [];
      popupVisible.value = true;

      // Reset the search state when showing directory contents
      isSearching.value = false;
      triggerByTab.value = false;

      // Force refresh of directory contents
      const searchPath = getSearchPath(value.value);
      emit('search', {
        path: searchPath,
        word: '',
      });
      return;
    }

    // Find the longest common prefix for auto-completion
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

      // Only update if we found a longer match than what's already typed
      if (
        prefix &&
        prefix.length > currentTerm.length &&
        prefix.toLowerCase().startsWith(currentTerm.toLowerCase())
      ) {
        value.value = basePath + prefix;

        // Trigger search with the new prefix
        emit('search', {
          path: addRootSlash(basePath),
          word: prefix,
        });
      } else if (allOptions.value.length > 1) {
        // If no better completion found but multiple options exist, show dropdown
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
      // If no better completion found but multiple options exist, show dropdown
      popupVisible.value = true;
    }

    // Always show options if there are multiple matches
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
        // emit('loading', true); // 已移除，父组件未监听此事件
        emit('goto', '/');
        return;
      }

      const targetPath = addRootSlash(v);

      // emit('loading', true); // 已移除，父组件未监听此事件
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
