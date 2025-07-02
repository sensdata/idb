<template>
  <a-modal
    v-model:visible="visible"
    :title="
      isEdit
        ? $t('app.crontab.category.action.edit')
        : $t('app.crontab.category.action.create')
    "
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="$t('app.crontab.category.form.name')"
        :rules="[
          {
            required: true,
            message: $t('app.crontab.category.form.name.required'),
          },
        ]"
      >
        <a-input
          v-model="formState.name"
          :placeholder="$t('app.crontab.category.form.name.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    createCrontabCategoryApi,
    updateCrontabCategoryApi,
  } from '@/api/crontab';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';

  const props = defineProps<{
    type: CRONTAB_TYPE;
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
      { required: true, message: t('app.crontab.category.form.name.required') },
    ],
  });

  const handleOk = async () => {
    try {
      await formRef.value?.validate();
      if (isEdit.value) {
        await updateCrontabCategoryApi({
          type: props.type,
          category: editingCategory.value!,
          new_name: formState.name,
        });
        Message.success(t('app.crontab.category.message.update_success'));
      } else {
        await createCrontabCategoryApi({
          type: props.type,
          category: formState.name,
        });
        Message.success(t('app.crontab.category.message.create_success'));
      }
      visible.value = false;
      emit('ok');
    } catch (err: any) {
      // 如果是因为数据为空或分类不存在导致的错误，不显示错误信息
      if (
        err?.message &&
        !err.message.includes('not found') &&
        !err.message.includes('empty') &&
        !err.message.includes('不存在')
      ) {
        Message.error(err.message);
      }
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
