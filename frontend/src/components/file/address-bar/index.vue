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
      <div class="root-symbol" @click="handleHome">
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
          @blur="() => handleBlur(removeRootSlash(props.path))"
        >
          <template #suffix>
            <div
              class="after"
              :class="{ 'after-highlight': pathChanged }"
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
        @scroll="handleScroll"
        @option-mouse-enter="handleOptionMouseEnter"
        @option-click="handleOptionClick"
      />
    </template>
  </a-dropdown>
</template>

<script lang="ts" setup>
  import { computed, ref, watch, onMounted, nextTick } from 'vue';
  import { IconHome, IconArrowRight } from '@arco-design/web-vue/es/icon';
  import { FileInfoEntity } from '@/entity/FileInfo';

  import DropdownContent, { DropdownOption } from './drop-down-content.vue';
  import useDropdownNavigation from './composables/use-dropdown-navigation';
  import useAddressBarSearch from './composables/use-address-bar-search';
  import usePathNavigation from './composables/use-path-navigation';
  import { removeRootSlash, addRootSlash } from './utils';
  import { EmitFn } from './types';

  const props = defineProps<{
    path: string;
    items?: FileInfoEntity[];
  }>();

  const emit = defineEmits<EmitFn>();

  const inputRef = ref();
  const dropdownRef = ref();
  const dropdownContentRef = ref();
  const value = ref('');
  const allOptions = ref<any[]>([]);
  const popupVisible = ref(false);
  const pathChanged = ref(false);

  const {
    currentSelectedIndex,
    hoverItem,
    preloadTimeoutId,
    handleKeyUp,
    handleKeyDown,
  } = useDropdownNavigation(allOptions, popupVisible, dropdownContentRef);

  const {
    isLoading,
    isSearching,
    searchWord,
    triggerByTab,
    computedPopupVisible,
    handleInputValueChange: originalHandleInputValueChange,
    handleScroll,
  } = useAddressBarSearch(
    value,
    emit,
    allOptions,
    popupVisible,
    dropdownContentRef
  );

  function handleInputValueChange() {
    pathChanged.value = addRootSlash(value.value) !== props.path;
    originalHandleInputValueChange();
  }

  const {
    userTyping,
    handleGo: originalHandleGo,
    handleHome,
    handleClear,
    handleTab: originalHandleTab,
    handleKeyEnter: originalHandleKeyEnter,
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

  function handleTab() {
    originalHandleTab();
    pathChanged.value = false;
    nextTick(() => {
      if (inputRef.value) {
        inputRef.value.focus();
      }
    });
  }

  function handlePopupVisibleChange(visible: boolean) {
    popupVisible.value = visible;
  }

  function handleOptionMouseEnter(item: DropdownOption, index: number) {
    currentSelectedIndex.value = index;
    inputRef.value = item.value;

    if (item.isDir) {
      hoverItem.value = item;

      if (preloadTimeoutId.value) {
        clearTimeout(preloadTimeoutId.value);
      }
    } else {
      hoverItem.value = null;
      if (preloadTimeoutId.value) {
        clearTimeout(preloadTimeoutId.value);
        preloadTimeoutId.value = null;
      }
    }
  }

  function handleOptionClick() {
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      originalHandleKeyEnter(null, currentSelectedIndex);
    } else {
      originalHandleGo();
    }
    pathChanged.value = false;
    nextTick(() => {
      if (inputRef.value) {
        inputRef.value.focus();
      }
    });
  }

  function handleKeyEnter(e: KeyboardEvent) {
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      originalHandleKeyEnter(e, currentSelectedIndex);
    } else {
      originalHandleGo();
    }
    pathChanged.value = false;
    nextTick(() => {
      if (inputRef.value) {
        inputRef.value.focus();
      }
    });
  }

  function handleGoButtonClick() {
    originalHandleGo();
    pathChanged.value = false;
    nextTick(() => {
      if (inputRef.value) {
        inputRef.value.focus();
      }
    });
  }

  function handleBlur(path: string) {
    originalHandleBlur(path);
  }

  function handleInputFocus() {
    if (!value.value) {
      value.value = removeRootSlash(props.path);
      nextTick(() => {
        if (inputRef.value) {
          inputRef.value.focus();
          inputRef.value.select(value.value.length, value.value.length);
        }
      });
    }
  }

  const validOptions = computed(() => {
    return (props.items || []).map((item) => ({
      value: item.name,
      label: item.name,
      isDir: item.is_dir,
      displayValue: item.is_dir ? `${item.name}/` : item.name,
    }));
  });

  watch(validOptions, (newOptions) => {
    if (isSearching.value) {
      const searchTerm = searchWord.value.toLowerCase();

      if (searchTerm) {
        const filteredOptions = newOptions.filter((option) =>
          option.value.toLowerCase().startsWith(searchTerm)
        );

        allOptions.value = [...filteredOptions];
      } else {
        allOptions.value = [...newOptions];
      }

      currentSelectedIndex.value = allOptions.value.length > 0 ? 0 : -1;
    }
    isSearching.value = false;
  });

  watch(
    () => props.path,
    (newPath) => {
      if (addRootSlash(value.value) !== newPath) {
        userTyping.value = false;
        value.value = removeRootSlash(newPath);
      }
    },
    { immediate: true }
  );

  watch(computedPopupVisible, (visible) => {
    if (visible && allOptions.value.length > 0) {
      currentSelectedIndex.value = 0;
    } else {
      currentSelectedIndex.value = -1;
    }
  });

  onMounted(() => {
    nextTick(() => {
      if (dropdownContentRef.value?.contentRef) {
        const contentElement = dropdownContentRef.value.contentRef;
        contentElement.addEventListener('scroll', handleScroll);
      }
    });
  });
