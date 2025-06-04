<template>
  <a-drawer
    v-model:visible="visible"
    :title="drawerTitle"
    :width="800"
    unmount-on-close
    @cancel="handleCancel"
  >
    <div class="form-drawer-content">
      <a-tabs v-model:active-key="activeTab" @change="handleTabChange">
        <a-tab-pane key="form" :title="$t('app.service.form.tab.form')">
          <service-form
            ref="serviceFormRef"
            v-model:form-data="formData"
            v-model:raw-content="rawContent"
            :type="params.type"
            :category="params.category"
            :is-edit="params.isEdit"
            :record="params.record"
            @category-change="handleCategoryChange"
            @change="handleFormChange"
          />
        </a-tab-pane>
        <a-tab-pane key="raw" :title="$t('app.service.form.tab.raw')">
          <service-raw
            ref="serviceRawRef"
            v-model:content="rawContent"
            :type="params.type"
            :category="params.category"
            :is-edit="params.isEdit"
            :record="params.record"
            @category-change="handleCategoryChange"
            @change="handleRawChange"
          />
        </a-tab-pane>
      </a-tabs>
    </div>
    <template #footer>
      <div class="drawer-footer">
        <a-button @click="handleCancel">{{ $t('common.cancel') }}</a-button>
        <a-button
          type="primary"
          :loading="loading"
          :disabled="!hasChanges"
          @click="handleSubmit"
        >
          {{ $t('common.save') }}
        </a-button>
      </div>
    </template>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, computed, nextTick, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { SERVICE_TYPE } from '@/config/enum';
  import { ServiceEntity } from '@/entity/Service';
  import {
    createServiceRawApi,
    updateServiceRawApi,
    getServiceDetailApi,
    createServiceCategoryApi,
    getServiceCategoryListApi,
  } from '@/api/service';
  import { useForm } from '@/hooks/use-form';
  import { useLogger } from '@/hooks/use-logger';
  import { useServiceFormState } from './hooks/use-service-form-state';
  import { useServiceModeSync } from './hooks/use-service-mode-sync';
  import { useServiceParser } from './hooks/use-service-parser';
  import ServiceForm from './service-form.vue';
  import ServiceRaw from './service-raw.vue';

  interface DrawerParams {
    type: SERVICE_TYPE;
    category: string;
    name?: string;
    isEdit: boolean;
    record?: ServiceEntity;
  }

  const emit = defineEmits<{
    ok: [];
    categoryChange: [category: string];
    categoryCreated: [category: string];
  }>();

  const { t } = useI18n();
  const { logDebug, logError } = useLogger('ServiceFormDrawer');

  // 使用状态管理hook
  const {
    visible,
    activeTab,
    params,
    formData,
    rawContent,
    hasChanges,
    originalRawContent,
    resetState,
    setFormChanged,
  } = useServiceFormState();

  // 组件引用
  const serviceFormRef = ref<InstanceType<typeof ServiceForm>>();
  const serviceRawRef = ref<InstanceType<typeof ServiceRaw>>();

  // 初始化 service parser
  const { parseServiceConfig, parseServiceConfigStructured } =
    useServiceParser();

  // 创建独立的loading状态
  const loading = ref(false);

  // 加载服务数据
  const loadServiceData = async () => {
    if (!params.value.record) return;

    try {
      loading.value = true;

      const category = params.value.category;
      const serviceName = params.value.record.name;

      if (!category || !serviceName) {
        Message.error(t('app.service.form.error.missing_category'));
        return;
      }

      // 从API获取最新的服务内容
      const latestContent = await getServiceDetailApi({
        type: params.value.type,
        category,
        name: serviceName,
      });

      // 设置原始内容为最新获取的内容
      rawContent.value = latestContent;
      originalRawContent.value = latestContent;

      // 从最新的原始配置解析出表单数据
      if (latestContent) {
        try {
          const parsedConfig = parseServiceConfig(latestContent);
          const structuredConfig = parseServiceConfigStructured(latestContent);

          // 传递解析后的数据给表单组件
          const initialFormData = {
            name: serviceName,
            category,
            parsedConfig,
            structuredConfig,
            originalContent: latestContent,
          };

          formData.value = initialFormData;
        } catch (error) {
          logError('解析服务配置失败', error as Error);
          Message.error(t('app.service.form.error.parse'));
        }
      }
    } catch (error) {
      logError('加载服务数据失败', error as Error);
      Message.error(t('app.service.form.error.load'));
    } finally {
      loading.value = false;
    }
  };

  // 使用通用表单hook处理提交
  const { submitForm, setFormRef } = useForm({
    initialValues: formData.value,
    async onSubmit() {
      // 设置loading状态
      loading.value = true;
      try {
        if (activeTab.value === 'form') {
          // 表单模式提交
          const currentFormData = await serviceFormRef.value?.getFormData();
          if (!currentFormData) return;

          if (params.value.isEdit) {
            await updateServiceRawApi({
              type: currentFormData.type,
              category: currentFormData.category,
              name: params.value.record?.name || currentFormData.name,
              new_name: currentFormData.name,
              content: currentFormData.content,
            });
          } else {
            await createServiceRawApi({
              type: currentFormData.type,
              category: currentFormData.category,
              name: currentFormData.name,
              content: currentFormData.content,
            });
          }
        } else {
          // 原始模式提交
          const isValid = await serviceRawRef.value?.validate();
          if (!isValid) return;

          const submitData = serviceRawRef.value?.getSubmitData();
          if (!submitData) return;

          if (params.value.isEdit) {
            await updateServiceRawApi({
              ...submitData,
              new_name: submitData.name,
            });
          } else {
            await createServiceRawApi(submitData);
          }
        }

        Message.success(
          params.value.isEdit
            ? t('app.service.form.success.update')
            : t('app.service.form.success.create')
        );

        hasChanges.value = false;

        // 如果是编辑模式，保存后不关闭弹框，只重置变更状态
        if (params.value.isEdit) {
          // 编辑模式：重新加载数据但保持弹框打开
          await loadServiceData();
          // 等待数据更新完成，然后强制表单组件刷新
          await nextTick();
          if (activeTab.value === 'form' && serviceFormRef.value) {
            // 对于表单模式，使用 setFormData 更新数据
            serviceFormRef.value.setFormData(formData.value);
          }
          // 对于原始模式，rawContent 已经在 loadServiceData 中更新了
          emit('ok'); // 通知父组件刷新列表，但不传递关闭信号
        } else {
          // 新建模式：保存后关闭弹框
          resetState();
          emit('ok');
        }
      } catch (error) {
        logError('保存失败', error as Error);
        Message.error(
          params.value.isEdit
            ? t('app.service.form.error.update')
            : t('app.service.form.error.create')
        );
      } finally {
        loading.value = false;
      }
    },
  });

  // 使用模式同步hook
  const { syncFormToRaw, syncRawToForm, checkFormChanges } = useServiceModeSync(
    formData,
    rawContent,
    activeTab,
    originalRawContent,
    serviceFormRef,
    serviceRawRef,
    setFormChanged
  );

  // 监听 serviceFormRef 变化，设置表单引用
  watch(
    () => serviceFormRef.value,
    (newRef: InstanceType<typeof ServiceForm> | undefined) => {
      if (newRef && newRef.formRef) {
        setFormRef(newRef.formRef);
      }
    },
    { immediate: true }
  );

  const drawerTitle = computed(() => {
    const action = params.value.isEdit
      ? t('app.service.form.title.edit')
      : t('app.service.form.title.create');
    return `${action} - ${t(`app.service.enum.type.${params.value.type}`)}`;
  });

  // 保存上一个标签页的状态
  const lastActiveTab = ref('form');

  // 监听标签页切换
  const handleTabChange = async (key: string | number) => {
    const newMode = String(key);
    const previousMode = lastActiveTab.value;

    logDebug(`模式切换: ${previousMode} -> ${newMode}`);

    // 如果没有真正的模式切换，直接返回
    if (previousMode === newMode) {
      return;
    }

    try {
      // 先显示加载状态
      loading.value = true;

      // 更新上一个标签页状态
      lastActiveTab.value = newMode;

      // 模式切换同步逻辑
      if (newMode === 'raw' && previousMode === 'form') {
        // 从表单模式切换到文件模式
        const success = await syncFormToRaw();
        if (!success) {
          throw new Error('从表单模式同步到文件模式失败');
        }
      } else if (newMode === 'form' && previousMode === 'raw') {
        // 从文件模式切换到表单模式
        logDebug('从文件模式切换到表单模式');

        // 1. 先同步原始内容到表单数据对象
        const success = await syncRawToForm();
        if (!success) {
          throw new Error('从文件模式同步到表单模式失败');
        }

        // 2. 确保等待数据同步完成
        await nextTick();

        // 3. 强制表单组件刷新数据
        if (serviceFormRef.value) {
          // 通过resetForm方法直接更新表单组件的内部状态
          serviceFormRef.value.resetForm(formData.value);

          // 4. 再次检查表单组件是否已更新
          await nextTick();
        }
      }
    } catch (error) {
      logError('模式切换失败', error as Error);
      Message.error(t('app.service.form.error.mode_switch'));
      // 如果切换失败，回退到原模式
      lastActiveTab.value = previousMode;
    } finally {
      // 隐藏加载状态
      loading.value = false;

      // 确保在切换后重新设置表单引用
      await nextTick();
      if (newMode === 'form' && serviceFormRef.value?.formRef) {
        setFormRef(serviceFormRef.value.formRef);
      }
    }
  };

  // 取消
  const handleCancel = () => {
    resetState();
  };

  // 显示抽屉
  const show = async (newParams?: Partial<DrawerParams>) => {
    if (newParams) {
      params.value = {
        ...params.value,
        ...newParams,
      };
    }

    visible.value = true;
    activeTab.value = 'form';
    lastActiveTab.value = 'form';
    hasChanges.value = false;

    await nextTick();

    // 如果是编辑模式，加载服务数据
    if (params.value.isEdit && params.value.record) {
      await loadServiceData();
    } else {
      // 新建模式，设置默认的表单数据
      const initialFormData = {
        name: '',
        category: params.value.category,
        parsedConfig: {},
      };

      formData.value = initialFormData;
      rawContent.value = '';
      originalRawContent.value = '';
    }

    // 确保在显示抽屉后设置表单引用
    await nextTick();
    if (activeTab.value === 'form' && serviceFormRef.value?.formRef) {
      setFormRef(serviceFormRef.value.formRef);
      // 刷新分类选项以确保显示最新的分类列表
      serviceFormRef.value.refreshCategories?.();
    }
  };

  // 确保分类存在（如果不存在则创建）
  const ensureCategoryExists = async (category: string): Promise<boolean> => {
    if (!category.trim()) return true;

    try {
      // 获取当前分类列表，检查分类是否已存在
      const response = await getServiceCategoryListApi({
        type: params.value.type,
        page: 1,
        page_size: 1000,
      });

      const existingCategories = response.items.map((item: any) => item.name);

      // 如果分类已存在，直接返回true
      if (existingCategories.includes(category)) {
        return true;
      }

      // 分类不存在，创建新分类
      await createServiceCategoryApi({
        type: params.value.type,
        category,
      });

      logDebug(`成功创建新分类: ${category}`);

      // 通知父组件刷新分类树
      emit('categoryCreated', category);

      // 刷新当前表单的分类选项列表
      if (
        activeTab.value === 'form' &&
        serviceFormRef.value?.refreshCategories
      ) {
        await serviceFormRef.value.refreshCategories(category);
      }

      return true;
    } catch (error) {
      logError('创建分类失败', error as Error);
      Message.error(t('app.service.form.error.create_category'));
      return false;
    }
  };

  // 处理分类变化
  const handleCategoryChange = async (category: string) => {
    // 更新参数中的分类
    params.value.category = category;

    // 如果分类不为空，确保分类存在
    if (category.trim()) {
      const categoryExists = await ensureCategoryExists(category);
      if (!categoryExists) {
        // 分类创建失败，不继续执行
        return;
      }
    }

    // 向父组件发射分类变化事件
    emit('categoryChange', category);
  };

  // 处理表单变化
  const handleFormChange = async () => {
    // 表单模式下检查变更并同步
    if (activeTab.value === 'form') {
      await syncFormToRaw();
    }
  };

  // 处理原始编辑器内容变化
  const handleRawChange = async () => {
    // 原始模式下直接标记为已变更
    if (activeTab.value === 'raw') {
      // 检查变更状态
      await checkFormChanges(originalRawContent.value);
    }
  };

  // 提交
  const handleSubmit = async () => {
    try {
      if (activeTab.value === 'form') {
        // 表单模式提交 - 先验证
        const isValid = await serviceFormRef.value?.validate();
        if (!isValid) return;
      }

      // 使用submitForm提交表单
      await submitForm();
    } catch (error) {
      // 错误已在submitForm中处理
    }
  };

  // 暴露方法给父组件
  defineExpose({
    show,
  });
</script>

<style scoped>
  .form-drawer-content {
    display: flex;
    flex-direction: column;
    min-height: calc(100vh - 120px);
  }

  .drawer-footer {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
</style>
