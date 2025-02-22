<template>
  <div
    class="tree-item-container"
    :class="{
      selected: selected?.path === item.path,
      ['level-' + level]: true,
    }"
    :style="{ paddingLeft: level * 8 + 'px' }"
    @click="handleClick"
  >
    <div class="tree-item-toggle">
      <span v-if="item.is_dir" @click.stop="handleToggle">
        <down-icon v-if="item.open" />
        <right-icon v-else />
      </span>
    </div>
    <div class="tree-item-content">
      <div class="tree-item-icon">
        <icon-render :item="item" />
      </div>
      <div class="tree-item-text truncate">
        {{ item.name }}
      </div>
    </div>
  </div>
  <div
    v-if="item.loading"
    class="tree-item-loading"
    :style="{ paddingLeft: level * 8 + 'px' }"
  >
    <a-spin :size="14" />
    <span>{{ $t('common.loading') }}</span>
  </div>
  <list-render
    v-if="item.is_dir && item.open && item.items?.length"
    :items="item.items"
    :level="level + 1"
  />
</template>

<script lang="ts" setup>
  import { inject, Ref } from 'vue';
  import DownIcon from '@/assets/icons/down.svg';
  import RightIcon from '@/assets/icons/direction-right.svg';
  import ListRender from './list-render.vue';
  import IconRender from './icon-render';
  import { FileTreeItem } from './type';

  const props = defineProps<{
    item: FileTreeItem;
    level: number;
  }>();
  const selected = inject<Ref<FileTreeItem | undefined | null>>('selected')!;
  const selectedChange = inject<(item: FileTreeItem) => void>('selectedChange');
  const openChange =
    inject<(item: FileTreeItem, open: boolean) => void>('openChange');

  function handleToggle() {
    openChange?.(props.item, !props.item.open);
  }

  function handleClick() {
    openChange?.(props.item, true);
    selectedChange?.(props.item);
  }
</script>

<style scoped>
  .tree-item-container {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    line-height: 32px;
    border-radius: 4px;
    cursor: pointer;
  }

  .tree-item-container:hover {
    background-color: var(--color-fill-1);
  }

  .tree-item-container.selected {
    background-color: var(--color-fill-2);
  }

  .tree-item-container.selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
    content: '';
  }

  .tree-item-toggle {
    width: 12px;
    height: 100%;
  }

  .tree-item-toggle span {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 100%;
  }

  .level-0 .tree-item-toggle span {
    border-radius: 4px 0 0 4px;
  }

  .tree-item-toggle span:hover {
    background-color: var(--color-fill-3);
  }

  .tree-item-toggle svg {
    width: 12px;
    height: 12px;
  }

  .tree-item-content {
    display: flex;
    flex: 1;
    align-items: center;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 5px 8px;
  }

  .tree-item-icon {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 5px 0;
  }

  .tree-item-icon svg {
    width: 14px;
    height: 14px;
  }

  .tree-item-text {
    flex: 1;
    min-width: 0;
    margin-left: 8px;
    font-size: 14px;
    line-height: 22px;
  }

  .tree-item-loading {
    display: flex;
    place-items: center flex-start;
    height: 32px;
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 32px;
  }

  .tree-item-loading :deep(.arco-spin) {
    margin-right: 6px;
    margin-left: 16px;
  }

  .tree-item-loading :deep(.arco-spin-icon) {
    color: var(--color-text-3);
  }
</style>
