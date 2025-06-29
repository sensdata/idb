<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.image.list.action.pull')"
    :ok-loading="loading"
    unmount-on-close
    @ok="onOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="image_name"
        :label="t('app.docker.image.form.image_name')"
        :rules="[
          {
            required: true,
            message: t('app.docker.image.form.image_name.required'),
          },
        ]"
      >
        <a-input
          v-model="formState.image_name"
          :placeholder="t('app.docker.image.form.image_name.placeholder')"
          :disabled="loading"
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import { pullImageApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    image_name: '',
  });
  const rules = {
    image_name: [
      {
        required: true,
        message: t('app.docker.image.form.image_name.required'),
      },
    ],
  };

  const show = () => {
    visible.value = true;
    formState.image_name = '';
    formRef.value?.resetFields();
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };
  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };
  const onOk = async () => {
    if (!(await validate())) {
      return;
    }

    try {
      loading.value = true;
      const result = await pullImageApi({ image_name: formState.image_name });
      if (result.success) {
        Message.success(
          t('app.docker.image.list.pull.success', {
            command: result.command,
          })
        );
      } else {
        Message.success(t('app.docker.image.list.pull.failed'));
      }
      emit('success');
      hide();
    } catch (e: any) {
      Message.error(e.message);
    } finally {
      loading.value = false;
    }
  };
  const onCancel = hide;

  defineExpose({ show });
</script>

<style scoped></style>
