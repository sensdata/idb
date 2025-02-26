<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.agent.installTitle')"
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
          <span :class="['log-message', log.level]">{{ log.message }}</span>
        </div>
        <div v-if="logs.length === 0" class="empty-log">
          {{ $t('manage.host.agent.waitingForLogs') }}
        </div>
      </div>
      <div class="log-footer">
        <a-button v-if="isCompleted" type="primary" @click="handleCancel">
          {{ $t('manage.host.agent.close') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useVisible from '@/hooks/visible';

  const emit = defineEmits(['ok', 'cancel']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();

  interface LogItem {
    time: number;
    message: string;
    level: 'info' | 'error' | 'warning' | 'success';
  }

  const logs = ref<LogItem[]>([]);
  const status = ref<'installing' | 'completed' | 'failed'>('installing');
  const logContentRef = ref<HTMLElement | null>(null);

  const isCompleted = computed(() => {
    return status.value === 'completed' || status.value === 'failed';
  });

  const statusText = computed(() => {
    switch (status.value) {
      case 'installing':
        return t('manage.host.agent.statusInstalling');
      case 'completed':
        return t('manage.host.agent.statusCompleted');
      case 'failed':
        return t('manage.host.agent.statusFailed');
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

  const addLog = (message: string, level: LogItem['level'] = 'info') => {
    logs.value.push({
      time: Date.now(),
      message,
      level,
    });
  };

  const clearLogs = () => {
    logs.value = [];
  };

  const setStatus = (newStatus: 'installing' | 'completed' | 'failed') => {
    status.value = newStatus;
  };

  const handleCancel = () => {
    if (isCompleted.value) {
      hide();
      emit('cancel');
    }
  };

  const reset = () => {
    clearLogs();
    setStatus('installing');
  };

  defineExpose({
    show,
    hide,
    reset,
    addLog,
    setStatus,
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
    color: var(--color-text-1);
  }

  .log-message.error {
    color: var(--color-danger-light-4);
  }

  .log-message.warning {
    color: var(--color-warning-light-4);
  }

  .log-message.success {
    color: var(--color-success-light-4);
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
