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
  import type { PortRule, PortRangeRule, SetPortRuleReq } from '@/api/nftables';
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

  // 将后端的端口区间规则转换为UI使用的PortRule格式
  const convertPortRuleSetToPortRule = (ruleSet: PortRangeRule): PortRule => {
    const portVal =
      ruleSet.port_start === ruleSet.port_end
        ? ruleSet.port_start
        : [ruleSet.port_start, ruleSet.port_end];

    // 从 rules 数组中提取第一条规则的 action 和 src_ip 用于表格显示
    const firstRule = ruleSet.rules?.[0];
    const action = firstRule?.action;
    const source = firstRule?.src_ip;

    return {
      port: portVal as any,
      protocol: ruleSet.protocol ?? 'tcp',
      rules: ruleSet.rules,
      action,
      source,
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
      const rulesLocal: SetPortRuleReq['rules'] = [];
      if (rule.rules && rule.rules.length) {
        rulesLocal.push(...rule.rules);
      } else if (rule.action) {
        rulesLocal.push({ type: 'default', action: rule.action });
      }

      // 支持：
      // - 单个端口
      // - 端口区间（a-b）
      // - 端口列表（逗号分隔 a,b 或 a,b,c...），总是作为多个独立端口逐条提交
      let batchPorts: number[] | null = null;
      let portStart: number | null = null;
      let portEnd: number | null = null;
      if (typeof rule.port === 'number') {
        portStart = rule.port;
        portEnd = rule.port;
      } else if (Array.isArray(rule.port)) {
        const arr = (rule.port as number[]).filter(
          (n) => typeof n === 'number'
        );
        if (arr.length === 0) {
          Message.error(t('app.nftables.message.operationFailed'));
          return false;
        }
        if (rule.portInputType === 'list') {
          // 逗号分隔：包含 2 个也按列表逐条提交
          batchPorts = Array.from(new Set(arr));
        } else if (arr.length === 2) {
          // 连字符分隔：区间
          portStart = Math.min(arr[0], arr[1]);
          portEnd = Math.max(arr[0], arr[1]);
        } else {
          portStart = arr[0];
          portEnd = arr[0];
        }
      }

      if (batchPorts && batchPorts.length) {
        await Promise.all(
          batchPorts.map((p) =>
            setPortRulesApi({
              port_start: p,
              port_end: p,
              rules: rulesLocal,
            })
          )
        );
        Message.success(t('app.nftables.message.configSaved'));
        return true;
      }

      if (portStart == null || portEnd == null) {
        Message.error(t('app.nftables.message.operationFailed'));
        return false;
      }

      await setPortRulesApi({
        port_start: portStart,
        port_end: portEnd,
        rules: rulesLocal,
      });

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

  // 计算端口区间工具函数
  const getPortRange = (
    p: PortRule['port']
  ): { start: number; end: number } | null => {
    if (typeof p === 'number') return { start: p, end: p };
    if (Array.isArray(p)) {
      const arr = p as number[];
      if (arr.length === 0) return null;
      if (arr.length === 1) return { start: arr[0], end: arr[0] };
      return { start: Math.min(arr[0], arr[1]), end: Math.max(arr[0], arr[1]) };
    }
    return null;
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
    saving.value = true; // 提交全流程展示 loading（删除旧 + 保存新 + 刷新列表）
    try {
      const isEditing = !!editingRule.value;

      // 若编辑时修改了端口区间，则先删除旧规则再保存新规则，避免产生重复记录
      if (isEditing) {
        const oldRange = getPortRange(editingRule.value!.port);
        const newRange = getPortRange(rule.port);
        if (
          oldRange &&
          newRange &&
          (oldRange.start !== newRange.start || oldRange.end !== newRange.end)
        ) {
          try {
            await deletePortRulesApi({
              port_start: oldRange.start,
              port_end: oldRange.end,
            });
          } catch (e) {
            // 删除失败不应阻断保存，继续尝试保存新规则
            logError('Failed to delete old port rule before update:', e);
          }
        }
      }

      // 保存规则到服务器（函数内会切换 saving，这里再次置为 true 保持连续 loading）
      const ok = await savePortRule(rule);
      if (!ok) return; // stop on validation or api failure
      saving.value = true;

      // 刷新端口规则列表
      await fetchPortRules();

      showRuleForm.value = false;
      editingRule.value = null;

      Message.success(
        isEditing
          ? t('app.nftables.message.ruleUpdated')
          : t('app.nftables.message.ruleAdded')
      );
    } catch (error) {
      // 错误已在 savePortRule 中处理
    } finally {
      saving.value = false;
    }
  };

  const handleRuleCancel = (): void => {
    showRuleForm.value = false;
    editingRule.value = null;
  };

  const handleRuleDelete = async (ruleToDelete: PortRule): Promise<void> => {
    try {
      let portStart: number;
      let portEnd: number;
      if (typeof ruleToDelete.port === 'number') {
        portStart = ruleToDelete.port;
        portEnd = ruleToDelete.port;
      } else if (Array.isArray(ruleToDelete.port)) {
        const arr = ruleToDelete.port as number[];
        if (arr.length === 1) {
          portStart = arr[0];
          portEnd = arr[0];
        } else {
          portStart = Math.min(arr[0], arr[1]);
          portEnd = Math.max(arr[0], arr[1]);
        }
      } else {
        Message.error(t('app.nftables.message.operationFailed'));
        return;
      }

      await deletePortRulesApi({
        port_start: portStart,
        port_end: portEnd,
      });

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
    min-height: 100vh;
    background: var(--color-bg-1);
    .page-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1.714rem; // 24px -> rem
      border-bottom: 1px solid var(--color-border-2);
      .header-left {
        .page-title {
          margin: 0;
          font-size: 1.286rem;
          font-weight: 500;
          color: var(--color-text-1);
        }
        .page-description {
          margin: 0;
          font-size: 1rem; // 14px -> rem (root 14px)
          color: var(--color-text-3);
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
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 28.571rem; // 400px -> rem
        padding: 4.286rem 1.714rem; // 60px 24px -> rem
        .loading-text {
          margin-top: 1.143rem; // 16px -> rem
          font-size: 1rem; // 14px -> rem
          font-weight: 400;
          color: var(--color-text-3);
        }
      }
      .rules-management-container {
        padding: 1.143rem; // 16px -> rem
        margin: 1.143rem 1.714rem 0; // 16px 24px 0 -> rem
        background: var(--color-bg-2);
        border-radius: 0.429rem; // 6px -> rem
      }
      .error-state {
        margin-top: 2.857rem; // 40px -> rem
        text-align: center;
      }
    }
  }

  // 响应式设计
  @media (width <= 1024px) {
    .ports-management-page {
      .page-header {
        padding-right: 1.143rem; // 16px -> rem
        padding-left: 1.143rem; // 16px -> rem
      }
      .page-content {
        .rules-management-container {
          margin-right: 1.143rem; // 16px -> rem
          margin-left: 1.143rem; // 16px -> rem
        }
      }
    }
  }

  @media (max-width: @screen-md) {
    // 768px -> 使用系统断点变量
    .ports-management-page {
      .page-header {
        flex-direction: column;
        gap: 1.143rem; // 16px -> rem
        align-items: flex-start;
        padding-right: 0.857rem; // 12px -> rem
        padding-left: 0.857rem; // 12px -> rem
        .header-right {
          align-self: flex-end;
        }
      }
      .page-content {
        .rules-management-container {
          padding: 0.857rem; // 12px -> rem
          margin-right: 0.857rem; // 12px -> rem
          margin-left: 0.857rem; // 12px -> rem
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
