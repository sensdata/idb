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
        <!-- 选中状态的紫色指示条 -->
        <div v-if="isSelected(item)" class="selection-indicator"></div>

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

  // 只显示根目录
  const rootItems = computed(() => {
    return props.items.filter(
      (item) => item.is_dir && (props.showHidden || !item.is_hidden)
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
    padding: 8px;
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
    padding: 0;
    margin: 0;
    overflow: visible;
    list-style: none;
    background: transparent;
    border: none;
  }

  /* 容器基础样式 */
  .tree-item-container {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    height: 32px;
    padding: 0 12px;
    margin: 0;
    cursor: pointer;
    border: none;
    border-radius: 0;
    transition: background-color 0.2s ease;
  }

  /* 悬停状态 - 移除container上的背景 */
  .tree-item:hover .tree-item-container {
    background-color: transparent;
  }

  /* 选中状态的紫色指示条 - 细长胶囊形状 */
  .selection-indicator {
    position: absolute;
    top: 4px;
    left: 4px;
    z-index: 10;
    display: block;
    width: 4px;
    height: 24px;
    background-color: #6241d4;
    border-radius: 2px;
  }

  /* 选中状态的背景 - 移除container上的背景，改为在content上设置 */
  .tree-item.selected .tree-item-container {
    background-color: transparent;
    border: none;
  }

  /* 强制覆盖任何可能的选中状态样式 */
  .tree-item.selected {
    background-color: transparent;
    border: none;
    border-left: none;
  }

  /* 内容区域样式 */
  .tree-item-content {
    display: flex;
    flex: 1;
    align-items: center;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 0 16px;
    margin: 0;
    border-radius: 4px;
    box-shadow: 0 1px 2px rgb(0 0 0 / 5%);
    transition: box-shadow 0.2s ease;
  }

  /* 悬停时的背景和阴影效果 */
  .tree-item:hover .tree-item-content {
    background-color: rgb(229 230 235 / 50%);
    box-shadow: 0 2px 8px rgb(0 0 0 / 10%);
  }

  /* 选中状态的背景和阴影效果 */
  .tree-item.selected .tree-item-content {
    background-color: #f2f3f5;
    box-shadow: 0 2px 12px rgb(98 65 212 / 15%);
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
    padding: 0;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
    line-height: 22px;
    color: #1d2129;
    white-space: nowrap;
  }

  /* 选中状态的文本颜色 */
  .tree-item.selected .tree-item-text {
    font-weight: 400;
    color: #1d2129;
  }

  /* 悬停时的文本颜色 */
  .tree-item:hover .tree-item-text {
    color: #1d2129;
  }

  /* 确保紫色指示条可见 - 使用伪元素作为备用方案 */
  .tree-item.selected::before {
    position: absolute;
    top: 4px;
    left: 4px;
    z-index: 15;
    display: block;
    width: 4px;
    height: 24px;
    content: '';
    background-color: #6241d4;
    border-radius: 2px;
  }
</style>
