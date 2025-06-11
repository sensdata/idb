/**
 * IP黑名单管理的组合式函数
 * @description 提供IP黑名单规则和配置管理的相关功能
 */

import { ref, computed, readonly, type Ref } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useHostStore } from '@/store';
import { useLogger } from '@/hooks/use-logger';
import {
  getIPBlacklistApi,
  addIPBlacklistApi,
  deleteIPBlacklistApi,
  getIPBlacklistConfigApi,
  updateIPBlacklistConfigApi,
  createIPBlacklistConfigApi,
  getDefaultIPBlacklistConfigTemplate,
  activateConfigApi,
  getFirewallStatusApi,
  type IPBlacklistRequest,
  type DeleteIPBlacklistRequest,
  type ConfigType,
} from '@/api/nftables';

export type ConfigModeType = 'form' | 'file';

// 常量定义
const CONFIG_CATEGORIES = {
  IP_BLACKLIST: 'ip-blacklist',
  MAIN_CONFIG: 'main.nft',
} as const;

const ACTIONS = {
  ACTIVATE: 'activate',
} as const;

/**
 * IP黑名单规则接口
 */
export interface IPBlacklistRule {
  ip: string;
  type: 'single' | 'cidr' | 'range';
  action: 'drop' | 'reject';
  description?: string;
  createdAt?: string;
}

/**
 * 获取默认IP黑名单配置模板
 * @returns 默认的 NFTables IP黑名单配置内容
 */
const getDefaultIPBlacklistConfig = (): string => {
  return getDefaultIPBlacklistConfigTemplate();
};

/**
 * IP地址格式验证
 */
function isValidIPAddress(ip: string): boolean {
  const ipRegex =
    /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  return ipRegex.test(ip);
}

/**
 * IP验证工具函数
 */
function validateIP(ip: string): boolean {
  // CIDR格式
  if (ip.includes('/')) {
    const [address, mask] = ip.split('/');
    const maskNum = parseInt(mask, 10);
    return isValidIPAddress(address) && maskNum >= 0 && maskNum <= 32;
  }

  // 范围格式
  if (ip.includes('-')) {
    const [start, end] = ip.split('-');
    return isValidIPAddress(start) && isValidIPAddress(end);
  }

  // 单个IP
  return isValidIPAddress(ip);
}

/**
 * IP黑名单管理组合式函数
 */
