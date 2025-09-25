<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.store.upgradeLog.title')"
    :footer="false"
    :mask-closable="false"
    :closable="true"
    width="600px"
    @cancel="handleCancel"
  >
    <div class="install-log">
      <div ref="logContentRef" class="log-content">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <span class="log-time"> {{ formatTime(log.time) }} </span>
          <span :class="['log-message', log.level]">{{ log.message }}</span>
        </div>
        <div v-if="logs.length === 0" class="empty-log">
          {{ $t('app.store.upgradeLog.waitingForLogs') }}
        </div>
      </div>
      <div class="log-footer">
        <a-button type="primary" @click="handleCancel">
          {{ $t('app.store.upgradeLog.close') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatTime } from '@/utils/format';
  import useVisible from '@/composables/visible';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { resolveApiUrl } from '@/helper/api-helper';

  const emit = defineEmits(['ok', 'cancel']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();

  interface LogItem {
    time: number;
    message: string;
    level: 'info' | 'error' | 'warn' | 'debug';
  }
  const logs = ref<LogItem[]>([]);
  const status = ref<'installing' | 'completed' | 'failed' | 'timeout'>(
    'installing'
  );
  const logContentRef = ref<HTMLElement | null>(null);

  const scrollToBottom = async () => {
    await nextTick();
    if (logContentRef.value) {
      // 使用 setTimeout 确保 DOM 完全更新后再滚动
      setTimeout(() => {
        if (logContentRef.value) {
          logContentRef.value.scrollTop = logContentRef.value.scrollHeight;
        }
      }, 10);
    }
  };

  watch(
    logs,
    () => {
      scrollToBottom();
    },
    { deep: true }
  );

  const addLog = (log: LogItem) => {
    logs.value.push(log);
    // 确保在添加日志后立即滚动到底部
    nextTick(() => {
      scrollToBottom();
    });
  };

  const clearLogs = () => {
    logs.value = [];
  };

  const setStatus = (
    newStatus: 'installing' | 'completed' | 'failed' | 'timeout'
  ) => {
    status.value = newStatus;
  };

  const reset = () => {
    clearLogs();
    setStatus('installing');
  };

  const processLogData = (data: string): LogItem => {
    let logMessage = data;
    let logTime = Date.now();
    let logLevel: 'info' | 'error' | 'warn' | 'debug' = 'info';

    if (data.trim().startsWith('{') && data.trim().endsWith('}')) {
      try {
        const logData = JSON.parse(data);
        if (logData && typeof logData === 'object') {
          if (logData.message) {
            logMessage = logData.message;
          }

          if (logData.timestamp) {
            logTime =
              typeof logData.timestamp === 'number'
                ? logData.timestamp
                : new Date(logData.timestamp).getTime() || Date.now();
          } else if (logData.time) {
            logTime =
              typeof logData.time === 'number'
                ? logData.time
                : new Date(logData.time).getTime() || Date.now();
          }

          if (logData.level) {
            const level = String(logData.level).toLowerCase();
            if (['error', 'warn', 'debug', 'info'].includes(level)) {
              logLevel = level as 'info' | 'error' | 'warn' | 'debug';
            }
          }
        }
      } catch (error) {
        console.warn('Failed to parse log JSON:', error);
      }
    } else if (data.includes('Status:')) {
      logMessage = data.trim();
    }

    return {
      time: logTime,
      message: logMessage,
      level: logLevel,
    };
  };

  const eventSourceRef = ref<EventSource | null>(null);
  const logFileLogs = (hostId: number, logPath: string) => {
    const url = resolveApiUrl(
      `logs/${hostId}/follow?path=${encodeURIComponent(logPath)}&whence=end`
    );
    reset();
    show();

    let heartbeat = Date.now();
    const eventSource = new EventSource(url);
    eventSourceRef.value = eventSource;
    let timer: number;

    eventSource.addEventListener('log', (event: Event) => {
      if (event instanceof MessageEvent) {
        if (event.data) {
          try {
            const rawData = event.data.trim();

            if (rawData.startsWith('{') && rawData.endsWith('}')) {
              addLog(processLogData(rawData));
            } else {
              const jsonMatch = rawData.match(/(\{.*\})/);
              if (jsonMatch && jsonMatch[1]) {
                addLog(processLogData(jsonMatch[1]));
              } else {
                addLog({
                  time: Date.now(),
                  message: rawData,
                  level: 'info',
                });
              }
            }
          } catch (error) {
            addLog({
              time: Date.now(),
              message: event.data,
              level: 'info',
            });
          }
        }
      }
    });

    eventSource.addEventListener('data', (event: Event) => {
      if (event instanceof MessageEvent) {
        if (event.data) {
          try {
            const rawData = event.data.trim();

            if (rawData.startsWith('{') && rawData.endsWith('}')) {
              addLog(processLogData(rawData));
            } else {
              const jsonMatch = rawData.match(/(\{.*\})/);
              if (jsonMatch && jsonMatch[1]) {
                addLog(processLogData(jsonMatch[1]));
              } else {
                addLog({
                  time: Date.now(),
                  message: rawData,
                  level: 'info',
                });
              }
            }
          } catch (error) {
            addLog({
              time: Date.now(),
              message: event.data,
              level: 'info',
            });
          }
        }
      }
    });

    eventSource.addEventListener('status', (event: Event) => {
      if (event instanceof MessageEvent) {
        const statusValue = event.data;

        if (statusValue === 'success') {
          clearInterval(timer);
          eventSource.close();
          setStatus('completed');
          Message.success(t('app.store.upgradeLog.success'));
          emit('ok');
        } else if (statusValue === 'failed' || statusValue === 'canceled') {
          clearInterval(timer);
          eventSource.close();
          setStatus('failed');
          showErrorWithDockerCheck(t('app.store.upgradeLog.failed'));
        } else if (statusValue === 'running') {
          addLog({
            time: Date.now(),
            message: t('app.store.upgradeLog.progress'),
            level: 'info',
          });
        } else {
          addLog({
            time: Date.now(),
            message: `Status: ${statusValue}`,
            level: 'info',
          });
        }
      }
    });

    eventSource.onmessage = (event) => {
      if (event.data) {
        try {
          const rawData = event.data.trim();

          if (rawData.startsWith('{') && rawData.endsWith('}')) {
            addLog(processLogData(rawData));
          } else {
            const jsonMatch = rawData.match(/(\{.*\})/);
            if (jsonMatch && jsonMatch[1]) {
              addLog(processLogData(jsonMatch[1]));
            } else {
              addLog({
                time: Date.now(),
                message: rawData,
                level: 'info',
              });
            }
          }
        } catch (error) {
          addLog({
            time: Date.now(),
            message: event.data,
            level: 'info',
          });
        }
      }
    };

    eventSource.addEventListener('heartbeat', () => {
      heartbeat = Date.now();
    });

    eventSource.addEventListener('close', () => {
      clearInterval(timer);
      eventSource.close();
      setStatus('completed');
      Message.success(t('app.store.upgradeLog.installSuccess'));
      emit('ok');
    });

    eventSource.addEventListener('error', (event) => {
      if (event.type === 'error') {
        clearInterval(timer);
        eventSource.close();
        setStatus('failed');
        showErrorWithDockerCheck(t('app.store.upgradeLog.installFailed'));
      }
    });

    timer = window.setInterval(() => {
      if (Date.now() - heartbeat > 60e3) {
        clearInterval(timer);
        eventSource.close();
        setStatus('timeout');
        showErrorWithDockerCheck(t('app.store.upgradeLog.installTimeout'));
      }
    }, 1000);

    eventSource.onopen = () => {
      addLog({
        time: Date.now(),
        message: t('app.store.upgradeLog.logConnected'),
        level: 'info',
      });
    };

    eventSource.onerror = () => {
      if (eventSource.readyState === EventSource.CLOSED) {
        clearInterval(timer);
        setStatus('failed');
        showErrorWithDockerCheck(t('app.store.upgradeLog.logConnectionFailed'));
      }
    };

    eventSource.addEventListener('end', () => {
      clearInterval(timer);
      eventSource.close();
      setStatus('completed');
      Message.success(t('app.store.upgradeLog.installSuccess'));
      emit('ok');
    });
  };

  const handleCancel = () => {
    hide();
    if (eventSourceRef.value) {
      eventSourceRef.value.close();
    }
    emit('cancel');
  };

  defineExpose({
    show,
    hide,
    reset,
    addLog,
    setStatus,
    logFileLogs,
  });
</script>

<style scoped>
  .install-log {
    display: flex;
    flex-direction: column;
    height: 400px;
  }

  .log-content {
    flex: 1;
    padding: 12px;
    overflow-y: auto;
    font-family: monospace;
    font-size: 14px;
    line-height: 1.5;
    background-color: var(--color-fill-2);
    border-radius: 4px;
    scroll-behavior: smooth;
  }

  .log-item {
    margin-bottom: 4px;
    word-break: break-all;
    white-space: pre-wrap;
  }

  .log-time {
    margin-right: 8px;
    color: var(--color-text-3);
  }

  .log-message {
    color: var(--color-text-2);
  }

  .log-message.debug {
    color: var(--color-text-3);
  }

  .log-message.info {
    color: var(--color-text-1);
  }

  .log-message.warn {
    color: var(--idbdusk-6);
  }

  .log-message.error {
    color: var(--idbred-6);
  }

  .empty-log {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--color-text-3);
  }

  .log-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
