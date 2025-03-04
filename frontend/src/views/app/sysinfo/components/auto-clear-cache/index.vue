<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.autoClearCache.title')"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <p class="text-gray-500 text-center mt-0 mb-4">
      {{ $t('app.sysinfo.autoClearCache.tip') }}
    </p>
    <a-form
      ref="formRef"
      :model="formState"
      :label-col-props="{ span: 8 }"
      :wrapper-col-props="{ span: 16 }"
    >
      <a-form-item
        field="enabled"
        :label="$t('app.sysinfo.autoClearCache.enabled')"
      >
        <a-switch v-model="formState.enabled" />
      </a-form-item>
      <a-form-item
        v-if="formState.enabled"
        field="interval"
        :label="$t('app.sysinfo.autoClearCache.interval')"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.autoClearCache.interval_required'),
          },
          {
            validator: validateInterval,
            message: $t('app.sysinfo.autoClearCache.interval_invalid'),
          },
        ]"
      >
        <div class="flex items-center">
          <a-input-number
            v-model="formState.interval"
            :min="1"
            :max="24"
            :step="1"
            class="w-20"
            hide-button
          />
          <span class="ml-2">{{ $t('app.sysinfo.autoClearCache.hours') }}</span>
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    setAutoClearMemoryCacheApi,
    getAutoClearMemoryCacheApi,
  } from '@/api/sysinfo';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const formState = reactive({
    enabled: false,
    interval: 1,
  });

  watch(
    () => formState.enabled,
    () => {
      formRef.value?.clearValidate();
    }
  );

  const validateInterval = (
    value: number,
    callback: (error?: string) => void
  ) => {
    if (formState.enabled && (value < 1 || value > 24)) {
      callback(t('app.sysinfo.autoClearCache.interval_invalid'));
    } else {
      callback();
    }
  };

  const setConfig = (intervalHours: number) => {
    if (intervalHours > 0) {
      formState.enabled = true;
      formState.interval = intervalHours;
    } else {
      formState.enabled = false;
      formState.interval = 1;
    }
  };

  const load = async () => {
    try {
      showLoading();
      const res = await getAutoClearMemoryCacheApi();
      setConfig(res.interval);
    } catch (err: any) {
      Message.error(err.message);
    } finally {
      hideLoading();
    }
  };

  const handleOk = async () => {
    try {
      const errors = await formRef.value.validate();
      if (errors) {
        return;
      }

      showLoading();
      const interval = formState.enabled ? formState.interval : 0;
      await setAutoClearMemoryCacheApi({ interval });
      Message.success(t('app.sysinfo.autoClearCache.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.autoClearCache.failed'));
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const reset = () => {
    formState.enabled = false;
    formState.interval = 1;
    formRef.value?.resetFields();
  };

  defineExpose({
    show,
    hide,
    reset,
    load,
    setConfig,
  });
</script>
