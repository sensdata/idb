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
      <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
        <a-form-item field="name" :label="t('app.crontab.form.name.label')">
          <a-input
            v-model="formState.name"
            :placeholder="t('app.crontab.form.name.placeholder')"
            :disabled="isEdit"
            class="form-input"
          />
        </a-form-item>

        <a-form-item
          field="category"
          :label="t('app.crontab.form.category.label')"
        >
          <a-select
            v-model="formState.category"
            :placeholder="t('app.crontab.form.category.placeholder')"
            :loading="categoryLoading"
            :options="categoryOptions"
            allow-clear
            allow-create
            class="form-input"
            @change="handleCategoryChange"
            @visible-change="(visible: boolean) => visible && fetchCategories()"
          />
        </a-form-item>

        <a-form-item>
          <template #label>
            {{ t('app.crontab.form.content_mode.label') }}
          </template>
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
          <a-form-item>
            <template #label>
              {{ t('app.crontab.form.script_source.label') }}
            </template>
            <a-select
              v-model="selectedScriptSourceCategory"
              :loading="scriptSourceCategoryLoading"
              :placeholder="t('app.crontab.form.script_category.placeholder')"
              :options="scriptSourceCategoryOptions"
              class="form-input"
              @change="handleScriptSourceCategoryChange"
            />
          </a-form-item>
          <a-form-item v-if="selectedScriptSourceCategory">
            <a-select
              v-model="selectedScript"
              :loading="scriptsLoading"
              :placeholder="t('app.crontab.form.script_name.placeholder')"
              :options="scriptOptions"
              :disabled="scriptOptions.length === 0"
              class="form-input"
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
          <a-form-item v-if="selectedScript">
            <template #label>
              {{ t('app.crontab.form.script_params.label') }}
            </template>
            <a-input
              v-model="scriptParams"
              :placeholder="t('app.crontab.form.script_params.placeholder')"
              class="form-input"
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
          <a-form-item field="command">
            <template #label>
              {{ t('app.crontab.form.command.label') }}
            </template>
            <a-input
              v-model="formState.command"
              :placeholder="t('app.crontab.form.command.placeholder')"
              class="form-input"
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
              class="form-code-editor"
              @blur="handleContentBlur"
            />
          </a-spin>
        </a-form-item>

        <a-form-item field="mark">
          <template #label>
            {{ t('app.crontab.form.mark.label') }}
          </template>
          <a-textarea
            v-model="formState.mark"
            :placeholder="t('app.crontab.form.mark.placeholder')"
            class="form-textarea"
            :rows="4"
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
  import usetCurrentHost from '@/composables/current-host';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { CrontabEntity } from '@/entity/Crontab';
  import ShellEditor from '@/components/shell-editor/index.vue';
  import PeriodInput from '../period-input/index.vue';

  // 引入自定义钩子函数
  import { useFormState } from './composables/use-form-state';
  import { useContentHandler } from './composables/use-content-handler';
  import { useScriptHandler } from './composables/use-script-handler';
  import { useEventHandlers } from './composables/use-event-handlers';
  import { useDataLoader } from './composables/use-data-loader';
  import { useFormSubmit } from './composables/use-form-submit';

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

    // 记录当前的更新状态，防止重复更新
    flags.isUpdatingFromPeriod.value = true;

    try {
      await updateContentWithParams(
        formState,
        selections.selectedScriptSourceCategory,
        selections.selectedScript,
        selections.scriptParams,
        currentHostId.value
      );
      refreshContentEditor();
    } catch (error) {
      console.error('更新脚本内容时发生错误:', error);
    } finally {
      // 确保状态标志被重置
      setTimeout(() => {
        flags.isUpdatingFromPeriod.value = false;
      }, 100);
    }
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
        // 优先级：如果从列表视图显式传递了类别，则使用它
        // 这确保表单使用与列表视图相同的类别
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
        (category: string) => {
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
      // 编辑模式：设置参数
      setParams({
        name: params.name || params.record?.name,
        type: params.type || props.type,
        category: params.category || '',
        isEdit: true,
        record: params.record,
      });
    } else {
      // 添加模式：设置默认参数
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

<style scoped lang="less">
  .form-input {
    width: 100%;
  }

  .form-textarea {
    width: 100%;
    resize: vertical;
  }

  .form-code-editor {
    width: 100%;
    min-height: 200px;
    background-color: #ffffff !important;
    border: 1px solid #e5e6eb;
    border-radius: 6px;

    // 强制覆盖所有可能的黑色背景
    * {
      background-color: #ffffff !important;
    }

    // 针对可能的深色元素进行全面覆盖
    div,
    span,
    pre,
    code {
      background-color: #ffffff !important;
      color: #1d2129 !important;
    }

    // 覆盖代码编辑器的深色主题
    :deep(.cm-editor) {
      background-color: #ffffff !important;
      color: #1d2129 !important;
    }

    :deep(.cm-content) {
      background-color: #ffffff !important;
      color: #1d2129 !important;
    }

    :deep(.cm-focused) {
      background-color: #ffffff !important;
    }

    :deep(.cm-scroller) {
      background-color: #ffffff !important;
    }

    // 行号区域样式覆盖
    :deep(.cm-gutters) {
      background-color: #ffffff !important;
      border-right: 1px solid #e5e6eb !important;
    }

    :deep(.cm-gutter) {
      background-color: #ffffff !important;
    }

    :deep(.cm-lineNumbers) {
      background-color: #ffffff !important;
    }

    :deep(.cm-lineNumbers .cm-gutterElement) {
      background-color: #ffffff !important;
      color: #86909c !important;
    }

    // Monaco Editor 样式覆盖（如果使用的是Monaco）
    :deep(.monaco-editor) {
      background-color: #ffffff !important;
    }

    :deep(.monaco-editor .margin) {
      background-color: #ffffff !important;
    }

    :deep(.monaco-editor .monaco-editor-background) {
      background-color: #ffffff !important;
    }

    :deep(.monaco-editor .margin-view-overlays) {
      background-color: #ffffff !important;
    }

    :deep(.monaco-editor .line-numbers) {
      background-color: #ffffff !important;
      color: #86909c !important;
    }

    // 通用行号样式覆盖（针对任何可能的编辑器）
    :deep(.line-numbers) {
      background-color: #ffffff !important;
      color: #86909c !important;
    }

    :deep(.gutter) {
      background-color: #ffffff !important;
    }

    :deep(.CodeMirror-gutters) {
      background-color: #ffffff !important;
      border-right: 1px solid #e5e6eb !important;
    }

    :deep(.CodeMirror-linenumber) {
      background-color: #ffffff !important;
      color: #86909c !important;
    }

    // 额外的深度样式覆盖，确保没有任何黑色残留
    :deep(.cm-theme-dark) {
      background-color: #ffffff !important;
    }

    :deep(.cm-editor.cm-focused) {
      outline: none !important;
    }

    :deep(.view-line) {
      background-color: #ffffff !important;
    }

    :deep(.current-line) {
      background-color: #ffffff !important;
    }

    // 针对可能的深色主题类
    :deep(.dark) {
      background-color: #ffffff !important;
    }

    :deep(.theme-dark) {
      background-color: #ffffff !important;
    }

    // 针对行背景
    :deep(.line) {
      background-color: #ffffff !important;
    }
  }

  :deep(.arco-form-item-label) {
    font-weight: 500;
    margin-bottom: 8px;
    color: #1d2129;
  }

  :deep(.arco-form-item) {
    margin-bottom: 20px;
  }

  :deep(.arco-radio-group) {
    .arco-radio-button {
      margin-right: 16px;
    }
  }
</style>
