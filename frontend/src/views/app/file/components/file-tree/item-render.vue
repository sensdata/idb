<template>
  <!-- 文件/文件夹项目容器 -->
  <div
    class="tree-item-container"
    :class="{
      'selected': selected?.path === item.path,
      ['level-' + level]: true,
      'is-directory': item.is_dir,
      'is-file': !item.is_dir,
    }"
    :style="{ paddingLeft: level * 8 + 'px' }"
    @click="handleClick"
  >
    <!-- 展开/折叠控制按钮区域 -->
    <div class="tree-item-toggle">
      <span v-if="item.is_dir" @click.stop="handleToggle">
        <component :is="item.open ? DownIcon : RightIcon" />
      </span>
    </div>
    <!-- 文件/文件夹内容区域 -->
    <div class="tree-item-content">
      <div class="tree-item-icon">
        <icon-render :item="item" />
      </div>
      <div class="tree-item-text truncate">
        {{ item.name }}
      </div>
    </div>
  </div>
  <!-- 加载中提示区域 -->
  <div
    v-if="item.loading"
    class="tree-item-loading"
    :style="{ paddingLeft: level * 8 + 'px' }"
  >
    <a-spin :size="14" />
    <span>{{ $t('common.loading') }}</span>
  </div>
  <!-- 子项列表渲染区域 -->
  <list-render
    v-if="showChildren"
    :items="item.items || []"
    :show-hidden="showHidden"
    :level="level + 1"
  />
</template>

<script lang="ts" setup>
  import { inject, Ref, computed } from 'vue';
  import DownIcon from '@/assets/icons/down.svg';
  import RightIcon from '@/assets/icons/direction-right.svg';
  import ListRender from './list-render.vue';
  import IconRender from './icon-render';
  import { FileTreeItem } from './type';

  const props = defineProps<{
    item: FileTreeItem; // 文件或文件夹项目
    showHidden?: boolean; // 是否显示隐藏文件
    level: number; // 层级深度
  }>();

  // 简化访问常用属性
  const { item } = props;

  // 从父组件中注入的状态和方法
  const selected = inject<Ref<FileTreeItem | undefined | null>>('selected')!;
  const selectedChange = inject<(item: FileTreeItem) => void>('selectedChange');
  const openChange =
    inject<(item: FileTreeItem, open: boolean) => void>('openChange');

  // 计算是否显示子项列表
  const showChildren = computed(
    () =>
      item.is_dir &&
      item.open &&
      Array.isArray(item.items) &&
      item.items.length > 0
  );

  /**
   * 处理文件夹展开/折叠按钮点击
   */
  function handleToggle() {
    openChange?.(item, !item.open);
  }

  /**
   * 处理项目点击事件
   */
  function handleClick() {
    // 对于文件夹，触发展开
    if (item.is_dir) {
      openChange?.(item, true);
    }

    // 更新选中项（无论是文件还是文件夹）
    selectedChange?.(item);
  }
</script>

<style scoped>
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
  .tree-item-container.selected {
    background-color: var(--color-fill-2);
  }

  /* 选中项左侧指示条 */
  .tree-item-container.selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
    content: '';
  }

  /* 展开/折叠按钮区域 */
  .tree-item-toggle {
    width: 12px;
    height: 100%;
  }

  .tree-item-toggle span {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 100%;
  }

  .level-0 .tree-item-toggle span {
    border-radius: 4px 0 0 4px;
  }

  .tree-item-toggle span:hover {
    background-color: var(--color-fill-3);
  }

  .tree-item-toggle svg {
    width: 12px;
    height: 12px;
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
    padding: 5px 0;
  }

  .tree-item-icon svg {
    width: 14px;
    height: 14px;
  }

  /* 文本样式 */
  .tree-item-text {
    flex: 1;
    min-width: 0;
    margin-left: 8px;
    font-size: 14px;
    line-height: 22px;
  }

  /* 加载中状态样式 */
  .tree-item-loading {
    display: flex;
    place-items: center flex-start;
    height: 32px;
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 32px;
  }

  .tree-item-loading :deep(.arco-spin) {
    margin-right: 6px;
    margin-left: 16px;
  }

  .tree-item-loading :deep(.arco-spin-icon) {
    color: var(--color-text-3);
  }
</style>
