<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="95vh"
    placement="bottom"
  >
    <template #title>{{ $t('components.terminal.title') }}</template>
    <terminal-tabs ref="tabsRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick } from 'vue';
  import { useHostStore } from '@/store';
  import TerminalTabs from './terminal-tabs.vue';

  const visible = defineModel('visible', {
    type: Boolean,
    required: true,
  });
  const hostStore = useHostStore();
  const tabsRef = ref<InstanceType<typeof TerminalTabs> | null>(null);

  const firstShow = ref(true);
  watch(visible, (val) => {
    if (val && firstShow.value) {
      firstShow.value = false;
      tabsRef?.value?.addItem({
        type: 'attach',
        host: hostStore.current!,
      });
    } else if (val) {
      nextTick(() => {
        tabsRef?.value?.focus();
      });
    }
  });
</script>
