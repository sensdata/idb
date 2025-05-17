<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="
      isEdit
        ? t('app.crontab.form.title.edit')
        : t('app.crontab.form.title.add')
    "
    unmountOnClose
    :ok-loading="submitLoading"
    @ok="handleOk"
    @cancel="handleCancel"
    @before-open="handleBeforeOpen"
    @before-close="handleBeforeClose"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form ref="formRef" :model="formState" :rules="rules">
        <a-form-item field="name" :label="t('app.crontab.form.name.label')">
          <a-input
            v-model="formState.name"
            class="w-[368px]"
            :placeholder="t('app.crontab.form.name.placeholder')"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item
          field="category"
          :label="t('app.crontab.form.category.label')"
        >
          <a-select
            v-model="formState.category"
            class="w-[368px]"
            :placeholder="t('app.crontab.form.category.placeholder')"
            :loading="categoryLoading"
            :options="categoryOptions"
            allow-clear
            allow-create
            @change="handleCategoryChange"
            @visible-change="(visible: boolean) => visible && fetchCategories()"
          />
        </a-form-item>
        <a-form-item :label="t('app.crontab.form.content_mode.label')">
          <a-radio-group
            v-model="formState.content_mode"
            :options="[
              {
                label: t('app.crontab.form.content_mode.direct'),
                value: 'direct',
              },
              {
                label: t('app.crontab.form.content_mode.script'),
                value: 'script',
              },
            ]"
            @change="handleContentModeChange"
          />
        </a-form-item>

        <template v-if="formState.content_mode === 'script'">
          <a-form-item :label="t('app.crontab.form.script_source.label')">
            <a-select
              v-model="selectedScriptSourceCategory"
              class="w-[368px]"
              :loading="scriptSourceCategoryLoading"
              :placeholder="t('app.crontab.form.script_category.placeholder')"
              :options="scriptSourceCategoryOptions"
              @change="handleScriptSourceCategoryChange"
            />
          </a-form-item>
          <a-form-item v-if="selectedScriptSourceCategory">
            <a-select
              v-model="selectedScript"
              class="w-[368px]"
              :loading="scriptsLoading"
              :placeholder="t('app.crontab.form.script_name.placeholder')"
              :options="scriptOptions"
              :disabled="scriptOptions.length === 0"
              @change="handleScriptChange"
            />
            <div v-if="scriptsLoading" class="text-sm mt-1 text-gray-500">
              {{ t('app.crontab.form.script_name.loading') }}
            </div>
            <div
              v-else-if="
                scriptOptions.length === 0 && selectedScriptSourceCategory
              "
              class="text-sm mt-1 text-gray-500"
            >
              {{ t('app.crontab.form.script_name.no_scripts') }}
            </div>
          </a-form-item>
          <a-form-item
            v-if="selectedScript"
            :label="t('app.crontab.form.script_params.label')"
          >
            <a-input
              v-model="scriptParams"
              class="w-[368px]"
              :placeholder="t('app.crontab.form.script_params.placeholder')"
              @change="updateScriptContent"
            />
          </a-form-item>
        </template>

        <a-form-item
          field="period_details"
          :label="t('app.crontab.form.period.label')"
        >
          <PeriodInput
            v-model="formState.period_details"
            @update:model-value="handlePeriodChange"
          />
        </a-form-item>

        <template v-if="formState.content_mode === 'direct'">
          <a-form-item
            field="command"
            :label="t('app.crontab.form.command.label')"
          >
            <a-input
              v-model="formState.command"
              :placeholder="t('app.crontab.form.command.placeholder')"
              @update:model-value="handleCommandChange"
            />
          </a-form-item>
        </template>

        <a-form-item
          field="content"
          :label="t('app.crontab.form.content.label')"
        >
          <a-spin :loading="loading" style="width: 100%">
            <ShellEditor
              :key="`content-${contentEditorKey}`"
              v-model="formState.content"
              :readonly="true"
              @blur="handleContentBlur"
            />
          </a-spin>
        </a-form-item>

        <a-form-item field="mark" :label="t('app.crontab.form.mark.label')">
          <a-textarea
            v-model="formState.mark"
            :placeholder="t('app.crontab.form.mark.placeholder')"
            @blur="handleMarkBlur"
          />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import usetCurrentHost from '@/hooks/current-host';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { CrontabEntity } from '@/entity/Crontab';
  import ShellEditor from '@/components/shell-editor/index.vue';
  import PeriodInput from '../period-input/index.vue';

  // 引入自定义钩子函数
  import { useFormState } from './hooks/use-form-state';
  import { useContentHandler } from './hooks/use-content-handler';
  import { useScriptHandler } from './hooks/use-script-handler';
  import { useEventHandlers } from './hooks/use-event-handlers';
  import { useDataLoader } from './hooks/use-data-loader';
  import { useFormSubmit } from './hooks/use-form-submit';

  const props = defineProps({
    type: {
      type: String as () => CRONTAB_TYPE,
      default: CRONTAB_TYPE.Local,
    },
  });

  const emit = defineEmits<{
    (e: 'ok', category?: string): void;
    (e: 'categoryChange', category: string): void;
  }>();

  // 基础状态
  const loading = ref(false);
  const visible = ref(false);
  const formRef = ref();
  const contentEditorKey = ref(0);

  const { t } = useI18n();

  const { currentHostId } = usetCurrentHost();

  const paramsRef = ref<{
    name?: string;
    type?: CRONTAB_TYPE;
    category?: string;
    isEdit?: boolean;
    record?: CrontabEntity;
  }>();

  // 表单状态和规则
  const { formState, createRules, flags, resetFormState } = useFormState();
  const rules = computed(() => createRules());
  const isEdit = computed(
    () => !!paramsRef.value?.isEdit || !!paramsRef.value?.name
  );

  const { updateContentWithPeriod, updateContentWithParams } =
    useContentHandler();

  // 脚本处理相关
  const scriptHandler = useScriptHandler(formState, flags, currentHostId);
  const {
    categoryLoading,
    scriptsLoading,
    categoryOptions,
    scriptOptions,
    fetchCategories,
    fetchScripts,
    handleCategoryChange,
    handleScriptChange,
    selectedScriptSourceCategory,
    selectedScript,
    scriptParams,
    scriptSourceCategoryLoading,
    scriptSourceCategoryOptions,
    fetchScriptSourceCategories,
    handleScriptSourceCategoryChange,
  } = scriptHandler;

  const selections = {
    selectedScriptSourceCategory,
    selectedScript,
    scriptParams,
  };

  const dataLoader = useDataLoader(
    formState,
    flags,
    selections,
    fetchScripts,
    currentHostId.value
  );

  const {
    handleContentBlur,
    handleMarkBlur,
    handleCommandChange,
    handlePeriodChange,
    initializePeriodDetails,
    handleContentModeChange: initializeContentModeChange,
  } = useEventHandlers(formState, flags, selections);

  const formSubmit = useFormSubmit(
    formState,
    flags,
    selections,
    currentHostId.value
  );
  const { submitLoading } = formSubmit;

  // 设置脚本模式
  const setupScriptMode = () => {
    selectedScriptSourceCategory.value = undefined;
    selectedScript.value = undefined;
    scriptParams.value = '';
    fetchScriptSourceCategories();
  };

  // 设置直接命令模式
  const setupDirectMode = () => {
    fetchCategories();
  };

  // 刷新内容编辑器
  const refreshContentEditor = () => {
    contentEditorKey.value++;
  };

  // 处理内容模式变更
  const handleContentModeChange = () => {
    formState.command = '';

    if (formState.content_mode === 'script') {
      setupScriptMode();
    } else {
      setupDirectMode();
    }

    initializeContentModeChange(fetchCategories);
    refreshContentEditor();
  };

  // 更新脚本内容
  const updateScriptContent = async () => {
    if (!selectedScriptSourceCategory.value || !selectedScript.value) return;

    await updateContentWithParams(
      formState,
      selections.selectedScriptSourceCategory,
      selections.selectedScript,
      selections.scriptParams,
      currentHostId.value
    );
    refreshContentEditor();
  };

  // 加载数据
  async function loadData() {
    loading.value = true;

    try {
      if (paramsRef.value?.record) {
        await dataLoader.loadDataFromRecord(paramsRef.value.record);
      } else if (paramsRef.value) {
        await dataLoader.loadData(paramsRef);
      }
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      loading.value = false;
    }
  }

  // 设置表单类型
  const setFormType = () => {
    formState.type = props.type;
  };

  // 加载编辑器数据
  const loadEditorData = async () => {
    if (isEdit.value) {
      if (paramsRef.value?.record) {
        await dataLoader.loadDataFromRecord(paramsRef.value.record);
        if (paramsRef.value?.category) {
          formState.category = paramsRef.value.category;
        }
      } else if (paramsRef.value) {
        await loadData();
      }
    } else {
      resetFormState();
      setFormType();
      if (paramsRef.value?.category) {
        formState.category = paramsRef.value.category;
      }
      initializePeriodDetails();
      formState.content = '';
      formState.command = '';
    }

    if (formState.content_mode === 'script' && selectedScript.value) {
      await updateContentWithParams(
        formState,
        selections.selectedScriptSourceCategory,
        selections.selectedScript,
        selections.scriptParams,
        currentHostId.value
      );
    } else {
      await updateContentWithPeriod(formState, flags, true);
    }

    refreshContentEditor();
  };

  // 设置参数
  const setParams = (params: {
    name?: string;
    type?: CRONTAB_TYPE;
    category?: string;
    isEdit?: boolean;
    record?: CrontabEntity;
  }) => {
    paramsRef.value = params;
  };

  // 处理表单提交
  const handleOk = async () => {
    try {
      await formSubmit.handleSubmit(
        formRef.value,
        () => {
          emit('ok', formState.category);
        },
        visible,
        isEdit.value,
        (category) => {
          emit('categoryChange', category);
        }
      );
    } catch (error) {
      console.error('提交表单时发生错误:', error);
    }
  };

  // 处理取消操作
  const handleCancel = () => {
    visible.value = false;
  };

  // 抽屉打开前处理
  const handleBeforeOpen = () => {
    fetchCategories();
    if (formState.content_mode === 'script') {
      fetchScriptSourceCategories();
    }
    if (!isEdit.value && paramsRef.value?.category) {
      formState.category = paramsRef.value.category;
    }
    setFormType();
  };

  // 抽屉关闭前处理
  const handleBeforeClose = () => {
    if (!isEdit.value) {
      resetFormState();
      setFormType();
      selectedScriptSourceCategory.value = undefined;
      selectedScript.value = undefined;
      scriptParams.value = '';
    }
    paramsRef.value = undefined;
  };

  // 显示抽屉
  const show = async (params?: {
    name?: string;
    type?: CRONTAB_TYPE;
    category?: string;
    isEdit?: boolean;
    record?: CrontabEntity;
  }) => {
    if (params?.record || params?.name) {
      setParams({
        name: params.name || params.record?.name,
        type: params.type || props.type,
        category: params.category || '',
        isEdit: true,
        record: params.record,
      });
    } else {
      paramsRef.value = {
        type: props.type,
        category: params?.category || '',
        isEdit: false,
      };
      formState.category = params?.category || '';
    }

    visible.value = true;
    await loadEditorData();
  };

  // 隐藏抽屉
  const hide = () => {
    visible.value = false;
  };

  // 监听表单类型变化
  watch(
    () => formState.type,
    () => {
      if (formState.content_mode === 'script') {
        setupScriptMode();
      }
      fetchCategories();
    }
  );

  // 暴露组件方法
  defineExpose({
    setParams,
    loadData,
    show,
    hide,
  });
</script>
