<template>
  <div class="ip-blacklist-management-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">{{ $t('app.nftables.ipBlacklist.pageTitle') }}</h1>
    </div>

    <div class="page-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <a-spin :size="28" />
        <div class="loading-text">{{ $t('common.loading') }}</div>
      </div>

      <!-- IP黑名单规则管理 -->
      <div v-else-if="!hasError" class="rules-management-container">
        <!-- IP黑名单规则列表 -->
        <IPBlacklistRuleList
          :ip-rules="ipBlacklistRules"
          :loading="loading"
          @add="handleRuleAdd"
          @delete="handleRuleDelete"
        />
      </div>

      <!-- 错误状态 -->
      <div v-else class="error-state">
        <a-result status="error" :title="$t('common.error.title')">
          <template #subtitle>
            {{ errorMessage }}
          </template>
          <template #extra>
            <a-button type="primary" @click="handleRetry">
              {{ $t('common.button.retry') }}
            </a-button>
          </template>
        </a-result>
      </div>
    </div>

    <!-- 添加规则抽屉 -->
    <a-drawer
      v-model:visible="showRuleForm"
      :title="$t('app.nftables.drawer.addIPRule')"
      :width="500"
      :footer="false"
      :mask-closable="true"
      placement="right"
      @cancel="handleRuleCancel"
    >
      <IPBlacklistRuleForm
        :loading="saving"
        @submit="handleRuleSubmit"
        @cancel="handleRuleCancel"
      />
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, computed } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useHostStore } from '@/store';
  import { useLogger } from '@/composables/use-logger';
  import {
    getIPBlacklistApi,
    addIPBlacklistApi,
    deleteIPBlacklistApi,
    type IPBlacklistRequest,
    type DeleteIPBlacklistRequest,
  } from '@/api/nftables';
  import IPBlacklistRuleForm from '../../components/ip-blacklist-rule-form.vue';
  import IPBlacklistRuleList from '../../components/ip-blacklist-rule-list.vue';

  // IP黑名单规则接口
  export interface IPBlacklistRule {
    ip: string;
    type: 'single' | 'cidr' | 'range';
    action: 'drop' | 'reject';
    description?: string;
    createdAt?: string;
  }

  // 国际化
  const { t } = useI18n();

  // 主机存储
  const hostStore = useHostStore();

  // 日志记录
  const { logError } = useLogger('NftablesIPBlacklistPage');

  // 响应式状态
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const ipBlacklistItems = ref<string[]>([]);
  const hasError = ref<boolean>(false);
  const errorMessage = ref<string>('');
  const showRuleForm = ref<boolean>(false);

  // IP地址格式验证
  const isValidIPAddress = (ip: string): boolean => {
    const ipRegex =
      /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return ipRegex.test(ip);
  };

  // IP验证工具函数
  const validateIP = (ip: string): boolean => {
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
  };

  // 转换IP字符串数组为规则对象数组
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

  // 计算属性：转换IP黑名单字符串数组为规则对象数组
  const ipBlacklistRules = computed((): IPBlacklistRule[] => {
    return convertToIPRules([...ipBlacklistItems.value]);
  });

  // 检查主机是否可用
  const checkHostAvailable = (): boolean => {
    if (!hostStore.currentId) {
      Message.error(t('common.error.noHostSelected'));
      return false;
    }
    return true;
  };

  // 获取IP黑名单列表
  const fetchIPBlacklist = async (): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      loading.value = true;
      hasError.value = false;

      const response = await getIPBlacklistApi();
      ipBlacklistItems.value = response.items || [];
    } catch (error) {
      logError('Failed to fetch IP blacklist:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.fetchConfigFailed');
      Message.error(t('app.nftables.message.operationFailed'));
      ipBlacklistItems.value = [];
    } finally {
      loading.value = false;
    }
  };

  // 保存IP黑名单规则
  const saveIPBlacklistRule = async (rule: IPBlacklistRule): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      saving.value = true;
      hasError.value = false;

      const request: IPBlacklistRequest = {
        ip: rule.ip,
      };

      await addIPBlacklistApi(request);

      Message.success(t('app.nftables.message.configSaved'));
    } catch (error) {
      logError('Failed to save IP blacklist rule:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.saveConfigFailed');
      Message.error(t('app.nftables.message.operationFailed'));
      throw error;
    } finally {
      saving.value = false;
    }
  };

  // 规则表单相关事件处理
  const handleRuleAdd = (): void => {
    showRuleForm.value = true;
  };

  const handleRuleSubmit = async (rule: IPBlacklistRule): Promise<void> => {
    try {
      // 保存规则到服务器
      await saveIPBlacklistRule(rule);

      // 刷新IP黑名单列表
      await fetchIPBlacklist();

      showRuleForm.value = false;

      Message.success(t('app.nftables.message.ruleAdded'));
    } catch (error) {
      // 错误已在 saveIPBlacklistRule 中处理
    }
  };

  const handleRuleCancel = (): void => {
    showRuleForm.value = false;
  };

  const handleRuleDelete = async (ip: string): Promise<void> => {
    if (!checkHostAvailable()) return;

    try {
      const request: DeleteIPBlacklistRequest = { ip };

      await deleteIPBlacklistApi(request);

      // 刷新IP黑名单列表
      await fetchIPBlacklist();

      Message.success(t('app.nftables.message.ruleDeleted'));
    } catch (error) {
      logError('Failed to delete IP blacklist rule:', error);
      Message.error(t('app.nftables.message.operationFailed'));
    }
  };

  // 重试处理
  const handleRetry = async (): Promise<void> => {
    await fetchIPBlacklist();
  };

  // 页面初始化
  onMounted(async () => {
    await fetchIPBlacklist();
  });
</script>

<style scoped lang="less">
  .ip-blacklist-management-page {
    background: var(--color-bg-1);
    min-height: 100vh;

    .page-header {
      padding: 24px;
      border-bottom: 1px solid var(--color-border-2);

      .page-title {
        font-size: 24px;
        font-weight: 600;
        color: var(--color-text-1);
        margin: 0 0 8px 0;
      }

      .page-description {
        font-size: 14px;
        color: var(--color-text-3);
        margin: 0;
      }
    }

    .page-content {
      padding: 16px 0;

      .loading-container {
        padding: 60px 24px;
        min-height: 400px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;

        .loading-text {
          margin-top: 16px;
          font-size: 14px;
          color: var(--color-text-3);
          font-weight: 400;
        }
      }

      .rules-management-container {
        background: var(--color-bg-2);
        border-radius: 6px;
        padding: 16px;
        margin: 16px 24px 0;
      }

      .error-state {
        margin-top: 40px;
        text-align: center;
      }
    }
  }

  // 响应式设计
  @media (max-width: 1024px) {
    .ip-blacklist-management-page {
      .page-header {
        padding-left: 16px;
        padding-right: 16px;
      }

      .page-content {
        .rules-management-container {
          margin-left: 16px;
          margin-right: 16px;
        }
      }
    }
  }

  @media (max-width: 768px) {
    .ip-blacklist-management-page {
      .page-header {
        padding-left: 12px;
        padding-right: 12px;
      }

      .page-content {
        .rules-management-container {
          margin-left: 12px;
          margin-right: 12px;
          padding: 12px;
        }
      }
    }
  }
</style>
