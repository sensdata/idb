import { ref, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useForm } from '@/composables/use-form';
import { useLogger } from '@/composables/use-logger';
import { LOGROTATE_FREQUENCY, LOGROTATE_TYPE } from '@/config/enum';
import { DEFAULT_LOGROTATE_CATEGORY } from '../../../constants';
import type { FormData } from '../types';

export function useFormState() {
  const { t } = useI18n();
  const { log } = useLogger('LogrotateFormState');

  const activeMode = ref<'form' | 'raw'>('form');
  const previousMode = ref<'form' | 'raw'>('form');
  const isEdit = ref(false);
  const currentType = ref<LOGROTATE_TYPE>(LOGROTATE_TYPE.Local);
  const originalName = ref('');
  const originalCategory = ref('');
  const originalFormData = ref<FormData | null>(null);
  const originalRawContent = ref('');

  // 初始表单数据
  const initialFormData: FormData = {
    name: '',
    category: DEFAULT_LOGROTATE_CATEGORY,
    path: '',
    frequency: LOGROTATE_FREQUENCY.Daily,
    count: 7,
    compress: false,
    delayCompress: false,
    missingOk: false,
    notIfEmpty: false,
    create: 'create 0644 root root',
    preRotate: '',
    postRotate: '',
  };

  const formRules = {
    name: [
      { required: true, message: t('app.logrotate.form.name_required') },
      {
        pattern: /^[a-zA-Z0-9_-]+$/,
        message: t('app.logrotate.form.name_pattern'),
      },
    ],
    category: [
      { required: true, message: t('app.logrotate.form.category_required') },
    ],
    path: [{ required: true, message: t('app.logrotate.form.path_required') }],
    count: [
      { required: true, message: t('app.logrotate.form.count_required') },
      { type: 'number', min: 1, message: t('app.logrotate.form.count_min') },
    ],
  };

  const frequencyOptions = [
    {
      label: t('app.logrotate.frequency.daily'),
      value: LOGROTATE_FREQUENCY.Daily,
    },
    {
      label: t('app.logrotate.frequency.weekly'),
      value: LOGROTATE_FREQUENCY.Weekly,
    },
    {
      label: t('app.logrotate.frequency.monthly'),
      value: LOGROTATE_FREQUENCY.Monthly,
    },
    {
      label: t('app.logrotate.frequency.yearly'),
      value: LOGROTATE_FREQUENCY.Yearly,
    },
  ];

  const {
    formRef,
    formData,
    resetForm,
    submitForm: submitFormData,
    updateForm,
  } = useForm<FormData>({
    initialValues: initialFormData,
    onSubmit: async () => {
      return Promise.resolve();
    },
  });

  const drawerTitle = computed(() =>
    isEdit.value
      ? t('app.logrotate.form.edit_title')
      : t('app.logrotate.form.create_title')
  );

  // 表单变更检测
  const isFormChanged = computed(() => {
    if (!originalFormData.value) {
      log('❌ 变更检测: originalFormData 为空');
      return false;
    }

    log('🔍 变更检测:', {
      currentMode: activeMode.value,
      originalFormData: originalFormData.value,
      currentFormData: formData,
    });

    if (activeMode.value === 'form') {
      const current = JSON.stringify(formData);
      const original = JSON.stringify(originalFormData.value);
      const hasChanged = current !== original;

      log('📝 表单模式变更检测:', { hasChanged });
      return hasChanged;
    }

    const hasChanged = false;
    log('🔧 文件模式变更检测:', { hasChanged });
    return hasChanged;
  });

  // 重置状态
  const resetState = async () => {
    resetForm();
    activeMode.value = 'form';
    previousMode.value = 'form';
    isEdit.value = false;
    originalName.value = '';
    originalCategory.value = '';
    await nextTick();
    originalFormData.value = JSON.parse(JSON.stringify(formData));
    originalRawContent.value = '';
  };

  // 更新原始状态
  const updateOriginalState = async () => {
    await nextTick();
    originalFormData.value = JSON.parse(JSON.stringify(formData));
  };

  return {
    activeMode,
    previousMode,
    isEdit,
    currentType,
    originalName,
    originalCategory,
    originalFormData,
    originalRawContent,
    formData,
    formRef,

    initialFormData,
    formRules,
    frequencyOptions,

    drawerTitle,
    isFormChanged,

    resetForm,
    resetState,
    updateForm,
    updateOriginalState,
    submitFormData,
  };
}
