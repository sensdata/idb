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
            :is-view="params.isView"
            :record="params.record"
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
            :is-view="params.isView"
            :record="params.record"
            @change="handleRawChange"
          />
        </a-tab-pane>
      </a-tabs>
    </div>
    <template #footer>
      <div class="drawer-footer">
        <a-button @click="handleCancel">{{ $t('common.cancel') }}</a-button>
        <a-button
          v-if="!params.isView"
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
    createServiceFormApi,
    updateServiceFormApi,
    getServiceDetailApi,
    DEFAULT_SERVICE_CATEGORY,
  } from '@/api/service';
  import { useForm } from '@/composables/use-form';
  import { useLogger } from '@/composables/use-logger';
  import { useServiceFormState } from './composables/use-service-form-state';
  import { useServiceModeSync } from './composables/use-service-mode-sync';
  import { useServiceParser } from './composables/use-service-parser';
  import ServiceForm from './service-form.vue';
  import ServiceRaw from './service-raw.vue';
  import { formatApiErrorMessage } from './utils';

  interface DrawerParams {
    type: SERVICE_TYPE;
    category?: string;
    name?: string;
    isEdit?: boolean;
    isView?: boolean;
    record?: ServiceEntity;
  }

  const emit = defineEmits<{
    ok: [];
  }>();

  const { t } = useI18n();
  const { logDebug, logError } = useLogger('ServiceFormDrawer');

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

  const serviceFormRef = ref<InstanceType<typeof ServiceForm>>();
  const serviceRawRef = ref<InstanceType<typeof ServiceRaw>>();

  const { parseServiceConfig, parseServiceConfigStructured } =
    useServiceParser();

  const loading = ref(false);

  const normalizeCategory = (type: SERVICE_TYPE, category?: string) => {
    if (type === SERVICE_TYPE.System) return '';
    return category && category.trim() ? category : DEFAULT_SERVICE_CATEGORY;
  };

  const loadServiceData = async () => {
    if (!params.value.record) return;

    try {
      loading.value = true;

      const category = normalizeCategory(
        params.value.type,
        params.value.category
      );
      const serviceName = params.value.record.name;

      if (!serviceName) {
        return;
      }

      const latestContent = await getServiceDetailApi({
        type: params.value.type,
        category,
        name: serviceName,
      });

      rawContent.value = latestContent;
      originalRawContent.value = latestContent;

      if (latestContent) {
        try {
          const parsedConfig = parseServiceConfig(latestContent);
          const structuredConfig = parseServiceConfigStructured(latestContent);

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

  const { submitForm, setFormRef } = useForm({
    initialValues: formData.value,
    async onSubmit() {
      if (params.value.isView) return;

      loading.value = true;
      try {
        if (activeTab.value === 'form') {
          const currentFormData = await serviceFormRef.value?.getFormData();
          if (!currentFormData) return;

          const formModel = serviceFormRef.value?.getFormModel();
          if (!formModel) return;

          const normalizedCategory = normalizeCategory(
            params.value.type,
            currentFormData.category
          );

          const formFields = [
            { key: 'Description', value: formModel.description || '' },
            { key: 'Type', value: formModel.serviceType || 'simple' },
            { key: 'ExecStart', value: formModel.execStart || '' },
            {
              key: 'WorkingDirectory',
              value: formModel.workingDirectory || '',
            },
            { key: 'User', value: formModel.user || 'root' },
            { key: 'Group', value: formModel.group || 'root' },
            { key: 'Environment', value: formModel.environment || '' },
          ];

          if (params.value.isEdit) {
            await updateServiceFormApi({
              type: currentFormData.type,
              category: normalizedCategory,
              name: params.value.record?.name || currentFormData.name,
              new_name: currentFormData.name,
              new_category: normalizedCategory,
              form: formFields,
            });
          } else {
            await createServiceFormApi({
              type: currentFormData.type,
              category: normalizedCategory,
              name: currentFormData.name,
              form: formFields,
            });
          }
        } else {
          const isValid = await serviceRawRef.value?.validate();
          if (!isValid) return;

          const submitData = serviceRawRef.value?.getSubmitData();
          if (!submitData) return;

          const normalizedCategory = normalizeCategory(
            params.value.type,
            submitData.category
          );

          if (params.value.isEdit) {
            await updateServiceRawApi({
              ...submitData,
              category: normalizedCategory,
              new_name: submitData.name,
            });
          } else {
            await createServiceRawApi({
              ...submitData,
              category: normalizedCategory,
            });
          }
        }

        Message.success(
          params.value.isEdit
            ? t('app.service.form.success.update')
            : t('app.service.form.success.create')
        );

        hasChanges.value = false;

        if (params.value.isEdit) {
          await loadServiceData();
          await nextTick();
          if (activeTab.value === 'form' && serviceFormRef.value) {
            serviceFormRef.value.setFormData(formData.value);
          }
          emit('ok');
        } else {
          resetState();
          emit('ok');
        }
      } catch (error: any) {
        logError('保存失败', error as Error);

        const defaultMessage = params.value.isEdit
          ? t('app.service.form.error.update')
          : t('app.service.form.error.create');
        const errorMessage = formatApiErrorMessage(error, defaultMessage);

        Message.error(errorMessage);
      } finally {
        loading.value = false;
      }
    },
  });

  const { syncFormToRaw, syncRawToForm, checkFormChanges } = useServiceModeSync(
    formData,
    rawContent,
    activeTab,
    originalRawContent,
    serviceFormRef,
    serviceRawRef,
    setFormChanged
  );

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
    let action = t('app.service.form.title.create');
    if (params.value.isView) {
      action = t('common.view');
    } else if (params.value.isEdit) {
      action = t('app.service.form.title.edit');
    }
    return `${action} - ${t(`app.service.enum.type.${params.value.type}`)}`;
  });

  const lastActiveTab = ref('form');

  const handleTabChange = async (key: string | number) => {
    const newMode = String(key);
    const previousMode = lastActiveTab.value;

    logDebug(`模式切换: ${previousMode} -> ${newMode}`);

    if (newMode === previousMode) {
      return;
    }

    try {
      if (previousMode === 'form' && newMode === 'raw') {
        await syncFormToRaw();
      } else if (previousMode === 'raw' && newMode === 'form') {
        await syncRawToForm();
      }

      lastActiveTab.value = newMode;

      if (newMode === 'form') {
        await nextTick();
      }
    } catch (error) {
      logError('模式切换失败', error as Error);
      Message.error(t('app.service.form.error.mode_switch'));
      activeTab.value = previousMode;
    }
  };

  const handleFormChange = async () => {
    if (activeTab.value !== 'form') return;
    await checkFormChanges(originalRawContent.value);
  };

  const handleRawChange = async () => {
    if (activeTab.value !== 'raw') return;
    await checkFormChanges(originalRawContent.value);
  };

  const resetDrawerState = () => {
    resetState();
    activeTab.value = 'form';
    lastActiveTab.value = 'form';
    loading.value = false;
  };

  const handleCancel = () => {
    resetDrawerState();
  };

  const handleSubmit = async () => {
    await submitForm();
  };

  const show = async (
    showParams: DrawerParams = { type: SERVICE_TYPE.Local }
  ) => {
    resetDrawerState();

    params.value = {
      type: showParams.type || SERVICE_TYPE.Local,
      category: normalizeCategory(
        showParams.type || SERVICE_TYPE.Local,
        showParams.category
      ),
      isEdit: showParams.isEdit || false,
      isView: showParams.isView || false,
      record: showParams.record,
    };

    visible.value = true;

    if (showParams.isEdit && showParams.record) {
      await loadServiceData();
    } else {
      formData.value = {
        name: '',
        category: normalizeCategory(params.value.type, params.value.category),
        parsedConfig: parseServiceConfig(''),
        structuredConfig: parseServiceConfigStructured(''),
      };
      rawContent.value = '';
      originalRawContent.value = '';
    }

    hasChanges.value = false;
  };

  defineExpose({
    show,
  });
</script>

<style scoped>
  .form-drawer-content {
    height: calc(100vh - 180px);
    overflow-y: auto;
  }

  .drawer-footer {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  :deep(.arco-tabs-content) {
    padding-top: 16px;
  }
</style>
