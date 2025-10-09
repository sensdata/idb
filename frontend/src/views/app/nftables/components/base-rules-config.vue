<template>
  <div class="base-rules-config">
    <!-- 当前策略显示 -->
    <div class="policy-display">
      <span class="policy-label"
        >{{ $t('app.nftables.baseRules.inputPolicy') }}:</span
      >
      <a-tag
        v-if="!loading"
        :color="getPolicyColor(currentPolicy)"
        class="policy-tag"
      >
        {{ getPolicyText(currentPolicy) }}
      </a-tag>
      <a-spin v-else :size="14" />
    </div>

    <!-- 配置按钮 -->
    <a-button
      type="text"
      size="small"
      :loading="saving"
      @click="showConfigModal = true"
    >
      <template #icon>
        <icon-settings />
      </template>
      {{ $t('common.button.config') }}
    </a-button>

    <!-- 配置模态框 -->
    <a-modal
      v-model:visible="showConfigModal"
      :title="$t('app.nftables.baseRules.title')"
      :width="480"
      :mask-closable="false"
      @ok="handleSave"
      @cancel="handleCancel"
    >
      <div class="config-content">
        <div class="config-description">
          <p>{{ $t('app.nftables.baseRules.inputPolicyDescription') }}</p>
        </div>

        <div class="policy-options">
          <a-radio-group v-model="selectedPolicy" direction="vertical">
            <a-radio value="accept">
              {{ $t('app.nftables.baseRules.accept') }}
              <icon-check-circle class="policy-icon accept" />
            </a-radio>

            <a-radio value="drop">
              {{ $t('app.nftables.baseRules.drop') }}
              <icon-close-circle class="policy-icon drop" />
            </a-radio>
          </a-radio-group>

          <!-- 动态显示说明文字 -->
          <div
            v-if="selectedPolicy"
            class="policy-description"
            style="
              padding: 1.143rem 1.429rem;
              margin-top: 1.714rem;
              background-color: var(--color-fill-1);
              border: 0.071rem solid var(--color-border-2);
              border-left: 0.286rem solid var(--color-primary-6);
              border-radius: 0.571rem;
            "
          >
            <div
              v-if="selectedPolicy === 'accept'"
              class="desc-content"
              style="
                font-size: 0.929rem;
                line-height: 1.6;
                color: var(--color-text-3);
                opacity: 0.9;
              "
            >
              💡 {{ $t('app.nftables.baseRules.acceptDesc') }}
            </div>
            <div
              v-else-if="selectedPolicy === 'drop'"
              class="desc-content"
              style="
                font-size: 0.929rem;
                line-height: 1.6;
                color: var(--color-text-3);
                opacity: 0.9;
              "
            >
              💡 {{ $t('app.nftables.baseRules.dropDesc') }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <a-button @click="handleCancel">
          {{ $t('common.button.cancel') }}
        </a-button>
        <a-button
          type="primary"
          :loading="saving"
          :disabled="selectedPolicy === currentPolicy"
          @click="handleSave"
        >
          {{ $t('common.button.save') }}
        </a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/composables/use-logger';
  import type { SetBaseRulesRequest } from '@/api/nftables';
  import { getBaseRulesApi, setBaseRulesApi } from '@/api/nftables';
  import {
    IconSettings,
    IconCheckCircle,
    IconCloseCircle,
  } from '@arco-design/web-vue/es/icon';
  import { useNftablesConfig } from '../composables/use-nftables-config';

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError, logInfo } = useLogger('BaseRulesConfig');

  // 获取配置应用后的刷新方法（关闭自动请求避免重复）
  const { handleConfigApplied } = useNftablesConfig({ autoFetch: false });

  // 响应式状态
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const showConfigModal = ref<boolean>(false);
  const currentPolicy = ref<'drop' | 'accept'>('accept');
  const selectedPolicy = ref<'drop' | 'accept'>('accept');

  // 获取策略颜色
  const getPolicyColor = (policy: string) => {
    switch (policy) {
      case 'accept':
        return 'green';
      case 'drop':
        return 'red';
      default:
        return 'gray';
    }
  };

  // 获取策略文本
  const getPolicyText = (policy: string) => {
    switch (policy) {
      case 'accept':
        return t('app.nftables.baseRules.accept');
      case 'drop':
        return t('app.nftables.baseRules.drop');
      default:
        return policy;
    }
  };

  // 获取基础规则
  const fetchBaseRules = async (): Promise<void> => {
    try {
      loading.value = true;

      const response = await getBaseRulesApi();
      currentPolicy.value = response.input_policy;
      selectedPolicy.value = response.input_policy;

      logInfo('Base rules fetched successfully:', response);
    } catch (error) {
      logError('Failed to fetch base rules:', error);
      Message.error(t('app.nftables.message.fetchBaseRulesFailed'));
    } finally {
      loading.value = false;
    }
  };

  // 保存基础规则
  const saveBaseRules = async (): Promise<void> => {
    try {
      saving.value = true;

      const request: SetBaseRulesRequest = {
        input_policy: selectedPolicy.value,
      };

      await setBaseRulesApi(request);
      currentPolicy.value = selectedPolicy.value;

      Message.success(t('app.nftables.message.baseRulesSaved'));
      logInfo('Base rules saved successfully:', request);

      // 配置应用成功后刷新当前配置列表
      await handleConfigApplied();
    } catch (error) {
      logError('Failed to save base rules:', error);
      Message.error(t('app.nftables.message.saveBaseRulesFailed'));
      throw error;
    } finally {
      saving.value = false;
    }
  };

  // 事件处理
  const handleSave = async (): Promise<void> => {
    try {
      await saveBaseRules();
      showConfigModal.value = false;
    } catch (error) {
      // 错误已在 saveBaseRules 中处理
    }
  };

  const handleCancel = (): void => {
    selectedPolicy.value = currentPolicy.value;
    showConfigModal.value = false;
  };

  // 页面初始化
  onMounted(async () => {
    await fetchBaseRules();
  });
