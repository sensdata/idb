import { defineStore } from 'pinia';
import { Message } from '@arco-design/web-vue';
import { getSSHConfig, operateSSH } from '@/api/ssh';
import { useLogger } from '@/composables/use-logger';
import { t } from '@/utils/i18n';
import type {
  SSHState,
  SSHStatus,
  SSHStoreConfig,
} from '@/views/app/ssh/types';
import {
  STATUS_MAP,
  BADGE_STATUS_MAP,
  COLOR_MAP,
  DEFAULT_FORM_CONFIG,
  CACHE_TTL,
} from './constants';

// 创建单例logger
const logger = useLogger('SSHStore');

// SSH服务管理状态仓库
const useSSHStore = defineStore('ssh', {
  state: (): SSHState => ({
    config: null,
    status: 'loading',
    loading: null,
    hostId: null,
    lastFetch: 0,
    formConfig: { ...DEFAULT_FORM_CONFIG },
    isConfigFetching: false,
    requestId: 0,
  }),

  getters: {
    isRunning(): boolean {
      return this.status === 'running';
    },
    isError(): boolean {
      return this.status === 'error';
    },
    isTransitioning(): boolean {
      return (
        this.status === 'starting' ||
        this.status === 'stopping' ||
        this.status === 'loading'
      );
    },
    statusBadge(): string {
      return BADGE_STATUS_MAP[this.status] || 'normal';
    },
    statusColor(): string {
      return COLOR_MAP[this.status] || 'gray';
    },
    // 添加缓存是否有效的判断
    isCacheValid(): boolean {
      return Date.now() - this.lastFetch < CACHE_TTL;
    },
  },

  actions: {
    // 设置当前主机ID
    setHostId(hostId: number) {
      this.hostId = hostId;
    },

    // 获取SSH配置和状态，统一在这里获取所有数据
    async fetchConfig() {
      if (!this.hostId) {
        return;
      }

      // 如果已经有请求在进行中，跳过
      if (this.isConfigFetching) {
        logger.log(
          `Config request ${this.requestId} already in progress, skipping duplicate request`
        );
        return;
      }

      // 如果缓存有效且不是处于过渡状态，直接返回
      if (this.isCacheValid && !this.isTransitioning) {
        logger.log(
          `Using cached SSH config, last fetch: ${new Date(
            this.lastFetch
          ).toISOString()}`
        );
        return;
      }

      // 设置请求锁
      this.isConfigFetching = true;
      this.requestId++;
      const currentRequestId = this.requestId;
      logger.log(`Starting config request ${currentRequestId}`);

      try {
        logger.log(
          `Fetching fresh SSH config for request ${currentRequestId}...`
        );
        const res = await getSSHConfig(this.hostId);

        this.config = res;
        this.lastFetch = Date.now();

        if (res?.status) {
          this.mapStatusFromConfig(res.status);
        } else {
          logger.logError('Invalid API response structure:', res);
          this.status = 'error';
        }

        // 解析API响应到表单配置
        this.parseFormConfig(res);
        logger.log(`Config request ${currentRequestId} completed successfully`);
      } catch (error) {
        logger.logError(
          `Failed to fetch SSH config for request ${currentRequestId}:`,
          error as string
        );
        this.status = 'error';
      } finally {
        // 释放请求锁
        this.isConfigFetching = false;
        logger.log(`Released lock for config request ${currentRequestId}`);
      }
    },

    // 将API响应解析为表单配置
    parseFormConfig(apiConfig: SSHStoreConfig | null) {
      if (!apiConfig) return;

      this.formConfig = {
        port: apiConfig.port || '22',
        listenAddress: apiConfig.listen_address || '0.0.0.0',
        permitRootLogin: apiConfig.permit_root_login || 'yes',
        passwordAuth: apiConfig.password_authentication === 'yes',
        keyAuth: apiConfig.pubkey_authentication === 'yes',
        reverseLookup: apiConfig.use_dns === 'yes',
      };
    },

    // 重置表单配置
    resetFormConfig() {
      this.formConfig = { ...DEFAULT_FORM_CONFIG };
    },

    // 重置整个store状态
    reset() {
      this.$reset();
    },

    // 获取SSH状态（向后兼容旧API）
    async fetchStatus() {
      await this.fetchConfig();
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

    // 强制刷新状态（操作后调用）
    async forceRefreshStatus() {
      // 重置缓存时间戳以强制刷新
      this.lastFetch = 0;
      await this.fetchConfig();
    },

    // 更新单个配置设置
    async updateConfigSetting(key: string, oldValue: string, newValue: string) {
      if (!this.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      try {
        const values = [{ key, old_value: oldValue, new_value: newValue }];

        await this.updateConfig(values);

        // 根据更新的设置，修改当前formConfig
        this.updateLocalFormConfig(key, newValue);
      } catch (error) {
        logger.logError(
          `Error updating SSH config for ${key}:`,
          error as string
        );
        throw error;
      }
    },

    // 更新本地表单配置
    updateLocalFormConfig(key: string, value: string) {
      switch (key) {
        case 'Port':
          this.formConfig.port = value;
          break;
        case 'ListenAddress':
          this.formConfig.listenAddress = value;
          break;
        case 'PermitRootLogin':
          this.formConfig.permitRootLogin = value;
          break;
        case 'PasswordAuthentication':
          this.formConfig.passwordAuth = value === 'yes';
          break;
        case 'PubkeyAuthentication':
          this.formConfig.keyAuth = value === 'yes';
          break;
        case 'UseDNS':
          this.formConfig.reverseLookup = value === 'yes';
          break;
        default:
          // 不做任何操作
          break;
      }
    },

    // 更新多个配置设置
    async updateConfig(
      values: Array<{ key: string; old_value: string; new_value: string }>
    ) {
      try {
        if (!this.hostId) {
          Message.error(t('app.ssh.error.noHost'));
          return;
        }

        // 调用API更新配置
        const { updateSSHConfig } = await import('@/api/ssh');
        await updateSSHConfig(this.hostId, values);

        // 更新后强制刷新配置
        await this.forceRefreshStatus();
      } catch (error) {
        logger.logError('Error updating SSH config:', error as string);
        throw error;
      }
    },

    // 更新端口配置
    async updatePort(newPort: string) {
      const oldPort = this.formConfig.port;
      await this.updateConfigSetting('Port', oldPort, newPort);
    },

    // 更新监听地址配置
    async updateListenAddress(newAddress: string) {
      const oldAddress = this.formConfig.listenAddress;
      await this.updateConfigSetting('ListenAddress', oldAddress, newAddress);
    },

    // 更新root登录配置
    async updateRootLogin(enabled: boolean) {
      const oldValue = this.formConfig.permitRootLogin;
      const newValue = enabled ? 'yes' : 'no';
      await this.updateConfigSetting('PermitRootLogin', oldValue, newValue);
    },

    // 更新密码认证配置
    async updatePasswordAuth(enabled: boolean) {
      const oldValue = this.formConfig.passwordAuth ? 'yes' : 'no';
      const newValue = enabled ? 'yes' : 'no';
      await this.updateConfigSetting(
        'PasswordAuthentication',
        oldValue,
        newValue
      );
    },

    // 更新密钥认证配置
    async updateKeyAuth(enabled: boolean) {
      const oldValue = this.formConfig.keyAuth ? 'yes' : 'no';
      const newValue = enabled ? 'yes' : 'no';
      await this.updateConfigSetting(
        'PubkeyAuthentication',
        oldValue,
        newValue
      );
    },

    // 更新反向查询配置
    async updateReverseLookup(enabled: boolean) {
      const oldValue = this.formConfig.reverseLookup ? 'yes' : 'no';
      const newValue = enabled ? 'yes' : 'no';
      await this.updateConfigSetting('UseDNS', oldValue, newValue);
    },

    // 重启SSH服务
    async restart() {
      if (!this.hostId) {
        logger.logError('Cannot restart: hostId is not set');
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

        // 强制刷新状态（不使用缓存）
        await this.forceRefreshStatus();

        // 显示成功消息
        Message.success(t(messages.success));
      } catch (error) {
        logger.logError(`Failed to ${operation} SSH server:`, error as string);
        Message.error(t(messages.failure));

        // 强制刷新状态（不使用缓存）
        await this.forceRefreshStatus();
      } finally {
        this.loading = null;
      }
    },
  },
});

export default useSSHStore;
