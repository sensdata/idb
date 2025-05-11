<template>
  <a-modal
    v-model:visible="visible"
    :title="
      isUpgrade
        ? $t('manage.host.installAgent.titleUpgrade')
        : $t('manage.host.installAgent.title')
    "
    :footer="false"
    :mask-closable="false"
    :closable="true"
    width="600px"
    @cancel="handleCancel"
  >
    <div class="install-log">
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
          <span :class="['log-message', log.level]">{{ log.message }}</span>
        </div>
        <div v-if="logs.length === 0" class="empty-log">
          {{ $t('manage.host.installAgent.waitingForLogs') }}
        </div>
      </div>
      <div class="log-footer">
        <a-button v-if="isCompleted" type="primary" @click="handleCancel">
          {{ $t('manage.host.installAgent.close') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatTime } from '@/utils/format';
  import useVisible from '@/hooks/visible';
  import { Message } from '@arco-design/web-vue';
  import { resolveApiUrl } from '@/helper/api-helper';
  import {
    installHostAgentApi,
    testHostAgentApi,
    upgradeHostAgentApi,
  } from '@/api/host';
  import { useConfirm } from '@/hooks/confirm';

  const emit = defineEmits(['ok', 'cancel']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();

  interface LogItem {
    time: number;
    message: string;
    level: 'info' | 'error' | 'warn' | 'debug';
  }

  const { confirm } = useConfirm();

  const isUpgrade = ref(false);
  const logs = ref<LogItem[]>([]);
  const status = ref<'installing' | 'completed' | 'failed' | 'timeout'>(
    'installing'
  );
  const logContentRef = ref<HTMLElement | null>(null);

  const isCompleted = computed(() => {
    return status.value === 'completed' || status.value === 'failed';
  });

  const statusText = computed(() => {
    switch (status.value) {
      case 'installing':
        return t('manage.host.installAgent.statusInstalling');
      case 'completed':
        return t('manage.host.installAgent.statusCompleted');
      case 'failed':
        return t('manage.host.installAgent.statusFailed');
      case 'timeout':
        return t('manage.host.installAgent.statusTimeout');
      default:
        return '';
    }
  });

  const statusColor = computed(() => {
    switch (status.value) {
      case 'installing':
        return 'orange';
      case 'completed':
        return 'green';
      case 'failed':
        return 'red';
      case 'timeout':
        return 'arcoblue';
      default:
        return 'arcoblue';
    }
  });

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

  const logFileLogs = (hostId: number, logPath: string) => {
    const url = resolveApiUrl(
      `logs/${hostId}/follow?path=${encodeURIComponent(logPath)}&whence=end`
    );
    reset();
    show();

    let heartbeat = Date.now();
    const eventSource = new EventSource(url);
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
          Message.success(t('manage.host.installAgent.installSuccess'));
          emit('ok');
        } else if (statusValue === 'failed' || statusValue === 'canceled') {
          clearInterval(timer);
          eventSource.close();
          setStatus('failed');
          Message.error(t('manage.host.installAgent.installFailed'));
        } else if (statusValue === 'running') {
          addLog({
            time: Date.now(),
            message: t('manage.host.installAgent.installing'),
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
      Message.success(t('manage.host.installAgent.installSuccess'));
      emit('ok');
    });

    eventSource.addEventListener('error', (event) => {
      if (event.type === 'error') {
        clearInterval(timer);
        eventSource.close();
        setStatus('failed');
        Message.error(t('manage.host.installAgent.installFailed'));
      }
    });

    timer = window.setInterval(() => {
      if (Date.now() - heartbeat > 60e3) {
        clearInterval(timer);
        eventSource.close();
        setStatus('timeout');
        Message.error(t('manage.host.installAgent.installTimeout'));
      }
    }, 1000);

    eventSource.onopen = () => {
      addLog({
        time: Date.now(),
        message: t('manage.host.installAgent.logConnected'),
        level: 'info',
      });
    };

    eventSource.onerror = () => {
      if (eventSource.readyState === EventSource.CLOSED) {
        clearInterval(timer);
        setStatus('failed');
        Message.error(t('manage.host.installAgent.logConnectionFailed'));
      }
    };

    eventSource.addEventListener('end', () => {
      clearInterval(timer);
      eventSource.close();
      setStatus('completed');
      Message.success(t('manage.host.installAgent.installSuccess'));
      emit('ok');
    });
  };

  const startInstall = async (hostId: number) => {
    try {
      const result = await installHostAgentApi(hostId);
      if (result.log_path) {
        isUpgrade.value = false;
        logFileLogs(result.log_host, result.log_path);
      } else {
        Message.error(t('manage.host.installAgent.installFailed'));
      }
    } catch (error) {
      Message.error(t('manage.host.installAgent.installFailed'));
    }
  };

  const startUpgrade = async (hostId: number) => {
    try {
      const result = await upgradeHostAgentApi(hostId);
      if (result.log_path) {
        isUpgrade.value = true;
        logFileLogs(result.log_host, result.log_path);
      } else {
        Message.error(t('manage.host.installAgent.upgradeFailed'));
      }
    } catch (error) {
      Message.error(t('manage.host.installAgent.upgradeFailed'));
    }
  };

  const confirmInstall = async (hostId: number) => {
    const confirmResult = await confirm(
      t('manage.host.installAgent.notInstalled')
    );
    if (confirmResult) {
      startInstall(hostId);
    }
  };

  const checkInstall = async (hostId: number) => {
    try {
      const result = await testHostAgentApi(hostId);
      if (!result.installed) {
        confirmInstall(hostId);
      }
      return result.installed;
    } catch (error) {
      return false;
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
    startInstall,
    startUpgrade,
    confirmInstall,
    checkInstall,
  });
</script>

<style scoped>
  .install-log {
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
    font-size: 14px;
    font-family: monospace;
    line-height: 1.5;
    background-color: var(--color-fill-2);
    border-radius: 4px;
  }

  .log-item {
    margin-bottom: 4px;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .log-time {
    margin-right: 8px;
    color: var(--color-text-3);
  }

  .log-message {
    color: var(--color-text-2);
  }

  .log-message.debug {
    color: #6c757d;
  }

  .log-message.info {
    color: #212529;
  }

  .log-message.warn {
    color: #ffc107;
  }

  .log-message.error {
    color: #dc3545;
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
