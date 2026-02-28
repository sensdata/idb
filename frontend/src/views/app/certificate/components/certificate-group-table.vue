<template>
  <div class="certificate-group-container">
    <div v-for="group in groups" :key="group.alias" class="group-card">
      <div class="group-card-header">
        <div class="group-meta">
          <div class="meta-item">
            <span class="meta-label">{{ $t('app.certificate.alias') }}:</span>
            <span class="meta-value">{{ group.alias }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label"
              >{{ $t('app.certificate.requester') }}:</span
            >
            <span class="meta-value">{{ getRequester(group.requester) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label"
              >{{ $t('app.certificate.certificates') }}:</span
            >
            <span class="meta-value">{{ getCertificateCount(group) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label"
              >{{ $t('app.certificate.expiresOn') }}:</span
            >
            <span class="meta-value">{{ getEarliestExpiryText(group) }}</span>
          </div>
          <div class="meta-item">
            <a-tag :color="getGroupStatusColor(group)">
              {{ getGroupStatusText(group) }}
            </a-tag>
          </div>
        </div>

        <div class="group-actions">
          <a-button
            type="text"
            size="small"
            class="action-link-btn"
            @click="$emit('importCertificate', group.alias)"
          >
            {{ $t('app.certificate.import') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            class="action-link-btn"
            @click="$emit('generateSelfSigned', group.alias)"
          >
            {{ $t('app.certificate.generateSelfSigned') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            class="action-link-btn"
            @click="$emit('viewPrivateKey', group.alias)"
          >
            {{ $t('app.certificate.viewPrivateKey') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            class="action-link-btn"
            @click="$emit('viewCsr', group.alias)"
          >
            {{ $t('app.certificate.viewCSR') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            status="danger"
            class="action-link-btn danger-link"
            @click="$emit('deleteGroup', group.alias)"
          >
            {{ $t('app.certificate.deleteGroup') }}
          </a-button>
        </div>
      </div>

      <div class="group-card-body">
        <a-table
          class="cert-table"
          :data="group.certificates || []"
          :loading="loading"
          :pagination="false"
          row-key="source"
          size="small"
          :bordered="false"
        >
          <template #columns>
            <a-table-column
              :title="$t('app.certificate.domain')"
              data-index="domain"
              :width="220"
            >
              <template #cell="{ record: cert }">
                <div class="cert-domain">{{ cert.domain || '-' }}</div>
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('app.certificate.altNames')"
              :width="260"
            >
              <template #cell="{ record: cert }">
                <div
                  v-if="cert.alt_names && cert.alt_names.length"
                  class="cert-alt-names"
                >
                  <a-tag
                    v-for="altName in cert.alt_names.slice(0, 2)"
                    :key="altName"
                    size="small"
                  >
                    {{ altName }}
                  </a-tag>
                  <a-tag v-if="cert.alt_names.length > 2" size="small">
                    +{{ cert.alt_names.length - 2 }}
                  </a-tag>
                </div>
                <span v-else>-</span>
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('app.certificate.expiresOn')"
              data-index="not_after"
              :width="180"
            >
              <template #cell="{ record: cert }">
                <span :class="getExpiryTextClass(cert.not_after)">
                  {{ getExpiryWithCountdown(cert.not_after) }}
                </span>
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('app.certificate.issuerCN')"
              :width="220"
            >
              <template #cell="{ record: cert }">
                {{ cert.issuer_cn || cert.issuer_organization || '-' }}
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('app.certificate.status')"
              data-index="status"
              :width="120"
              align="center"
            >
              <template #cell="{ record: cert }">
                <a-tag :color="getCertificateStatusColor(cert.status)">
                  {{ getCertificateStatusText(cert.status) }}
                </a-tag>
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('app.certificate.chainStatus')"
              :width="120"
              align="center"
            >
              <template #cell="{ record: cert }">
                <a-tag :color="getChainStatusColor(cert)">
                  {{ getChainStatusText(cert) }}
                </a-tag>
              </template>
            </a-table-column>

            <a-table-column
              :title="$t('common.operations')"
              :width="220"
              align="right"
            >
              <template #cell="{ record: cert }">
                <div class="cert-operations">
                  <a-button
                    type="text"
                    size="small"
                    class="action-link-btn"
                    @click="$emit('viewCertificate', cert.source)"
                  >
                    {{ $t('app.certificate.viewCertificate') }}
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    class="action-link-btn"
                    @click="$emit('completeChain', cert.source)"
                  >
                    {{ $t('app.certificate.completeChain') }}
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    class="action-link-btn"
                    @click="handleDeleteCertificate(cert.source)"
                  >
                    {{ $t('app.certificate.deleteCertificate') }}
                  </a-button>
                </div>
              </template>
            </a-table-column>
          </template>

          <template #empty>
            <a-empty :description="$t('app.certificate.noCertificates')" />
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import dayjs from 'dayjs';
  import type {
    CertificateGroup,
    CertificateSimpleInfo,
  } from '@/api/certificate';

  interface Props {
    loading: boolean;
    groups: CertificateGroup[];
  }

  const props = defineProps<Props>();

  const emit = defineEmits<{
    (e: 'viewCertificate', source: string): void;
    (e: 'generateSelfSigned', alias: string): void;
    (e: 'importCertificate', alias: string): void;
    (e: 'completeChain', source: string): void;
    (e: 'deleteGroup', alias: string): void;
    (e: 'deleteCertificate', source: string): void;
    (e: 'viewPrivateKey', alias: string): void;
    (e: 'viewCsr', alias: string): void;
  }>();

  const { t } = useI18n();

  const groups = computed(() => props.groups || []);

  const handleDeleteCertificate = (source: string) => {
    emit('deleteCertificate', source);
  };

  const getRequester = (requester?: string) => {
    const normalized = requester ? String(requester).trim() : '';
    return normalized || 'NULL';
  };

  const getCertificateCount = (group: CertificateGroup) =>
    (group.certificates || []).length;

  const getEarliestExpiry = (group: CertificateGroup) => {
    const certs = group.certificates || [];
    if (!certs.length) return null;

    const validDates = certs
      .map((cert) => dayjs(cert.not_after))
      .filter((date) => date.isValid())
      .sort((a, b) => a.valueOf() - b.valueOf());

    return validDates.length ? validDates[0] : null;
  };

  const getEarliestExpiryText = (group: CertificateGroup) => {
    const expiry = getEarliestExpiry(group);
    if (!expiry) return '-';

    const days = expiry.startOf('day').diff(dayjs().startOf('day'), 'day');
    if (days >= 0) {
      return `${expiry.format('YYYY/MM/DD')} (${days}d)`;
    }
    return `${expiry.format('YYYY/MM/DD')} (-${Math.abs(days)}d)`;
  };

  const getGroupStatus = (group: CertificateGroup) => {
    const certs = group.certificates || [];
    if (!certs.length) return 'unknown';

    if (certs.some((cert) => cert.status === 'expired')) return 'expired';
    if (certs.some((cert) => cert.status === 'expiring_soon')) {
      return 'expiring_soon';
    }
    if (certs.some((cert) => cert.status === 'valid')) return 'valid';
    return 'unknown';
  };

  function getCertificateStatusColor(status: string) {
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
  }

  function getCertificateStatusText(status: string) {
    switch (status) {
      case 'valid':
        return t('app.certificate.status.valid');
      case 'expired':
        return t('app.certificate.status.expired');
      case 'expiring_soon':
        return t('app.certificate.status.expiringSoon', { days: 30 });
      default:
        return t('app.certificate.status.unknown');
    }
  }

  const resolveChainCompletion = (cert: CertificateSimpleInfo) => {
    const candidates = [
      cert.chain_complete,
      cert.chain_completed,
      cert.is_chain_complete,
      cert.complete_chain,
    ];
    const raw = candidates.find(
      (value) => value !== undefined && value !== null
    );
    if (raw === undefined) return null;

    if (typeof raw === 'boolean') return raw;
    if (typeof raw === 'number') return raw === 1;
    if (typeof raw === 'string') {
      const value = raw.trim().toLowerCase();
      if (['1', 'true', 'yes', 'ok', 'completed'].includes(value)) return true;
      if (['0', 'false', 'no', 'none', 'incomplete'].includes(value))
        return false;
    }

    return null;
  };

  const getChainStatusColor = (cert: CertificateSimpleInfo) => {
    const completed = resolveChainCompletion(cert);
    if (completed === true) return 'green';
    if (completed === false) return 'orange';
    return 'gray';
  };

  const getChainStatusText = (cert: CertificateSimpleInfo) => {
    const completed = resolveChainCompletion(cert);
    if (completed === true) return t('app.certificate.chainStatus.completed');
    if (completed === false) return t('app.certificate.chainStatus.incomplete');
    return t('app.certificate.chainStatus.unknown');
  };

  const getGroupStatusColor = (group: CertificateGroup) => {
    return getCertificateStatusColor(getGroupStatus(group));
  };

  const getGroupStatusText = (group: CertificateGroup) => {
    return getCertificateStatusText(getGroupStatus(group));
  };

  const getExpiryWithCountdown = (dateString: string) => {
    if (!dateString) return '-';
    const date = dayjs(dateString);
    if (!date.isValid()) return '-';

    const days = date.startOf('day').diff(dayjs().startOf('day'), 'day');
    if (days >= 0) {
      return `${date.format('YYYY/MM/DD')} (${days}d)`;
    }
    return `${date.format('YYYY/MM/DD')} (-${Math.abs(days)}d)`;
  };

  const getExpiryTextClass = (dateString: string) => {
    if (!dateString) return '';
    const date = dayjs(dateString);
    if (!date.isValid()) return '';

    const days = date.startOf('day').diff(dayjs().startOf('day'), 'day');
    if (days < 0) return 'expiry-text expiry-text-expired';
    if (days <= 30) return 'expiry-text expiry-text-expiring';
    return 'expiry-text';
  };
</script>

<style scoped>
  .certificate-group-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
  }

  .group-card {
    overflow: hidden;
    background: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 0.5rem;
  }

  .group-card-header {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    justify-content: space-between;
    padding: 0.625rem 0.75rem;
    background: var(--color-fill-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  .group-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1rem;
    align-items: center;
  }

  .meta-item {
    display: flex;
    gap: 0.25rem;
    align-items: center;
  }

  .meta-label {
    font-size: 0.8125rem;
    color: var(--color-text-3);
  }

  .meta-value {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .group-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.125rem;
    align-items: center;
    justify-content: flex-end;
  }

  .group-card-body {
    padding: 0.5rem;
  }

  .cert-domain {
    font-weight: 600;
    color: var(--color-text-1);
  }

  .cert-alt-names {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .expiry-text {
    color: var(--color-text-1);
  }

  .expiry-text-expiring {
    font-weight: 600;
    color: rgb(var(--warning-6));
  }

  .expiry-text-expired {
    font-weight: 600;
    color: rgb(var(--danger-6));
  }

  .cert-operations {
    display: flex;
    flex-wrap: wrap;
    gap: 0.125rem;
    align-items: center;
    justify-content: flex-end;
  }

  .action-link-btn {
    color: rgb(var(--primary-6));
  }

  .action-link-btn.danger-link {
    color: rgb(var(--danger-6));
  }

  .action-link-btn.danger-link:hover {
    color: rgb(var(--danger-7));
  }

  :deep(.cert-table .arco-table-th) {
    background: var(--color-fill-1);
  }

  :deep(.cert-table .arco-table-td) {
    background: var(--color-bg-1);
  }
</style>
