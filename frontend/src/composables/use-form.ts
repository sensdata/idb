import { ref, reactive } from 'vue';
import { Message } from '@arco-design/web-vue';
import type { FormInstance } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useLogger } from '@/composables/use-logger';

export interface FormOptions<T> {
  initialValues: T;
  onSubmit?: (values: T) => Promise<void> | void;
  onError?: (error: Error) => void;
  validateMessage?: string;
}

export function useForm<T extends Record<string, any>>(
  options: FormOptions<T>
) {
  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const formData = reactive<T>({ ...options.initialValues });
  const loading = ref(false);
  const { logWarn, logError } = useLogger('FormHook');

  // 重置表单为初始值
  const resetForm = () => {
    // 使用深拷贝确保不会影响原始数据
    Object.keys(options.initialValues).forEach((key) => {
      // @ts-ignore
      formData[key] = options.initialValues[key];
    });

    if (formRef.value) {
      formRef.value.resetFields();
    }
  };

  // 设置表单引用
  const setFormRef = (el: FormInstance) => {
    if (el) {
      formRef.value = el;
    } else {
      logWarn('Form reference is invalid');
    }
  };

  // 提交表单
  const submitForm = async () => {
    if (!formRef.value) {
      logError('Form reference is not set');
      throw new Error('Form reference is not set');
    }

    try {
      loading.value = true;
      const errors = await formRef.value.validate();
      if (errors) {
        return;
      }

      if (options.onSubmit) {
        await options.onSubmit(formData as T);
      } else {
        logWarn('No submit handler provided');
      }
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      logError('Form validation or submission failed:', err);
      Message.error(options.validateMessage || t('common.form.validateFailed'));
      options.onError?.(err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  // 更新表单数据
  const updateForm = (values: Partial<T>) => {
    Object.assign(formData, values);
  };

  return {
    formRef,
    formData,
    loading,
    resetForm,
    setFormRef,
    submitForm,
    updateForm,
  };
}
