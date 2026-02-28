<template>
  <div class="certificate-page-container">
    <div class="header-container">
      <div class="header-actions">
        <a-button type="primary" @click="showCreateGroupModal">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.certificate.createGroup') }}
        </a-button>
        <a-button @click="showImportModal()">
          <template #icon>
            <icon-import />
          </template>
          {{ $t('app.certificate.import') }}
        </a-button>
      </div>
    </div>

    <div class="content-container">
      <div class="certificate-content">
        <CertificateGroupTable
          :loading="loading"
          :groups="certificateGroups"
          @view-certificate="handleViewCertificate"
          @generate-self-signed="handleGenerateSelfSigned"
          @complete-chain="handleCompleteChain"
          @delete-group="handleDeleteGroup"
          @delete-certificate="handleDeleteCertificate"
          @view-private-key="handleViewPrivateKey"
          @view-csr="handleViewCSR"
          @import-certificate="handleImportCertificateModal"
        />
      </div>
    </div>

    <!-- 创建证书组抽屉 -->
    <CreateCertificateDrawer
      v-model:visible="createGroupModalVisible"
      :loading="submitLoading"
      @ok="handleCreateGroup"
    />

    <!-- 导入/更新证书抽屉 -->
    <UpdateCertificateModal
      v-model:visible="importModalVisible"
      :alias="selectedAlias"
      :loading="submitLoading"
      @ok="handleImportCertificate"
    />

    <!-- 生成自签名证书弹窗 -->
    <SelfSignedCertificateModal
      v-model:visible="selfSignedModalVisible"
      :alias="selectedAlias"
      :loading="submitLoading"
      @ok="handleGenerateSelfSignedCertificate"
    />

    <!-- 证书详情弹窗 -->
    <CertificateDetailModal
      v-model:visible="detailModalVisible"
      :certificate="selectedCertificate"
      :source="selectedSource"
      :loading="detailLoading"
      @complete-chain="handleCompleteChainFromDetail"
    />

    <!-- 私钥信息弹窗 -->
    <PrivateKeyModal
      v-model:visible="privateKeyModalVisible"
      :private-key="selectedPrivateKey"
      :loading="privateKeyLoading"
    />

    <!-- CSR信息弹窗 -->
    <CSRModal
      v-model:visible="csrModalVisible"
      :csr="selectedCSR"
      :loading="csrLoading"
    />
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { IconPlus, IconImport } from '@arco-design/web-vue/es/icon';
  import useCurrentHost from '@/composables/current-host';
  import { useApiWithLoading } from '@/composables/use-api-with-loading';
  import useLoading from '@/composables/loading';
  import { useLogger } from '@/composables/use-logger';
  import { useConfirm } from '@/composables/confirm';

  import {
    getCertificateGroups,
    createCertificateGroup,
    deleteCertificateGroup,
    deleteCertificate,
    generateSelfSignedCertificate,
    completeCertificateChain,
    getCertificateInfo,
    getPrivateKeyInfo,
    getCSRInfo,
    updateCertificate,
    type CertificateGroup,
    type CertificateInfo,
    type PrivateKeyInfo,
    type CSRInfo,
    type CreateGroupRequest,
    type SelfSignedRequest,
  } from '@/api/certificate';

  import CertificateGroupTable from './components/certificate-group-table.vue';
  import CreateCertificateDrawer from './components/create-certificate-drawer.vue';
  import UpdateCertificateModal from './components/update-certificate-modal.vue';
  import SelfSignedCertificateModal from './components/self-signed-certificate-modal.vue';
  import CertificateDetailModal from './components/certificate-detail-modal.vue';
  import PrivateKeyModal from './components/private-key-modal.vue';
  import CSRModal from './components/csr-modal.vue';

  const { t } = useI18n();
  const { currentHostId } = useCurrentHost();
  const { setLoading } = useLoading();
  const { executeApi } = useApiWithLoading(setLoading);
  const { logError } = useLogger();
  const { confirm } = useConfirm();

  const loading = ref(false);
  const submitLoading = ref(false);
  const detailLoading = ref(false);
  const privateKeyLoading = ref(false);
  const csrLoading = ref(false);

  const certificateGroups = ref<CertificateGroup[]>([]);

  const createGroupModalVisible = ref(false);
  const importModalVisible = ref(false);
  const selfSignedModalVisible = ref(false);
  const detailModalVisible = ref(false);
  const privateKeyModalVisible = ref(false);
  const csrModalVisible = ref(false);

  const selectedAlias = ref('');
  const selectedSource = ref('');
  const selectedCertificate = ref<CertificateInfo | null>(null);
  const selectedPrivateKey = ref<PrivateKeyInfo | null>(null);
  const selectedCSR = ref<CSRInfo | null>(null);

  const getHostIdOrNotify = (notify = true) => {
    const hostId = currentHostId.value;
    if (!hostId) {
      if (notify) {
        Message.error(t('app.certificate.error.noHost'));
      }
      return null;
    }
    return hostId as number;
  };

  const fetchCertificateGroups = async (notifyWhenNoHost = true) => {
    const hostId = getHostIdOrNotify(notifyWhenNoHost);
    if (!hostId) {
      return;
    }

    try {
      loading.value = true;
      const response = await executeApi(() => getCertificateGroups(hostId), {
        errorMessage: t('app.certificate.loadError'),
      });

      if (response && response.items) {
        certificateGroups.value = response.items.map((group) => ({
          ...group,
          certificates: group.certificates || [],
        }));
      } else {
        certificateGroups.value = [];
      }
    } catch (error) {
      logError('Failed to fetch certificate groups:', error);
    } finally {
      loading.value = false;
    }
  };

  const showCreateGroupModal = () => {
    createGroupModalVisible.value = true;
  };

  const showImportModal = (alias = '') => {
    selectedAlias.value = alias;
    importModalVisible.value = true;
  };

  const handleImportCertificateModal = (alias: string) => {
    showImportModal(alias);
  };

  const handleCreateGroup = async (form: CreateGroupRequest) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(() => createCertificateGroup(hostId, form), {
        onSuccess: () => {
          Message.success(t('app.certificate.createSuccess'));
          createGroupModalVisible.value = false;
          fetchCertificateGroups();
        },
        onError: (error) => {
          const errorMessage =
            error.message || t('app.certificate.createError');
          Message.error(errorMessage);
          logError('Failed to create certificate group:', error);
        },
      });
    } catch (error) {
      logError('Failed to create certificate group:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  const handleImportCertificate = async (formData: FormData) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(() => updateCertificate(hostId, formData), {
        onSuccess: () => {
          Message.success(t('app.certificate.importSuccess'));
          importModalVisible.value = false;
          fetchCertificateGroups();
        },
        onError: (error) => {
          const errorMessage =
            error.message || t('app.certificate.importError');
          Message.error(errorMessage);
          logError('Failed to import certificate:', error);
        },
      });
    } catch (error) {
      logError('Failed to import certificate:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  const handleGenerateSelfSigned = (alias: string) => {
    selectedAlias.value = alias;
    selfSignedModalVisible.value = true;
  };

  const handleGenerateSelfSignedCertificate = async (
    form: SelfSignedRequest
  ) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(() => generateSelfSignedCertificate(hostId, form), {
        onSuccess: () => {
          Message.success(t('app.certificate.generateSuccess'));
          selfSignedModalVisible.value = false;
          fetchCertificateGroups();
        },
        onError: (error) => {
          const errorMessage =
            error.message || t('app.certificate.generateError');
          Message.error(errorMessage);
          logError('Failed to generate self-signed certificate:', error);
        },
      });
    } catch (error) {
      logError('Failed to generate self-signed certificate:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  const handleViewCertificate = async (source: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      detailLoading.value = true;
      selectedSource.value = source;

      const response = await executeApi(
        () => getCertificateInfo(hostId, source),
        {
          errorMessage: t('app.certificate.loadDetailError'),
        }
      );

      if (response) {
        selectedCertificate.value = response;
        detailModalVisible.value = true;
      }
    } catch (error) {
      logError('Failed to get certificate info:', error);
    } finally {
      detailLoading.value = false;
    }
  };

  const handleCompleteChain = async (source: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      await executeApi(() => completeCertificateChain(hostId, { source }), {
        onSuccess: () => Message.success(t('app.certificate.completeSuccess')),
        errorMessage: t('app.certificate.completeError'),
      });

      await fetchCertificateGroups();
    } catch (error) {
      logError('Failed to complete certificate chain:', error);
    }
  };

  const handleCompleteChainFromDetail = async () => {
    if (selectedSource.value) {
      await handleCompleteChain(selectedSource.value);
      detailModalVisible.value = false;
    }
  };

  const handleDeleteGroup = async (alias: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    const confirmed = await confirm({
      title: t('app.certificate.deleteGroupConfirm.title'),
      content: t('app.certificate.deleteGroupConfirm.content', { alias }),
    });

    if (!confirmed) return;

    try {
      await executeApi(() => deleteCertificateGroup(hostId, { alias }), {
        onSuccess: () => Message.success(t('app.certificate.deleteSuccess')),
        errorMessage: t('app.certificate.deleteError'),
      });

      await fetchCertificateGroups();
    } catch (error) {
      logError('Failed to delete certificate group:', error);
    }
  };

  const handleDeleteCertificate = async (source: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    const confirmed = await confirm({
      title: t('app.certificate.deleteCertificateConfirm.title'),
      content: t('app.certificate.deleteCertificateConfirm.content'),
    });

    if (!confirmed) return;

    try {
      await executeApi(() => deleteCertificate(hostId, { source }), {
        onSuccess: () => Message.success(t('app.certificate.deleteSuccess')),
        errorMessage: t('app.certificate.deleteError'),
      });

      await fetchCertificateGroups();
    } catch (error) {
      logError('Failed to delete certificate:', error);
    }
  };

  const handleViewPrivateKey = async (alias: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      privateKeyLoading.value = true;

      const response = await executeApi(
        () => getPrivateKeyInfo(hostId, alias),
        {
          errorMessage: t('app.certificate.loadPrivateKeyError'),
        }
      );

      if (response) {
        selectedPrivateKey.value = response;
        privateKeyModalVisible.value = true;
      }
    } catch (error) {
      logError('Failed to get private key info:', error);
    } finally {
      privateKeyLoading.value = false;
    }
  };

  const handleViewCSR = async (alias: string) => {
    const hostId = getHostIdOrNotify();
    if (!hostId) {
      return;
    }

    try {
      csrLoading.value = true;

      const response = await executeApi(() => getCSRInfo(hostId, alias), {
        onError: () => {
          Message.error(t('app.certificate.loadCSRError'));
        },
      });

      if (response) {
        selectedCSR.value = response;
        csrModalVisible.value = true;
      }
    } catch (error) {
      selectedCSR.value = null;
      csrModalVisible.value = false;
      logError('Failed to get CSR info:', error);
    } finally {
      csrLoading.value = false;
    }
  };

  watch(
    () => currentHostId.value,
    () => {
      fetchCertificateGroups(false);
    },
    { immediate: true }
  );
</script>

<style scoped>
  .certificate-page-container {
    padding: 1.67rem;
  }

  .header-container {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-right: 1.67rem;
    margin-bottom: 1.67rem;
  }

  .page-title {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 1rem;
  }

  .content-container {
    padding: 1.67rem;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 0.67rem;
  }
</style>
