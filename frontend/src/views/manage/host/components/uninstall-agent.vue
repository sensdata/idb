<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.uninstallAgent.title')"
    :footer="false"
    :mask-closable="false"
    :closable="true"
    width="600px"
    @cancel="handleCancel"
  >
    <div class="uninstall-log">
      <div class="log-header">
        <div class="status">
          <a-tag :color="statusColor">{{ statusText }}</a-tag>
        </div>
        <div v-if="!isCompleted" class="progress">
          <a-progress
            :percent="100"
            :show-text="false"
            status="warning"
            :stroke-width="4"
            animation
          />
        </div>
      </div>
      <div ref="logContentRef" class="log-content">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <span class="log-time"> {{ formatTime(log.time) }} </span>
          <span :class="['log-message', log.level]">{{
            formatLogMessage(log.message)
          }}</span>
        </div>
        <div v-if="logs.length === 0" class="empty-log">
          {{ $t('manage.host.uninstallAgent.waitingForLogs') }}
        </div>
      </div>
      <div class="log-footer">
        <a-button v-if="isCompleted" type="primary" @click="handleCancel">
          {{ $t('manage.host.uninstallAgent.close') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatTime } from '@/utils/format';
  import useVisible from '@/composables/visible';
  import { Message } from '@arco-design/web-vue';
  import { resolveApiUrl } from '@/helper/api-helper';
  import { uninstallHostAgentApi } from '@/api/host';

  const emit = defineEmits(['ok', 'cancel']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();

  interface LogItem {
    time: number;
    message: string;
    level: 'info' | 'error' | 'warn' | 'debug';
  }

  const logs = ref<LogItem[]>([]);
  const status = ref<'uninstalling' | 'completed' | 'failed' | 'timeout'>(
    'uninstalling'
  );
  const logContentRef = ref<HTMLElement | null>(null);

  const isCompleted = computed(() => {
    return status.value === 'completed' || status.value === 'failed';
  });

  const statusText = computed(() => {
    switch (status.value) {
      case 'uninstalling':
        return t('manage.host.uninstallAgent.statusUninstalling');
      case 'completed':
        return t('manage.host.uninstallAgent.statusCompleted');
      case 'failed':
        return t('manage.host.uninstallAgent.statusFailed');
      case 'timeout':
        return t('manage.host.uninstallAgent.statusTimeout');
      default:
        return '';
    }
  });

  const statusColor = computed(() => {
    switch (status.value) {
      case 'uninstalling':
        return 'rgb(var(--warning-6))';
      case 'completed':
        return 'rgb(var(--success-6))';
      case 'failed':
        return 'rgb(var(--danger-6))';
      case 'timeout':
        return 'rgb(var(--primary-6))';
      default:
        return 'rgb(var(--primary-6))';
    }
  });

  const formatLogMessage = (message: string): string => {
    const escaped = message
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');

    return escaped;
  };

  const scrollToBottom = async () => {
    await nextTick();
    if (logContentRef.value) {
      logContentRef.value.scrollTop = logContentRef.value.scrollHeight;
    }
  };

  watch(logs, () => {
    scrollToBottom();
  });

  const addLog = (log: LogItem) => {
    logs.value.push(log);
  };

  const clearLogs = () => {
    logs.value = [];
  };

  const setStatus = (
    newStatus: 'uninstalling' | 'completed' | 'failed' | 'timeout'
  ) => {
    status.value = newStatus;
  };

  const reset = () => {
    clearLogs();
    setStatus('uninstalling');
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

  const logFileLogs = (hostId: number, logPath: string) => {
    const url = resolveApiUrl(
      `logs/${hostId}/follow?path=${encodeURIComponent(logPath)}&whence=end`
    );

    let heartbeat = Date.now();
    const eventSource = new EventSource(url);
    let timer: number;

    const handleLogEvent = (event: MessageEvent) => {
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

    eventSource.addEventListener('log', handleLogEvent);
    eventSource.addEventListener('data', handleLogEvent);
    eventSource.onmessage = handleLogEvent;

    eventSource.addEventListener('status', (event: Event) => {
      if (event instanceof MessageEvent) {
        const statusValue = event.data;

        if (statusValue === 'success') {
          clearInterval(timer);
          eventSource.close();
          setStatus('completed');
          Message.success(t('manage.host.uninstallAgent.uninstallSuccess'));
          emit('ok');
        } else if (['failed', 'canceled'].includes(statusValue)) {
          clearInterval(timer);
          eventSource.close();
          setStatus('failed');
          Message.error(t('manage.host.uninstallAgent.uninstallFailed'));
        } else if (statusValue === 'running') {
          addLog({
            time: Date.now(),
            message: t('manage.host.uninstallAgent.statusUninstalling'),
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

    eventSource.addEventListener('heartbeat', () => {
      heartbeat = Date.now();
    });

    const handleClose = () => {
      clearInterval(timer);
      eventSource.close();
      setStatus('completed');
      Message.success(t('manage.host.uninstallAgent.uninstallSuccess'));
      emit('ok');
    };

    eventSource.addEventListener('close', handleClose);
    eventSource.addEventListener('end', handleClose);

    eventSource.onerror = () => {
      if (eventSource.readyState === EventSource.CLOSED) {
        clearInterval(timer);
        setStatus('failed');
        Message.error(t('manage.host.uninstallAgent.logConnectionFailed'));
      }
    };

    eventSource.addEventListener('error', (event) => {
      if (event.type === 'error') {
        clearInterval(timer);
        eventSource.close();
        setStatus('failed');
        Message.error(t('manage.host.uninstallAgent.uninstallFailed'));
      }
    });

    timer = window.setInterval(() => {
      if (Date.now() - heartbeat > 60e3) {
        clearInterval(timer);
        eventSource.close();
        setStatus('timeout');
        Message.error(t('manage.host.uninstallAgent.uninstallTimeout'));
      }
    }, 1000);

    eventSource.onopen = () => {
      addLog({
        time: Date.now(),
        message: t('manage.host.uninstallAgent.logConnected'),
        level: 'info',
      });
    };
  };

  const startUninstall = async (id: number) => {
    reset();
    show();
    const result = await uninstallHostAgentApi(id).catch((error) => {
      Message.error(
        error?.message || t('manage.host.list.uninstallAgent.failed')
      );
      return null;
    });

    if (result?.log_path) {
      logFileLogs(result.log_host, result.log_path);
    } else if (result) {
      Message.error(t('manage.host.list.uninstallAgent.failed'));
    }
  };

  const handleCancel = () => {
    if (isCompleted.value) {
      hide();
      emit('cancel');
    }
  };

  defineExpose({
    show,
    hide,
    reset,
    addLog,
    setStatus,
    logFileLogs,
    startUninstall,
  });
</script>

<style scoped>
  .uninstall-log {
    display: flex;
    flex-direction: column;
    height: 400px;
  }

  .log-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .progress {
    flex: 1;
    margin-left: 12px;
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
