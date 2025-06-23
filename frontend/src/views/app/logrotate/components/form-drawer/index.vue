<template>
  <a-drawer
    v-model:visible="visible"
    :title="drawerTitle"
    :width="DRAWER_WIDTH"
    :footer="false"
    unmount-on-close
  >
    <div class="form-drawer">
      <a-tabs v-model:active-key="activeMode" @change="handleModeChange">
        <a-tab-pane key="form" :title="$t('app.logrotate.mode.form')">
          <FormTab
            ref="formRef"
            :form-data="formData"
            :form-rules="formRules"
            :frequency-options="frequencyOptions"
            :category-loading="categoryLoading"
            :category-options="categoryOptions"
            :is-edit="isEdit"
            @category-change="handleCategoryChange"
            @category-visible-change="handleCategoryVisibleChangeWrapper"
            @update-form-data="handleUpdateFormData"
          />
        </a-tab-pane>

        <a-tab-pane key="raw" :title="$t('app.logrotate.mode.raw')">
          <RawTab
            v-model:content="rawContent"
            :extensions="logrotateExtensions"
          />
        </a-tab-pane>
      </a-tabs>

      <div class="drawer-footer">
        <a-space>
          <a-button @click="handleCancel">
            {{ $t('common.cancel') }}
          </a-button>
          <a-button
            type="primary"
            :loading="loading"
            :disabled="!isFormChanged"
            @click="handleSubmit"
          >
            {{ $t('common.save') }}
          </a-button>
        </a-space>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import useEditorConfig from '@/components/code-editor/composables/use-editor-config';
  import { LOGROTATE_TYPE } from '@/config/enum';

  // 组合式API
  import useLoading from '@/composables/loading';
  import useCurrentHost from '@/composables/current-host';
  import useVisible from '@/composables/visible';
  import { useLogger } from '@/composables/use-logger';
  import { useRawContentParser } from './composables/use-raw-content-parser';
  import { useCategories } from './composables/use-categories';

  import { useFormState } from './composables/use-form-state';
  import { useModeManager } from './composables/use-mode-manager';
  import { useLogrotateApi } from './composables/use-logrotate-api';

  import FormTab from './form-tab.vue';
  import RawTab from './raw-tab.vue';

  import type { ShowParams } from './types';

  const DRAWER_WIDTH = 800;
  const { t } = useI18n();

  const emit = defineEmits<{
    ok: [];
    categoryChange: [category: string];
  }>();

  const { log } = useLogger('LogrotateFormDrawer');
  const { loading, setLoading } = useLoading();
  const { currentHostId } = useCurrentHost();
  const { visible, show: showDrawer, hide: hideDrawer } = useVisible();

  // 表单状态管理
  const {
    activeMode,
    previousMode,
    isEdit,
    currentType,
    originalName,
    originalCategory,
    originalFormData,
    originalRawContent,
    formData,
    formRef,
    formRules,
    frequencyOptions,
    drawerTitle,
    resetForm,
    resetState,
    updateForm,
    updateOriginalState,
    submitFormData,
  } = useFormState();

  // 原始内容解析器
  const { rawContent, generateRawContent, parseRawContentToForm } =
    useRawContentParser();

  // 分类管理
  const {
    categoryLoading,
    categoryOptions,
    handleCategoryChange: handleCategoryChangeInternal,
    handleCategoryVisibleChange,
    ensureCategoryInOptions,
  } = useCategories();

  // API操作
  const { loadContent, submitLogrotate } = useLogrotateApi(setLoading);

  // 模式管理
  const { handleModeChange } = useModeManager(
    activeMode,
    previousMode,
    generateRawContent,
    parseRawContentToForm,
    updateForm,
    formData
  );

  // 编辑器扩展
  const { getLogrotateExtensions } = useEditorConfig(ref(null));
  const logrotateExtensions = getLogrotateExtensions();

  // 取消操作
  const handleCancel = () => {
    hideDrawer();
    resetState();
    rawContent.value = '';
  };

  const handleCategoryChange = () => {
    handleCategoryChangeInternal();
  };

  const handleCategoryVisibleChangeWrapper = (isVisible: boolean) => {
    handleCategoryVisibleChange(isVisible, currentType.value);
  };

  const handleUpdateFormData = (field: string, value: any) => {
    updateForm({ [field]: value });
  };

  // 判断原始内容是否已更改
  const isRawContentChanged = computed(
    () => rawContent.value !== originalRawContent.value
  );

  // 判断表单是否已更改
  const isFormChanged = computed(() => {
    if (activeMode.value === 'raw') {
      return isRawContentChanged.value;
    }
    return formData.category && originalFormData.value
      ? JSON.stringify(formData) !== JSON.stringify(originalFormData.value)
      : false;
  });

  // 提交表单数据
  const handleSubmit = async () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error(t('common.host_id_required'));
      return;
    }

    try {
      // 表单模式下验证表单
      if (activeMode.value === 'form') {
        await submitFormData();
      }

      // 提交到API
      const successMessage = await submitLogrotate(
        activeMode.value,
        formData,
        rawContent.value,
        isEdit.value,
        currentType.value,
        originalCategory.value,
        originalName.value,
        Number(hostId)
      );

      Message.success(successMessage);

      // 如果分类已更改，通知父组件
      if (formData.category !== originalCategory.value) {
        emit('categoryChange', formData.category);
      }

      emit('ok');
      handleCancel();
    } catch (error) {
      log('提交失败:', error);
    }
  };

  // 加载数据并显示抽屉
  const show = async (params?: ShowParams) => {
    showDrawer();
    activeMode.value = 'form';
    previousMode.value = 'form';

    if (params) {
      currentType.value = params.type || LOGROTATE_TYPE.Local;
      isEdit.value = params.isEdit || false;

      if (params.isEdit && params.record) {
        originalName.value = params.record.name;
        originalCategory.value = params.record.category;

        // 设置基本信息
        updateForm({
          name: params.record.name,
          category: params.record.category,
          path: '',
          frequency: formData.frequency,
          count: 7,
          compress: false,
          delayCompress: false,
          missingOk: false,
          notIfEmpty: false,
          create: '',
          preRotate: '',
          postRotate: '',
        });

        // 加载原始内容并解析
        try {
          const hostId = currentHostId.value;
          if (!hostId) {
            throw new Error('Host ID is required');
          }

          rawContent.value = await loadContent(
            params.record.type,
            params.record.category,
            params.record.name,
            Number(hostId)
          );

          // 将原始内容解析为表单字段
          if (rawContent.value) {
            const parsedData = parseRawContentToForm(formData);
            if (parsedData) {
              updateForm(parsedData);
            }
          }

          // 设置原始表单数据作为"原始状态"
          await updateOriginalState();
          originalRawContent.value = rawContent.value;
          log('📋 原始数据已设置', {
            name: originalName.value,
            category: originalCategory.value,
          });
        } catch (error) {
          log('加载内容失败:', error);
        }
      } else {
        resetForm();
        updateForm({ category: params.category || '' });
        // 设置原始表单数据
        await updateOriginalState();
        originalRawContent.value = rawContent.value;
      }
    } else {
      resetForm();
      await updateOriginalState();
      originalRawContent.value = rawContent.value;
    }

    // 确保当前分类在选项中
    if (formData.category) {
      ensureCategoryInOptions(formData.category);
    }
  };

  // 监听表单数据变化以同步到文件模式
  watch(
    formData,
    (newData) => {
      if (activeMode.value === 'form' && newData.path && newData.name) {
        generateRawContent(newData);
      }
    },
    { deep: true }
  );

  // 当原始内容更改时，确保更改检测正常工作
  watch(rawContent, () => {
    log('📄 原始内容已更新');
  });

  // 暴露方法
  defineExpose({
    show,
    hide: handleCancel,
  });
</script>

<style scoped>
  .form-drawer {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .form-drawer :deep(.arco-tabs) {
    display: flex;
    flex: 1;
    flex-direction: column;
    height: 100%;
  }

  .form-drawer :deep(.arco-tabs-content) {
    flex: 1;
    height: 100%;
  }

  .form-drawer :deep(.arco-tabs-pane) {
    height: 100%;
  }

  .drawer-footer {
    display: flex;
    flex-shrink: 0;
    justify-content: flex-end;
    padding-top: 16px;
    margin-top: 24px;
    border-top: 1px solid var(--color-border-2);
  }
</style>
