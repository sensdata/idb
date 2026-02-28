<template>
  <div class="breadcrumb-address-bar">
    <!-- 面包屑导航容器 -->
    <div
      class="breadcrumb-container"
      :class="{
        'has-content': hasContent,
        'is-focused': isFocused,
      }"
    >
      <div class="address-rows">
        <!-- 当前目录行 -->
        <div class="current-row">
          <span class="zone-badge zone-badge-current">{{
            $t('components.file.addressBar.current.label')
          }}</span>
          <div ref="breadcrumbRef" class="custom-breadcrumb">
            <!-- Home 项 -->
            <div
              class="breadcrumb-item home-item"
              @click="handlePathClick('/')"
            >
              <icon-home class="home-icon" />
              <span class="breadcrumb-text">{{ rootDisplayName }}</span>
            </div>

            <!-- 路径段 -->
            <template
              v-for="(segment, index) in pathSegments"
              :key="segment.path"
            >
              <icon-right class="breadcrumb-separator" />
              <div
                class="breadcrumb-item"
                :class="
                  index === pathSegments.length - 1
                    ? 'current-item'
                    : 'path-item'
                "
                :title="segment.path"
                @click="
                  index !== pathSegments.length - 1
                    ? handleSegmentClick(index)
                    : undefined
                "
              >
                <span class="breadcrumb-text">{{ segment.name }}</span>
              </div>
            </template>
          </div>
        </div>

        <!-- 输入目录行 -->
        <div class="input-row">
          <span class="zone-badge zone-badge-input">{{
            $t('components.file.addressBar.input.label')
          }}</span>
          <div class="path-input-container">
            <a-input
              ref="pathInputRef"
              v-model="inputValue"
              :placeholder="$t('components.file.addressBar.input.placeholder')"
              class="path-input"
              size="small"
              @keydown.enter="handleInputEnter"
              @keydown.escape="handleInputEscape"
              @keydown.up="handleKeyUp"
              @keydown.down="handleKeyDown"
              @keydown.tab="handleTab"
              @blur="handleInputBlur"
              @focus="handleInputFocus"
              @input="handleInputValueChange"
            />
          </div>
          <div
            class="clear-button"
            :class="clearButtonClasses"
            @click="handleClearButtonClick"
          >
            <icon-close />
          </div>
          <div
            class="go-button"
            :class="goButtonClasses"
            @click="handleGoButtonClick"
          >
            <icon-arrow-right />
          </div>
        </div>
      </div>
    </div>

    <!-- 下拉搜索建议 -->
    <a-dropdown
      ref="dropdownRef"
      :popup-visible="computedPopupVisible"
      auto-fit-popup-width
      prevent-focus
      :click-to-close="false"
      :popup-offset="4"
      @popup-visible-change="handlePopupVisibleChange"
    >
      <div></div>
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
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue';
  import {
    IconHome,
    IconRight,
    IconArrowRight,
    IconClose,
  } from '@arco-design/web-vue/es/icon';
  import type { FileInfoEntity } from '@/entity/FileInfo';
  import { useLogger } from '@/composables/use-logger';

  import DropdownContent, { DropdownOption } from './drop-down-content.vue';
  import useDropdownNavigation from './composables/use-dropdown-navigation';
  import useAddressBarSearch from './composables/use-address-bar-search';
  import usePathNavigation from './composables/use-path-navigation';
  import { resolveLinuxPathInput } from './utils';
  import type { EmitFn } from './types';

  // ==================== Props & Emits ====================
  interface Props {
    path: string;
    items?: FileInfoEntity[];
  }

  const props = defineProps<Props>();
  const emit = defineEmits<EmitFn>();

  // ==================== Template Refs ====================
  const pathInputRef = ref<HTMLElement | null>(null);
  const breadcrumbRef = ref<HTMLElement | null>(null);
  const dropdownRef = ref();
  const dropdownContentRef = ref();

  // ==================== Reactive State ====================
  const inputValue = ref('');
  const allOptions = ref<DropdownOption[]>([]);
  const popupVisible = ref(false);
  const isFocused = ref(false);
  const shouldRefocusAfterGoto = ref(false);
  const refocusFallbackTimer = ref<number | null>(null);

  // ==================== Logger ====================
  const { logDebug } = useLogger('AddressBar');

  // ==================== Computed Properties ====================
  const rootDisplayName = computed(() => {
    return '/';
  });

  // 解析路径为面包屑段
  const pathSegments = computed(() => {
    if (!props.path || props.path === '/') {
      return [];
    }

    const cleanPath = props.path.replace(/^\/+|\/+$/g, '');
    if (!cleanPath) {
      return [];
    }

    const segments = cleanPath.split('/').filter(Boolean);
    return segments.map((name, index) => ({
      name,
      path: '/' + segments.slice(0, index + 1).join('/'),
    }));
  });

  // 检查是否有输入内容
  const hasContent = computed(() => inputValue.value.trim().length > 0);

  // 调试：Go按钮的class状态
  const goButtonClasses = computed(() => {
    const classes = {
      'go-button--disabled': !inputValue.value.trim(),
      'go-button--active': isFocused.value || hasContent.value,
    };
    logDebug(
      'Go button classes:',
      classes,
      'inputValue:',
      inputValue.value,
      'isFocused:',
      isFocused.value,
      'hasContent:',
      hasContent.value
    );
    return classes;
  });

  const clearButtonClasses = computed(() => ({
    'clear-button--disabled': !inputValue.value.trim(),
  }));

  // ==================== Helper Functions ====================
  const resolveTargetPath = (rawInput: string) => {
    return resolveLinuxPathInput(rawInput, props.path);
  };

  const focusPathInput = () => {
    const inputHost = pathInputRef.value as any;
    const inputEl: HTMLInputElement | null =
      inputHost?.$el?.querySelector('input') ||
      inputHost?.querySelector?.('input') ||
      null;
    inputEl?.focus();
  };

  const requestRefocusAfterGoto = () => {
    if (isFocused.value) {
      shouldRefocusAfterGoto.value = true;
      if (refocusFallbackTimer.value) {
        window.clearTimeout(refocusFallbackTimer.value);
      }
      refocusFallbackTimer.value = window.setTimeout(() => {
        shouldRefocusAfterGoto.value = false;
        refocusFallbackTimer.value = null;
      }, 1500);
    }
  };

  // 更新输入框显示值
  const updateInputValue = () => {
    // 默认状态下输入框应该为空，用于输入子目录或文件名
    inputValue.value = '';
  };

  // ==================== Composables ====================
  const {
    currentSelectedIndex,
    setSelectedIndex,
    setHoverItem,
    resetSelection,
    handleKeyUp: dropdownKeyUp,
    handleKeyDown: dropdownKeyDown,
  } = useDropdownNavigation(allOptions, popupVisible);

  const {
    isLoading,
    isSearching,
    searchWord,
    triggerByTab,
    computedPopupVisible,
    handleInputValueChange: originalHandleInputValueChange,
    handleScroll,
  } = useAddressBarSearch({
    value: inputValue,
    currentPath: computed(() => props.path), // 传递当前路径
    emit,
    allOptions,
    popupVisible,
  });

  const pathNavigation = usePathNavigation(
    inputValue,
    pathInputRef,
    emit,
    allOptions,
    popupVisible,
    isSearching,
    triggerByTab,
    computed(() => props.path) // 传递当前路径
  );

  // ==================== Event Handlers ====================

  // 处理输入回车
  const handleInputEnter = () => {
    // Enter 优先按当前输入执行导航（更接近 shell 行为）
    const trimmedValue = inputValue.value.trim();
    if (!trimmedValue) {
      // 如果输入为空，不做任何操作
      return;
    }

    const targetPath = resolveTargetPath(trimmedValue);
    requestRefocusAfterGoto();
    emit('goto', targetPath);

    // 清理下拉状态
    allOptions.value = [];
    popupVisible.value = false;
    isSearching.value = false;
    resetSelection();

    // 清空输入框
    inputValue.value = '';
  };

  // 处理路径点击
  const handlePathClick = (path: string) => {
    emit('goto', path);
  };

  // 处理路径段点击
  const handleSegmentClick = (segmentIndex: number) => {
    const targetPath = pathSegments.value[segmentIndex].path;
    emit('goto', targetPath);
  };

  // 处理 ESC 键
  const handleInputEscape = () => {
    // ESC键清空输入框
    inputValue.value = '';
    popupVisible.value = false;
  };

  const handleClearButtonClick = () => {
    if (!inputValue.value.trim()) return;
    handleInputEscape();
    focusPathInput();
  };

  // 处理Go按钮点击
  const handleGoButtonClick = () => {
    const trimmedValue = inputValue.value.trim();
    if (!trimmedValue) {
      return;
    }

    const targetPath = resolveTargetPath(trimmedValue);
    requestRefocusAfterGoto();
    emit('goto', targetPath);

    // 清空输入框
    inputValue.value = '';

    // 隐藏下拉框
    popupVisible.value = false;
    allOptions.value = [];
    isSearching.value = false;
    resetSelection();
  };

  // 处理输入框失焦
  const handleInputBlur = () => {
    if (shouldRefocusAfterGoto.value) {
      return;
    }
    // 设置失焦状态
    isFocused.value = false;
    logDebug(
      'Input blur - isFocused:',
      isFocused.value,
      'hasContent:',
      hasContent.value
    );

    // 延迟隐藏，让用户有时间点击下拉选项
    setTimeout(() => {
      if (!popupVisible.value) {
        // 若未修改路径，失焦后回到空输入态
        if (inputValue.value.trim() === props.path) {
          inputValue.value = '';
        }
      }
    }, 200);
  };

  // 处理输入框聚焦
  const handleInputFocus = () => {
    // 聚焦时切换为完整路径编辑模式，支持直接修改任意路径段
    inputValue.value = props.path || '/';

    // 设置聚焦状态
    isFocused.value = true;
    logDebug(
      'Input focus - isFocused:',
      isFocused.value,
      'hasContent:',
      hasContent.value
    );

    // 自动全选，便于用户快速覆盖编辑整条路径
    nextTick(() => {
      const inputHost = pathInputRef.value as any;
      const inputEl: HTMLInputElement | null =
        inputHost?.$el?.querySelector('input') ||
        inputHost?.querySelector?.('input') ||
        null;
      inputEl?.select();
    });
  };

  // 处理上下键导航
  const handleKeyUp = (event: KeyboardEvent) => {
    event.preventDefault();
    dropdownKeyUp();
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    event.preventDefault();
    dropdownKeyDown();
  };

  // 处理Tab键
  const handleTab = (event: KeyboardEvent) => {
    event.preventDefault();
    pathNavigation.handleTab();
  };

  // 处理下拉选项输入变化
  const handleInputValueChange = () => {
    originalHandleInputValueChange();
  };

  // 处理下拉框可见性变化
  const handlePopupVisibleChange = (visible: boolean) => {
    popupVisible.value = visible;
  };

  // 处理下拉选项悬停
  const handleOptionMouseEnter = (item: DropdownOption, index: number) => {
    setSelectedIndex(index);
    if (item.isDir) {
      setHoverItem(item);
    } else {
      setHoverItem(null);
    }
  };

  // 处理下拉选项点击
  const handleOptionClick = (item: DropdownOption) => {
    const basePath = inputValue.value.substring(
      0,
      inputValue.value.lastIndexOf('/') + 1
    );
    const selectedValue = item.isDir ? `${item.value}/` : item.value;
    inputValue.value = basePath + selectedValue;

    // 清理下拉状态
    allOptions.value = [];
    isSearching.value = false;
    popupVisible.value = false;
    resetSelection();
  };

  // 处理滚动事件
  const handleScrollEvent = (event: Event) => {
    const target = event.target as HTMLElement;
    if (target) {
      handleScroll(target);
    }
  };

  // ==================== Watchers ====================
  watch(
    () => props.path,
    (newPath) => {
      logDebug('props.path changed:', newPath);

      // 路径变化时，保持输入框为空，不自动填充
      // 用户应该能够在新路径下输入子目录或文件名
      inputValue.value = '';
      nextTick(() => {
        if (breadcrumbRef.value) {
          breadcrumbRef.value.scrollLeft = breadcrumbRef.value.scrollWidth;
        }
        if (shouldRefocusAfterGoto.value) {
          focusPathInput();
          shouldRefocusAfterGoto.value = false;
          if (refocusFallbackTimer.value) {
            window.clearTimeout(refocusFallbackTimer.value);
            refocusFallbackTimer.value = null;
          }
        }
      });
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

  // 监听输入值变化触发搜索
  watch(inputValue, (newValue) => {
    logDebug(
      'inputValue changed:',
      newValue,
      'hasContent:',
      hasContent.value,
      'isFocused:',
      isFocused.value
    );
    if (newValue.trim()) {
      handleInputValueChange();
    } else {
      // 输入为空时，清空选项和隐藏下拉框
      allOptions.value = [];
      popupVisible.value = false;
      isSearching.value = false;
    }
  });

  // 监听搜索结果变化，转换为下拉选项并重置loading状态
  watch(
    () => props.items,
    (newItems) => {
      logDebug(
        'props.items changed, isSearching:',
        isSearching.value,
        'items length:',
        newItems?.length
      );

      // 将搜索结果转换为下拉选项，客户端过滤确保只显示以搜索词开头的结果
      if (newItems && newItems.length > 0) {
        const searchTerm = searchWord.value.trim().toLowerCase();
        const filteredItems = newItems.filter(
          (item) =>
            searchTerm === '' || item.name.toLowerCase().startsWith(searchTerm)
        );

        allOptions.value = filteredItems.map((item) => ({
          value: item.name,
          label: item.name,
          isDir: item.is_dir,
          displayValue: item.is_dir ? `${item.name}/` : item.name,
        }));

        // 保持下拉框可见
        if (inputValue.value.trim()) {
          popupVisible.value = true;
        }

        logDebug(
          'Filtered and converted items to options:',
          allOptions.value.length,
          'from',
          newItems.length
        );
      } else if (newItems) {
        // 如果返回空数组，清空选项但保持下拉框可见以显示"无结果"
        allOptions.value = [];
        if (inputValue.value.trim()) {
          popupVisible.value = true;
        }
      }

      // 当搜索结果更新时，重置loading状态
      if (newItems && isSearching.value) {
        logDebug('Resetting isSearching to false due to search results update');
        isSearching.value = false;
      }
    },
    { deep: true }
  );

  // ==================== Lifecycle ====================
  onMounted(() => {
    logDebug('BreadcrumbAddressBar mounted');
    logDebug(
      'Initial state - isFocused:',
      isFocused.value,
      'hasContent:',
      hasContent.value,
      'inputValue:',
      inputValue.value
    );
    updateInputValue();
  });

  onUnmounted(() => {
    if (refocusFallbackTimer.value) {
      window.clearTimeout(refocusFallbackTimer.value);
      refocusFallbackTimer.value = null;
    }
    logDebug('BreadcrumbAddressBar unmounted');
  });
</script>

<style scoped lang="less">
  @import url('@/assets/style/breakpoint.less');

  // Variables
  @container-height: 2.286rem;
  @border-width: 0.071rem;
  @border-radius: var(--border-radius-small, 0.286rem);
  @transition-duration: 0.2s;
  @padding-sm: 0.143rem;
  @padding-md: 0.286rem;
  @padding-lg: 0.286rem;
  @padding-xl: 0.857rem;
  @font-size-base: 1rem;
  @font-size-sm: 0.857rem;
  @font-size-xs: 0.714rem;

  // Mixins
  .flex-center() {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .transition(@property: all, @duration: @transition-duration, @timing: ease) {
    transition: @property @duration @timing;
  }

  .button-state(@bg-color, @border-color, @text-color: currentcolor) {
    color: @text-color;
    background: @bg-color;
    border-color: @border-color;
  }

  .breadcrumb-address-bar {
    position: relative;
    width: 100%;
  }

  .breadcrumb-container {
    position: relative;
    padding: @padding-sm;
    overflow: hidden;
    background-color: var(--color-bg-2);
    border: @border-width solid var(--color-border-2);
    border-radius: @border-radius;
    .transition();
    &:hover {
      border-color: var(--color-border-3);
    }
    &.is-focused,
    &.has-content {
      background: var(--color-fill-2);
      border-color: var(--color-border-3);
    }
  }

  .address-rows {
    display: flex;
    flex-direction: column;
    gap: @padding-sm;
    width: 100%;
  }

  .current-row,
  .input-row {
    display: flex;
    gap: @padding-sm;
    align-items: center;
    min-width: 0;
    padding: 0 @padding-sm;
    border-radius: @padding-sm;
  }

  .current-row {
    height: @container-height;
    background: var(--color-fill-1);
    border: @border-width solid var(--color-border-2);
  }

  .input-row {
    height: @container-height;
    background: var(--color-fill-1);
    border: @border-width solid var(--color-border-2);
  }

  // 自定义面包屑容器
  .custom-breadcrumb {
    display: flex;
    flex: 1 1 auto;
    align-items: center;
    min-width: 0;
    height: 100%;
    overflow: auto hidden;
    scrollbar-width: none;
  }

  .custom-breadcrumb::-webkit-scrollbar {
    display: none;
  }

  // 面包屑项通用样式
  .breadcrumb-item {
    .flex-center();

    padding: 0 @padding-lg;
    font-family: Roboto, -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: @font-size-base;
    line-height: 1.571rem;
    color: var(--color-text-2);
    cursor: pointer;
    border-radius: @padding-sm;
    .transition();
  }

  // Home 项样式 - 始终灰色背景
  .home-item {
    height: 1.857rem;
    padding: 0 @padding-lg;
    background-color: var(--color-fill-2);
    &:hover {
      background-color: var(--color-fill-3);
    }
  }

  // 路径项样式
  .path-item {
    &:hover {
      background-color: var(--color-fill-2);
    }
  }

  // 当前目录项样式 - 不可点击，无悬停效果
  .current-item {
    color: var(--color-text-1);
    cursor: default;
    &:hover {
      background-color: transparent;
    }
  }

  // 面包屑分隔符
  .breadcrumb-separator {
    flex-shrink: 0;
    margin: 0 @padding-md;
    font-size: @font-size-sm;
    color: var(--color-text-3);
  }

  // 面包屑文本
  .breadcrumb-text {
    white-space: nowrap;
    user-select: none;
  }

  // Home 图标
  .home-icon {
    margin-right: @padding-md;
    font-size: @font-size-base;
  }

  // 输入容器样式 - 完全透明，无边框
  .path-input-container {
    display: flex;
    flex: 1;
    align-items: center;
    min-width: 8.571rem;
    background: transparent;
    border: none;

    // 针对 arco-input-wrapper 移除所有默认样式 - 合并所有状态
    :deep(.arco-input-wrapper) {
      &,
      &:hover,
      &:focus,
      &:focus-within,
      &:active,
      &.arco-input-focus,
      &.arco-input-wrapper-focused,
      &.arco-input-wrapper-focus {
        padding: 0 !important;
        outline: none !important;
        background: transparent !important;
        border: none !important;
        box-shadow: none !important;
      }
    }
  }

  .zone-badge {
    flex-shrink: 0;
    padding: 0 0.429rem;
    font-size: @font-size-xs;
    line-height: 1.429rem;
    border-radius: 0.286rem;
  }

  .zone-badge-current {
    color: var(--color-text-2);
    background: var(--color-fill-2);
  }

  .zone-badge-input {
    color: rgb(var(--success-7));
    background: rgb(var(--success-1));
  }

  // 输入框样式
  .breadcrumb-container .path-input {
    flex: 1;
    min-width: 0;
    font-family: Roboto, -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: @font-size-base;
    border-radius: 0;
  }

  // 合并输入框所有状态的样式
  .path-input {
    :deep(.arco-input) {
      &,
      &:hover,
      &:focus {
        // 预留清除按钮空间，避免文字和右侧X重叠
        padding: 0 2rem 0 @padding-lg;
        font-size: @font-size-base;
        line-height: 1.571rem;
        outline: none !important;
        background: transparent !important;
        border: none !important;
        box-shadow: none !important;
      }

      color: var(--color-text-3);
      &:focus {
        color: var(--color-text-1);
      }
      &::placeholder {
        color: var(--color-text-3);
      }
    }
    :deep(.arco-input-wrapper) {
      position: relative;
      padding: 0 !important;
      background: transparent !important;
      border: none !important;
      box-shadow: none !important;
    }
  }

  .clear-button {
    flex-shrink: 0;
    .flex-center();

    width: 1.857rem;
    height: 1.857rem;
    color: var(--color-text-2);
    cursor: pointer;
    background: var(--color-bg-1);
    border: @border-width solid var(--color-border-2);
    border-radius: @padding-sm;
    .transition();
    &:hover:not(.clear-button--disabled) {
      background: var(--color-fill-2);
      border-color: var(--color-border-3);
    }
    &:active:not(.clear-button--disabled) {
      background: var(--color-fill-1);
      border-color: var(--color-border-3);
    }
    &.clear-button--disabled {
      color: var(--color-text-4);
      cursor: not-allowed;
      background: var(--color-fill-1);
      border-color: var(--color-border-2);
    }
    :deep(.arco-icon) {
      font-size: @font-size-base;
    }
  }

  // Go按钮样式 - 增加高度
  .go-button {
    flex-shrink: 0;
    .flex-center();

    width: 1.857rem;
    height: 1.857rem;
    color: var(--color-text-2);
    cursor: pointer;
    background: var(--color-bg-1);
    border: @border-width solid var(--color-border-2);
    border-radius: @padding-sm;
    .transition();
    &--disabled {
      color: var(--color-text-4);
      cursor: not-allowed;
      background: var(--color-fill-1);
      border-color: var(--color-border-2);
    }

    // 默认悬停和激活状态 - 仅在没有激活类时应用
    &:hover:not(.go-button--disabled, .go-button--active) {
      background: var(--color-fill-2);
      border-color: var(--color-border-3);
    }
    &:active:not(.go-button--disabled, .go-button--active) {
      background: var(--color-fill-1);
      border-color: var(--color-border-3);
    }
    :deep(.arco-icon) {
      font-size: @font-size-base;
      color: currentcolor;
    }
  }

  // Go按钮激活状态 - 灰色主题，使用更高权重选择器减少 !important
  .breadcrumb-container .go-button--active {
    color: #fff;
    cursor: pointer;
    background: var(--color-text-3);
    border: @border-width solid var(--color-text-3);
    &:hover {
      background: var(--color-text-2);
      border-color: var(--color-text-2);
    }
    &:active {
      background: var(--color-text-4);
      border-color: var(--color-text-4);
    }
    :deep(.arco-icon) {
      color: #fff;
    }
  }

  // 响应式设计
  @media (max-width: @screen-md) {
    .home-item {
      padding: 0 0.429rem;
    }
    .home-icon {
      font-size: @font-size-sm;
    }
    .path-input-container {
      flex: 1;
      min-width: 7.143rem;
    }
  }

  @media (max-width: @screen-sm) {
    .breadcrumb-text {
      max-width: 5.714rem;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .path-input-container {
      flex: 1;
      min-width: 5.714rem;
    }
    .zone-badge {
      display: none;
    }
  }

  // 深色主题适配
  @media (prefers-color-scheme: dark) {
    .breadcrumb-container {
      background-color: var(--color-bg-3);
      border-color: var(--color-border-3);
    }
  }
</style>
