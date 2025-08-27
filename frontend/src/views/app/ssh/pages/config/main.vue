<template>
  <div class="ssh-page-container">
    <div class="header-container">
      <h2 class="page-title">{{ $t('app.ssh.config.title') }}</h2>
      <ssh-status class="ssh-status-container" />
    </div>

    <a-alert
      v-if="anyLoading && !pageLoading"
      type="info"
      class="operation-alert"
      banner
    >
      <template #icon><icon-loading /></template>
      {{ $t('app.ssh.savingChanges') }}
    </a-alert>

    <div v-if="pageLoading" class="loading-container">
      <a-spin :size="30" />
      <p class="loading-text">{{ $t('app.ssh.loading') }}</p>
    </div>

    <div v-else class="content-container">
      <mode-switcher :mode="configMode" @mode-change="tryChangeMode" />

      <visual-config-panel
        v-if="configMode === 'visual'"
        :sshConfig="sshConfig"
        :loadingStates="loadingStates"
        :rootEnabled="rootEnabled"
        :rootUserDisplay="rootUserDisplay"
        @open-setting="openSettingModal"
        @update-password-auth="handlePasswordAuthChange"
        @update-key-auth="handleKeyAuthChange"
        @update-reverse-lookup="handleReverseLookupChange"
        @save-root="saveRootSetting"
      />

      <source-config-panel
        v-else
        ref="sourceEditorRef"
        v-model:config="sourceConfig"
        :original-config="originalSourceConfig"
        :loading="loadingStates.sourceConfig"
        @save="handleSaveSourceConfig"
        @reset="resetSourceConfig"
        @update-form="updateFormFromSourceConfig"
      />
    </div>

    <setting-modals
      :portModalVisible="portModalVisible"
      :listenModalVisible="listenModalVisible"
      :rootModalVisible="rootModalVisible"
      :port="sshConfig.port"
      :listenAddress="sshConfig.listenAddress"
      :rootEnabled="rootEnabled"
      @update:port-modal-visible="portModalVisible = $event"
      @update:listen-modal-visible="listenModalVisible = $event"
      @update:root-modal-visible="rootModalVisible = $event"
      @save-port="savePortSetting"
      @save-listen="saveListenSetting"
      @save-root="saveRootSetting"
    />

    <unsaved-changes-modal
      :visible="unsavedChangesModalVisible"
      @cancel="handleUnsavedChangesCancel"
      @confirm="handleUnsavedChangesConfirm"
    />
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconLoading } from '@arco-design/web-vue/es/icon';
  import { Message } from '@arco-design/web-vue';
  import useSSHStore from '../../store';
  import SshStatus from '../../components/ssh-status/index.vue';
  import { useSSHConfig } from './composables/use-ssh-config';
  import { useLoadingState } from './composables/use-loading-state';
  import ModeSwitcher from './components/mode-switcher.vue';
  import VisualConfigPanel from './components/visual-config-panel.vue';
  import SourceConfigPanel from './components/source-config-panel.vue';
  import SettingModals from './components/setting-modals.vue';
  import UnsavedChangesModal from './components/unsaved-changes-modal.vue';
  import type { EditorRefType, ConfigMode } from './types';

  const { t } = useI18n();
  const sshStore = useSSHStore();

  const pageLoading = ref<boolean>(true);
  const dataInitialized = ref<boolean>(false);
  const configMode = ref<ConfigMode>('visual');

  const sourceEditorRef = ref<EditorRefType | null>(null);

  const portModalVisible = ref<boolean>(false);
  const listenModalVisible = ref<boolean>(false);
  const rootModalVisible = ref<boolean>(false);
  const unsavedChangesModalVisible = ref<boolean>(false);
  const pendingModeChange = ref<ConfigMode | null>(null);

  const { loadingStates, anyLoading } = useLoadingState();

  /* SSH配置管理 */
  const {
    sshConfig,
    sourceConfig,
    originalSourceConfig,
    fetchSSHConfigContent,
    updateFormFromSourceConfig,
    saveSourceConfig,
    resetSourceConfig,
    refreshSourceConfig,
  } = useSSHConfig(sshStore, loadingStates);

  /* 计算属性 */
  const rootEnabled = computed<boolean>(() => {
    return sshConfig.value.permitRootLogin === 'yes';
  });

  const rootUserDisplay = computed<string>(() => {
    return rootEnabled.value
      ? t('app.ssh.rootModal.allow')
      : t('app.ssh.rootModal.deny');
  });

  /* 初始化SSH数据 */
  const initializeSSHData = async (): Promise<void> => {
    if (dataInitialized.value || !sshStore.hostId) return;

    try {
      pageLoading.value = true;

      await Promise.all([sshStore.fetchConfig(), fetchSSHConfigContent()]);

      dataInitialized.value = true;
    } catch (error) {
      Message.error(t('app.ssh.error.fetchFailed'));
    } finally {
      pageLoading.value = false;
    }
  };

  watch(
    () => sshStore.hostId,
    async (newVal) => {
      if (newVal && !dataInitialized.value) {
        await initializeSSHData();
      }
    },
    { immediate: true }
  );

  /* 模式切换处理 */
  const tryChangeMode = (newMode: ConfigMode): void => {
    if (configMode.value === 'source' && newMode === 'visual') {
      if (sourceEditorRef.value?.checkUnsavedChanges()) {
        pendingModeChange.value = newMode;
        unsavedChangesModalVisible.value = true;
        return;
      }
    }

    if (
      newMode === 'source' &&
      (!sourceConfig.value || sourceConfig.value.trim() === '')
    ) {
      fetchSSHConfigContent().catch(() => {
        Message.error(t('app.ssh.error.fetchFailed'));
      });
    }

    configMode.value = newMode;
  };

  const handleUnsavedChangesConfirm = (): void => {
    unsavedChangesModalVisible.value = false;
    if (pendingModeChange.value) {
      configMode.value = pendingModeChange.value;
      pendingModeChange.value = null;
    }
  };

  const handleUnsavedChangesCancel = (): void => {
    unsavedChangesModalVisible.value = false;
    pendingModeChange.value = null;
  };

  /* 设置处理函数 */
  const savePortSetting = async (newPort: string): Promise<void> => {
    try {
      if (!sshStore.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      loadingStates.port = true;

      await sshStore.updatePort(newPort);
      await refreshSourceConfig();
      Message.success(t('app.ssh.portModal.saveSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.portModal.saveError'));
    } finally {
      loadingStates.port = false;
    }
  };

  const saveListenSetting = async (newAddress: string): Promise<void> => {
    try {
      if (!sshStore.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      loadingStates.listen = true;

      await sshStore.updateListenAddress(newAddress);
      await refreshSourceConfig();
      Message.success(t('app.ssh.listenModal.saveSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.listenModal.saveError'));
    } finally {
      loadingStates.listen = false;
    }
  };

  const saveRootSetting = async (enabled: boolean): Promise<void> => {
    try {
      if (!sshStore.hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      loadingStates.root = true;

      await sshStore.updateRootLogin(enabled);
      await refreshSourceConfig();
      Message.success(t('app.ssh.rootModal.saveSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.rootModal.saveError'));
    } finally {
      loadingStates.root = false;
    }
  };

  /* 开关处理函数 */
  const handlePasswordAuthChange = async (value: boolean): Promise<void> => {
    try {
      if (!sshStore.hostId) return;

      loadingStates.passwordAuth = true;
      await sshStore.updatePasswordAuth(value);
      await refreshSourceConfig();
      Message.success(t('app.ssh.config.updateSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.config.updateError'));
    } finally {
      loadingStates.passwordAuth = false;
    }
  };

  const handleKeyAuthChange = async (value: boolean): Promise<void> => {
    try {
      if (!sshStore.hostId) return;

      loadingStates.keyAuth = true;
      await sshStore.updateKeyAuth(value);
      await refreshSourceConfig();
      Message.success(t('app.ssh.config.updateSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.config.updateError'));
    } finally {
      loadingStates.keyAuth = false;
    }
  };

  const handleReverseLookupChange = async (value: boolean): Promise<void> => {
    try {
      if (!sshStore.hostId) return;

      loadingStates.reverseLookup = true;
      await sshStore.updateReverseLookup(value);
      await refreshSourceConfig();
      Message.success(t('app.ssh.config.updateSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.config.updateError'));
    } finally {
      loadingStates.reverseLookup = false;
    }
  };

  const openSettingModal = (type: 'port' | 'listen' | 'root'): void => {
    if (type === 'port') {
      portModalVisible.value = true;
    } else if (type === 'listen') {
      listenModalVisible.value = true;
    } else if (type === 'root') {
      rootModalVisible.value = true;
    }
  };

  /* 处理源码编辑器保存 */
  const handleSaveSourceConfig = async (): Promise<void> => {
    await saveSourceConfig();
  };
</script>

<style scoped lang="less">
  .ssh-page-container {
    padding: 0 16px;
    background-color: var(--color-bg-2);
    border-radius: 6px;
    position: relative;
    border: 1px solid var(--color-border-2);
    box-shadow: 0 2px 8px var(--color-fill-2);
  }

  .header-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 0;
    margin-bottom: 16px;
  }

  .operation-alert {
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
    background-color: var(--color-bg-2);
    border-radius: 4px;
    border: 1px solid var(--color-border-2);
    padding: 16px 20px;
    margin-bottom: 16px;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--color-bg-2);
    border-radius: 4px;
    border: 1px solid var(--color-border-2);
    padding: 40px 20px;
    margin-bottom: 16px;
    min-height: 200px;

    .loading-text {
      margin-top: 16px;
      color: var(--color-text-2);
      font-size: 14px;
    }
  }
</style>
