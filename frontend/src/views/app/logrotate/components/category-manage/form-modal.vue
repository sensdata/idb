<template>
  <a-modal
    v-model:visible="visible"
    :title="
      isEdit
        ? $t('app.logrotate.category.manage.edit_title')
        : $t('app.logrotate.category.manage.create_title')
    "
    :confirm-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="$t('app.logrotate.category.manage.form.name')"
      >
        <a-input
          v-model="formState.name"
          :placeholder="
            $t('app.logrotate.category.manage.form.name_placeholder')
          "
          :disabled="loading"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    createLogrotateCategoryApi,
    updateLogrotateCategoryApi,
  } from '@/api/logrotate';
  import { LOGROTATE_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';
  import useCurrentHost from '@/hooks/current-host';
  import type { FormInstance } from '@arco-design/web-vue';

  // 接口定义
  interface FormState {
    name: string;
  }

  interface CategoryData {
    name: string;
  }

  // Props 和 Emits 定义
  const props = defineProps<{
    type: LOGROTATE_TYPE;
  }>();

  const emit = defineEmits<{
    ok: [];
  }>();

  // Composables
  const { t } = useI18n();
  const { currentHostId } = useCurrentHost();

  // 响应式数据
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const editingCategory = ref<string | null>(null);

  const formState = reactive<FormState>({
    name: '',
  });

  // 计算属性
  const isEdit = computed(() => !!editingCategory.value);

  const rules = computed(() => ({
    name: [
      {
        required: true,
        message: t('app.logrotate.category.manage.form.name_required'),
      },
    ],
  }));

  // 重置表单的辅助函数
  const resetForm = () => {
    formState.name = '';
    editingCategory.value = null;
    formRef.value?.resetFields();
  };

  // 方法
  const handleOk = async () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error(
        t('common.error.host_id_required') || 'Host ID is required'
      );
      return;
    }

    try {
      loading.value = true;
      const errors = await formRef.value?.validate();
      if (errors) {
        return;
      }

      if (isEdit.value && editingCategory.value) {
        await updateLogrotateCategoryApi(
          props.type,
          editingCategory.value,
          formState.name,
          hostId
        );
        Message.success(
          t('app.logrotate.category.manage.message.update_success')
        );
      } else {
        await createLogrotateCategoryApi(props.type, formState.name, hostId);
        Message.success(
          t('app.logrotate.category.manage.message.create_success')
        );
      }

      visible.value = false;
      emit('ok');
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : t('common.error.operation_failed') || 'Operation failed';
      Message.error(errorMessage);
    } finally {
      loading.value = false;
    }
  };

  const handleCancel = () => {
    visible.value = false;
    resetForm();
  };

  const show = async () => {
    visible.value = true;
    resetForm();
    // 等待DOM更新后再重置表单，确保表单正确初始化
    await nextTick();
    formRef.value?.resetFields();
  };

  const setData = (params: CategoryData) => {
    editingCategory.value = params.name;
    formState.name = params.name;
  };

  // 暴露方法
  defineExpose({
    show,
    setData,
  });
</script>
