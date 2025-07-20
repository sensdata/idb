<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.import')"
    :width="700"
    :footer="true"
    @cancel="handleCancel"
  >
    <template #footer>
      <div class="drawer-footer">
        <a-button @click="handleCancel">
          {{ $t('common.cancel') }}
        </a-button>
        <a-button
          type="primary"
          :loading="loading"
          :disabled="!isFormValid"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </a-button>
      </div>
    </template>
    <div class="drawer-content">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        @submit="handleSubmit"
      >
        <!-- 基本信息 -->
        <a-form-item
          field="alias"
          :label="$t('app.certificate.form.alias')"
          required
        >
          <a-input
            v-model="form.alias"
            :placeholder="$t('app.certificate.form.aliasPlaceholder')"
          />
        </a-form-item>

        <!-- 私钥导入 -->
        <a-divider>{{ $t('app.certificate.import.privateKey') }}</a-divider>
        <a-form-item
          field="key_type"
          :label="$t('app.certificate.import.keyType')"
          required
        >
          <a-radio-group v-model="form.key_type">
            <a-radio :value="0">{{
              $t('app.certificate.import.fileUpload')
            }}</a-radio>
            <a-radio :value="1">{{
              $t('app.certificate.import.textInput')
            }}</a-radio>
            <a-radio :value="2">{{
              $t('app.certificate.import.localPath')
            }}</a-radio>
          </a-radio-group>
        </a-form-item>

        <!-- 私钥文件上传 -->
        <a-form-item
          v-if="form.key_type === 0"
          field="key_file"
          :label="$t('app.certificate.import.keyFile')"
          required
        >
          <a-upload
            :file-list="keyFileList"
            :show-file-list="true"
            :auto-upload="false"
            accept=".key,.pem"
            @change="handleKeyFileChange"
          >
            <template #upload-button>
              <a-button>
                <template #icon>
                  <icon-upload />
                </template>
                {{ $t('app.certificate.import.selectFile') }}
              </a-button>
            </template>
          </a-upload>
        </a-form-item>

        <!-- 私钥文本输入 -->
        <a-form-item
          v-if="form.key_type === 1"
          field="key_content"
          :label="$t('app.certificate.import.keyContent')"
          required
        >
          <a-textarea
            v-model="form.key_content"
            :placeholder="$t('app.certificate.import.keyContentPlaceholder')"
            :rows="6"
          />
        </a-form-item>

        <!-- 私钥本地路径 -->
        <a-form-item
          v-if="form.key_type === 2"
          field="key_path"
          :label="$t('app.certificate.import.keyPath')"
          required
        >
          <FileSelector
            v-model="form.key_path"
            type="file"
            :placeholder="$t('app.certificate.import.keyPathPlaceholder')"
          />
        </a-form-item>

        <!-- 证书导入 -->
        <a-divider>{{ $t('app.certificate.import.certificate') }}</a-divider>
        <a-form-item
          field="ca_type"
          :label="$t('app.certificate.import.certificateType')"
          required
        >
          <a-radio-group v-model="form.ca_type">
            <a-radio :value="0">{{
              $t('app.certificate.import.fileUpload')
            }}</a-radio>
            <a-radio :value="1">{{
              $t('app.certificate.import.textInput')
            }}</a-radio>
            <a-radio :value="2">{{
              $t('app.certificate.import.localPath')
            }}</a-radio>
          </a-radio-group>
        </a-form-item>

        <!-- 证书文件上传 -->
        <a-form-item
          v-if="form.ca_type === 0"
          field="ca_file"
          :label="$t('app.certificate.import.certificateFile')"
          required
        >
          <a-upload
            :file-list="caFileList"
            :show-file-list="true"
            :auto-upload="false"
            accept=".crt,.pem,.cer"
            @change="handleCaFileChange"
          >
            <template #upload-button>
              <a-button>
                <template #icon>
                  <icon-upload />
                </template>
                {{ $t('app.certificate.import.selectFile') }}
              </a-button>
            </template>
          </a-upload>
        </a-form-item>

        <!-- 证书文本输入 -->
        <a-form-item
          v-if="form.ca_type === 1"
          field="ca_content"
          :label="$t('app.certificate.import.certificateContent')"
          required
        >
          <a-textarea
            v-model="form.ca_content"
            :placeholder="
              $t('app.certificate.import.certificateContentPlaceholder')
            "
            :rows="6"
          />
        </a-form-item>

        <!-- 证书本地路径 -->
        <a-form-item
          v-if="form.ca_type === 2"
          field="ca_path"
          :label="$t('app.certificate.import.certificatePath')"
          required
        >
          <FileSelector
            v-model="form.ca_path"
            type="file"
            :placeholder="
              $t('app.certificate.import.certificatePathPlaceholder')
            "
          />
        </a-form-item>

        <!-- CSR导入（可选） -->
        <a-divider
          >{{ $t('app.certificate.import.csr') }} ({{
            $t('common.optional')
          }})</a-divider
        >
        <a-form-item
          field="csr_type"
          :label="$t('app.certificate.import.csrType')"
        >
          <a-radio-group v-model="form.csr_type">
            <a-radio :value="0">{{
              $t('app.certificate.import.fileUpload')
            }}</a-radio>
            <a-radio :value="1">{{
              $t('app.certificate.import.textInput')
            }}</a-radio>
            <a-radio :value="2">{{
              $t('app.certificate.import.localPath')
            }}</a-radio>
          </a-radio-group>
        </a-form-item>

        <!-- CSR文件上传 -->
        <a-form-item
          v-if="form.csr_type === 0"
          field="csr_file"
          :label="$t('app.certificate.import.csrFile')"
        >
          <a-upload
            :file-list="csrFileList"
            :show-file-list="true"
            :auto-upload="false"
            accept=".csr,.pem"
            @change="handleCsrFileChange"
          >
            <template #upload-button>
              <a-button>
                <template #icon>
                  <icon-upload />
                </template>
                {{ $t('app.certificate.import.selectFile') }}
              </a-button>
            </template>
          </a-upload>
        </a-form-item>

        <!-- CSR文本输入 -->
        <a-form-item
          v-if="form.csr_type === 1"
          field="csr_content"
          :label="$t('app.certificate.import.csrContent')"
        >
          <a-textarea
            v-model="form.csr_content"
            :placeholder="$t('app.certificate.import.csrContentPlaceholder')"
            :rows="6"
          />
        </a-form-item>

        <!-- CSR本地路径 -->
        <a-form-item
          v-if="form.csr_type === 2"
          field="csr_path"
          :label="$t('app.certificate.import.csrPath')"
        >
          <FileSelector
            v-model="form.csr_path"
            type="file"
            :placeholder="$t('app.certificate.import.csrPathPlaceholder')"
          />
        </a-form-item>
      </a-form>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconUpload } from '@arco-design/web-vue/es/icon';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FileItem } from '@arco-design/web-vue/es/upload';
  import FileSelector from '@/components/file/file-selector/index.vue';

  // Props 定义
  interface Props {
    visible: boolean;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'ok', formData: FormData): void;
  }>();

  const { t } = useI18n();

  // 响应式数据
  const formRef = ref<FormInstance>();
  const form = ref({
    alias: '',
    key_type: 0,
    key_file: null as File | null,
    key_content: '',
    key_path: '',
    ca_type: 0,
    ca_file: null as File | null,
    ca_content: '',
    ca_path: '',
    csr_type: 0,
    csr_file: null as File | null,
    csr_content: '',
    csr_path: '',
  });

  const keyFileList = ref<FileItem[]>([]);
  const caFileList = ref<FileItem[]>([]);
  const csrFileList = ref<FileItem[]>([]);

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 表单验证状态
  const isFormValid = computed(() => {
    // 基本字段验证
    const hasAlias = form.value.alias.trim() !== '';

    // 私钥验证
    let hasValidKey = false;
    if (form.value.key_type === 0) {
      hasValidKey = !!form.value.key_file;
    } else if (form.value.key_type === 1) {
      hasValidKey = form.value.key_content.trim() !== '';
    } else if (form.value.key_type === 2) {
      hasValidKey = form.value.key_path.trim() !== '';
    }

    // 证书验证
    let hasValidCert = false;
    if (form.value.ca_type === 0) {
      hasValidCert = !!form.value.ca_file;
    } else if (form.value.ca_type === 1) {
      hasValidCert = form.value.ca_content.trim() !== '';
    } else if (form.value.ca_type === 2) {
      hasValidCert = form.value.ca_path.trim() !== '';
    }

    return hasAlias && hasValidKey && hasValidCert;
  });

  // 表单验证规则
  const rules = {
    alias: [
      {
        required: true,
        message: t('app.certificate.form.aliasRequired'),
      },
      {
        match: /^[a-zA-Z0-9_-]+$/,
        message: t('app.certificate.form.aliasFormat'),
      },
    ],
    key_file: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.key_type === 0 && !form.value.key_file) {
            callback(t('app.certificate.import.keyFileRequired'));
          } else {
            callback();
          }
        },
      },
    ],
    key_content: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.key_type === 1 && !form.value.key_content) {
            callback(t('app.certificate.import.keyContentRequired'));
          } else {
            callback();
          }
        },
      },
    ],
    key_path: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.key_type === 2 && !form.value.key_path) {
            callback(t('app.certificate.import.keyPathRequired'));
          } else {
            callback();
          }
        },
      },
    ],
    ca_file: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.ca_type === 0 && !form.value.ca_file) {
            callback(t('app.certificate.import.certificateFileRequired'));
          } else {
            callback();
          }
        },
      },
    ],
    ca_content: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.ca_type === 1 && !form.value.ca_content) {
            callback(t('app.certificate.import.certificateContentRequired'));
          } else {
            callback();
          }
        },
      },
    ],
    ca_path: [
      {
        validator: (_: any, callback: any) => {
          if (form.value.ca_type === 2 && !form.value.ca_path) {
            callback(t('app.certificate.import.certificatePathRequired'));
          } else {
            callback();
          }
        },
      },
    ],
  };

  // 文件上传处理
  const handleKeyFileChange = (fileList: FileItem[]) => {
    keyFileList.value = fileList;
    if (fileList.length > 0) {
      form.value.key_file = fileList[0].file as File;
    } else {
      form.value.key_file = null;
    }
  };

  const handleCaFileChange = (fileList: FileItem[]) => {
    caFileList.value = fileList;
    if (fileList.length > 0) {
      form.value.ca_file = fileList[0].file as File;
    } else {
      form.value.ca_file = null;
    }
  };

  const handleCsrFileChange = (fileList: FileItem[]) => {
    csrFileList.value = fileList;
    if (fileList.length > 0) {
      form.value.csr_file = fileList[0].file as File;
    } else {
      form.value.csr_file = null;
    }
  };

  // 重置表单
  const resetForm = () => {
    form.value = {
      alias: '',
      key_type: 0,
      key_file: null,
      key_content: '',
      key_path: '',
      ca_type: 0,
      ca_file: null,
      ca_content: '',
      ca_path: '',
      csr_type: 0,
      csr_file: null,
      csr_content: '',
      csr_path: '',
    };
    keyFileList.value = [];
    caFileList.value = [];
    csrFileList.value = [];
    formRef.value?.resetFields();
  };

  // 处理提交
  const handleSubmit = async () => {
    try {
      const errors = await formRef.value?.validate();
      if (!errors) {
        // 验证通过，没有错误
        const formData = new FormData();

        // 添加基本信息
        formData.append('alias', form.value.alias);

        // 添加私钥信息
        formData.append('key_type', form.value.key_type.toString());
        if (form.value.key_type === 0 && form.value.key_file) {
          formData.append('key_file', form.value.key_file);
        } else if (form.value.key_type === 1) {
          formData.append('key_content', form.value.key_content);
        } else if (form.value.key_type === 2) {
          formData.append('key_path', form.value.key_path);
        }

        // 添加证书信息
        formData.append('ca_type', form.value.ca_type.toString());
        if (form.value.ca_type === 0 && form.value.ca_file) {
          formData.append('ca_file', form.value.ca_file);
        } else if (form.value.ca_type === 1) {
          formData.append('ca_content', form.value.ca_content);
        } else if (form.value.ca_type === 2) {
          formData.append('ca_path', form.value.ca_path);
        }

        // 添加CSR信息（如果有）
        if (form.value.csr_type !== undefined) {
          formData.append('csr_type', form.value.csr_type.toString());
          if (form.value.csr_type === 0 && form.value.csr_file) {
            formData.append('csr_file', form.value.csr_file);
          } else if (form.value.csr_type === 1 && form.value.csr_content) {
            formData.append('csr_content', form.value.csr_content);
          } else if (form.value.csr_type === 2 && form.value.csr_path) {
            formData.append('csr_path', form.value.csr_path);
          }
        }

        emit('ok', formData);
      }
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  // 处理取消
  const handleCancel = () => {
    drawerVisible.value = false;
  };

  // 监听弹窗显示状态
  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        resetForm();
      }
    }
  );
</script>

<style scoped>
  .drawer-content {
    height: 100%;
    padding: 1.67rem;
    overflow-y: auto;
  }

  .drawer-footer {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    padding: 1.33rem 0;
  }

  :deep(.arco-form-item) {
    margin-bottom: 1.67rem;
  }

  :deep(.arco-form-item-label) {
    margin-bottom: 0.67rem;
    font-weight: 500;
  }

  :deep(.arco-divider) {
    margin: 2.67rem 0 1.67rem 0;
    font-size: 1.33rem;
    font-weight: 600;
  }

  :deep(.arco-input) {
    height: 3.33rem;
  }

  :deep(.arco-textarea) {
    min-height: 10rem;
  }

  :deep(.arco-radio-group) {
    margin-bottom: 1.33rem;
  }

  :deep(.arco-upload) {
    width: 100%;
  }
</style>
