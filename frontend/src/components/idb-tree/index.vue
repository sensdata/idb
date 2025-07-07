<template>
  <div class="idb-tree" :class="[`idb-tree--${mode}`]">
    <!-- 空状态 -->
    <div v-if="visibleItems.length === 0" class="idb-tree__empty">
      <span class="idb-tree__empty-text">{{ defaultEmptyText }}</span>
      <span
        v-if="showCreate"
        class="idb-tree__empty-create"
        @click="handleCreate"
      >
        {{ defaultCreateText }}
      </span>
    </div>

    <!-- 树形列表 -->
    <div v-else class="idb-tree__list">
      <tree-item
        v-for="item in visibleItems"
        :key="item.id"
        :item="item"
        :level="0"
        :selected="internalSelected"
        :show-icon="showIcon"
        :show-toggle="mode === 'tree' && showToggle"
        :show-hidden="showHidden"
        :default-icon="defaultIcon"
        :expand-icon="expandIcon"
        :collapse-icon="collapseIcon"
        @select="handleSelect"
        @double-click="handleDoubleClick"
        @toggle="handleToggle"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import DownIcon from '@/assets/icons/down.svg';
  import RightIcon from '@/assets/icons/direction-right.svg';
  import TreeItem from './tree-item.vue';
  import { TreeItem as TreeItemType, TreeProps, SelectedValue } from './types';

  type Props = TreeProps;

  interface Emits {
    (e: 'update:selected', value: SelectedValue): void;
    (e: 'select', item: TreeItemType): void;
    (e: 'doubleClick', item: TreeItemType): void;
    (e: 'toggle', item: TreeItemType, expanded: boolean): void;
    (e: 'create'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    mode: 'list',
    items: () => [],
    multiple: false,
    showIcon: true,
    showToggle: true,
    showHidden: false,
    emptyText: '',
    createText: '',
    showCreate: false,
    defaultIcon: FolderIcon,
    expandIcon: RightIcon,
    collapseIcon: DownIcon,
  });

  const emit = defineEmits<Emits>();

  // 国际化
  const { t } = useI18n();

  // 响应式数据
  const internalSelected = ref<SelectedValue>(props.selected || null);
  const internalItems = ref<TreeItemType[]>([]);

  // 深拷贝函数，避免修改原始props
  const deepCloneItems = (items: TreeItemType[]): TreeItemType[] => {
    return items.map((item) => ({
      ...item,
      children: item.children ? deepCloneItems(item.children) : undefined,
    }));
  };

  // 计算默认翻译文本
  const defaultEmptyText = computed(
    () => props.emptyText || t('tree.emptyText')
  );

  const defaultCreateText = computed(
    () => props.createText || t('tree.createText')
  );

  // 监听外部选中值变化
  watch(
    () => props.selected,
    (value) => {
      internalSelected.value = value || null;
    },
    { immediate: true }
  );

  // 监听 items 变化，创建内部副本
  watch(
    () => props.items,
    (newItems) => {
      internalItems.value = deepCloneItems(newItems);
    },
    { immediate: true, deep: true }
  );

  // 计算属性
  const visibleItems = computed(() => {
    return props.showHidden
      ? internalItems.value
      : internalItems.value.filter((item) => !item.hidden);
  });

  // 事件处理
  const handleSelect = (item: TreeItemType) => {
    internalSelected.value = item;
    emit('update:selected', item);
    emit('select', item);
  };

  const handleDoubleClick = (item: TreeItemType) => {
    emit('doubleClick', item);
  };

  const handleToggle = (item: TreeItemType, expanded: boolean) => {
    // 更新内部状态
    const updateItemExpanded = (items: TreeItemType[]): boolean => {
      for (const internalItem of items) {
        if (internalItem.id === item.id) {
          internalItem.expanded = expanded;
          return true;
        }
        if (
          internalItem.children &&
          updateItemExpanded(internalItem.children)
        ) {
          return true;
        }
      }
      return false;
    };

    updateItemExpanded(internalItems.value);
    emit('toggle', item, expanded);
  };

  const handleCreate = () => {
    emit('create');
  };

  // 暴露给父组件的方法
  const refresh = () => {
    // 可以在这里添加刷新逻辑
    // console.log('IdbTree refresh');
  };

  const selectItem = (id: string | number) => {
    const findItem = (items: TreeItemType[]): TreeItemType | null => {
      for (const item of items) {
        if (item.id === id) {
          return item;
        }
        if (item.children) {
          const found = findItem(item.children);
          if (found) return found;
        }
      }
      return null;
    };

    const item = findItem(internalItems.value);
    if (item) {
      handleSelect(item);
    }
  };

  const expandItem = (id: string | number) => {
    const findAndExpand = (items: TreeItemType[]): boolean => {
      for (const item of items) {
        if (item.id === id) {
          item.expanded = true;
          return true;
        }
        if (item.children && findAndExpand(item.children)) {
          return true;
        }
      }
      return false;
    };

    findAndExpand(internalItems.value);
  };

  const collapseItem = (id: string | number) => {
    const findAndCollapse = (items: TreeItemType[]): boolean => {
      for (const item of items) {
        if (item.id === id) {
          item.expanded = false;
          return true;
        }
        if (item.children && findAndCollapse(item.children)) {
          return true;
        }
      }
      return false;
    };

    findAndCollapse(internalItems.value);
  };

  const getSelected = () => {
    if (typeof internalSelected.value === 'object') {
      return internalSelected.value;
    }
    return null;
  };

  defineExpose({
    refresh,
    selectItem,
    expandItem,
    collapseItem,
    getSelected,
  });
</script>

<style scoped>
  .idb-tree {
    position: relative;
    width: 100%;
  }

  .idb-tree--list {
    /* 列表模式样式 */
  }

  .idb-tree--tree {
    /* 树模式样式 */
  }

  .idb-tree__empty {
    padding: 16px;
    font-size: 14px;
    color: var(--color-text-3);
    text-align: center;
  }

  .idb-tree__empty-text {
    margin-right: 4px;
  }

  .idb-tree__empty-create {
    color: rgb(var(--primary-6));
    cursor: pointer;
    transition: color 0.2s ease;
  }

  .idb-tree__empty-create:hover {
    color: rgb(var(--primary-5));
  }

  .idb-tree__list {
    /* 列表容器样式 */
  }

  /* 与现有样式保持一致 */
  .idb-tree {
    padding: 8px 4px 8px 16px;
  }

  /* 为不同模式设置不同的内边距 */
  .idb-tree--list {
    padding: 8px 4px 8px 16px;
  }

  .idb-tree--tree {
    padding: 8px 4px 8px 16px;
  }

  /* 树形连接线效果 */
  .idb-tree--tree:hover :deep(.tree-level-line) {
    background-color: var(--color-border-2);
  }
</style>
