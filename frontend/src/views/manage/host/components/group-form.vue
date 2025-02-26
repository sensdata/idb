<template>
  <a-modal
    v-model:visible="visible"
    :title="
      $t(
        isEdit
          ? 'manage.host.group.form.title.edit'
          : 'manage.host.group.form.title.create'
      )
    "
    width="400px"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="model"
      :rules="rules"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 18 }"
    >
      <a-form-item
        field="group_name"
        :label="$t('manage.host.group.form.name.label')"
      >
        <a-input
          v-model="model.group_name"
          :placeholder="$t('manage.host.group.form.name.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { createHostGroupApi, updateHostGroupApi } from '@/api/host';

  interface HostGroup {
    id: number;
    group_name: string;
  }

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const model = reactive({
    id: 0,
    group_name: '',
  });

  const isEdit = computed(() => model.id > 0);

  const rules = computed(() => ({
    group_name: [
      { required: true, message: t('manage.host.group.form.name.required') },
    ],
  }));

  const getData = () => {
    const data = toRaw(model);
    return data;
  };

  const reset = () => {
    model.id = 0;
    model.group_name = '';
    formRef.value?.resetFields();
    formRef.value?.clearValidate();
  };

  const setData = (data: HostGroup) => {
    model.id = data.id;
    model.group_name = data.group_name;
  };

  const validate = () => {
    return formRef.value.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      const data = getData();
      if (isEdit.value) {
        await updateHostGroupApi(data);
        Message.success(t('manage.host.group.form.update.success'));
      } else {
        await createHostGroupApi(data);
        Message.success(t('manage.host.group.form.save.success'));
      }
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err);
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
    reset,
    setData,
  });
</script>
