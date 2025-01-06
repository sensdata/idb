<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="90vh"
    placement="bottom"
  >
    <template #title>{{ $t('components.terminal.title') }}</template>
    <terminal-tabs ref="tabsRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useHostStore } from '@/store';
  import { getTerminalSessionsApi } from '@/api/terminal';
  import TerminalTabs from './terminal-tabs.vue';

  const visible = defineModel('visible', {
    type: Boolean,
    required: true,
  });
  const hostStore = useHostStore();
  const tabsRef = ref<InstanceType<typeof TerminalTabs> | null>(null);

  async function addFirstSession() {
    const sessions = await getTerminalSessionsApi(hostStore.currentId!);

    if (sessions.length) {
      tabsRef?.value.attachSession(sessions[0]);
    } else {
      tabsRef?.value.addSession(hostStore.currentId!);
    }
  }

  const firstShow = ref(true);
  watch(visible, (val) => {
    if (val && firstShow.value) {
      firstShow.value = false;
      addFirstSession();
    }
  });
</script>
