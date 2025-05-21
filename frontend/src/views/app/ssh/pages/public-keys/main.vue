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
  import { ref, onMounted, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useHostStore } from '@/store';
  import { useApiWithLoading } from '@/hooks/use-api-with-loading';
  import useLoading from '@/hooks/loading';
  import { useLogger } from '@/hooks/use-logger';
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

  // 状态管理
  const { loading, setLoading } = useLoading(false);
  const { executeApi } = useApiWithLoading(setLoading);
  const submitLoading = ref(false);
  const addKeyModalVisible = ref(false);
  const removeConfirmVisible = ref(false);
  const keys = ref<AuthKeyInfo[]>([]);
  const tableRef = ref();
  const currentRecord = ref<AuthKeyInfo | null>(null);

  // 获取SSH密钥
  const fetchKeys = async () => {
    if (!currentHostId.value) {
      Message.error(t('app.ssh.error.noHost'));
      return;
    }

    try {
      const response = await executeApi(
        () => getAuthorizedKeys(currentHostId.value as number),
        {
          errorMessage: t('app.ssh.publicKeys.loadError'),
        }
      );

      if (response && response.items) {
        keys.value = response.items.map((item: any) => {
          return {
            algorithm: item.algorithm || '',
            key: item.key || '',
            comment: item.comment || '',
            content:
              item.content ||
              `${item.algorithm} ${item.key} ${item.comment}`.trim(),
          };
        });
      } else {
        keys.value = [];
      }
    } catch (error) {
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

  // 确认删除
  const confirmRemove = async () => {
    if (currentRecord.value) {
      submitLoading.value = true;
      try {
        // 使用完整的密钥内容进行删除
        await deleteAuthorizedKey(
          currentHostId.value as number,
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
    submitLoading.value = true;
    try {
      await addAuthorizedKey(currentHostId.value as number, keyData.content);
      Message.success(t('app.ssh.publicKeys.addSuccess'));
      addKeyModalVisible.value = false;
      await fetchKeys();
    } catch (error) {
      Message.error(t('app.ssh.publicKeys.addError'));
    } finally {
      submitLoading.value = false;
    }
  };

  // 组件挂载时加载数据
  onMounted(() => {
    if (currentHostId.value) {
      fetchKeys();
    }
  });

  // 监听主机ID变化，重新加载数据
  watch(
    () => currentHostId.value,
    (newId) => {
      if (newId) {
        fetchKeys();
      }
    }
  );
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

  .public-keys-content {
    margin-top: 8px;
  }
</style>
