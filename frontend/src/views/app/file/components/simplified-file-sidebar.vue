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

<style scoped lang="less">
  @import '@/assets/style/mixin.less';

  .simplified-file-sidebar {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 5;
    width: 208px;
    height: calc(100vh - 240px); /* 改为固定的视口相对高度，避免依赖父容器 */
    padding: 4px 0; /* 改为左右无padding，让紫色指示条能贴边显示 */
    overflow: hidden auto; /* 隐藏水平滚动条 */ /* 明确指定垂直滚动 */
    border-right: 1px solid var(--color-border-2);
    transition: width 0.3s ease;

    /* Apply custom scrollbar styling using mixin */
    .custom-scrollbar();
  }

  /* 平板设备 */
  @media screen and (width <= 991px) {
    .simplified-file-sidebar {
      width: 180px;
    }
  }

  /* 小型平板 */
  @media screen and (width <= 768px) {
    .simplified-file-sidebar {
      width: 160px;
    }
  }

  /* 手机设备 */
  @media screen and (width <= 576px) {
    .simplified-file-sidebar {
      width: 140px;
      padding: 4px 4px;
    }
  }

  /* 小型手机 */
  @media screen and (width <= 480px) {
    .simplified-file-sidebar {
      width: 120px;
      padding: 4px 2px;
    }
  }
</style>
