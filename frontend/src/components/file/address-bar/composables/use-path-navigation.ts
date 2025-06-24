import {
  ref,
  Ref,
  computed,
  onUnmounted,
  inject,
  provide,
  readonly,
  watchEffect,
} from 'vue';
import type { ComponentPublicInstance } from 'vue';
import useFileStore from '@/views/app/file/store/file-store';
import {
  addRootSlash,
  findCommonPrefix,
  getSearchPath,
  getSearchTerm,
} from '../utils';
import { DropdownOption } from './use-dropdown-navigation';
import { EmitFn } from '../types';

// 使用Symbol作为injection key，确保类型安全
const NAVIGATION_STATE_KEY = Symbol('navigationState');

interface NavigationState {
  isUserNavigation: boolean;
  userExpectedPath: string;
  timestamp: number;
}

interface NavigationStateProvider {
  getState: () => NavigationState;
  setState: (targetPath: string) => void;
  clearState: () => void;
  checkNavigationComplete: (path: string) => boolean;
}

// 创建导航状态提供者
function createNavigationStateProvider(): NavigationStateProvider {
  const state = ref<NavigationState>({
    isUserNavigation: false,
    userExpectedPath: '',
    timestamp: 0,
  });

  // 先定义clearState函数
  const clearState = () => {
    state.value = {
      isUserNavigation: false,
      userExpectedPath: '',
      timestamp: 0,
    };
  };

  // 使用computed来自动检查状态是否过期
  const isStateExpired = computed(() => {
    if (!state.value.isUserNavigation) return false;
    return Date.now() - state.value.timestamp > 5000; // 5秒过期
  });

  // 监听状态过期，自动清理
  const stopWatcher = watchEffect(() => {
    if (isStateExpired.value) {
      clearState();
    }
  });

  const setState = (targetPath: string) => {
    state.value = {
      isUserNavigation: true,
      userExpectedPath: targetPath,
      timestamp: Date.now(),
    };
  };

  // 获取当前有效状态
  const getState = (): NavigationState => {
    // 如果状态已过期，返回清空的状态
    if (isStateExpired.value) {
      return {
        isUserNavigation: false,
        userExpectedPath: '',
        timestamp: 0,
      };
    }
    return state.value;
  };

  const checkNavigationComplete = (path: string) => {
    const currentState = getState();
    if (
      currentState.isUserNavigation &&
      currentState.userExpectedPath === path
    ) {
      clearState();
      return true;
    }
    return false;
  };

  // 确保清理watcher
  onUnmounted(() => {
    stopWatcher();
    clearState();
  });

  return {
    getState,
    setState,
    clearState,
    checkNavigationComplete,
  };
}

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

// 辅助函数：安全地触发输入框失焦
function safeBlurInput(
  inputRef: Ref<ComponentPublicInstance | HTMLElement | null>
) {
  if (inputRef.value) {
    // 处理Vue组件实例
    if ('$el' in inputRef.value) {
      const element = inputRef.value.$el;
      if (element && typeof element.blur === 'function') {
        element.blur();
      }
    }
    // 处理DOM元素
    else if (typeof inputRef.value.blur === 'function') {
      inputRef.value.blur();
    }
  }
}

export default function usePathNavigation(
  value: Ref<string>,
  inputRef: Ref<ComponentPublicInstance | HTMLElement | null>,
  emit: EmitFn,
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>,
  isSearching: Ref<boolean>,
  triggerByTab: Ref<boolean>
) {
  const lastTabTime = ref(0);
  const userTyping = ref(true);

  // 尝试注入导航状态提供者，如果不存在则创建新的
  let navigationProvider =
    inject<NavigationStateProvider>(NAVIGATION_STATE_KEY);

  if (!navigationProvider) {
    navigationProvider = createNavigationStateProvider();
    provide(NAVIGATION_STATE_KEY, navigationProvider);
  }

  // 从provider获取状态
  const navigationState = computed(() => navigationProvider!.getState());
  const isUserNavigation = computed(
    () => navigationState.value.isUserNavigation
  );
  const userExpectedPath = computed(
    () => navigationState.value.userExpectedPath
  );

  // 设置用户导航状态
  const setUserNavigation = (targetPath: string) => {
    navigationProvider!.setState(targetPath);
  };

  // 清除用户导航状态
  const clearUserNavigation = () => {
    navigationProvider!.clearState();
  };

  // 检查是否应该忽略路径变化
  const shouldIgnorePathChange = (newPath: string) => {
    return isUserNavigation.value && userExpectedPath.value !== newPath;
  };

  // 检查用户导航是否完成
  const checkNavigationComplete = (newPath: string) => {
    return navigationProvider!.checkNavigationComplete(newPath);
  };

  /**
   * 导航到输入的路径
   */
  const handleGo = () => {
    const trimmedValue = value.value.trim();

    if (!trimmedValue) {
      emit('goto', '/');
      return;
    }

    const targetPath = addRootSlash(trimmedValue);
    setUserNavigation(targetPath);
    emit('goto', targetPath);
  };

  const handleHome = () => {
    value.value = '';
    emit('goto', '/');
  };

  const handleClear = () => {
    value.value = '';
    emit('clear');
    emit('search', {
      path: '/',
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

    // 不再自动触发导航，只更新输入框值
    // 用户需要再次按Enter来触发路径跳转
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

      const searchPath = getSearchPath(value.value);
      const searchTerm = getSearchTerm(value.value);

      emit('search', {
        path: searchPath,
        word: searchTerm,
      });
      return;
    }

    // 单个选项时只更新输入框值，不触发导航
    if (allOptions.value.length === 1) {
      const option = allOptions.value[0];
      const { basePath } = parseCurrentPath(value.value);
      const selectedValue = buildSelectedValue(option);

      value.value = basePath + selectedValue;

      // 清理下拉状态，但不触发导航
      allOptions.value = [];
      isSearching.value = false;
      triggerByTab.value = false;
      popupVisible.value = false;

      // 不再自动触发导航，只更新输入框值
      // 用户需要再次按Enter或点击Go按钮来触发路径跳转
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

  /**
   * 处理回车键
   */
  const handleKeyEnter = (
    e: KeyboardEvent | null,
    currentSelectedIndex?: number
  ) => {
    if (
      popupVisible.value &&
      currentSelectedIndex !== undefined &&
      currentSelectedIndex >= 0 &&
      currentSelectedIndex < allOptions.value.length
    ) {
      const selectedOption = allOptions.value[currentSelectedIndex];
      handleOptionClick(selectedOption);
      return;
    }

    handleGo();
  };

  const handleBlur = (path: string) => {
    if (value.value.trim() === '') {
      value.value = path;
    }
  };

  return {
    // 状态 - 使用readonly防止外部修改
    lastTabTime: readonly(lastTabTime),
    userTyping: readonly(userTyping),
    isUserNavigation,
    userExpectedPath,

    // 方法
    setUserNavigation,
    clearUserNavigation,
    shouldIgnorePathChange,
    checkNavigationComplete,
    handleGo,
    handleHome,
    handleClear,
    handleOptionClick,
    handleTab,
    handleKeyEnter,
    handleBlur,
  };
}
