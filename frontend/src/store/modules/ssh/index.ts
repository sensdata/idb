import { defineStore } from 'pinia';
import { Message } from '@arco-design/web-vue';
import { getSSHConfig, operateSSH } from '@/api/ssh';
import { useLogger } from '@/utils/hooks/use-logger';
import { t } from '@/utils/i18n';
import { SSHState, SSHStatus } from './types';

const { logError } = useLogger();

// API状态与前端状态的映射
const STATUS_MAP: Record<string, SSHStatus> = {
  enable: 'running',
  running: 'running',
  disable: 'stopped',
  disabled: 'stopped',
  stopped: 'stopped',
  failed: 'error',
};

// 状态对应的徽章样式
const BADGE_STATUS_MAP: Record<SSHStatus, string> = {
  running: 'success',
  stopped: 'danger',
  starting: 'warning',
  stopping: 'warning',
  error: 'danger',
  unhealthy: 'warning',
};

// 状态对应的颜色
const COLOR_MAP: Record<SSHStatus, string> = {
  running: 'green',
  stopped: 'red',
  starting: 'orange',
  stopping: 'orange',
  error: 'red',
  unhealthy: 'orange',
};

// SSH服务管理状态仓库
const useSSHStore = defineStore('ssh', {
  state: (): SSHState => ({
    config: null,
    status: 'stopped',
    loading: null,
    hostId: null,
  }),

  getters: {
    isRunning(): boolean {
      return this.status === 'running';
    },
    isError(): boolean {
      return this.status === 'error';
    },
    isTransitioning(): boolean {
      return this.status === 'starting' || this.status === 'stopping';
    },
    statusBadge(): string {
      return BADGE_STATUS_MAP[this.status] || 'normal';
    },
    statusColor(): string {
      return COLOR_MAP[this.status] || 'gray';
    },
  },

  actions: {
    // 设置当前主机ID
    setHostId(hostId: number) {
      this.hostId = hostId;
    },

    // 获取SSH配置和状态
    async fetchStatus() {
      if (!this.hostId) return;

      try {
        const res = await getSSHConfig(this.hostId);
        this.config = res;

        if (res?.status) {
          this.mapStatusFromConfig(res.status);
        } else {
          logError('Invalid API response structure:', res);
          this.status = 'error';
        }
      } catch (error) {
        logError('Failed to fetch SSH status:', error);
        this.status = 'error';
      }
    },

    // 将API状态映射到前端状态
    mapStatusFromConfig(apiStatus: string) {
      if (!apiStatus) {
        this.status = 'stopped';
        return;
      }

      const status = apiStatus.toLowerCase();
      this.status = STATUS_MAP[status] || 'stopped';
    },

    // 重启SSH服务
    async restart() {
      if (!this.hostId) {
        logError('Cannot restart: hostId is not set');
        return;
      }

      await this.executeOperation('restart', 'restart', {
        starting: 'starting',
        success: 'app.ssh.status.restartSuccess',
        failure: 'app.ssh.status.restartFailed',
      });
    },

    // 停止SSH服务
    async stop() {
      if (!this.hostId) return;

      await this.executeOperation('stop', 'stop', {
        starting: 'stopping',
        success: 'app.ssh.status.stopSuccess',
        failure: 'app.ssh.status.stopFailed',
      });
    },

    // 重载SSH服务配置
    async reload() {
      if (!this.hostId) return;

      await this.executeOperation('reload', 'reload', {
        success: 'app.ssh.status.reloadSuccess',
        failure: 'app.ssh.status.reloadFailed',
      });
    },

    // 统一处理SSH操作的方法
    async executeOperation(
      operation: 'enable' | 'disable' | 'stop' | 'reload' | 'restart',
      loadingState: string,
      messages: {
        starting?: SSHStatus;
        success: string;
        failure: string;
      }
    ) {
      if (!this.hostId) return;

      try {
        this.loading = loadingState;

        // 设置初始状态
        if (messages.starting) {
          this.status = messages.starting;
        }

        // 执行操作
        await operateSSH(this.hostId, operation);

        // 刷新状态
        await this.fetchStatus();

        // 显示成功消息
        Message.success(t(messages.success));
      } catch (error) {
        logError(`Failed to ${operation} SSH server:`, error);
        Message.error(t(messages.failure));
        await this.fetchStatus();
      } finally {
        this.loading = null;
      }
    },
  },
});

export default useSSHStore;
