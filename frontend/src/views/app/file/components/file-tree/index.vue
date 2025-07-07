<template>
  <div class="file-tree">
    <list-render
      :items="props.items"
      :show-hidden="props.showHidden"
      :level="0"
    />
  </div>
</template>

<script lang="ts" setup>
  import { provide, ref, watch } from 'vue';
  import ListRender from './list-render.vue';
  import { FileTreeItem } from './type';

  /**
   * 组件属性定义
   */
  const props = defineProps<{
    items: FileTreeItem[]; // 文件树顶层项目列表
    showHidden?: boolean; // 是否显示隐藏文件
    selected?: FileTreeItem | null; // 当前选中的项目
    selectedChange: (item: FileTreeItem) => void; // 选中项变化回调
    openChange: (item: FileTreeItem, open: boolean) => void; // 展开状态变化回调
    doubleClickChange?: (item: FileTreeItem) => void; // 双击项目回调
  }>();

  // 创建响应式选中项引用
  const selectedItem = ref(props.selected);

  // 监听外部传入的selected变化
  watch(
    () => props.selected,
    (value) => {
      selectedItem.value = value;
    }
  );

  // 向子组件提供状态和回调函数
  provide('selected', selectedItem);
  provide('selectedChange', props.selectedChange);
  provide('openChange', props.openChange);
  provide('doubleClickChange', props.doubleClickChange);
</script>

<style scoped>
  .file-tree {
    padding: 8px 4px 8px 16px;
  }

  /* 鼠标悬停时显示层级连接线 */
  .file-tree:hover :deep(.tree-level-line) {
    background-color: var(--color-border-2);
  }
</style>
