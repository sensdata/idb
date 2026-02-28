<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.certificateDetail')"
    :width="820"
    :footer="false"
    class="certificate-detail-drawer"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="certificate" class="certificate-detail">
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="1"
          bordered
        >
          <a-descriptions-item :label="$t('app.certificate.domain')">
            {{ certificate.domain }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.status')">
            <a-tag :color="getStatusColor()">
              {{ getStatusText() }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.notBefore')">
            {{ formatDate(certificate.not_before) }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.notAfter')">
            {{ formatDate(certificate.not_after) }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.keyAlgorithm')">
            {{ certificate.key_algorithm }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.keySize')">
            {{ certificate.key_size }} bits
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.isCA')">
            <a-tag
              :color="
                certificate.is_ca
                  ? 'rgb(var(--success-6))'
                  : 'rgb(var(--gray-6, 120,120,120))'
              "
            >
              {{ certificate.is_ca ? $t('common.yes') : $t('common.no') }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.source')">
            <div class="source-cell">
              <a-typography-text code class="source-text">{{
                certificateSource || '-'
              }}</a-typography-text>
              <div v-if="certificateSource" class="source-actions">
                <a-button type="text" size="small" @click="copySourcePath">
                  <template #icon>
                    <icon-copy />
                  </template>
                  {{ $t('app.certificate.copySourcePath') }}
                </a-button>
                <a-button type="text" size="small" @click="handleViewFile">
                  <template #icon>
                    <icon-folder />
                  </template>
                  {{ $t('app.certificate.viewFile') }}
                </a-button>
              </div>
            </div>
          </a-descriptions-item>
        </a-descriptions>

        <a-descriptions
          :title="$t('app.certificate.subjectInfo')"
          :column="1"
          bordered
          class="mt-6"
        >
          <a-descriptions-item :label="$t('app.certificate.country')">
            {{ certificate.country || '-' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.organization')">
            {{ certificate.organization || '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <a-descriptions
          :title="$t('app.certificate.issuerInfo')"
          :column="1"
          bordered
          class="mt-6"
        >
          <a-descriptions-item :label="$t('app.certificate.issuerCN')">
            {{ certificate.issuer_cn || '-' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('app.certificate.issuerCountry')">
            {{ certificate.issuer_country || '-' }}
          </a-descriptions-item>
          <a-descriptions-item
            :label="$t('app.certificate.issuerOrganization')"
          >
            {{ certificate.issuer_organization || '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <div
          v-if="certificate.alt_names && certificate.alt_names.length > 0"
          class="section-block"
        >
          <h4 class="section-title">{{ $t('app.certificate.altNames') }}</h4>
          <div class="alt-names-container">
            <a-tag v-for="altName in certificate.alt_names" :key="altName">
              {{ altName }}
            </a-tag>
          </div>
        </div>

        <div class="section-block">
          <h4 class="section-title">{{ $t('app.certificate.pemContent') }}</h4>
          <a-textarea
            :model-value="formattedPem"
            :auto-size="{ minRows: 10, maxRows: 10 }"
            readonly
            class="pem-content"
          />
        </div>

        <div class="action-bar">
          <a-button @click="copyPemContent">
            <template #icon>
              <icon-copy />
            </template>
            {{ $t('app.certificate.copyPEM') }}
          </a-button>
          <a-button type="primary" @click="handleCompleteChain">
            <template #icon>
              <icon-link />
            </template>
            {{ $t('app.certificate.completeChain') }}
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
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { IconCopy, IconLink, IconFolder } from '@arco-design/web-vue/es/icon';
  import { useClipboard } from '@/composables/use-clipboard';
  import { createFileRoute } from '@/utils/file-route';
  import type { CertificateInfo } from '@/api/certificate';

  interface Props {
    visible: boolean;
    certificate?: CertificateInfo | null;
    source?: string;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'completeChain'): void;
  }>();

  const { t } = useI18n();
  const router = useRouter();
  const { copyText } = useClipboard();

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  const getStatusColor = () => {
    if (!props.certificate) return 'rgb(var(--color-text-4))';

    const now = new Date();
    const notAfter = new Date(props.certificate.not_after);
    const notBefore = new Date(props.certificate.not_before);

    if (now > notAfter) {
      return 'rgb(var(--danger-6))';
    }
    if (now < notBefore) {
      return 'rgb(var(--warning-6))';
    }
    const daysUntilExpiry = Math.ceil(
      (notAfter.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
    );
    if (daysUntilExpiry <= 30) {
      return 'rgb(var(--warning-6))';
    }
    return 'rgb(var(--success-6))';
  };

  const getStatusText = () => {
    if (!props.certificate) return t('app.certificate.status.unknown');

    const now = new Date();
    const notAfter = new Date(props.certificate.not_after);
    const notBefore = new Date(props.certificate.not_before);

    if (now > notAfter) {
      return t('app.certificate.status.expired');
    }
    if (now < notBefore) {
      return t('app.certificate.status.notYetValid');
    }
    const daysUntilExpiry = Math.ceil(
      (notAfter.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
    );
    if (daysUntilExpiry <= 30) {
      return t('app.certificate.status.expiringSoon', {
        days: daysUntilExpiry,
      });
    }
    return t('app.certificate.status.valid');
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const formattedPem = computed(() => {
    const pem = props.certificate?.pem || '';
    return pem.includes('\\n') ? pem.replace(/\\n/g, '\n') : pem;
  });

  const copyPemContent = async () => {
    if (formattedPem.value) {
      try {
        await copyText(formattedPem.value);
        Message.success(t('app.certificate.copySuccess'));
      } catch (error) {
        Message.error(t('app.certificate.copyError'));
      }
    }
  };

  const handleCompleteChain = () => {
    emit('completeChain');
  };

  const getDirectoryPath = (filePath: string) => {
    const lastSlashIndex = filePath.lastIndexOf('/');
    return lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';
  };

  const certificateSource = computed(() => {
    return props.source || props.certificate?.source || '';
  });

  const copySourcePath = async () => {
    if (!certificateSource.value) {
      Message.warning(t('app.certificate.pathUnavailable'));
      return;
    }
    try {
      await copyText(certificateSource.value);
      Message.success(t('app.certificate.copySuccess'));
    } catch (error) {
      Message.error(t('app.certificate.copyError'));
    }
  };

  const handleViewFile = () => {
    if (certificateSource.value) {
      const directoryPath = getDirectoryPath(certificateSource.value);
      const route = createFileRoute(directoryPath);
      router.push(route);
      drawerVisible.value = false;
    }
  };
</script>

<style scoped>
  .certificate-detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-height: calc(100vh - 9rem);
    padding-bottom: 0.25rem;
    overflow-y: auto;
  }

  .alt-names-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .section-block {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .section-title {
    margin: 0;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .pem-content {
    font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
    font-size: 0.8125rem;
    line-height: 1.35;
  }

  :deep(.arco-descriptions-title) {
    margin-bottom: 0.75rem;
    font-size: 1rem;
    font-weight: 600;
  }

  .source-cell {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    align-items: flex-start;
  }

  .source-text {
    max-width: 100%;
    overflow-wrap: anywhere;
  }

  .source-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .action-bar {
    position: sticky;
    bottom: 0;
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    padding-top: 0.75rem;
    margin-top: 0.25rem;
    background: linear-gradient(
      to bottom,
      rgb(255 255 255 / 10%),
      var(--color-bg-1) 40%
    );
  }

  :deep(.certificate-detail-drawer .arco-drawer-body) {
    padding-bottom: 0.75rem;
  }
</style>
