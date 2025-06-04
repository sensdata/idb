<template>
  <a-drawer
    v-model:visible="visible"
    :width="500"
    :title="t('app.docker.image.list.operation.push')"
    :ok-loading="loading"
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :rules="rules" :model="formState" layout="vertical">
      <a-form-item field="name" :label="t('app.docker.image.form.name')">
        <a-input
          v-model="formState.name"
          :placeholder="t('app.docker.image.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item field="tag_name" :label="t('app.docker.image.form.tag')">
        <a-select
          v-model="formState.tag_name"
          :placeholder="t('app.docker.image.form.tag.placeholder')"
          allow-clear
        >
          <a-option v-for="tag in tagOptions" :key="tag" :value="tag">
            {{ tag }}
          </a-option>
        </a-select>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue';
  import { Message } from '@arco-design/web-vue';
  import { pushImageApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    name: '',
    tag_name: '',
  });
  let currentRecord: any = null;
  const tagOptions = ref<string[]>([]);

  const show = (record?: any) => {
    currentRecord = record;
    tagOptions.value = Array.isArray(record?.tags) ? record.tags : [];
    formState.tag_name = tagOptions.value[0] || '';
    formState.name = '';
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };
  const rules = {
    name: [
      {
        required: true,
        message: t('app.docker.image.form.name.required'),
      },
    ],
    tag_name: [
      {
        required: true,
        message: t('app.docker.image.form.tag.required'),
      },
    ],
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const onBeforeOk = async () => {
    if (!currentRecord?.id) {
      return false;
    }
    if (!(await validate())) {
      return false;
    }

    try {
      loading.value = true;
      await pushImageApi(currentRecord.id, {
        name: formState.name,
        tag_name: formState.tag_name,
      });
      Message.success(t('common.message.operationSuccess'));
      emit('success');
      return true;
    } catch (e: any) {
      Message.error(e?.message || t('common.message.operationError'));
    } finally {
      loading.value = false;
    }

    return false;
  };
  const onCancel = hide;

  defineExpose({ show });
</script>

<style scoped></style>
