<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="100vh"
    placement="bottom"
  >
    <template #title>{{ $t('components.terminalDrawer.title') }}</template>
    <terminal-tabs ref="tabsRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useHostStore } from '@/store';
  import TerminalTabs from '@/components/terminal-tabs/index.vue';

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
      tabsRef?.value?.addTerm(hostStore.default);
    }
  });
</script>
