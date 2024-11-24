<template>
  <div class="file-tree">
    <list-render :items="props.items" :level="0" />
  </div>
</template>

<script lang="ts" setup>
  import { provide, ref } from 'vue';
  import { FileItem } from './type';
  import ListRender from './list-render.vue';

  const selected = ref(null);
  provide('selected', selected);

  const props = defineProps<{
    items: FileItem[];
  }>();
</script>

<style scoped>
  .file-tree:hover :deep(.tree-level-line) {
    background-color: var(--color-border-2);
  }
</style>
