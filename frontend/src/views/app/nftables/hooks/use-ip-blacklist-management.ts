/**
 * IP黑名单管理的组合式函数
 * @description 提供IP黑名单规则管理的相关功能
 */

import { ref, computed, readonly } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useHostStore } from '@/store';
import { useLogger } from '@/hooks/use-logger';
import {
  getIPBlacklistApi,
  addIPBlacklistApi,
  deleteIPBlacklistApi,
  getFirewallStatusApi,
  type IPBlacklistRequest,
  type DeleteIPBlacklistRequest,
} from '@/api/nftables';

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
  const ipBlacklist = ref<string[]>([]);

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
   * 保存IP黑名单规则
   */
  const saveIPBlacklistRule = async (rule: IPBlacklistRule): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;
      const request: IPBlacklistRequest = {
        ip: rule.ip,
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
   * 初始化
   */
  const initialize = async (): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      // 获取防火墙状态
      const statusResponse = await getFirewallStatusApi();
      logInfo('Firewall status:', statusResponse);

      // 加载IP黑名单数据
      await fetchIPBlacklist();
    } catch (error) {
      handleApiError(error, 'initialize');
    }
  };

  // 返回响应式数据和方法
  return {
    // 状态
    loading: readonly(loading),
    saving: readonly(saving),
    ipBlacklist: readonly(ipBlacklist),

    // 转换函数
    convertToIPRules,

    // 方法
    saveIPBlacklistRule,
    deleteIPBlacklistRule,
    initialize,
    fetchIPBlacklist,
  };
}
