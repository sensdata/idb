<template>
  <div class="certificate-group-container">
    <idb-table
      ref="tableRef"
      :loading="loading"
      :dataSource="dataSource"
      row-key="alias"
      :pagination="false"
      :columns="columns"
      @reload="handleReload"
    >
      <template #leftActions>
        <a-button type="primary" @click="$emit('reload')">
          <template #icon>
            <icon-refresh />
          </template>
          {{ $t('common.refresh') }}
        </a-button>
      </template>

      <template #alias="{ record }">
        <div class="alias-cell">
          <a-typography-text strong>{{ record.alias }}</a-typography-text>
          <a-typography-text type="secondary" class="requester">
            {{ $t('app.certificate.requester') }}: {{ record.requester }}
          </a-typography-text>
        </div>
      </template>

      <template #certificates="{ record }">
        <div class="certificates-cell">
          <div
            v-if="!record.certificates || record.certificates.length === 0"
            class="no-certificates"
          >
            <a-typography-text type="secondary">
              {{ $t('app.certificate.noCertificates') }}
            </a-typography-text>
          </div>
          <div v-else class="certificate-list">
            <div
              v-for="cert in record.certificates"
              :key="cert.source"
              class="certificate-item"
            >
              <div class="certificate-info">
                <a-typography-text strong>{{ cert.domain }}</a-typography-text>
                <div
                  v-if="cert.alt_names && cert.alt_names.length > 0"
                  class="alt-names"
                >
                  <a-tag
                    v-for="altName in cert.alt_names.slice(0, 3)"
                    :key="altName"
                    size="small"
                  >
                    {{ altName }}
                  </a-tag>
                  <a-tag v-if="cert.alt_names.length > 3" size="small">
                    +{{ cert.alt_names.length - 3 }}
                  </a-tag>
                </div>
              </div>
              <div class="certificate-meta">
                <a-tag :color="getCertificateStatusColor(cert.status)">
                  {{ getCertificateStatusText(cert.status) }}
                </a-tag>
                <a-typography-text type="secondary" class="expiry">
                  {{ $t('app.certificate.expiresOn') }}:
                  {{ formatDate(cert.not_after) }}
                </a-typography-text>
              </div>
              <div class="certificate-actions">
                <a-button
                  type="text"
                  size="small"
                  @click="$emit('viewDetail', cert.source)"
                >
                  {{ $t('app.certificate.viewDetail') }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  @click="$emit('completeChain', cert.source)"
                >
                  {{ $t('app.certificate.completeChain') }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  status="danger"
                  @click="$emit('deleteCertificate', cert.source)"
                >
                  {{ $t('common.delete') }}
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #operations="{ record }">
        <idb-table-operation
          type="button"
          :options="getOperationOptions(record)"
        />
      </template>
    </idb-table>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconRefresh } from '@arco-design/web-vue/es/icon';
  import type { Column } from '@/components/idb-table/types';
  import type { CertificateGroup } from '@/api/certificate';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';

  // Props 定义
  interface Props {
    loading: boolean;
    groups: CertificateGroup[];
  }

  const props = defineProps<Props>();

  // 事件定义
  const emit = defineEmits<{
    (e: 'reload'): void;
    (e: 'viewDetail', source: string): void;
    (e: 'generateSelfSigned', alias: string): void;
    (e: 'completeChain', source: string): void;
    (e: 'deleteGroup', alias: string): void;
    (e: 'deleteCertificate', source: string): void;
    (e: 'viewPrivateKey', alias: string): void;
    (e: 'viewCSR', alias: string): void;
  }>();

  const { t } = useI18n();

  // 表格列定义
  const columns = computed<Column[]>(() => [
    {
      title: t('app.certificate.alias'),
      dataIndex: 'alias',
      key: 'alias',
      width: 200,
      align: 'left',
      slotName: 'alias',
    },
    {
      title: t('app.certificate.certificates'),
      dataIndex: 'certificates',
      key: 'certificates',
      align: 'left',
      slotName: 'certificates',
    },
    {
      title: t('common.operations'),
      key: 'operations',
      width: 200,
      align: 'left',
      slotName: 'operations',
    },
  ]);

  // 数据源
  const dataSource = computed(() => ({
    items: props.groups,
    page: 1,
    page_size: props.groups.length,
    total: props.groups.length,
  }));

  // 获取操作选项
  const getOperationOptions = (record: CertificateGroup) => [
    {
      text: t('app.certificate.generateSelfSigned'),
      click: () => emit('generateSelfSigned', record.alias),
    },
    {
      text: t('app.certificate.viewPrivateKey'),
      click: () => emit('viewPrivateKey', record.alias),
    },
    {
      text: t('app.certificate.viewCSR'),
      click: () => emit('viewCSR', record.alias),
    },
    {
      text: t('app.certificate.deleteGroup'),
      status: 'danger' as const,
      click: () => emit('deleteGroup', record.alias),
    },
  ];

  // 处理刷新
  const handleReload = () => {
    emit('reload');
  };

  // 获取证书状态颜色
  const getCertificateStatusColor = (status: string) => {
    switch (status) {
      case 'valid':
        return 'green';
      case 'expired':
        return 'red';
      case 'expiring_soon':
        return 'orange';
      default:
        return 'gray';
    }
  };

  // 获取证书状态文本
  const getCertificateStatusText = (status: string) => {
    switch (status) {
      case 'valid':
        return t('app.certificate.status.valid');
      case 'expired':
        return t('app.certificate.status.expired');
      case 'expiring_soon':
        return t('app.certificate.status.expiringSoon');
      default:
        return t('app.certificate.status.unknown');
    }
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    try {
      const date = new Date(dateString);
      if (Number.isNaN(date.getTime())) return '-';
      return date.toLocaleDateString();
    } catch (error) {
      return '-';
    }
  };
</script>

<style scoped>
  .certificate-group-container {
    width: 100%;
  }

  .alias-cell {
    display: flex;
    flex-direction: column;
    gap: 0.33rem;
  }

  .requester {
    font-size: 1rem;
  }

  .certificates-cell {
    width: 100%;
  }

  .no-certificates {
    padding: 1.67rem 1.67rem 1.67rem 0;
    text-align: left;
  }

  .certificate-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .certificate-item {
    padding: 1rem;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
  }

  .certificate-info {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 0.67rem;
  }

  .alt-names {
    display: flex;
    flex-wrap: wrap;
    gap: 0.33rem;
  }

  .certificate-meta {
    display: flex;
    gap: 1rem;
    align-items: center;
    margin-bottom: 0.67rem;
  }

  .expiry {
    font-size: 1rem;
  }

  .certificate-actions {
    display: flex;
    gap: 0.67rem;
  }
</style>
