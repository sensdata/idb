<template>
  <a-dropdown
    ref="dropdownRef"
    :popup-visible="computedPopupVisible"
    trigger="focus"
    auto-fit-popup-width
    prevent-focus
    :click-to-close="false"
    :popup-offset="4"
    @popup-visible-change="handlePopupVisibleChange"
  >
    <div class="address-bar-container">
      <div class="root-symbol" @click="handleHomeClick">
        <icon-home />
        <span class="root-slash">/</span>
      </div>
      <div class="input-wrapper">
        <a-input
          ref="inputRef"
          v-model="value"
          :placeholder="$t('components.file.addressBar.input.placeholder')"
          class="address-bar"
          allow-clear
          @clear="handleClear"
          @input="handleInputValueChange"
          @keydown.tab.prevent="handleTab"
          @keydown.up.prevent="handleKeyUp"
          @keydown.down.prevent="handleKeyDown"
          @keydown.enter.prevent="handleKeyEnter"
          @focus="handleInputFocus"
          @blur="handleInputBlur"
        >
          <template #suffix>
            <div
              class="go-button"
              :class="{ 'go-button--active': isInputFocused }"
              @mousedown.stop
              @click="handleGoButtonClick"
            >
              <icon-arrow-right />
            </div>
          </template>
        </a-input>
      </div>
    </div>
    <template #content>
      <dropdown-content
        ref="dropdownContentRef"
        :options="allOptions"
        :selected-index="currentSelectedIndex"
        :is-searching="isSearching"
        :is-loading="isLoading"
        @scroll="handleScrollEvent"
        @option-mouse-enter="handleOptionMouseEnter"
        @option-click="handleOptionClick"
      />
    </template>
  </a-dropdown>
</template>

