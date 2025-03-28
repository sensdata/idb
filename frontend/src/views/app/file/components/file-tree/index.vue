<template>
  <div class="file-tree">
    <list-render :items="props.items" :show-hidden="showHidden" :level="0" />
  </div>
</template>

<script lang="ts" setup>
  import { provide, ref, watch } from 'vue';
  import ListRender from './list-render.vue';
  import { FileTreeItem } from './type';

  const props = defineProps<{
    items: FileTreeItem[];
    showHidden?: boolean;
    selected?: FileTreeItem | null;
    selectedChange: (item: FileTreeItem) => void;
    openChange: (item: FileTreeItem, open: boolean) => void;
  }>();

  const selected = ref(props.selected);
  watch(
    () => props.selected,
    (value) => {
      selected.value = value;
    }
  );
  provide('selected', selected);
  provide('selectedChange', props.selectedChange);
  provide('openChange', props.openChange);
</script>

<style scoped>
  .file-tree {
    padding-left: 8px;
  }

  .file-tree:hover :deep(.tree-level-line) {
    background-color: var(--color-border-2);
  }
</style>
