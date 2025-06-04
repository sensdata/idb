<template>
  <a-drawer
    v-model:visible="visible"
    :title="$t('app.service.logs.title')"
    :width="800"
    unmount-on-close
    @cancel="handleDrawerClose"
  >
    <div class="logs-content">
      <div class="logs-header">
        <a-space>
          <a-button :loading="isConnecting" @click="handleRefresh">
            <template #icon>
              <icon-refresh />
            </template>
            {{ $t('common.refresh') }}
          </a-button>
          <a-switch v-model="autoRefresh" @change="handleAutoRefreshChange">
            <template #checked>{{
              $t('app.service.logs.auto_refresh')
            }}</template>
            <template #unchecked>{{ $t('app.service.logs.manual') }}</template>
          </a-switch>
        </a-space>
      </div>

      <div class="logs-viewer">
        <!-- 连接中状态 -->
        <div v-if="isConnecting" class="status-container">
          <a-spin :size="24" />
          <p class="status-message">{{ $t('app.service.logs.connecting') }}</p>
        </div>

        <!-- 连接错误状态 -->
        <div v-else-if="hasConnectionError" class="status-container">
          <a-result
            status="warning"
            :title="$t('app.service.logs.connection_error')"
          >
            <template #subtitle>
              {{ connectionErrorMessage }}
            </template>
            <template #extra>
              <a-button type="primary" @click="handleRefresh">
                {{ $t('app.service.logs.retry') }}
              </a-button>
            </template>
          </a-result>
        </div>

        <!-- 正常日志显示（包括无日志的情况） -->
        <div v-else class="log-container">
          <pre class="log-content">{{ logContent }}</pre>
          <!-- 当没有实际日志数据时的提示 -->
          <div v-if="showNoLogsHint" class="no-logs-hint">
            <a-divider />
            <div class="hint-content">
              <icon-info-circle class="hint-icon" />
              <span>{{ $t('app.service.logs.no_entries_hint') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, onUnmounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { SERVICE_TYPE } from '@/config/enum';
  import { resolveApiUrl } from '@/helper/api-helper';
  import { ServiceLogStreamParams } from '@/api/service';

  interface LogsParams {
    type: SERVICE_TYPE;
    category: string;
    name: string;
  }

  const { t } = useI18n();

  const visible = ref(false);
  const autoRefresh = ref(false);
  const logContent = ref('');
  const currentParams = ref<LogsParams>();
  const isConnecting = ref(false);
  const hasConnectionError = ref(false);
  const connectionErrorMessage = ref('');
  const showNoLogsHint = ref(false);
  const hasReceivedData = ref(false);
  const connectionTimeout = ref<number | undefined>();

  let eventSource: EventSource | null = null;

  const resetState = () => {
    logContent.value = '';
    isConnecting.value = false;
    hasConnectionError.value = false;
    connectionErrorMessage.value = '';
    showNoLogsHint.value = false;
    hasReceivedData.value = false;

    if (connectionTimeout.value) {
      clearTimeout(connectionTimeout.value);
      connectionTimeout.value = undefined;
    }
  };

  const loadLogs = () => {
    if (!currentParams.value) return;

    resetState();
    isConnecting.value = true;

    // 关闭之前的连接
    if (eventSource) {
      eventSource.close();
    }

    // 构建SSE URL
    const streamParams: ServiceLogStreamParams = {
      type: currentParams.value.type,
      category: currentParams.value.category,
      name: currentParams.value.name,
      follow: autoRefresh.value,
      tail: 100, // 默认显示最后100行
    };

    const url = resolveApiUrl('services/{host}/logs/tail', streamParams);

    // 创建EventSource连接
    eventSource = new EventSource(url);

    // 设置连接超时
    connectionTimeout.value = window.setTimeout(() => {
      if (isConnecting.value) {
        hasConnectionError.value = true;
        connectionErrorMessage.value = t('app.service.logs.connection_timeout');
        isConnecting.value = false;
      }
    }, 10000); // 10秒超时

    eventSource.addEventListener('open', () => {
      isConnecting.value = false;
      hasConnectionError.value = false;

      if (connectionTimeout.value) {
        clearTimeout(connectionTimeout.value);
        connectionTimeout.value = undefined;
      }

      logContent.value += `[${new Date().toLocaleString()}] ${t(
        'app.service.logs.connected',
        { name: currentParams.value?.name }
      )}\n`;

      // 设置一个延迟检查，如果连接后5秒内没有收到任何日志数据，显示无日志提示
      setTimeout(() => {
        if (!hasReceivedData.value && !hasConnectionError.value) {
          logContent.value += `[${new Date().toLocaleString()}] ${t(
            'app.service.logs.no_entries_info'
          )}\n`;
          showNoLogsHint.value = true;
        }
      }, 5000); // 延长到5秒，给更多时间等待日志
    });

    eventSource.addEventListener('log', (event: Event) => {
      if (event instanceof MessageEvent) {
        if (event.data) {
          hasReceivedData.value = true;
          showNoLogsHint.value = false; // 有数据时隐藏提示

          try {
            const rawData = event.data.trim();

            // 处理日志数据
            if (rawData.startsWith('{') && rawData.endsWith('}')) {
              const logData = JSON.parse(rawData);
              logContent.value += `[${new Date(
                logData.timestamp || Date.now()
              ).toLocaleString()}] ${logData.message || rawData}\n`;
            } else {
              logContent.value += `[${new Date().toLocaleString()}] ${
                event.data
              }\n`;
            }
          } catch (error) {
            logContent.value += `[${new Date().toLocaleString()}] ${
              event.data
            }\n`;
          }
        }
      }
    });

    eventSource.addEventListener('status', (event: Event) => {
      if (event instanceof MessageEvent) {
        const status = event.data;

        if (status === 'close' || status === 'end') {
          // 正常关闭，如果没有接收到任何数据，显示无日志提示
          if (!hasReceivedData.value) {
            showNoLogsHint.value = true;
          }
        }
      }
    });

    eventSource.addEventListener('error', (event) => {
      console.error('SSE connection error:', event);
      isConnecting.value = false;

      if (connectionTimeout.value) {
        clearTimeout(connectionTimeout.value);
        connectionTimeout.value = undefined;
      }

      // 如果还没有接收到任何数据，且eventSource状态为关闭，很可能是因为没有日志
      if (
        !hasReceivedData.value &&
        eventSource?.readyState === EventSource.CLOSED
      ) {
        showNoLogsHint.value = true;
        logContent.value += `[${new Date().toLocaleString()}] ${t(
          'app.service.logs.no_entries_info'
        )}\n`;
      } else {
        hasConnectionError.value = true;
        connectionErrorMessage.value = t('app.service.logs.connection_failed');
      }
    });

    // 处理EventSource状态变化
    const checkConnectionState = () => {
      if (
        eventSource?.readyState === EventSource.CLOSED &&
        isConnecting.value
      ) {
        isConnecting.value = false;
        if (!hasReceivedData.value) {
          showNoLogsHint.value = true;
        }
      }
    };

    // 定期检查连接状态
    const stateCheckInterval = setInterval(() => {
      checkConnectionState();
      if (!eventSource || eventSource.readyState === EventSource.CLOSED) {
        clearInterval(stateCheckInterval);
      }
    }, 1000);
  };

  const show = (params: LogsParams) => {
    currentParams.value = params;
    visible.value = true;
    loadLogs();
  };

  const handleRefresh = () => {
    if (currentParams.value) {
      loadLogs();
    }
  };

  // 监听自动刷新开关变化
  const handleAutoRefreshChange = () => {
    if (currentParams.value) {
      loadLogs();
    }
  };

  // 抽屉关闭时也关闭连接
  const handleDrawerClose = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    resetState();
  };

  // 组件卸载时关闭连接
  onUnmounted(() => {
    if (eventSource) {
      eventSource.close();
    }
    if (connectionTimeout.value) {
      clearTimeout(connectionTimeout.value);
    }
  });

  defineExpose({
    show,
  });
</script>

<style scoped>
  .logs-content {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 120px);
  }

  .logs-header {
    padding-bottom: 16px;
    margin-bottom: 16px;
    border-bottom: 1px solid var(--color-border-2);
  }

  .logs-viewer {
    position: relative;
    flex: 1;
    overflow: hidden;
    background: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .log-container {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .log-content {
    flex: 1;
    padding: 16px;
    margin: 0;
    overflow: auto;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 14px;
    line-height: 1.5;
    color: var(--color-text-1);
    word-break: break-word;
    white-space: pre-wrap;
  }

  .status-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 200px;
    padding: 32px;
  }

  .status-message {
    margin-top: 16px;
    font-size: 14px;
    color: var(--color-text-2);
  }

  .no-logs-hint {
    padding: 16px;
    background: var(--color-fill-1);
    border-top: 1px solid var(--color-border-2);
  }

  .hint-content {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    color: var(--color-text-3);
  }

  .hint-icon {
    font-size: 16px;
    color: var(--color-text-4);
  }
</style>
