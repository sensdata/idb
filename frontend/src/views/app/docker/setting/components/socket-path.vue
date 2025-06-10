<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.setting.socketPath.title')"
    destroy-on-close
    @close="handleClose"
    @before-ok="handleBeforeOk"
  >
    <a-spin :loading="loading" class="w-full">
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="socketPath"
              :label="t('app.docker.setting.socketPath.socket_path')"
            >
              <a-input
                v-model="form.socketPath"
                :placeholder="
                  t('app.docker.setting.socketPath.socket_path.placeholder')
                "
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { updateDockerConfApi } from '@/api/docker';
  import { useConfirm } from '@/hooks/confirm';

  const { t } = useI18n();
  const emit = defineEmits(['ok']);

  const loading = ref(false);
  const visible = ref(false);
  const formRef = ref();

  const form = reactive({
    socketPath: '',
  });

  const rules = reactive({
    socketPath: [
      {
        required: true,
        message: t('app.docker.setting.socketPath.socket_path.required'),
        trigger: 'blur',
      },
      {
        validator(value: string, callback: any) {
          if (!/^\/(.+)$|^tcp:\/\/.+:.+$/i.test(value)) {
            callback(t('app.docker.setting.socketPath.socket_path.format'));
            return;
          }
          callback();
        },
        trigger: 'blur',
      },
    ],
  });

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  const setData = (params: any) => {
    form.socketPath = params.socket_path || '';
  };

  const handleClose = () => {
    visible.value = false;
  };

  const { confirm } = useConfirm();

  const handleBeforeOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) {
      return false;
    }
    const socketPath = form.socketPath.trim();
    if (
      await confirm({
        title: t('app.docker.setting.socketPath.confirm.title'),
        content: t('app.docker.setting.socketPath.confirm.content'),
      })
    ) {
      loading.value = true;
      try {
        await updateDockerConfApi({ key: 'socket_path', value: socketPath });
        emit('ok');
      } catch (err: any) {
        loading.value = false;
        Message.error(err.message);
        return false;
      }
      return true;
    }
    return false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