</script>

<style scoped lang="less">
  .base-rules-config {
    display: flex;
    align-items: center;
    gap: 0.857rem;

    .policy-display {
      display: flex;
      align-items: center;
      gap: 0.429rem;
      font-size: 1rem;

      .policy-label {
        color: var(--color-text-2);
        white-space: nowrap;
      }

      .policy-tag {
        font-size: 0.857rem;
        font-weight: 500;
      }
    }

    .config-content {
      .config-description {
        margin-bottom: 1.429rem;
        padding: 0.857rem;
        background: var(--color-fill-1);
        border-radius: 0.429rem;

        p {
          margin: 0;
          font-size: 1rem;
          color: var(--color-text-2);
          line-height: 1.5;
        }
      }

      .policy-options {
        margin-bottom: 1.143rem;

        :deep(.arco-radio) {
          margin-bottom: 0.857rem;
          display: flex;
          align-items: center;
          gap: 0.571rem;

          &:last-child {
            margin-bottom: 1.143rem;
          }

          .arco-radio-label {
            display: flex;
            align-items: center;
            gap: 0.571rem;
            padding-left: 0.571rem;
          }
        }

        .policy-icon {
          font-size: 1.143rem;
          flex-shrink: 0;

          &.accept {
            color: var(--color-success-6);
          }

          &.drop {
            color: var(--color-danger-6);
          }
        }

        .policy-description {
          margin-top: 1.714rem;
          padding: 1.143rem 1.429rem;
          background-color: var(--color-fill-1);
          border-radius: 0.571rem;
          border: 0.071rem solid var(--color-border-2);
          position: relative;

          &::before {
            content: '';
            position: absolute;
            left: 0;
            top: 0;
            bottom: 0;
            width: 0.286rem;
            background-color: var(--color-primary-6);
            border-radius: 0.286rem 0 0 0.286rem;
          }

          .desc-content {
            font-size: 0.929rem;
            color: var(--color-text-3);
            line-height: 1.6;
            opacity: 0.9;

            &::before {
              content: '💡 ';
              margin-right: 0.429rem;
            }
          }
        }
      }
    }
  }

  // 响应式设计
  @media (max-width: 54.857rem) {
    .base-rules-config {
      .policy-display {
        .policy-label {
          display: none;
        }
      }
    }
  }
</style>
