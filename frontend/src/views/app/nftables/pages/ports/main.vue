<template>
  <div class="ports-management-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">{{ $t('app.nftables.ports.pageTitle') }}</h1>
      </div>
      <div class="header-right">
        <!-- 基础规则配置 -->
        <base-rules-config />
      </div>
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
      :width="'35.714rem'"
      :mask-closable="true"
      placement="right"
      @cancel="handleRuleCancel"
    >
      <port-rule-form
        ref="portRuleFormRef"
        :loading="saving"
        :editing-rule="editingRule"
        @submit="handleRuleSubmit"
        @cancel="handleRuleCancel"
      />
      <template #footer>
        <div class="drawer-footer">
          <a-space>
            <a-button @click="handleRuleCancel">{{
              $t('common.cancel')
            }}</a-button>
            <a-button
              type="primary"
              :loading="saving"
              @click="portRuleFormRef?.submitForm?.()"
            >
              {{
                editingRule
                  ? $t('app.nftables.form.updateRule')
                  : $t('app.nftables.button.addRule')
              }}
            </a-button>
          </a-space>
        </div>
      </template>
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
  import BaseRulesConfig from '../../components/base-rules-config.vue';

  const portRuleFormRef = ref<InstanceType<typeof PortRuleForm>>();

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
  const savePortRule = async (rule: PortRule): Promise<boolean> => {
    try {
      saving.value = true;
      hasError.value = false;

      // 构造规则（避免嵌套三元）
      const rulesLocal: SetPortRuleApiParams['rules'] = [];
      if (rule.rules && rule.rules.length) {
        rulesLocal.push(...rule.rules);
      } else if (rule.action) {
        rulesLocal.push({ type: 'default', action: rule.action });
      }

      // 当前版本不支持批量端口提交：如检测到端口段或多个端口，直接提示并中止
      let portValue: number | null = null;
      if (typeof rule.port === 'number') {
        portValue = rule.port;
      } else if (Array.isArray(rule.port)) {
        const arr = rule.port as number[];
        if (arr.length !== 1) {
          Message.error(
            (t && t('app.nftables.message.batchNotSupported')) ||
              '当前版本暂不支持批量端口提交，请选择单个端口'
          );
          return false; // avoid double error message
        }
        portValue = arr[0];
      }

      // 逐个端口保存
      if (portValue == null) {
        Message.error(t('app.nftables.message.operationFailed'));
        return false; // avoid double error message
      }

      await setPortRulesApi({ port: portValue, rules: rulesLocal });

      Message.success(t('app.nftables.message.configSaved'));

      // 端口页无需全局刷新，交由调用方按需刷新（避免多余的 process/port 请求）
      return true;
    } catch (error) {
      logError('Failed to save port rule:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.error.saveConfigFailed');
      Message.error(t('app.nftables.message.operationFailed'));
      return false;
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
      const ok = await savePortRule(rule);
      if (!ok) return; // stop on validation or api failure

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
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1.714rem; // 24px -> rem
      border-bottom: 1px solid var(--color-border-2);

      .header-left {
        .page-title {
          font-size: 1.714rem; // 24px -> rem
          font-weight: 600;
          color: var(--color-text-1);
          margin: 0 0 0.571rem 0; // 8px -> rem
        }

        .page-description {
          font-size: 1rem; // 14px -> rem (root 14px)
          color: var(--color-text-3);
          margin: 0;
        }
      }

      .header-right {
        display: flex;
        align-items: center;
      }
    }

    .page-content {
      padding: 1.143rem 0; // 16px -> rem

      .loading-container {
        padding: 4.286rem 1.714rem; // 60px 24px -> rem
        min-height: 28.571rem; // 400px -> rem
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;

        .loading-text {
          margin-top: 1.143rem; // 16px -> rem
          font-size: 1rem; // 14px -> rem
          color: var(--color-text-3);
          font-weight: 400;
        }
      }

      .rules-management-container {
        background: var(--color-bg-2);
        border-radius: 0.429rem; // 6px -> rem
        padding: 1.143rem; // 16px -> rem
        margin: 1.143rem 1.714rem 0; // 16px 24px 0 -> rem
      }

      .error-state {
        margin-top: 2.857rem; // 40px -> rem
        text-align: center;
      }
    }
  }

  // 响应式设计
  @media (max-width: 1024px) {
    .ports-management-page {
      .page-header {
        padding-left: 1.143rem; // 16px -> rem
        padding-right: 1.143rem; // 16px -> rem
      }

      .page-content {
        .rules-management-container {
          margin-left: 1.143rem; // 16px -> rem
          margin-right: 1.143rem; // 16px -> rem
        }
      }
    }
  }

  @media (max-width: @screen-md) {
    // 768px -> 使用系统断点变量
    .ports-management-page {
      .page-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 1.143rem; // 16px -> rem
        padding-left: 0.857rem; // 12px -> rem
        padding-right: 0.857rem; // 12px -> rem

        .header-right {
          align-self: flex-end;
        }
      }

      .page-content {
        .rules-management-container {
          margin-left: 0.857rem; // 12px -> rem
          margin-right: 0.857rem; // 12px -> rem
          padding: 0.857rem; // 12px -> rem
        }
      }
    }
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
    margin-top: 24px;
    border-top: 1px solid var(--color-border-2);
  }
</style>
