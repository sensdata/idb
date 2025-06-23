<template>
  <div class="ssh-status">
    <div v-if="sshStore.status !== 'loading'" class="status-display">
      <a-space>
        <a-badge :status="sshStore.statusBadge as BadgeStatus" />
        <a-tag :color="sshStore.statusColor" size="medium">{{
          sshStatusText
        }}</a-tag>
        <a-tag
          v-if="sshStore.config"
          :color="sshStore.config.auto_start ? 'green' : 'gray'"
          size="medium"
        >
          {{ autoStartText }}
        </a-tag>
      </a-space>
    </div>
    <div class="status-actions">
      <a-button
        type="outline"
        size="small"
        :disabled="
          sshStore.status === 'loading' ||
          (sshStore.status !== 'running' && sshStore.status !== 'unhealthy')
        "
        :loading="sshStore.loading === 'stop'"
        @click="stopSshServer"
      >
        {{ $t('app.ssh.status.stop') }}
      </a-button>
      <a-button
        type="outline"
        size="small"
        :disabled="
          sshStore.status === 'loading' ||
          (sshStore.status !== 'running' && sshStore.status !== 'unhealthy')
        "
        :loading="sshStore.loading === 'reload'"
        @click="reloadSshServer"
      >
        {{ $t('app.ssh.status.reload') }}
      </a-button>
      <a-button
        type="primary"
        size="small"
        :disabled="sshStore.isTransitioning"
        :loading="sshStore.loading === 'restart'"
        @click="restartSshServer"
      >
        {{ $t('app.ssh.status.restart') }}
      </a-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, onUnmounted, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import useSSHStore from '@/views/app/ssh/store';
  import { useHostStore } from '@/store';
  import { useLogger } from '@/composables/use-logger';
  import { usePolling } from '@/composables/use-polling';

  type BadgeStatus = 'success' | 'warning' | 'danger' | 'normal' | 'processing';

  const POLLING_INTERVAL = 10000;
  const INITIAL_DELAY = 200;

  const { logDebug, logWarn, logError } = useLogger();

  const { t } = useI18n();
  const route = useRoute();
  const sshStore = useSSHStore();
  const hostStore = useHostStore();

  // 主机ID - 使用query.id而不是params.host
  const hostId = computed<number>(() => {
    return +(route.query.id || hostStore.currentId || 0);
  });

  // SSH状态文本
  const sshStatusText = computed<string>(() => {
    switch (sshStore.status) {
      case 'running':
        return t('app.ssh.status.running');
      case 'stopped':
        return t('app.ssh.status.stopped');
      case 'starting':
        return t('app.ssh.status.starting');
      case 'stopping':
        return t('app.ssh.status.stopping');
      case 'error':
        return t('app.ssh.status.error');
      case 'unhealthy':
        return t('app.ssh.status.unhealthy');
      case 'loading':
        return '';
      default:
        return t('app.ssh.status.unknown');
    }
  });

  // 自动启动状态文本
  const autoStartText = computed<string>(() => {
    if (!sshStore.config) return '';
    return sshStore.config.auto_start
      ? t('app.ssh.status.autoStartEnabled')
      : t('app.ssh.status.autoStartDisabled');
  });

  // SSH服务控制方法
  const stopSshServer = (): void => {
    sshStore.stop();
  };

  const reloadSshServer = (): void => {
    sshStore.reload();
  };

  const restartSshServer = async (): Promise<void> => {
    logDebug('Restart button clicked, current status:', sshStore.status);
    logDebug('Restart button clicked, host ID:', hostId.value);

    if (!hostId.value || hostId.value <= 0) {
      logError('Cannot restart: Invalid host ID', hostId.value);
      return;
    }

    sshStore.setHostId(hostId.value);
    logDebug('Calling sshStore.restart()');
    await sshStore.restart();
    logDebug('SSH restart completed successfully');
  };

  // 使用自定义轮询组合函数
  const { startPolling, stopPolling } = usePolling({
    pollingFunction: () => sshStore.fetchStatus(),
    interval: POLLING_INTERVAL,
    immediate: true,
    initialDelay: INITIAL_DELAY,
    onBeforeStart: (id: number) => {
      logDebug('Starting SSH status polling for host ID:', id);
      sshStore.setHostId(id);
    },
    onInvalidId: (id: number) => {
      logWarn('Cannot start polling: Invalid host ID', id);
    },
  });

  // 组件挂载时启动轮询
  onMounted(() => {
    logDebug('SSH status component mounted, host ID:', hostId.value);
    startPolling(hostId.value);
  });

  // 主机ID变化时更新
  watch(hostId, (newId, oldId) => {
    logDebug('Host ID changed from', oldId, 'to', newId);
    if (newId && newId > 0) {
      stopPolling();
      sshStore.setHostId(newId);
      startPolling(newId);
    } else {
      logWarn('Invalid host ID detected:', newId);
      stopPolling();
    }
  });

  // 组件卸载时停止轮询
  onUnmounted(() => {
    stopPolling();
  });
</script>

<style scoped lang="less">
  .ssh-status {
    display: flex;
    align-items: center;
    justify-content: flex-end;

    .status-display {
      margin-right: 24px;

      :deep(.arco-tag) {
        font-size: 14px;
        font-weight: 500;
      }
    }

    .status-actions {
      display: flex;
      gap: 16px;

      .arco-btn {
        min-width: 60px;
      }
    }
  }
</style>
