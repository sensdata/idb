<template>
  <div :class="typedStyles.service_form">
    <a-form
      ref="formRef"
      :model="formModel"
      layout="vertical"
      :class="typedStyles.service_form_container"
    >
      <BasicInfoSection
        ref="basicSectionRef"
        :form-model="formModel"
        :type="type"
        :initial-category="category"
        :styles="typedStyles"
        @category-change="handleCategoryChange"
        @update:form-model="updateFormModel"
      />

      <ExecutionSection
        :form-model="formModel"
        :styles="typedStyles"
        @open-environment-editor="openEnvironmentEditor"
        @update:form-model="updateFormModel"
      />

      <AdvancedSection
        :form-model="formModel"
        :styles="typedStyles"
        @update:form-model="updateFormModel"
      />
    </a-form>

    <environment-editor
      ref="environmentEditorRef"
      @update="handleEnvironmentUpdate"
    />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick } from 'vue';
  import { SERVICE_TYPE } from '@/config/enum';
  import { ServiceEntity } from '@/entity/Service';
  import { useLogger } from '@/hooks/use-logger';
  import EnvironmentEditor from './environment-editor.vue';
  import {
    BasicInfoSection,
    ExecutionSection,
    AdvancedSection,
  } from './sections';
  import { useFormModel } from './hooks/use-form-model';
  import styles from './styles/index.module.css';

  const props = defineProps<{
    type: SERVICE_TYPE;
    category: string;
    isEdit: boolean;
    record?: ServiceEntity;
  }>();

  const emit = defineEmits<{
    categoryChange: [category: string];
    change: [];
  }>();

  const formData = defineModel<any>('formData', { default: () => ({}) });
  const rawContent = defineModel<string>('rawContent', { default: '' });

  // 日志记录
  const { logError } = useLogger('ServiceForm');

  // 表单引用
  const formRef = ref();
  const basicSectionRef = ref();

  // 使用表单模型钩子
  const { formModel, setFormData, getFormData } = useFormModel(
    props.type,
    props.category,
    props.isEdit,
    props.record
  );

  // 环境变量编辑器引用
  const environmentEditorRef = ref<InstanceType<typeof EnvironmentEditor>>();

  // 更新表单模型
  const updateFormModel = (newModel: any) => {
    formModel.value = newModel;
  };

  // 处理分类变化
  const handleCategoryChange = (category: string) => {
    emit('categoryChange', category);
  };

  // 环境变量编辑器方法
  const openEnvironmentEditor = () => {
    environmentEditorRef.value?.show(formModel.value.environment);
  };

  const handleEnvironmentUpdate = (value: string) => {
    const updatedModel = { ...formModel.value };
    updatedModel.environment = value;
    updateFormModel(updatedModel);
  };

  // 验证表单
  const validate = async () => {
    try {
      await formRef.value?.validate();
      return true;
    } catch {
      return false;
    }
  };

  // 监听外部数据变化
  watch(
    () => formData.value,
    (newValue) => {
      if (newValue) {
        // 使用Vue.nextTick确保DOM更新
        nextTick(() => {
          setFormData(newValue);
          // 确保分类在选项中
          if (formModel.value.category) {
            basicSectionRef.value?.ensureCategoryInOptions?.(
              formModel.value.category
            );
          }
        });
      }
    },
    { immediate: true, deep: true }
  );

  // 监听表单模型变化
  watch(
    () => formModel.value,
    () => {
      emit('change');
    },
    { deep: true }
  );

  // 获取表单数据处理
  const getFormDataHandler = () => {
    return getFormData(
      rawContent.value || formData.value?.parsedConfig?.rawContent || ''
    );
  };

  // 重置表单数据（在文件模式到表单模式切换时调用）
  const resetForm = (newFormData: any) => {
    if (!newFormData || !newFormData.parsedConfig) {
      logError('resetForm: 传入的表单数据不完整', newFormData);
      return;
    }

    try {
      // 确保parsedConfig存在且包含必要字段
      const config = newFormData.parsedConfig;

      // 创建一个全新的表单模型对象以触发视图更新
      const updatedFormModel = {
        name: newFormData.name || '',
        category: newFormData.category || props.category || '',
        description: config.description || '',
        serviceType: config.serviceType || 'simple',
        execStart: config.execStart || '',
        workingDirectory: config.workingDirectory || '',
        user: config.user || 'root',
        group: config.group || 'root',
        environment: config.environment || '',
        execStop: config.execStop || '',
        execReload: config.execReload || '',
        restart: config.restart || 'no',
        restartSec: Number(config.restartSec) || 0,
        timeoutStartSec: Number(config.timeoutStartSec) || 90,
        timeoutStopSec: Number(config.timeoutStopSec) || 90,
        rawContent: config.rawContent || newFormData.originalContent || '',
      };

      // 直接更新formModel以触发所有子组件刷新
      formModel.value = { ...updatedFormModel };

      // 手动强制所有子组件刷新
      nextTick(() => {
        // 确保子组件接收到更新
        emit('change');

        // 确保分类选项包含当前分类
        if (formModel.value.category && basicSectionRef.value) {
          basicSectionRef.value.ensureCategoryInOptions?.(
            formModel.value.category
          );
        }
      });
    } catch (error) {
      logError('表单重置失败:', error as Error);
    }
  };

  // 获取当前表单模型
  const getFormModel = () => {
    return { ...formModel.value };
  };

  // 刷新分类选项
  const refreshCategories = async (category?: string) => {
    if (basicSectionRef.value?.refreshCategoriesAndEnsure) {
      await basicSectionRef.value.refreshCategoriesAndEnsure(category);
    }
  };

  // CSS模块类型处理（解决类型错误）
  interface ServiceFormStyles {
    service_form: string;
    service_form_container: string;
    form_section: string;
    section_title: string;
    form_item_with_help: string;
    command_form_item: string;
    form_help: string;
    form_row: string;
    form_item_half: string;
    [key: string]: string;
  }

  const typedStyles = styles as unknown as ServiceFormStyles;

  defineExpose({
    validate,
    getFormData: getFormDataHandler,
    setFormData,
    resetForm,
    getFormModel,
    refreshCategories,
    formRef,
  });
</script>

<style scoped>
  /* 使用深度选择器覆盖Arco Design组件样式 */
  .service_form
    :deep(.service-type-group .arco-radio-group-button .arco-radio) {
    background-color: #fff;
  }

  .service_form
    :deep(
      .service-type-group
        .arco-radio-group-button
        .arco-radio
        .arco-radio-button
    ) {
    background-color: #fff;
    border-color: var(--color-border-2);
  }

  .service_form
    :deep(.service-type-group .arco-radio-group-button .arco-radio-checked) {
    background-color: var(--color-primary-light-1);
  }

  .service_form
    :deep(
      .service-type-group
        .arco-radio-group-button
        .arco-radio-checked
        .arco-radio-button
    ) {
    background-color: var(--color-primary-light-1);
    border-color: var(--color-primary-light-3);
  }
</style>
