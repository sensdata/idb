<template>
  <div class="host-sidebar">
    <div class="sidebar-header">
      <div class="header-content">
        <icon-desktop class="header-icon" />
        <h3>{{ $t('components.terminal.workspace.hosts') }}</h3>
      </div>
      <div class="header-actions">
        <a-button
          type="text"
          size="small"
          :loading="loading"
          class="refresh-btn"
          :aria-label="$t('components.terminal.workspace.refreshHosts')"
          @click="refreshHosts"
        >
          <template #icon>
            <icon-refresh :spin="loading" />
          </template>
        </a-button>
        <a-button
          type="text"
          size="small"
          class="refresh-btn"
          :aria-label="$t('components.terminal.workspace.collapseSidebar')"
          @click="handleCollapse"
        >
          <template #icon>
            <icon-menu-fold />
          </template>
        </a-button>
      </div>
    </div>

    <div class="host-list" role="list">
      <div
        v-for="host in hosts"
        :key="host.id"
        class="host-item"
        :class="{ active: host.id === currentHostId }"
        role="listitem"
        tabindex="0"
        :aria-selected="host.id === currentHostId"
        @click="handleHostSelect(host)"
        @keydown.enter="handleHostSelect(host)"
        @keydown.space.prevent="handleHostSelect(host)"
      >
        <div class="host-avatar">
          <icon-desktop />
        </div>
        <div class="host-info">
          <div class="host-name">{{ host.name }}</div>
          <div class="host-addr">{{ host.addr }}</div>
        </div>
        <div class="host-status">
          <div
            class="status-dot"
            :class="hostStatusMap.get(host.id)?.class || 'unknown'"
            :title="hostStatusMap.get(host.id)?.tooltip || ''"
          ></div>
        </div>
      </div>
    </div>

    <div v-if="hosts.length === 0 && !loading" class="empty-state">
      <div class="empty-content">
        <icon-desktop class="empty-icon" />
        <p class="empty-text">
          {{ $t('components.terminal.workspace.noHosts') }}
        </p>
      </div>
    </div>

    <div v-if="error" class="error-state">
      <a-alert
        type="error"
        :message="$t('components.terminal.workspace.loadError')"
        :description="error"
        show-icon
        closable
        @close="clearError"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useHostStore } from '@/store';
  import { HostEntity } from '@/entity/Host';
  import { Message } from '@arco-design/web-vue';
  import { usePolling } from '@/composables/use-polling';
  import { useLogger } from '@/composables/use-logger';

  const STATUS_CLASSES = {
    ONLINE: 'online',
    OFFLINE: 'offline',
    UNKNOWN: 'unknown',
  } as const;

  const ONLINE_STATUSES = ['online', 'connected', 'installed'];
  const OFFLINE_STATUSES = ['offline', 'disconnected', 'not installed'];

  const { t } = useI18n();
  const hostStore = useHostStore();
  const { logError } = useLogger();

  interface Props {
    currentHostId?: number;
  }

  defineProps<Props>();

  const emit = defineEmits<{
    hostSelect: [host: HostEntity];
    collapse: [];
  }>();

  const handleCollapse = (): void => {
    emit('collapse');
  };

  const loading = ref(false);
  const error = ref('');

  const hosts = computed(() => hostStore.items);

  const getStatusClass = (host: HostEntity): string => {
    if (!host.agent_status) {
      return STATUS_CLASSES.UNKNOWN;
    }

    const connected = host.agent_status.connected?.toLowerCase();
    if (connected === 'online') {
      return STATUS_CLASSES.ONLINE;
    }
    if (connected === 'offline') {
      return STATUS_CLASSES.OFFLINE;
    }

    const status = host.agent_status.status?.toLowerCase();
    if (status) {
      if (ONLINE_STATUSES.includes(status)) {
        return connected === 'online'
          ? STATUS_CLASSES.ONLINE
          : STATUS_CLASSES.OFFLINE;
      }
      if (OFFLINE_STATUSES.includes(status)) {
        return STATUS_CLASSES.OFFLINE;
      }
    }

    return STATUS_CLASSES.UNKNOWN;
  };

  const getStatusTooltip = (host: HostEntity): string => {
    const statusClass = getStatusClass(host);
    switch (statusClass) {
      case STATUS_CLASSES.ONLINE:
        return t('components.terminal.workspace.statusOnline');
      case STATUS_CLASSES.OFFLINE:
        return t('components.terminal.workspace.statusOffline');
      default:
        return t('components.terminal.workspace.statusUnknown');
    }
  };

  // 优化：将状态计算逻辑提取为computed，避免在模板中重复计算
  const hostStatusMap = computed(() => {
    const map = new Map<number, { class: string; tooltip: string }>();
    hosts.value.forEach((host) => {
      const statusClass = getStatusClass(host);
      map.set(host.id, {
        class: statusClass,
        tooltip: getStatusTooltip(host),
      });
    });
    return map;
  });

  const handleHostSelect = (host: HostEntity): void => {
    emit('hostSelect', host);
  };

  const clearError = (): void => {
    error.value = '';
  };

  const refreshHosts = async (): Promise<void> => {
    if (loading.value) return;

    loading.value = true;
    error.value = '';

    try {
      await hostStore.load();
      Message.success(t('components.terminal.workspace.refreshSuccess'));
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      Message.error(t('components.terminal.workspace.refreshFailed'));
      logError('Failed to refresh hosts:', err);
    } finally {
      loading.value = false;
    }
  };

  // 自动刷新主机列表
  const { startPolling, stopPolling } = usePolling({
    pollingFunction: async () => {
      if (!loading.value) {
        await hostStore.load();
      }
    },
    interval: 30000,
    immediate: false,
    onError: (err) => {
      logError('Auto refresh failed:', err);
    },
  });

  onMounted(() => {
    // 始终在挂载时刷新一次，确保新增主机能及时显示
    refreshHosts();
    startPolling(1);
  });

  // 添加组件卸载时的清理逻辑
  onUnmounted(() => {
    stopPolling();
  });
