<template>
  <a-drawer
    :visible="visible"
    :title="$t('app.rsync.client.logs.title')"
    :width="900"
    unmount-on-close
    @cancel="handleClose"
  >
    <div class="logs-wrapper">
      <!-- 日志文件列表 -->
      <div class="logs-sidebar">
        <div class="sidebar-header">
          {{ $t('app.rsync.client.logs.fileList') }}
        </div>
        <a-spin :loading="loading" class="sidebar-content">
          <a-empty v-if="!loading && logList.length === 0" />
          <div v-else class="log-file-list">
            <div
              v-for="log in logList"
              :key="log.path"
              class="log-file-item"
              :class="{ active: selectedLog?.path === log.path }"
              @click="handleSelectLog(log)"
            >
              <icon-file class="log-icon" />
              <span class="log-name">{{ getLogFileName(log.path) }}</span>
            </div>
          </div>
          <div v-if="logList.length > 0" class="pagination-wrapper">
            <a-pagination
              :current="pagination.current"
              :page-size="pagination.pageSize"
              :total="pagination.total"
              size="mini"
              simple
              @change="handlePageChange"
            />
          </div>
        </a-spin>
      </div>

      <!-- 日志内容 -->
      <div class="logs-content">
        <div class="content-header">
          <span v-if="selectedLog" class="selected-log-path">
            {{ selectedLog.path }}
          </span>
          <span v-else class="no-selection">
            {{ $t('app.rsync.client.logs.selectFile') }}
          </span>
        </div>
        <div class="content-viewer">
          <a-spin v-if="isLoading" class="connecting-spin">
            <template #tip>
              {{ $t('app.rsync.client.logs.loading') }}
            </template>
          </a-spin>
          <a-empty
            v-else-if="!selectedLog"
            :description="$t('app.rsync.client.logs.selectFile')"
          />
          <pre v-else ref="logContentRef" class="log-content">{{
            logContent
          }}</pre>
        </div>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch, onUnmounted, nextTick } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { getRsyncClientTaskLogListApi } from '@/api/database';
  import { getFileDetailApi } from '@/api/file';
  import { RsyncTaskLog } from '@/entity/Database';
  import { resolveApiUrl } from '@/helper/api-helper';

  interface Props {
    visible: boolean;
    taskId: string;
    taskState?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    taskId: '',
    taskState: '',
  });

  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  const { t } = useI18n();
  const loading = ref(false);
  const logList = ref<RsyncTaskLog[]>([]);
  const selectedLog = ref<RsyncTaskLog | null>(null);
  const logContent = ref('');
  const logContentRef = ref<HTMLElement | null>(null);
  const isLoading = ref(false);

  // 使用递增的 session ID 来标识每次加载
  let currentSessionId = 0;
  let eventSource: EventSource | null = null;

  const pagination = ref({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const getLogFileName = (path: string) => {
    return path.split('/').pop() || path;
  };

  const scrollToBottom = () => {
    nextTick(() => {
      if (logContentRef.value) {
        logContentRef.value.scrollTop = logContentRef.value.scrollHeight;
      }
    });
  };

  const closeEventSource = () => {
    if (eventSource) {
      try {
        eventSource.close();
      } catch (e) {
        // ignore
      }
      eventSource = null;
    }
  };

  const loadLogContent = async (logPath: string) => {
    // 1. 关闭旧的 SSE 连接
    closeEventSource();

    // 2. 清空内容
    logContent.value = '';
    isLoading.value = true;

    // 记录当前要加载的路径，用于验证
    const targetPath = logPath;

    // 根据任务状态决定策略：运行中任务只用 SSE，其它任务只用 API
    const isRunningTask = props.taskState === 'running';

    if (isRunningTask) {
      // 运行中任务：使用 SSE 从头开始跟踪日志
      try {
        const url = resolveApiUrl(
          `logs/{host}/follow?path=${encodeURIComponent(logPath)}&whence=start`
        );

        const es = new EventSource(url);
        eventSource = es;

        es.addEventListener('log', (event: MessageEvent) => {
          // 如果当前 eventSource 已经不是这个连接，忽略消息
          if (eventSource !== es) {
            es.close();
            return;
          }
          if (event.data) {
            // 第一次收到数据时结束 loading
            if (isLoading.value) {
              isLoading.value = false;
            }
            logContent.value += `${event.data}\n`;
            scrollToBottom();
          }
        });

        es.addEventListener('error', () => {
          if (eventSource !== es) {
            es.close();
          }
        });
      } catch (error) {
        if (selectedLog.value?.path === targetPath) {
          isLoading.value = false;
          Message.error(t('app.rsync.client.logs.loadFailed'));
        }
      }

      return;
    }

    // 非运行中任务：使用文件 API 读取完整内容，不再打开 SSE，避免重复
    try {
      const res = await getFileDetailApi({ path: logPath, expand: true });

      // 检查是否已经切换到其他日志
      if (selectedLog.value?.path !== targetPath) {
        return;
      }

      // 设置日志内容
      logContent.value = res.content || '';
      isLoading.value = false;
      scrollToBottom();
    } catch (error) {
      if (selectedLog.value?.path === targetPath) {
        isLoading.value = false;
        Message.error(t('app.rsync.client.logs.loadFailed'));
      }
    }
  };

  const handleSelectLog = (log: RsyncTaskLog) => {
    // 如果点击的是同一个日志，不重新加载
    if (selectedLog.value?.path === log.path) {
      return;
    }
    selectedLog.value = log;
    loadLogContent(log.path);
  };

  const fetchLogs = async () => {
    if (!props.taskId) return;

    loading.value = true;
    try {
      const res = await getRsyncClientTaskLogListApi({
        id: props.taskId,
        page: pagination.value.current,
        page_size: pagination.value.pageSize,
      });
      // 按文件名倒序排列，新的日志显示在上方
      const logs = res.logs || [];
      logs.sort((a, b) => {
        const nameA = a.path.split('/').pop() || '';
        const nameB = b.path.split('/').pop() || '';
        return nameB.localeCompare(nameA);
      });
      logList.value = logs;
      pagination.value.total = res.total || 0;

      // 默认选择第一个日志（最新的）
      if (logList.value.length > 0 && !selectedLog.value) {
        handleSelectLog(logList.value[0]);
      }
    } catch (error) {
      Message.error(t('app.rsync.client.logs.fetchFailed'));
    } finally {
      loading.value = false;
    }
  };

  const handlePageChange = (page: number) => {
    pagination.value.current = page;
    fetchLogs();
  };

  const handleClose = () => {
    currentSessionId += 1;
    closeEventSource();
    selectedLog.value = null;
    logContent.value = '';
    emit('update:visible', false);
  };

  watch(
    () => props.visible,
    (val) => {
      if (val && props.taskId) {
        pagination.value.current = 1;
        selectedLog.value = null;
        logContent.value = '';
        currentSessionId += 1;
        closeEventSource();
        fetchLogs();
      } else if (!val) {
        // 关闭抽屉时清理 SSE 连接
        currentSessionId += 1;
        closeEventSource();
      }
    }
  );

  onUnmounted(() => {
    currentSessionId += 1;
    closeEventSource();
  });
</script>

<style scoped lang="less">
  .logs-wrapper {
    display: flex;
    height: calc(100vh - 8.57rem);
    gap: 1.14rem;
  }

  .logs-sidebar {
    width: 20rem;
    flex-shrink: 0;
    border: 0.07rem solid var(--color-border-2);
    border-radius: 0.29rem;
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    padding: 0.86rem 1.14rem;
    font-weight: 500;
    border-bottom: 0.07rem solid var(--color-border-2);
    background: var(--color-fill-1);
  }

  .sidebar-content {
    flex: 1;
    overflow: auto;
    padding: 0.57rem;
  }

  .log-file-list {
    display: flex;
    flex-direction: column;
    gap: 0.29rem;
  }

  .log-file-item {
    display: flex;
    align-items: center;
    gap: 0.57rem;
    padding: 0.57rem 0.86rem;
    border-radius: 0.29rem;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: var(--color-fill-2);
    }

    &.active {
      background: rgb(var(--primary-6));
      color: #fff;

      .log-icon {
        color: #fff;
      }
    }
  }

  .log-icon {
    color: var(--color-text-3);
    font-size: 1rem;
  }

  .log-name {
    font-size: 0.93rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pagination-wrapper {
    padding: 0.57rem 0;
    border-top: 0.07rem solid var(--color-border-2);
    margin-top: 0.57rem;
  }

  .logs-content {
    flex: 1;
    border: 0.07rem solid var(--color-border-2);
    border-radius: 0.29rem;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .content-header {
    padding: 0.86rem 1.14rem;
    border-bottom: 0.07rem solid var(--color-border-2);
    background: var(--color-fill-1);
    font-family: monospace;
    font-size: 0.86rem;
  }

  .selected-log-path {
    color: var(--color-text-2);
  }

  .no-selection {
    color: var(--color-text-3);
  }

  .content-viewer {
    flex: 1;
    overflow: auto;
    position: relative;
  }

  .connecting-spin {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .log-content {
    margin: 0;
    padding: 1.14rem;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.86rem;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--color-bg-2);
    min-height: 100%;
  }
</style>
