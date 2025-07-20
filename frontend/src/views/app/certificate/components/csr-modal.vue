<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.csrInfo')"
    :width="500"
    :footer="false"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="csr" class="csr-detail">
        <!-- 基本信息 -->
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="2"
          bordered
        >
          <a-descriptions-item :label="$t('app.certificate.commonName')">
            {{ csr.common_name }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.country')">
            {{ csr.country || '-' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.organization')">
            {{ csr.organization || '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <!-- 邮箱地址 -->
        <div v-if="csr.email_addresses.length > 0" class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.emailAddresses') }}</h4>
          <div class="email-addresses-container">
            <a-tag
              v-for="email in csr.email_addresses"
              :key="email"
              class="mb-2 mr-2"
            >
              {{ email }}
            </a-tag>
          </div>
        </div>

        <!-- CSR内容 -->
        <div class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.csrContent') }}</h4>
          <a-textarea
            :model-value="csr.pem"
            :rows="12"
            readonly
            class="font-mono text-sm"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="mt-6 flex justify-end gap-3">
          <a-button @click="copyCSR">
            <template #icon>
              <icon-copy />
            </template>
            {{ $t('app.certificate.copyCSR') }}
          </a-button>
          <a-button @click="drawerVisible = false">
            {{ $t('common.close') }}
          </a-button>
        </div>
      </div>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { IconCopy } from '@arco-design/web-vue/es/icon';
  import { useClipboard } from '@/composables/use-clipboard';
  import type { CSRInfo } from '@/api/certificate';

  // Props 定义
  interface Props {
    visible: boolean;
    csr?: CSRInfo | null;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  const { t } = useI18n();
  const { copyText } = useClipboard();

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 复制CSR内容
  const copyCSR = async () => {
    if (props.csr?.pem) {
      try {
        await copyText(props.csr.pem);
        Message.success(t('app.certificate.copySuccess'));
      } catch (error) {
        Message.error(t('app.certificate.copyError'));
      }
    }
  };
</script>

<style scoped>
  .csr-detail {
    max-height: 70vh;
    overflow-y: auto;
  }

  .email-addresses-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.67rem;
  }

  :deep(.arco-descriptions-title) {
    margin-bottom: 1rem;
    font-size: 1.33rem;
    font-weight: 600;
  }

  :deep(.arco-textarea) {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
  }

  .mt-6 {
    margin-top: 2rem;
  }

  .mb-2 {
    margin-bottom: 0.67rem;
  }

  .mr-2 {
    margin-right: 0.67rem;
  }

  .mb-3 {
    margin-bottom: 1rem;
  }

  h4 {
    margin: 0;
    font-size: 1.33rem;
    font-weight: 600;
  }
</style>
