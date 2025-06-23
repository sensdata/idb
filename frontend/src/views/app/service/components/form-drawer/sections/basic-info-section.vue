<template>
  <div :class="styles.form_section">
    <h3 :class="styles.section_title">{{
      $t('app.service.form.section.basic')
    }}</h3>

    <div :class="styles.form_row">
      <a-form-item
        field="name"
        :label="$t('app.service.form.field.name')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.name.required'),
          },
        ]"
        :class="styles.form_item_half"
      >
        <a-input
          :model-value="formModel.name"
          :placeholder="$t('app.service.form.field.name')"
          allow-clear
          @update:model-value="updateField('name', $event)"
        >
          <template #prefix>
            <icon-tag />
          </template>
        </a-input>
      </a-form-item>

      <a-form-item
        field="category"
        :label="$t('app.service.form.field.category')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.category.required'),
          },
        ]"
        :class="styles.form_item_half"
      >
        <a-select
          :model-value="formModel.category"
          :placeholder="$t('app.service.form.field.category.placeholder')"
          :loading="categoryLoading"
          :options="categoryOptions"
          :disabled="isEdit"
          :allow-clear="!isEdit"
          :allow-create="!isEdit"
          @change="handleCategoryChange"
          @update:model-value="updateField('category', $event)"
          @visible-change="handleCategoryVisibleChange"
        >
          <template #prefix>
            <icon-folder />
          </template>
        </a-select>
      </a-form-item>
    </div>

    <a-form-item
      field="description"
      :label="$t('app.service.form.field.description')"
      :rules="[
        {
          required: true,
          message: $t('app.service.form.validate.description.required'),
        },
      ]"
    >
      <a-textarea
        :model-value="formModel.description"
        :placeholder="$t('app.service.form.field.description')"
        :auto-size="{ minRows: 2, maxRows: 3 }"
        @update:model-value="updateField('description', $event)"
      />
    </a-form-item>

    <a-form-item
      field="serviceType"
      :label="$t('app.service.form.field.service_type')"
      :rules="[
        {
          required: true,
          message: $t('app.service.form.validate.service_type.required'),
        },
      ]"
    >
      <div :class="styles.service_type_cards">
        <div
          v-for="serviceType in serviceTypes"
          :key="serviceType.value"
          :class="[
            styles.service_type_card,
            {
              [styles.service_type_card_active]:
                formModel.serviceType === serviceType.value,
            },
          ]"
          @click="updateField('serviceType', serviceType.value)"
        >
          <div :class="styles.service_type_card_content">
            <span :class="styles.service_type_card_title">{{
              serviceType.label
            }}</span>
            <div :class="styles.service_type_card_desc">{{
              serviceType.description
            }}</div>
          </div>
        </div>
      </div>
    </a-form-item>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { getServiceCategoryListApi } from '@/api/service';
  import { useLogger } from '@/composables/use-logger';
  import type { SERVICE_TYPE } from '@/config/enum';

  const props = defineProps<{
    type: SERVICE_TYPE;
    initialCategory: string;
    isEdit: boolean;
    styles: any;
  }>();

  // 使用defineModel实现表单数据的双向绑定
  const formModel = defineModel<any>('formModel');

  const emit = defineEmits<{
    categoryChange: [category: string];
  }>();

  const { t } = useI18n();
  const { log, logError } = useLogger('BasicInfoSection');

  // 服务类型选项
  const serviceTypes = computed(() =>
    [
      {
        value: 'simple',
        label: 'Simple',
        description: 'app.service.form.service_type.simple',
      },
      {
        value: 'forking',
        label: 'Forking',
        description: 'app.service.form.service_type.forking',
      },
      {
        value: 'oneshot',
        label: 'Oneshot',
        description: 'app.service.form.service_type.oneshot',
      },
      {
        value: 'notify',
        label: 'Notify',
        description: 'app.service.form.service_type.notify',
      },
    ].map((type) => ({
      ...type,
      description: t(type.description),
    }))
  );

  // 分类相关状态
  const categoryLoading = ref(false);
  const categoryOptions = ref<Array<{ label: string; value: string }>>([]);

  // 加载分类列表
  const fetchCategories = async () => {
    categoryLoading.value = true;
    try {
      const response = await getServiceCategoryListApi({
        type: props.type,
        page: 1,
        page_size: 100,
      });
      categoryOptions.value = response.items.map((category: any) => ({
        label: category.name,
        value: category.name,
      }));
      log('fetchCategories: 加载分类列表', categoryOptions.value);
    } catch (error) {
      logError('fetchCategories: 加载分类失败', error);
      categoryOptions.value = [];
    } finally {
      categoryLoading.value = false;
    }
  };

  // 处理分类变化
  const handleCategoryChange = (
    value:
      | string
      | number
      | boolean
      | Record<string, any>
      | (string | number | boolean | Record<string, any>)[]
  ) => {
    const categoryValue = String(value);
    emit('categoryChange', categoryValue);
  };

  // 处理分类下拉框显示状态变化
  const handleCategoryVisibleChange = (visible: boolean) => {
    if (visible) {
      fetchCategories();
    }
  };

  // 确保分类在选项中存在
  const ensureCategoryInOptions = (category: string) => {
    if (!category) return;

    // 如果分类已存在于选项中则直接返回
    if (categoryOptions.value.some((option) => option.value === category)) {
      return;
    }

    // 如果分类不存在，则添加到选项列表
    categoryOptions.value.unshift({
      label: category,
      value: category,
    });
  };

  // 更新表单字段
  const updateField = (field: string, value: any) => {
    formModel.value = {
      ...formModel.value,
      [field]: value,
    };
  };

  // 在组件挂载时确保初始分类在选项中
  onMounted(async () => {
    // 先加载分类列表
    await fetchCategories();

    if (formModel.value.category) {
      ensureCategoryInOptions(formModel.value.category);
    }
  });

  // 监听分类prop变化
  watch(
    () => props.initialCategory,
    (newCategory) => {
      if (newCategory && newCategory !== formModel.value.category) {
        updateField('category', newCategory);
        ensureCategoryInOptions(newCategory);
      }
    },
    { immediate: true }
  );

  // 刷新分类选项并确保特定分类在选项中
  const refreshCategoriesAndEnsure = async (category?: string) => {
    await fetchCategories();
    if (category) {
      ensureCategoryInOptions(category);
    }
  };

  defineExpose({
    fetchCategories,
    ensureCategoryInOptions,
    refreshCategoriesAndEnsure,
  });
</script>
