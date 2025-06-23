import { ref, Ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import {
  createUpdateCrontabRawApi,
  createCrontabCategoryApi,
  getCrontabCategoryListApi,
} from '@/api/crontab';
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

  const ensureCategoryExists = async (category: string): Promise<boolean> => {
    if (!category.trim()) return true;

    try {
      // 获取当前分类列表
      const resp = await getCrontabCategoryListApi({
        type: formState.type,
        page: 1,
        page_size: 1000,
      });

      const existingCategories = resp.items.map((item) => item.name);

      // 如果已存在，直接返回true
      if (existingCategories.includes(category)) {
        return true;
      }

      // 需要创建新分类
      await createCrontabCategoryApi({
        type: formState.type,
        category,
      });
      return true;
    } catch (err) {
      console.error('分类处理失败:', err);
      return false;
    }
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

      // 步骤2: 确保分类存在
      formState.category = formState.category || '';
      const originalCategory = formState.category.trim();
      const categoryCreated = await ensureCategoryExists(originalCategory);

      if (!categoryCreated) {
        throw new Error(t('app.crontab.form.category.create.failed'));
      }

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

      // 通知分类变更
      if (originalCategory && onCategoryChange) {
        onCategoryChange(originalCategory);
      }

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
