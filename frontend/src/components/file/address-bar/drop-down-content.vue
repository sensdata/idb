<template>
  <div ref="contentRef" class="dropdown-content" @scroll="onScroll">
    <div v-if="isSearching" class="searching-indicator">
      <a-spin :size="20" />
      <span class="ml-2">{{ $t('common.loading') }}</span>
    </div>
    <div v-else-if="options.length === 0" class="no-results">
      {{
        $t('app.file.list.message.noResults') ||
        $t('common.noResults') ||
        '没有找到结果'
      }}
    </div>
    <a-doption
      v-for="(item, index) of options"
      v-show="!isSearching"
      :key="item.value"
      :ref="
        (el) => {
          if (el) optionRefs[index] = el;
        }
      "
      :value="item"
      :class="{ 'option-selected': index === selectedIndex }"
      @mouseenter="onOptionMouseEnter(item, index)"
      @click="onOptionClick(item)"
    >
      <div class="file-option">
        <span class="file-icon">
          <icon-folder v-if="item.isDir" />
          <icon-file v-else />
        </span>
        <span :class="{ 'dir-label': item.isDir }">{{ item.label }}</span>
      </div>
    </a-doption>
    <div v-if="isLoading" class="loading-more">
      <a-spin :size="20" />
      <span class="ml-2">{{ $t('common.loading') }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { IconFolder, IconFile } from '@arco-design/web-vue/es/icon';
  import type { ComponentPublicInstance } from 'vue';

  export interface DropdownOption {
    value: string;
    label: string;
    isDir?: boolean;
    displayValue?: string;
  }

  const props = defineProps<{
    options: DropdownOption[];
    selectedIndex: number;
    isSearching: boolean;
    isLoading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'scroll', event: Event): void;
    (e: 'optionMouseEnter', item: DropdownOption, index: number): void;
    (e: 'optionClick', item: DropdownOption): void;
  }>();

  const contentRef = ref<HTMLElement | null>(null);
  const optionRefs = ref<(ComponentPublicInstance | Element | null)[]>([]);

  // 当选项变化时重置 optionRefs 数组
  watch(
    () => props.options.length,
    () => {
      optionRefs.value = [];
    }
  );

  function onScroll(event: Event) {
    emit('scroll', event);
  }

  function onOptionMouseEnter(item: DropdownOption, index: number) {
    emit('optionMouseEnter', item, index);
  }

  function onOptionClick(item: DropdownOption) {
    emit('optionClick', item);
  }

  defineExpose({
    contentRef,
    optionRefs,
  });
</script>

<style scoped>
  .dropdown-content {
    max-height: 300px;
    overflow-y: auto;
  }

  .loading-more {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 8px 0;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .searching-indicator {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    color: var(--color-text-3);
  }

  .no-results {
    padding: 16px;
    color: var(--color-text-3);
    text-align: center;
  }

  .option-selected {
    background-color: var(--color-fill-2) !important;
  }

  .option-selected :deep(.arco-dropdown-option-content) {
    font-weight: 500;
  }

  .file-option {
    display: flex;
    align-items: center;
  }

  .file-icon {
    margin-right: 8px;
  }

  .dir-label {
    color: rgb(var(--primary-6));
    font-weight: 500;
  }
</style>
