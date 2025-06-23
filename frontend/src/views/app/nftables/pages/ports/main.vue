<template>
  <div class="ports-management-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">{{ $t('app.nftables.ports.pageTitle') }}</h1>
    </div>

    <div class="page-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <a-spin :size="28" />
        <div class="loading-text">{{ $t('common.loading') }}</div>
      </div>

      <!-- 端口规则管理 -->
      <div v-else-if="!hasError" class="rules-management-container">
        <!-- 端口规则列表 -->
        <port-rule-list
          :port-rules="portRules"
          :loading="loading"
          @add="handleRuleAdd"
          @edit="handleRuleEdit"
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

    <!-- 添加/编辑规则抽屉 -->
    <a-drawer
      v-model:visible="showRuleForm"
      :title="
        editingRule
          ? $t('app.nftables.drawer.editRule')
          : $t('app.nftables.drawer.addRule')
      "
      :width="500"
      :footer="false"
      :mask-closable="true"
      placement="right"
      @cancel="handleRuleCancel"
    >
      <port-rule-form
        :loading="saving"
        :editing-rule="editingRule"
        @submit="handleRuleSubmit"
        @cancel="handleRuleCancel"
      />
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/composables/use-logger';
  import type {
    PortRule,
    PortRuleSet,
    SetPortRuleApiParams,
  } from '@/api/nftables';
  import {
    getPortRulesApi,
    setPortRulesApi,
    deletePortRulesApi,
  } from '@/api/nftables';
  import PortRuleForm from '../../components/port-rule-form.vue';
  import PortRuleList from '../../components/port-rule-list.vue';

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError } = useLogger('NftablesPortsPage');

  // 响应式状态
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const portRules = ref<PortRule[]>([]);
  const hasError = ref<boolean>(false);
  const errorMessage = ref<string>('');
  const showRuleForm = ref<boolean>(false);
  const editingRule = ref<PortRule | null>(null);

  // 将PortRuleSet转换为PortRule格式
  const convertPortRuleSetToPortRule = (ruleSet: PortRuleSet): PortRule => {
    return {
      port: ruleSet.port,
      protocol: 'tcp', // 根据API文档，当前只支持TCP
      rules: ruleSet.rules,
      // 为了兼容性，如果只有一个默认规则，也设置action
      action:
        ruleSet.rules.length === 1 && ruleSet.rules[0].type === 'default'
          ? ruleSet.rules[0].action
          : undefined,
    };
  };

  // 获取端口规则列表
  const fetchPortRules = async (): Promise<void> => {
    try {
      loading.value = true;
      hasError.value = false;

      const response = await getPortRulesApi();

      // 转换API返回的数据格式
      portRules.value = response.items.map(convertPortRuleSetToPortRule);
    } catch (error) {
      logError('Failed to fetch port rules:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.fetchConfigFailed');
      Message.error(t('app.nftables.message.operationFailed'));
      portRules.value = [];
    } finally {
      loading.value = false;
    }
  };

  // 保存单个端口规则
  const savePortRule = async (rule: PortRule): Promise<void> => {
    try {
      saving.value = true;
      hasError.value = false;

      // 构造API参数
      const apiParams: SetPortRuleApiParams = {
        port: typeof rule.port === 'number' ? rule.port : rule.port[0],
        rules: rule.rules || [],
      };

      // 如果没有高级规则但有基础action，创建默认规则
      if (!apiParams.rules.length && rule.action) {
        apiParams.rules = [
          {
            type: 'default',
            action: rule.action,
          },
        ];
      }

      await setPortRulesApi(apiParams);

      Message.success(t('app.nftables.message.configSaved'));
    } catch (error) {
      logError('Failed to save port rule:', error);
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
    editingRule.value = null;
    showRuleForm.value = true;
  };

  const handleRuleEdit = (rule: PortRule): void => {
    editingRule.value = { ...rule };
    showRuleForm.value = true;
  };

  const handleRuleSubmit = async (rule: PortRule): Promise<void> => {
    try {
      // 保存规则到服务器
      await savePortRule(rule);

      // 刷新端口规则列表
      await fetchPortRules();

      showRuleForm.value = false;
      editingRule.value = null;

      Message.success(
        editingRule.value
          ? t('app.nftables.message.ruleUpdated')
          : t('app.nftables.message.ruleAdded')
      );
    } catch (error) {
      // 错误已在 savePortRule 中处理
    }
  };

  const handleRuleCancel = (): void => {
    showRuleForm.value = false;
    editingRule.value = null;
  };

  const handleRuleDelete = async (ruleToDelete: PortRule): Promise<void> => {
    try {
      const port =
        typeof ruleToDelete.port === 'number'
          ? ruleToDelete.port
          : ruleToDelete.port[0];

      await deletePortRulesApi({ port });

      // 刷新端口规则列表
      await fetchPortRules();

      Message.success(t('app.nftables.message.ruleDeleted'));
    } catch (error) {
      logError('Failed to delete port rule:', error);
      Message.error(t('app.nftables.message.operationFailed'));
    }
  };

  // 重试处理
  const handleRetry = async (): Promise<void> => {
    await fetchPortRules();
  };

  // 页面初始化
  onMounted(async () => {
    await fetchPortRules();
  });
</script>

<style scoped lang="less">
  .ports-management-page {
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
    .ports-management-page {
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
    .ports-management-page {
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
