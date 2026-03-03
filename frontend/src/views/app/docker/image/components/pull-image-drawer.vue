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
    <a-alert class="mb-4 pull-guide-alert" type="info" :show-icon="true">
      <div class="guide-title">{{
        t('app.docker.image.pull.guide.title')
      }}</div>
      <div class="guide-desc">{{ t('app.docker.image.pull.guide.desc') }}</div>
      <a-link @click="openDockerHub">
        {{ t('app.docker.image.pull.guide.link') }}
      </a-link>
    </a-alert>
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
  const openDockerHub = () => {
    window.open('https://hub.docker.com/search?type=image', '_blank');
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
        Message.error(t('app.docker.image.list.pull.failed'));
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

<style scoped>
  .pull-guide-alert .guide-title {
    margin-bottom: 0.25rem;
    font-weight: 600;
  }

  .pull-guide-alert .guide-desc {
    margin-bottom: 0.25rem;
    font-size: 13px;
    line-height: 1.5;
  }
</style>
