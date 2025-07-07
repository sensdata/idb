<template>
  <div class="category-tree" :class="[`category-tree--${mode}`]">
    <!-- 分类标题 -->
    <div v-if="showTitle" class="category-tree__header">
      <div class="category-tree__header-title">
        {{ title || t('category.title') }}
      </div>
    </div>

    <!-- 主要的树形组件 -->
    <idb-tree
      ref="treeRef"
      v-bind="treeProps"
      :items="computedItems"
      :selected="selectedItem"
      @update:selected="handleSelectChange"
      @select="handleSelect"
      @double-click="handleDoubleClick"
      @toggle="handleToggle"
      @create="handleCreate"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import IdbTree from './index.vue';
  import { TreeItem, TreeProps, SelectedValue } from './types';

  interface Props extends Omit<TreeProps, 'items'> {
    /** 分类数据源 */
    categories?: string[];
    /** 是否显示标题 */
    showTitle?: boolean;
    /** 标题文本 */
    title?: string;
    /** 选中的分类名称 */
    selectedCategory?: string;
  }

  interface Emits {
    (e: 'update:selected', value: SelectedValue): void;
    (e: 'update:selectedCategory', value: string): void;
    (e: 'select', item: TreeItem): void;
    (e: 'doubleClick', item: TreeItem): void;
    (e: 'toggle', item: TreeItem, expanded: boolean): void;
    (e: 'create'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    mode: 'list',
    multiple: false,
    showIcon: true,
    showToggle: false,
    showHidden: false,
    emptyText: '',
    createText: '',
    showCreate: true,
    defaultIcon: FolderIcon,
    showTitle: true,
    categories: () => [],
  });

  const emit = defineEmits<Emits>();

  // 国际化
  const { t } = useI18n();

  // 组件引用
  const treeRef = ref();

  // 计算默认翻译文本
  const defaultEmptyText = computed(
    () => props.emptyText || t('tree.categoryEmpty')
  );

  const defaultCreateText = computed(
    () => props.createText || t('tree.categoryCreateText')
  );

  // 计算属性
  const computedItems = computed<TreeItem[]>(() => {
    return (props.categories || []).map((category) => ({
      id: category,
      label: category,
      icon: props.defaultIcon,
    }));
  });

  const selectedItem = computed<TreeItem | null>(() => {
    if (!props.selectedCategory) return null;
    return (
      computedItems.value.find((item) => item.id === props.selectedCategory) ||
      null
    );
  });

  const treeProps = computed(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { categories, showTitle, title, selectedCategory, ...rest } = props;
    return {
      ...rest,
      emptyText: defaultEmptyText.value,
      createText: defaultCreateText.value,
    };
  });

  // 事件处理
  const handleSelectChange = (value: any) => {
    const item = value as TreeItem | null;
    const categoryName = item ? String(item.id) : '';

    emit('update:selected', item);
    emit('update:selectedCategory', categoryName);
  };

  const handleSelect = (item: TreeItem) => {
    emit('select', item);
  };

  const handleDoubleClick = (item: TreeItem) => {
    emit('doubleClick', item);
  };

  const handleToggle = (item: TreeItem, expanded: boolean) => {
    emit('toggle', item, expanded);
  };

  const handleCreate = () => {
    emit('create');
  };

  // 监听器
  watch(
    () => props.selectedCategory,
    () => {
      // 可以在这里添加其他处理逻辑
    }
  );

  // 暴露的方法
  const refresh = () => {
    treeRef.value?.refresh();
  };

  const selectItem = (id: string | number) => {
    treeRef.value?.selectItem(id);
  };

  const expandItem = (id: string | number) => {
    treeRef.value?.expandItem(id);
  };

  const collapseItem = (id: string | number) => {
    treeRef.value?.collapseItem(id);
  };

  const getSelected = (): TreeItem | null => {
    return treeRef.value?.getSelected() || null;
  };

  const selectCategory = (categoryName: string) => {
    const item = computedItems.value.find(
      (categoryItem) => categoryItem.id === categoryName
    );
    if (item) {
      emit('update:selectedCategory', categoryName);
    }
  };

  const getCategories = (): string[] => {
    return props.categories || [];
  };

  const getSelectedCategory = (): string | null => {
    return props.selectedCategory || null;
  };

  defineExpose({
    refresh,
    selectItem,
    expandItem,
    collapseItem,
    getSelected,
    selectCategory,
    getCategories,
    getSelectedCategory,
  });
</script>

<style scoped>
  .category-tree {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .category-tree__header {
    padding: 8px 12px;
    background-color: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  .category-tree__header-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .category-tree--managing {
    /* 分类管理模式的样式 */
  }

  .category-tree :deep(.idb-tree) {
    flex: 1;
    overflow: hidden;
  }
</style>