</script>

<style scoped>
  .host-sidebar {
    display: flex;
    flex-direction: column;
    width: 216px; /* 调整宽度以适应200px的item尺寸 */
    height: 100%;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2); /* 添加完整边框 */
    border-radius: var(--border-radius-small);
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 16px 12px; /* 左右间距调整为16px - 符合设计图 */
    background: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2); /* 保留底部边框，用于分隔头部和列表 */
  }

  .header-content {
    display: flex;
    align-items: center;
  }

  .header-actions {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .header-icon {
    margin-right: 16px; /* 图标和文字之间16px间距 - 符合设计图 */
    font-size: 16px;
    color: var(--color-primary-6);
  }

  .sidebar-header h3 {
    margin: 0;
    font-weight: 600;
    color: var(--color-text-1);

    @apply text-base;
  }

  .refresh-btn {
    border-radius: 6px;
    transition: all 0.2s ease;
  }

  .refresh-btn:hover {
    background: var(--color-fill-2);
    transform: scale(1.05);
  }

  .host-list {
    flex: 1;
    padding: 8px;
    overflow-y: auto;
  }

  .host-item {
    position: relative;
    display: flex;
    align-items: center;
    padding: 3px 8px 7px 8px; /* 上padding:3px, 下padding:7px */
    margin-bottom: 4px;
    cursor: pointer;
    background: var(--color-bg-1);
    border: 1px solid transparent;
    border-radius: var(--border-radius-small);
    transition: all 0.2s ease;
  }

  .host-item::before {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 3px;
    content: '';
    background: transparent;
    transition: background 0.2s ease;
  }

  .host-item:hover {
    border-color: var(--color-border-3);
  }

  .host-item.active {
    background: var(--color-primary-light-1);
    border-color: var(--color-primary-light-3);
  }

  .host-item.active::before {
    background: var(--color-primary-6);
  }

  .host-avatar {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 24px; /* 减少avatar尺寸 */
    height: 24px; /* 减少avatar尺寸 */
    margin-right: 8px;
    color: var(--color-text-2);
    background: var(--color-fill-2);

    @apply text-sm;
  }

  .host-item.active .host-avatar {
    color: var(--color-primary-6);
    background: var(--color-primary-light-2);
  }

  .host-info {
    flex: 1;
    min-width: 0;
    margin: 0 12px 0 4px;
  }

  .host-name {
    margin-bottom: 1px; /* 减少间距 */
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 600;
    line-height: 1.2; /* 减少行高 */
    color: var(--color-text-1);
    white-space: nowrap;

    @apply text-base;
  }

  .host-addr {
    display: inline-block;
    max-width: 100%;
    padding: 2px 4px 1px 4px;
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    line-height: 1.1; /* 添加紧凑的行高 */
    color: var(--color-text-3);
    white-space: nowrap;
    background: var(--color-fill-1);

    @apply text-sm;
  }

  .host-status {
    flex-shrink: 0;
  }

  .status-dot {
    position: relative;
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .status-dot.online {
    background: #00b42a;
    box-shadow: 0 0 0 1px rgb(0 180 42 / 25%);
  }

  .status-dot.online::after {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 6px;
    height: 6px;
    content: '';
    background: rgb(0 180 42 / 25%);
    border-radius: 50%;
    transform: translate(-50%, -50%);
    animation: pulse 2s infinite;
  }

  .status-dot.offline {
    background: #f53f3f;
    box-shadow: 0 0 0 1.5px rgb(245 63 63 / 25%);
  }

  .status-dot.unknown {
    background: #86909c;
    box-shadow: 0 0 0 1.5px rgb(134 144 156 / 25%);
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 0.8;
      transform: translate(-50%, -50%) scale(1);
    }
    50% {
      opacity: 0.3;
      transform: translate(-50%, -50%) scale(1.8);
    }
  }

  .host-item:focus {
    outline: 2px solid var(--color-primary-6);
    outline-offset: -2px;
  }

  .empty-state {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    text-align: center;
  }

  .empty-icon {
    margin-bottom: 16px;
    font-size: 48px;
    color: var(--color-text-4);
    opacity: 0.6;
  }

  .empty-text {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--color-text-3);
  }

  .error-state {
    padding: 16px;
    margin: 12px;
  }

  @media (width <= 768px) {
    .host-sidebar {
      width: 200px;
    }
    .host-item {
      padding: 8px 6px;
    }
    .host-avatar {
      width: 24px;
      height: 24px;
      margin-right: 6px;
      font-size: 12px;
    }
    .host-name {
      font-size: 12px;
    }
    .host-addr {
      font-size: 10px;
    }
  }
</style>
