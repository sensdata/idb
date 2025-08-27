<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.certificateDetail')"
    :width="700"
    :footer="false"
  >
    <a-spin :loading="loading" class="w-full">
      <div v-if="certificate" class="certificate-detail">
        <!-- 基本信息 -->
        <a-descriptions
          :title="$t('app.certificate.basicInfo')"
          :column="2"
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
            <a-typography-text code>{{ certificate.source }}</a-typography-text>
          </a-descriptions-item>
        </a-descriptions>

        <!-- 主体信息 -->
        <a-descriptions
          :title="$t('app.certificate.subjectInfo')"
          :column="2"
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

        <!-- 签发机构信息 -->
        <a-descriptions
          :title="$t('app.certificate.issuerInfo')"
          :column="2"
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
            :span="2"
          >
            {{ certificate.issuer_organization || '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <!-- 备用域名 -->
        <div v-if="certificate.alt_names.length > 0" class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.altNames') }}</h4>
          <div class="alt-names-container">
            <a-tag
              v-for="altName in certificate.alt_names"
              :key="altName"
              class="mb-2 mr-2"
            >
              {{ altName }}
            </a-tag>
          </div>
        </div>

        <!-- 证书内容 -->
        <div class="mt-6">
          <h4 class="mb-3">{{ $t('app.certificate.pemContent') }}</h4>
          <a-textarea
            :model-value="certificate.pem"
            :rows="10"
            readonly
            class="font-mono text-sm"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="mt-6 flex justify-end gap-3">
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
  import { Message } from '@arco-design/web-vue';
  import { IconCopy, IconLink } from '@arco-design/web-vue/es/icon';
  import { useClipboard } from '@/composables/use-clipboard';
  import type { CertificateInfo } from '@/api/certificate';

  // Props 定义
  interface Props {
    visible: boolean;
    certificate?: CertificateInfo | null;
    source?: string;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'completeChain'): void;
  }>();

  const { t } = useI18n();
  const { copyText } = useClipboard();

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 获取证书状态颜色
  const getStatusColor = () => {
    if (!props.certificate) return 'rgb(var(--color-text-4))';

    const now = new Date();
    const notAfter = new Date(props.certificate.not_after);
    const notBefore = new Date(props.certificate.not_before);

    if (now > notAfter) {
      return 'rgb(var(--danger-6))'; // 已过期
    }
    if (now < notBefore) {
      return 'rgb(var(--warning-6))'; // 尚未生效
    }
    const daysUntilExpiry = Math.ceil(
      (notAfter.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
    );
    if (daysUntilExpiry <= 30) {
      return 'rgb(var(--warning-6))'; // 即将过期
    }
    return 'rgb(var(--success-6))'; // 有效
  };

  // 获取证书状态文本
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

  // 格式化日期
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  // 复制PEM内容
  const copyPemContent = async () => {
    if (props.certificate?.pem) {
      try {
        await copyText(props.certificate.pem);
        Message.success(t('app.certificate.copySuccess'));
      } catch (error) {
        Message.error(t('app.certificate.copyError'));
      }
    }
  };

  // 处理补齐证书链
  const handleCompleteChain = () => {
    emit('completeChain');
  };
</script>

<style scoped>
  .certificate-detail {
    max-height: 70vh;
    overflow-y: auto;
  }

  .alt-names-container {
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
