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
    <a-alert class="mb-4" type="info" :show-icon="true">
      <div class="guide-title">{{
        t('app.docker.volume.create.guide.title')
      }}</div>
      <div class="guide-desc">{{
        t('app.docker.volume.create.guide.desc')
      }}</div>
    </a-alert>
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
    driver: 'local',
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
    formState.driver = 'local';
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
    let success = false;
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
      success = true;
    } catch (e: any) {
      Message.error(e?.message || t('app.docker.volume.create.failed'));
    } finally {
      loading.value = false;
    }
    return success;
  };
  const onCancel = hide;
  defineExpose({ show });
</script>

<style scoped>
  .guide-title {
    margin-bottom: 0.25rem;
    font-weight: 600;
  }

  .guide-desc {
    font-size: 13px;
    line-height: 1.5;
  }
</style>
