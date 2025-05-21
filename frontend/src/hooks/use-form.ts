import { ref, reactive } from 'vue';
import { Message } from '@arco-design/web-vue';
import type { FormInstance } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';

export interface FormOptions<T> {
  initialValues: T;
  onSubmit?: (values: T) => Promise<void> | void;
  onError?: (error: Error) => void;
  validateMessage?: string;
}

interface ResetOptions {
  validate?: boolean;
  deep?: boolean;
}

export function useForm<T extends Record<string, any>>(
  options: FormOptions<T>
) {
  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const formData = reactive<T>({ ...options.initialValues });
  const loading = ref(false);

  // 重置表单为初始值
  const resetForm = (opt?: ResetOptions) => {
    // 使用深拷贝确保不会影响原始数据
    Object.keys(options.initialValues).forEach((key) => {
      formData[key] = options.initialValues[key];
    });

    if (formRef.value) {
      formRef.value.resetFields();
    }
  };

  // 设置表单引用
  const setFormRef = (el: FormInstance) => {
    formRef.value = el;
  };

  // 提交表单
  const submitForm = async () => {
    if (!formRef.value) {
      return;
    }

    try {
      loading.value = true;
      await formRef.value.validate();
      await options.onSubmit?.(formData as T);
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
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
