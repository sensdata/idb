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
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item
        field="path"
        :label="t('app.docker.image.form.path')"
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

    try {
      loading.value = true;
      await importImageApi({ path: formState.path });
      Message.success(t('app.docker.image.list.import.success'));
      emit('success');
      hide();
    } catch (e: any) {
      if (e?.message) {
        Message.error(e.message);
      }
    } finally {
      loading.value = false;
    }

    return true;
  };
  const onCancel = hide;

  defineExpose({ show });
</script>

<style scoped></style>
