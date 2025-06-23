/**
 * nftables配置页面的组合式函数
 */

import { ref, watch } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useHostStore } from '@/store';
import { useLogger } from '@/composables/use-logger';
import {
  getProcessStatusApi,
  getFirewallStatusApi,
  getPortRulesApi,
  switchFirewallApi,
  installApi,
  type ProcessStatus,
  type NftablesStatus,
  type PortRuleSet,
} from '@/api/nftables';
import { isNftablesActive, isIptablesActive } from '../utils/firewall-status';

export function useNftablesConfig() {
  const { t } = useI18n();
  const hostStore = useHostStore();
  const { logError } = useLogger('NftablesConfig');

  // 响应式状态
  const loading = ref<boolean>(false);
  const statusLoading = ref<boolean>(false);
  const switchLoading = ref<boolean>(false);
  const processData = ref<ProcessStatus[]>([]);
  const firewallStatus = ref<NftablesStatus | null>(null);
  const portRules = ref<PortRuleSet[]>([]);
  const openPorts = ref<Set<number>>(new Set());

  // 获取进程状态数据
  const fetchProcessData = async (showSuccess = false): Promise<void> => {
    if (!hostStore.currentId) return;

    try {
      loading.value = true;
      const response = await getProcessStatusApi();

      // 处理API返回的数据：{total: number, items: ProcessStatus[]}
      processData.value = response.items || [];

      if (showSuccess) {
        Message.success(
          t('app.nftables.message.fetchSuccess', {
            count: processData.value.length,
          })
        );
      }
    } catch (error) {
      logError('Failed to fetch process data:', error);
      Message.error(t('app.nftables.message.fetchFailed'));
      processData.value = [];
    } finally {
      loading.value = false;
    }
  };

  // 获取防火墙状态数据
  const fetchFirewallStatus = async (showSuccess = false): Promise<void> => {
    if (!hostStore.currentId) return;

    try {
      statusLoading.value = true;
      const response = await getFirewallStatusApi();
      firewallStatus.value = response;

      if (showSuccess) {
        Message.success(t('app.nftables.message.statusFetchSuccess'));
      }
    } catch (error) {
      logError('Failed to fetch firewall status:', error);
      Message.error(t('app.nftables.message.statusFetchFailed'));
      firewallStatus.value = null;
    } finally {
      statusLoading.value = false;
    }
  };

  // 获取端口规则数据
  const fetchPortRules = async (): Promise<void> => {
    if (!hostStore.currentId) return;

    try {
      const response = await getPortRulesApi();
      portRules.value = response.items || [];

      // 从端口规则中提取开放端口
      const ports = new Set<number>();
      portRules.value.forEach((rule) => {
        // 只有规则中包含 accept 动作的端口才算作开放端口
        if (rule.rules.some((r) => r.action === 'accept')) {
          ports.add(rule.port);
        }
      });
      openPorts.value = ports;
    } catch (error) {
      logError('Failed to fetch port rules:', error);
      // 静默失败，不显示错误消息，因为可能是nftables未配置
      portRules.value = [];
      openPorts.value = new Set();
    }
  };

  // 状态刷新
  const handleStatusRefresh = () => {
    fetchFirewallStatus(true);
  };

  // 处理切换操作
  const handleSwitch = async (): Promise<void> => {
    if (!firewallStatus.value || !hostStore.currentId) return;

    try {
      switchLoading.value = true;

      const active = firewallStatus.value.active;
      const status = firewallStatus.value.status;

      // 当前是nftables，切换到iptables
      if (isNftablesActive(active)) {
        await switchFirewallApi({ option: 'iptables' });
        Message.success(t('app.nftables.message.switchToIptablesSuccess'));
      }
      // 当前是iptables，切换到nftables
      else if (isIptablesActive(active)) {
        // 如果nftables未安装，先安装
        if (status !== 'installed') {
          Message.info(t('app.nftables.message.installingNftables'));
          await installApi();
          Message.success(t('app.nftables.message.installSuccess'));
        }

        // 切换到nftables
        await switchFirewallApi({ option: 'nftables' });
        Message.success(t('app.nftables.message.switchToNftablesSuccess'));
      }

      // 切换成功后刷新状态
      await fetchFirewallStatus();
    } catch (error) {
      logError('Failed to switch firewall:', error);
      Message.error(t('app.nftables.message.switchFailed'));
    } finally {
      switchLoading.value = false;
    }
  };

  // 表格重新加载
  const handleReload = async () => {
    await Promise.all([fetchProcessData(true), fetchPortRules()]);
  };

  // 监听主机ID变化
  watch(
    () => hostStore.currentId,
    async (newId) => {
      if (newId) {
        await Promise.all([
          fetchFirewallStatus(),
          fetchProcessData(),
          fetchPortRules(),
        ]);
      }
    },
    { immediate: true }
  );

  return {
    // 状态
    loading,
    statusLoading,
    switchLoading,
    processData,
    firewallStatus,
    portRules,
    openPorts,

    // 方法
    fetchProcessData,
    fetchFirewallStatus,
    fetchPortRules,
    handleStatusRefresh,
    handleSwitch,
    handleReload,
  };
}
