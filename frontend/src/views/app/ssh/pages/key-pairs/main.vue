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
          <a-tag :color="isEnabled(record) ? 'green' : 'gray'">
            {{
              isEnabled(record)
                ? $t('app.ssh.keyPairs.enabled')
                : $t('app.ssh.keyPairs.disabled')
            }}
          </a-tag>
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
      :key-bits-options="KEY_BITS_OPTIONS"
      :form="keyForm.formData"
      @confirm="handleGenerateConfirm"
      @update:visible="generateModalVisible = $event"
      @form-ref-updated="keyForm.setFormRef"
      @update:form="keyForm.updateForm"
    />
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, GlobalComponents } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/composables/loading';
  import { useForm } from '@/composables/use-form';
  import { useConfirm } from '@/composables/confirm';
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

  const ENCRYPTION_OPTIONS: EncryptionOption[] = [
    { label: 'RSA', value: 'rsa' },
    { label: 'ED25519', value: 'ed25519' },
    { label: 'ECDSA', value: 'ecdsa' },
    { label: 'DSA', value: 'dsa' },
  ];

  const KEY_BITS_OPTIONS: KeyBitsOption[] = [
    { label: '1024', value: 1024 },
    { label: '2048', value: 2048 },
  ];

  const { t } = useI18n();
  const hostStore = useHostStore();
  const { loading, setLoading } = useLoading(false);
  const generateModalVisible = ref<boolean>(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const { logError } = useLogger('KeyPairsMain');

  // 使用自定义Hook处理API调用的加载状态
  const { executeApi } = useApiWithLoading(setLoading);

  // 表单默认值
  const defaultFormState: GenerateKeyForm = {
    key_name: '',
    encryption_mode: 'rsa',
    key_bits: 2048,
    password: '',
    enable: true,
  };

  // 使用表单Hook
  const keyForm = useForm<GenerateKeyForm>({
    initialValues: defaultFormState,
    onSubmit: async (values) => {
      await generateSSHKey(hostStore.currentId as number, {
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

  // 文件下载工具函数
  const downloadFile = (data: Blob | any, filename: string): void => {
    const blob = data instanceof Blob ? data : new Blob([data]);
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename;
    link.click();
    URL.revokeObjectURL(link.href);
  };

  // 启用/禁用SSH密钥
  const handleToggleEnable = async (
    record: SSHKeyRecord,
    enabled: boolean
  ): Promise<void> => {
    await executeApi(
      async () => {
        await toggleSSHKeyEnabled(hostStore.currentId as number, {
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

  // 下载私钥
  const handleDownload = async (record: SSHKeyRecord): Promise<void> => {
    await executeApi(
      async () => {
        const response = await downloadSSHKey(
          hostStore.currentId as number,
          record.private_key_path
        );
        downloadFile(response.data, record.key_name);
        Message.success(t('app.ssh.keyPairs.downloadSuccess'));
      },
      { errorMessage: t('app.ssh.keyPairs.operationFailed') }
    );
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
          await deleteSSHKey(hostStore.currentId as number, record.key_name);
          Message.success(t('app.ssh.keyPairs.deleteSuccess'));
          gridRef.value?.reload();
        },
        { errorMessage: t('app.ssh.keyPairs.operationFailed') }
      );
    }
  };

  // 获取操作选项
  const getOperationOptions = (record: SSHKeyRecord) => [
    {
      text: t('app.ssh.keyPairs.download'),
      disabled: record.status !== SSHKeyStatus.ENABLED,
      click: () => handleDownload(record),
    },
    {
      text:
        record.status === SSHKeyStatus.ENABLED
          ? t('app.ssh.keyPairs.disable')
          : t('app.ssh.keyPairs.enable'),
      click: () =>
        handleToggleEnable(record, record.status !== SSHKeyStatus.ENABLED),
    },
    {
      text: t('app.ssh.keyPairs.delete'),
      status: 'danger' as const,
      click: () => handleDelete(record),
    },
  ];

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
      title: t('common.table.operation'),
      dataIndex: 'operation',
      width: 240,
      align: 'left' as const,
      slotName: 'operation',
      fixed: 'right' as const,
    },
  ]);

  // 获取SSH密钥列表
  const getSSHKeysApi = async (
    params: ApiListParams
  ): Promise<ApiListResult<SSHKeyRecord>> => {
    return executeApi(() => getSSHKeys(hostStore.currentId as number, params), {
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
</script>

<style scoped lang="less">
  .ssh-page-container {
    padding: 0 16px;
    background-color: var(--color-bg-2);
    border-radius: 6px;
    position: relative;
    border: 1px solid var(--color-border-2);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .header-container {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 12px 0;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 500;
    color: var(--color-text-1);
    margin: 0;
  }

  .content-container {
    background-color: #fff;
    border-radius: 4px;
    border: 1px solid var(--color-border-2);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    padding: 16px 20px;
    margin-bottom: 16px;
  }

  .ssh-password-table {
    width: 100%;
  }
</style>