<script lang="ts" setup>
  import { computed, ref, watch, onMounted, nextTick, onUnmounted } from 'vue';
  import { IconHome, IconArrowRight } from '@arco-design/web-vue/es/icon';
  import type { FileInfoEntity } from '@/entity/FileInfo';
  import { useLogger } from '@/composables/use-logger';
  import { useAddressBarStore } from './store/address-bar-store';

  import DropdownContent, { DropdownOption } from './drop-down-content.vue';
  import useDropdownNavigation from './composables/use-dropdown-navigation';
  import useAddressBarSearch from './composables/use-address-bar-search';
  import usePathNavigation from './composables/use-path-navigation';
  import { removeRootSlash, addRootSlash } from './utils';
  import type { EmitFn } from './types';

  // ==================== Props & Emits ====================
  interface Props {
    path: string;
    items?: FileInfoEntity[];
  }

  const props = defineProps<Props>();
  const emit = defineEmits<EmitFn>();

  // ==================== Template Refs ====================
  const inputRef = ref<HTMLElement | null>(null);
  const dropdownRef = ref();
  const dropdownContentRef = ref();

  // ==================== Store ====================
  const addressBarStore = useAddressBarStore();

  // ==================== Reactive State ====================
  const value = ref('');
  const allOptions = ref<DropdownOption[]>([]);
  const popupVisible = ref(false);
  const isInputFocused = ref(false);

  // ==================== Logger ====================
  const { logWarn, logDebug } = useLogger('AddressBar');

  // ==================== Helper Functions ====================
  const focusInputWithSelection = () => {
    nextTick(() => {
      if (inputRef.value) {
        try {
          const input = inputRef.value as HTMLInputElement;
          input.focus();
          const length = value.value.length;
          input.setSelectionRange(length, length);
        } catch (error) {
          logWarn('Failed to set focus with selection:', error);
        }
      }
    });
  };

  // ==================== Composables ====================
  const {
    currentSelectedIndex,
    handleKeyUp,
    handleKeyDown,
    setSelectedIndex,
    setHoverItem,
    resetSelection,
  } = useDropdownNavigation(allOptions, popupVisible);

  const {
    isLoading,
    isSearching,
    searchWord,
    triggerByTab,
    computedPopupVisible,
    handleInputValueChange: originalHandleInputValueChange,
    handleSearchComplete,
    handleScroll,
  } = useAddressBarSearch({
    value,
    emit,
    allOptions,
    popupVisible,
  });

  const {
    handleHome,
    handleClear,
    handleTab: originalHandleTab,
    handleBlur: originalHandleBlur,
  } = usePathNavigation(
    value,
    inputRef,
    emit,
    allOptions,
    popupVisible,
    isSearching,
    triggerByTab
  );

  // ==================== Computed Properties ====================
  const validOptions = computed<DropdownOption[]>(() => {
    return (props.items || []).map((item) => ({
      value: item.name,
      label: item.name,
      isDir: item.is_dir,
      displayValue: item.is_dir ? `${item.name}/` : item.name,
    }));
  });

  // ==================== Event Handlers ====================
  // 定义 handleGo 函数在其他函数之前，避免 no-use-before-define 错误
  const handleGo = () => {
    const trimmedValue = value.value.trim();
    if (!trimmedValue) {
      emit('goto', '/');
      return;
    }

    const targetPath = addRootSlash(trimmedValue);

    // 开始导航，store 会保持当前输入值
    addressBarStore.startNavigation(targetPath);

    emit('goto', targetPath);
  };

  const handleInputValueChange = () => {
    // 更新 store 中的用户输入值
    addressBarStore.setUserInput(value.value);
    originalHandleInputValueChange();
  };

  const handleTab = () => {
    originalHandleTab();
    // Maintain focus after tab completion
    focusInputWithSelection();
  };

  const handlePopupVisibleChange = (visible: boolean) => {
    popupVisible.value = visible;
  };

  const handleOptionMouseEnter = (item: DropdownOption, index: number) => {
    setSelectedIndex(index);

    if (item.isDir) {
      setHoverItem(item);
    } else {
      setHoverItem(null);
    }
  };

  const handleOptionClick = () => {
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      // 鼠标点击选项时，只更新输入框值，不触发导航
      const selectedOption = allOptions.value[currentSelectedIndex.value];
      const basePath = value.value.substring(
        0,
        value.value.lastIndexOf('/') + 1
      );
      const selectedValue = selectedOption.isDir
        ? `${selectedOption.value}/`
        : selectedOption.value;
      value.value = basePath + selectedValue;

      // 更新 store
      addressBarStore.setUserInput(value.value);

      // 清理下拉状态
      allOptions.value = [];
      isSearching.value = false;
      popupVisible.value = false;
      resetSelection();
    } else {
      handleGo();
    }
  };

  const handleKeyEnter = () => {
    logDebug('handleKeyEnter called:', {
      popupVisible: popupVisible.value,
      currentSelectedIndex: currentSelectedIndex.value,
      allOptionsLength: allOptions.value.length,
      inputValue: value.value,
      timestamp: new Date().toISOString(),
    });

    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      logDebug(
        'Enter: using dropdown option selection - only updating input value'
      );
      // 当有下拉选项被选中时，只更新输入框值，不触发导航
      const selectedOption = allOptions.value[currentSelectedIndex.value];
      const basePath = value.value.substring(
        0,
        value.value.lastIndexOf('/') + 1
      );
      const selectedValue = selectedOption.isDir
        ? `${selectedOption.value}/`
        : selectedOption.value;
      value.value = basePath + selectedValue;

      // 更新 store
      addressBarStore.setUserInput(value.value);

      // 清理下拉状态
      allOptions.value = [];
      isSearching.value = false;
      popupVisible.value = false;
      resetSelection();
    } else {
      logDebug('Enter: calling handleGo directly');
      handleGo();

      // 失焦输入框
      nextTick(() => {
        if (inputRef.value) {
          (inputRef.value as HTMLInputElement).blur();
        }
      });
    }
  };

  const handleGoButtonClick = () => {
    handleGo();
  };

  const handleBlur = (path: string) => {
    originalHandleBlur(path);
  };

  const handleInputFocus = () => {
    isInputFocused.value = true;
    if (!value.value) {
      value.value = addressBarStore.getDisplayValue(props.path);
      focusInputWithSelection();
    }
  };

  const handleInputBlur = () => {
    isInputFocused.value = false;
    handleBlur(removeRootSlash(props.path));
  };

  const handleHomeClick = () => {
    handleHome();
  };

  const handleScrollEvent = (event: Event) => {
    const target = event.target as HTMLElement;
    if (target) {
      handleScroll(target);
    }
  };

  // ==================== Watchers ====================
  watch(
    validOptions,
    (newOptions) => {
      logDebug('validOptions changed:', {
        newOptionsLength: newOptions.length,
        isSearching: isSearching.value,
        searchWord: searchWord.value,
        firstFewOptions: newOptions.slice(0, 3),
      });

      if (isSearching.value && searchWord.value) {
        const searchTerm = searchWord.value.toLowerCase();
        const filteredOptions = newOptions.filter((option) =>
          option.value.toLowerCase().startsWith(searchTerm)
        );
        allOptions.value = [...filteredOptions];

        logDebug('Search filtering applied:', {
          searchTerm,
          filteredCount: filteredOptions.length,
          originalCount: newOptions.length,
        });

        // 搜索完成后，重置搜索状态
        handleSearchComplete();
      } else {
        allOptions.value = [...newOptions];
      }

      if (allOptions.value.length > 0 && popupVisible.value) {
        setSelectedIndex(0);
      } else {
        resetSelection();
      }
    },
    { immediate: true }
  );

  watch(
    () => props.path,
    (newPath) => {
      logDebug('props.path changed:', {
        newPath,
        shouldIgnore: addressBarStore.shouldIgnorePathChange(newPath),
        isNavigating: addressBarStore.isUserNavigating,
        expectedPath: addressBarStore.expectedPath,
      });

      // 检查是否应该忽略路径变化
      if (addressBarStore.shouldIgnorePathChange(newPath)) {
        logDebug('Ignoring path change during navigation');
        return;
      }

      // 检查导航是否完成
      if (addressBarStore.completeNavigation(newPath)) {
        logDebug('Navigation completed successfully');
      }

      // 更新输入框显示值
      const displayValue = addressBarStore.getDisplayValue(newPath);
      if (value.value !== displayValue) {
        value.value = displayValue;
      }
    },
    { immediate: true }
  );

  watch(computedPopupVisible, (visible) => {
    if (visible && allOptions.value.length > 0) {
      setSelectedIndex(0);
    } else {
      resetSelection();
    }
  });

  // ==================== Lifecycle ====================
  onMounted(() => {
    logDebug('AddressBar mounted');

    // 初始化显示值
    if (!value.value.trim() && props.path) {
      value.value = addressBarStore.getDisplayValue(props.path);
    }
  });

  onUnmounted(() => {
    logDebug('AddressBar unmounted');
    // 组件卸载时不重置 store，让状态保持
  });
