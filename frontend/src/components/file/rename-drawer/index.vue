<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.renameDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item
        field="path"
        :label="$t('components.file.renameDrawer.path')"
      >
        <span>{{ pwd }}</span>
      </a-form-item>
      <a-form-item
        field="name"
        :label="$t('components.file.renameDrawer.name')"
      >
        <a-input v-model="formState.name" />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/composables/loading';
  import { renameFileApi } from '@/api/file';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const formRef = ref();
  const formState = reactive({
    path: '',
    name: '',
  });

  const pwd = computed(() => formState.path.split('/').slice(0, -1).join('/'));

  const rules = {
    name: {
      required: true,
      message: t('components.file.renameDrawer.name_required'),
    },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  const showLoading = () => {
    setLoading(true);
  };

  const hideLoading = () => {
    setLoading(false);
  };

  const setData = (data: { path: string }) => {
    formState.path = data.path;
    formState.name = data.path.split('/').pop() || '';
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      await renameFileApi({
        source: formState.path,
        name: formState.name,
      });
      Message.success(t('components.file.renameDrawer.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
