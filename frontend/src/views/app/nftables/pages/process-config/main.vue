<!--
  NFTables配置页面 - 使用GET /nftables/{host}/port接口获取端口信息
  已优化为使用专用的端口规则API，不再解析原始配置文件
-->
<template>
  <div class="nftables-page-container">
    <div class="header-container">
      <h2 class="page-title">{{ $t('app.nftables.config.title') }}</h2>

      <!-- 防火墙状态显示组件 -->
      <firewall-status-header
        :firewall-status="firewallStatus"
        :status-loading="statusLoading"
        :switch-loading="switchLoading"
        @switch="handleSwitch"
        @refresh="handleStatusRefresh"
      />
    </div>

    <!-- 进程状态表格组件 -->
    <process-status-table
      :process-data="processData"
      :open-ports="openPorts"
      :port-rules="portRules"
      :loading="loading"
      @reload="handleReload"
      @add-port-rule="handleAddPortRule"
      @edit-port-rule="handleEditPortRule"
    />

    <!-- 端口规则编辑抽屉 -->
    <a-drawer
      v-model:visible="showRuleForm"
      :title="
        editingRule
          ? $t('app.nftables.drawer.editRule')
          : $t('app.nftables.drawer.addRule')
      "
      :width="700"
      :mask-closable="true"
      placement="right"
      class="port-rule-drawer"
      @cancel="handleRuleCancel"
    >
      <port-rule-form
        ref="portRuleFormRef"
        :loading="saving"
        :editing-rule="editingRule"
        :port-readonly="!!editingRule"
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

<script lang="ts" setup>
  import { ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/composables/use-logger';
  import {
    setPortRulesApi,
    type PortRule,
    type SetPortRuleApiParams,
  } from '@/api/nftables';
  import { useNftablesConfig } from '../../composables/use-nftables-config';
  import FirewallStatusHeader from '../../components/firewall-status-header.vue';
  import ProcessStatusTable from '../../components/process-status-table.vue';
  import PortRuleForm from '../../components/port-rule-form.vue';

  const portRuleFormRef = ref<InstanceType<typeof PortRuleForm>>();

  const { t } = useI18n();
  const { logError } = useLogger('NftablesConfigPage');

  // 使用组合式函数获取所有状态和方法
  const {
    loading,
    statusLoading,
    switchLoading,
    processData,
    firewallStatus,
    portRules,
    openPorts,
    handleStatusRefresh,
    handleSwitch,
    handleReload,
    fetchPortRules,
  } = useNftablesConfig();

  // 端口规则编辑相关状态
  const showRuleForm = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const editingRule = ref<PortRule | null>(null);

  // 将PortRuleSet转换为PortRule格式
  const convertPortRuleSetToPortRule = (port: number): PortRule => {
    const ruleSet = portRules.value.find((rule) => rule.port === port);

    if (ruleSet) {
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
    }

    // 返回新规则的默认值
    return {
      port,
      protocol: 'tcp',
      action: 'accept',
      rules: [],
    };
  };

  // 保存端口规则
  const savePortRule = async (rule: PortRule): Promise<boolean> => {
    try {
      saving.value = true;

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

      if (portValue == null) {
        Message.error(t('app.nftables.message.operationFailed'));
        return false; // avoid double error message
      }

      await setPortRulesApi({ port: portValue, rules: rulesLocal });

      Message.success(t('app.nftables.message.configSaved'));
      return true;
    } catch (error) {
      logError('Failed to save port rule:', error);
      Message.error(t('app.nftables.message.operationFailed'));
      return false;
    } finally {
      saving.value = false;
    }
  };

  // 添加指定端口的规则
  const handleAddPortRule = (port: number): void => {
    editingRule.value = convertPortRuleSetToPortRule(port);
    editingRule.value.port = port;
    showRuleForm.value = true;
  };

  // 编辑端口规则
  const handleEditPortRule = (port: number): void => {
    editingRule.value = { ...convertPortRuleSetToPortRule(port) };
    showRuleForm.value = true;
  };

  // 提交规则表单
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

  // 取消规则编辑
  const handleRuleCancel = (): void => {
    showRuleForm.value = false;
    editingRule.value = null;
  };
</script>

<style scoped lang="less">
  .nftables-page-container {
    padding: 0 16px;
    position: relative;
    background: var(--color-bg-1);
  }

  .header-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    margin-bottom: 20px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 500;
    color: var(--color-text-1);
    margin: 0;
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .nftables-page-container {
      padding: 0 12px;
    }

    .header-container {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
  }

  @media (max-width: 480px) {
    .header-container {
      gap: 8px;
    }
  }

  /* Drawer边框样式 */
  :deep(.port-rule-drawer .arco-drawer-body) {
    border-left: 1px solid var(--color-border-2);
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
    margin-top: 24px;
    border-top: 1px solid var(--color-border-2);
  }
</style>
