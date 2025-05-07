<template>
  <div class="simplified-file-sidebar">
    <simplified-file-tree
      :items="items"
      :show-hidden="showHidden"
      :current="current"
      @item-select="onItemSelect"
      @item-double-click="onItemDoubleClick"
    />
  </div>
</template>

<script lang="ts" setup>
  import { SimpleFileInfoEntity } from '@/entity/FileInfo';
  import SimplifiedFileTree from './simplified-file-tree.vue';
  import { FileTreeItem } from './file-tree/type';

  defineProps({
    items: {
      type: Array as () => FileTreeItem[],
      required: true,
    },
    showHidden: {
      type: Boolean,
      default: false,
    },
    current: {
      type: Object as () => SimpleFileInfoEntity | null,
      default: null,
    },
  });

  const emit = defineEmits(['itemSelect', 'itemDoubleClick']);

  const onItemSelect = (item: FileTreeItem) => {
    emit('itemSelect', item);
  };

  const onItemDoubleClick = (item: FileTreeItem) => {
    emit('itemDoubleClick', item);
  };
</script>

<style scoped>
  .simplified-file-sidebar {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 240px;
    height: 100%;
    padding: 4px 8px;
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
  }
</style>
