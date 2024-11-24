<template>
  <div
    class="tree-item-container"
    :class="{ selected: selected === item.path }"
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
    v-if="loading"
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
  import useLoading from '@/hooks/loading';
  import DownIcon from '@/assets/icons/down.svg';
  import RightIcon from '@/assets/icons/direction-right.svg';
  import ListRender from './list-render.vue';
  import IconRender from './icon-render';
  import { FileItem } from './type';

  const props = defineProps<{
    item: FileItem;
    level: number;
  }>();
  const selected = inject<Ref<string>>('selected')!;

  const { loading, setLoading } = useLoading(false);

  function loadChildren() {
    setLoading(true);
    window.setTimeout(() => {
      Object.assign(props.item, {
        items: [
          {
            name: props.item.name + '-1',
            path: `idb-prd/apps/my-sql/aaa/aaa1/aaa1-2/aaa1-2${Math.random()
              .toString(32)
              .slice(2)}`,
            is_dir: true,
          },
        ],
      });
      Object.assign(props.item, { open: true });
      setLoading(false);
    }, 3000);
  }

  function handleToggle() {
    if (!props.item.is_dir) {
      return;
    }

    if (props.item.open) {
      Object.assign(props.item, { open: false });
      return;
    }

    if (props.item.items) {
      Object.assign(props.item, { open: true });
      return;
    }

    loadChildren();
  }

  function handleClick() {
    handleToggle();
    selected.value = props.item.path;
  }
</script>

<style scoped>
  .tree-item-container {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    overflow: hidden;
    line-height: 32px;
    border-radius: 4px;
    cursor: pointer;
  }

  .tree-item-container:hover,
  .tree-item-container.selected {
    background-color: var(--color-fill-2);
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
