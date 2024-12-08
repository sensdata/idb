<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('app.file.ownerDrawer.title')"
    unmountOnClose
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form :model="formState" :rules="rules">
      <a-form-item field="path" :label="$t('app.file.ownerDrawer.path')">
        <span>{{ formState.path }}</span>
      </a-form-item>
      <a-form-item field="user" :label="$t('app.file.ownerDrawer.user')">
        <a-input v-model="formState.user" class="w-60" />
      </a-form-item>
      <a-form-item field="group" :label="$t('app.file.ownerDrawer.group')">
        <a-input v-model="formState.group" class="w-60" />
      </a-form-item>
      <a-form-item field="sub" label=" ">
        <a-checkbox v-model="formState.sub">
          {{ $t('app.file.ownerDrawer.sub') }}
        </a-checkbox>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/hooks/loading';
  import { FileInfoEntity } from '@/entity/FileInfo';

  const { t } = useI18n();

  const formState = reactive({
    path: '',
    user: '',
    group: '',
    sub: true,
  });

  const rules = {
    user: { required: true, message: t('app.file.ownerDrawer.userRequired') },
    group: { required: true, message: t('app.file.ownerDrawer.groupRequired') },
  };

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const setData = (data: FileInfoEntity) => {
    formState.path = data.path;
    formState.user = data.user;
    formState.group = data.group;
  };

  const handleOk = () => {};
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
