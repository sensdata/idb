<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.agent.form.title')"
    width="540px"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="model" :rules="rules">
      <a-form-item
        field="agent_addr"
        :label="$t('manage.host.agent.form.addr.label')"
      >
        <a-input
          v-model="model.agent_addr"
          :placeholder="$t('manage.host.agent.form.addr.placeholder')"
        />
      </a-form-item>
      <a-form-item
        field="agent_port"
        :label="$t('manage.host.agent.form.port.label')"
      >
        <a-input-number
          v-model="model.agent_port"
          :placeholder="$t('manage.host.agent.form.port.placeholder')"
          :min="1"
          :max="65535"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { updateHostAgentApi, getHostInfoApi } from '@/api/host';
  import useVisible from '@/composables/visible';
  import useLoading from '@/composables/loading';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();

  const model = reactive({
    id: 0,
    agent_addr: '',
    agent_port: 0,
  });

  const validatePort = (value: number, callback: (error?: string) => void) => {
    if (!value) {
      callback(t('manage.host.agent.form.port.required'));
      return;
    }
    if (value < 1 || value > 65535) {
      callback(t('manage.host.agent.form.port.invalid'));
      return;
    }
    callback();
  };

  const rules = computed(() => ({
    agent_addr: [
      { required: true, message: t('manage.host.agent.form.addr.required') },
    ],
    agent_port: [{ validator: validatePort, trigger: 'blur' }],
  }));

  const getData = () => {
    const data = toRaw(model);
    return {
      agent_addr: data.agent_addr,
      agent_port: data.agent_port,
    };
  };

  const reset = () => {
    formRef.value?.resetFields();
    formRef.value?.clearValidate();
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

      // 手动验证 agent_addr 非空
      if (!model.agent_addr || model.agent_addr.trim() === '') {
        Message.error(t('manage.host.agent.form.addr.required'));
        return;
      }

      // 手动验证 agent_port 合法
      if (!model.agent_port || model.agent_port < 1 || model.agent_port > 65535) {
        Message.error(t('manage.host.agent.form.port.invalid'));
        return;
      }

      showLoading();
      const data = getData();
      await updateHostAgentApi(model.id, data);
      Message.success(t('manage.host.form.save.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const load = async (hostId: number) => {
    try {
      showLoading();
      const data = await getHostInfoApi(hostId);
      model.id = data.id;
      model.agent_addr = data.agent_addr || '';
      model.agent_port = data.agent_port || 0;
    } catch (err: any) {
      Message.error(t('manage.host.form.load.failed'));
    } finally {
      hideLoading();
    }
  };

  defineExpose({
    show,
    hide,
    reset,
    load,
  });
</script>
