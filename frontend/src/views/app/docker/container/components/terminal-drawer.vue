<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="90vh"
    placement="bottom"
    unmount-on-close
  >
    <template #title>{{ $t('manage.host.terminal.title') }}</template>
    <terminal
      v-if="termVisible && hostId"
      path="docker/{host}/containers/terminal"
      :hostId="hostId!"
    />
  </a-drawer>
</template>

<script setup lang="ts">
  import { nextTick, ref } from 'vue';
  import Terminal from '@/components/terminal/terminal.vue';
  import { useHostStore } from '@/store';

  const { currentId: hostId } = useHostStore();
  const visible = ref(false);
  const termVisible = ref(false);
  function show() {
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
