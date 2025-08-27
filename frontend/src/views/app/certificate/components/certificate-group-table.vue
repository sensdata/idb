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
            <!-- Multiple certificates summary -->
            <div
              v-if="record.certificates.length > 1"
              class="certificates-summary"
            >
              <div class="summary-line">
                <span class="cert-count">
                  {{ record.certificates.length }}
                  {{ $t('app.certificate.certificatesCount') }}
                </span>
                <div class="status-tags">
                  <a-tag
                    v-for="status in getStatusSummary(record.certificates)"
                    :key="status.type"
                    :color="getCertificateStatusColor(status.type)"
                    size="small"
                  >
                    {{ getCertificateStatusText(status.type) }} ({{
                      status.count
                    }})
                  </a-tag>
                </div>
                <a-button
                  type="primary"
                  size="small"
                  class="view-details-btn"
                  @click="openCertificateDetailsModal(record)"
                >
                  {{ $t('app.certificate.viewDetails') }}
                </a-button>
              </div>
              <div class="domains-preview">
                {{ getDomainPreview(record.certificates) }}
              </div>
            </div>

            <!-- Single certificate display -->
            <div v-else class="single-certificate">
              <div
                v-for="cert in record.certificates"
                :key="cert.source"
                class="certificate-row"
              >
                <div class="cert-main-info">
                  <div class="domain-line">
                    <span class="domain-name">{{ cert.domain }}</span>
                    <a-tag
                      :color="getCertificateStatusColor(cert.status)"
                      size="small"
                    >
                      {{ getCertificateStatusText(cert.status) }}
                    </a-tag>
                  </div>
                  <div class="cert-meta">
                    <span class="expiry-info">
                      {{ $t('app.certificate.expiresOn') }}:
                      {{ formatDate(cert.not_after) }}
                    </span>
                    <div
                      v-if="cert.alt_names && cert.alt_names.length > 0"
                      class="alt-names-inline"
                    >
                      <a-tag
                        v-for="altName in cert.alt_names.slice(0, 2)"
                        :key="altName"
                        size="small"
                        class="alt-tag"
                      >
                        {{ altName }}
                      </a-tag>
                      <a-tag
                        v-if="cert.alt_names.length > 2"
                        size="small"
                        class="alt-tag"
                      >
                        +{{ cert.alt_names.length - 2 }}
                      </a-tag>
                    </div>
                  </div>
                </div>
                <div class="cert-actions">
                  <a-dropdown trigger="hover" position="bottom">
                    <a-button type="text" size="small">
                      <template #icon>
                        <icon-more />
                      </template>
                    </a-button>
                    <template #content>
                      <!-- 只有一个证书时，只显示必要的操作 -->
                      <template v-if="record.certificates.length === 1">
                        <a-doption
                          @click="$emit('updateCertificate', record.alias)"
                        >
                          {{ $t('app.certificate.updateCertificate') }}
                        </a-doption>
                        <a-doption
                          class="danger-option"
                          @click="handleDeleteCertificate(cert.source)"
                        >
                          {{ $t('common.delete') }}
                        </a-doption>
                      </template>
                      <!-- 多个证书时，显示所有操作 -->
                      <template v-else>
                        <a-doption @click="$emit('viewDetail', cert.source)">
                          {{ $t('app.certificate.viewDetail') }}
                        </a-doption>
                        <a-doption @click="$emit('completeChain', cert.source)">
                          {{ $t('app.certificate.completeChain') }}
                        </a-doption>
                        <a-doption
                          @click="$emit('updateCertificate', record.alias)"
                        >
                          {{ $t('app.certificate.updateCertificate') }}
                        </a-doption>
                        <a-doption
                          class="danger-option"
                          @click="handleDeleteCertificate(cert.source)"
                        >
                          {{ $t('common.delete') }}
                        </a-doption>
                      </template>
                    </template>
                  </a-dropdown>
                </div>
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

    <!-- Certificate Details Drawer -->
    <a-drawer
      v-model:visible="certificateDetailsDrawerVisible"
      :title="drawerTitle"
      :width="drawerWidth"
      placement="right"
      class="certificate-details-drawer"
      :mask-closable="true"
    >
      <div v-if="selectedGroup" class="drawer-content">
        <a-table
          :data="selectedGroup.certificates"
          :pagination="false"
          :bordered="false"
          size="small"
        >
          <template #columns>
            <a-table-column
              :title="$t('app.certificate.domain')"
              data-index="domain"
              :width="200"
            >
              <template #cell="{ record }">
                <div class="domain-cell">
                  <div class="domain-name">{{ record.domain }}</div>
                  <div class="domain-meta">
                    <a-tag
                      :color="getCertificateStatusColor(record.status)"
                      size="small"
                    >
                      {{ getCertificateStatusText(record.status) }}
                    </a-tag>
                    <span class="expiry-date">{{
                      formatDate(record.not_after)
                    }}</span>
                  </div>
                </div>
              </template>
            </a-table-column>
            <a-table-column
              :title="$t('common.operations')"
              :width="120"
              align="right"
            >
              <template #cell="{ record }">
                <div class="table-actions">
                  <a-button
                    type="text"
                    size="small"
                    @click="$emit('updateCertificate', selectedGroup.alias)"
                  >
                    {{ $t('common.edit') }}
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    status="danger"
                    @click="handleDeleteCertificate(record.source)"
                  >
                    {{ $t('common.delete') }}
                  </a-button>
                </div>
              </template>
            </a-table-column>
          </template>
        </a-table>
      </div>
    </a-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
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
    (e: 'updateCertificate', alias: string): void;
    (e: 'deleteGroup', alias: string): void;
    (e: 'deleteCertificate', source: string): void;
    (e: 'viewPrivateKey', alias: string): void;
    (e: 'viewCSR', alias: string): void;
  }>();

  const { t } = useI18n();

  // 证书详情抽屉状态
  const certificateDetailsDrawerVisible = ref(false);
  const selectedGroup = ref<CertificateGroup | null>(null);

  // 打开证书详情抽屉
  const openCertificateDetailsModal = (group: CertificateGroup) => {
    selectedGroup.value = group;
    certificateDetailsDrawerVisible.value = true;
  };

  // 抽屉标题
  const drawerTitle = computed(() => {
    if (!selectedGroup.value) return '';
    return `${selectedGroup.value.alias} - ${t(
      'app.certificate.certificates'
    )}`;
  });

  // 抽屉宽度
  const drawerWidth = computed(() => 700);

  // 处理删除证书
  const handleDeleteCertificate = (source: string) => {
    emit('deleteCertificate', source);
    // 不自动关闭抽屉，让用户手动关闭
  };

  // 监听groups变化，更新selectedGroup数据
  watch(
    () => props.groups,
    (newGroups: CertificateGroup[]) => {
      if (selectedGroup.value && newGroups) {
        // 找到对应的更新后的组数据
        const updatedGroup = newGroups.find(
          (group: CertificateGroup) =>
            group.alias === selectedGroup.value?.alias
        );
        if (updatedGroup) {
          selectedGroup.value = updatedGroup;
        }
      }
    },
    { deep: true }
  );

  // 获取状态汇总
  const getStatusSummary = (certificates: any[]) => {
    const statusCount = certificates.reduce((acc, cert) => {
      acc[cert.status] = (acc[cert.status] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    return Object.entries(statusCount).map(([type, count]) => ({
      type,
      count,
    }));
  };

  // 获取域名预览
  const getDomainPreview = (certificates: any[]) => {
    const domains = certificates.map((cert) => cert.domain);
    if (domains.length <= 2) {
      return domains.join(', ');
    }
    return `${domains.slice(0, 2).join(', ')} +${domains.length - 2}`;
  };

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
        return 'rgb(var(--success-6))';
      case 'expired':
        return 'rgb(var(--danger-6))';
      case 'expiring_soon':
        return 'rgb(var(--warning-6))';
      default:
        return 'rgb(var(--color-text-4))';
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
    width: 100%;
  }

  .certificates-summary {
    margin-bottom: 0.75rem;
  }

  .summary-line {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .cert-count {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-text-1);
    white-space: nowrap;
  }

  .status-tags {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 0.375rem;
  }

  .view-details-btn {
    padding: 0.25rem 0.75rem;
    font-size: 0.8125rem;
    white-space: nowrap;
  }

  .domains-preview {
    font-size: 0.8125rem;
    font-style: italic;
    line-height: 1.4;
    color: var(--color-text-3);
  }

  .single-certificate {
    margin-top: 0;
  }

  .certificate-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 0.5rem 0;
  }

  .cert-main-info {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.5rem;
  }

  .domain-line {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    align-items: center;
  }

  .domain-name {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .cert-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: center;
  }

  .expiry-info {
    font-size: 0.8125rem;
    color: var(--color-text-3);
  }

  .alt-names-inline {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .alt-tag {
    font-size: 0.75rem;
  }

  .cert-actions {
    flex-shrink: 0;
    align-self: flex-start;
    margin-left: 1rem;
  }

  /* Certificate Details Modal */
  .certificate-details-modal {
    max-height: 60vh;
    overflow-y: auto;
  }

  .modal-cert-item {
    padding: 1.5rem;
    margin-bottom: 1rem;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 0.5rem;
  }

  .modal-cert-item:last-child {
    margin-bottom: 0;
  }

  .modal-cert-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  .modal-domain-info {
    display: flex;
    gap: 0.75rem;
    align-items: center;
  }

  .modal-domain-name {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .modal-cert-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .modal-cert-details {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .modal-cert-meta {
    display: flex;
    gap: 1rem;
    align-items: center;
  }

  .modal-expiry-info {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-3);
    background: var(--color-fill-2);
    border-radius: 0.25rem;
  }

  .modal-alt-names {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .alt-names-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-2);
  }

  .alt-names-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
  }

  .modal-alt-tag {
    font-size: 0.75rem;
    color: var(--idblue-6);
    background: var(--idblue-1);
    border: 1px solid var(--idblue-2);
  }

  :deep(.arco-dropdown-option.danger-option) {
    color: var(--idbred-6);
  }

  :deep(.arco-dropdown-option.danger-option:hover) {
    color: var(--idbred-4);
    background-color: var(--idbred-1);
  }

  /* Certificate Details Drawer */
  .drawer-content {
    padding: 0;
  }

  .domain-cell {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .domain-name {
    font-weight: 500;
    color: var(--color-text-1);
  }

  .domain-meta {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .expiry-date {
    font-size: 0.75rem;
    color: var(--color-text-3);
  }

  .table-actions {
    display: flex;
    gap: 0.25rem;
    justify-content: flex-end;
  }

  /* Ensure drawer width is respected */
  :deep(.certificate-details-drawer .arco-drawer) {
    width: 700px !important;
    min-width: 700px !important;
    max-width: 700px !important;
  }

  :deep(.certificate-details-drawer .arco-drawer-content) {
    width: 700px !important;
    min-width: 700px !important;
    max-width: 700px !important;
  }

  :deep(.certificate-details-drawer .arco-drawer-wrapper) {
    width: 700px !important;
    min-width: 700px !important;
    max-width: 700px !important;
  }

  /* Override any responsive behavior */
  @media (width <= 768px) {
    :deep(.certificate-details-drawer .arco-drawer) {
      width: 700px !important;
      max-width: 700px !important;
    }
  }
</style>
