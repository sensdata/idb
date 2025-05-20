<template>
  <div class="auth-key-container">
    <idb-table
      ref="tableRef"
      :loading="loading"
      :dataSource="dataSource"
      row-key="key"
      :pagination="false"
      :columns="columns"
    >
      <template #leftActions>
        <a-button
          type="primary"
          class="generate-key-btn"
          @click="showAddKeyModal"
        >
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.ssh.authKey.add') }}
        </a-button>
      </template>

      <template #operations="{ record }">
        <a-space>
          <a-button type="text" status="danger" @click="handleRemove(record)">
            {{ $t('app.ssh.authKey.remove') }}
          </a-button>
        </a-space>
      </template>
    </idb-table>

    <!-- 添加密钥弹窗 -->
    <a-modal
      v-model:visible="addKeyModalVisible"
      :title="$t('app.ssh.authKey.modal.title')"
      role="dialog"
      @ok="handleAddKey"
      @cancel="addKeyModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <div id="key-content-label" class="modal-label">{{
            $t('app.ssh.authKey.modal.content')
          }}</div>
          <div class="modal-input-wrapper">
            <a-textarea
              v-model="addKeyForm.content"
              :placeholder="$t('app.ssh.authKey.modal.placeholder')"
              :auto-size="{ minRows: 4, maxRows: 8 }"
              aria-labelledby="key-content-label"
              :status="
                addKeyForm.validated && addKeyForm.error ? 'error' : undefined
              "
            />
            <div id="key-description" class="modal-field-description">
              {{ addKeyForm.error || $t('app.ssh.authKey.modal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <!-- 确认删除弹窗 -->
    <a-modal
      v-model:visible="removeConfirmVisible"
      :title="$t('app.ssh.authKey.removeModal.title')"
      role="dialog"
      @ok="confirmRemove"
      @cancel="removeConfirmVisible = false"
    >
      <p>{{ $t('app.ssh.authKey.removeModal.content') }}</p>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { IconPlus } from '@arco-design/web-vue/es/icon';
  import type { Column } from '@/components/idb-table/types';
  import type { ApiListResult } from '@/types/global';

  const { t } = useI18n();

  // SSH密钥验证逻辑
  const useSSHKeyValidator = () => {
    // 密钥表单状态
    const keyForm = reactive({
      content: '',
      validated: false,
      error: '',
      parsed: {
        algorithm: '',
        key: '',
        comment: '',
      },
    });

    // 验证密钥内容
    const validateKey = () => {
      keyForm.validated = true;

      // 验证是否为空
      if (!keyForm.content.trim()) {
        keyForm.error = t('app.ssh.authKey.modal.emptyError');
        return false;
      }

      // 解析密钥内容
      const parts = keyForm.content.trim().split(' ');
      if (parts.length < 2) {
        keyForm.error = t('app.ssh.authKey.modal.formatError');
        return false;
      }

      // 更新解析结果
      keyForm.parsed = {
        algorithm: parts[0],
        key: parts[1],
        comment: parts.slice(2).join(' ') || '',
      };

      keyForm.error = '';
      return true;
    };

    // 重置表单状态
    const resetForm = () => {
      keyForm.content = '';
      keyForm.validated = false;
      keyForm.error = '';
    };

    // 监听内容变化，重置验证状态
    watch(
      () => keyForm.content,
      () => {
        if (keyForm.validated) {
          validateKey();
        }
      }
    );

    return {
      keyForm,
      validateKey,
      resetForm,
    };
  };

  // SSH密钥管理逻辑
  const useSSHKeys = () => {
    const keys = ref<any[]>([]);
    const loading = ref(false);

    const fetchKeys = async () => {
      loading.value = true;
      try {
        // 模拟API调用
        await new Promise((resolve) => {
          setTimeout(resolve, 500);
        });
        keys.value = [
          {
            algorithm: 'ssh-rsa',
            key: 'AAAAB3NzaC1yc2EAAAADAQABAAABAQDcYWVOEj2a9zUpgEEAYnwhJ9h/9x/NG4x+qJMMm5Zz5jDT3mQlrG0bXmJH3hb5RPkc+hiD9YGrBGZtUzvWTY1zJcMj7xQULbLTDd8GFoFsJUmZ/mT9qFSkfkpW1VsGXVATLJE10MLltU0IFWshjBHN9q1TcZ0yFzTJ5D9CJ7E2i9HYgERrO5OQf2JrucKMQBcKnGNYgF9UiCbBl1JF9FWJnVQwGj4p/F5sTWkmAqGo16mr0LLlpLN3aEzxvK+wNSXmrksSlbUtA6EX6kLR8qLNeBL0KXzTnwALDFW6ck7H8yXnvjO3D3Y2nSGcxRHCFbQMmY5kcCBG4jKALQ+jlRmX',
            comment: 'user@example.com',
          },
          {
            algorithm: 'ssh-ed25519',
            key: 'AAAAC3NzaC1lZDI1NTE5AAAAIKhxHPSLRcJJwPYVfxsgzuj+Q0zijj2yjqxiLkEwNcQ/',
            comment: 'backup-key',
          },
          {
            algorithm: 'ssh-rsa',
            key: 'AAAAB3NzaC1yc2EAAAADAQABAAABAQC57C+lS8CWcZZOMWnqKOzr7PiU6B2auIXHACe9LrpVhaNqu+A1efm17WIuKIXYcJZLt1r/28NfQjpNnNTlYsRwJkZnBmVjoH8QlvS1Y2mLw3ai6JBBXeKf9J/IwQMZoYha5AkKg5X4XEYnQUKQwNYTRQSMMJgEuyjeV4m3yJ0Xcsa+G6oxBMlRW5gQJ2te',
            comment: 'gitlab_key',
          },
          {
            algorithm: 'ssh-rsa',
            key: 'AAAAB3NzaC1yc2EAAAADAQABAAABgQDHvlRoXaebX9q3xCjgXQCK/kZF2ZPHT0gSVWIxE5U+5gdjhi4XbV/H+X7Rc0oj0o1fTS1OHJ/y+AKyDW1w/7gOnROzAVlmPdOCSn5/JyQIMnMqxZszVaf6wfMH5i4cVIEclhGPVTZYwEJKX+PvHcMUJ+Q7LMRUx8ESm5j0vJ1XDV33AyYSX7Yfkty6EAAn50gk2seKD2zTLfQAYpuTMAuXN5c7jJoY0J0PfWTHwGNR+KQtBLK1dyTf5SLj+P9DDrHZZC55jLbqjGWDWJpGdwv7JRKmGRtQJYqQBzIVbRQlY38RSlTpwGcJMitG/rRkoaXyyaWtEHWkq7I78X5YZC6ij9Ud0PFW8Qq4sEVkrRvzlxKcnngSGXTYdWUvQmuLs+AzKGRdgSJuLvgFV5K1UF+Vk5WsJEbIvV+uyJ1E0/7fQACbzJeZpqEN9nZnr2yz+iyPULYwyYxRGmdNUkm3Dy4sfTYcqzOJpzuNrA1CbOUQw3k1bpycPXbSvxXTMMM=',
            comment: 'deploy@server1',
          },
          {
            algorithm: 'ecdsa-sha2-nistp256',
            key: 'AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBH+WgxSXpzsUUdgYtFnvYZ6hGJsbaOKN7kMnYF/GG7h4EBnBYGGPxV1GVjuA0p6X+IXpz8rHMP4XVZXpY3o8ys0=',
            comment: 'ci@jenkins',
          },
          {
            algorithm: 'ssh-rsa',
            key: 'AAAAB3NzaC1yc2EAAAADAQABAAABAQCqfC9iXIKx/jvS7CZtRXA2Y0SRYs+UKPkOJxD18JmcIbAUw3Ye/NkGXVTRgeCdi6po5iXR/CzC9ZUXVnVnvZ0bxA2yUOsK0P14MpXYQ0b8eK5G0vVLUXLCH2cNNO8SK0DTxZvMZgVcMZfJgAi1QNh8qvMi8C6Jx9XbIheGVGULnQwmHiUCUyLZRPtGC6jUBwdKHah9gpeEtnV7K2A9bjk1vjq9UjCnY0BgKQEK/6EFKpA4kcPEKnLWNc3V5R3aN+Nj9VGA/X2p5M+qrp5i3Y3/wQ8eEUAgRmnVUTFtLk7Ug+3yZoy3jYbcQlrGeM6v0SnLp/KtRR0S/rPFAFfdS1Wv',
            comment: 'developer@workstation',
          },
          {
            algorithm: 'ssh-ed25519',
            key: 'AAAAC3NzaC1lZDI1NTE5AAAAIDTsEMXuCKYU6f5ikK5+xYL9QeLv3BgWzG7RQfTEzn0i',
            comment: 'backup@server2',
          },
        ];
      } catch (error) {
        console.error('Failed to fetch keys:', error);
      } finally {
        loading.value = false;
      }
    };

    // 添加新SSH密钥
    const addKey = async (keyData: any) => {
      await new Promise((resolve) => {
        setTimeout(resolve, 300);
      });

      // 使用不可变更新模式
      keys.value = [...keys.value, keyData];
      return keyData;
    };

    // 删除SSH密钥
    const removeKey = async (keyToRemove: string) => {
      await new Promise((resolve) => {
        setTimeout(resolve, 300);
      });

      // 使用不可变更新模式
      keys.value = keys.value.filter((item) => item.key !== keyToRemove);
      return true;
    };

    return {
      keys,
      loading,
      fetchKeys,
      addKey,
      removeKey,
    };
  };

  // 使用SSH密钥验证组合函数
  const { keyForm: addKeyForm, validateKey, resetForm } = useSSHKeyValidator();

  // 使用SSH密钥管理组合函数
  const { keys, loading, addKey, removeKey, fetchKeys } = useSSHKeys();

  // 表格数据源
  const dataSource = computed<ApiListResult<any>>(() => ({
    items: keys.value,
    total: keys.value.length,
    page: 1,
    page_size: 20,
  }));

  const tableRef = ref();

  // 模态框状态
  const addKeyModalVisible = ref(false);
  const removeConfirmVisible = ref(false);

  // 当前操作的记录
  const currentRecord = ref<any>(null);

  // 表格列定义
  const columns = ref<Column[]>([
    {
      title: t('app.ssh.authKey.columns.algorithm'),
      dataIndex: 'algorithm',
      width: 100,
    },
    {
      title: t('app.ssh.authKey.columns.key'),
      dataIndex: 'key',
      ellipsis: true,
    },
    {
      title: t('app.ssh.authKey.columns.comment'),
      dataIndex: 'comment',
      width: 150,
    },
    {
      title: t('app.ssh.authKey.columns.operations'),
      dataIndex: 'operations',
      width: 100,
      slotName: 'operations',
    },
  ]);

  // 显示添加密钥弹窗
  const showAddKeyModal = () => {
    resetForm();
    addKeyModalVisible.value = true;
  };

  // 添加密钥
  const handleAddKey = async () => {
    // 使用提取的验证逻辑
    if (!validateKey()) {
      return;
    }

    try {
      await addKey(addKeyForm.parsed);
      addKeyModalVisible.value = false;
      Message.success(t('app.ssh.authKey.addSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.authKey.addError'));
    }
  };

  // 处理删除密钥
  const handleRemove = (record: any) => {
    currentRecord.value = record;
    removeConfirmVisible.value = true;
  };

  // 确认删除
  const confirmRemove = async () => {
    if (currentRecord.value) {
      try {
        await removeKey(currentRecord.value.key);
        Message.success(t('app.ssh.authKey.removeSuccess'));
      } catch (error) {
        Message.error(t('app.ssh.authKey.removeError'));
      }
    }
    removeConfirmVisible.value = false;
  };

  // 组件挂载时加载数据
  onMounted(() => {
    fetchKeys();
  });
</script>

<style scoped lang="less">
  .auth-key-container {
    padding: 0;
  }

  .generate-key-btn {
    background-color: #6c52fa;

    &:hover,
    &:focus {
      background-color: #8b74ff;
    }
  }

  .modal-form-wrapper {
    padding: 0 20px;
  }

  .modal-form-item {
    display: flex;
    margin-bottom: 20px;
  }

  .modal-label {
    width: 80px;
    margin-right: 20px;
    color: var(--color-text-1);
    font-weight: 500;
    line-height: 32px;
    text-align: right;
  }

  .modal-input-wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .modal-field-description {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
