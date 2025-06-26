<template>
  <div ref="domRef" class="xterm-container" />
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount, shallowRef, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { debounce } from 'lodash';
  import { API_BASE_URL } from '@/helper/api-helper';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { useConfirm } from '@/composables/confirm';
  import { installTerminalApi } from '@/api/terminal';
  import { serializeQueryParams } from '@/utils';
  import { useAppStore } from '@/store';
  import { MsgType, ReceiveMsgDo, SendMsgDo } from './type';
  import '@xterm/xterm/css/xterm.css';

  const { t } = useI18n();
  const appStore = useAppStore();

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

  // 计算终端主题
  const terminalTheme = computed(() => {
    const isDark = appStore.theme === 'dark';

    if (isDark) {
      // 深色主题
      return {
        background: '#1a1a1a',
        foreground: '#e8e8e8',
        cursor: '#ffffff',
        cursorAccent: '#1a1a1a',
        selectionBackground: 'rgba(255, 255, 255, 0.3)',
        selectionForeground: undefined,
        // ANSI 颜色
        black: '#2e3436',
        red: '#cc0000',
        green: '#4e9a06',
        yellow: '#c4a000',
        blue: '#3465a4',
        magenta: '#75507b',
        cyan: '#06989a',
        white: '#d3d7cf',
        // 亮色 ANSI 颜色
        brightBlack: '#555753',
        brightRed: '#ef2929',
        brightGreen: '#8ae234',
        brightYellow: '#fce94f',
        brightBlue: '#729fcf',
        brightMagenta: '#ad7fa8',
        brightCyan: '#34e2e2',
        brightWhite: '#eeeeec',
      };
    }
    // 亮色主题 - 也使用黑色背景
    return {
      background: '#1a1a1a',
      foreground: '#e8e8e8',
      cursor: '#ffffff',
      cursorAccent: '#1a1a1a',
      selectionBackground: 'rgba(255, 255, 255, 0.3)',
      selectionForeground: undefined,
      // ANSI 颜色 - 明亮配色适合黑色背景
      black: '#2e3436',
      red: '#cc0000',
      green: '#4e9a06',
      yellow: '#c4a000',
      blue: '#3465a4',
      magenta: '#75507b',
      cyan: '#06989a',
      white: '#d3d7cf',
      // 亮色 ANSI 颜色 - 在黑色背景上的明亮色彩
      brightBlack: '#555753',
      brightRed: '#ef2929',
      brightGreen: '#8ae234',
      brightYellow: '#fce94f',
      brightBlue: '#729fcf',
      brightMagenta: '#ad7fa8',
      brightCyan: '#34e2e2',
      brightWhite: '#ffffff',
    };
  });

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
      scrollback: 5000,
      // 设置主题
      theme: terminalTheme.value,
      // 优化渲染
      convertEol: true,
      disableStdin: false,
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

  // 监听主题变化并更新终端主题
  const updateTerminalTheme = () => {
    if (termRef.value) {
      const options = termRef.value.options;
      options.theme = terminalTheme.value;
      termRef.value.refresh(0, termRef.value.rows - 1);
    }
  };

  // 监听应用主题变化
  const stopWatchingTheme = appStore.$subscribe(() => {
    updateTerminalTheme();
  });

  function dispose() {
    removeResizeListener();
    disconnectWs();
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
    stopWatchingTheme();
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

  /* 强制覆盖默认颜色样式，解决白条问题 - 始终使用深色背景 */
  .xterm-container :deep(.xterm-bg-257) {
    background-color: #1a1a1a !important;
  }

  .xterm-container :deep(.xterm-fg-257) {
    color: #e8e8e8 !important;
  }
</style>
