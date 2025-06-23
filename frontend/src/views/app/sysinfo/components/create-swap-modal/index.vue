<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.createSwap.title')"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState">
      <a-form-item class="mb-0">
        <p class="text-gray-500">
          {{ $t('app.sysinfo.createSwap.input_tip') }}
        </p>
      </a-form-item>
      <a-form-item
        field="size"
        :label="$t('app.sysinfo.createSwap.size')"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.createSwap.size_required'),
          },
        ]"
      >
        <div class="flex items-center">
          <a-input-number
            v-model="formState.size"
            :min="1"
            :max="16"
            :step="1"
            class="flex-1"
            hide-button
          />
          <a-select v-model="formState.unit" class="ml-2 w-24">
            <a-option value="GB">GB</a-option>
            <a-option value="MB">MB</a-option>
          </a-select>
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { createSwapApi } from '@/api/sysinfo';
  import useVisible from '@/composables/visible';
  import useLoading from '@/composables/loading';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const formState = reactive({
    unit: 'GB',
    size: 1,
  });

  const handleOk = async () => {
    try {
      const errors = await formRef.value.validate();
      if (errors) {
        return;
      }
      let size = 0;
      switch (formState.unit) {
        case 'GB':
          size = formState.size * 1024;
          break;
        case 'MB':
        default:
          size = formState.size;
          break;
      }
      showLoading();
      await createSwapApi({ size });
      Message.success(t('app.sysinfo.createSwap.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.createSwap.failed'));
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const reset = () => {
    formState.size = 1;
    formState.unit = 'GB';
    formRef.value?.resetFields();
  };

  defineExpose({
    show,
    hide,
    reset,
  });
</script>
