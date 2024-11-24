<template>
  <div class="tree-view">
    <ul class="tree-list">
      <li v-for="item of props.items" :key="item.path" class="tree-item">
        <item-render :level="level" :item="item" />
      </li>
    </ul>
    <div
      v-if="level > 0"
      class="tree-level-line"
      :style="{ left: level * 8 + 'px' }"
    ></div>
  </div>
</template>

<script lang="ts" setup>
  import { FileItem } from './type';
  import ItemRender from './item-render.vue';

  const props = defineProps<{
    items: FileItem[];
    level: number;
  }>();
</script>

<script lang="ts">
  export default {
    name: 'ListRender',
  };
</script>

<style scoped>
  .tree-view {
    position: relative;
  }

  .tree-list {
    padding-left: 0;
    list-style: none;
  }

  .tree-level-line {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 1px;
    font-size: 0;
    background-color: transparent;
  }
</style>