</script>

<style scoped>
  .address-bar-container {
    display: flex;
    align-items: stretch;
    width: 100%;
    background-color: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: var(--border-radius-small, 4px);
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .address-bar-container:focus-within {
    border-color: rgb(var(--primary-6));
  }

  .input-wrapper {
    position: relative;
    display: flex;
    flex: 1;
    min-width: 0;
  }

  .address-bar {
    width: 100%;
    border: none;
  }

  :deep(.arco-input-wrapper.address-bar) {
    padding-right: 0 !important;
    border: none;
  }

  .address-bar :deep(.arco-input-wrapper):focus-within {
    border-color: rgb(var(--primary-6));
  }

  .address-bar :deep(.arco-input) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .address-bar :deep(.arco-input-inner-wrapper) {
    padding-right: 0;
  }

  .address-bar :deep(.arco-input-clear-btn) {
    margin-right: var(--spacing-medium, 10px);
  }

  .root-symbol {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    min-width: 36px;
    padding: 0 var(--spacing-small, 8px);
    cursor: pointer;
    background-color: var(--color-fill-2);
    border-right: 1px solid var(--color-border-2);
    transition: background-color 0.2s ease;
  }

  .root-symbol:hover {
    background-color: var(--color-fill-3);
  }

  .root-slash {
    margin-left: var(--spacing-mini, 4px);
    font-weight: 600;
    color: var(--color-text-1);
    user-select: none;
  }

  .go-button {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 100%;
    min-height: 32px;
    margin-right: -1px;
    cursor: pointer;
    border-left: 1px solid var(--color-border-2);
    transition: all 0.2s ease;
  }

  .go-button--active {
    z-index: 10;
    color: white;
    background-color: rgb(var(--primary-6));
    border: none;
    border-radius: var(--border-radius-small, 2px);
    box-shadow: 0 2px 8px rgb(var(--primary-6) / 30%);
    transform: scale(1.02);
  }

  .go-button--active:hover {
    background-color: rgb(var(--primary-5));
  }

  .go-button--active:active {
    background-color: rgb(var(--primary-7));
    transform: scale(0.98);
  }

  /* Responsive Design */
  @media (width <= 768px) {
    .root-symbol {
      min-width: 28px;
      padding: 0 6px;
    }
    .root-slash {
      display: none;
    }
    .go-button {
      width: 32px;
    }
  }

  @media (width <= 576px) {
    .root-symbol {
      min-width: 24px;
      padding: 0 4px;
    }
    .go-button {
      width: 28px;
    }
  }
</style>
