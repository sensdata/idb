<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.createFileDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :rules="rules"
      :labelAlign="'left'"
    >
      <a-form-item
        field="name"
        :label="$t('components.file.createFileDrawer.name')"
        label-col-flex="80px"
      >
        <a-input v-model="formState.name" />
      </a-form-item>
      <a-form-item field="is_link" label=" " label-col-flex="80px">
        <a-checkbox v-model="formState.is_link">
          {{ $t('components.file.createFileDrawer.is_link') }}
        </a-checkbox>
      </a-form-item>
      <a-form-item
        v-if="formState.is_link"
        field="is_link"
        :label="$t('components.file.createFileDrawer.link_type')"
        label-col-flex="80px"
      >
        <a-radio-group v-model="formState.link_type">
          <a-radio value="soft">{{
            $t('components.file.createFileDrawer.soft')
          }}</a-radio>
          <a-radio value="hard">{{
            $t('components.file.createFileDrawer.hard')
          }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item
        v-if="formState.is_link"
        field="link_path"
        :label="$t('components.file.createFileDrawer.link_path')"
        label-col-flex="85px"
      >
        <a-input v-model="formState.link_path" />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/composables/loading';
  import { createFileApi } from '@/api/file';
  import { useHostStore } from '@/store';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const hostStore = useHostStore();
  const formRef = ref();
  const formState = reactive({
    name: '',
    pwd: '',
    is_link: false,
    link_type: 'soft',
    link_path: '',
  });

  const rules = {
    name: {
      required: true,
      message: t('components.file.createFileDrawer.name_required'),
    },
    link_path: {
      required: true,
      message: t('components.file.createFileDrawer.link_path_required'),
    },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const setData = (data: { pwd: string }) => {
    formState.pwd = data.pwd;
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    if (!(await validate())) {
      return;
    }

    if (loading.value) {
      return;
    }
    setLoading(true);
    try {
      await createFileApi({
        host: hostStore.currentId ?? hostStore.defaultId,
        source: formState.pwd + '/' + formState.name,
        is_dir: false,
        is_link: formState.is_link,
        is_symlink: formState.link_type === 'soft',
        link_path: formState.link_path,
      });
      visible.value = false;
      Message.success(t('components.file.createFileDrawer.success'));
      emit('ok');
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
