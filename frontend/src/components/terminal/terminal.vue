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
  const fitRef = shallowRef<FitAddon>();
  const timerRef = shallowRef<number>();
  const sessionIdRef = ref<string>();
  const wsMessageQueue: Partial<SendMsgDo>[] = [];
  const { confirm } = useConfirm();
  const MAX_WS_QUEUE_SIZE = 100;
  let suppressCloseNotice = false;

  // 终端主题配置 - 始终使用深色主题，不跟随系统主题变化
  // 使用硬编码颜色值，确保在黑色背景上所有文字清晰可见
  const terminalTheme = computed(() => {
    return {
      // 深色终端背景
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4',
      cursorAccent: '#1e1e1e',
      selectionBackground: 'rgba(255, 255, 255, 0.3)',
      // ANSI 标准颜色 - 适合深色背景
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      // ANSI 亮色 - 更亮的变体，确保在深色背景上清晰可见
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#ffffff',
    };
  });

  function onWsClose(ev: CloseEvent) {
    if (suppressCloseNotice) {
      suppressCloseNotice = false;
      return;
    }
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

  function enqueueWsMsg(payload: Partial<SendMsgDo>) {
    // resize/heartbeat 属于瞬时状态，不需要排队
    if (payload.type === MsgType.Resize || payload.type === MsgType.Heartbeat) {
      return;
    }

    if (wsMessageQueue.length >= MAX_WS_QUEUE_SIZE) {
      wsMessageQueue.shift();
    }

    wsMessageQueue.push(payload);
  }

  function flushWsQueue() {
    if (!isWsOpen() || wsMessageQueue.length === 0) {
      return;
    }

    while (wsMessageQueue.length > 0) {
      const payload = wsMessageQueue.shift();
      if (!payload) {
        continue;
      }
      wsRef.value?.send(
        JSON.stringify({
          ...(props.type === 'session' ? { session: sessionIdRef.value } : {}),
          ...payload,
        })
      );
    }
  }

  function sendWsMsg(payload: Partial<SendMsgDo>) {
    if (isWsOpen()) {
      wsRef.value?.send(
        JSON.stringify({
          ...(props.type === 'session' ? { session: sessionIdRef.value } : {}),
          ...payload,
        })
      );
      return;
    }

    enqueueWsMsg(payload);
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

  // 处理窗口大小变化时的终端尺寸调整
  const handleResize = () => {
    fitRef.value?.fit();
    if (termRef.value) {
      const { cols, rows } = termRef.value;
      sendWsMsg({ type: MsgType.Resize, cols, rows });
    }
  };

  const debouncedResize = debounce(handleResize, 150);

  let resizeObserver: ResizeObserver | undefined;

  // 使用类型系统解决循环引用问题
  type WebSocketInitializer = (term: Terminal) => void;
  let initWebSocket: WebSocketInitializer;

  // 使用const定义onWsMsgReceived函数
  const onWsMsgReceived = async (ev: MessageEvent) => {
    const msg: ReceiveMsgDo = JSON.parse(ev.data);
    if (msg.code != null && msg.code !== 200 && msg.code !== 0) {
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
        // latencyRef.value = Date.now() - msg.timestamp!;
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
      suppressCloseNotice = true;
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
      emit('wsopen');
      flushWsQueue();
      if (props.sendHeartbeat) {
        autoSendHeartbeat();
      }
      handleResize();
    };
    wsRef.value.onmessage = onWsMsgReceived;
  };

  function disconnectWs() {
    if (wsRef.value && wsRef.value.readyState === WebSocket.OPEN) {
      suppressCloseNotice = true;
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
    if (domRef.value) {
      resizeObserver = new ResizeObserver(() => {
        debouncedResize();
      });
      resizeObserver.observe(domRef.value);
    }
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
    resizeObserver?.disconnect();
    resizeObserver = undefined;
    disconnectWs();
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
    wsMessageQueue.splice(0);
    stopWatchingTheme();
    termRef.value?.dispose();
  }

  // 聚焦终端
  const focus = () => {
    termRef.value?.focus();
  };

  // 适配终端尺寸
  const fit = () => {
    fitRef.value?.fit();
  };

  // 强制重新适配终端尺寸，解决标签页切换后的显示问题
  const forceRefit = () => {
    if (!fitRef.value || !termRef.value) return;

    // 使用双重 requestAnimationFrame 确保DOM完全更新
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        fitRef.value?.fit();
        if (termRef.value) {
          const { cols, rows } = termRef.value;
          sendWsMsg({ type: MsgType.Resize, cols, rows });
        }
      });
    });
  };

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
    fit,
    forceRefit,
  });
</script>

<style scoped>
  .xterm-container {
    width: 100%;
    height: 100%;
    background-color: #1e1e1e;
  }

  .xterm-container :deep(.xterm) {
    padding: 8px 16px;
  }

  /* 覆盖 xterm.js 默认的前景/背景色类，确保与终端主题一致 */
  .xterm-container :deep(.xterm-bg-257) {
    background-color: transparent !important;
  }

  .xterm-container :deep(.xterm-fg-257) {
    color: #d4d4d4 !important;
  }
</style>
