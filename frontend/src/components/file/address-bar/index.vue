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
      <!-- 自定义面包屑组件 -->
      <div class="custom-breadcrumb">
        <!-- Home 项 -->
        <div class="breadcrumb-item home-item" @click="handlePathClick('/')">
          <icon-home class="home-icon" />
          <span class="breadcrumb-text">{{ rootDisplayName }}</span>
        </div>

        <!-- 路径段 - 第一个段前不显示分隔符 -->
        <template v-for="(segment, index) in pathSegments" :key="segment.path">
          <!-- 分隔符：只有非第一个路径段才显示 -->
          <icon-right v-if="index > 0" class="breadcrumb-separator" />

          <!-- 路径段项 -->
          <div
            class="breadcrumb-item"
            :class="
              index === pathSegments.length - 1 ? 'current-item' : 'path-item'
            "
            @click="
              index !== pathSegments.length - 1
                ? handleSegmentClick(index)
                : undefined
            "
          >
            <span class="breadcrumb-text">{{ segment.name }}</span>
          </div>
        </template>

        <!-- 在最后一个路径段后面添加分隔符 -->
        <icon-right
          v-if="pathSegments.length > 0"
          class="breadcrumb-separator"
        />
      </div>

      <!-- 路径输入框 - 无边框，透明背景 -->
      <div class="path-input-container">
        <a-input
          ref="pathInputRef"
          v-model="inputValue"
          placeholder="输入目录或文件名..."
          class="path-input"
          size="small"
          allow-clear
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

      <!-- Go按钮 - 相对于breadcrumb-container定位，高度匹配整个容器 -->
      <div
        class="go-button"
        :class="goButtonClasses"
        @click="handleGoButtonClick"
      >
        <icon-arrow-right />
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
  import { computed, ref, watch, onMounted, onUnmounted } from 'vue';
  import {
    IconHome,
    IconRight,
    IconArrowRight,
  } from '@arco-design/web-vue/es/icon';
  import type { FileInfoEntity } from '@/entity/FileInfo';
  import { useLogger } from '@/composables/use-logger';

  import DropdownContent, { DropdownOption } from './drop-down-content.vue';
  import useDropdownNavigation from './composables/use-dropdown-navigation';
  import useAddressBarSearch from './composables/use-address-bar-search';
  import usePathNavigation from './composables/use-path-navigation';
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
  const dropdownRef = ref();
  const dropdownContentRef = ref();

  // ==================== Reactive State ====================
  const inputValue = ref('');
  const allOptions = ref<DropdownOption[]>([]);
  const popupVisible = ref(false);
  const isFocused = ref(false);

  // ==================== Logger ====================
  const { logDebug } = useLogger('AddressBar');

  // ==================== Computed Properties ====================
  const rootDisplayName = computed(() => {
    // 可以根据当前主机或项目名称来显示，这里先用固定值
    return '';
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

  // ==================== Helper Functions ====================

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
    // 如果有下拉框选项且有选中项，使用选中的选项
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      const selectedOption = allOptions.value[currentSelectedIndex.value];
      const selectedValue = selectedOption.isDir
        ? `${selectedOption.value}/`
        : selectedOption.value;

      // 构建完整路径：当前路径 + 选中的选项
      const currentPath = props.path === '/' ? '' : props.path;
      const targetPath = `${currentPath}/${selectedValue}`.replace(/\/+/g, '/');
      emit('goto', targetPath);

      // 清理下拉状态
      allOptions.value = [];
      popupVisible.value = false;
      isSearching.value = false;
      resetSelection();
      inputValue.value = '';
      return;
    }

    // 如果没有选中项，使用输入的文本
    const trimmedValue = inputValue.value.trim();
    if (!trimmedValue) {
      // 如果输入为空，不做任何操作
      return;
    }

    // 构建完整路径：当前路径 + 输入的子路径
    const currentPath = props.path === '/' ? '' : props.path;
    const targetPath = `${currentPath}/${trimmedValue}`.replace(/\/+/g, '/');
    emit('goto', targetPath);

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

  // 处理Go按钮点击
  const handleGoButtonClick = () => {
    const trimmedValue = inputValue.value.trim();
    if (!trimmedValue) {
      return;
    }

    // 构建完整路径：当前路径 + 输入的子路径
    const currentPath = props.path === '/' ? '' : props.path;
    const targetPath = `${currentPath}/${trimmedValue}`.replace(/\/+/g, '/');
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
        // 失焦时保持输入内容，不自动清空
        // inputValue.value 保持用户输入的内容
      }
    }, 200);
  };

  // 处理输入框聚焦
  const handleInputFocus = () => {
    // 设置聚焦状态
    isFocused.value = true;
    logDebug(
      'Input focus - isFocused:',
      isFocused.value,
      'hasContent:',
      hasContent.value
    );

    // 聚焦时保持当前输入内容，不做修改
    // 让用户继续输入子目录或文件名
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
  const handleOptionClick = () => {
    if (
      popupVisible.value &&
      currentSelectedIndex.value >= 0 &&
      currentSelectedIndex.value < allOptions.value.length
    ) {
      const selectedOption = allOptions.value[currentSelectedIndex.value];
      const basePath = inputValue.value.substring(
        0,
        inputValue.value.lastIndexOf('/') + 1
      );
      const selectedValue = selectedOption.isDir
        ? `${selectedOption.value}/`
        : selectedOption.value;
      inputValue.value = basePath + selectedValue;

      // 清理下拉状态
      allOptions.value = [];
      isSearching.value = false;
      popupVisible.value = false;
      resetSelection();
    }
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
        const searchTerm = inputValue.value.trim().toLowerCase();
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
    logDebug('BreadcrumbAddressBar unmounted');
  });
