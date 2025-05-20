<template>
  <div class="ssh-page-container">
    <div class="header-container">
      <h2 class="page-title">SSH 配置</h2>
      <ssh-status class="ssh-status-container" />
    </div>
    <div class="content-container">
      <!-- Mode Switch -->
      <div class="mode-switch">
        <a-radio-group v-model="configMode" type="button">
          <a-radio value="visual">{{ $t('app.ssh.mode.visual') }}</a-radio>
          <a-radio value="source">{{ $t('app.ssh.mode.source') }}</a-radio>
        </a-radio-group>
      </div>

      <!-- Visual Mode -->
      <div v-if="configMode === 'visual'" class="config-content">
        <div class="form-item">
          <div class="label">{{ $t('app.ssh.port.label') }}</div>
          <div class="content with-input-group">
            <div class="input-group">
              <a-input
                v-model="sshConfig.port"
                placeholder="22"
                disabled
                class="short-input"
              />
              <div class="actions">
                <a-button
                  type="text"
                  class="setting-btn"
                  @click="openSettingModal('port')"
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
          <div class="label">{{ $t('app.ssh.listen.label') }}</div>
          <div class="content with-input-group">
            <div class="input-group">
              <a-input
                v-model="sshConfig.listenAddress"
                placeholder="0.0.0.0"
                disabled
                class="short-input"
              />
              <div class="actions">
                <a-button
                  type="text"
                  class="setting-btn"
                  @click="openSettingModal('listen')"
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
          <div class="label">{{ $t('app.ssh.root.label') }}</div>
          <div class="content with-input-group">
            <div class="input-group">
              <a-input
                v-model="rootUserDisplay"
                :placeholder="t('app.ssh.rootModal.allow')"
                disabled
                class="short-input"
              />
              <div class="actions">
                <a-button
                  type="text"
                  class="setting-btn"
                  @click="openSettingModal('root')"
                >
                  <icon-settings />
                  <span>{{ $t('app.ssh.btn.setting') }}</span>
                </a-button>
              </div>
            </div>
            <div class="description">{{ $t('app.ssh.root.description') }}</div>
          </div>
        </div>

        <div class="form-item">
          <div class="label">{{ $t('app.ssh.password.label') }}</div>
          <div class="content">
            <a-switch v-model="sshConfig.passwordAuth" />
            <div class="description">
              {{ $t('app.ssh.password.description') }}
            </div>
          </div>
        </div>

        <div class="form-item">
          <div class="label">{{ $t('app.ssh.key.label') }}</div>
          <div class="content">
            <a-switch v-model="sshConfig.keyAuth" />
            <div class="description">{{ $t('app.ssh.key.description') }}</div>
          </div>
        </div>

        <div class="form-item">
          <div class="label">{{ $t('app.ssh.passwordInfo.label') }}</div>
          <div class="content">
            <div class="password-info-placeholder"></div>
          </div>
        </div>

        <div class="form-item">
          <div class="label">{{ $t('app.ssh.reverse.label') }}</div>
          <div class="content">
            <a-switch v-model="sshConfig.reverseLookup" />
            <div class="description">
              {{ $t('app.ssh.reverse.description') }}
            </div>
          </div>
        </div>

        <div class="form-item">
          <div class="label">{{ $t('app.ssh.autostart.label') }}</div>
          <div class="content">
            <a-switch v-model="sshConfig.autoStart" />
          </div>
        </div>
      </div>

      <!-- Source Text Mode -->
      <div v-else class="source-mode">
        <source-editor
          v-model:config="sourceConfig"
          :original-config="originalSourceConfig"
          @save="saveSourceConfig"
          @reset="resetSourceConfig"
          @update-form="updateFormFromSourceConfig"
        />
      </div>
    </div>

    <!-- Modals -->
    <port-settings-modal
      v-model:visible="portModalVisible"
      :port="sshConfig.port"
      @save="savePortSetting"
    />

    <listen-settings-modal
      v-model:visible="listenModalVisible"
      :address="sshConfig.listenAddress"
      @save="saveListenSetting"
    />

    <root-settings-modal
      v-model:visible="rootModalVisible"
      :enabled="rootEnabled"
      @save="saveRootSetting"
    />
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconSettings } from '@arco-design/web-vue/es/icon';
  import { Message, Modal } from '@arco-design/web-vue';
  import { getSSHConfigContent, updateSSHConfigContent } from '@/api/ssh';
  import useSSHStore from '../../store';
  import { useSSHConfig } from './hooks/use-ssh-config';
  import SourceEditor from './components/source-editor.vue';
  import PortSettingsModal from './components/port-settings-modal.vue';
  import ListenSettingsModal from './components/listen-settings-modal.vue';
  import RootSettingsModal from './components/root-settings-modal.vue';
  import SshStatus from '../../components/ssh-status/index.vue';

  const { t } = useI18n();
  const sshStore = useSSHStore();

  // 配置模式：visual（可视化）/ source（源文件）
  const configMode = ref('visual');

  // 源文件模式的配置文本
  const sourceConfig = ref('');
  const originalSourceConfig = ref('');

  // 弹窗显示状态
  const portModalVisible = ref(false);
  const listenModalVisible = ref(false);
  const rootModalVisible = ref(false);

  // 使用composable管理SSH配置
  const { sshConfig, fetchConfig, updateConfig } = useSSHConfig(
    sshStore.hostId || 0
  );

  // 计算属性
  const rootEnabled = computed(() => {
    return sshConfig.value.permitRootLogin === 'yes';
  });

  const rootUserDisplay = computed(() => {
    return rootEnabled.value
      ? t('app.ssh.rootModal.allow')
      : t('app.ssh.rootModal.deny');
  });

  // 从源文件配置更新表单数据
  const updateFormFromSourceConfig = () => {
    try {
      if (!sourceConfig.value || sourceConfig.value.trim() === '') {
        Message.warning(t('app.ssh.source.emptyConfig'));
        return;
      }

      const lines = sourceConfig.value.split('\n');
      let hasUpdates = false;

      // 解析每一行配置
      for (const line of lines) {
        const trimmedLine = line.trim();
        if (trimmedLine.startsWith('#') || trimmedLine === '') continue;

        try {
          if (trimmedLine.startsWith('Port ')) {
            sshConfig.value.port = trimmedLine.replace('Port ', '').trim();
            hasUpdates = true;
          } else if (trimmedLine.startsWith('ListenAddress ')) {
            sshConfig.value.listenAddress = trimmedLine
              .replace('ListenAddress ', '')
              .trim();
            hasUpdates = true;
          } else if (trimmedLine.startsWith('PermitRootLogin ')) {
            const value = trimmedLine
              .replace('PermitRootLogin ', '')
              .trim()
              .toLowerCase();
            sshConfig.value.permitRootLogin = value;
            hasUpdates = true;
          } else if (trimmedLine.startsWith('PasswordAuthentication ')) {
            const value = trimmedLine
              .replace('PasswordAuthentication ', '')
              .trim()
              .toLowerCase();
            sshConfig.value.passwordAuth = value === 'yes';
            hasUpdates = true;
          } else if (trimmedLine.startsWith('PubkeyAuthentication ')) {
            const value = trimmedLine
              .replace('PubkeyAuthentication ', '')
              .trim()
              .toLowerCase();
            sshConfig.value.keyAuth = value === 'yes';
            hasUpdates = true;
          } else if (trimmedLine.startsWith('UseDNS ')) {
            const value = trimmedLine
              .replace('UseDNS ', '')
              .trim()
              .toLowerCase();
            sshConfig.value.reverseLookup = value === 'yes';
            hasUpdates = true;
          }
        } catch (lineError) {
          console.error('Error parsing line:', trimmedLine, lineError);
          // Continue processing other lines
        }
      }

      if (hasUpdates) {
        Message.success(t('app.ssh.source.parseSuccess'));
      } else {
        Message.warning(t('app.ssh.source.noChanges'));
      }
    } catch (error) {
      console.error('Error parsing SSH config:', error);
      Message.error(t('app.ssh.source.parseError'));
    }
  };

  // 获取SSH配置
  const fetchSSHConfig = async () => {
    try {
      if (!sshStore.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      await fetchConfig();

      // 获取SSH配置文件内容
      const contentRes = await getSSHConfigContent(sshStore.hostId);
      // 保存原始配置文件内容
      sourceConfig.value = contentRes.content || '';
      originalSourceConfig.value = contentRes.content || '';
    } catch (error) {
      console.error('Error fetching SSH config:', error);
      Message.error(t('app.ssh.error.fetchFailed'));
    }
  };

  // 初始化数据
  onMounted(async () => {
    try {
      await fetchSSHConfig();
    } catch (error) {
      // 错误处理
      Message.error(t('app.ssh.error.fetchFailed'));
    }
  });

  // 监听模式切换
  watch(configMode, (newMode, oldMode) => {
    if (newMode === 'visual' && oldMode === 'source') {
      // 从源文件模式切换到表单模式前先提示用户
      Modal.confirm({
        title: t('app.ssh.mode.switchConfirmTitle'),
        content: t('app.ssh.mode.switchConfirmContent'),
        onOk: () => {
          // 用户确认切换，解析源文件更新表单
          updateFormFromSourceConfig();
        },
        onCancel: () => {
          // 用户取消切换，保持在源文件模式
          configMode.value = 'source';
        },
      });
    }
  });

  // 保存源文件配置
  const saveSourceConfig = async () => {
    try {
      if (!sshStore.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      // 从源文件更新表单数据
      updateFormFromSourceConfig();

      // 保存配置文件内容
      await updateSSHConfigContent(sshStore.hostId, sourceConfig.value);

      originalSourceConfig.value = sourceConfig.value;
      Message.success(t('app.ssh.source.saveSuccess'));
    } catch (error) {
      console.error('Error saving SSH config:', error);
      Message.error(t('app.ssh.source.saveError'));
    }
  };

  // 重置源文件配置
  const resetSourceConfig = () => {
    sourceConfig.value = originalSourceConfig.value;
    Message.info(t('app.ssh.source.resetSuccess'));
  };

  // 打开设置弹窗函数
  const openSettingModal = (type: 'port' | 'listen' | 'root') => {
    if (type === 'port') {
      portModalVisible.value = true;
    } else if (type === 'listen') {
      listenModalVisible.value = true;
    } else if (type === 'root') {
      rootModalVisible.value = true;
    }
  };

  // 保存各项设置的函数
  const savePortSetting = async (newPort: string) => {
    sshConfig.value.port = newPort;
    portModalVisible.value = false;

    // 保存配置
    if (sshStore.hostId) {
      try {
        await updateConfig();
        Message.success(t('app.ssh.portModal.saveSuccess'));
      } catch (error) {
        console.error('Error saving port setting:', error);
        Message.error(t('app.ssh.portModal.saveError'));
      }
    }
  };

  const saveListenSetting = async (newAddress: string) => {
    sshConfig.value.listenAddress = newAddress;
    listenModalVisible.value = false;

    // 保存配置
    if (sshStore.hostId) {
      try {
        await updateConfig();
        Message.success(t('app.ssh.listenModal.saveSuccess'));
      } catch (error) {
        console.error('Error saving listen setting:', error);
        Message.error(t('app.ssh.listenModal.saveError'));
      }
    }
  };

  const saveRootSetting = async (enabled: boolean) => {
    sshConfig.value.permitRootLogin = enabled ? 'yes' : 'no';
    rootModalVisible.value = false;

    // 保存配置
    if (sshStore.hostId) {
      try {
        await updateConfig();
        Message.success(t('app.ssh.rootModal.saveSuccess'));
      } catch (error) {
        console.error('Error saving root setting:', error);
        Message.error(t('app.ssh.rootModal.saveError'));
      }
    }
  };
</script>

<style scoped lang="less">
  .ssh-page-container {
    padding: 0 16px;
    background-color: var(--color-bg-2);
    border-radius: 6px;
    position: relative;
    border: 1px solid var(--color-border-2);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .header-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 0;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 500;
    color: var(--color-text-1);
    margin: 0;
  }

  .ssh-status-container {
    margin-bottom: 0;
  }

  .content-container {
    background-color: #fff;
    border-radius: 4px;
    border: 1px solid var(--color-border-2);
    padding: 16px 20px;
    margin-bottom: 16px;
  }

  .mode-switch {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 16px;
  }

  .config-content {
    margin-top: 8px;
  }

  .form-item {
    display: flex;
    align-items: flex-start;
    margin-bottom: 24px;

    .label {
      flex-shrink: 0;
      width: 100px;
      margin-right: 16px;
      color: var(--color-text-1);
      font-weight: 500;
      line-height: 32px;
      text-align: right;
    }

    .content {
      display: flex;
      flex: 1;

      &.with-input-group {
        flex-direction: column;

        .input-group {
          display: flex;
          align-items: center;
          margin-bottom: 4px;

          .short-input {
            width: 200px;
          }
        }

        .description {
          color: var(--color-text-3);
          font-size: 12px;
        }
      }

      &:not(.with-input-group) {
        align-items: center;

        .description {
          margin-left: 8px;
          color: var(--color-text-3);
          font-size: 12px;
        }
      }
    }

    .actions {
      display: flex;
      flex-shrink: 0;
      align-items: center;
      justify-content: center;
      height: 32px;
      margin-left: 8px;
      text-align: center;

      .setting-btn {
        padding: 0 8px;
        color: #8250df; /* 紫色按钮文字 */
        background: none;
        border: none;
        box-shadow: none;

        &:hover {
          color: #9e77e3; /* 悬停时的浅紫色 */
          background: none;
        }
      }
    }

    .password-info-placeholder {
      width: 200px;
      height: 32px;
      background-color: var(--color-fill-2);
      border-radius: 4px;
    }
  }

  // 源文件模式样式
  .source-mode {
    margin-top: 16px;
  }
</style>
