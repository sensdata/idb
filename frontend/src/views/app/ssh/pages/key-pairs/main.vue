<template>
  <div class="ssh-page-container">
    <div class="header-container">
      <h2 class="page-title">{{ $t('app.ssh.keyPairs.title') }}</h2>
    </div>
    <div class="content-container">
      <idb-table
        ref="gridRef"
        class="ssh-password-table"
        :loading="loading"
        :columns="columns"
        :fetch="getSSHKeysApi"
        :has-search="false"
        :pagination="false"
      >
        <template #leftActions>
          <a-button type="primary" @click="handleGenerateKey">
            {{ $t('app.ssh.keyPairs.generateKey') }}
          </a-button>
        </template>
        <template #enabled="{ record }">
          <a-tag
            :color="
              isEnabled(record) ? 'rgb(var(--success-6))' : 'rgb(var(--gray-7))'
            "
          >
            {{
              isEnabled(record)
                ? $t('app.ssh.keyPairs.enabled')
                : $t('app.ssh.keyPairs.disabled')
            }}
          </a-tag>
        </template>
        <template #security="{ record }">
          <a-tooltip :content="getSecurityAssessment(record).reason">
            <a-tag :color="getSecurityAssessment(record).color">
              {{ getSecurityAssessment(record).label }}
            </a-tag>
          </a-tooltip>
        </template>
        <template #operation="{ record }">
          <idb-table-operation
            type="button"
            :options="getOperationOptions(record)"
          />
        </template>
      </idb-table>
    </div>

    <KeyGenerateModal
      :visible="generateModalVisible"
      :loading="keyFormLoading"
      :encryption-options="ENCRYPTION_OPTIONS"
      :key-bits-options="KEY_BITS_OPTIONS_MAP"
      :form="keyForm.formData"
      @confirm="handleGenerateConfirm"
      @update:visible="generateModalVisible = $event"
      @form-ref-updated="keyForm.setFormRef"
      @update:form="keyForm.updateForm"
    />

    <a-modal
      v-model:visible="publicKeyModalVisible"
      :title="
        $t('app.ssh.keyPairs.publicKeyModal.title', {
          keyName: viewingPublicKeyName,
        })
      "
      :ok-loading="publicKeyLoading"
      :ok-text="$t('app.ssh.keyPairs.copyPublicKey')"
      :cancel-text="$t('common.close')"
      @ok="handleCopyCurrentPublicKey"
    >
      <a-textarea
        :model-value="viewingPublicKeyContent"
        readonly
        :auto-size="{ minRows: 6, maxRows: 12 }"
      />
    </a-modal>

    <a-modal
      v-model:visible="privateKeyModalVisible"
      :title="
        $t('app.ssh.keyPairs.privateKeyModal.title', {
          keyName: viewingPrivateKeyName,
        })
      "
      :footer="false"
    >
      <a-textarea
        :model-value="viewingPrivateKeyContent"
        readonly
        :auto-size="{ minRows: 8, maxRows: 16 }"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch, GlobalComponents } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/composables/loading';
  import { useForm } from '@/composables/use-form';
  import { useConfirm } from '@/composables/confirm';
  import { useClipboard } from '@/composables/use-clipboard';
  import useHostStore from '@/store/modules/host';
  import { ApiListParams, ApiListResult } from '@/types/global';
  import { useLogger } from '@/composables/use-logger';
  import {
    getSSHKeys,
    generateSSHKey,
    downloadSSHKey,
    toggleSSHKeyEnabled,
    deleteSSHKey,
  } from '@/api/ssh';
  import { useApiWithLoading } from '@/composables/use-api-with-loading';
  import {
    SSHKeyRecord,
    GenerateKeyForm,
    EncryptionMode,
    KeyBits,
    SSHKeyStatus,
  } from '@/views/app/ssh/types';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import KeyGenerateModal from './components/key-generate-modal.vue';

  interface EncryptionOption {
    label: string;
    value: EncryptionMode;
  }

  interface KeyBitsOption {
    label: string;
    value: KeyBits;
  }

  type SecurityLevel = 'recommended' | 'acceptable' | 'warning' | 'danger';

  interface SecurityAssessment {
    level: SecurityLevel;
    label: string;
    reason: string;
    color: string;
  }

  const ENCRYPTION_OPTIONS: EncryptionOption[] = [
    { label: 'ED25519', value: 'ed25519' },
    { label: 'RSA', value: 'rsa' },
    { label: 'ECDSA', value: 'ecdsa' },
  ];

  const KEY_BITS_OPTIONS_MAP: Record<EncryptionMode, KeyBitsOption[]> = {
    ed25519: [{ label: '256', value: 256 }],
    rsa: [
      { label: '3072', value: 3072 },
      { label: '2048', value: 2048 },
      { label: '4096', value: 4096 },
    ],
    ecdsa: [
      { label: '256', value: 256 },
      { label: '384', value: 384 },
      { label: '521', value: 521 },
    ],
  };

  const { t } = useI18n();
  const hostStore = useHostStore();
  const currentHostId = computed(
    () => hostStore.currentId as number | undefined
  );
  const { copyText } = useClipboard();
  const { loading, setLoading } = useLoading(false);
  const generateModalVisible = ref<boolean>(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const { logError } = useLogger('KeyPairsMain');
  const publicKeyModalVisible = ref(false);
  const publicKeyLoading = ref(false);
  const viewingPublicKeyContent = ref('');
  const viewingPublicKeyName = ref('');
  const privateKeyModalVisible = ref(false);
  const privateKeyLoading = ref(false);
  const viewingPrivateKeyContent = ref('');
  const viewingPrivateKeyName = ref('');

  // 使用自定义Hook处理API调用的加载状态
  const { executeApi } = useApiWithLoading(setLoading);

  // 表单默认值
  const defaultFormState: GenerateKeyForm = {
    key_name: '',
    encryption_mode: 'ed25519',
    key_bits: 256,
    password: '',
    enable: true,
  };

  // 使用表单Hook
  const keyForm = useForm<GenerateKeyForm>({
    initialValues: defaultFormState,
    onSubmit: async (values) => {
      if (!currentHostId.value) {
        throw new Error(t('app.ssh.error.noHost'));
      }

      await generateSSHKey(currentHostId.value, {
        key_name: values.key_name,
        encryption_mode: values.encryption_mode,
        key_bits: values.key_bits,
        password: values.password || undefined,
        enable: values.enable,
      });

      Message.success(t('app.ssh.keyPairs.generateSuccess'));
      generateModalVisible.value = false;
      gridRef.value?.reload();
    },
    onError: (error) => {
      Message.error(error.message);
    },
    validateMessage: t('app.ssh.keyPairs.generateValidationFailed'),
  });

  // 将 Ref<boolean> 转换为 boolean，解决类型错误
  const keyFormLoading = computed(() => keyForm.loading.value);

  // 检查SSH密钥是否启用
  const isEnabled = (record: SSHKeyRecord): boolean => {
    return record.status === SSHKeyStatus.ENABLED;
  };

  const getKeyAlgorithm = (record: SSHKeyRecord): string => {
    const source = `${record.key_name || ''} ${
      record.private_key_path || ''
    }`.toLowerCase();
    if (source.includes('ed25519')) return 'ed25519';
    if (source.includes('ecdsa')) return 'ecdsa';
    if (source.includes('rsa')) return 'rsa';
    if (source.includes('dsa') || source.includes('dss')) return 'dsa';
    return 'unknown';
  };

  const getSecurityAssessment = (record: SSHKeyRecord): SecurityAssessment => {
    const algorithm = getKeyAlgorithm(record);
    const bits = Number(record.key_bits) || 0;
    const colorMap: Record<SecurityLevel, string> = {
      recommended: 'rgb(var(--success-6))',
      acceptable: 'rgb(var(--arcoblue-6))',
      warning: 'rgb(var(--warning-6))',
      danger: 'rgb(var(--danger-6))',
    };

    if (algorithm === 'ed25519') {
      return {
        level: 'recommended',
        label: t('app.ssh.keyPairs.security.recommended'),
        reason: t('app.ssh.keyPairs.security.reason.ed25519'),
        color: colorMap.recommended,
      };
    }

    if (algorithm === 'ecdsa') {
      return {
        level: 'acceptable',
        label: t('app.ssh.keyPairs.security.acceptable'),
        reason: t('app.ssh.keyPairs.security.reason.ecdsa'),
        color: colorMap.acceptable,
      };
    }

    if (algorithm === 'rsa') {
      if (bits >= 3072) {
        return {
          level: 'recommended',
          label: t('app.ssh.keyPairs.security.recommended'),
          reason: t('app.ssh.keyPairs.security.reason.rsa3072'),
          color: colorMap.recommended,
        };
      }
      if (bits >= 2048) {
        return {
          level: 'acceptable',
          label: t('app.ssh.keyPairs.security.acceptable'),
          reason: t('app.ssh.keyPairs.security.reason.rsa2048'),
          color: colorMap.acceptable,
        };
      }
      return {
        level: 'danger',
        label: t('app.ssh.keyPairs.security.weak'),
        reason: t('app.ssh.keyPairs.security.reason.rsaWeak'),
        color: colorMap.danger,
      };
    }

    if (algorithm === 'dsa') {
      return {
        level: 'danger',
        label: t('app.ssh.keyPairs.security.weak'),
        reason: t('app.ssh.keyPairs.security.reason.dsa'),
        color: colorMap.danger,
      };
    }

    return {
      level: 'warning',
      label: t('app.ssh.keyPairs.security.unknown'),
      reason: t('app.ssh.keyPairs.security.reason.unknown'),
      color: colorMap.warning,
    };
  };

  const getPublicKeySourcePath = (record: SSHKeyRecord): string => {
    if (!record.private_key_path) return '';
    return record.private_key_path.endsWith('.pub')
      ? record.private_key_path
      : `${record.private_key_path}.pub`;
  };

  const fetchPublicKeyContent = async (
    record: SSHKeyRecord
  ): Promise<string> => {
    if (!currentHostId.value) {
      throw new Error(t('app.ssh.error.noHost'));
    }

    const source = getPublicKeySourcePath(record);
    if (!source) {
      throw new Error(t('app.ssh.keyPairs.publicKeyLoadFailed'));
    }

    const response = await downloadSSHKey(currentHostId.value, source);
    const fileData = response?.data;
    const blob = fileData instanceof Blob ? fileData : new Blob([fileData]);
    const content = (await blob.text()).trim();
    if (!content) {
      throw new Error(t('app.ssh.keyPairs.publicKeyLoadFailed'));
    }
    return content;
  };

  const handleViewPublicKey = async (record: SSHKeyRecord): Promise<void> => {
    publicKeyLoading.value = true;
    try {
      const content = await fetchPublicKeyContent(record);
      viewingPublicKeyName.value = record.key_name;
      viewingPublicKeyContent.value = content;
      publicKeyModalVisible.value = true;
    } catch (error) {
      logError('Failed to view public key:', error);
      Message.error(
        error instanceof Error
          ? error.message
          : t('app.ssh.keyPairs.publicKeyLoadFailed')
      );
    } finally {
      publicKeyLoading.value = false;
    }
  };

  const handleCopyPublicKey = async (record: SSHKeyRecord): Promise<void> => {
    publicKeyLoading.value = true;
    try {
      const content =
        viewingPublicKeyName.value === record.key_name &&
        viewingPublicKeyContent.value
          ? viewingPublicKeyContent.value
          : await fetchPublicKeyContent(record);
      await copyText(content);
      Message.success(t('app.ssh.keyPairs.publicKeyCopySuccess'));
    } catch (error) {
      logError('Failed to copy public key:', error);
      Message.error(t('app.ssh.keyPairs.publicKeyCopyFailed'));
    } finally {
      publicKeyLoading.value = false;
    }
  };

  const handleCopyCurrentPublicKey = async (): Promise<void> => {
    if (!viewingPublicKeyContent.value) {
      Message.error(t('app.ssh.keyPairs.publicKeyCopyFailed'));
      return;
    }

    publicKeyLoading.value = true;
    try {
      await copyText(viewingPublicKeyContent.value);
      Message.success(t('app.ssh.keyPairs.publicKeyCopySuccess'));
    } catch (error) {
      Message.error(t('app.ssh.keyPairs.publicKeyCopyFailed'));
    } finally {
      publicKeyLoading.value = false;
    }
  };

  // 启用/禁用SSH密钥
  const handleToggleEnable = async (
    record: SSHKeyRecord,
    enabled: boolean
  ): Promise<void> => {
    await executeApi(
      async () => {
        if (!currentHostId.value) {
          Message.error(t('app.ssh.error.noHost'));
          return;
        }

        await toggleSSHKeyEnabled(currentHostId.value, {
          key_name: record.key_name,
          enable: enabled,
        });

        // 不直接修改记录，而是重新加载数据
        gridRef.value?.reload();

        Message.success(
          enabled
            ? t('app.ssh.keyPairs.enableSuccess')
            : t('app.ssh.keyPairs.disableSuccess')
        );
      },
      { errorMessage: t('app.ssh.keyPairs.operationFailed') }
    );
  };

  const handleViewPrivateKey = async (record: SSHKeyRecord): Promise<void> => {
    privateKeyLoading.value = true;
    try {
      if (!currentHostId.value) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }
      if (!record.private_key_path) {
        Message.error(t('app.ssh.keyPairs.privateKeyLoadFailed'));
        return;
      }

      const response = await downloadSSHKey(
        currentHostId.value,
        record.private_key_path
      );
      const fileData = response?.data;
      const blob = fileData instanceof Blob ? fileData : new Blob([fileData]);
      const content = (await blob.text()).trim();
      if (!content) {
        Message.error(t('app.ssh.keyPairs.privateKeyLoadFailed'));
        return;
      }

      viewingPrivateKeyName.value = record.key_name;
      viewingPrivateKeyContent.value = content;
      privateKeyModalVisible.value = true;
    } catch (error) {
      logError('Failed to view private key:', error);
      Message.error(t('app.ssh.keyPairs.privateKeyLoadFailed'));
    } finally {
      privateKeyLoading.value = false;
    }
  };

  // 删除SSH密钥
  const { confirm } = useConfirm();
  const handleDelete = async (record: SSHKeyRecord): Promise<void> => {
    if (
      await confirm({
        content: t('app.ssh.keyPairs.deleteConfirm', {
          keyName: record.key_name,
        }),
      })
    ) {
      await executeApi(
        async () => {
          if (!currentHostId.value) {
            Message.error(t('app.ssh.error.noHost'));
            return;
          }

          await deleteSSHKey(currentHostId.value, record.key_name);
          Message.success(t('app.ssh.keyPairs.deleteSuccess'));
          gridRef.value?.reload();
        },
        { errorMessage: t('app.ssh.keyPairs.operationFailed') }
      );
    }
  };

  // 获取操作选项
  const getOperationOptions = (record: SSHKeyRecord) => {
    const enabled = record.status === SSHKeyStatus.ENABLED;

    return [
      {
        text: t('app.ssh.keyPairs.view'),
        type: 'text' as const,
        status: 'normal' as const,
        disabled: !record.private_key_path,
        click: () => handleViewPrivateKey(record),
      },
      {
        text: t('app.ssh.keyPairs.viewPublicKey'),
        disabled: !record.private_key_path,
        click: () => handleViewPublicKey(record),
      },
      {
        text: t('app.ssh.keyPairs.copyPublicKey'),
        disabled: !record.private_key_path,
        click: () => handleCopyPublicKey(record),
      },
      {
        text: enabled
          ? t('app.ssh.keyPairs.disable')
          : t('app.ssh.keyPairs.enable'),
        status: enabled ? ('warning' as const) : ('success' as const),
        click: () => handleToggleEnable(record, !enabled),
      },
      {
        text: t('app.ssh.keyPairs.delete'),
        status: 'danger' as const,
        click: () => handleDelete(record),
      },
    ];
  };

  // 表格列定义
  const columns = computed(() => [
    {
      title: t('app.ssh.keyPairs.columns.keyName'),
      dataIndex: 'key_name',
      width: 150,
    },
    {
      title: t('app.ssh.keyPairs.columns.user'),
      dataIndex: 'user',
      width: 120,
    },
    {
      title: t('app.ssh.keyPairs.columns.keyBits'),
      dataIndex: 'key_bits',
      width: 100,
    },
    {
      title: t('app.ssh.keyPairs.columns.keyPath'),
      dataIndex: 'private_key_path',
      width: 200,
    },
    {
      title: t('app.ssh.keyPairs.columns.fingerprint'),
      dataIndex: 'fingerprint',
      width: 180,
      ellipsis: true,
    },
    {
      title: t('app.ssh.keyPairs.columns.status'),
      dataIndex: 'status',
      width: 100,
      slotName: 'enabled',
    },
    {
      title: t('app.ssh.keyPairs.columns.security'),
      dataIndex: 'security',
      width: 120,
      slotName: 'security',
    },
    {
      title: t('common.table.operation'),
      dataIndex: 'operation',
      width: 320,
      align: 'left' as const,
      slotName: 'operation',
      fixed: 'right' as const,
    },
  ]);

  // 获取SSH密钥列表
  const getSSHKeysApi = async (
    params: ApiListParams
  ): Promise<ApiListResult<SSHKeyRecord>> => {
    const hostId = currentHostId.value;
    if (!hostId) {
      return {
        total: 0,
        items: [],
        page: params.page || 1,
        page_size: params.page_size || 20,
      };
    }

    return executeApi(() => getSSHKeys(hostId, params), {
      errorMessage: t('app.ssh.keyPairs.operationFailed'),
      defaultValue: {
        total: 0,
        items: [],
        page: params.page || 1,
        page_size: params.page_size || 20,
      },
    });
  };

  // 生成新密钥对
  const handleGenerateKey = (): void => {
    keyForm.resetForm();
    generateModalVisible.value = true;
  };

  const handleGenerateConfirm = async (): Promise<void> => {
    try {
      await keyForm.submitForm();
    } catch (error) {
      logError('Form submission failed:', error);
      Message.error(t('app.ssh.keyPairs.generateValidationFailed'));
    }
  };

  watch(currentHostId, () => {
    gridRef.value?.reload();
  });
</script>

<style scoped lang="less">
  .ssh-page-container {
    position: relative;
    padding: 0 16px;
    background-color: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    box-shadow: 0 2px 8px var(--color-fill-2);
  }

  .header-container {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 12px 0;
    margin-bottom: 16px;
  }

  .page-title {
    margin: 0;
    font-size: 18px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .content-container {
    padding: 16px 20px;
    margin-bottom: 16px;
    background-color: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
    box-shadow: 0 2px 8px var(--color-fill-2);
  }

  .ssh-password-table {
    width: 100%;
  }

  .ssh-password-table
    :deep(.idb-table-operation.type-button .arco-btn[disabled]) {
    color: var(--color-text-3) !important;
    cursor: not-allowed;
    opacity: 1;
  }

  .ssh-password-table
    :deep(
      .idb-table-operation.type-button .arco-btn[disabled] .arco-btn-content
    ) {
    color: inherit;
  }
</style>
