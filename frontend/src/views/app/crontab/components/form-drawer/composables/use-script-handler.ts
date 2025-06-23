import { ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message, SelectOption } from '@arco-design/web-vue';
import {
  getScriptCategoryListApi,
  getScriptListApi,
  getScriptDetailApi,
} from '@/api/script';
import { getCrontabCategoryListApi } from '@/api/crontab';
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

  // 脚本源分类状态 - 新增
  const scriptSourceCategoryLoading = ref(false);
  const scriptSourceCategoryOptions = ref<SelectOption[]>([]);
  const selectedScriptSourceCategory = ref<string>();

  // 获取计划任务分类
  const fetchCategories = async () => {
    categoryLoading.value = true;
    categoryOptions.value = [];
    try {
      const res = await getCrontabCategoryListApi({
        type: formState.type,
        page: 1,
        page_size: 1000,
      });

      categoryOptions.value = res.items.map((item) => {
        return {
          label: item.name,
          value: item.name,
        };
      });

      if (
        formState.category &&
        !categoryOptions.value.some(
          (opt) => typeof opt === 'object' && opt.value === formState.category
        )
      ) {
        categoryOptions.value.push({
          label: formState.category,
          value: formState.category,
        });
      }

      if (categoryOptions.value.length === 0) {
        Message.info(t('app.crontab.form.script_category.no_categories'));
      }
    } catch (err) {
      // 错误已捕获
    } finally {
      categoryLoading.value = false;
    }
  };

  // 获取脚本分类
  const fetchScriptSourceCategories = async () => {
    scriptSourceCategoryLoading.value = true;
    scriptSourceCategoryOptions.value = [];
    try {
      // 从crontab类型转换为对应的script类型
      const scriptType =
        formState.type === CRONTAB_TYPE.Global
          ? SCRIPT_TYPE.Global
          : SCRIPT_TYPE.Local;

      const result = await getScriptCategoryListApi({
        type: scriptType,
        page: 1,
        page_size: 100,
        host: currentHostId.value,
      } as any);

      if (result && result.items && Array.isArray(result.items)) {
        scriptSourceCategoryOptions.value = result.items.map((item) => ({
          label: item.name,
          value: item.name,
        }));

        if (scriptSourceCategoryOptions.value.length === 0) {
          Message.info(t('app.crontab.form.script_category.no_categories'));
        }
      } else {
        Message.error(t('app.crontab.form.script_category.invalid_response'));
      }
    } catch (error) {
      Message.error(t('app.crontab.form.script_category.fetch_error'));
    } finally {
      scriptSourceCategoryLoading.value = false;
    }
  };

  // 获取选定分类的脚本列表
  const fetchScripts = async () => {
    if (!selectedScriptSourceCategory.value) {
      scriptOptions.value = [];
      return;
    }

    scriptsLoading.value = true;
    scriptOptions.value = [];
    try {
      const scriptType =
        formState.type === CRONTAB_TYPE.Global
          ? SCRIPT_TYPE.Global
          : SCRIPT_TYPE.Local;

      const result = await getScriptListApi({
        type: scriptType,
        category: selectedScriptSourceCategory.value,
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
        Message.error(t('app.crontab.form.script_name.invalid_response'));
      }
    } catch (error) {
      Message.error(t('app.crontab.form.script_name.fetch_error'));
    } finally {
      scriptsLoading.value = false;
    }
  };

  // 获取脚本内容
  const fetchScriptContent = async () => {
    if (!selectedScriptSourceCategory.value || !selectedScript.value) {
      formState.content = '';
      scriptContent.value = '';
      return;
    }

    try {
      const script = await getScriptDetailApi({
        name: selectedScript.value,
        category: selectedScriptSourceCategory.value,
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
          selectedScriptSourceCategory,
          selectedScript,
          scriptParams,
          currentHostId.value
        );
      } else {
        Message.error(t('app.crontab.form.script_content.invalid_response'));
        formState.content = '';
        scriptContent.value = '';
      }
    } catch (error) {
      Message.error(t('app.crontab.form.script_content.fetch_error'));
      formState.content = '';
      scriptContent.value = '';
    }
  };

  // 处理分类变更
  const handleCategoryChange = (
    value:
      | string
      | number
      | Record<string, any>
      | (string | number | Record<string, any>)[]
      | boolean
  ) => {
    if (value !== undefined && value !== null) {
      const categoryValue = String(value);
      formState.category = categoryValue;
    } else {
      formState.category = '';
    }
  };

  // 处理脚本源分类变更
  const handleScriptSourceCategoryChange = () => {
    selectedScript.value = undefined;
    formState.content = '';
    scriptContent.value = '';
    fetchScripts();
  };

  // 处理脚本变更
  const handleScriptChange = async () => {
    flags.isUpdatingFromPeriod.value = true;

    try {
      await fetchScriptContent();

      await nextTick();

      if (selectedScript.value && selectedScriptSourceCategory.value) {
        await updateContentWithParams(
          formState,
          selectedScriptSourceCategory,
          selectedScript,
          scriptParams,
          currentHostId.value
        );
      }
    } catch (error) {
      // 错误已静默处理
    } finally {
      setTimeout(() => {
        flags.isUpdatingFromPeriod.value = false;
      }, 100);
    }
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
    scriptSourceCategoryLoading,
    scriptSourceCategoryOptions,
    fetchScriptSourceCategories,
    selectedScriptSourceCategory,
    handleScriptSourceCategoryChange,
  };
};
