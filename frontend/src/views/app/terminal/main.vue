<template>
  <terminal-tabs ref="tabsRef" />
</template>

<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { useHostStore } from '@/store';
  import TerminalTabs from '@/components/terminal-tabs/index.vue';

  const hostStore = useHostStore();
  const tabsRef = ref<InstanceType<typeof TerminalTabs> | null>(null);

  onMounted(() => {
    tabsRef?.value?.addTerm(hostStore.current!);
  });
</script>
