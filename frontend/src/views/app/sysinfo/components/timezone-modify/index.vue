<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.overview.timezone_modify_title')"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState">
      <a-form-item
        class="mb-0"
        field="timezone"
        :label="$t('app.sysinfo.overview.timezone_modify_label')"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.overview.timezone_modify_required'),
          },
        ]"
      >
        <a-input
          v-model="formState.timezone"
          :placeholder="$t('app.sysinfo.overview.timezone_modify_placeholder')"
        />
      </a-form-item>
      <a-form-item>
        <p class="text-gray-500 text-sm">
          {{ $t('app.sysinfo.overview.timezone_modify_example') }}
        </p>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { updateTimeZoneApi } from '@/api/sysinfo';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const formState = reactive({
    timezone: '',
  });

  const handleOk = async () => {
    try {
      const errors = await formRef.value.validate();
      if (errors) {
        return;
      }

      showLoading();
      await updateTimeZoneApi({ timezone: formState.timezone });
      Message.success(t('app.sysinfo.overview.timezone_modify_success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(
        err.message || t('app.sysinfo.overview.timezone_modify_failed')
      );
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const setTimeZone = (timezone: string) => {
    formState.timezone = timezone;
  };

  const reset = () => {
    formState.timezone = '';
    formRef.value?.resetFields();
  };

  defineExpose({
    show,
    hide,
    reset,
    setTimeZone,
  });
</script>
