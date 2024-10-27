<template>
  <div class="fixed-footer-bar" :style="`width: calc(100% - ${menuWidth}px);`">
    <a-row class="h-full">
      <a-col :span="12">
        <slot name="left"></slot>
      </a-col>
      <a-col :span="12" class="flex h-full items-center justify-end">
        <slot name="right"></slot>
      </a-col>
    </a-row>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useAppStore } from '@/store';

  const appStore = useAppStore();
  const menuWidth = computed(() => {
    return appStore.menuCollapse ? 48 : appStore.menuWidth;
  });
</script>

<style lang="less" scoped>
  .fixed-footer-bar {
    position: fixed;
    right: 0;
    bottom: 0;
    z-index: 99;
    height: 60px;
    padding: 0 24px;
    line-height: 60px;
    background: var(--color-bg-2);
    border-top: 1px solid var(--color-border);
  }
</style>
