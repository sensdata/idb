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
  @import url('@/assets/style/mixin.less');

  .simplified-file-sidebar {
    flex: 0 0 208px;
    align-self: stretch;
    width: 208px;
    min-height: 0;
    padding: 4px 0;
    overflow: hidden auto;
    background-color: var(--color-fill-1);
    border-right: 1px solid var(--color-border-2);
    transition: width 0.3s ease;

    /* Apply custom scrollbar styling using mixin */
    .custom-scrollbar();
  }

  /* 平板设备 */
  @media screen and (width <= 991px) {
    .simplified-file-sidebar {
      flex-basis: 180px;
      width: 180px;
    }
  }

  /* 小型平板 */
  @media screen and (width <= 768px) {
    .simplified-file-sidebar {
      flex-basis: 160px;
      width: 160px;
    }
  }

  /* 手机设备 */
  @media screen and (width <= 576px) {
    .simplified-file-sidebar {
      flex-basis: 140px;
      width: 140px;
      padding: 4px 4px;
    }
  }

  /* 小型手机 */
  @media screen and (width <= 480px) {
    .simplified-file-sidebar {
      flex-basis: 120px;
      width: 120px;
      padding: 4px 2px;
    }
  }
</style>
