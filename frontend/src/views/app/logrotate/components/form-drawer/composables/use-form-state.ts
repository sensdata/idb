import { ref, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useForm } from '@/composables/use-form';
import { LOGROTATE_FREQUENCY, LOGROTATE_TYPE } from '@/config/enum';
import { DEFAULT_LOGROTATE_CATEGORY } from '../../../constants';
import type { ActiveMode, FormData } from '../types';

export function useFormState() {
  const { t } = useI18n();

  const activeMode = ref<ActiveMode>('overview');
  const previousMode = ref<ActiveMode>('overview');
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

  const { formData, resetForm, updateForm } = useForm<FormData>({
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

  // 重置状态
  const resetState = async () => {
    resetForm();
    activeMode.value = 'overview';
    previousMode.value = 'overview';
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

    initialFormData,
    frequencyOptions,

    drawerTitle,

    resetForm,
    resetState,
    updateForm,
    updateOriginalState,
  };
}
