<template>
  <a-drawer
    v-model:visible="visible"
    :width="DRAWER_WIDTH"
    :footer="false"
    unmount-on-close
  >
    <template #title>
      <a-space size="small">
        <span>{{ drawerTitle }}</span>
        <a-tag v-if="isSystemType" color="arcoblue">
          {{ $t('app.logrotate.form.readonly_tag') }}
        </a-tag>
      </a-space>
    </template>

    <div class="form-drawer">
      <a-tabs v-model:active-key="activeMode" @change="handleTabChange">
        <a-tab-pane key="overview" :title="$t('app.logrotate.mode.overview')">
          <ConfigOverview
            :form-data="formData"
            :raw-content="rawContent"
            :editing="overviewEditing"
            :is-edit="isEdit"
            :current-type="currentType"
            :frequency-options="frequencyOptions"
            :host-id="currentHostId ? Number(currentHostId) : undefined"
            @update-form-data="handleUpdateFormData"
          />
        </a-tab-pane>

        <a-tab-pane key="raw" :title="$t('app.logrotate.mode.raw')">
          <RawTab v-model:content="rawContent" :readonly="!overviewEditing" />
        </a-tab-pane>
      </a-tabs>

      <div class="drawer-footer">
        <a-space>
          <a-button @click="handleCancel">
            {{ $t('common.cancel') }}
          </a-button>
          <a-button
            v-if="activeMode === 'overview' && !overviewEditing"
            type="primary"
            :disabled="isSystemType"
            @click="overviewEditing = true"
          >
            {{ $t('app.logrotate.overview.edit_button') }}
          </a-button>
          <a-button
            v-if="activeMode === 'overview' && overviewEditing"
            @click="overviewEditing = false"
          >
            {{ $t('app.logrotate.overview.view_button') }}
          </a-button>
          <a-button
            type="primary"
            :loading="loading"
            :disabled="
              isSystemType ||
              !isFormChanged ||
              (activeMode === 'overview' && !overviewEditing) ||
              (activeMode === 'raw' && !overviewEditing)
            "
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
  import { computed, ref, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { LOGROTATE_TYPE } from '@/config/enum';

  import useLoading from '@/composables/loading';
  import useCurrentHost from '@/composables/current-host';
  import useVisible from '@/composables/visible';
  import { useLogger } from '@/composables/use-logger';
  import { DEFAULT_LOGROTATE_CATEGORY } from '../../constants';
  import { useRawContentParser } from './composables/use-raw-content-parser';
  import { useFormState } from './composables/use-form-state';
  import { useLogrotateApi } from './composables/use-logrotate-api';

  import ConfigOverview from './config-overview.vue';
  import RawTab from './raw-tab.vue';

  import type { ActiveMode, ShowParams } from './types';

  const DRAWER_WIDTH = 960;
  const { t } = useI18n();

  const emit = defineEmits<{
    ok: [];
  }>();
  const overviewEditing = ref(false);

  const { log } = useLogger('LogrotateConfigViewDrawer');
  const { loading, setLoading } = useLoading();
  const { currentHostId } = useCurrentHost();
  const { visible, show: showDrawer, hide: hideDrawer } = useVisible();

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
    frequencyOptions,
    drawerTitle,
    resetForm,
    resetState,
    updateForm,
    updateOriginalState,
  } = useFormState();

  const isSystemType = computed(
    () => currentType.value === LOGROTATE_TYPE.System
  );

  const { rawContent, generateRawContent, parseRawContentToForm } =
    useRawContentParser();

  const { loadContent, submitLogrotate } = useLogrotateApi(setLoading);

  const handleCancel = () => {
    hideDrawer();
    resetState();
    overviewEditing.value = false;
    rawContent.value = '';
  };

  const handleUpdateFormData = (field: string, value: any) => {
    updateForm({ [field]: value });
  };

  const isRawContentChanged = computed(
    () => rawContent.value !== originalRawContent.value
  );

  const isFormChanged = computed(() => {
    if (activeMode.value === 'raw') {
      return isRawContentChanged.value;
    }
    return originalFormData.value
      ? JSON.stringify(formData) !== JSON.stringify(originalFormData.value)
      : false;
  });

  const handleTabChange = async (mode: string | number) => {
    const targetMode = String(mode) as ActiveMode;
    const currentMode = previousMode.value;

    if (targetMode === currentMode) {
      return;
    }

    if (targetMode === 'raw' && currentMode === 'overview') {
      // 仅在进入编辑态后才将结构化字段写回 raw，避免只读查看触发漂移
      if (overviewEditing.value && formData.path && formData.name) {
        generateRawContent(formData);
      }
    }

    if (targetMode === 'overview' && currentMode === 'raw') {
      const parsedData = parseRawContentToForm(formData);
      if (parsedData) {
        updateForm(parsedData);
      }
    }

    previousMode.value = targetMode;
    activeMode.value = targetMode;
  };

  const validateForOverviewSubmit = (): boolean => {
    if (!formData.name?.trim()) {
      Message.warning(t('app.logrotate.form.name_required'));
      return false;
    }

    if (!/^[a-zA-Z0-9_-]+$/.test(formData.name.trim())) {
      Message.warning(t('app.logrotate.form.name_pattern'));
      return false;
    }

    if (!formData.path?.trim()) {
      Message.warning(t('app.logrotate.form.path_required'));
      return false;
    }

    if (!formData.count || formData.count < 1) {
      Message.warning(t('app.logrotate.form.count_min'));
      return false;
    }

    return true;
  };

  const handleSubmit = async () => {
    if (isSystemType.value) {
      Message.warning(t('app.logrotate.form.system_readonly'));
      return;
    }

    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error(t('common.host_id_required'));
      return;
    }

    try {
      if (activeMode.value === 'raw' && !overviewEditing.value) {
        return;
      }

      if (activeMode.value === 'overview' && !validateForOverviewSubmit()) {
        return;
      }

      const successMessage = await submitLogrotate(
        activeMode.value === 'raw' ? 'raw' : 'form',
        formData,
        rawContent.value,
        isEdit.value,
        currentType.value,
        originalCategory.value,
        originalName.value,
        Number(hostId)
      );

      Message.success(successMessage);
      emit('ok');
      handleCancel();
    } catch (error) {
      log('提交失败:', error);
    }
  };

  const show = async (params?: ShowParams) => {
    showDrawer();
    activeMode.value = 'overview';
    previousMode.value = 'overview';
    overviewEditing.value = !params?.isEdit;

    if (params) {
      currentType.value = params.type || LOGROTATE_TYPE.Local;
      isEdit.value = params.isEdit || false;

      if (params.isEdit && params.record) {
        originalName.value = params.record.name;
        originalCategory.value = params.record.category;

        updateForm({
          name: params.record.name,
          category:
            params.record.category ||
            (currentType.value === LOGROTATE_TYPE.System
              ? ''
              : DEFAULT_LOGROTATE_CATEGORY),
          path: '',
          frequency: formData.frequency,
          count: 7,
          compress: false,
          delayCompress: false,
          missingOk: false,
          notIfEmpty: false,
          create: 'create 0644 root root',
          preRotate: '',
          postRotate: '',
        });

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

          if (rawContent.value) {
            const parsedData = parseRawContentToForm(formData);
            if (parsedData) {
              updateForm(parsedData);
            }
          }

          await updateOriginalState();
          originalRawContent.value = rawContent.value;
        } catch (error) {
          log('加载内容失败:', error);
        }
      } else {
        resetForm();
        updateForm({
          category:
            params.type === LOGROTATE_TYPE.System
              ? ''
              : DEFAULT_LOGROTATE_CATEGORY,
        });
        await updateOriginalState();
        originalRawContent.value = rawContent.value;
      }
    } else {
      resetForm();
      updateForm({
        category:
          currentType.value === LOGROTATE_TYPE.System
            ? ''
            : DEFAULT_LOGROTATE_CATEGORY,
      });
      await updateOriginalState();
      originalRawContent.value = rawContent.value;
    }
  };

  watch(
    formData,
    (newData) => {
      if (activeMode.value === 'overview' && newData.path && newData.name) {
        generateRawContent(newData);
      }
    },
    { deep: true }
  );

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
    min-height: 0;
  }

  .form-drawer :deep(.arco-tabs-content) {
    flex: 1;
    min-height: 0;
    overflow: hidden auto;
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
