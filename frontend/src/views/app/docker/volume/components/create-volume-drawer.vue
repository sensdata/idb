<template>
  <a-drawer
    v-model:visible="visible"
    :width="500"
    :title="t('app.docker.volume.create.title')"
    :ok-loading="loading"
    unmount-on-close
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="t('app.docker.volume.create.form.name')"
      >
        <a-input
          v-model="formState.name"
          :placeholder="t('app.docker.volume.create.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item
        field="driver"
        :label="t('app.docker.volume.create.form.driver')"
      >
        <a-tag :color="'rgb(var(--success-6))'">local</a-tag>
      </a-form-item>
      <a-form-item
        field="options"
        :label="t('app.docker.volume.create.form.options')"
      >
        <a-textarea
          v-model="formState.optionsText"
          :placeholder="t('app.docker.volume.create.form.options.placeholder')"
          :auto-size="{
            minRows: 3,
            maxRows: 6,
          }"
        />
      </a-form-item>
      <a-form-item
        field="labels"
        :label="t('app.docker.volume.create.form.labels')"
      >
        <a-textarea
          v-model="formState.labelsText"
          :placeholder="t('app.docker.volume.create.form.labels.placeholder')"
          :auto-size="{
            minRows: 3,
            maxRows: 6,
          }"
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue';
  import { Message } from '@arco-design/web-vue';
  import { createVolumeApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    name: '',
    driver: '',
    optionsText: '',
    labelsText: '',
  });
  const rules = {
    name: [
      {
        required: true,
        message: t('app.docker.volume.create.form.name.required'),
      },
    ],
    driver: [
      {
        required: true,
        message: t('app.docker.volume.create.form.driver.required'),
      },
    ],
  };
  const show = () => {
    visible.value = true;
    formState.name = '';
    formState.driver = '';
    formState.optionsText = '';
    formState.labelsText = '';
    formRef.value?.resetFields();
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };
  const onBeforeOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) {
      return false;
    }
    try {
      loading.value = true;
      await createVolumeApi({
        name: formState.name,
        driver: formState.driver,
        options: formState.optionsText
          .split('\n')
          .map((line) => line.trim())
          .filter(Boolean),
        labels: formState.labelsText
          .split('\n')
          .map((line) => line.trim())
          .filter(Boolean),
      });
      Message.success(t('app.docker.volume.create.success'));
      emit('success');
      hide();
    } catch (e: any) {
      Message.error(e?.message || t('app.docker.volume.create.failed'));
    } finally {
      loading.value = false;
    }
    return true;
  };
  const onCancel = hide;
  defineExpose({ show });
</script>

<style scoped></style>