</script>

<style scoped>
  .breadcrumb-address-bar {
    position: relative;
    width: 100%;
  }

  .breadcrumb-container {
    position: relative;
    display: flex;
    align-items: center;
    height: 32px;
    padding: 0;
    overflow: hidden;
    background-color: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: var(--border-radius-small, 4px);
    transition: all 0.2s ease;
  }

  /* 悬停状态 */
  .breadcrumb-container:hover {
    border-color: var(--color-border-3);
  }

  /* 聚焦状态 - 紫色主题 */
  .breadcrumb-container.is-focused {
    background: var(--color-primary-light-5);
    border-color: var(--color-primary-light-3);
  }

  /* 有内容时的状态 - 紫色主题 */
  .breadcrumb-container.has-content {
    background: var(--color-primary-light-5);
    border-color: var(--color-primary-light-3);
  }

  /* 自定义面包屑容器 */
  .custom-breadcrumb {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    min-width: 0;
    height: 100%;
  }

  /* 面包屑项通用样式 */
  .breadcrumb-item {
    display: flex;
    align-items: center;
    padding: 2px 8px;
    font-family: Roboto, -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: 14px;
    line-height: 22px;
    color: rgb(var(--primary-6)); /* #6241D4 */
    cursor: pointer;
    border-radius: 2px;
    transition: all 0.2s ease;
  }

  /* Home 项样式 - 始终灰色背景 */
  .home-item {
    align-self: stretch;
    height: 32px;
    padding: 0 8px 0 12px;
    background-color: var(--color-fill-2); /* #E5E6EB */
  }

  .home-item:hover {
    background-color: var(--color-fill-3); /* 悬停时稍深 */
  }

  /* 路径项样式 */
  .path-item:hover {
    background-color: var(--color-fill-2); /* #F2F3F5 */
  }

  /* 当前目录项样式 - 不可点击，无悬停效果 */
  .current-item {
    color: var(--color-text-1); /* 使用深色文本颜色表示当前位置 */
    cursor: default;
  }

  .current-item:hover {
    background-color: transparent; /* 无悬停背景 */
  }

  /* 面包屑分隔符 */
  .breadcrumb-separator {
    flex-shrink: 0;
    margin: 0 4px;
    font-size: 12px;
    color: var(--color-text-3); /* 更淡的颜色 */
  }

  /* 面包屑文本 */
  .breadcrumb-text {
    white-space: nowrap;
    user-select: none;
  }

  /* Home 图标 */
  .home-icon {
    margin-right: 4px;
    font-size: 14px;
  }

  /* 输入容器样式 - 完全透明，无边框 */
  .path-input-container {
    display: flex;
    flex: 1;
    align-items: center;
    min-width: 120px;
    background: transparent;
    border: none;
  }

  /* 针对 arco-input-wrapper 移除所有默认样式 - 合并所有状态 */
  .path-input-container :deep(.arco-input-wrapper),
  .path-input-container :deep(.arco-input-wrapper:hover),
  .path-input-container :deep(.arco-input-wrapper:focus),
  .path-input-container :deep(.arco-input-wrapper:focus-within),
  .path-input-container :deep(.arco-input-wrapper:active),
  .path-input-container :deep(.arco-input-wrapper.arco-input-focus),
  .path-input-container :deep(.arco-input-wrapper.arco-input-wrapper-focused),
  .path-input-container :deep(.arco-input-wrapper.arco-input-wrapper-focus) {
    padding: 0 !important;
    outline: none !important;
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
  }

  /* 输入框样式 */
  .breadcrumb-container .path-input {
    flex: 1;
    min-width: 0;
    padding-right: 65px; /* 增加右侧padding为Go按钮和清除按钮留出空间 */
    font-family: Roboto, -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: 14px;
    border-radius: 0;
  }

  /* 合并输入框所有状态的样式 */
  .path-input :deep(.arco-input),
  .path-input :deep(.arco-input:hover),
  .path-input :deep(.arco-input:focus) {
    padding: 2px 65px 2px 8px; /* 增加右侧padding */
    font-size: 14px;
    line-height: 22px;
    outline: none !important;
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
  }

  .path-input :deep(.arco-input) {
    color: var(--color-text-3);
  }

  .path-input :deep(.arco-input:focus) {
    color: var(--color-text-1);
  }

  .path-input :deep(.arco-input::placeholder) {
    color: var(--color-text-3);
  }

  .path-input :deep(.arco-input-wrapper) {
    padding: 0 !important;
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
  }

  /* 清除按钮样式调整 */
  .path-input :deep(.arco-input-clear-btn) {
    position: absolute;
    top: 50%;
    right: 35px; /* 定位在Go按钮左边 */
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    color: var(--color-text-3);
    background: transparent;
    border-radius: 2px;
    transform: translateY(-50%);
  }

  .path-input :deep(.arco-input-clear-btn:hover) {
    color: var(--color-text-2);
    background: var(--color-fill-2);
  }

  /* Go按钮样式 - 增加高度 */
  .go-button {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    color: var(--color-text-2);
    cursor: pointer;
    background: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 2px;
    transition: all 0.2s ease;
  }

  .go-button--disabled {
    color: var(--color-text-4);
    cursor: not-allowed;
    background: var(--color-fill-1);
    border-color: var(--color-border-2);
  }

  /* Go按钮激活状态 - 紫色主题，使用更高权重选择器减少 !important */
  .breadcrumb-container .go-button--active {
    color: white;
    cursor: pointer;
    background: rgb(var(--primary-6));
    border: 1.5px solid rgb(var(--primary-6));
    box-shadow: 0 0 12px rgba(var(--primary-6), 0.4);
    animation: button-glow 2s ease-in-out infinite alternate;
  }

  .breadcrumb-container .go-button--active:hover {
    background: rgb(var(--primary-5));
    border-color: rgb(var(--primary-5));
    box-shadow: 0 0 16px rgba(var(--primary-5), 0.6);
    animation: none;
  }

  .breadcrumb-container .go-button--active:active {
    background: rgb(var(--primary-7));
    border-color: rgb(var(--primary-7));
    box-shadow: 0 0 8px rgba(var(--primary-7), 0.8);
    animation: none;
  }

  /* 呼吸动画效果 */
  @keyframes button-glow {
    0% {
      box-shadow: 0 0 8px rgba(var(--primary-6), 0.3);
    }
    100% {
      box-shadow: 0 0 16px rgba(var(--primary-6), 0.6);
    }
  }

  /* 默认悬停和激活状态 - 仅在没有激活类时应用 */
  .go-button:hover:not(.go-button--disabled, .go-button--active) {
    background: var(--color-fill-2);
    border-color: var(--color-border-3);
  }

  .go-button:active:not(.go-button--disabled, .go-button--active) {
    background: var(--color-fill-1);
    border-color: var(--color-border-3);
  }

  .go-button :deep(.arco-icon) {
    font-size: 14px;
    color: currentcolor;
  }

  /* 响应式设计 */
  @media (width <= 768px) {
    .breadcrumb-container {
      height: 32px;
      padding: 0;
    }
    .home-item {
      padding: 0 6px 0 8px;
    }
    .home-icon-overlay {
      left: 0;
    }
    :deep(.arco-breadcrumb-item) {
      font-size: 13px;
    }
    :deep(.arco-breadcrumb-item-link) {
      padding: 2px 6px;
    }
    .home-icon {
      font-size: 12px;
    }
    .home-text {
      font-size: 13px;
    }
    :deep(.arco-breadcrumb-item-separator) {
      margin: 0 2px;
      font-size: 10px;
    }
    .path-input-container {
      flex: 1;
      min-width: 100px;
    }
    .go-button {
      top: 2px;
      right: 2px;
      bottom: 2px;
      width: 28px;
    }
    .path-input {
      padding-right: 55px; /* 调整移动端padding */
    }
    .path-input :deep(.arco-input) {
      padding: 2px 55px 2px 8px; /* 调整移动端padding */
    }
    .path-input :deep(.arco-input-clear-btn) {
      right: 30px; /* 移动端清除按钮位置 */
    }
  }

  @media (width <= 576px) {
    .breadcrumb-container {
      height: 32px;
      padding: 0;
    }
    .home-item {
      padding: 0 6px 0 6px;
    }
    .home-icon-overlay {
      left: 0; /* 调整小屏幕左边距 */
    }
    .home-text {
      font-size: 12px;
    }
    .breadcrumb-text {
      max-width: 80px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .path-input-container {
      flex: 1;
      min-width: 80px;
    }
    .go-button {
      top: 1px;
      right: 1px;
      bottom: 1px;
      width: 24px;
    }
    .go-button :deep(.arco-icon) {
      font-size: 12px;
    }
    .path-input {
      padding-right: 50px; /* 调整小屏幕padding */
    }
    .path-input :deep(.arco-input) {
      padding: 2px 50px 2px 6px; /* 调整小屏幕padding */
    }
    .path-input :deep(.arco-input-clear-btn) {
      right: 26px; /* 小屏幕清除按钮位置 */
    }
  }

  /* 深色主题适配 */
  @media (prefers-color-scheme: dark) {
    .breadcrumb-container {
      background-color: var(--color-bg-3);
      border-color: var(--color-border-3);
    }
  }
</style>
