<template>
  <div
    class="idb-tree-item"
    :class="{
      'is-selected': isSelected,
      'is-disabled': item.disabled,
      'is-loading': item.loading,
      [`level-${level}`]: true,
    }"
    :style="{ paddingLeft: `${level * 12}px` }"
    @click="handleClick"
    @dblclick="handleDoubleClick"
  >
    <!-- 展开/折叠按钮 -->
    <div v-if="showToggle && hasChildren" class="idb-tree-item__toggle">
      <span class="idb-tree-item__toggle-btn" @click.stop="handleToggle">
        <component :is="item.expanded ? collapseIcon : expandIcon" />
      </span>
    </div>
    <div v-else-if="showToggle" class="idb-tree-item__toggle">
      <!-- 占位符，保持对齐 -->
    </div>

    <!-- 内容区域 -->
    <div class="idb-tree-item__content">
      <!-- 图标 -->
      <div v-if="showIcon" class="idb-tree-item__icon">
        <component :is="itemIcon" />
      </div>

      <!-- 文本 -->
      <div class="idb-tree-item__text">
        {{ item.label }}
      </div>

      <!-- 加载中指示器 -->
      <div v-if="item.loading" class="idb-tree-item__loading">
        <a-spin :size="14" />
      </div>
    </div>
  </div>

  <!-- 子项列表 -->
  <div v-if="hasChildren && item.expanded" class="idb-tree-item__children">
    <tree-item
      v-for="child in visibleChildren"
      :key="child.id"
      :item="child"
      :level="level + 1"
      :selected="selected"
      :show-icon="showIcon"
      :show-toggle="showToggle"
      :show-hidden="showHidden"
      :default-icon="defaultIcon"
      :expand-icon="expandIcon"
      :collapse-icon="collapseIcon"
      @select="$emit('select', $event)"
      @double-click="$emit('doubleClick', $event)"
      @toggle="handleChildToggle"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';

  // 内联类型定义，避免跨文件导入问题
  interface TreeItem {
    /** 唯一标识 */
    id: string | number;
    /** 显示文本 */
    label: string;
    /** 图标组件或图标名 */
    icon?: string | any;
    /** 是否禁用 */
    disabled?: boolean;
    /** 是否加载中 */
    loading?: boolean;
    /** 是否展开（仅树模式） */
    expanded?: boolean;
    /** 子项列表（仅树模式） */
    children?: TreeItem[];
    /** 是否隐藏 */
    hidden?: boolean;
    /** 扩展数据 */
    [key: string]: any;
  }

  type SelectedValue = string | number | TreeItem | null;

  interface Props {
    item: TreeItem;
    level?: number;
    selected?: SelectedValue;
    showIcon?: boolean;
    showToggle?: boolean;
    showHidden?: boolean;
    defaultIcon?: string | any;
    expandIcon?: string | any;
    collapseIcon?: string | any;
  }

  interface Emits {
    (e: 'select', item: TreeItem): void;
    (e: 'doubleClick', item: TreeItem): void;
    (e: 'toggle', item: TreeItem, expanded: boolean): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    level: 0,
    showIcon: true,
    showToggle: true,
    showHidden: false,
  });

  const emit = defineEmits<Emits>();

  // 计算属性
  const isSelected = computed(() => {
    if (!props.selected) return false;

    if (typeof props.selected === 'object') {
      return props.selected.id === props.item.id;
    }

    return props.selected === props.item.id;
  });

  const hasChildren = computed(() => {
    return props.item.children && props.item.children.length > 0;
  });

  const visibleChildren = computed(() => {
    if (!props.item.children) return [];

    return props.showHidden
      ? props.item.children
      : props.item.children.filter((child) => !child.hidden);
  });

  const itemIcon = computed(() => {
    if (props.item.icon) {
      return props.item.icon;
    }
    return props.defaultIcon;
  });

  // 事件处理
  const handleClick = () => {
    if (props.item.disabled) return;
    emit('select', props.item);
  };

  const handleDoubleClick = () => {
    if (props.item.disabled) return;
    emit('doubleClick', props.item);
  };

  const handleToggle = () => {
    if (props.item.disabled || !hasChildren.value) return;

    const newExpanded = !props.item.expanded;
    // 通过事件通知父组件更新状态，不直接修改props
    emit('toggle', props.item, newExpanded);
  };

  const handleChildToggle = (item: TreeItem, expanded: boolean) => {
    emit('toggle', item, expanded);
  };
</script>

<style scoped>
  .idb-tree-item {
    position: relative;
    display: flex;
    align-items: center;
    height: 32px;
    margin-bottom: 8px;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.2s ease;
  }

  .idb-tree-item:hover {
    background-color: var(--color-fill-1);
  }

  .idb-tree-item.is-selected {
    background-color: var(--color-fill-2);
  }

  .idb-tree-item.is-selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    content: '';
    background-color: rgb(var(--primary-6));
    border-radius: 2px;
  }

  .idb-tree-item.is-disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  .idb-tree-item.is-loading {
    cursor: wait;
  }

  .idb-tree-item__toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    margin-right: 4px;
  }

  .idb-tree-item__toggle-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 2px;
    transition: background-color 0.2s ease;
  }

  .idb-tree-item__toggle-btn:hover {
    background-color: var(--color-fill-3);
  }

  .idb-tree-item__toggle-btn :deep(svg) {
    width: 12px;
    height: 12px;
  }

  .idb-tree-item__content {
    display: flex;
    flex: 1;
    align-items: center;
    min-width: 0;
    padding: 0 12px;
  }

  .idb-tree-item__icon {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    margin-right: 8px;
  }

  .idb-tree-item__icon :deep(svg) {
    width: 14px;
    height: 14px;
  }

  .idb-tree-item__text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
    line-height: 22px;
    white-space: nowrap;
  }

  .idb-tree-item__loading {
    flex-shrink: 0;
    margin-left: 8px;
  }

  .idb-tree-item__children {
    /* 子项容器，无需额外样式 */
  }

  /* 层级缩进样式 */
  .idb-tree-item.level-0 {
    padding-left: 8px;
  }

  .idb-tree-item.level-1 {
    padding-left: 20px;
  }

  .idb-tree-item.level-2 {
    padding-left: 32px;
  }

  .idb-tree-item.level-3 {
    padding-left: 44px;
  }
</style>
