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

  // 计算终端主题 - 使用CSS变量替代硬编码颜色
  const terminalTheme = computed(() => {
    const isDark = appStore.theme === 'dark';

    if (isDark) {
      // 深色主题 - 使用CSS变量
      return {
        background: 'var(--color-bg-1)',
        foreground: 'var(--color-text-1)',
        cursor: 'var(--color-text-1)',
        cursorAccent: 'var(--color-bg-1)',
        selectionBackground: 'var(--color-fill-3)',
        selectionForeground: undefined,
        // ANSI 颜色 - 使用品牌色和系统色
        black: 'var(--color-text-4)',
        red: 'var(--idbred-6)',
        green: 'var(--idbgreen-6)',
        yellow: 'var(--idbdusk-6)',
        blue: 'var(--idblue-6)',
        magenta: 'var(--idbautumn-6)',
        cyan: 'var(--idbturquoise-6)',
        white: 'var(--color-text-2)',
        // 亮色 ANSI 颜色 - 使用浅色变体
        brightBlack: 'var(--color-text-3)',
        brightRed: 'var(--idbred-5)',
        brightGreen: 'var(--idbgreen-5)',
        brightYellow: 'var(--idbdusk-5)',
        brightBlue: 'var(--idblue-5)',
        brightMagenta: 'var(--idbautumn-5)',
        brightCyan: 'var(--idbturquoise-5)',
        brightWhite: 'var(--color-text-1)',
      };
    }
    // 亮色主题 - 使用浅色背景
    return {
      background: 'var(--color-bg-2)',
      foreground: 'var(--color-text-1)',
      cursor: 'var(--color-text-1)',
      cursorAccent: 'var(--color-bg-2)',
      selectionBackground: 'var(--color-fill-2)',
      selectionForeground: undefined,
      // ANSI 颜色 - 适合浅色背景的深色配色
      black: 'var(--color-text-1)',
      red: 'var(--idbred-6)',
      green: 'var(--idbgreen-6)',
      yellow: 'var(--idbdusk-6)',
      blue: 'var(--idblue-6)',
      magenta: 'var(--idbautumn-6)',
      cyan: 'var(--idbturquoise-6)',
      white: 'var(--color-text-4)',
      // 亮色 ANSI 颜色 - 使用深色变体以在浅色背景上显示
      brightBlack: 'var(--color-text-2)',
      brightRed: 'var(--idbred-4)',
      brightGreen: 'var(--idbgreen-4)',
      brightYellow: 'var(--idbdusk-4)',
      brightBlue: 'var(--idblue-4)',
      brightMagenta: 'var(--idbautumn-4)',
      brightCyan: 'var(--idbturquoise-4)',
      brightWhite: 'var(--color-text-1)',
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
    if (
      msg.code != null &&
      msg.code !== 200 &&
      msg.type !== MsgType.Heartbeat
    ) {
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
      handleResize();
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
  }

  .xterm-container :deep(.xterm) {
    padding: 8px 16px;
  }

  /* 强制覆盖默认颜色样式，使用CSS变量 */
  .xterm-container :deep(.xterm-bg-257) {
    background-color: transparent !important;
  }

  .xterm-container :deep(.xterm-fg-257) {
    color: var(--color-text-1) !important;
  }
</style>
