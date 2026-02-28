<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.csrInfo')"
    :width="820"
    :footer="false"
    class="csr-drawer"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="csr" class="csr-detail">
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="1"
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

        <div class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.csrContent') }}</h4>
          <a-textarea
            :model-value="formattedCsrPem"
            :rows="12"
            readonly
            class="font-mono text-sm"
          />
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <a-button :disabled="!formattedCsrPem" @click="copyCSR">
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

      <a-empty
        v-else
        class="csr-empty"
        :description="$t('app.certificate.csrUnavailable')"
      />
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

  interface Props {
    visible: boolean;
    csr?: CSRInfo | null;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  const { t } = useI18n();
  const { copyText } = useClipboard();

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  const formattedCsrPem = computed(() => {
    const pem = props.csr?.pem || '';
    return pem.includes('\\n') ? pem.replace(/\\n/g, '\n') : pem;
  });

  const copyCSR = async () => {
    if (!formattedCsrPem.value) {
      Message.warning(t('app.certificate.csrUnavailable'));
      return;
    }
    try {
      await copyText(formattedCsrPem.value);
      Message.success(t('app.certificate.copySuccess'));
    } catch (error) {
      Message.error(t('app.certificate.copyError'));
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

  .csr-empty {
    padding: 2.5rem 0 2rem;
  }

  :deep(.csr-drawer .arco-drawer-body) {
    padding-bottom: 0.75rem;
  }
</style>
