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
    // 完整文件树
    items: FileTreeItem[];
    // 是否显示隐藏文件/文件夹
    showHidden: boolean;
    // 当前选中的项目
    current: SimpleFileInfoEntity | null;
  }>();

  const emit = defineEmits(['itemSelect', 'itemDoubleClick']);

  // 只显示根目录，并按字母表排序
  const rootItems = computed(() => {
    return props.items
      .filter((item) => item.is_dir && (props.showHidden || !item.is_hidden))
      .sort((a, b) =>
        a.name.localeCompare(b.name, undefined, {
          numeric: true,
          caseFirst: 'lower',
        })
      );
  });

  // 检查项目是否被选中（通过比较路径）
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

  // 处理点击事件 - 发出选择事件
  const handleItemClick = (item: FileTreeItem) => {
    emit('itemSelect', item);
  };

  // 处理双击事件 - 发出双击事件
  const handleItemDoubleClick = (item: FileTreeItem) => {
    emit('itemDoubleClick', item);
  };
</script>

<style scoped>
  .simplified-file-tree {
    position: relative;
    width: 100%;
    padding: 8px 4px 8px 16px;
    margin: 0;
  }

  .tree-list {
    position: relative;
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .tree-item {
    position: relative;
    width: 100%;
    height: 32px;
    padding: 0 12px;
    margin-bottom: 8px;
    overflow: visible;
    cursor: pointer;
    list-style: none;
    background: transparent;
    border: none;
    border-radius: 4px;
    transition: background-color 0.2s ease;
  }

  /* 悬停状态 */
  .tree-item:hover {
    background-color: var(--color-fill-1);
  }

  /* 选中状态背景 */
  .tree-item.selected {
    background-color: var(--color-fill-2);
  }

  /* 选中状态的紫色指示条 */
  .tree-item.selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    content: '';
    background-color: rgb(var(--primary-6));
    border-radius: 2px;
  }

  /* 容器基础样式 */
  .tree-item-container {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    height: 32px;
    padding: 0;
    margin: 0;
    background: transparent;
    border: none;
    border-radius: 0;
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
    margin: 0;
    background: transparent;
  }

  /* 图标样式 */
  .tree-item-icon {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    margin-right: 8px;
  }

  .tree-item-icon :deep(svg) {
    width: 14px;
    height: 14px;
  }

  /* 文本样式 */
  .tree-item-text {
    flex: 1;
    min-width: 0;
    margin-left: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
    line-height: 22px;
    white-space: nowrap;
  }
</style>
