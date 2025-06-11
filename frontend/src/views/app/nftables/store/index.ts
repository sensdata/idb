import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { createLogger } from '@/utils/logger';

export interface NftablesConfig {
  auto_start: boolean;
  version: string;
  installed: boolean;
  configured: boolean;
}

export interface NftablesStatus {
  running: boolean;
  installed: boolean;
  version: string;
  rules_count: number;
  last_reload: string;
}

export type ServiceStatus =
  | 'running'
  | 'stopped'
  | 'starting'
  | 'stopping'
  | 'error'
  | 'unhealthy'
  | 'loading'
  | 'unknown';
export type LoadingType = 'start' | 'stop' | 'restart' | 'reload' | null;

const useNftablesStore = defineStore('nftables', () => {
  const { t } = useI18n();
  const logger = createLogger('NftablesStore');

  // 状态
  const hostId = ref<number>(0);
  const status = ref<ServiceStatus>('loading');
  const loading = ref<LoadingType>(null);
  const config = ref<NftablesConfig | null>(null);
  const statusInfo = ref<NftablesStatus | null>(null);

  // 计算属性
  const statusBadge = computed(() => {
    switch (status.value) {
      case 'running':
        return 'success';
      case 'stopped':
        return 'normal';
      case 'starting':
      case 'stopping':
        return 'processing';
      case 'error':
      case 'unhealthy':
        return 'danger';
      default:
        return 'normal';
    }
  });

  const statusColor = computed(() => {
    switch (status.value) {
      case 'running':
        return 'green';
      case 'stopped':
        return 'gray';
      case 'starting':
      case 'stopping':
        return 'blue';
      case 'error':
      case 'unhealthy':
        return 'red';
      default:
        return 'gray';
    }
  });

  const isTransitioning = computed(() => {
    return (
      status.value === 'starting' ||
      status.value === 'stopping' ||
      loading.value !== null
    );
  });

  // 方法
  const setHostId = (id: number) => {
    hostId.value = id;
  };

  const fetchStatus = async (): Promise<void> => {
    if (!hostId.value) return;

    try {
      // 模拟API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 500);
      });

      // 模拟状态数据
      const mockStatus: NftablesStatus = {
        running: Math.random() > 0.3,
        installed: Math.random() > 0.2,
        version: '1.0.4',
        rules_count: Math.floor(Math.random() * 20) + 5,
        last_reload: new Date().toISOString(),
      };

      statusInfo.value = mockStatus;
      status.value = mockStatus.running ? 'running' : 'stopped';
    } catch (error) {
      status.value = 'error';
      logger.logError('Failed to fetch NFTables status:', error);
    }
  };

  const fetchConfig = async (): Promise<void> => {
    if (!hostId.value) return;

    try {
      // 模拟API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 300);
      });

      // 模拟配置数据
      const mockConfig: NftablesConfig = {
        auto_start: Math.random() > 0.5,
        version: '1.0.4',
        installed: Math.random() > 0.2,
        configured: Math.random() > 0.4,
      };

      config.value = mockConfig;
    } catch (error) {
      logger.logError('Failed to fetch NFTables config:', error);
      Message.error(t('app.nftables.message.configLoadFailed'));
    }
  };

  const start = async (): Promise<void> => {
    if (!hostId.value || loading.value) return;

    try {
      loading.value = 'start';
      status.value = 'starting';

      // 模拟API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 2000);
      });

      status.value = 'running';
      Message.success(t('app.nftables.message.serviceStarted'));
    } catch (error) {
      status.value = 'error';
      Message.error(t('app.nftables.message.startFailed'));
      logger.logError('Failed to start NFTables:', error);
    } finally {
      loading.value = null;
    }
  };

  const stop = async (): Promise<void> => {
    if (!hostId.value || loading.value) return;

    try {
      loading.value = 'stop';
      status.value = 'stopping';

      // 模拟API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 1500);
      });

      status.value = 'stopped';
      Message.success(t('app.nftables.message.serviceStopped'));
    } catch (error) {
      status.value = 'error';
      Message.error(t('app.nftables.message.stopFailed'));
      logger.logError('Failed to stop NFTables:', error);
    } finally {
      loading.value = null;
    }
  };

  const restart = async (): Promise<void> => {
    if (!hostId.value || loading.value) return;

    try {
      loading.value = 'restart';
      status.value = 'stopping';

      // 模拟停止
      await new Promise((resolve) => {
        setTimeout(resolve, 1000);
      });

      status.value = 'starting';

      // 模拟启动
      await new Promise((resolve) => {
        setTimeout(resolve, 1500);
      });

      status.value = 'running';
      Message.success(t('app.nftables.message.serviceRestarted'));
    } catch (error) {
      status.value = 'error';
      Message.error(t('app.nftables.message.restartFailed'));
      logger.logError('Failed to restart NFTables:', error);
    } finally {
      loading.value = null;
    }
  };

  const reload = async (): Promise<void> => {
    if (!hostId.value || loading.value) return;

    try {
      loading.value = 'reload';

      // 模拟API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 1000);
      });

      Message.success(t('app.nftables.message.configReloaded'));

      // 重新获取状态
      await fetchStatus();
    } catch (error) {
      Message.error(t('app.nftables.message.reloadFailed'));
      logger.logError('Failed to reload NFTables:', error);
    } finally {
      loading.value = null;
    }
  };

  return {
    // 状态
    hostId,
    status,
    loading,
    config,
    statusInfo,

    // 计算属性
    statusBadge,
    statusColor,
    isTransitioning,

    // 方法
    setHostId,
    fetchStatus,
    fetchConfig,
    start,
    stop,
    restart,
    reload,
  };
});

export default useNftablesStore;
