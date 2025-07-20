<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.updateCertificate')"
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

        <!-- 补齐证书链选项 -->
        <a-divider>{{ $t('app.certificate.import.options') }}</a-divider>
        <a-form-item field="complete_chain">
          <a-checkbox v-model="form.complete_chain">
            {{ $t('app.certificate.import.completeChain') }}
          </a-checkbox>
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
    alias?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    alias: '',
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
    ca_type: 0,
    ca_file: null as File | null,
    ca_content: '',
    ca_path: '',
    complete_chain: false,
  });

  const caFileList = ref<FileItem[]>([]);

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 表单验证状态
  const isFormValid = computed(() => {
    // 基本字段验证
    const hasAlias = form.value.alias.trim() !== '';

    // 证书验证
    let hasValidCert = false;
    if (form.value.ca_type === 0) {
      hasValidCert = !!form.value.ca_file;
    } else if (form.value.ca_type === 1) {
      hasValidCert = form.value.ca_content.trim() !== '';
    } else if (form.value.ca_type === 2) {
      hasValidCert = form.value.ca_path.trim() !== '';
    }

    return hasAlias && hasValidCert;
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
  const handleCaFileChange = (fileList: FileItem[]) => {
    caFileList.value = fileList;
    if (fileList.length > 0) {
      form.value.ca_file = fileList[0].file as File;
    } else {
      form.value.ca_file = null;
    }
  };

  // 重置表单
  const resetForm = () => {
    form.value = {
      alias: props.alias || '',
      ca_type: 0,
      ca_file: null,
      ca_content: '',
      ca_path: '',
      complete_chain: false,
    };
    caFileList.value = [];
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

        // 添加证书信息
        formData.append('ca_type', form.value.ca_type.toString());
        if (form.value.ca_type === 0 && form.value.ca_file) {
          formData.append('ca_file', form.value.ca_file);
        } else if (form.value.ca_type === 1) {
          formData.append('ca_content', form.value.ca_content);
        } else if (form.value.ca_type === 2) {
          formData.append('ca_path', form.value.ca_path);
        }

        // 添加补齐证书链选项
        formData.append('complete_chain', form.value.complete_chain.toString());

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

  // 监听 alias prop 变化
  watch(
    () => props.alias,
    (newAlias) => {
      if (newAlias) {
        form.value.alias = newAlias;
      }
    },
    { immediate: true }
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
