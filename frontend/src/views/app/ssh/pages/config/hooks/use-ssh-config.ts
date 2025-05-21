import { ref, computed } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { getSSHConfigContent, updateSSHConfigContent } from '@/api/ssh';
import { useLogger } from '@/utils/hooks/use-logger';
import type { SSHFormConfig } from '@/views/app/ssh/store/types';
import type { LoadingStates } from '../types';

// 定义Store类型接口，包含所需方法和属性
interface SSHStoreInterface {
  hostId: number | null;
  formConfig: SSHFormConfig;
}

export function useSSHConfig(
  sshStore: SSHStoreInterface,
  loadingStates: LoadingStates
) {
  const { t } = useI18n();
  const { logDebug, logInfo, logWarn, logError } = useLogger('SSHConfig');

  // 源配置内容
  const sourceConfig = ref<string>('');
  const originalSourceConfig = ref<string>('');

  const sshConfig = computed(() => sshStore.formConfig);

  // 使用ref管理内部状态
  const isContentFetching = ref(false);

  // 缓存30秒
  const configContentCache = {
    content: '',
    timestamp: 0,
    isValid: () => Date.now() - configContentCache.timestamp < 30000,
  };

  // 获取SSH配置内容
  const fetchSSHConfigContent = async (): Promise<void> => {
    try {
      if (!sshStore.hostId) {
        logWarn('无法获取SSH配置：未设置主机ID');
        Message.error(t('app.ssh.error.noHost'));
        return;
      }
      if (isContentFetching.value) {
        logDebug('已有请求正在进行中，跳过重复请求');
        return;
      }
      if (configContentCache.isValid() && configContentCache.content) {
        logInfo('使用缓存的SSH配置内容');
        sourceConfig.value = configContentCache.content;
        originalSourceConfig.value = configContentCache.content;
        return;
      }

      logInfo(`开始获取主机(${sshStore.hostId})的SSH配置内容`);
      isContentFetching.value = true;
      loadingStates.sourceConfig = true;

      const contentRes = await getSSHConfigContent(sshStore.hostId);
      if (!contentRes || !contentRes.content) {
        logError('SSH配置内容为空');
        Message.error(t('app.ssh.error.emptyConfig'));
        return;
      }

      logInfo('成功获取SSH配置内容');
      configContentCache.content = contentRes.content;
      configContentCache.timestamp = Date.now();
      sourceConfig.value = contentRes.content;
      originalSourceConfig.value = contentRes.content;
    } catch (error) {
      logError('获取SSH配置内容出错:', error);
      Message.error(t('app.ssh.error.fetchError'));
    } finally {
      isContentFetching.value = false;
      loadingStates.sourceConfig = false;
    }
  };

  // 从源配置内容更新表单
  const updateFormFromSourceConfig = (showMessage = true): void => {
    try {
      if (!sourceConfig.value || sourceConfig.value.trim() === '') {
        logWarn('源配置内容为空，无法解析');
        if (showMessage) {
          Message.warning(t('app.ssh.source.emptyConfig'));
        }
        return;
      }

      logInfo('开始从源配置更新表单数据');
      const lines = sourceConfig.value.split('\n');
      let hasUpdates = false;
      const updatedConfig: SSHFormConfig = { ...sshStore.formConfig };

      for (const line of lines) {
        const trimmedLine = line.trim();
        if (trimmedLine.startsWith('#') || trimmedLine === '') continue;
        try {
          if (trimmedLine.startsWith('Port ')) {
            updatedConfig.port = trimmedLine.replace('Port ', '').trim();
            hasUpdates = true;
            logDebug(`解析端口配置: ${updatedConfig.port}`);
          } else if (trimmedLine.startsWith('ListenAddress ')) {
            updatedConfig.listenAddress = trimmedLine
              .replace('ListenAddress ', '')
              .trim();
            hasUpdates = true;
            logDebug(`解析监听地址: ${updatedConfig.listenAddress}`);
          } else if (trimmedLine.startsWith('PermitRootLogin ')) {
            const value = trimmedLine
              .replace('PermitRootLogin ', '')
              .trim()
              .toLowerCase();
            updatedConfig.permitRootLogin = value;
            hasUpdates = true;
            logDebug(`解析Root登录权限: ${value}`);
          } else if (trimmedLine.startsWith('PasswordAuthentication ')) {
            const value = trimmedLine
              .replace('PasswordAuthentication ', '')
              .trim()
              .toLowerCase();
            updatedConfig.passwordAuth = value === 'yes';
            hasUpdates = true;
            logDebug(`解析密码认证: ${value}`);
          } else if (trimmedLine.startsWith('PubkeyAuthentication ')) {
            const value = trimmedLine
              .replace('PubkeyAuthentication ', '')
              .trim()
              .toLowerCase();
            updatedConfig.keyAuth = value === 'yes';
            hasUpdates = true;
            logDebug(`解析密钥认证: ${value}`);
          } else if (trimmedLine.startsWith('UseDNS ')) {
            const value = trimmedLine
              .replace('UseDNS ', '')
              .trim()
              .toLowerCase();
            updatedConfig.reverseLookup = value === 'yes';
            hasUpdates = true;
            logDebug(`解析DNS反查: ${value}`);
          }
        } catch (lineError) {
          logWarn(`解析配置行出错: ${trimmedLine}`, lineError);
        }
      }

      if (hasUpdates) {
        logInfo('成功从源配置更新表单数据');
        sshStore.formConfig = updatedConfig;
        if (showMessage) {
          Message.success(t('app.ssh.source.parseSuccess'));
        }
      } else {
        logWarn('没有从源配置检测到任何更改');
        if (showMessage) {
          Message.warning(t('app.ssh.source.noChanges'));
        }
      }
    } catch (error) {
      logError('从源配置更新表单出错:', error);
      if (showMessage) {
        Message.error(t('app.ssh.source.parseError'));
      }
    }
  };

  // 保存源配置内容
  const saveSourceConfig = async (): Promise<void> => {
    try {
      if (!sshStore.hostId) {
        logWarn('无法保存SSH配置：未设置主机ID');
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      logInfo(`开始保存主机(${sshStore.hostId})的SSH配置内容`);
      loadingStates.sourceConfig = true;

      await updateSSHConfigContent(sshStore.hostId, sourceConfig.value);
      configContentCache.content = sourceConfig.value;
      configContentCache.timestamp = Date.now();
      originalSourceConfig.value = sourceConfig.value;

      // 更新表单数据，确保视图模式数据与源文件一致，但不显示消息
      updateFormFromSourceConfig(false);

      logInfo('成功保存SSH配置内容');
      Message.success(t('app.ssh.source.saveSuccess'));
    } catch (error) {
      logError('保存SSH配置内容出错:', error);
      Message.error(t('app.ssh.source.saveError'));
    } finally {
      loadingStates.sourceConfig = false;
    }
  };

  // 重置源配置内容
  const resetSourceConfig = (): void => {
    logInfo('重置源配置内容');
    sourceConfig.value = originalSourceConfig.value;
    Message.info(t('app.ssh.source.resetSuccess'));
  };

  // 表单更新后刷新源配置内容
  const refreshSourceConfig = async (): Promise<void> => {
    if (loadingStates.sourceConfig) {
      logDebug('加载中，跳过刷新源配置内容');
      return;
    }

    try {
      logInfo('强制刷新源配置内容');
      configContentCache.timestamp = 0;
      if (sshStore.hostId) {
        await fetchSSHConfigContent();
      }
    } catch (error) {
      logError('刷新源配置内容出错:', error);
      // 错误已在 fetchSSHConfigContent 内部处理，无需额外处理
    }
  };

  return {
    sshConfig,
    sourceConfig,
    originalSourceConfig,
    fetchSSHConfigContent,
    updateFormFromSourceConfig,
    saveSourceConfig,
    resetSourceConfig,
    refreshSourceConfig,
  };
}
