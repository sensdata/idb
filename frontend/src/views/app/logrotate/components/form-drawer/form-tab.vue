<template>
  <a-form ref="formRef" :model="formData" :rules="formRules" layout="vertical">
    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item
          field="name"
          :label="$t('app.logrotate.form.name')"
          required
        >
          <a-input
            :model-value="formData.name"
            :placeholder="$t('app.logrotate.form.name_placeholder')"
            :disabled="isEdit"
            @update:model-value="(value) => updateFormData('name', value)"
          />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item
          field="category"
          :label="$t('app.logrotate.form.category')"
          required
        >
          <a-select
            :model-value="formData.category"
            :placeholder="$t('app.logrotate.form.category_placeholder')"
            :loading="categoryLoading"
            :options="categoryOptions"
            allow-clear
            allow-create
            @change="handleCategoryChange"
            @visible-change="handleCategoryVisibleChange"
            @update:model-value="(value) => updateFormData('category', value)"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-form-item field="path" :label="$t('app.logrotate.form.path')" required>
      <a-input
        :model-value="formData.path"
        :placeholder="$t('app.logrotate.form.path_placeholder')"
        @update:model-value="(value) => updateFormData('path', value)"
      />
    </a-form-item>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item
          field="frequency"
          :label="$t('app.logrotate.form.frequency')"
          required
        >
          <a-select
            :model-value="formData.frequency"
            :placeholder="$t('app.logrotate.form.frequency_placeholder')"
            @update:model-value="(value) => updateFormData('frequency', value)"
          >
            <a-option
              v-for="freq in frequencyOptions"
              :key="freq.value"
              :value="freq.value"
            >
              {{ freq.label }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item
          field="count"
          :label="$t('app.logrotate.form.count')"
          required
        >
          <a-input-number
            :model-value="formData.count"
            :placeholder="$t('app.logrotate.form.count_placeholder')"
            :min="1"
            :precision="0"
            style="width: 100%"
            @update:model-value="(value) => updateFormData('count', value)"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-form-item field="create" :label="$t('app.logrotate.form.create')">
      <PermissionInput
        :model-value="formData.create"
        @update:model-value="(value) => updateFormData('create', value)"
      />
    </a-form-item>

    <a-row :gutter="16">
      <a-col :span="8">
        <a-form-item field="compress">
          <a-checkbox
            :model-value="formData.compress"
            @update:model-value="(value) => updateFormData('compress', value)"
          >
            {{ $t('app.logrotate.form.compress') }}
          </a-checkbox>
        </a-form-item>
      </a-col>
      <a-col :span="8">
        <a-form-item field="delayCompress">
          <a-checkbox
            :model-value="formData.delayCompress"
            @update:model-value="
              (value) => updateFormData('delayCompress', value)
            "
          >
            {{ $t('app.logrotate.form.delay_compress') }}
          </a-checkbox>
        </a-form-item>
      </a-col>
      <a-col :span="8">
        <a-form-item field="missingOk">
          <a-checkbox
            :model-value="formData.missingOk"
            @update:model-value="(value) => updateFormData('missingOk', value)"
          >
            {{ $t('app.logrotate.form.missing_ok') }}
          </a-checkbox>
        </a-form-item>
      </a-col>
    </a-row>

    <a-form-item
      field="notIfEmpty"
      class="checkbox-group-item"
      style="margin-top: -8px; margin-bottom: 12px"
    >
      <a-checkbox
        :model-value="formData.notIfEmpty"
        @update:model-value="(value) => updateFormData('notIfEmpty', value)"
      >
        {{ $t('app.logrotate.form.not_if_empty') }}
      </a-checkbox>
    </a-form-item>

    <a-form-item field="preRotate" :label="$t('app.logrotate.form.pre_rotate')">
      <a-textarea
        :model-value="formData.preRotate"
        :placeholder="$t('app.logrotate.form.pre_rotate_placeholder')"
        :rows="3"
        @update:model-value="(value) => updateFormData('preRotate', value)"
      />
    </a-form-item>

    <a-form-item
      field="postRotate"
      :label="$t('app.logrotate.form.post_rotate')"
    >
      <a-textarea
        :model-value="formData.postRotate"
        :placeholder="$t('app.logrotate.form.post_rotate_placeholder')"
        :rows="3"
        @update:model-value="(value) => updateFormData('postRotate', value)"
      />
    </a-form-item>
  </a-form>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import type { FormData, CategoryValue, SelectOption } from './types';
  import PermissionInput from './permission-input.vue';

  interface Props {
    formData: FormData;
    formRules: Record<string, any>;
    frequencyOptions: SelectOption[];
    categoryLoading: boolean;
    categoryOptions: SelectOption[];
    isEdit: boolean;
  }

  const emit = defineEmits<{
    categoryChange: [category: CategoryValue];
    categoryVisibleChange: [visible: boolean];
    updateFormData: [field: keyof FormData, value: any];
  }>();

  defineProps<Props>();

  const formRef = ref();

  const updateFormData = (field: keyof FormData, value: any) => {
    emit('updateFormData', field, value);
  };

  const handleCategoryChange = (category: CategoryValue) => {
    emit('categoryChange', category);
    updateFormData('category', category);
  };

  const handleCategoryVisibleChange = (visible: boolean) => {
    emit('categoryVisibleChange', visible);
  };

  defineExpose({
    validate: () => formRef.value?.validate(),
    clearValidate: () => formRef.value?.clearValidate(),
    resetFields: () => formRef.value?.resetFields(),
  });
</script>

<style scoped lang="less">
  /* 轻量样式调整，不干扰默认功能 */
  :deep(.arco-checkbox) {
    margin-bottom: 0.286rem;
  }

  :deep(.arco-checkbox-label) {
    font-size: 1rem;
  }

  /* 减少文件权限表单项的下边距 */
  :deep(.arco-form-item.arco-form-item-layout-vertical) {
    margin-bottom: 1.143rem; /* 16px / 14px */
  }

  /* 特别减少文件权限表单项的下边距 */
  :deep(.arco-form-item.arco-form-item-layout-vertical:nth-of-type(4)) {
    margin-bottom: 0.357rem; /* 5px / 14px */
  }

  /* 更具体的选择器，确保样式生效 */
  :deep(.checkbox-group-item) {
    margin-top: -1.5rem !important;
    margin-bottom: 0.8rem !important;
  }

  /* 选中状态颜色调整 - 对于第三方组件库需要使用!important来确保样式生效 */
  :deep(.arco-checkbox-checked .arco-checkbox-icon) {
    background-color: var(--idblue-6) !important;
    border-color: var(--idblue-6) !important;
  }

  :deep(.arco-checkbox-checked .arco-checkbox-icon .arco-checkbox-icon-check) {
    color: var(--idb-brand-text) !important;
  }

  /* 悬停状态 */
  :deep(.arco-checkbox:not(.arco-checkbox-disabled):hover .arco-checkbox-icon) {
    border-color: var(--idblue-6) !important;
  }

  /* 聚焦状态 */
  :deep(
      .arco-checkbox:not(.arco-checkbox-disabled).arco-checkbox-focus
        .arco-checkbox-icon
    ) {
    border-color: var(--idblue-6) !important;
    box-shadow: 0 0 0 2px var(--idblue-1) !important;
  }
</style>
