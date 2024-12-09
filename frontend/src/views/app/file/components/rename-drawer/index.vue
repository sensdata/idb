<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('app.file.renameDrawer.title')"
    unmountOnClose
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form :model="formState" :rules="rules">
        <a-form-item field="path" :label="$t('app.file.renameDrawer.path')">
          <span>{{ pwd }}</span>
        </a-form-item>
        <a-form-item field="name" :label="$t('app.file.renameDrawer.name')">
          <a-input v-model="formState.name" />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/hooks/loading';
  import { renameFileApi } from '@/api/file';

  const emit = defineEmits(['success']);

  const { t } = useI18n();

  const formState = reactive({
    path: '',
    name: '',
  });

  const pwd = computed(() => formState.path.split('/').slice(0, -1).join('/'));

  const rules = {
    name: {
      required: true,
      message: t('app.file.renameDrawer.name_required'),
    },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const setData = (data: { path: string }) => {
    formState.path = data.path;
  };

  const handleOk = async () => {
    setLoading(true);
    try {
      await renameFileApi({
        source: formState.path,
        name: formState.name,
      });
      visible.value = false;
      emit('success');
    } finally {
      setLoading(false);
    }
  };
  const handleCancel = () => {
    visible.value = false;
  };

  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
