<template>
  <div ref="domRef" class="xterm-container" />
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount, shallowRef } from 'vue';
  import { API_BASE_URL } from '@/helper/api-helper';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { debounce } from 'lodash';
  import { MsgDo, MsgType } from './type';
  import '@xterm/xterm/css/xterm.css';

  const props = defineProps<{
    path?: string;
    hostId: number;
    sendHeartbeat?: boolean;
    type?: 'session' | 'ssh';
  }>();

  const emit = defineEmits<{
    (e: 'session', data: { sessionId: string; sessionName: string }): void;
    (e: 'wsopen'): void;
  }>();

  const domRef = ref<HTMLDivElement>();
  const wsRef = shallowRef<WebSocket>();
  const termRef = shallowRef<Terminal>();
  const latencyRef = ref<number>(0);
  const fitRef = shallowRef<FitAddon>();
  const timerRef = shallowRef<number>();
  const sessionIdRef = ref<string>();

  function isWsOpen() {
    return wsRef.value && wsRef.value.readyState === WebSocket.OPEN;
  }

  function sendWsMsg(payload: MsgDo) {
    if (isWsOpen()) {
      wsRef.value?.send(
        JSON.stringify({
          ...(props.type === 'session' ? { session: sessionIdRef.value } : {}),
          ...payload,
        })
      );
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
        termRef.value?.write(msg.data!);
        break;
      case MsgType.Heartbeat:
        latencyRef.value = Date.now() - msg.timestamp!;
        break;
      // server notify client current session name
      case MsgType.Attach:
      case MsgType.Start:
        emit('session', {
          sessionId: msg.session!,
          sessionName: msg.data!,
        });
        sessionIdRef.value = msg.session;
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
    const path = (props.path || 'terminals/{host}/ssh/start').replace(
      '{host}',
      String(props.hostId)
    );
    wsRef.value = new WebSocket(
      `${protocol}://${window.location.host}${API_BASE_URL}${path}`
    );
    wsRef.value.onerror = onWsError;
    wsRef.value.onclose = onWsClose;
    wsRef.value.onopen = () => {
      // eslint-disable-next-line no-console
      console.log(`terminal connected, host: ${props.hostId}`);
      emit('wsopen');
      if (props.sendHeartbeat) {
        autoSendHeartbeat();
      }
    };
    wsRef.value.onmessage = onWsMsgReceived;
  }

  function disconnectWs() {
    if (wsRef.value && wsRef.value.readyState === WebSocket.OPEN) {
      wsRef.value.close();
    }
  }

  function initTerminal() {
    termRef.value = new Terminal({
      lineHeight: 1.2,
      fontSize: 14,
      fontFamily:
        "'Lucida Console', 'DejaVu Sans Mono', 'Everson Mono', FreeMono, Menlo, Terminal, monospace",
      cursorStyle: 'underline',
      cursorBlink: true,
      scrollback: 100,
    });
    fitRef.value = new FitAddon();
    termRef.value.loadAddon(fitRef.value);
    termRef.value.open(domRef.value!);
    termRef.value.onData((data) => {
      sendWsMsg({
        type: MsgType.Cmd,
        data,
      });
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

  defineExpose({
    sendWsMsg,
    dispose,
  });
</script>

<style scoped>
  .xterm-container {
    width: 100%;
    height: 100%;
  }

  .xterm-container :deep(.xterm) {
    padding: 8px 16px;
  }
</style>
