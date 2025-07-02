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
    z-index: 5;
    width: 240px;
    height: 100%;
    padding: 4px 0; /* 改为左右无padding，让紫色指示条能贴边显示 */
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
    transition: width 0.3s ease;
  }

  /* 平板设备 */
  @media screen and (width <= 991px) {
    .simplified-file-sidebar {
      width: 200px;
    }
  }

  /* 小型平板 */
  @media screen and (width <= 768px) {
    .simplified-file-sidebar {
      width: 180px;
    }
  }

  /* 手机设备 */
  @media screen and (width <= 576px) {
    .simplified-file-sidebar {
      width: 150px;
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
