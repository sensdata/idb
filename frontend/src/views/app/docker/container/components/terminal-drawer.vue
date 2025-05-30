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
      :hostId="hostId!"
      path="docker/{host}/containers/terminal"
      @wsopen="handleWsOpen"
    />
  </a-drawer>
</template>

<script setup lang="ts">
  import { nextTick, ref } from 'vue';
  import Terminal from '@/components/terminal/terminal.vue';
  import { useHostStore } from '@/store';
  import { MsgType } from '@/components/terminal/type';

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

  function handleWsOpen() {
    console.log('handleWsOpen', conatinerIdRef.value);
    termRef.value?.sendWsMsg({
      type: MsgType.Start,
      session: conatinerIdRef.value,
    });
  }

  defineExpose({
    show,
    hide,
  });
</script>
