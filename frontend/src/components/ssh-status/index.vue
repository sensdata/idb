<template>
  <div class="ssh-status">
    <div class="status-display">
      <a-space>
        <a-badge :status="statusBadge" />
        <a-tag :color="statusColor" size="medium">{{ sshStatusText }}</a-tag>
      </a-space>
    </div>
    <div class="status-actions">
      <a-button
        type="outline"
        size="small"
        :disabled="sshStatus !== 'running' && sshStatus !== 'unhealthy'"
        @click="stopSshServer"
      >
        {{ $t('app.ssh.status.stop') }}
      </a-button>
      <a-button
        type="outline"
        size="small"
        :disabled="sshStatus !== 'running' && sshStatus !== 'unhealthy'"
        @click="reloadSshServer"
      >
        {{ $t('app.ssh.status.reload') }}
      </a-button>
      <a-button
        type="primary"
        size="small"
        :disabled="sshStatus === 'starting' || sshStatus === 'stopping'"
        @click="restartSshServer"
      >
        {{ $t('app.ssh.status.restart') }}
      </a-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import zhCN from './locale/zh-CN';
  import enUS from './locale/en-US';

  // 注册国际化消息
  const messages = {
    'zh-CN': zhCN,
    'en-US': enUS,
  };

  const { t } = useI18n({ messages });

  // SSH 服务状态类型
  type SshStatusType =
    | 'running'
    | 'stopped'
    | 'starting'
    | 'stopping'
    | 'error'
    | 'unhealthy';
  const sshStatus = ref<SshStatusType>('stopped');

  const sshStatusText = computed(() => {
    switch (sshStatus.value) {
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
      default:
        return t('app.ssh.status.unknown');
    }
  });

  // 状态徽章和颜色
  const statusBadge = computed(() => {
    switch (sshStatus.value) {
      case 'running':
        return 'success';
      case 'stopped':
        return 'danger';
      case 'starting':
      case 'stopping':
        return 'warning';
      case 'error':
        return 'danger';
      case 'unhealthy':
        return 'warning';
      default:
        return 'normal';
    }
  });

  const statusColor = computed(() => {
    switch (sshStatus.value) {
      case 'running':
        return 'green';
      case 'stopped':
        return 'red';
      case 'starting':
      case 'stopping':
        return 'orange';
      case 'error':
        return 'red';
      case 'unhealthy':
        return 'orange';
      default:
        return 'gray';
    }
  });

  // SSH 服务控制方法
  const stopSshServer = async () => {
    try {
      sshStatus.value = 'stopping';
      // TODO: 实现停止SSH服务器的实际API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 1000);
      });
      sshStatus.value = 'stopped';
      Message.success(t('app.ssh.status.stopSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.status.stopFailed'));
      sshStatus.value = 'running';
    }
  };

  const reloadSshServer = async () => {
    try {
      // TODO: 实现重载SSH服务器的实际API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 1000);
      });
      Message.success(t('app.ssh.status.reloadSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.status.reloadFailed'));
    }
  };

  const restartSshServer = async () => {
    try {
      sshStatus.value = 'stopping';
      // TODO: 实现重启SSH服务器的实际API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 500);
      });
      sshStatus.value = 'starting';
      await new Promise((resolve) => {
        setTimeout(resolve, 1000);
      });
      sshStatus.value = 'running';
      Message.success(t('app.ssh.status.restartSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.status.restartFailed'));
      // 尝试确定当前状态或默认为已停止
      sshStatus.value = 'stopped';
    }
  };

  // 组件挂载时检查SSH服务器的初始状态
  onMounted(async () => {
    try {
      // TODO: 实现获取SSH服务器状态的实际API调用
      await new Promise((resolve) => {
        setTimeout(resolve, 500);
      });
      sshStatus.value = 'running';
    } catch (error) {
      sshStatus.value = 'stopped';
    }
  });
</script>

<style scoped lang="less">
  .ssh-status {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-right: 90px;

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
        padding: 0 16px;
        height: 32px;
        font-size: 14px;
      }
    }
  }
</style>
