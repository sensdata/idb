<template>
  <a-modal
    v-model:visible="visible"
    :title="
      isEdit
        ? $t('app.script.category.action.edit')
        : $t('app.script.category.action.create')
    "
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="$t('app.script.category.form.name')"
        :rules="[
          {
            required: true,
            message: $t('app.script.category.form.name.required'),
          },
        ]"
      >
        <a-input
          v-model="formState.name"
          :placeholder="$t('app.script.category.form.name.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    createScriptCategoryApi,
    updateScriptCategoryApi,
  } from '@/api/script';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';

  const props = defineProps<{
    type: SCRIPT_TYPE;
  }>();

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const visible = ref(false);
  const formRef = ref();
  const editingCategory = ref<string | null>(null);
  const isEdit = computed(() => !!editingCategory.value);
  const formState = reactive({
    name: '',
  });
  const rules = ref({
    name: [
      { required: true, message: t('app.script.category.form.name.required') },
    ],
  });

  const handleOk = async () => {
    try {
      await formRef.value?.validate();
      if (isEdit.value) {
        await updateScriptCategoryApi({
          type: props.type,
          category: editingCategory.value!,
          new_name: formState.name,
        });
        Message.success(t('app.script.category.message.update_success'));
      } else {
        await createScriptCategoryApi({
          type: props.type,
          category: formState.name,
        });
        Message.success(t('app.script.category.message.create_success'));
      }
      visible.value = false;
      emit('ok');
    } catch (err: any) {
      Message.error(err?.message);
    }
  };

  const handleCancel = () => {
    visible.value = false;
    formRef.value?.resetFields();
  };

  const show = () => {
    visible.value = true;
    formRef.value?.resetFields();
  };

  const setData = (params: { name: string }) => {
    editingCategory.value = params.name;
    formState.name = params.name;
  };

  defineExpose({
    show,
    setData,
  });
</script>
