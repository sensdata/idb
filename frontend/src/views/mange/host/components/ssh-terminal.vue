<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="90vh"
    placement="bottom"
    unmount-on-close
  >
    <template #title>{{ $t('manage.host.terminal.title') }}</template>
    <terminal v-if="termVisible" :host-id="hostId!" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { nextTick, ref } from 'vue';
  import Terminal from '@/components/terminal/terminal.vue';

  const hostId = ref<number>();
  const visible = ref(false);
  const termVisible = ref(false);
  function show(host: number) {
    hostId.value = host;
    visible.value = true;
    nextTick(() => {
      termVisible.value = true;
    });
  }

  function hide() {
    visible.value = false;
  }

  defineExpose({
    show,
    hide,
  });
</script>
