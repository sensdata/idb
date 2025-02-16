<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.ownerDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item field="path" :label="$t('components.file.ownerDrawer.path')">
        <span>{{ formState.path }}</span>
      </a-form-item>
      <a-form-item field="user" :label="$t('components.file.ownerDrawer.user')">
        <a-input v-model="formState.user" class="w-60" />
      </a-form-item>
      <a-form-item field="group" :label="$t('components.file.ownerDrawer.group')">
        <a-input v-model="formState.group" class="w-60" />
      </a-form-item>
      <a-form-item field="sub" label=" ">
        <a-checkbox v-model="formState.sub">
          {{ $t('components.file.ownerDrawer.sub') }}
        </a-checkbox>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useLoading from '@/hooks/loading';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { updateFileOwnerApi } from '@/api/file';

  const { t } = useI18n();

  const emit = defineEmits(['ok']);

  const formRef = ref();
  const formState = reactive({
    path: '',
    user: '',
    group: '',
    sub: true,
  });

  const rules = {
    user: { required: true, message: t('components.file.ownerDrawer.userRequired') },
    group: { required: true, message: t('components.file.ownerDrawer.groupRequired') },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const setData = (data: FileInfoEntity) => {
    formState.path = data.path;
    formState.user = data.user;
    formState.group = data.group;
  };

  const getData = () => {
    return {
      source: formState.path,
      user: formState.user,
      group: formState.group,
      sub: formState.sub,
    };
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleBeforeOk = async (done: any) => {
    if (await validate()) {
      try {
        setLoading(true);
        const data = getData();
        await updateFileOwnerApi(data);
        done();
        Message.success(t('components.file.modeDrawer.message.success'));
        emit('ok');
        return true;
      } finally {
        setLoading(false);
      }
    }
    return false;
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
