<template>
  <div class="ssh-password-container">
    <idb-table
      ref="gridRef"
      class="ssh-password-table"
      :loading="loading"
      :columns="columns"
      :fetch="getSSHKeysApi"
      :has-search="true"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleGenerateKey">
          {{ $t('app.ssh.password.generateKey') }}
        </a-button>
      </template>
      <template #passwordStatus="{ record }">
        <a-tag v-if="record.password" color="green">
          {{ $t('app.ssh.password.hasPassword') }}
        </a-tag>
        <a-tag v-else color="gray">
          {{ $t('app.ssh.password.noPassword') }}
        </a-tag>
      </template>
      <template #enabled="{ record }">
        <a-switch
          :model-value="!!record.enabled"
          :disabled="loading"
          @change="(value) => handleToggleEnable(record, Boolean(value))"
        />
      </template>
      <template #operation="{ record }">
        <div class="operation">
          <a-button
            type="text"
            size="small"
            :disabled="!record.enabled"
            @click="handleDownload(record)"
          >
            {{ $t('app.ssh.password.download') }}
          </a-button>
          <a-button type="text" size="small" @click="handleSetPassword(record)">
            {{
              record.password
                ? $t('app.ssh.password.update')
                : $t('app.ssh.password.set')
            }}
          </a-button>
          <a-button
            v-if="record.password"
            type="text"
            size="small"
            status="danger"
            @click="handleClearPassword(record)"
          >
            {{ $t('app.ssh.password.clear') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            status="danger"
            @click="handleDelete(record)"
          >
            {{ $t('app.ssh.password.delete') }}
          </a-button>
        </div>
      </template>
    </idb-table>

    <!-- 生成密钥弹窗 -->
    <a-modal
      v-model:visible="generateModalVisible"
      :title="$t('app.ssh.password.generateModal.title')"
      :ok-loading="modalLoading"
      @ok="handleGenerateConfirm"
      @cancel="generateModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <a-form
          ref="generateFormRef"
          :model="generateForm"
          label-align="right"
          :label-col-props="{ span: 6 }"
          :wrapper-col-props="{ span: 18 }"
        >
          <a-form-item
            field="key_name"
            :label="$t('app.ssh.password.generateModal.keyName')"
            :rules="[
              {
                required: true,
                message: $t('app.ssh.password.generateModal.keyNameRequired'),
              },
            ]"
          >
            <a-input v-model="generateForm.key_name" />
          </a-form-item>
          <a-form-item
            field="encryption_mode"
            :label="$t('app.ssh.password.generateModal.encryptionMode')"
            :rules="[
              {
                required: true,
                message: $t(
                  'app.ssh.password.generateModal.encryptionModeRequired'
                ),
              },
            ]"
          >
            <a-select v-model="generateForm.encryption_mode">
              <a-option value="rsa">RSA</a-option>
              <a-option value="ed25519">ED25519</a-option>
              <a-option value="ecdsa">ECDSA</a-option>
              <a-option value="dsa">DSA</a-option>
            </a-select>
          </a-form-item>
          <a-form-item
            field="key_bits"
            :label="$t('app.ssh.password.generateModal.keyBits')"
            :rules="[
              {
                required: true,
                message: $t('app.ssh.password.generateModal.keyBitsRequired'),
              },
            ]"
          >
            <a-select v-model="generateForm.key_bits">
              <a-option :value="1024">1024</a-option>
              <a-option :value="2048">2048</a-option>
            </a-select>
          </a-form-item>
          <a-form-item
            field="password"
            :label="$t('app.ssh.password.generateModal.password')"
          >
            <a-input-password v-model="generateForm.password" allow-clear />
          </a-form-item>
          <a-form-item
            field="enable"
            :label="$t('app.ssh.password.generateModal.enable')"
          >
            <a-switch v-model="generateForm.enable" />
          </a-form-item>
        </a-form>
      </div>
    </a-modal>

    <!-- 设置/更新密码弹窗 -->
    <a-modal
      v-model:visible="passwordModalVisible"
      :title="
        currentRecord && 'password' in currentRecord && currentRecord.password
          ? $t('app.ssh.password.updateModal.title')
          : $t('app.ssh.password.setModal.title')
      "
      :ok-loading="modalLoading"
      @ok="handlePasswordConfirm"
      @cancel="passwordModalVisible = false"
    >
      <div class="modal-form-wrapper">
        <a-form
          ref="passwordFormRef"
          :model="passwordForm"
          label-align="right"
          :label-col-props="{ span: 6 }"
          :wrapper-col-props="{ span: 18 }"
        >
          <a-form-item
            v-if="
              currentRecord &&
              'password' in currentRecord &&
              currentRecord.password
            "
            field="old_password"
            :label="$t('app.ssh.password.updateModal.oldPassword')"
            :rules="[
              {
                required: true,
                message: $t('app.ssh.password.updateModal.oldPasswordRequired'),
              },
            ]"
          >
            <a-input-password v-model="passwordForm.old_password" allow-clear />
          </a-form-item>
          <a-form-item
            :field="
              currentRecord &&
              'password' in currentRecord &&
              currentRecord.password
                ? 'new_password'
                : 'password'
            "
            :label="
              currentRecord &&
              'password' in currentRecord &&
              currentRecord.password
                ? $t('app.ssh.password.updateModal.newPassword')
                : $t('app.ssh.password.setModal.password')
            "
            :rules="[
              {
                required: true,
                message:
                  currentRecord &&
                  'password' in currentRecord &&
                  currentRecord.password
                    ? $t('app.ssh.password.updateModal.newPasswordRequired')
                    : $t('app.ssh.password.setModal.passwordRequired'),
              },
            ]"
          >
            <a-input-password
              :model-value="
                currentRecord &&
                'password' in currentRecord &&
                currentRecord.password
                  ? passwordForm.new_password
                  : passwordForm.password
              "
              allow-clear
              @update:model-value="(val: string) => { 
                if (currentRecord && 'password' in currentRecord && currentRecord.password) {
                  passwordForm.new_password = val;
                } else {
                  passwordForm.password = val;
                }
              }"
            />
          </a-form-item>
        </a-form>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, GlobalComponents } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/hooks/loading';
  import { useConfirm } from '@/hooks/confirm';
  import useHostStore from '@/store/modules/host';
  import { ApiListParams, ApiListResult } from '@/types/global';

  interface SSHKeyRecord {
    key_name: string;
    encryption_mode: string;
    key_bits: number;
    password: boolean;
    created_at: string;
    enabled: boolean;
  }

  const { t } = useI18n();
  const hostStore = useHostStore();
  const { loading, setLoading } = useLoading(false);
  const modalLoading = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const generateFormRef = ref();
  const passwordFormRef = ref();

  // 表格列定义
  const columns = [
    {
      title: t('app.ssh.password.columns.keyName'),
      dataIndex: 'key_name',
      width: 150,
    },
    {
      title: t('app.ssh.password.columns.encryptionMode'),
      dataIndex: 'encryption_mode',
      width: 120,
    },
    {
      title: t('app.ssh.password.columns.keyBits'),
      dataIndex: 'key_bits',
      width: 100,
    },
    {
      title: t('app.ssh.password.columns.password'),
      dataIndex: 'passwordStatus',
      width: 120,
      slotName: 'passwordStatus',
    },
    {
      title: t('app.ssh.password.columns.createTime'),
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: t('app.ssh.password.columns.enabled'),
      dataIndex: 'enabled',
      width: 100,
      slotName: 'enabled',
    },
    {
      title: t('common.table.operation'),
      dataIndex: 'operation',
      width: 240,
      align: 'center' as const,
      slotName: 'operation',
      fixed: 'right' as const,
    },
  ];

  // 生成密钥相关
  const generateModalVisible = ref(false);
  const generateForm = reactive({
    key_name: '',
    encryption_mode: 'rsa' as 'rsa' | 'ed25519' | 'ecdsa' | 'dsa',
    key_bits: 2048 as 1024 | 2048,
    password: '',
    enable: true,
  });

  // 设置密码相关
  const passwordModalVisible = ref(false);
  const currentRecord = ref<SSHKeyRecord | null>(null);
  const passwordForm = reactive({
    key_name: '',
    password: '',
    old_password: '',
    new_password: '',
  });

  // Mock API 函数
  // 实际项目中应替换为真实的API调用
  const getSSHKeysApi = async (
    params: ApiListParams
  ): Promise<ApiListResult<SSHKeyRecord>> => {
    setLoading(true);
    try {
      // 模拟API延迟
      await new Promise<void>((resolve) => {
        setTimeout(resolve, 500);
      });

      // 假数据
      const mockRecords = [
        {
          key_name: 'id_rsa',
          encryption_mode: 'rsa',
          key_bits: 2048,
          password: true,
          created_at: '2023-06-01 10:30:45',
          enabled: true,
        },
        {
          key_name: 'gitlab_key',
          encryption_mode: 'ed25519',
          key_bits: 1024,
          password: false,
          created_at: '2023-05-20 15:22:10',
          enabled: true,
        },
        {
          key_name: 'backup_key',
          encryption_mode: 'rsa',
          key_bits: 2048,
          password: true,
          created_at: '2023-04-15 09:12:33',
          enabled: false,
        },
      ];

      const filteredItems = mockRecords.filter(
        (item) =>
          !params.keyword ||
          item.key_name.toLowerCase().includes(params.keyword.toLowerCase())
      );

      return {
        total: filteredItems.length,
        items: filteredItems,
        page: params.page || 1,
        page_size: params.page_size || 20,
      };
    } finally {
      setLoading(false);
    }
  };

  // 生成密钥
  const handleGenerateKey = () => {
    generateForm.key_name = '';
    generateForm.encryption_mode = 'rsa';
    generateForm.key_bits = 2048;
    generateForm.password = '';
    generateForm.enable = true;
    generateModalVisible.value = true;
  };

  const handleGenerateConfirm = async () => {
    try {
      await generateFormRef.value.validate();
      modalLoading.value = true;

      // 模拟API调用
      await new Promise<void>((resolve) => {
        setTimeout(resolve, 800);
      });

      Message.success(t('app.ssh.password.generateSuccess'));
      generateModalVisible.value = false;
      gridRef.value?.reload();
    } catch (error) {
      // 表单验证失败
    } finally {
      modalLoading.value = false;
    }
  };

  // 启用/禁用密钥
  const handleToggleEnable = async (record: SSHKeyRecord, enabled: boolean) => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise<void>((resolve) => {
        setTimeout(resolve, 500);
      });

      Message.success(
        enabled
          ? t('app.ssh.password.enableSuccess')
          : t('app.ssh.password.disableSuccess')
      );
    } catch (error) {
      // 恢复状态
      record.enabled = !enabled;
      Message.error(t('app.ssh.password.operationFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 下载密钥
  const handleDownload = async (record: SSHKeyRecord) => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise<void>((resolve) => {
        setTimeout(resolve, 500);
      });

      Message.success(t('app.ssh.password.downloadSuccess'));
    } catch (error) {
      Message.error(t('app.ssh.password.operationFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 设置/更新密码
  const handleSetPassword = (record: SSHKeyRecord) => {
    currentRecord.value = record;
    passwordForm.key_name = record.key_name;
    passwordForm.password = '';
    passwordForm.old_password = '';
    passwordForm.new_password = '';
    passwordModalVisible.value = true;
  };

  const handlePasswordConfirm = async () => {
    try {
      await passwordFormRef.value.validate();
      modalLoading.value = true;

      // 模拟API调用
      await new Promise<void>((resolve) => {
        setTimeout(resolve, 800);
      });

      const isUpdate = currentRecord.value && currentRecord.value.password;
      Message.success(
        isUpdate
          ? t('app.ssh.password.updateSuccess')
          : t('app.ssh.password.setSuccess')
      );
      passwordModalVisible.value = false;
      gridRef.value?.reload();
    } catch (error) {
      // 表单验证失败
    } finally {
      modalLoading.value = false;
    }
  };

  // 清除密码
  const { confirm } = useConfirm();
  const handleClearPassword = async (record: SSHKeyRecord) => {
    if (
      await confirm({
        content: t('app.ssh.password.clearConfirm', {
          keyName: record.key_name,
        }),
      })
    ) {
      setLoading(true);
      try {
        // 模拟API调用
        await new Promise<void>((resolve) => {
          setTimeout(resolve, 500);
        });

        Message.success(t('app.ssh.password.clearSuccess'));
        gridRef.value?.reload();
      } catch (error) {
        Message.error(t('app.ssh.password.operationFailed'));
      } finally {
        setLoading(false);
      }
    }
  };

  // 删除密钥
  const handleDelete = async (record: SSHKeyRecord) => {
    if (
      await confirm({
        content: t('app.ssh.password.deleteConfirm', {
          keyName: record.key_name,
        }),
      })
    ) {
      setLoading(true);
      try {
        // 模拟API调用
        await new Promise<void>((resolve) => {
          setTimeout(resolve, 500);
        });

        Message.success(t('app.ssh.password.deleteSuccess'));
        gridRef.value?.reload();
      } catch (error) {
        Message.error(t('app.ssh.password.operationFailed'));
      } finally {
        setLoading(false);
      }
    }
  };
</script>

<style scoped lang="less">
  .ssh-password-container {
    width: 100%;

    .operation {
      display: flex;
      justify-content: center;

      :deep(.arco-btn-size-small) {
        padding-right: 4px;
        padding-left: 4px;
      }
    }
  }

  .modal-form-wrapper {
    padding: 20px 0;
  }
</style>
