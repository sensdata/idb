<template>
  <a-modal
    :visible="visible"
    :title="$t('app.ssh.publicKeys.modal.addPublicKey')"
    role="dialog"
    :mask-closable="false"
    :unmount-on-close="false"
    :ok-loading="isSubmitting"
    :footer-class="'key-modal-footer'"
    :body-class="'key-modal-body'"
    :width="600"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <div class="modal-form-wrapper">
      <a-form
        ref="formRef"
        :model="keyForm.formData"
        label-align="right"
        :label-col-props="{ span: 4 }"
        :wrapper-col-props="{ span: 20 }"
      >
        <a-form-item
          field="content"
          :label="$t('app.ssh.publicKeys.modal.content')"
          :rules="[
            {
              required: true,
              message: $t('app.ssh.publicKeys.modal.emptyError'),
            },
            {
              validator: validateKeyContent,
            },
          ]"
        >
          <a-textarea
            v-model="keyForm.formData.content"
            :placeholder="$t('app.ssh.publicKeys.modal.placeholder')"
            :auto-size="{ minRows: 6, maxRows: 10 }"
            class="key-textarea"
          />
        </a-form-item>
        <div class="description-wrapper">
          <div class="modal-field-description">
            {{ $t('app.ssh.publicKeys.modal.description') }}
          </div>
        </div>
      </a-form>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, defineProps, defineEmits, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue';
  import { useForm } from '@/hooks/use-form';
  import { KeyForm, ParsedKey } from '@/views/app/ssh/types';
  import { useLogger } from '@/hooks/use-logger';

  const props = defineProps<{
    visible: boolean;
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'ok', data: { content: string; parsed: ParsedKey }): void;
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const { logError } = useLogger('PublicKeyAddModal');
  const formRef = ref<FormInstance>();
  const isSubmitting = ref(false);

  // 表单默认值
  const defaultFormState: KeyForm = {
    content: '',
  };

  // 解析SSH密钥内容
  const parseKeyContent = (content: string): ParsedKey => {
    const parts = content.trim().split(' ');
    return {
      algorithm: parts[0] || '',
      key: parts[1] || '',
      comment: parts.slice(2).join(' ') || '',
    };
  };

  // 使用表单Hook
  const keyForm = useForm<KeyForm>({
    initialValues: defaultFormState,
    onSubmit: async (values) => {
      const content = values.content.trim();
      const parsed = parseKeyContent(content);
      emit('ok', { content, parsed });
    },
  });

  // 监听表单引用
  watch(formRef, (newRef) => {
    if (newRef) {
      keyForm.setFormRef(newRef);
    }
  });

  // 验证密钥内容
  const validateKeyContent = (
    value: string,
    callback: (error?: string) => void
  ) => {
    if (!value.trim()) {
      return callback(t('app.ssh.publicKeys.modal.emptyError'));
    }

    // 解析密钥内容
    const parts = value.trim().split(' ');
    if (parts.length < 2) {
      return callback(t('app.ssh.publicKeys.modal.formatError'));
    }

    // 检查算法类型是否合法
    const validAlgorithms = [
      'ssh-rsa',
      'ssh-ed25519',
      'ecdsa-sha2-nistp256',
      'ecdsa-sha2-nistp384',
      'ecdsa-sha2-nistp521',
      'ssh-dss',
    ];
    if (!validAlgorithms.includes(parts[0])) {
      return callback(t('app.ssh.publicKeys.modal.invalidAlgorithm'));
    }

    // 检查密钥部分是否是有效的Base64
    const keyPart = parts[1];
    try {
      // 尝试验证是否为有效的Base64
      if (!/^[A-Za-z0-9+/]+={0,2}$/.test(keyPart)) {
        return callback(t('app.ssh.publicKeys.modal.invalidKeyFormat'));
      }

      // 确保密钥长度合理
      if (keyPart.length < 20) {
        return callback(t('app.ssh.publicKeys.modal.keyTooShort'));
      }
    } catch (e) {
      return callback(t('app.ssh.publicKeys.modal.invalidKeyFormat'));
    }

    callback();
    return true;
  };

  // 处理提交
  const handleOk = async () => {
    if (!formRef.value) {
      logError('Form reference is not set');
      return;
    }

    try {
      // 验证表单
      const errors = await formRef.value.validate();

      // 如果有验证错误，直接返回，不执行后续操作
      if (errors) {
        return;
      }

      // 验证通过，设置loading状态
      isSubmitting.value = true;

      // 处理表单提交
      const content = keyForm.formData.content.trim();
      const parsed = parseKeyContent(content);
      emit('ok', { content, parsed });
    } catch (error) {
      logError('Form validation failed:', error);
      // 验证失败，不做其他操作
    }
  };

  // 处理取消
  const handleCancel = () => {
    keyForm.resetForm();
    emit('update:visible', false);
  };

  // 监听弹窗可见性变化，当弹窗打开时重置表单
  watch(
    () => props.visible,
    (newVisible) => {
      if (newVisible) {
        keyForm.resetForm();
        isSubmitting.value = false;
      }
    }
  );

  // 监听loading属性变化
  watch(
    () => props.loading,
    (newLoading) => {
      // 当外部loading结束时，也结束内部isSubmitting状态
      if (!newLoading) {
        isSubmitting.value = false;
      }
    }
  );
</script>

<style scoped lang="less">
  .modal-form-wrapper {
    padding: 0 16px;
  }

  .description-wrapper {
    padding-left: calc(4 / 24 * 100%);
    margin-top: 0;
    margin-bottom: 24px;
  }

  .modal-field-description {
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 1.5;
    padding-top: 12px;
  }

  .error-text {
    color: #f53f3f !important;
    font-weight: 500;
  }

  :deep(.key-modal-body) {
    padding: 20px 4px;
  }

  :deep(.key-modal-footer) {
    padding: 16px 20px;
    border-top: 1px solid var(--color-border-2);
  }

  :deep(.key-textarea) {
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
    font-size: 13px;
    line-height: 1.5;

    .arco-textarea-wrapper {
      padding: 8px;
    }

    &:focus,
    &:hover {
      border-color: var(--color-primary-6);
    }
  }
</style>
