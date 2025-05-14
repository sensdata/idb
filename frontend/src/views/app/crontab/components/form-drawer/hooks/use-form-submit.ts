import { ref, Ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { createUpdateCrontabRawApi } from '@/api/crontab';
import { useContentHandler } from './use-content-handler';
import { FormState, StateFlags } from './use-form-state';

interface ScriptSelections {
  selectedCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

export const useFormSubmit = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections
) => {
  const { t } = useI18n();
  const submitLoading = ref(false);

  const { updateContentWithPeriod, updateMarkInScriptMode } =
    useContentHandler();

  const validate = async (formRef: any) => {
    if (!formRef) return false;

    return formRef.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleSubmit = async (
    formRef: any,
    onSuccess: () => void,
    visible: Ref<boolean>
  ) => {
    if (!(await validate(formRef))) {
      return;
    }

    if (submitLoading.value) {
      return;
    }

    submitLoading.value = true;
    try {
      // 确保内容包含周期信息和标记
      flags.isUpdatingFromPeriod.value = true;
      try {
        if (formState.content_mode === 'script') {
          // 脚本模式下，直接内联更新内容以确保包含标记
          if (
            selections.selectedScript.value &&
            selections.selectedCategory.value
          ) {
            updateMarkInScriptMode(
              formState,
              formState.mark,
              selections.selectedCategory,
              selections.selectedScript,
              selections.scriptParams
            );
          } else {
            // 未选择脚本时，使用通用方法
            updateContentWithPeriod(formState, flags, true);
          }
        } else {
          // 直接输入模式下，使用通用方法
          updateContentWithPeriod(formState, flags, true);
        }
      } finally {
        await nextTick();
        flags.isUpdatingFromPeriod.value = false;
      }

      // 使用原始端点创建/更新定时任务
      await createUpdateCrontabRawApi({
        type: formState.type,
        category: 'crontab', // 定时任务文件的类别固定为'crontab'
        name: formState.name,
        content: formState.content,
      });

      visible.value = false;
      Message.success(t('app.crontab.form.success'));
      onSuccess();
    } catch (error) {
      console.error('Error submitting form:', error);
      Message.error(t('app.crontab.form.error'));
    } finally {
      submitLoading.value = false;
    }
  };

  return {
    submitLoading,
    validate,
    handleSubmit,
  };
};
