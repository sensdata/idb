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
  import { TASK_STATUS } from '@/config/enum';
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

  const logTaskMsgs = (taskId: string) => {
    const url = resolveApiUrl(`tasks/${taskId}/logs`);
    reset();
    show();

    let heartbeat = Date.now();
    const eventSource = new EventSource(url);

    // 处理日志事件
    eventSource.addEventListener('log', (event: Event) => {
      try {
        if (event instanceof MessageEvent) {
          const logData = JSON.parse(event.data);
          if (
            logData &&
            typeof logData === 'object' &&
            logData.level &&
            logData.message
          ) {
            // 如果日志数据包含级别和消息
            const level = logData.level.toLowerCase();
            // 将level映射到组件支持的级别
            let mappedLevel: 'info' | 'error' | 'warn' | 'debug' = 'info';

            if (level === 'debug') {
              mappedLevel = 'debug';
            } else if (level === 'info') {
              mappedLevel = 'info';
            } else if (level === 'warn') {
              mappedLevel = 'warn';
            } else if (level === 'error') {
              mappedLevel = 'error';
            }

            addLog({
              time: logData.timestamp,
              message: logData.message,
              level: mappedLevel,
            });
          } else {
            // 如果是普通字符串，默认为info级别
            addLog({
              time: Date.now(),
              message: event.data,
              level: 'info',
            });
          }
        }
      } catch (e) {
        // 如果解析JSON失败，则按原样显示为info级别
        if (event instanceof MessageEvent) {
          addLog({
            time: Date.now(),
            message: event.data,
            level: 'info',
          });
        }
      }
    });

    eventSource.addEventListener('heartbeat', () => {
      heartbeat = Date.now();
    });

    const timer = window.setInterval(() => {
      if (Date.now() - heartbeat > 60e3) {
        clearInterval(timer);
        eventSource.close();
        setStatus('timeout');
        Message.error(t('manage.host.installAgent.installTimeout'));
      }
    }, 1000);

    eventSource.addEventListener('status', (event: Event) => {
      if (event instanceof MessageEvent) {
        switch (event.data) {
          case TASK_STATUS.Success:
            clearInterval(timer);
            eventSource.close();
            setStatus('completed');
            Message.success(t('manage.host.installAgent.installSuccess'));
            break;
          case TASK_STATUS.Failed:
          case TASK_STATUS.Canceled:
            clearInterval(timer);
            eventSource.close();
            setStatus('failed');
            Message.error(t('manage.host.installAgent.installFailed'));
            break;
          default:
            break;
        }
      }
    });

    // 全局错误处理
    eventSource.addEventListener('error', () => {
      clearInterval(timer);
      eventSource.close();
      setStatus('failed');
      Message.error(t('manage.host.installAgent.installFailed'));
    });
  };

  const startInstall = async (hostId: number) => {
    try {
      const result = await installHostAgentApi(hostId);
      if (result.task_id) {
        isUpgrade.value = false;
        logTaskMsgs(result.task_id);
      } else {
        Message.error(t('manage.host.installAgent.installFailed'));
      }
    } catch (error) {
      Message.error(t('manage.host.installAgent.installFailed'));
      console.error('Failed to install agent:', error);
    }
  };

  const startUpgrade = async (hostId: number) => {
    try {
      const result = await upgradeHostAgentApi(hostId);
      if (result.task_id) {
        isUpgrade.value = true;
        logTaskMsgs(result.task_id);
      } else {
        Message.error(t('manage.host.installAgent.upgradeFailed'));
      }
    } catch (error) {
      Message.error(t('manage.host.installAgent.upgradeFailed'));
      console.error('Failed to upgrade agent:', error);
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
      console.error('Failed to check agent:', error);
    }

    return false;
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
    logTaskMsgs,
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
