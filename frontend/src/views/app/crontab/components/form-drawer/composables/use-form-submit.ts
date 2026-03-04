import { ref, Ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { createUpdateCrontabRawApi } from '@/api/crontab';
import { useContentHandler } from './use-content-handler';
import { FormState, StateFlags } from './use-form-state';

interface ScriptSelections {
  selectedScriptSourceCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

export const useFormSubmit = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections,
  currentHostId?: any
) => {
  const DEFAULT_CRONTAB_CATEGORY = 'default';
  const { t } = useI18n();
  const submitLoading = ref(false);

  const {
    updateContentWithPeriod,
    updateMarkInScriptMode,
    updateContentFromMark,
  } = useContentHandler();

  const validate = async (formRef: any) => {
    if (!formRef) return false;

    return formRef.validate().then((errors: any) => {
      return !errors;
    });
  };

  /**
   * 更新提交前的内容，确保备注和周期信息被正确更新到content中
   */
  const updateContentBeforeSubmit = async (): Promise<void> => {
    flags.isUpdatingFromPeriod.value = true;

    try {
      // 在提交前先确保备注内容被更新到content中
      await updateContentFromMark(formState);

      // 根据内容模式选择不同的更新方法
      if (formState.content_mode === 'script') {
        if (
          selections.selectedScript.value &&
          selections.selectedScriptSourceCategory.value
        ) {
          await updateMarkInScriptMode(
            formState,
            formState.mark,
            selections.selectedScriptSourceCategory,
            selections.selectedScript,
            selections.scriptParams,
            currentHostId
          );
        } else {
          await updateContentWithPeriod(formState, flags, true);
        }
      } else {
        await updateContentWithPeriod(formState, flags, true);
      }

      // 确保内容更新完成
      await nextTick();
    } finally {
      await nextTick();
      flags.isUpdatingFromPeriod.value = false;
    }
  };

  /**
   * 处理表单提交
   * 简化的错误处理流程，更清晰的执行步骤
   */
  const handleSubmit = async (
    formRef: any,
    onSuccess: () => void,
    visible: Ref<boolean>,
    isEdit: boolean,
    onCategoryChange?: (category: string) => void
  ) => {
    // 表单验证
    if (!(await validate(formRef))) {
      return;
    }

    // 防止重复提交
    if (submitLoading.value) {
      return;
    }

    submitLoading.value = true;

    try {
      // 步骤1: 更新内容（确保包含周期信息和标记）
      await updateContentBeforeSubmit();

      // 步骤2: 固定使用default分类（分类对用户侧收敛）
      const originalCategory = DEFAULT_CRONTAB_CATEGORY;
      formState.category = originalCategory;

      // 步骤3: 创建/更新定时任务
      await createUpdateCrontabRawApi({
        type: formState.type,
        category: originalCategory,
        name: formState.name,
        content: formState.content,
        isEdit,
      });

      // 步骤4: 处理成功响应
      visible.value = false;

      if (onCategoryChange) onCategoryChange(originalCategory);

      Message.success(t('app.crontab.form.success'));
      onSuccess();
    } catch (error: any) {
      console.error('Error submitting form:', error);
      Message.error(error?.message || t('app.crontab.form.error'));
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
