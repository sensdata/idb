<template>
  <a-modal
    :visible="visible"
    :title="$t('app.ssh.keyPairs.generateModal.title')"
    :ok-loading="loading"
    @ok="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="modal-form-wrapper">
      <a-form
        ref="formRef"
        :model="formData"
        label-align="right"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 18 }"
      >
        <a-form-item
          field="key_name"
          :label="$t('app.ssh.keyPairs.generateModal.keyName')"
          :rules="[
            {
              required: true,
              message: $t('app.ssh.keyPairs.generateModal.keyNameRequired'),
            },
          ]"
        >
          <a-input v-model="formData.key_name" />
        </a-form-item>
        <a-form-item
          field="encryption_mode"
          :label="$t('app.ssh.keyPairs.generateModal.encryptionMode')"
          :rules="[
            {
              required: true,
              message: $t(
                'app.ssh.keyPairs.generateModal.encryptionModeRequired'
              ),
            },
          ]"
        >
          <a-select v-model="formData.encryption_mode">
            <a-option
              v-for="option in encryptionOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          field="key_bits"
          :label="$t('app.ssh.keyPairs.generateModal.keyBits')"
          :rules="[
            {
              required: true,
              message: $t('app.ssh.keyPairs.generateModal.keyBitsRequired'),
            },
          ]"
        >
          <a-select v-model="formData.key_bits">
            <a-option
              v-for="option in keyBitsOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          field="password"
          :label="$t('app.ssh.keyPairs.generateModal.password')"
        >
          <a-input-password v-model="formData.password" allow-clear />
        </a-form-item>
        <a-form-item
          field="enable"
          :label="$t('app.ssh.keyPairs.generateModal.enable')"
        >
          <a-switch v-model="formData.enable" />
        </a-form-item>
      </a-form>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
  import { defineProps, defineEmits, computed, ref, watch } from 'vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import {
    GenerateKeyForm,
    EncryptionMode,
    KeyBits,
  } from '@/views/app/ssh/types';
  import { useLogger } from '@/hooks/use-logger';

  interface EncryptionOption {
    label: string;
    value: EncryptionMode;
  }

  interface KeyBitsOption {
    label: string;
    value: KeyBits;
  }

  const props = defineProps<{
    visible: boolean;
    loading: boolean;
    encryptionOptions: EncryptionOption[];
    keyBitsOptions: KeyBitsOption[];
    form: GenerateKeyForm;
  }>();

  const emit = defineEmits<{
    (e: 'confirm'): void;
    (e: 'update:visible', value: boolean): void;
    (e: 'update:loading', value: boolean): void;
    (e: 'update:form', form: GenerateKeyForm): void;
    (e: 'formRefUpdated', formRef: FormInstance): void;
  }>();

  const formRef = ref<FormInstance | null>(null);
  const { logError } = useLogger('KeyGenerateModal');

  watch(
    formRef,
    (newVal) => {
      if (newVal) {
        emit('formRefUpdated', newVal);
      }
    },
    { immediate: true }
  );

  // 使用计算属性处理表单数据，避免直接修改prop
  const formData = computed({
    get: () => props.form,
    set: (newValue) => {
      emit('update:form', newValue);
    },
  });

  const handleConfirm = async () => {
    if (!formRef.value) {
      logError('Form reference is missing');
      return;
    }

    try {
      await formRef.value.validate();
      emit('confirm');
    } catch (error) {
      logError('Form validation failed:', error);
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };
</script>

<style scoped lang="less">
  .modal-form-wrapper {
    padding: 20px 0;
  }
</style>
