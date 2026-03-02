<template>
  <div class="ssh-page-container">
    <div class="header-container">
      <h2 class="page-title">{{ $t('app.ssh.publicKeys.title') }}</h2>
    </div>
    <div class="content-container">
      <div class="public-keys-content">
        <PublicKeysTable
          ref="tableRef"
          :loading="loading"
          :keys="keys"
          @reload="fetchKeys"
          @show-add-modal="showAddKeyModal"
          @view="handleView"
          @copy="handleCopy"
          @remove="handleRemove"
        />
      </div>
    </div>

    <!-- 添加公钥弹窗 -->
    <PublicKeyAddModal
      v-model:visible="addKeyModalVisible"
      :loading="submitLoading"
      @ok="handleAddKey"
    />

    <a-modal
      v-model:visible="viewModalVisible"
      :title="$t('app.ssh.publicKeys.viewModal.title')"
      :footer="false"
      role="dialog"
    >
      <a-textarea
        :model-value="viewingContent"
        readonly
        :auto-size="{ minRows: 6, maxRows: 12 }"
      />
    </a-modal>

    <!-- 确认删除弹窗 -->
    <a-modal
      v-model:visible="removeConfirmVisible"
      :title="$t('app.ssh.publicKeys.removeModal.title')"
      role="dialog"
      @ok="confirmRemove"
      @cancel="removeConfirmVisible = false"
    >
      <p>{{ $t('app.ssh.publicKeys.removeModal.content') }}</p>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useHostStore } from '@/store';
  import { useApiWithLoading } from '@/composables/use-api-with-loading';
  import { useClipboard } from '@/composables/use-clipboard';
  import useLoading from '@/composables/loading';
  import { useLogger } from '@/composables/use-logger';
  import {
    getAuthorizedKeys,
    addAuthorizedKey,
    deleteAuthorizedKey,
  } from '@/api/ssh';
  import { AuthKeyInfo, ParsedKey } from '@/views/app/ssh/types';
  import PublicKeysTable from './components/public-keys-table.vue';
  import PublicKeyAddModal from './components/public-key-add-modal.vue';

  const { t } = useI18n();
  const hostStore = useHostStore();
  const currentHostId = computed(() => hostStore.currentId);
  const { logError } = useLogger('PublicKeysMain');
  const { copyText } = useClipboard();

  // 状态管理
  const { loading, setLoading } = useLoading(false);
  const { executeApi } = useApiWithLoading(setLoading);
  const submitLoading = ref(false);
  const addKeyModalVisible = ref(false);
  const removeConfirmVisible = ref(false);
  const keys = ref<AuthKeyInfo[]>([]);
  const currentRecord = ref<AuthKeyInfo | null>(null);
  const fetchRequestId = ref(0);
  const viewModalVisible = ref(false);
  const viewingContent = ref('');
  const VALID_ALGORITHMS = new Set([
    'ssh-rsa',
    'ssh-ed25519',
    'ecdsa-sha2-nistp256',
    'ecdsa-sha2-nistp384',
    'ecdsa-sha2-nistp521',
    'ssh-dss',
  ]);

  const getFingerprint = async (keyBase64: string): Promise<string> => {
    try {
      const cleaned = keyBase64.trim();
      if (!cleaned) return '-';
      const padded = cleaned.padEnd(
        cleaned.length + ((4 - (cleaned.length % 4)) % 4),
        '='
      );
      const binary = atob(padded);
      const bytes = new Uint8Array(binary.length);
      for (let i = 0; i < binary.length; i += 1) {
        bytes[i] = binary.charCodeAt(i);
      }
      if (!window.crypto?.subtle) {
        return '-';
      }
      const digest = await window.crypto.subtle.digest('SHA-256', bytes);
      const digestBytes = new Uint8Array(digest);
      let digestBinary = '';
      for (let i = 0; i < digestBytes.length; i += 1) {
        digestBinary += String.fromCharCode(digestBytes[i]);
      }
      const base64 = btoa(digestBinary).replace(/=+$/g, '');
      return `SHA256:${base64}`;
    } catch (error) {
      return '-';
    }
  };

  const normalizeAuthorizedKey = async (
    item: any
  ): Promise<AuthKeyInfo | null> => {
    const raw = String(
      item.content ||
        `${item.algorithm || ''} ${item.key || ''} ${item.comment || ''}`
    ).trim();
    if (!raw || raw.startsWith('#')) {
      return null;
    }

    // authorized_keys may include options before algorithm.
    const parts = raw.split(/\s+/).filter(Boolean);
    const algorithmIndex = parts.findIndex((part) =>
      VALID_ALGORITHMS.has(part)
    );
    if (algorithmIndex < 0 || algorithmIndex + 1 >= parts.length) {
      return null;
    }

    const algorithm = parts[algorithmIndex];
    const key = parts[algorithmIndex + 1];
    if (!/^[A-Za-z0-9+/]+={0,2}$/.test(key)) {
      return null;
    }

    const comment = parts.slice(algorithmIndex + 2).join(' ');
    const content = raw;
    const fingerprint = await getFingerprint(key);

    return {
      algorithm,
      key,
      comment,
      fingerprint,
      content,
    };
  };

  // 获取SSH密钥
  const fetchKeys = async () => {
    if (!currentHostId.value) {
      keys.value = [];
      return;
    }

    const requestId = ++fetchRequestId.value;

    try {
      const response = await executeApi(
        () => getAuthorizedKeys(currentHostId.value as number),
        {
          errorMessage: t('app.ssh.publicKeys.loadError'),
        }
      );

      if (requestId !== fetchRequestId.value) {
        return;
      }

      if (response && response.items) {
        const parsed = await Promise.all(
          response.items.map((item: any) => normalizeAuthorizedKey(item))
        );
        if (requestId !== fetchRequestId.value) {
          return;
        }
        keys.value = parsed.filter(
          (item): item is AuthKeyInfo => item !== null
        );
      } else {
        keys.value = [];
      }
    } catch (error) {
      if (requestId !== fetchRequestId.value) {
        return;
      }
      logError('Failed to fetch keys:', error);
    }
  };

  // 显示添加密钥弹窗
  const showAddKeyModal = () => {
    addKeyModalVisible.value = true;
  };

  // 处理删除密钥
  const handleRemove = (record: any) => {
    currentRecord.value = record;
    removeConfirmVisible.value = true;
  };

  const handleView = (record: AuthKeyInfo) => {
    viewingContent.value = record.content || '';
    viewModalVisible.value = true;
  };

  const handleCopy = async (record: AuthKeyInfo) => {
    if (!record.content) {
      Message.error(t('app.ssh.publicKeys.copyError'));
      return;
    }

    try {
      await copyText(record.content);
      Message.success(t('app.ssh.publicKeys.copySuccess'));
    } catch (error) {
      Message.error(t('app.ssh.publicKeys.copyError'));
    }
  };

  // 确认删除
  const confirmRemove = async () => {
    if (currentRecord.value) {
      if (!currentHostId.value) {
        Message.error(t('app.ssh.error.noHost'));
        removeConfirmVisible.value = false;
        return;
      }

      submitLoading.value = true;
      try {
        // 使用完整的密钥内容进行删除
        await deleteAuthorizedKey(
          currentHostId.value,
          currentRecord.value.content ||
            `${currentRecord.value.algorithm} ${currentRecord.value.key} ${currentRecord.value.comment}`.trim()
        );
        Message.success(t('app.ssh.publicKeys.removeSuccess'));
        await fetchKeys();
      } catch (error) {
        Message.error(t('app.ssh.publicKeys.removeError'));
      } finally {
        submitLoading.value = false;
      }
    }
    removeConfirmVisible.value = false;
  };

  // 添加公钥
  const handleAddKey = async (keyData: {
    content: string;
    parsed: ParsedKey;
  }) => {
    if (!currentHostId.value) {
      Message.error(t('app.ssh.error.noHost'));
      return;
    }

    submitLoading.value = true;
    try {
      await addAuthorizedKey(currentHostId.value, keyData.content);
      Message.success(t('app.ssh.publicKeys.addSuccess'));
      addKeyModalVisible.value = false;
      await fetchKeys();
    } catch (error) {
      Message.error(t('app.ssh.publicKeys.addError'));
    } finally {
      submitLoading.value = false;
    }
  };

  // 监听主机ID变化，重新加载数据
  watch(
    () => currentHostId.value,
    (newId) => {
      if (newId) {
        fetchKeys();
      } else {
        keys.value = [];
      }
    },
    { immediate: true }
  );
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

  .public-keys-content {
    margin-top: 8px;
  }
</style>
