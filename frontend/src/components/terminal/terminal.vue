<template>
  <div ref="domRef" class="xterm-container" />
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount, shallowRef } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { API_BASE_URL } from '@/helper/api-helper';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { useConfirm } from '@/composables/confirm';
  import { installTerminalApi } from '@/api/terminal';
  import { debounce } from 'lodash';
  import { serializeQueryParams } from '@/utils';
  import { MsgType, ReceiveMsgDo, SendMsgDo } from './type';
  import '@xterm/xterm/css/xterm.css';

  const { t } = useI18n();

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
  const { confirm } = useConfirm();

  function onWsClose(ev: CloseEvent) {
    termRef.value?.write(`\x1b[31mConnection closed: ${ev.reason}\x1b[m\r\n`);
  }

  function onWsError(ev: any) {
    const message = ev.message || 'Connection error';
    if (termRef.value) {
      termRef.value.write(`\x1b[31m${message}\x1b[m\r\n`);
    }
  }

  function isWsOpen() {
    return wsRef.value && wsRef.value.readyState === WebSocket.OPEN;
  }

  function sendWsMsg(payload: Partial<SendMsgDo>) {
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
      sendWsMsg({ type: MsgType.Resize, cols, rows });
    }
  };
  const onResizeDebounce = debounce(onResize, 500);
  function addResizeListener() {
    window.addEventListener('resize', onResizeDebounce);
  }
  function removeResizeListener() {
    window.removeEventListener('resize', onResizeDebounce);
  }

  // 使用类型系统解决循环引用问题
  type WebSocketInitializer = (term: Terminal) => void;
  let initWebSocket: WebSocketInitializer;

  // 使用const定义onWsMsgReceived函数
  const onWsMsgReceived = async (ev: MessageEvent) => {
    const msg: ReceiveMsgDo = JSON.parse(ev.data);
    if (msg.code != null && msg.code !== 200) {
      termRef.value?.write(`\x1b[31m${msg.msg}\x1b[m\r\n`);
      if (msg.msg === 'ErrNotInstalled') {
        if (await confirm(t('components.terminal.session.confirmInstall'))) {
          termRef.value?.write('installing...\r\n');
          try {
            await installTerminalApi(props.hostId);
            termRef.value?.write('install success\r\n');
            // 安装成功后重新连接
            if (termRef.value) {
              initWebSocket(termRef.value);
            }
          } catch (error) {
            termRef.value?.write(
              `\x1b[31mInstallation failed: ${error}\x1b[m\r\n`
            );
          }
          return;
        }
      }
      return;
    }

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
  };

  // 定义initWebSocket函数
  initWebSocket = (term: Terminal) => {
    // 关闭已存在的连接
    if (wsRef.value && wsRef.value.readyState === WebSocket.OPEN) {
      wsRef.value.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    let path = (props.path || 'terminals/{host}/ssh/start').replace(
      '{host}',
      String(props.hostId)
    );
    path += path.indexOf('?') > -1 ? '&' : '?';
    path += serializeQueryParams({
      cols: term.cols,
      rows: term.rows,
    });

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
      onResize();
    };
    wsRef.value.onmessage = onWsMsgReceived;
  };

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
    termRef.value.focus();
    addResizeListener();
    initWebSocket(termRef.value);
  }

  function dispose() {
    removeResizeListener();
    disconnectWs();
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
    termRef.value?.dispose();
  }

  function focus() {
    termRef.value?.focus();
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
    focus,
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
