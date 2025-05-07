<template>
  <div class="simplified-file-tree">
    <ul class="tree-list">
      <li
        v-for="item of rootItems"
        :key="item.path"
        class="tree-item"
        :class="{ selected: isSelected(item) }"
        @click="handleItemClick(item)"
        @dblclick="handleItemDoubleClick(item)"
      >
        <div class="tree-item-container">
          <!-- 用于对齐的元素，模拟折叠按钮区域 -->
          <div class="tree-item-toggle"></div>

          <!-- 内容区域 -->
          <div class="tree-item-content">
            <div class="tree-item-icon">
              <folder-icon />
            </div>
            <div class="tree-item-text">
              {{ item.name }}
            </div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { SimpleFileInfoEntity } from '@/entity/FileInfo';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { FileTreeItem } from './file-tree/type';

  const props = defineProps<{
    // The complete tree
    items: FileTreeItem[];
    // If showHidden is true, hidden files/folders will be displayed
    showHidden: boolean;
    // Currently selected item
    current: SimpleFileInfoEntity | null;
  }>();

  const emit = defineEmits(['itemSelect', 'itemDoubleClick']);

  // Only show root directories
  const rootItems = computed(() => {
    return props.items.filter(
      (item) => item.is_dir && (props.showHidden || !item.is_hidden)
    );
  });

  // Check if an item is selected by comparing paths
  const isSelected = (item: FileTreeItem) => {
    if (!props.current) return false;

    // 获取当前路径和项目路径
    const currentPath = props.current.path;
    const itemPath = item.path;

    // 确保是精确匹配或者子目录匹配，避免前缀相同但不相关的目录被选中
    // 例如：/lib 和 /lib.usr-is-merged 不应该互相影响
    if (currentPath === itemPath) {
      // 精确匹配 - 当前项就是选中的项
      return true;
    }

    if (currentPath.startsWith(itemPath + '/')) {
      // 子目录匹配 - 确保只有真正的父目录被高亮
      // 添加'/'确保是真正的子目录，避免前缀问题
      return true;
    }

    return false;
  };

  // Handle item click - emit selection event
  const handleItemClick = (item: FileTreeItem) => {
    emit('itemSelect', item);
  };

  // Handle double click - emit double click event
  const handleItemDoubleClick = (item: FileTreeItem) => {
    emit('itemDoubleClick', item);
  };
</script>

<style scoped>
  .simplified-file-tree {
    padding: 4px;
  }

  .tree-list {
    margin: 0;
    padding-left: 0;
    list-style: none;
  }

  .tree-item {
    position: relative;
  }

  .tree-item-container {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    line-height: 32px;
    border-radius: 4px;
    cursor: pointer;
  }

  /* 鼠标悬停高亮效果 */
  .tree-item-container:hover {
    background-color: var(--color-fill-1);
  }

  /* 选中项高亮效果 */
  .tree-item.selected .tree-item-container {
    background-color: var(--color-fill-2);
  }

  /* 选中项左侧指示条 */
  .tree-item.selected .tree-item-container::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
    content: '';
  }

  /* 展开/折叠按钮区域，保留用于对齐 */
  .tree-item-toggle {
    width: 16px;
    height: 100%;
  }

  /* 内容区域样式 */
  .tree-item-content {
    display: flex;
    flex: 1;
    align-items: center;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 5px 8px;
  }

  /* 图标样式 */
  .tree-item-icon {
    display: flex;
    align-items: center;
    height: 100%;
    margin-right: 8px;
    padding: 5px 0;
  }

  .tree-item-icon :deep(svg) {
    width: 16px;
    height: 16px;
  }

  /* 文本样式 */
  .tree-item-text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-size: 14px;
    line-height: 22px;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
</style>
