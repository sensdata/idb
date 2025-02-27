<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('components.timeModify.title')"
    :ok-loading="loading"
    @before-ok="handleBeforeOk"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 18 }"
    >
      <a-form-item :label="$t('components.timeModify.current_time')">
        <span>{{ currentTime }}</span>
      </a-form-item>
      <a-form-item
        :label="$t('components.timeModify.new_time')"
        field="newTime"
        :rules="[
          {
            required: true,
            message: $t('components.timeModify.new_time_required'),
          },
        ]"
      >
        <a-date-picker
          v-model="formState.newTime"
          show-time
          format="YYYY-MM-DD HH:mm:ss"
          value-format="YYYY-MM-DD HH:mm:ss"
          :placeholder="$t('components.timeModify.select_time')"
          style="width: 100%"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { updateTimeApi } from '@/api/sysinfo';
  import dayjs from 'dayjs';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref<FormInstance>();
  const formState = reactive({
    newTime: undefined as string | undefined,
  });
  const currentTime = ref('');

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    try {
      await formRef.value?.validate();

      if (!formState.newTime) {
        Message.error(t('components.timeModify.new_time_required'));
        done(false);
        return;
      }

      showLoading();

      try {
        await updateTimeApi({
          timestamp: dayjs(formState.newTime, 'YYYY-MM-DD HH:mm:ss').unix(),
        });

        Message.success(t('components.timeModify.update_success'));
        emit('ok');
        done(true);
      } catch (err: any) {
        Message.error(err.message || t('components.timeModify.update_failed'));
        done(false);
      } finally {
        hideLoading();
      }
    } catch (err) {
      done(false);
    }
  };

  const setCurrentTime = (time: string) => {
    currentTime.value = time;
    formState.newTime = time;
  };

  defineExpose({
    show,
    hide,
    setCurrentTime,
  });
</script>
