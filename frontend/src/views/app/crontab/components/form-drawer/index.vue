<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="
      isEdit
        ? $t('app.crontab.form.title.edit')
        : $t('app.crontab.form.title.add')
    "
    unmountOnClose
    :ok-loading="submitLoading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form ref="formRef" :model="formState" :rules="rules">
        <a-form-item field="name" :label="$t('app.crontab.form.name.label')">
          <a-input
            v-model="formState.name"
            class="w-[368px]"
            :placeholder="$t('app.crontab.form.name.placeholder')"
          />
        </a-form-item>
        <a-form-item field="type" :label="$t('app.crontab.form.type.label')">
          <a-radio-group v-model="formState.type" :options="typeOptions" />
        </a-form-item>
        <a-form-item :label="$t('app.crontab.form.content_mode.label')">
          <a-radio-group
            v-model="formState.content_mode"
            :options="[
              {
                label: $t('app.crontab.form.content_mode.direct'),
                value: 'direct',
              },
              {
                label: $t('app.crontab.form.content_mode.script'),
                value: 'script',
              },
            ]"
            @change="handleContentModeChange"
          />
        </a-form-item>

        <template v-if="formState.content_mode === 'script'">
          <a-form-item :label="$t('app.crontab.form.script_source.label')">
            <a-select
              v-model="selectedCategory"
              class="w-[368px]"
              :loading="categoryLoading"
              :placeholder="$t('app.crontab.form.script_category.placeholder')"
              :options="categoryOptions"
              @change="handleCategoryChange"
            />
          </a-form-item>
          <a-form-item v-if="selectedCategory">
            <a-select
              v-model="selectedScript"
              class="w-[368px]"
              :loading="scriptsLoading"
              :placeholder="$t('app.crontab.form.script_name.placeholder')"
              :options="scriptOptions"
              :disabled="scriptOptions.length === 0"
              @change="handleScriptChange"
            />
            <div v-if="scriptsLoading" class="text-sm mt-1 text-gray-500">
              {{ $t('app.crontab.form.script_name.loading') }}
            </div>
            <div
              v-else-if="scriptOptions.length === 0 && selectedCategory"
              class="text-sm mt-1 text-gray-500"
            >
              {{ $t('app.crontab.form.script_name.no_scripts') }}
            </div>
          </a-form-item>
          <a-form-item
            v-if="selectedScript"
            :label="$t('app.crontab.form.script_params.label')"
          >
            <a-input
              v-model="scriptParams"
              class="w-[368px]"
              :placeholder="$t('app.crontab.form.script_params.placeholder')"
              @change="
                () => {
                  if (selectedCategory && selectedScript && scriptParams) {
                    updateContentWithParams(
                      formState,
                      selections.selectedCategory,
                      selections.selectedScript,
                      selections.scriptParams
                    );
                  }
                }
              "
            />
          </a-form-item>
        </template>

        <a-form-item
          field="period_details"
          :label="$t('app.crontab.form.period.label')"
        >
          <PeriodInput
            v-model="formState.period_details"
            @update:model-value="handlePeriodChange"
          />
        </a-form-item>

        <a-form-item
          field="content"
          :label="$t('app.crontab.form.content.label')"
        >
          <ShellEditor
            v-model="formState.content"
            :readonly="formState.content_mode === 'script'"
            @update:model-value="handleContentChange"
            @blur="handleContentBlur"
          />
        </a-form-item>

        <a-form-item field="mark" :label="$t('app.crontab.form.mark.label')">
          <a-textarea
            v-model="formState.mark"
            :placeholder="$t('app.crontab.form.mark.placeholder')"
            @update:model-value="handleMarkValueChange"
            @change="handleMarkValueChange"
            @blur="handleMarkBlur"
            @input="handleMarkInput"
          />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, nextTick, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import usetCurrentHost from '@/hooks/current-host';
  import ShellEditor from '@/components/shell-editor/index.vue';
  import PeriodInput from '../period-input/index.vue';

  import { useFormState } from './hooks/use-form-state';
  import { useContentHandler } from './hooks/use-content-handler';
  import { useScriptHandler } from './hooks/use-script-handler';
  import { useEventHandlers } from './hooks/use-event-handlers';
  import { useDataLoader } from './hooks/use-data-loader';
  import { useFormSubmit } from './hooks/use-form-submit';

  const emit = defineEmits(['ok']);

  const loading = ref(false);
  const visible = ref(false);
  const formRef = ref();

  // We still need useI18n() for template translations but don't need to use 't' directly in script
  useI18n();

  const { currentHostId } = usetCurrentHost();

  const { formState, createRules, getTypeOptions, flags } = useFormState();

  const rules = createRules();
  const typeOptions = ref(getTypeOptions());

  const {
    updateContentWithPeriod,
    updateContentWithParams,
    updateMarkInScriptMode,
  } = useContentHandler();

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
    selectedCategory,
    selectedScript,
    scriptParams,
  } = scriptHandler;

  const selections = {
    selectedCategory,
    selectedScript,
    scriptParams,
  };

  const {
    handleContentChange,
    handleMarkInput,
    handleMarkBlur,
    handleMarkValueChange,
    handleContentBlur,
    handlePeriodChange,
    initializePeriodDetails,
    handleContentModeChange: initializeContentModeChange,
  } = useEventHandlers(formState, flags, selections);

  const handleContentModeChange = () => {
    initializeContentModeChange(fetchCategories);
  };

  const dataLoader = useDataLoader(formState, flags, selections, fetchScripts);

  const formSubmit = useFormSubmit(formState, flags, selections);

  const { submitLoading } = formSubmit;

  const paramsRef = ref<{ id: number }>();
  const isEdit = computed(() => !!paramsRef.value?.id);

  watch(
    () => visible.value,
    (val) => {
      if (val) {
        fetchCategories();
        if (!isEdit.value) {
          initializePeriodDetails();

          formState.content = '';

          if (formState.content_mode === 'script' && selectedScript.value) {
            updateContentWithParams(
              formState,
              selections.selectedCategory,
              selections.selectedScript,
              selections.scriptParams
            );
          } else {
            updateContentWithPeriod(formState, flags, true);
          }
        }
      }
    }
  );

  watch(
    () => formState.type,
    () => {
      if (formState.content_mode === 'script') {
        selectedCategory.value = undefined;
        selectedScript.value = undefined;
        scriptParams.value = '';
        formState.content = '';

        fetchCategories();
      }
    }
  );

  watch(
    () => formState.content,
    (newContent) => {
      if (
        formState.content_mode === 'direct' &&
        !flags.isUpdatingFromPeriod.value &&
        !flags.isInitialLoad.value
      ) {
        handleContentChange(newContent);
      }
    }
  );

  watch(
    () => formState.mark,
    (newMark) => {
      flags.isInitialLoad.value = false;

      const wasUpdating = flags.isUpdatingFromPeriod.value;

      if (!wasUpdating) {
        flags.isUpdatingFromPeriod.value = true;
      }

      try {
        if (
          formState.content_mode === 'script' &&
          selectedScript.value &&
          selectedCategory.value
        ) {
          updateMarkInScriptMode(
            formState,
            newMark,
            selections.selectedCategory,
            selections.selectedScript,
            selections.scriptParams
          );
        } else {
          updateContentWithPeriod(formState, flags, true);
        }
      } finally {
        if (!wasUpdating) {
          nextTick(() => {
            flags.isUpdatingFromPeriod.value = false;
          });
        }
      }
    },
    { immediate: false }
  );

  watch(
    () => formState.period_details,
    () => {
      if (flags.userEditingContent.value || flags.isUpdatingFromPeriod.value) {
        return;
      }

      if (formState.content_mode === 'script') {
        if (selectedScript.value && selectedCategory.value) {
          updateContentWithParams(
            formState,
            selections.selectedCategory,
            selections.selectedScript,
            selections.scriptParams
          );
        } else {
          updateContentWithPeriod(formState, flags, true);
        }
      } else {
        updateContentWithPeriod(formState, flags, true);
      }
    },
    { deep: true, immediate: true }
  );

  const setParams = (params: { id: number }) => {
    paramsRef.value = params;
  };

  async function load() {
    loading.value = true;

    try {
      await dataLoader.loadData(paramsRef);
    } finally {
      loading.value = false;
    }
  }

  const handleOk = async () => {
    await formSubmit.handleSubmit(formRef.value, () => emit('ok'), visible);
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  defineExpose({
    setParams,
    load,
    show,
    hide,
  });
</script>
