<template>
  <div class="tree-view">
    <ul class="tree-list">
      <template v-for="item of props.items" :key="item.path">
        <!-- 只显示满足条件的项目（根据showHidden控制） -->
        <li v-show="props.showHidden || !item.is_hidden" class="tree-item">
          <item-render
            :item="item"
            :show-hidden="props.showHidden"
            :level="props.level"
          />
        </li>
      </template>
    </ul>

    <!-- 垂直连接线，用于显示层级关系 -->
    <div
      v-if="props.level > 0"
      class="tree-level-line"
      :style="{ left: props.level * 8 + 'px' }"
    ></div>
  </div>
</template>

<script lang="ts" setup>
  import { FileTreeItem } from './type';
  import ItemRender from './item-render.vue';

  /**
   * 组件属性定义
   * @param items - 要渲染的文件/文件夹数组
   * @param showHidden - 是否显示隐藏文件
   * @param level - 当前层级深度，用于计算缩进和连接线
   */
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
  /* 树视图容器 */
  .tree-view {
    position: relative;
  }

  /* 列表样式重置 */
  .tree-list {
    margin: 0;
    padding-left: 0;
    list-style: none;
  }

  /* 垂直连接线 */
  .tree-level-line {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 1px;
    font-size: 0;
    background-color: transparent;
  }
</style>
