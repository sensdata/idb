<template>
  <div class="tree-view">
    <ul class="tree-list">
      <template v-for="item of props.items" :key="item.path">
        <li v-if="showHidden || !item.is_hidden" class="tree-item">
          <item-render :level="level" :item="item" :show-hidden="showHidden" />
        </li>
      </template>
    </ul>
    <div
      v-if="level > 0"
      class="tree-level-line"
      :style="{ left: level * 8 + 'px' }"
    ></div>
  </div>
</template>

<script lang="ts" setup>
  import { FileTreeItem } from './type';
  import ItemRender from './item-render.vue';

  const props = defineProps<{
    items: FileTreeItem[];
    showHidden?: boolean;
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
    margin: 0;
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
