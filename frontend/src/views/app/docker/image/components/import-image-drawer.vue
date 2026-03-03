<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.image.list.action.import')"
    :ok-loading="loading"
    unmount-on-close
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-alert class="mb-4" type="info" :show-icon="true">
      <div class="guide-title">{{
        t('app.docker.image.import.guide.title')
      }}</div>
      <div class="guide-desc">{{
        t('app.docker.image.import.guide.desc')
      }}</div>
    </a-alert>
    <a-form
      ref="formRef"
      :model="formState"
      :rules="rules"
      :labelAlign="'left'"
    >
      <a-form-item
        field="path"
        :label="t('app.docker.image.form.path')"
        label-col-flex="80px"
        :rules="[
          {
            required: true,
            message: t('app.docker.image.form.path.required'),
          },
        ]"
      >
        <file-selector
          v-model="formState.path"
          :initial-path="formState.path"
          type="file"
          :placeholder="$t('app.docker.image.form.path.placeholder')"
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
  import { importImageApi } from '@/api/docker';
  import FileSelector from '@/components/file/file-selector/index.vue';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    path: '',
  });
  const rules = {
    path: [
      {
        required: true,
        message: t('app.docker.image.form.path.required'),
      },
    ],
  };

  const show = () => {
    visible.value = true;
    formState.path = '';
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
  const onBeforeOk = async () => {
    if (!(await validate())) {
      return false;
    }

    let success = false;
    try {
      loading.value = true;
      await importImageApi({ path: formState.path });
      Message.success(t('app.docker.image.list.import.success'));
      emit('success');
      hide();
      success = true;
    } catch (e: any) {
      if (e?.message) {
        Message.error(e.message);
      }
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
