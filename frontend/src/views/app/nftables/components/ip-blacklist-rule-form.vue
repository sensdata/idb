<template>
  <div class="ip-blacklist-rule-form">
    <a-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      layout="vertical"
      @submit-success="handleSubmit"
    >
      <!-- IP地址 -->
      <a-form-item
        :label="$t('app.nftables.ipBlacklist.rules.ip')"
        field="ip"
        required
      >
        <a-input
          v-model="formData.ip"
          :placeholder="$t('app.nftables.ipBlacklist.rules.ipPlaceholder')"
          allow-clear
        />
        <template #extra>
          <div class="form-extra">
            {{ $t('app.nftables.ipBlacklist.rules.ipFormatHint') }}
          </div>
        </template>
      </a-form-item>

      <!-- 提示信息 -->
      <a-alert type="info" :closable="false" class="info-alert">
        <template #icon>
          <icon-info-circle />
        </template>
        {{ $t('app.nftables.ipBlacklist.rules.dropHint') }}
      </a-alert>
    </a-form>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { validateIPFormat, getIPType } from '@/utils';
  import { useLogger } from '@/composables/use-logger';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import { IconInfoCircle } from '@arco-design/web-vue/es/icon';
  // IP黑名单规则接口
  export interface IPBlacklistRule {
    ip: string;
    type: 'single' | 'cidr' | 'range';
    action: 'drop' | 'reject';
    description?: string;
    createdAt?: string;
  }

  // 常量定义
  const DEFAULT_ACTION = 'drop' as const;

  // 类型定义
  interface Props {
    loading?: boolean;
  }

  interface Emits {
    (e: 'submit', rule: IPBlacklistRule): void;
    (e: 'cancel'): void;
  }

  interface FormData {
    ip: string;
  }

  // Props 和 Emits 定义
  withDefaults(defineProps<Props>(), {
    loading: false,
  });

  const emit = defineEmits<Emits>();

  const { t } = useI18n();
  const { logError } = useLogger('IPBlacklistRuleForm');

  // 响应式数据
  const formRef = ref<FormInstance>();

  const formData = reactive<FormData>({
    ip: '',
  });

  // 计算属性
  const ipType = computed(() => getIPType(formData.ip));

  // 表单验证规则
  const formRules = computed(
    (): Record<keyof FormData, FieldRule[]> => ({
      ip: [
        {
          required: true,
          message: t('common.form.required'),
        },
        {
          validator: (value: string, callback: (error?: string) => void) => {
            if (!value) {
              callback();
              return;
            }

            if (validateIPFormat(value)) {
              callback();
              return;
            }
            callback(t('app.nftables.ipBlacklist.rules.invalidIPFormat'));
          },
        },
      ],
    })
  );

  // 方法定义
  const resetForm = (): void => {
    formData.ip = '';

    // 重置表单验证状态
    nextTick(() => {
      formRef.value?.clearValidate();
    });
  };

  const handleSubmit = async (): Promise<void> => {
    try {
      const errors = await formRef.value?.validate();
      if (errors) {
        logError('Form validation failed:', errors);
        return;
      }

      const rule: IPBlacklistRule = {
        ip: formData.ip.trim(),
        type: ipType.value,
        action: DEFAULT_ACTION,
        description: '',
        createdAt: new Date().toISOString(),
      };

      emit('submit', rule);
    } catch (error) {
      logError('Form submission failed:', error);
    }
  };

  // 暴露给模板的方法（如果需要）
  defineExpose({
    resetForm,
    validate: () => formRef.value?.validate(),
  });
</script>

<style scoped lang="less">
  .ip-blacklist-rule-form {
    .form-extra {
      font-size: 12px;
      color: var(--color-text-3);
      margin-top: 4px;
      line-height: 1.4;
    }

    .info-alert {
      margin-bottom: 16px;
    }

    :deep(.arco-radio) {
      display: flex;
      align-items: center;
      margin-bottom: 12px;
      padding: 8px 0;

      .arco-radio-label {
        padding-left: 8px;
        flex: 1;
      }
    }

    :deep(.arco-textarea) {
      resize: vertical;
    }
  }
</style>
