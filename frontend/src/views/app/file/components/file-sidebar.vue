<template>
  <div class="file-sidebar">
    <file-tree
      :items="foldersOnlyTree"
      :show-hidden="showHidden"
      :selected="current"
      :selected-change="onTreeItemSelect"
      :open-change="onTreeItemOpenChange"
      :double-click-change="onTreeItemDoubleClick"
    />
  </div>
</template>

<script lang="ts" setup>
  import { SimpleFileInfoEntity } from '@/entity/FileInfo';
  import FileTree from './file-tree/index.vue';
  import { FileTreeItem } from './file-tree/type';

  defineProps({
    foldersOnlyTree: {
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

  const emit = defineEmits([
    'treeItemSelect',
    'treeItemOpenChange',
    'treeItemDoubleClick',
  ]);

  const onTreeItemSelect = (record: FileTreeItem) => {
    emit('treeItemSelect', record);
  };

  const onTreeItemOpenChange = (item: FileTreeItem, open: boolean) => {
    emit('treeItemOpenChange', item, open);
  };

  const onTreeItemDoubleClick = (record: FileTreeItem) => {
    emit('treeItemDoubleClick', record);
  };
</script>

<style scoped>
  .file-sidebar {
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
