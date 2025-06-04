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
      ref="termRef"
      type="session"
      :hostId="hostId!"
      path="docker/{host}/containers/terminal"
      @wsopen="handleWsOpen"
    />
  </a-drawer>
</template>

<script setup lang="ts">
  import { nextTick, onBeforeUnmount, ref, watch } from 'vue';
  import Terminal from '@/components/terminal/terminal.vue';
  import { useHostStore } from '@/store';
  import { MsgType } from '@/components/terminal/type';
  import { quitContainerTerminalApi } from '@/api/docker';

  const { currentId: hostId } = useHostStore();
  const visible = ref(false);
  const termRef = ref<InstanceType<typeof Terminal>>();
  const termVisible = ref(false);
  const conatinerIdRef = ref('');

  function show(containerId: string) {
    conatinerIdRef.value = containerId;
    visible.value = true;
    nextTick(() => {
      termVisible.value = true;
    });
  }

  function hide() {
    visible.value = false;
  }

  function handleQuit() {
    quitContainerTerminalApi({
      session: conatinerIdRef.value,
      type: 'screen',
    });
  }

  function handleWsOpen() {
    termRef.value?.sendWsMsg({
      type: MsgType.Start,
      session: conatinerIdRef.value,
      data: '/bin/bash',
    });
  }

  watch(visible, (v) => {
    if (!v) {
      handleQuit();
    }
  });

  onBeforeUnmount(handleQuit);

  defineExpose({
    show,
    hide,
  });
</script>
