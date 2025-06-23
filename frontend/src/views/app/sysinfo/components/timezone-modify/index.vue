<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.timezoneModify.title')"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState">
      <a-form-item
        class="mb-0"
        field="timezone"
        :label="$t('app.sysinfo.timezoneModify.label')"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.timezoneModify.required'),
          },
        ]"
      >
        <a-select
          v-model="formState.timezone"
          :placeholder="$t('app.sysinfo.timezoneModify.placeholder')"
          :loading="timezoneLoading"
          :options="timezoneOptions"
          allow-search
          allow-clear
        />
      </a-form-item>
      <!-- <a-form-item>
        <p class="text-gray-500 text-sm">
          {{ $t('app.sysinfo.timezoneModify.example') }}
        </p>
      </a-form-item> -->
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOptionData } from '@arco-design/web-vue';
  import { updateTimeZoneApi } from '@/api/sysinfo';
  import { getTimezonesApi } from '@/api/settings';
  import useVisible from '@/composables/visible';
  import useLoading from '@/composables/loading';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const formState = reactive({
    timezone: '',
  });

  const timezoneOptions = ref<SelectOptionData[]>([]);
  const timezoneLoading = ref(false);
  const getTimeZoneOptions = async () => {
    timezoneLoading.value = true;
    try {
      const res = await getTimezonesApi({
        page: 1,
        page_size: 1000,
      });
      timezoneOptions.value = res.items.map((item) => ({
        label: item.utc,
        value: item.utc,
      }));
    } finally {
      timezoneLoading.value = false;
    }
  };
  onMounted(() => {
    getTimeZoneOptions();
  });

  const handleOk = async () => {
    try {
      const errors = await formRef.value.validate();
      if (errors) {
        return;
      }

      showLoading();
      await updateTimeZoneApi({ timezone: formState.timezone });
      Message.success(t('app.sysinfo.timezoneModify.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.timezoneModify.failed'));
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
