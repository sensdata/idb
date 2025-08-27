<template>
  <div class="config-content">
    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.port.label') }} <span class="config-key">(Port)</span>
      </div>
      <div class="content with-input-group">
        <div class="input-group">
          <a-input
            :model-value="sshConfig.port"
            placeholder="22"
            disabled
            class="short-input"
          />
          <div class="actions">
            <a-button
              type="text"
              class="setting-btn"
              :loading="loadingStates.port"
              :disabled="loadingStates.port"
              @click.stop="openSetting('port')"
            >
              <icon-settings />
              <span>{{ $t('app.ssh.btn.setting') }}</span>
            </a-button>
          </div>
        </div>
        <div class="description">{{ $t('app.ssh.port.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.listen.label') }}
        <span class="config-key">(ListenAddress)</span>
      </div>
      <div class="content with-input-group">
        <div class="input-group">
          <a-input
            :model-value="sshConfig.listenAddress"
            placeholder="0.0.0.0"
            disabled
            class="short-input"
          />
          <div class="actions">
            <a-button
              type="text"
              class="setting-btn"
              :loading="loadingStates.listen"
              :disabled="loadingStates.listen"
              @click.stop="openSetting('listen')"
            >
              <icon-settings />
              <span>{{ $t('app.ssh.btn.setting') }}</span>
            </a-button>
          </div>
        </div>
        <div class="description">
          {{ $t('app.ssh.listen.description') }}
        </div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.rootModal.label') }}
        <span class="config-key">(PermitRootLogin)</span>
      </div>
      <div class="content">
        <a-switch
          :model-value="rootEnabled"
          :loading="loadingStates.root"
          :disabled="loadingStates.root"
          @change="handleRootEnabledChange"
        />
        <div class="description">
          {{ $t('app.ssh.rootModal.description') }}
        </div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.password.label') }}
        <span class="config-key">(PasswordAuthentication)</span>
      </div>
      <div class="content">
        <a-switch
          :model-value="sshConfig.passwordAuth"
          :loading="loadingStates.passwordAuth"
          :disabled="loadingStates.passwordAuth"
          @change="handlePasswordAuthChange"
        />
        <div class="description">
          {{ $t('app.ssh.password.description') }}
        </div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.key.label') }}
        <span class="config-key">(PubkeyAuthentication)</span>
      </div>
      <div class="content">
        <a-switch
          :model-value="sshConfig.keyAuth"
          :loading="loadingStates.keyAuth"
          :disabled="loadingStates.keyAuth"
          @change="handleKeyAuthChange"
        />
        <div class="description">{{ $t('app.ssh.key.description') }}</div>
      </div>
    </div>

    <div class="form-item">
      <div class="label">
        {{ $t('app.ssh.reverse.label') }}
        <span class="config-key">(UseDNS)</span>
      </div>
      <div class="content">
        <a-switch
          :model-value="sshConfig.reverseLookup"
          :loading="loadingStates.reverseLookup"
          :disabled="loadingStates.reverseLookup"
          @change="handleReverseLookupChange"
        />
        <div class="description">
          {{ $t('app.ssh.reverse.description') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { IconSettings } from '@arco-design/web-vue/es/icon';
  import type { SSHConfig, LoadingStates } from '../types';

  defineProps<{
    sshConfig: SSHConfig;
    loadingStates: LoadingStates;
    rootEnabled: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'openSetting', type: 'port' | 'listen'): void;
    (e: 'updatePasswordAuth', value: boolean): void;
    (e: 'updateKeyAuth', value: boolean): void;
    (e: 'updateReverseLookup', value: boolean): void;
    (e: 'saveRoot', enabled: boolean): void;
  }>();

  /** 打开设置弹窗 */
  const openSetting = (type: 'port' | 'listen') => {
    emit('openSetting', type);
  };

  /** 切换 root 登录开关 */
  const handleRootEnabledChange = (value: string | number | boolean) => {
    emit('saveRoot', Boolean(value));
  };

  /** 切换密码认证开关 */
  const handlePasswordAuthChange = (value: string | number | boolean) => {
    emit('updatePasswordAuth', Boolean(value));
  };

  /** 切换密钥认证开关 */
  const handleKeyAuthChange = (value: string | number | boolean) => {
    emit('updateKeyAuth', Boolean(value));
  };

  /** 切换反向 DNS 查询开关 */
  const handleReverseLookupChange = (value: string | number | boolean) => {
    emit('updateReverseLookup', Boolean(value));
  };
</script>

<style scoped lang="less">
  .config-content {
    margin-top: 8px;
    max-width: 900px;
  }

  .form-item {
    display: flex;
    align-items: flex-start;
    margin-bottom: 24px;

    .label {
      flex-shrink: 0;
      width: 220px;
      margin-right: 24px;
      color: var(--color-text-1);
      font-weight: 500;
      line-height: 32px;
      text-align: left;
      overflow: visible;
      white-space: normal;

      .config-key {
        font-size: 12px;
        font-weight: normal;
        color: var(--color-text-3);
        display: inline;
      }
    }

    .content {
      display: flex;
      flex: 1;
      max-width: 650px;

      &.with-input-group {
        flex-direction: column;

        .input-group {
          display: flex;
          align-items: center;
          margin-bottom: 4px;

          .short-input {
            width: 350px;
          }
        }

        .description {
          color: var(--color-text-3);
          font-size: 12px;
          max-width: 600px;
          line-height: 1.5;
        }
      }

      &:not(.with-input-group) {
        align-items: center;

        .description {
          margin-left: 16px;
          color: var(--color-text-3);
          font-size: 12px;
          max-width: 550px;
          line-height: 1.5;
        }
      }
    }

    .actions {
      display: flex;
      flex-shrink: 0;
      align-items: center;
      justify-content: center;
      height: 32px;
      margin-left: 16px;
      text-align: center;

      .setting-btn {
        padding: 0 8px;
        color: rgb(var(--primary-6));
        background: none;
        border: none;
        box-shadow: none;

        &:hover {
          color: rgb(var(--primary-5));
          background: none;
        }
      }
    }
  }
</style>
