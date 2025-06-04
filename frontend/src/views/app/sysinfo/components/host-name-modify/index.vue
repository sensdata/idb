<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.hostNameModify.title')"
    :ok-loading="loading"
    :width="500"
    @before-ok="handleBeforeOk"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :label-col-props="{ span: 5 }"
      :wrapper-col-props="{ span: 18 }"
    >
      <a-form-item
        :label="$t('app.sysinfo.hostNameModify.host_name')"
        field="host_name"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.hostNameModify.host_name_required'),
          },
          {
            validator: validateHostName,
            message: $t('app.sysinfo.hostNameModify.host_name_invalid'),
          },
        ]"
      >
        <a-input
          v-model="formState.host_name"
          :placeholder="$t('app.sysinfo.hostNameModify.host_name_placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { updateHostNameApi } from '@/api/sysinfo';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref<FormInstance>();
  const formState = reactive({
    host_name: '',
  });

  // 验证主机名格式
  const validateHostName = (
    value: string,
    callback: (error?: string) => void
  ) => {
    // 主机名规则：只允许字母、数字、连字符，不能以连字符开头或结尾
    const hostNameRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/;
    if (!hostNameRegex.test(value)) {
      callback(t('app.sysinfo.hostNameModify.host_name_invalid'));
      return;
    }
    callback();
  };

  // 提交前验证
  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    try {
      const errors = await formRef.value?.validate();
      if (errors) {
        done(false);
        return;
      }

      showLoading();
      try {
        await updateHostNameApi({
          host_name: formState.host_name,
        });

        Message.success(t('app.sysinfo.hostNameModify.update_success'));
        emit('ok');
        done(true);
      } catch (err: any) {
        Message.error(
          err.message || t('app.sysinfo.hostNameModify.update_failed')
        );
        done(false);
      } finally {
        hideLoading();
      }
    } catch (err) {
      done(false);
    }
  };

  // 设置主机名
  const setHostName = (hostName: string) => {
    formState.host_name = hostName;
  };

  defineExpose({
    show,
    hide,
    setHostName,
  });
</script>