</script>

<style scoped>
  .address-bar-container {
    display: flex;
    align-items: stretch;
    width: 100%;
    background-color: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
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

  .address-bar :deep(.arco-input-wrapper) {
    border: none;
  }

  .address-bar :deep(.arco-input) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .address-bar :deep(.arco-input-inner-wrapper) {
    padding-right: 0;
  }

  .address-bar :deep(.arco-input-prefix) {
    padding-right: 4px;
  }

  .address-bar :deep(.arco-input-suffix) {
    padding: 0;
  }

  .address-bar :deep(.arco-input-clear-btn) {
    margin-right: 10px;
  }

  .root-symbol {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    min-width: 36px;
    padding: 0 8px;
    cursor: pointer;
    background-color: var(--color-fill-2);
    border-right: 1px solid var(--color-border-2);
  }

  .root-slash {
    margin-left: 4px;
    font-weight: bold;
    color: var(--color-text-1);
    user-select: none;
  }

  .after {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 100%;
    min-height: 32px;
    cursor: pointer;
    border-left: 1px solid var(--color-border-2);
    transition: all 0.2s ease;
  }

  .after-highlight {
    z-index: 10;
    color: white;
    background-color: #6e41e2;
    border: none;
    border-radius: 2px;
    box-shadow: 0 0 8px rgb(110 65 226 / 50%);
    transform: scale(1.05);
    animation: purple-pulse 1.2s infinite;
  }

  .after-highlight:hover {
    background-color: #7a50e7;
  }

  .after-highlight:active {
    background-color: #5f38c3;
  }

  .after-highlight :deep(svg) {
    font-weight: bold;
    color: white;
  }

  @keyframes purple-pulse {
    0% {
      background-color: #6e41e2;
      box-shadow: 0 0 8px rgb(110 65 226 / 50%);
    }
    50% {
      background-color: #8257ef;
      box-shadow: 0 0 15px rgb(130 87 239 / 70%);
    }
    100% {
      background-color: #6e41e2;
      box-shadow: 0 0 8px rgb(110 65 226 / 50%);
    }
  }

  /* 平板和移动设备适配 */
  @media screen and (width <= 768px) {
    .root-symbol {
      min-width: 28px;
      padding: 0 6px;
    }
    .root-slash {
      display: none;
    }
    .after {
      width: 32px;
    }
  }

  @media screen and (width <= 576px) {
    .root-symbol {
      min-width: 24px;
      padding: 0 4px;
    }
    .after {
      width: 28px;
    }
  }
</style>
