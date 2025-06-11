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
      :footer="false"
      :mask-closable="true"
      placement="right"
      class="port-rule-drawer"
      @cancel="handleRuleCancel"
    >
      <port-rule-form
        :loading="saving"
        :editing-rule="editingRule"
        :port-readonly="!!editingRule"
        @submit="handleRuleSubmit"
        @cancel="handleRuleCancel"
      />
    </a-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/hooks/use-logger';
  import {
    setPortRulesApi,
    type PortRule,
    type SetPortRuleApiParams,
  } from '@/api/nftables';
  import { useNftablesConfig } from '../../hooks/use-nftables-config';
  import FirewallStatusHeader from '../../components/firewall-status-header.vue';
  import ProcessStatusTable from '../../components/process-status-table.vue';
  import PortRuleForm from '../../components/port-rule-form.vue';

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
  const savePortRule = async (rule: PortRule): Promise<void> => {
    try {
      saving.value = true;

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
      Message.error(t('app.nftables.message.operationFailed'));
      throw error;
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
</style>
