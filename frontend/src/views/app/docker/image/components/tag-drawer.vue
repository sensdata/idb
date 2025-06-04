<template>
  <a-drawer
    v-model:visible="visible"
    :width="500"
    :title="t('app.docker.image.list.operation.tag')"
    :ok-loading="loading"
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :rules="rules" :model="formState" layout="vertical">
      <a-form-item field="tag" :label="t('app.docker.image.form.tag')">
        <a-input
          v-model="formState.tag"
          :placeholder="t('app.docker.image.form.tag.placeholder')"
          :max="1"
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
  import { setImageTagApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    tag: '',
  });
  let currentRecord: any = null;

  const show = (record?: any) => {
    currentRecord = record;
    formState.tag = record.tags?.[0] || '';
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };

  const rules = {
    tag: [
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
      await setImageTagApi({
        source_id: currentRecord.id,
        target_name: formState.tag.trim(),
      });
      Message.success(t('common.message.operationSuccess'));
      emit('success');
      hide();
    } catch (e: any) {
      Message.error(e?.message || t('common.message.operationError'));
    } finally {
      loading.value = false;
    }

    return true;
  };
  const onCancel = hide;

  defineExpose({ show });
</script>

<style scoped></style>
