import { reactive, ref, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { SelectOption } from '@arco-design/web-vue';
import { RadioOption } from '@arco-design/web-vue/es/radio/interface';
import { PeriodDetailDo } from '@/entity/Crontab';
import { CRONTAB_KIND, CRONTAB_TYPE } from '@/config/enum';

// 定义表单状态接口
export interface FormState {
  name: string;
  type: CRONTAB_TYPE;
  kind: CRONTAB_KIND;
  content: string;
  content_mode: 'direct' | 'script';
  period_details: PeriodDetailDo[];
  mark: string;
  command: string;
  user: string;
  category: string;
  id?: number;
}

// 状态标志接口
export interface StateFlags {
  isInitialLoad: Ref<boolean>;
  isUpdatingFromPeriod: Ref<boolean>;
  userEditingContent: Ref<boolean>;
}

// 创建默认表单状态
const createDefaultFormState = (): FormState => ({
  name: '',
  type: CRONTAB_TYPE.Local,
  kind: CRONTAB_KIND.Shell,
  content: '',
  content_mode: 'direct',
  period_details: [],
  mark: '',
  command: '',
  user: 'root',
  category: '',
});

export const useFormState = () => {
  const { t } = useI18n();

  // 表单状态
  const formState = reactive<FormState>(createDefaultFormState());

  // 重置表单状态的方法
  const resetFormState = () => {
    const defaultState = createDefaultFormState();
    Object.keys(formState).forEach((key) => {
      // @ts-ignore
      formState[key] = defaultState[key];
    });
  };

  // 表单验证规则
  const createRules = () => {
    // Define a type for the rules object
    type ValidationRules = {
      [key: string]: {
        required: boolean;
        message: string;
        type?: 'array' | 'string' | 'number' | 'boolean' | 'object';
      }[];
    };

    const rules: ValidationRules = {
      name: [{ required: true, message: t('app.crontab.form.name.required') }],
      category: [
        { required: true, message: t('app.crontab.form.category.required') },
      ],
      period_details: [
        {
          required: true,
          message: t('app.crontab.form.period.required'),
          type: 'array' as const,
        },
      ],
      content: [
        { required: true, message: t('app.crontab.form.content.required') },
      ],
    };

    return rules;
  };

  // 类型选项
  const getTypeOptions = (): RadioOption[] => {
    return [
      {
        label: t('app.crontab.enum.type.local'),
        value: CRONTAB_TYPE.Local,
      },
      {
        label: t('app.crontab.enum.type.global'),
        value: CRONTAB_TYPE.Global,
      },
    ];
  };

  // 脚本相关状态
  const selectedCategory = ref<string>();
  const selectedScript = ref<string>();
  const scriptParams = ref('');
  const scriptContent = ref('');
  const categoryLoading = ref(false);
  const scriptsLoading = ref(false);
  const categoryOptions = ref<SelectOption[]>([]);
  const scriptOptions = ref<SelectOption[]>([]);

  // 状态跟踪标志
  const isInitialLoad = ref(true);
  const isUpdatingFromPeriod = ref(false);
  const userEditingContent = ref(false);

  // 创建标志对象
  const flags: StateFlags = {
    isInitialLoad,
    isUpdatingFromPeriod,
    userEditingContent,
  };

  return {
    formState,
    createRules,
    getTypeOptions,
    selectedCategory,
    selectedScript,
    scriptParams,
    scriptContent,
    categoryLoading,
    scriptsLoading,
    categoryOptions,
    scriptOptions,
    flags,
    isInitialLoad,
    isUpdatingFromPeriod,
    userEditingContent,
    resetFormState,
  };
};
