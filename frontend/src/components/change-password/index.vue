<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('components.changePassword.title')"
    :ok-loading="loading"
    @cancel="handleCancel"
    @ok="handleOk"
  >
    <a-form
      ref="formRef"
      :model="model"
      :rules="rules"
      :label-col-props="formLayoutProps.labelColProps"
      :wrapper-col-props="formLayoutProps.wrapperColProps"
      label-align="left"
    >
      <a-form-item
        field="old_password"
        :label="$t('components.changePassword.old')"
      >
        <a-input-password
          v-model="model.old_password"
          :placeholder="$t('components.changePassword.oldPlaceholder')"
          autocomplete="current-password"
          allow-clear
        />
      </a-form-item>
      <a-form-item
        field="new_password"
        :label="$t('components.changePassword.new')"
      >
        <a-input-password
          v-model="model.new_password"
          :placeholder="$t('components.changePassword.newPlaceholder')"
          autocomplete="new-password"
          allow-clear
        />
      </a-form-item>
      <a-form-item
        field="confirm_password"
        :label="$t('components.changePassword.confirm')"
      >
        <a-input-password
          v-model="model.confirm_password"
          :placeholder="$t('components.changePassword.confirmPlaceholder')"
          autocomplete="new-password"
          allow-clear
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { changePasswordApi } from '@/api/user';
  import { useUserStore } from '@/store';
  import useVisible from '@/composables/visible';
  import useLoading from '@/composables/loading';
  import useLocale from '@/composables/locale';

  const userStore = useUserStore();
  const { t } = useI18n();
  const { currentLocale } = useLocale();
  const { visible, show } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  // 根据当前语言动态调整表单布局
  const formLayoutProps = computed(() => {
    // 英文标签较长，需要更多空间
    if (currentLocale.value === 'en-US') {
      return {
        labelColProps: { span: 8 },
        wrapperColProps: { span: 16 },
      };
    }
    // 中文标签较短，可以使用较小的空间
    return {
      labelColProps: { span: 5 },
      wrapperColProps: { span: 18 },
    };
  });

  const formRef = ref();
  const model = reactive({
    old_password: '',
    new_password: '',
    confirm_password: '',
  });

  const rules = {
    old_password: [
      { required: true, message: t('components.changePassword.oldRequired') },
    ],
    new_password: [
      { required: true, message: t('components.changePassword.newRequired') },
      { minLength: 6, message: t('components.changePassword.lengthError') },
    ],
    confirm_password: [
      {
        required: true,
        message: t('components.changePassword.confirmRequired'),
      },
      {
        validator: (value: string, cb: (error?: string) => void) => {
          if (value !== model.new_password) {
            cb(t('components.changePassword.notMatch'));
          }
          cb();
        },
      },
    ],
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    if (!formRef.value) {
      return;
    }
    if (!(await validate())) {
      return;
    }

    try {
      showLoading();
      await changePasswordApi({
        id: userStore.id,
        old_password: model.old_password,
        password: model.new_password,
      });
      Message.success(t('components.changePassword.success'));
      visible.value = true;
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    model.old_password = '';
    model.new_password = '';
    model.confirm_password = '';
    formRef.value?.clearValidate();
  };

  defineExpose({
    show,
  });
</script>