export function useIPBlacklistManagement() {
  const { t } = useI18n();
  const hostStore = useHostStore();
  const { logError, logInfo } = useLogger('IPBlacklistManagement');

  // 状态管理
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const configMode = ref<ConfigModeType>('form');
  const configType = ref<ConfigType>('local');
  const ipBlacklist = ref<string[]>([]);
  const configContent = ref<string>('');
  const isConfigExist = ref<boolean>(false);
  const activeConfigType = ref<ConfigType>('local');

  // 计算属性
  const isFormMode = computed(() => configMode.value === 'form');
  const isFileMode = computed(() => configMode.value === 'file');

  /**
   * 检查主机是否可用
   */
  const checkHostAvailable = (): boolean => {
    if (!hostStore.currentId) {
      Message.error(t('common.error.noHostSelected'));
      return false;
    }
    return true;
  };

  /**
   * 处理API错误
   */
  const handleApiError = (
    error: any,
    operation: string,
    showMessage = true
  ): void => {
    const errorMsg =
      error?.response?.data?.message ||
      error?.message ||
      t('common.error.unknown');
    logError(`Failed to ${operation}:`, error);
    if (showMessage) {
      Message.error(`${operation} 失败: ${errorMsg}`);
    }
  };

  /**
   * 转换IP字符串数组为规则对象数组
   */
  const convertToIPRules = (ipList: string[]): IPBlacklistRule[] => {
    return ipList
      .filter((ip) => {
        const isValid = validateIP(ip);
        if (!isValid) {
          logError(`Invalid IP format: ${ip}`);
        }
        return isValid;
      })
      .map((ip) => {
        let type: 'single' | 'cidr' | 'range' = 'single';

        if (ip.includes('/')) {
          type = 'cidr';
        } else if (ip.includes('-')) {
          type = 'range';
        }

        return {
          ip,
          type,
          action: 'drop',
          description: '',
          createdAt: new Date().toISOString(),
        };
      });
  };

  /**
   * 获取IP黑名单数据
   * @param showSuccess 是否显示成功消息
   */
  const fetchIPBlacklist = async (showSuccess = false): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      loading.value = true;
      const response = await getIPBlacklistApi();
      ipBlacklist.value = response.items || [];

      if (showSuccess) {
        Message.success(
          t('app.nftables.message.fetchSuccess', {
            count: ipBlacklist.value.length,
          })
        );
      }

      logInfo(`Fetched ${ipBlacklist.value.length} IP blacklist rules`);
    } catch (error) {
      handleApiError(error, 'fetch IP blacklist');
      ipBlacklist.value = [];
    } finally {
      loading.value = false;
    }
  };

  /**
   * 获取配置文件内容
   */
  const fetchConfigContent = async (): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      loading.value = true;
      const content = await getIPBlacklistConfigApi(configType.value);
      configContent.value = content;
      isConfigExist.value = true;
      logInfo(`Fetched config content for type: ${configType.value}`);
    } catch (error) {
      // 如果配置文件不存在，使用默认配置
      configContent.value = getDefaultIPBlacklistConfig();
      isConfigExist.value = false;
      logInfo('Using default config content');
    } finally {
      loading.value = false;
    }
  };

  /**
   * 保存IP黑名单规则
   */
  const saveIPBlacklistRule = async (rule: IPBlacklistRule): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;
      const request: IPBlacklistRequest = {
        ip: rule.ip,
        description: rule.description,
      };

      await addIPBlacklistApi(request);
      await fetchIPBlacklist(); // 重新获取列表
      logInfo(`Added IP blacklist rule: ${rule.ip}`);
    } catch (error) {
      handleApiError(error, 'save IP blacklist rule');
      throw error;
    } finally {
      saving.value = false;
    }
  };

  /**
   * 删除IP黑名单规则
   */
  const deleteIPBlacklistRule = async (ip: string): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;
      const request: DeleteIPBlacklistRequest = { ip };
      await deleteIPBlacklistApi(request);
      await fetchIPBlacklist(); // 重新获取列表
      logInfo(`Deleted IP blacklist rule: ${ip}`);
    } catch (error) {
      handleApiError(error, 'delete IP blacklist rule');
      throw error;
    } finally {
      saving.value = false;
    }
  };

  /**
   * 保存配置文件内容
   */
  const saveConfigContent = async (content: string): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;

      if (isConfigExist.value) {
        await updateIPBlacklistConfigApi(content, configType.value);
      } else {
        await createIPBlacklistConfigApi(content, configType.value);
        isConfigExist.value = true;
      }

      configContent.value = content;
      logInfo(`Saved config content for type: ${configType.value}`);
    } catch (error) {
      handleApiError(error, 'save config content');
      throw error;
    } finally {
      saving.value = false;
    }
  };

  /**
   * 切换配置模式
   */
  const switchConfigMode = (mode: ConfigModeType): void => {
    configMode.value = mode;
    Message.success(
      t('app.nftables.message.switchedToMode', {
        mode: t(`app.nftables.mode.${mode}`),
      })
    );
    logInfo(`Switched to ${mode} mode`);
  };

  /**
   * 切换配置类型
   */
  const switchConfigType = async (type: ConfigType): Promise<void> => {
    configType.value = type;
    Message.success(
      t('app.nftables.message.switchedToConfigType', {
        type: t(`app.nftables.configType.${type}`),
      })
    );

    // 重新获取对应类型的配置
    if (isFileMode.value) {
      await fetchConfigContent();
    } else {
      await fetchIPBlacklist();
    }

    logInfo(`Switched to ${type} config type`);
  };

  /**
   * 激活配置
   */
  const activateConfig = async (): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;
      await activateConfigApi({
        type: configType.value,
        category: CONFIG_CATEGORIES.IP_BLACKLIST,
        name: CONFIG_CATEGORIES.MAIN_CONFIG,
        action: ACTIONS.ACTIVATE,
      });

      activeConfigType.value = configType.value;
      Message.success(t('app.nftables.message.configApplied'));
      logInfo(`Activated ${configType.value} config`);
    } catch (error) {
      handleApiError(error, 'activate config');
      throw error;
    } finally {
      saving.value = false;
    }
  };

  /**
   * 初始化
   */
  const initialize = async (): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      // 获取防火墙状态
      const statusResponse = await getFirewallStatusApi();
      logInfo('Firewall status:', statusResponse);

      // 根据当前模式加载数据
      if (isFormMode.value) {
        await fetchIPBlacklist();
      } else {
        await fetchConfigContent();
      }
    } catch (error) {
      handleApiError(error, 'initialize');
    }
  };

  /**
   * 处理配置刷新
   */
  const handleConfigRefresh = async (): Promise<void> => {
    await fetchConfigContent();
  };

  /**
   * 处理配置保存
   */
  const handleConfigSave = async (content: string): Promise<void> => {
    await saveConfigContent(content);
  };

  // 返回响应式数据和方法
  return {
    // 状态
    loading: readonly(loading),
    saving: readonly(saving),
    configMode: readonly(configMode),
    configType: readonly(configType),
    ipBlacklist: readonly(ipBlacklist),
    configContent: readonly(configContent),
    isConfigExist: readonly(isConfigExist),
    activeConfigType: readonly(activeConfigType),

    // 计算属性
    isFormMode,
    isFileMode,

    // 转换函数
    convertToIPRules,

    // 方法
    saveIPBlacklistRule,
    deleteIPBlacklistRule,
    saveConfigContent,
    switchConfigMode,
    switchConfigType,
    activateConfig,
    initialize,
    fetchConfigContent: handleConfigRefresh,
    handleConfigSave,
  };
}
