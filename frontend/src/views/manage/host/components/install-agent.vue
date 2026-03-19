<template>
  <a-modal
    v-model:visible="visible"
    :title="modalTitle"
    :footer="false"
    :mask-closable="false"
    :closable="isCompleted"
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
  import useVisible from '@/composables/visible';
  import { Message } from '@arco-design/web-vue';
  import { resolveApiUrl } from '@/helper/api-helper';
  import {
    installHostAgentApi,
    testHostAgentApi,
    upgradeHostAgentApi,
  } from '@/api/host';
  import { useConfirm } from '@/composables/confirm';

  const emit = defineEmits(['ok', 'cancel']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();

  interface LogItem {
    time: number;
    message: string;
    level: 'info' | 'error' | 'warn' | 'debug';
  }

  interface BatchUpgradeHost {
    id: number;
    name: string;
  }

  interface StreamLogOptions {
    resetBefore?: boolean;
    prefix?: string;
    successMessage?: string;
    failedMessage?: string;
    timeoutMessage?: string;
    emitOk?: boolean;
  }

  const { confirm } = useConfirm();

  const isUpgrade = ref(false);
  const isBatchUpgrade = ref(false);
  const batchRunning = ref(false);
  const logs = ref<LogItem[]>([]);
  const status = ref<'installing' | 'completed' | 'failed' | 'timeout'>(
    'installing'
  );
  const logContentRef = ref<HTMLElement | null>(null);

  const isCompleted = computed(() => {
    if (batchRunning.value) {
      return false;
    }
    return (
      status.value === 'completed' ||
      status.value === 'failed' ||
      status.value === 'timeout'
    );
  });

  const modalTitle = computed(() => {
    if (isBatchUpgrade.value) {
      return t('manage.host.installAgent.titleBatchUpgrade');
    }
    return isUpgrade.value
      ? t('manage.host.installAgent.titleUpgrade')
      : t('manage.host.installAgent.title');
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

  const scrollToBottom = async () => {
    await nextTick();
    if (logContentRef.value) {
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
    isUpgrade.value = false;
    isBatchUpgrade.value = false;
    batchRunning.value = false;
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

  const buildLogItem = (rawData: string): LogItem => {
    const data = rawData.trim();
    if (data.startsWith('{') && data.endsWith('}')) {
      return processLogData(data);
    }
    const jsonMatch = data.match(/(\{.*\})/);
    if (jsonMatch && jsonMatch[1]) {
      return processLogData(jsonMatch[1]);
    }
    return {
      time: Date.now(),
      message: data,
      level: 'info',
    };
  };

  const appendLogMessage = (rawData: string, prefix?: string) => {
    const logItem = buildLogItem(rawData);
    addLog({
      ...logItem,
      message: prefix ? `[${prefix}] ${logItem.message}` : logItem.message,
    });
  };

  const logFileLogs = (
    hostId: number,
    logPath: string,
    options: StreamLogOptions = {}
  ) => {
    const url = resolveApiUrl(
      `logs/${hostId}/follow?path=${encodeURIComponent(logPath)}&whence=end`
    );
    const {
      resetBefore = true,
      prefix,
      successMessage = t('manage.host.installAgent.installSuccess'),
      failedMessage = t('manage.host.installAgent.installFailed'),
      timeoutMessage = t('manage.host.installAgent.installTimeout'),
      emitOk = true,
    } = options;

    if (resetBefore) {
      reset();
    }
    show();

    let heartbeat = Date.now();
    const eventSource = new EventSource(url);
    let timer = 0;
    let finished = false;

    const finish = (
      finalStatus: 'completed' | 'failed' | 'timeout',
      toastMessage?: string
    ) => {
      if (finished) {
        return;
      }
      finished = true;
      clearInterval(timer);
      eventSource.close();
      setStatus(finalStatus);

      if (toastMessage) {
        if (finalStatus === 'completed') {
          Message.success(toastMessage);
        } else {
          Message.error(toastMessage);
        }
      }

      if (finalStatus === 'completed' && emitOk) {
        emit('ok');
      }
    };

    const handleMessage = (event: MessageEvent) => {
      if (!event.data) {
        return;
      }
      try {
        appendLogMessage(event.data, prefix);
      } catch (error) {
        addLog({
          time: Date.now(),
          message: prefix ? `[${prefix}] ${event.data}` : event.data,
          level: 'info',
        });
      }
    };

    return new Promise<'completed' | 'failed' | 'timeout'>((resolve) => {
      const resolveAndFinish = (
        finalStatus: 'completed' | 'failed' | 'timeout',
        toastMessage?: string
      ) => {
        finish(finalStatus, toastMessage);
        resolve(finalStatus);
      };

      eventSource.addEventListener('log', (event: Event) => {
        if (event instanceof MessageEvent) {
          handleMessage(event);
        }
      });

      eventSource.addEventListener('data', (event: Event) => {
        if (event instanceof MessageEvent) {
          handleMessage(event);
        }
      });

      eventSource.addEventListener('status', (event: Event) => {
        if (!(event instanceof MessageEvent)) {
          return;
        }
        const statusValue = event.data;

        if (statusValue === 'success') {
          resolveAndFinish('completed', successMessage);
          return;
        }
        if (statusValue === 'failed' || statusValue === 'canceled') {
          resolveAndFinish('failed', failedMessage);
          return;
        }
        if (statusValue === 'running') {
          addLog({
            time: Date.now(),
            message: prefix
              ? `[${prefix}] ${t('manage.host.installAgent.installing')}`
              : t('manage.host.installAgent.installing'),
            level: 'info',
          });
          return;
        }

        addLog({
          time: Date.now(),
          message: prefix
            ? `[${prefix}] Status: ${statusValue}`
            : `Status: ${statusValue}`,
          level: 'info',
        });
      });

      eventSource.addEventListener('heartbeat', () => {
        heartbeat = Date.now();
      });

      eventSource.addEventListener('close', () => {
        resolveAndFinish('completed', successMessage);
      });

      eventSource.addEventListener('end', () => {
        resolveAndFinish('completed', successMessage);
      });

      eventSource.addEventListener('error', (event) => {
        if (event.type === 'error') {
          resolveAndFinish('failed', failedMessage);
        }
      });

      eventSource.onmessage = (event) => {
        handleMessage(event);
      };

      eventSource.onopen = () => {
        addLog({
          time: Date.now(),
          message: prefix
            ? `[${prefix}] ${t('manage.host.installAgent.logConnected')}`
            : t('manage.host.installAgent.logConnected'),
          level: 'info',
        });
      };

      eventSource.onerror = () => {
        if (eventSource.readyState === EventSource.CLOSED) {
          resolveAndFinish(
            'failed',
            t('manage.host.installAgent.logConnectionFailed')
          );
        }
      };

      timer = window.setInterval(() => {
        if (Date.now() - heartbeat > 60e3) {
          resolveAndFinish('timeout', timeoutMessage);
        }
      }, 1000);
    });
  };

  const startInstall = async (hostId: number) => {
    try {
      const result = await installHostAgentApi(hostId);
      if (result.log_path) {
        isUpgrade.value = false;
        await logFileLogs(result.log_host, result.log_path);
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
        await logFileLogs(result.log_host, result.log_path, {
          successMessage: t('manage.host.installAgent.upgradeSuccess'),
          failedMessage: t('manage.host.installAgent.upgradeFailed'),
          timeoutMessage: t('manage.host.installAgent.upgradeTimeout'),
        });
      } else {
        Message.error(t('manage.host.installAgent.upgradeFailed'));
      }
    } catch (error) {
      Message.error(t('manage.host.installAgent.upgradeFailed'));
    }
  };

  const startBatchUpgrade = async (hosts: BatchUpgradeHost[]) => {
    if (!hosts.length) {
      Message.info(t('manage.host.list.batchUpgrade.empty'));
      return;
    }

    reset();
    show();
    isUpgrade.value = true;
    isBatchUpgrade.value = true;
    batchRunning.value = true;

    let successCount = 0;
    let failedCount = 0;

    addLog({
      time: Date.now(),
      message: t('manage.host.list.batchUpgrade.start', {
        count: hosts.length,
      }),
      level: 'info',
    });

    /* eslint-disable no-await-in-loop */
    for (const host of hosts) {
      setStatus('installing');
      addLog({
        time: Date.now(),
        message: t('manage.host.list.batchUpgrade.processing', {
          name: host.name,
        }),
        level: 'info',
      });

      try {
        const result = await upgradeHostAgentApi(host.id);
        if (!result.log_path) {
          failedCount += 1;
          addLog({
            time: Date.now(),
            message: t('manage.host.list.batchUpgrade.itemFailed', {
              name: host.name,
            }),
            level: 'error',
          });
          continue;
        }

        const finalStatus = await logFileLogs(
          result.log_host,
          result.log_path,
          {
            resetBefore: false,
            prefix: host.name,
            successMessage: '',
            failedMessage: '',
            timeoutMessage: '',
            emitOk: false,
          }
        );

        if (finalStatus === 'completed') {
          successCount += 1;
          addLog({
            time: Date.now(),
            message: t('manage.host.list.batchUpgrade.itemSuccess', {
              name: host.name,
            }),
            level: 'info',
          });
        } else {
          failedCount += 1;
          addLog({
            time: Date.now(),
            message: t('manage.host.list.batchUpgrade.itemFailed', {
              name: host.name,
            }),
            level: 'error',
          });
        }
      } catch (error) {
        failedCount += 1;
        addLog({
          time: Date.now(),
          message: t('manage.host.list.batchUpgrade.itemFailed', {
            name: host.name,
          }),
          level: 'error',
        });
      }
    }
    /* eslint-enable no-await-in-loop */

    batchRunning.value = false;
    setStatus(failedCount > 0 ? 'failed' : 'completed');
    addLog({
      time: Date.now(),
      message: t('manage.host.list.batchUpgrade.summary', {
        success: successCount,
        failed: failedCount,
      }),
      level: failedCount > 0 ? 'warn' : 'info',
    });
    emit('ok');

    if (failedCount > 0) {
      Message.warning(
        t('manage.host.list.batchUpgrade.doneWithFailure', {
          success: successCount,
          failed: failedCount,
        })
      );
    } else {
      Message.success(
        t('manage.host.list.batchUpgrade.done', { count: successCount })
      );
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
    startBatchUpgrade,
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
