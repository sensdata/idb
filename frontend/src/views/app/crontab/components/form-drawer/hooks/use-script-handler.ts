import { ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message, SelectOption } from '@arco-design/web-vue';
import {
  getScriptCategoryListApi,
  getScriptListApi,
  getScriptDetailApi,
} from '@/api/script';
import { SCRIPT_TYPE, CRONTAB_TYPE } from '@/config/enum';
import { ScriptEntity } from '@/entity/Script';
import { useContentHandler } from './use-content-handler';
import { StateFlags } from './use-form-state';
import type { FormState } from './use-form-state';

export const useScriptHandler = (
  formState: FormState,
  flags: StateFlags,
  currentHostId: any
) => {
  const { t } = useI18n();
  const { updateContentWithParams } = useContentHandler();

  // 脚本选择状态
  const selectedCategory = ref<string>();
  const selectedScript = ref<string>();
  const scriptParams = ref('');
  const scriptContent = ref('');
  const categoryLoading = ref(false);
  const scriptsLoading = ref(false);
  const categoryOptions = ref<SelectOption[]>([]);
  const scriptOptions = ref<SelectOption[]>([]);

  // 获取脚本分类
  const fetchCategories = async () => {
    categoryLoading.value = true;
    categoryOptions.value = [];
    try {
      const result = await getScriptCategoryListApi({
        type:
          formState.type === CRONTAB_TYPE.Global
            ? SCRIPT_TYPE.Global
            : SCRIPT_TYPE.Local,
        page: 1,
        page_size: 100,
        host: currentHostId.value,
      } as any);

      if (result && result.items && Array.isArray(result.items)) {
        categoryOptions.value = result.items.map((item) => ({
          label: item.name,
          value: item.name,
        }));

        if (categoryOptions.value.length === 0) {
          Message.info(t('app.crontab.form.script_category.no_categories'));
        }
      } else {
        console.error('Invalid categories response format:', result);
        Message.error(t('app.crontab.form.script_category.invalid_response'));
      }
    } catch (error) {
      console.error('Failed to fetch script categories:', error);
      Message.error(t('app.crontab.form.script_category.fetch_error'));
    } finally {
      categoryLoading.value = false;
    }
  };

  // 获取选定分类的脚本列表
  const fetchScripts = async () => {
    if (!selectedCategory.value) {
      scriptOptions.value = [];
      return;
    }

    scriptsLoading.value = true;
    scriptOptions.value = [];
    try {
      const result = await getScriptListApi({
        type:
          formState.type === CRONTAB_TYPE.Global
            ? SCRIPT_TYPE.Global
            : SCRIPT_TYPE.Local,
        category: selectedCategory.value,
        page: 1,
        page_size: 100,
        host: currentHostId.value,
      } as any);

      if (result && result.items && Array.isArray(result.items)) {
        scriptOptions.value = result.items.map((item: ScriptEntity) => ({
          label: item.name,
          value: item.name,
        }));

        if (scriptOptions.value.length === 0) {
          Message.info(t('app.crontab.form.script_name.no_scripts'));
        }
      } else {
        console.error('Invalid scripts response format:', result);
        Message.error(t('app.crontab.form.script_name.invalid_response'));
      }
    } catch (error) {
      console.error('Failed to fetch scripts:', error);
      Message.error(t('app.crontab.form.script_name.fetch_error'));
    } finally {
      scriptsLoading.value = false;
    }
  };

  // 获取脚本内容
  const fetchScriptContent = async () => {
    if (!selectedCategory.value || !selectedScript.value) {
      formState.content = '';
      scriptContent.value = '';
      return;
    }

    try {
      const script = await getScriptDetailApi({
        name: selectedScript.value,
        category: selectedCategory.value,
        type:
          formState.type === CRONTAB_TYPE.Global
            ? SCRIPT_TYPE.Global
            : SCRIPT_TYPE.Local,
        host: currentHostId.value,
      } as any);

      if (script && script.content !== undefined) {
        scriptContent.value = script.content;

        updateContentWithParams(
          formState,
          selectedCategory,
          selectedScript,
          scriptParams
        );
      } else {
        console.error('Invalid script content response:', script);
        Message.error(t('app.crontab.form.script_content.invalid_response'));
        formState.content = '';
        scriptContent.value = '';
      }
    } catch (error) {
      console.error('Failed to fetch script content:', error);
      Message.error(t('app.crontab.form.script_content.fetch_error'));
      formState.content = '';
      scriptContent.value = '';
    }
  };

  // 处理分类变更
  const handleCategoryChange = () => {
    selectedScript.value = undefined;
    formState.content = '';
    scriptContent.value = '';
    fetchScripts();
  };

  // 处理脚本变更
  const handleScriptChange = () => {
    fetchScriptContent().then(() => {
      // 选择脚本时，确保包含周期和标记
      flags.isUpdatingFromPeriod.value = true;
      try {
        if (selectedScript.value && selectedCategory.value) {
          updateContentWithParams(
            formState,
            selectedCategory,
            selectedScript,
            scriptParams
          );
        }
      } catch (error) {
        console.error('Error updating content with script params:', error);
      } finally {
        nextTick(() => {
          flags.isUpdatingFromPeriod.value = false;
        });
      }
    });
  };

  return {
    selectedCategory,
    selectedScript,
    scriptParams,
    scriptContent,
    categoryLoading,
    scriptsLoading,
    categoryOptions,
    scriptOptions,
    fetchCategories,
    fetchScripts,
    fetchScriptContent,
    handleCategoryChange,
    handleScriptChange,
  };
};
