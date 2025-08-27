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
        <a-button @click="showImportModal">
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
          ref="tableRef"
          :loading="loading"
          :groups="certificateGroups"
          @reload="fetchCertificateGroups"
          @view-detail="handleViewDetail"
          @generate-self-signed="handleGenerateSelfSigned"
          @complete-chain="handleCompleteChain"
          @update-certificate="handleUpdateCertificateModal"
          @delete-group="handleDeleteGroup"
          @delete-certificate="handleDeleteCertificate"
          @view-private-key="handleViewPrivateKey"
          @view-csr="handleViewCSR"
        />
      </div>
    </div>

    <!-- 创建证书组抽屉 -->
    <CreateCertificateDrawer
      v-model:visible="createGroupModalVisible"
      :loading="submitLoading"
      @ok="handleCreateGroup"
    />

    <!-- 导入证书弹窗 -->
    <ImportCertificateModal
      v-model:visible="importModalVisible"
      :loading="submitLoading"
      @ok="handleImportCertificate"
    />

    <!-- 更新证书弹窗 -->
    <UpdateCertificateModal
      v-model:visible="updateModalVisible"
      :loading="submitLoading"
      :alias="selectedAlias"
      @ok="handleUpdateCertificate"
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
  import { ref, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { IconPlus, IconImport } from '@arco-design/web-vue/es/icon';
  import useHostStore from '@/store/modules/host';
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
    importCertificate,
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
  import ImportCertificateModal from './components/import-certificate-modal.vue';
  import UpdateCertificateModal from './components/update-certificate-modal.vue';
  import SelfSignedCertificateModal from './components/self-signed-certificate-modal.vue';
  import CertificateDetailModal from './components/certificate-detail-modal.vue';
  import PrivateKeyModal from './components/private-key-modal.vue';
  import CSRModal from './components/csr-modal.vue';

  const { t } = useI18n();
  const hostStore = useHostStore();
  const { setLoading } = useLoading();
  const { executeApi } = useApiWithLoading(setLoading);
  const { logError } = useLogger();
  const { confirm } = useConfirm();

  // 响应式数据
  const loading = ref(false);
  const submitLoading = ref(false);
  const detailLoading = ref(false);
  const privateKeyLoading = ref(false);
  const csrLoading = ref(false);

  const certificateGroups = ref<CertificateGroup[]>([]);

  // 弹窗状态
  const createGroupModalVisible = ref(false);
  const importModalVisible = ref(false);
  const updateModalVisible = ref(false);
  const selfSignedModalVisible = ref(false);
  const detailModalVisible = ref(false);
  const privateKeyModalVisible = ref(false);
  const csrModalVisible = ref(false);

  // 选中的数据
  const selectedAlias = ref('');
  const selectedSource = ref('');
  const selectedCertificate = ref<CertificateInfo | null>(null);
  const selectedPrivateKey = ref<PrivateKeyInfo | null>(null);
  const selectedCSR = ref<CSRInfo | null>(null);

  // 计算属性
  const currentHostId = computed(() => hostStore.current?.id);

  // 获取证书组列表
  const fetchCertificateGroups = async () => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      loading.value = true;
      const response = await executeApi(
        () => getCertificateGroups(currentHostId.value as number),
        {
          errorMessage: t('app.certificate.loadError'),
        }
      );

      if (response && response.items) {
        // Ensure each certificate group has a proper certificates array
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

  // 显示创建证书组弹窗
  const showCreateGroupModal = () => {
    createGroupModalVisible.value = true;
  };

  // 显示导入证书弹窗
  const showImportModal = () => {
    importModalVisible.value = true;
  };

  // 显示更新证书弹窗
  const handleUpdateCertificateModal = (alias: string) => {
    selectedAlias.value = alias;
    updateModalVisible.value = true;
  };

  // 处理创建证书组
  const handleCreateGroup = async (form: CreateGroupRequest) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(
        () => createCertificateGroup(currentHostId.value as number, form),
        {
          onSuccess: () => {
            Message.success(t('app.certificate.createSuccess'));
            // 只有成功时才关闭弹窗和刷新列表
            createGroupModalVisible.value = false;
            fetchCertificateGroups();
          },
          onError: (error) => {
            // 显示具体的错误信息
            const errorMessage =
              error.message || t('app.certificate.createError');
            Message.error(errorMessage);
            logError('Failed to create certificate group:', error);
          },
        }
      );
    } catch (error) {
      // executeApi 已经处理了错误，这里不需要额外处理
      logError('Failed to create certificate group:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  // 处理导入证书
  const handleImportCertificate = async (formData: FormData) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(
        () => importCertificate(currentHostId.value as number, formData),
        {
          onSuccess: () => {
            Message.success(t('app.certificate.importSuccess'));
            // 只有成功时才关闭弹窗和刷新列表
            importModalVisible.value = false;
            fetchCertificateGroups();
          },
          onError: (error) => {
            // 显示具体的错误信息
            const errorMessage =
              error.message || t('app.certificate.importError');
            Message.error(errorMessage);
            logError('Failed to import certificate:', error);
          },
        }
      );
    } catch (error) {
      // executeApi 已经处理了错误，这里不需要额外处理
      logError('Failed to import certificate:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  // 处理更新证书
  const handleUpdateCertificate = async (formData: FormData) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(
        () => updateCertificate(currentHostId.value as number, formData),
        {
          onSuccess: () => {
            Message.success(t('app.certificate.updateSuccess'));
            // 只有成功时才关闭弹窗和刷新列表
            updateModalVisible.value = false;
            fetchCertificateGroups();
          },
          onError: (error) => {
            // 显示具体的错误信息
            const errorMessage =
              error.message || t('app.certificate.updateError');
            Message.error(errorMessage);
            logError('Failed to update certificate:', error);
          },
        }
      );
    } catch (error) {
      // executeApi 已经处理了错误，这里不需要额外处理
      logError('Failed to update certificate:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  // 处理生成自签名证书
  const handleGenerateSelfSigned = (alias: string) => {
    selectedAlias.value = alias;
    selfSignedModalVisible.value = true;
  };

  const handleGenerateSelfSignedCertificate = async (
    form: SelfSignedRequest
  ) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      submitLoading.value = true;
      await executeApi(
        () =>
          generateSelfSignedCertificate(currentHostId.value as number, form),
        {
          onSuccess: () => {
            Message.success(t('app.certificate.generateSuccess'));
            // 只有成功时才关闭弹窗和刷新列表
            selfSignedModalVisible.value = false;
            fetchCertificateGroups();
          },
          onError: (error) => {
            // 显示具体的错误信息
            const errorMessage =
              error.message || t('app.certificate.generateError');
            Message.error(errorMessage);
            logError('Failed to generate self-signed certificate:', error);
          },
        }
      );
    } catch (error) {
      // executeApi 已经处理了错误，这里不需要额外处理
      logError('Failed to generate self-signed certificate:', error);
    } finally {
      submitLoading.value = false;
    }
  };

  // 处理查看证书详情
  const handleViewDetail = async (source: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      detailLoading.value = true;
      selectedSource.value = source;

      const response = await executeApi(
        () => getCertificateInfo(currentHostId.value as number, source),
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

  // 处理补齐证书链
  const handleCompleteChain = async (source: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      await executeApi(
        () =>
          completeCertificateChain(currentHostId.value as number, { source }),
        {
          onSuccess: () =>
            Message.success(t('app.certificate.completeSuccess')),
          errorMessage: t('app.certificate.completeError'),
        }
      );

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

  // 处理删除证书组
  const handleDeleteGroup = async (alias: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    const confirmed = await confirm({
      title: t('app.certificate.deleteGroupConfirm.title'),
      content: t('app.certificate.deleteGroupConfirm.content', { alias }),
    });

    if (!confirmed) return;

    try {
      await executeApi(
        () => deleteCertificateGroup(currentHostId.value as number, { alias }),
        {
          onSuccess: () => Message.success(t('app.certificate.deleteSuccess')),
          errorMessage: t('app.certificate.deleteError'),
        }
      );

      await fetchCertificateGroups();
    } catch (error) {
      logError('Failed to delete certificate group:', error);
    }
  };

  // 处理删除证书
  const handleDeleteCertificate = async (source: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    const confirmed = await confirm({
      title: t('app.certificate.deleteCertificateConfirm.title'),
      content: t('app.certificate.deleteCertificateConfirm.content'),
    });

    if (!confirmed) return;

    try {
      await executeApi(
        () => deleteCertificate(currentHostId.value as number, { source }),
        {
          onSuccess: () => Message.success(t('app.certificate.deleteSuccess')),
          errorMessage: t('app.certificate.deleteError'),
        }
      );

      await fetchCertificateGroups();
    } catch (error) {
      logError('Failed to delete certificate:', error);
    }
  };

  // 处理查看私钥
  const handleViewPrivateKey = async (alias: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      privateKeyLoading.value = true;

      const response = await executeApi(
        () => getPrivateKeyInfo(currentHostId.value as number, alias),
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

  // 处理查看CSR
  const handleViewCSR = async (alias: string) => {
    if (!currentHostId.value) {
      Message.error(t('app.certificate.error.noHost'));
      return;
    }

    try {
      csrLoading.value = true;

      const response = await executeApi(
        () => getCSRInfo(currentHostId.value as number, alias),
        {
          errorMessage: t('app.certificate.loadCSRError'),
        }
      );

      if (response) {
        selectedCSR.value = response;
        csrModalVisible.value = true;
      }
    } catch (error) {
      logError('Failed to get CSR info:', error);
    } finally {
      csrLoading.value = false;
    }
  };

  // 生命周期
  onMounted(() => {
    fetchCertificateGroups();
  });
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
