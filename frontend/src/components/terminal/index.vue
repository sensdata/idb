<template>
  <div ref="domRef" />
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { AttachAddon } from '@xterm/addon-attach';
  import { debounce } from 'lodash';

  const props = defineProps<{
    hostId: number;
  }>();

  const { t } = useI18n();

  const domRef = ref<HTMLDivElement>();
  const wsRef = ref<WebSocket>();
  const termRef = ref<Terminal>();
  const fitRef = ref<FitAddon>();
  const attachRef = ref<AttachAddon>();

  const onResize = debounce(() => {
    fitRef.value?.fit();
  }, 800);
  function addResizeListener() {
    window.addEventListener('resize', onResize);
  }
  function removeResizeListener() {
    window.removeEventListener('resize', onResize);
  }

  function initTerminal() {
    termRef.value = new Terminal({
      lineHeight: 1.2,
      fontSize: 12,
      fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
      cursorStyle: 'underline',
      cursorBlink: true,
      scrollback: 100,
    });
    attachRef.value = new AttachAddon(wsRef.value!);
    fitRef.value = new FitAddon();
    termRef.value.loadAddon(attachRef.value);
    termRef.value.loadAddon(fitRef.value);
    termRef.value.open(domRef.value!);
    fitRef.value.fit();
    addResizeListener();
  }

  function initWs() {
    wsRef.value = new WebSocket(`ws://localhost:3000?id=${props.hostId}`);
    wsRef.value.onopen = () => {
      initTerminal();
    };
    wsRef.value.onerror = (e) => {
      // eslint-disable-next-line no-console
      console.warn('WebSocket error', e);
      Message.error(t('components.xterm.connectError'));
    };
  }

  function disconnectWs() {
    if (wsRef.value && wsRef.value.readyState === WebSocket.OPEN) {
      wsRef.value.close();
    }
  }

  onMounted(() => {
    initWs();
  });
  onBeforeUnmount(() => {
    removeResizeListener();
    disconnectWs();
  });
</script>

<style lang="scss" scoped>
  #terminal {
    width: 100%;
    height: 100%;
  }
</style>
