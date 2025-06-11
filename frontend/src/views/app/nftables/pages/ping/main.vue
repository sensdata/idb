<template>
  <div class="ping-configuration-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">{{ $t('app.nftables.ping.pageTitle') }}</h1>
      <p class="page-description">{{ $t('app.nftables.ping.description') }}</p>
    </div>

    <div class="page-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <a-spin :size="28" />
        <div class="loading-text">{{ $t('app.nftables.ping.loading') }}</div>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="hasError" class="error-state">
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

      <!-- 主要内容 -->
      <div v-else class="main-content">
        <!-- 当前状态卡片 -->
        <a-card
          class="status-card"
          :title="$t('app.nftables.ping.currentStatus')"
        >
          <div class="status-content">
            <div class="status-indicator">
              <a-badge
                :status="pingStatus.allowed ? 'success' : 'danger'"
                :text="
                  pingStatus.allowed
                    ? $t('app.nftables.ping.allowed')
                    : $t('app.nftables.ping.blocked')
                "
              />
            </div>
            <div class="status-description">
              {{
                pingStatus.allowed
                  ? $t('app.nftables.ping.statusDescription.allowed')
                  : $t('app.nftables.ping.statusDescription.blocked')
              }}
            </div>
          </div>
        </a-card>

        <!-- 配置设置卡片 -->
        <a-card
          class="config-card"
          :title="$t('app.nftables.ping.configureTitle')"
        >
          <div class="config-content">
            <a-form :model="formData" layout="vertical" @submit="handleSubmit">
              <a-form-item>
                <template #label>
                  <div class="form-label-section">
                    <div class="form-label">{{
                      $t('app.nftables.ping.enableLabel')
                    }}</div>
                    <div class="form-description">{{
                      $t('app.nftables.ping.enableHelp')
                    }}</div>
                  </div>
                </template>
                <div class="switch-wrapper">
                  <a-switch
                    v-model="formData.allowed"
                    :loading="saving"
                    :disabled="saving"
                    size="medium"
                  />
                </div>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  :loading="saving"
                  :disabled="!hasChanges"
                  @click="handleSubmit"
                >
                  <template #icon>
                    <icon-check />
                  </template>
                  {{
                    saving
                      ? $t('app.nftables.ping.saving')
                      : $t('app.nftables.ping.applySettings')
                  }}
                </a-button>
              </a-form-item>
            </a-form>
          </div>
        </a-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { IconCheck } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/hooks/use-logger';
  import type { PingStatus, SetPingStatusRequest } from '@/api/nftables';
  import { getPingStatusApi, setPingStatusApi } from '@/api/nftables';

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError } = useLogger('NftablesPingPage');

  // 响应式状态
  const loading = ref<boolean>(false);
  const saving = ref<boolean>(false);
  const hasError = ref<boolean>(false);
  const errorMessage = ref<string>('');
  const pingStatus = ref<PingStatus>({
    allowed: false,
  });

  // 表单数据
  const formData = ref<SetPingStatusRequest>({
    allowed: false,
  });

  // 计算属性：是否有变更
  const hasChanges = computed(() => {
    return formData.value.allowed !== pingStatus.value.allowed;
  });

  // 获取ping状态
  const fetchPingStatus = async (): Promise<void> => {
    try {
      loading.value = true;
      hasError.value = false;

      const response = await getPingStatusApi();
      pingStatus.value = response;
      formData.value.allowed = response.allowed;
    } catch (error) {
      logError('Failed to fetch ping status:', error);
      hasError.value = true;
      errorMessage.value = t('app.nftables.ping.loadFailed');
      Message.error(t('app.nftables.ping.loadFailed'));
    } finally {
      loading.value = false;
    }
  };

  // 设置ping状态
  const setPingStatus = async (allowed: boolean): Promise<void> => {
    try {
      saving.value = true;

      const requestData: SetPingStatusRequest = {
        allowed,
      };

      await setPingStatusApi(requestData);

      // 更新本地状态
      pingStatus.value.allowed = allowed;
      formData.value.allowed = allowed;

      Message.success(t('app.nftables.ping.saveSuccess'));
    } catch (error) {
      logError('Failed to set ping status:', error);
      Message.error(t('app.nftables.ping.saveFailed'));

      // 恢复原来的状态
      formData.value.allowed = pingStatus.value.allowed;
      throw error;
    } finally {
      saving.value = false;
    }
  };

  // 提交表单
  const handleSubmit = async (): Promise<void> => {
    if (!hasChanges.value) {
      return;
    }

    try {
      await setPingStatus(formData.value.allowed);
    } catch (error) {
      // 错误已在 setPingStatus 中处理
    }
  };

  // 重试
  const handleRetry = (): void => {
    fetchPingStatus();
  };

  // 组件挂载时获取数据
  onMounted(() => {
    fetchPingStatus();
  });
</script>

<style scoped lang="less">
  .ping-configuration-page {
    padding: 20px;
    background-color: var(--color-bg-1);
    min-height: 100vh;
  }

  .page-header {
    margin-bottom: 24px;

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
    max-width: 800px;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
    text-align: center;

    .loading-text {
      margin-top: 12px;
      color: var(--color-text-3);
      font-size: 14px;
    }
  }

  .error-state {
    margin-top: 40px;
  }

  .main-content {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .status-card {
    .status-content {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .status-indicator {
      :deep(.arco-badge-text) {
        font-size: 16px;
        font-weight: 500;
      }
    }

    .status-description {
      color: var(--color-text-2);
      font-size: 14px;
      line-height: 1.6;
    }
  }

  .config-card {
    .config-content {
      .form-label-section {
        .form-label {
          font-size: 14px;
          font-weight: 500;
          color: var(--color-text-1);
          margin-bottom: 4px;
        }

        .form-description {
          font-size: 12px;
          color: var(--color-text-3);
          line-height: 1.5;
          margin-bottom: 0;
        }
      }

      .switch-wrapper {
        margin-top: 8px;
      }

      :deep(.arco-btn) {
        margin-top: 16px;
      }
    }
  }

  // 响应式设计
  @media (max-width: 768px) {
    .ping-configuration-page {
      padding: 16px;
    }

    .page-header {
      .page-title {
        font-size: 20px;
      }
    }

    .main-content {
      gap: 16px;
    }
  }
</style>
