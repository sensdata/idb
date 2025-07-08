<template>
  <a-modal
    v-model:visible="visible"
    :title="isEdit ? $t('category.action.edit') : $t('category.action.create')"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="$t('category.form.name')"
        :rules="[
          {
            required: true,
            message: $t('category.form.name.required'),
          },
        ]"
      >
        <a-input
          v-model="formState.name"
          :placeholder="$t('category.form.name.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { CategoryManageConfig } from '../types/category';

  interface Props {
    config: CategoryManageConfig;
  }

  const props = defineProps<Props>();
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
    name: [{ required: true, message: t('category.form.name.required') }],
  });

  const handleOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) {
      return;
    }
    try {
      const params = {
        category: isEdit.value ? editingCategory.value! : formState.name,
        ...(props.config.params || {}),
      };

      if (isEdit.value) {
        await props.config.api.update({
          ...params,
          new_name: formState.name,
        });
        Message.success(t('category.message.update_success'));
      } else {
        await props.config.api.create(params);
        Message.success(t('category.message.create_success'));
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
