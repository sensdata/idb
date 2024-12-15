<template>
  <div class="category-tree">
    <div
      v-for="cat of items"
      :key="cat"
      class="item"
      :class="{ selected: selected === cat }"
      @click="handleClick(cat)"
    >
      <div class="item-icon">
        <folder-icon />
      </div>
      <div class="item-text truncate">
        {{ cat == null ? $t('app.script.category.all') : cat }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { PropType } from 'vue';

  defineProps<{
    items: Array<string | null>;
  }>();

  const selected = defineModel('selected', {
    type: [String, null] as PropType<string | null>,
    required: true,
  });

  function handleClick(cat: string | null) {
    selected.value = cat;
  }
</script>

<style scoped>
  .category-tree {
    padding-left: 8px;
  }

  .item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    margin-bottom: 8px;
    padding-left: 10px;
    line-height: 32px;
    border-radius: 4px;
    cursor: pointer;
  }

  .item:hover {
    background-color: var(--color-fill-1);
  }

  .item.selected {
    background-color: var(--color-fill-2);
  }

  .item.selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
    content: '';
  }

  .item-icon {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 5px 0;
  }

  .item-icon svg {
    width: 14px;
    height: 14px;
  }

  .item-text {
    flex: 1;
    min-width: 0;
    margin-left: 8px;
    font-size: 14px;
    line-height: 22px;
  }
</style>
