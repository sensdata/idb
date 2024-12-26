<template>
  <div ref="domRef" class="xterm-container" />
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount, shallowRef } from 'vue';
  import { API_BASE_URL } from '@/helper/api-helper';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { debounce } from 'lodash';
  import '@xterm/xterm/css/xterm.css';

  enum MsgType {
    Heartbeat = 'heartbeat',
    Cmd = 'cmd',
  }

  interface MsgDo {
    msg_id: string;
    type: MsgType;
    sign: string;
    data: string;
    timestamp: number;
    nonce: string;
    version: string;
    checksum: string;
  }

  const props = defineProps<{
    hostId: number;
  }>();

  const domRef = ref<HTMLDivElement>();
  const wsRef = shallowRef<WebSocket>();
  const termRef = shallowRef<Terminal>();
  const latencyRef = ref<number>(0);
  const fitRef = shallowRef<FitAddon>();
  const timerRef = shallowRef<number>();

  function isWsOpen() {
    return wsRef.value && wsRef.value.readyState === WebSocket.OPEN;
  }

  function sendWsMsg(payload: {
    type: MsgType;
    data?: string;
    cols?: number;
    rows?: number;
    timestamp?: number;
  }) {
    if (isWsOpen()) {
      wsRef.value?.send(JSON.stringify(payload));
    }
  }

  function autoSendHeartbeat() {
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
    timerRef.value = window.setInterval(() => {
      if (isWsOpen()) {
        sendWsMsg({ type: MsgType.Heartbeat, timestamp: Date.now() });
      }
    }, 5e3);
  }

  const onResize = () => {
    fitRef.value?.fit();
    if (termRef.value) {
      const { cols, rows } = termRef.value;
      sendWsMsg({ type: MsgType.Cmd, cols, rows });
    }
  };
  const onResizeDebounce = debounce(onResize, 500);
  function addResizeListener() {
    window.addEventListener('resize', onResizeDebounce);
  }
  function removeResizeListener() {
    window.removeEventListener('resize', onResizeDebounce);
  }

  function onWsMsgReceived(ev: MessageEvent) {
    const msg: MsgDo = JSON.parse(ev.data);
    switch (msg.type) {
      case MsgType.Cmd:
        termRef.value?.write(msg.data);
        break;
      case MsgType.Heartbeat:
        latencyRef.value = Date.now() - msg.timestamp;
        break;
      default:
        break;
    }
  }

  function onWsClose(ev: CloseEvent) {
    termRef.value?.write(`\x1b[31mConnection closed: ${ev.reason}\x1b[m\r\n`);
  }

  function onWsError(ev: any) {
    const message = ev.message || 'Connection error';
    if (termRef.value) {
      termRef.value.write(`\x1b[31m${message}\x1b[m\r\n`);
    }
  }

  function initWs() {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    wsRef.value = new WebSocket(
      `${protocol}://${window.location.host}${API_BASE_URL}ws/terminals?host_id=${props.hostId}`
    );
    wsRef.value.onerror = onWsError;
    wsRef.value.onclose = onWsClose;
    wsRef.value.onmessage = onWsMsgReceived;
    autoSendHeartbeat();
  }

  function disconnectWs() {
    if (wsRef.value && wsRef.value.readyState === WebSocket.OPEN) {
      wsRef.value.close();
    }
  }

  function initTerminal() {
    termRef.value = new Terminal({
      lineHeight: 1.2,
      fontSize: 12,
      fontFamily: "Roboto, Monaco, Menlo, Consolas, 'Courier New', monospace",
      cursorStyle: 'underline',
      cursorBlink: true,
      scrollback: 100,
    });
    fitRef.value = new FitAddon();
    termRef.value.loadAddon(fitRef.value);
    termRef.value.open(domRef.value!);
    termRef.value.onData((data) => {
      sendWsMsg({ type: MsgType.Cmd, data });
    });
    fitRef.value.fit();
    addResizeListener();
    initWs();
  }

  function dispose() {
    removeResizeListener();
    disconnectWs();
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
    termRef.value?.dispose();
  }

  onMounted(() => {
    initTerminal();
  });
  onBeforeUnmount(() => {
    dispose();
  });
</script>

<style scoped>
  .xterm-container {
    width: 100%;
    height: 100%;
  }
</style>
